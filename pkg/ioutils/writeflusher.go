package ioutils

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

// WriteFlusher wraps the Write and Flush operation ensuring that every write
// is a flush. In addition, the Close method can be called to intercept
// Read/Write calls if the targets lifecycle has already ended.
type WriteFlusher struct {
	w           io.Writer
	flusher     flusher
	flushed     chan struct{}
	flushedOnce sync.Once
	closed      chan struct{}
	closeLock   sync.Mutex
}

type flusher interface {
	Flush()
}

var errWriteFlusherClosed = io.EOF

func (wf *WriteFlusher) Write(b []byte) (n int, err error) {
	select {
	case <-wf.closed:
		return 0, errWriteFlusherClosed
	default:
	}

	n, err = wf.w.Write(b)
	wf.Flush() // every write is a flush.
	return n, err
}

// Flush the stream immediately.
func (wf *WriteFlusher) Flush() {
	select {
	case <-wf.closed:
		return
	default:
	}

	wf.flushedOnce.Do(func() {
		close(wf.flushed)
	})
	wf.flusher.Flush()
}

// Flushed returns the state of flushed.
// If it's flushed, return true, or else it return false.
func (wf *WriteFlusher) Flushed() bool {
	// BUG(stevvooe): Remove this method. Its use is inherently racy. Seems to
	// be used to detect whether or a response code has been issued or not.
	// Another hook should be used instead.
	var flushed bool
	select {
	case <-wf.flushed:
		flushed = true
	default:
	}
	return flushed
}

// Close closes the write flusher, disallowing any further writes to the
// target. After the flusher is closed, all calls to write or flush will
// result in an error.
func (wf *WriteFlusher) Close() error {
	wf.closeLock.Lock()
	defer wf.closeLock.Unlock()

	select {
	case <-wf.closed:
		return errWriteFlusherClosed
	default:
		close(wf.closed)
	}
	return nil
}

// NewWriteFlusher returns a new WriteFlusher.
func NewWriteFlusher(w io.Writer) *WriteFlusher {
	var fl flusher
	if f, ok := w.(flusher); ok {
		fl = f
	} else {
		fl = &NopFlusher{}
	}
	return &WriteFlusher{w: w, flusher: fl, closed: make(chan struct{}), flushed: make(chan struct{})}
}
