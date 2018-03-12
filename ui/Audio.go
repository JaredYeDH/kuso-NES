package ui

import "github.com/gordonklaus/portaudio"

type Audio struct {
	stream     *portaudio.Stream
	sampleRate float64
	outChans   int
	channel    chan float32
}

func NewAudio() *Audio {
	channel := make(chan float32, 44100) // 44.1kHz
	return &Audio{channel: channel}
}

func (a *Audio) CallBack(out []float32) {
	var o float32
	for i := range out {
		if i%a.outChans == 0 {
			select {
			case sample := <-a.channel:
				o = sample
			default:
				o = 0
			}
		}
		out[i] = o
	}
}

func (a *Audio) Start() error {
	hostapi, err := portaudio.DefaultHostApi()
	if err != nil {
		return err
	}
	para := portaudio.HighLatencyParameters(nil, hostapi.DefaultOutputDevice)
	stream, err := portaudio.OpenStream(para, a.CallBack)
	if err != nil {
		return err
	}
	a.stream = stream
	a.sampleRate = para.SampleRate
	a.outChans = para.Output.Channels
	return nil
}

func (a *Audio) Stop() {
	a.stream.Close()
}
