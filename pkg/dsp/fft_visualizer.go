package dsp

import (
	"math"
	"math/cmplx"
	"sync"
)

func FFT(x []complex128) {
	n := len(x)
	if n <= 1 {
		return
	}

	j := 0
	for i := 0; i < n-1; i++ {
		if i < j {
			x[i], x[j] = x[j], x[i]
		}
		k := n >> 1
		for k <= j {
			j -= k
			k >>= 1
		}
		j += k
	}

	for l := 2; l <= n; l <<= 1 {
		halfL := l >> 1
		theta := -2.0 * math.Pi / float64(l)
		wStep := cmplx.Rect(1, theta)
		for i := 0; i < n; i += l {
			w := complex(1.0, 0.0)
			for m := 0; m < halfL; m++ {
				u := x[i+m]
				v := x[i+m+halfL] * w
				x[i+m] = u + v
				x[i+m+halfL] = u - v
				w *= wStep
			}
		}
	}
}

type VisualizerRack struct {
	mu           sync.Mutex
	windowSize   int
	bandsCount   int
	fftBuffer    []complex128
	windowWindow []float64
	LastSpectrum []float64
	LastScopeL   []float64
	LastScopeR   []float64
}

func NewVisualizerRack(windowSize, bandsCount int) *VisualizerRack {
	v := &VisualizerRack{
		windowSize:   windowSize,
		bandsCount:   bandsCount,
		fftBuffer:    make([]complex128, windowSize),
		windowWindow: make([]float64, windowSize),
		LastSpectrum: make([]float64, bandsCount),
		LastScopeL:   make([]float64, 64),
		LastScopeR:   make([]float64, 64),
	}
	for i := 0; i < windowSize; i++ {
		v.windowWindow[i] = 0.5 * (1.0 - math.Cos(2.0*math.Pi*float64(i)/float64(windowSize-1)))
	}
	return v
}

func (v *VisualizerRack) FeedAudio(samples [][2]float64) {
	v.mu.Lock()
	defer v.mu.Unlock()

	sLen := len(samples)
	if sLen == 0 {
		return
	}

	scopeLen := len(v.LastScopeL)
	step := sLen / scopeLen
	if step < 1 {
		step = 1
	}
	for i := 0; i < scopeLen && i*step < sLen; i++ {
		v.LastScopeL[i] = samples[i*step][0]
		v.LastScopeR[i] = samples[i*step][1]
	}

	copyCount := v.windowSize
	if sLen < copyCount {
		copyCount = sLen
	}
	for i := 0; i < copyCount; i++ {
		mono := (samples[i][0] + samples[i][1]) * 0.5
		v.fftBuffer[i] = complex(mono*v.windowWindow[i], 0)
	}
	for i := copyCount; i < v.windowSize; i++ {
		v.fftBuffer[i] = 0
	}

	FFT(v.fftBuffer)

	halfWindow := v.windowSize / 2
	samplesPerBand := halfWindow / v.bandsCount
	if samplesPerBand < 1 {
		samplesPerBand = 1
	}

	for b := 0; b < v.bandsCount; b++ {
		var magSum float64
		start := b * samplesPerBand
		end := start + samplesPerBand
		if end > halfWindow {
			end = halfWindow
		}

		for k := start; k < end; k++ {
			magSum += cmplx.Abs(v.fftBuffer[k])
		}

		avgMag := magSum / float64(end-start)
		v.LastSpectrum[b] = v.LastSpectrum[b]*0.4 + avgMag*0.6
	}
}

func (v *VisualizerRack) Snapshot() ([]float64, []float64, []float64) {
	v.mu.Lock()
	defer v.mu.Unlock()

	spec := make([]float64, len(v.LastSpectrum))
	copy(spec, v.LastSpectrum)
	scopeL := make([]float64, len(v.LastScopeL))
	copy(scopeL, v.LastScopeL)
	scopeR := make([]float64, len(v.LastScopeR))
	copy(scopeR, v.LastScopeR)

	return spec, scopeL, scopeR
}
