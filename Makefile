all: gooper

gooper: pkg
	go build gooper

pkg:
	go get code.google.com/p/portaudio-go/portaudio

clean:
	rm gooper

.PHONY: all clean
