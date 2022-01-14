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

import "io"

// NopWriter represents a type which write operation is nop.
type NopWriter struct{}

func (*NopWriter) Write(buf []byte) (int, error) {
	return len(buf), nil
}

type nopWriteCloser struct {
	io.Writer
}

func (w *nopWriteCloser) Close() error { return nil }

// NopWriteCloser returns a nopWriteCloser.
func NopWriteCloser(w io.Writer) io.WriteCloser {
	return &nopWriteCloser{w}
}

// NopFlusher represents a type which flush operation is nop.
type NopFlusher struct{}

// Flush is a nop operation.
func (f *NopFlusher) Flush() {}

type writeCloserWrapper struct {
	io.Writer
	closer func() error
}

func (r *writeCloserWrapper) Close() error {
	return r.closer()
}

// NewWriteCloserWrapper returns a new io.WriteCloser.
func NewWriteCloserWrapper(r io.Writer, closer func() error) io.WriteCloser {
	return &writeCloserWrapper{
		Writer: r,
		closer: closer,
	}
}

// WriteCounter wraps a concrete io.Writer and hold a count of the number
// of bytes written to the writer during a "session".
// This can be convenient when write return is masked
// (e.g., json.Encoder.Encode())
type WriteCounter struct {
	Count  int64
	Writer io.Writer
}

// NewWriteCounter returns a new WriteCounter.
func NewWriteCounter(w io.Writer) *WriteCounter {
	return &WriteCounter{
		Writer: w,
	}
}

func (wc *WriteCounter) Write(p []byte) (count int, err error) {
	count, err = wc.Writer.Write(p)
	wc.Count += int64(count)
	return
}
