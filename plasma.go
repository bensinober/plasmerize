package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"math"
	"math/rand"
	"strconv"
	"time"

	_ "net/http/pprof"

	driver "gitlab.com/bensinober/rtmididrv"
	"gitlab.com/gomidi/midi/mid"
	"gocv.io/x/gocv"
)

type Plasmer struct {
	Midi    *mid.Writer
	Cam     *gocv.VideoCapture
	NotesOn map[uint8]midiNote
	NotesIn chan map[uint8]midiNote
}

type midiNote struct {
	Note, Velo uint8
	Timer      *time.Timer
}

func newPlasmer(mw *mid.Writer, max int, cam *gocv.VideoCapture) *Plasmer {
	return &Plasmer{
		Midi:    mw,
		Cam:     cam,
		NotesOn: make(map[uint8]midiNote, max),
		NotesIn: make(chan map[uint8]midiNote, max),
	}
}

func findCentroid(img gocv.Mat, ctr []image.Point) image.Point {
	mat := gocv.NewMatWithSize(img.Rows(), img.Cols(), gocv.MatTypeCV8U)
	gocv.FillPoly(&mat, [][]image.Point{ctr}, color.RGBA{R: 255, G: 255, B: 255, A: 255})
	m := gocv.Moments(mat, false) // binaryImage = false
	cx := int(m["m10"] / m["m00"])
	cy := int(m["m01"] / m["m00"])
	return image.Point{X: cx, Y: cy}
}

/* opt 1) straight vectors x,y adapted to midi 0-127 range */
func vectorTouchToMidi(p image.Point) midiNote {
	n := uint8((p.X - 20) * 125 / 600)
	v := uint8((p.Y - 20) * 125 / 440)
	fmt.Printf("vectorTouchToMidi: (%dx,%dy) -> (note %d, velo %d)\n", p.X, p.Y, n, v)
	return midiNote{Note: n, Velo: v}
}

/* opt 2) angle and distance from center */
func angleDistToMidi(p image.Point) midiNote {
	b := math.Abs(float64(p.X) - 320.0) // length of a
	a := math.Abs(float64(p.Y) - 240.0) // length of b
	/* angle */
	var ang float64
	if p.Y > 240 {
		ang = math.Atan(b/a) * 180 / math.Pi
	} else {
		ang = (math.Atan(a/b) * 180 / math.Pi) + 90
	}

	/* dist */
	dist := math.Sqrt(math.Pow(b, 2) + math.Pow(a, 2))
	/* note, velo */
	n := uint8(ang / 180 * 127)
	v := uint8(dist / 240 * 127)
	fmt.Printf("angleDistToMidi: (%dx,%dy) -> (ang %f, dist %f) -> (note %d, velo %d)\n", p.X, p.Y, ang, dist, n, v)
	return midiNote{Note: n, Velo: v}
}

func notesPressed(img gocv.Mat, pts [][]image.Point) map[uint8]midiNote {
	nts := make(map[uint8]midiNote, 10)
	for _, ctrd := range pts {
		c := findCentroid(img, ctrd)
		n := angleDistToMidi(c)
		nts[n.Note] = n
	}
	return nts
}

// remove smaller contours within an X-sized area
func filterContours(pts [][]image.Point) [][]image.Point {
	minArea := 40.0
	var filtered [][]image.Point
	for _, c := range pts {

		area := gocv.ContourArea(c)
		if area > minArea {
			filtered = append(filtered, c)
		}
	}
	return filtered
}

/* incoming notes channel */
func (p *Plasmer) handlePressedNotes() {
	for nts := range p.NotesIn {
		for i, note := range nts {
			timer := time.NewTimer(time.Second * 3)
			if active, ok := p.NotesOn[i]; ok {
				active.Timer = timer
			} else {
				note.Timer = timer
				p.NotesOn[i] = note
				p.Midi.NoteOn(note.Note, note.Velo)
				fmt.Printf("pressing note %d, velo %d\n", note.Note, note.Velo)
			}
		}
	}
}

func (p *Plasmer) expireNotes() {
	for {
		for i, note := range p.NotesOn {
			select {
			case <-note.Timer.C:
				fmt.Printf("releasing note %d\n", i)
				p.Midi.NoteOff(i)
				delete(p.NotesOn, i)
			}
		}
	}
}

