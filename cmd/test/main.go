package main

import (
	"log"
	"os"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
)

var moneyStreamer beep.StreamSeekCloser
var moneySampleRate beep.SampleRate
var receivedStreamer beep.StreamSeekCloser
var receivedSampleRate beep.SampleRate

var sampleRate = beep.SampleRate(22050)

func openStream(path string) (beep.StreamSeekCloser, beep.SampleRate) {
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	st, format, err := wav.Decode(f)
	if err != nil {
		log.Fatal(err)
	}

	return st, format.SampleRate
}

func main() {
	moneyStreamer, moneySampleRate = openStream("money.wav")
	receivedStreamer, receivedSampleRate = openStream("received.wav")

	speaker.Init(sampleRate, sampleRate.N(time.Second/10))

	moneyStreamer.Seek(0)
	speaker.Play(beep.Resample(4, moneySampleRate, sampleRate, moneyStreamer))

	/*receivedStreamer.Seek(0)
	speaker.Play(beep.Resample(4, receivedSampleRate, sampleRate, receivedStreamer))*/

	select {}
}
