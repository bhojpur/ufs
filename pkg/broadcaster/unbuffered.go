package broadcaster

// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

import (
	"io"
	"sync"
)

// Unbuffered accumulates multiple io.WriteCloser by stream.
type Unbuffered struct {
	mu      sync.Mutex
	writers []io.WriteCloser
}

// Add adds new io.WriteCloser.
func (w *Unbuffered) Add(writer io.WriteCloser) {
	w.mu.Lock()
	w.writers = append(w.writers, writer)
	w.mu.Unlock()
}

// Write writes bytes to all writers. Failed writers will be evicted during
// this call.
func (w *Unbuffered) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	var evict []int
	for i, sw := range w.writers {
		if n, err := sw.Write(p); err != nil || n != len(p) {
			// On error, evict the writer
			evict = append(evict, i)
		}
	}
	for n, i := range evict {
		w.writers = append(w.writers[:i-n], w.writers[i-n+1:]...)
	}
	w.mu.Unlock()
	return len(p), nil
}

// Clean closes and removes all writers. Last non-eol-terminated part of data
// will be saved.
func (w *Unbuffered) Clean() error {
	w.mu.Lock()
	for _, sw := range w.writers {
		sw.Close()
	}
	w.writers = nil
	w.mu.Unlock()
	return nil
}