func (p *Plasmer) readCam() {
	window1 := gocv.NewWindow("plasma points detector 1")
	window2 := gocv.NewWindow("plasma points detector 2")
	window3 := gocv.NewWindow("plasma points detector 3")
	defer window1.Close()
	defer window2.Close()
	defer window3.Close()
	img := gocv.NewMat()
	img2 := gocv.NewMat()
	hueImg := gocv.NewMat()
	mask := gocv.NewMat()
	defer img.Close()
	defer img2.Close()
	defer hueImg.Close()
	defer mask.Close()
	green := color.RGBA{0, 255, 0, 0}
	for {
		if ok := p.Cam.Read(&img); !ok {
			fmt.Printf("Device closed: %v\n", p.Cam)
			return
		}
		if img.Empty() {
			continue
		}
		//img.CopyTo(&img2)
		gocv.Flip(img, &img2, 1)                                                   // flip horizontally
		gocv.CvtColor(img2, &hueImg, gocv.ColorBGRToHSV)                           // convert to hue
		gocv.Circle(&img2, image.Pt(320, 330), 68, color.RGBA{0, 0, 255, 0}, -1)   // exclude center
		gocv.Circle(&hueImg, image.Pt(320, 330), 68, color.RGBA{0, 0, 255, 0}, -1) // exclude center

		// HUE-SATURATION-VUE spectrum: https://i.stack.imgur.com/gyuw4.png
		// extract the pinkish red hue range to mask Mat
		gocv.InRangeWithScalar(hueImg, gocv.NewScalar(150.0, 100.0, 250.0, 0.0), gocv.NewScalar(170.0, 255.0, 255.0, 0.0), &mask)
		ctrs := gocv.FindContours(mask, gocv.RetrievalExternal, gocv.ChainApproxSimple)
		fCtrs := filterContours(ctrs)
		for _, ctr := range ctrs {
			gocv.Circle(&img2, (ctr[0]), 4, green, 2)
		}

		n := notesPressed(img2, fCtrs)
		p.NotesIn <- n

		window1.IMShow(img2)
		window2.IMShow(mask)
		window3.IMShow(hueImg)
		if window1.WaitKey(1) == 27 {
			break
		}
	}
}

func main() {
	cam := flag.String("cam", "", "address of webcam, http or id int")
	max := flag.Int("max", 8, "max no of simultaneous notes")
	test := flag.Bool("test", false, "run test mode")
	flag.Parse()
	// device or url to mjpeg stream
	// go run plasma.go -cam 0 / http://root:pass@192.168.1.2/mjpg/1/video.mjpg
	var webcam *gocv.VideoCapture
	if *cam != "" {
		if i, err := strconv.Atoi(*cam); err == nil {
			webcam, err = gocv.OpenVideoCapture(i)
			if err != nil {
				fmt.Printf("Error opening video capture device: %v\n", err)
				return
			}
		} else {
			webcam, err = gocv.OpenVideoCapture(*cam)
			if err != nil {
				fmt.Printf("Error opening video capture device: %v\n", err)
				return
			}
		}
	}
	fmt.Println(webcam)
	defer webcam.Close()

	// midi setup
	drv, err := driver.New()
	if err != nil {
		panic(err)
	}
	defer drv.Close()
	outs, err := drv.Outs()
	//fmt.Printf("%v", outs)
	if err != nil {
		panic(err)
	}
	if err := outs[0].Open(); err != nil {
		panic(err)
	}
	midiWr := mid.ConnectOut(outs[0])
	midiWr.SetChannel(0)
	p := newPlasmer(midiWr, *max, webcam)

	go p.handlePressedNotes()
	go p.expireNotes()
	if *cam != "" {
		p.readCam()
	}
	if *test {
		window := gocv.NewWindow("plasma points detector")
		img := gocv.NewMatWithSize(640, 480, gocv.MatTypeCV8U)
		defer img.Close()
		gocv.Circle(&img, image.Pt(330, 280), 220, color.RGBA{0, 0, 255, 0}, -1) // test
		for i := 0; i < 12; i++ {
			var pts []image.Point
			var ctrs [][]image.Point
			x := rand.Intn(600) + 20
			y := rand.Intn(420) + 20
			pt := image.Point{X: x, Y: y}
			pts = append(pts, pt)
			ctrs = append(ctrs, pts)
			fmt.Printf("Random touch: (%dx, %dy)\n", x, y)
			gocv.Circle(&img, image.Pt(x, y), 10, color.RGBA{0, 0, 0, 0}, -1) // test
			window.IMShow(img)
			n := notesPressed(img, ctrs)
			p.NotesIn <- n
			time.Sleep(1000 * 1000 * 500 * time.Nanosecond)
			if window.WaitKey(1) == 27 {
				break
			}
		}
		time.Sleep(5 * time.Second) // for allowing channels to finish reading
	}
}