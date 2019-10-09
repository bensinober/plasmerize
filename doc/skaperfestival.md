# PLAN SKAPERFESTIVALEN

## Utstyr

PC med ubuntu bionic
Plasmakule m/strøm
Webkamera type Axs M3045
POE adapter (til kamera)
Switch
Launchpad MK mini
https://www.websynths.com/ (web synth)

## Install

low-latency kernel +diverse miditools og golang:

```
sudo apt-get update && install -y git zynaddsubfx timidity linux-lowlatency qsynth sfarkxtc libglib2.0-dev libportmidi-dev
```
make sure user is allowed access to usb and audio devices:
```
sudo usermod -a -G plugdev,audio,dialout <user>
```

Install Golang
Go 1.11 (Siden midistøtte er litt flaky fom. 1.12)
eller go 1.13 og bruk mitt fork av rtmididrv
```
wget https://dl.google.com/go/go1.11.13.linux-amd64.tar.gz
sudo tar -C /usr/local -xvf go1.11.13.linux-amd64.tar.gz
echo -n "PATH=$PATH:/usr/local/go/bin" >> ~/.bashrc
```

Checkout repo:

	git clone https://github.com/bensinober/plasmerize
	cd plasmerize
	go get -v ./...

OpenCV
	go get -u -d gocv.io/x/gocv
	cd $GOPATH/src/gocv.io/x/gocv
	make install

Midi (RealTimeMidi)
    go get -u gitlab.com/gomidi/midi
	go get -u gitlab.com/bensinober/rtmididrv
