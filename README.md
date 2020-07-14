# plasmerize

Controlling Midi signals with a Tesla coil Plasma Ball - how convenient!

In short, it is a program to:

* register touches and process via OpenCV (computer vision)
* correlate and triangulate significant touches
* generate Midi Signals and send to Midi Through or device

![](https://github.com/bensinober/plasmerize/blob/master/doc/plasmaball.png?raw=true|width=100)

# Prerequisites

* Golang
* OpenCv
* RtMidi
* low-latency kernel (recommended)

## Ubuntu install:

```
sudo apt-get update && sudo apt-get install -y git zynaddsubfx timidity linux-lowlatency qsynth sfarkxtc libglib2.0-dev libportmidi-dev
```
make sure user is allowed access to usb and audio devices:
```
sudo usermod -a -G plugdev,audio,dialout <user>
```

*NB* If you dont use low latency kernel drivers you'll need to insert some sleep to cam read loop

## Golang and OpenCV

**Go 1.13.1**

```
wget https://dl.google.com/go/go1.13.1.linux-amd64.tar.gz
sudo tar -C /usr/local -xvf go1.13.1.linux-amd64.tar.gz
echo -n "PATH=$PATH:/usr/local/go/bin" >> ~/.bashrc
```

**OpenCV (Computer Vision)**

```
go get -u -d gocv.io/x/gocv
cd $GOPATH/src/gocv.io/x/gocv
make install
```

**RtMidi (Real-Time Midi)**

```
go get -u gitlab.com/gomidi/midi
go get -u gitlab.com/gomidi/midi/reader
go get -u gitlab.com/gomidi/midi/writer
go get -u gitlab.com/gomidi/rtmididrv
#go get -u gitlab.com/bensinober/rtmididrv
```

## Documentation

run with a usb cam on usb port 0 and midi out on device 0, channel 1, and separate dmx out on device 0, channel 2
```go run plasma.go -cam=0 -mid=0 -dmx=0```

run some test notes
```go run plasma.go -test```


Start zynaddsubfx with a sample sound for real time midi instrument
```
zynaddsubfx -a -L /usr/share/zynaddsubfx/banks/Arpeggios/0001-Arpeggio1.xiz
```

Optionally start virtual midi device
```
sudo modprobe snd-virmidi
```

Connect midi ports with aconnect
```
aconnect 24:0 129:0
```

## How does it work?

The webcam sends still images to process with openCV, which converts, saturizes and lowers to hue image.

The hue image is then filtered so a small range of intense pink/red pixels (=touches) are binarized to white.

Countours are registered and trigonometry used to calculate angels and distance from centre of ball.

The various touches are then transformed to midi signals and sent raw to any device given.

Also a subset of angles (360 / 4) are sent to another channel for use by DMX, along with velocity.
