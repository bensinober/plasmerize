# plasmerize

Controlling Midi signals with a Tesla coil Plasma Ball - how convenient!

In short, it is a program to:

* register touches and process via OpenCV (computer vision)
* correlate and triangulate significant touches
* generate Midi Signals and send to Midi Through or device

# Prerequisites

* Golang
* OpenCv
* RtMidi
* low-latency kernel (recommended)

## Ubuntu install:

```
sudo apt-get update && install -y git zynaddsubfx timidity linux-lowlatency qsynth sfarkxtc libglib2.0-dev libportmidi-dev
```
make sure user is allowed access to usb and audio devices:
```
sudo usermod -a -G plugdev,audio,dialout <user>
```

## Golang and OpenCV

Go 1.13.1

	wget https://dl.google.com/go/go1.13.1.linux-amd64.tar.gz
	sudo tar -C /usr/local -xvf go1.13.1.linux-amd64.tar.gz
	echo -n "PATH=$PATH:/usr/local/go/bin" >> ~/.bashrc

OpenCV
	go get -u -d gocv.io/x/gocv
	cd $GOPATH/src/gocv.io/x/gocv
	make install

RtMidi
    go get -u gitlab.com/gomidi/midi
	go get -u gitlab.com/bensinober/rtmididrv

## Documentation

run with a usb cam
	go run plasma.go -cam=0

