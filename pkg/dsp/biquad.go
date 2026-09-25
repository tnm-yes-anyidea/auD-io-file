package dsp

import (
	"math"
)

type FilterType int

const (
	FilterPeaking FilterType = iota
	FilterLowShelf
	FilterHighShelf
)

type Biquad struct {
	b0, b1, b2 float64
	a1, a2     float64
	x1L, x2L   float64
	y1L, y2L   float64
	x1R, x2R   float64
	y1R, y2R   float64
}

func (b *Biquad) Configure(fType FilterType, freq, gainDB, q, sampleRate float64) {
	A := math.Pow(10, gainDB/40.0)
	w0 := 2.0 * math.Pi * freq / sampleRate
	alpha := math.Sin(w0) / (2.0 * q)
	cosW0 := math.Cos(w0)

	var a0 float64

	switch fType {
	case FilterPeaking:
		b.b0 = 1.0 + alpha*A
		b.b1 = -2.0 * cosW0
		b.b2 = 1.0 - alpha*A
		a0 = 1.0 + alpha/A
		b.a1 = -2.0 * cosW0
		b.a2 = 1.0 - alpha/A

	case FilterLowShelf:
		twoSqrtAAlpha := 2.0 * math.Sqrt(A) * alpha
		b.b0 = A * ((A + 1.0) - (A-1.0)*cosW0 + twoSqrtAAlpha)
		b.b1 = 2.0 * A * ((A - 1.0) - (A+1.0)*cosW0)
		b.b2 = A * ((A + 1.0) - (A-1.0)*cosW0 - twoSqrtAAlpha)
		a0 = (A + 1.0) + (A-1.0)*cosW0 + twoSqrtAAlpha
		b.a1 = -2.0 * ((A - 1.0) + (A+1.0)*cosW0)
		b.a2 = (A + 1.0) + (A-1.0)*cosW0 - twoSqrtAAlpha

	case FilterHighShelf:
		twoSqrtAAlpha := 2.0 * math.Sqrt(A) * alpha
		b.b0 = A * ((A + 1.0) + (A-1.0)*cosW0 + twoSqrtAAlpha)
		b.b1 = -2.0 * A * ((A - 1.0) + (A+1.0)*cosW0)
		b.b2 = A * ((A + 1.0) + (A-1.0)*cosW0 - twoSqrtAAlpha)
		a0 = (A + 1.0) - (A-1.0)*cosW0 + twoSqrtAAlpha
		b.a1 = 2.0 * ((A - 1.0) - (A+1.0)*cosW0)
		b.a2 = (A + 1.0) - (A-1.0)*cosW0 - twoSqrtAAlpha
	}

	b.b0 /= a0
	b.b1 /= a0
	b.b2 /= a0
	b.a1 /= a0
	b.a2 /= a0
}

func (b *Biquad) Process(left, right float64) (float64, float64) {
	outL := b.b0*left + b.b1*b.x1L + b.b2*b.x2L - b.a1*b.y1L - b.a2*b.y2L
	b.x2L, b.x1L = b.x1L, left
	b.y2L, b.y1L = b.y1L, outL

	outR := b.b0*right + b.b1*b.x1R + b.b2*b.x2R - b.a1*b.y1R - b.a2*b.y2R
	b.x2R, b.x1R = b.x1R, right
	b.y2R, b.y1R = b.y1R, outR

	return outL, outR
}

type ParametricEQ struct {
	Bands [5]Biquad
}

func NewParametricEQ(sampleRate float64) *ParametricEQ {
	eq := &ParametricEQ{}
	eq.Bands[0].Configure(FilterLowShelf, 60, 0, 0.707, sampleRate)
	eq.Bands[1].Configure(FilterPeaking, 250, 0, 1.4, sampleRate)
	eq.Bands[2].Configure(FilterPeaking, 1000, 0, 1.4, sampleRate)
	eq.Bands[3].Configure(FilterPeaking, 4000, 0, 1.4, sampleRate)
	eq.Bands[4].Configure(FilterHighShelf, 12000, 0, 0.707, sampleRate)
	return eq
}

func (eq *ParametricEQ) UpdateBand(bandIndex int, gainDB float64, sampleRate float64) {
	freqs := []float64{60, 250, 1000, 4000, 12000}
	types := []FilterType{FilterLowShelf, FilterPeaking, FilterPeaking, FilterPeaking, FilterHighShelf}
	if bandIndex >= 0 && bandIndex < 5 {
		eq.Bands[bandIndex].Configure(types[bandIndex], freqs[bandIndex], gainDB, 1.2, sampleRate)
	}
}

func (eq *ParametricEQ) ProcessSample(l, r float64) (float64, float64) {
	curL, curR := l, r
	for i := 0; i < 5; i++ {
		curL, curR = eq.Bands[i].Process(curL, curR)
	}
	return curL, curR
}
