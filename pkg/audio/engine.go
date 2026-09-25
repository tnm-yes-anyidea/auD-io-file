package audio

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/effects"
	"github.com/gopxl/beep/flac"
	"github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/speaker"
	"github.com/gopxl/beep/vorbis"
	"github.com/gopxl/beep/wav"

	"github.com/tnm-yes-anyidea/auD-io-file/pkg/config"
	"github.com/tnm-yes-anyidea/auD-io-file/pkg/dsp"
)

type RepeatMode int

const (
	RepeatOff RepeatMode = iota
	RepeatAll
	RepeatOne
)

type Engine struct {
	mu           sync.Mutex
	cfg          *config.ProfileConfig
	ctrl         *beep.Ctrl
	volume       *effects.Volume
	streamer     beep.StreamSeekCloser
	format       beep.Format
	eq           *dsp.ParametricEQ
	Visualizer   *dsp.VisualizerRack
	IsPlaying    bool
	CurrentTrack string
	VolumeLevel  float64
	
	// Queue and State
	Queue        []string
	QueueIndex   int
	Repeat       RepeatMode
	OnTrackEnd   func() // Callback for UI updates
}

type DSPWrapper struct {
	source     beep.Streamer
	eq         *dsp.ParametricEQ
	visualizer *dsp.VisualizerRack
}

func (d *DSPWrapper) Stream(samples [][2]float64) (int, bool) {
	n, ok := d.source.Stream(samples)
	if n > 0 {
		for i := 0; i < n; i++ {
			samples[i][0], samples[i][1] = d.eq.ProcessSample(samples[i][0], samples[i][1])
		}
		d.visualizer.FeedAudio(samples[:n])
	}
	return n, ok
}
func (d *DSPWrapper) Err() error { return d.source.Err() }

func NewEngine(cfg *config.ProfileConfig) (*Engine, error) {
	sr := beep.SampleRate(cfg.TargetSampleRate)
	if err := speaker.Init(sr, sr.N(time.Millisecond*100)); err != nil {
		return nil, err
	}

	return &Engine{
		cfg:         cfg,
		Visualizer:  dsp.NewVisualizerRack(cfg.FFTWindowSize, cfg.SpectrumBands),
		eq:          dsp.NewParametricEQ(float64(cfg.TargetSampleRate)),
		VolumeLevel: 0,
		Queue:       make([]string, 0),
		Repeat:      RepeatOff,
	}, nil
}

func (e *Engine) PlayQueueIndex(index int) error {
	e.mu.Lock()
	if index < 0 || index >= len(e.Queue) {
		e.mu.Unlock()
		return fmt.Errorf("index out of bounds")
	}
	filePath := e.Queue[index]
	e.QueueIndex = index
	e.mu.Unlock()
	return e.Play(filePath)
}

func (e *Engine) Play(filePath string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.streamer != nil { _ = e.streamer.Close() }

	f, err := os.Open(filePath)
	if err != nil { return err }

	ext := strings.ToLower(filepath.Ext(filePath))
	var streamer beep.StreamSeekCloser
	var format beep.Format

	switch ext {
	case ".flac": streamer, format, err = flac.Decode(f)
	case ".mp3":  streamer, format, err = mp3.Decode(f)
	case ".ogg":  streamer, format, err = vorbis.Decode(f)
	case ".wav":  streamer, format, err = wav.Decode(f)
	default: f.Close(); return fmt.Errorf("unsupported")
	}
	if err != nil { return err }

	e.streamer = streamer
	e.format = format
	e.CurrentTrack = filePath

	var finalStreamer beep.Streamer = streamer
	if format.SampleRate != beep.SampleRate(e.cfg.TargetSampleRate) {
		finalStreamer = beep.Resample(4, format.SampleRate, beep.SampleRate(e.cfg.TargetSampleRate), streamer)
	}

	dspNode := &DSPWrapper{source: finalStreamer, eq: e.eq, visualizer: e.Visualizer}
	
	// Callback when track finishes
	seq := beep.Seq(dspNode, beep.Callback(func() {
		go e.trackFinished()
	}))

	e.volume = &effects.Volume{Streamer: seq, Base: 2, Volume: e.VolumeLevel}
	e.ctrl = &beep.Ctrl{Streamer: e.volume, Paused: false}

	speaker.Clear()
	speaker.Play(e.ctrl)
	e.IsPlaying = true

	return nil
}

func (e *Engine) trackFinished() {
	e.mu.Lock()
	repeat := e.Repeat
	idx := e.QueueIndex
	queueLen := len(e.Queue)
	e.mu.Unlock()

	if queueLen == 0 { return }

	if repeat == RepeatOne {
		_ = e.PlayQueueIndex(idx)
	} else if idx+1 < queueLen {
		_ = e.PlayQueueIndex(idx + 1)
	} else if repeat == RepeatAll {
		_ = e.PlayQueueIndex(0)
	} else {
		e.Stop()
	}

	if e.OnTrackEnd != nil { e.OnTrackEnd() }
}

func (e *Engine) Seek(d time.Duration) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.streamer == nil { return }
	
	speaker.Lock()
	pos := e.format.SampleRate.N(d)
	if pos < 0 { pos = 0 }
	if pos >= e.streamer.Len() { pos = e.streamer.Len() - 1 }
	_ = e.streamer.Seek(pos)
	speaker.Unlock()
}

func (e *Engine) SeekRelative(d time.Duration) {
	pos := e.Position() + d
	e.Seek(pos)
}

func (e *Engine) Position() time.Duration {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.streamer == nil { return 0 }
	return e.format.SampleRate.D(e.streamer.Position())
}

func (e *Engine) Length() time.Duration {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.streamer == nil { return 0 }
	return e.format.SampleRate.D(e.streamer.Len())
}

func (e *Engine) TogglePause() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.ctrl != nil {
		speaker.Lock()
		e.ctrl.Paused = !e.ctrl.Paused
		e.IsPlaying = !e.ctrl.Paused
		speaker.Unlock()
	}
}

func (e *Engine) Stop() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.ctrl != nil {
		speaker.Lock()
		e.ctrl.Paused = true
		e.IsPlaying = false
		e.CurrentTrack = ""
		speaker.Unlock()
	}
}

func (e *Engine) SetVolume(delta float64) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.VolumeLevel += delta
	if e.VolumeLevel > 5.0 { e.VolumeLevel = 5.0 }
	if e.VolumeLevel < -10.0 { e.VolumeLevel = -10.0 }
	if e.volume != nil {
		speaker.Lock()
		e.volume.Volume = e.VolumeLevel
		speaker.Unlock()
	}
}

func (e *Engine) SetEQGain(bandIndex int, gainDB float64) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.eq.UpdateBand(bandIndex, gainDB, float64(e.cfg.TargetSampleRate))
}
