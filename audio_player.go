package main

import (
	"os"
	"path/filepath"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
)

type AudioPlayer struct {
	Dir        string
	SampleRate beep.SampleRate
	Audio      *beep.Ctrl
}

func NewAudioPlayer(dir string) *AudioPlayer {
	return &AudioPlayer{Dir: dir}
}

func (p *AudioPlayer) Play(sound string) {
	if p.Audio != nil {
		p.Audio.Paused = true
	}
	file, err := os.Open(filepath.Join(p.Dir, sound))
	HandleErr(err)
	streamer, format, err := wav.Decode(file)
	HandleErr(err)
	p.Audio = &beep.Ctrl{Streamer: streamer, Paused: false}
	if p.SampleRate != format.SampleRate {
		err = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
		HandleErr(err)
	}
	speaker.Play(p.Audio)
}
