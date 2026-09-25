package dsp

import (
	"sync"
)

type RingBuffer struct {
	mu       sync.Mutex
	buffer   [][2]float64
	capacity int
	readPos  int
	writePos int
	size     int
}

func NewRingBuffer(capacity int) *RingBuffer {
	return &RingBuffer{
		buffer:   make([][2]float64, capacity),
		capacity: capacity,
	}
}

func (r *RingBuffer) Write(samples [][2]float64) int {
	r.mu.Lock()
	defer r.mu.Unlock()

	written := 0
	for _, sample := range samples {
		if r.size == r.capacity {
			break
		}
		r.buffer[r.writePos] = sample
		r.writePos = (r.writePos + 1) % r.capacity
		r.size++
		written++
	}
	return written
}

func (r *RingBuffer) Read(out [][2]float64) int {
	r.mu.Lock()
	defer r.mu.Unlock()

	read := 0
	for i := 0; i < len(out); i++ {
		if r.size == 0 {
			break
		}
		out[i] = r.buffer[r.readPos]
		r.readPos = (r.readPos + 1) % r.capacity
		r.size--
		read++
	}
	return read
}

func (r *RingBuffer) Available() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.size
}
