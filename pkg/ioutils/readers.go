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
	"context"
	"io"

	// make sure crypto.SHA256, crypto.sha512 and crypto.SHA384 are registered
	_ "crypto/sha256"
	_ "crypto/sha512"
)

// ReadCloserWrapper wraps an io.Reader, and implements an io.ReadCloser
// It calls the given callback function when closed. It should be constructed
// with NewReadCloserWrapper
type ReadCloserWrapper struct {
	io.Reader
	closer func() error
}

// Close calls back the passed closer function
func (r *ReadCloserWrapper) Close() error {
	return r.closer()
}

// NewReadCloserWrapper returns a new io.ReadCloser.
func NewReadCloserWrapper(r io.Reader, closer func() error) io.ReadCloser {
	return &ReadCloserWrapper{
		Reader: r,
		closer: closer,
	}
}

type readerErrWrapper struct {
	reader io.Reader
	closer func()
}

func (r *readerErrWrapper) Read(p []byte) (int, error) {
	n, err := r.reader.Read(p)
	if err != nil {
		r.closer()
	}
	return n, err
}

// NewReaderErrWrapper returns a new io.Reader.
func NewReaderErrWrapper(r io.Reader, closer func()) io.Reader {
	return &readerErrWrapper{
		reader: r,
		closer: closer,
	}
}

// OnEOFReader wraps an io.ReadCloser and a function
// the function will run at the end of file or close the file.
type OnEOFReader struct {
	Rc io.ReadCloser
	Fn func()
}

func (r *OnEOFReader) Read(p []byte) (n int, err error) {
	n, err = r.Rc.Read(p)
	if err == io.EOF {
		r.runFunc()
	}
	return
}

// Close closes the file and run the function.
func (r *OnEOFReader) Close() error {
	err := r.Rc.Close()
	r.runFunc()
	return err
}

func (r *OnEOFReader) runFunc() {
	if fn := r.Fn; fn != nil {
		fn()
		r.Fn = nil
	}
}

// cancelReadCloser wraps an io.ReadCloser with a context for cancelling read
// operations.
type cancelReadCloser struct {
	cancel func()
	pR     *io.PipeReader // Stream to read from
	pW     *io.PipeWriter
}

// NewCancelReadCloser creates a wrapper that closes the ReadCloser when the
// context is cancelled. The returned io.ReadCloser must be closed when it is
// no longer needed.
func NewCancelReadCloser(ctx context.Context, in io.ReadCloser) io.ReadCloser {
	pR, pW := io.Pipe()

	// Create a context used to signal when the pipe is closed
	doneCtx, cancel := context.WithCancel(context.Background())

	p := &cancelReadCloser{
		cancel: cancel,
		pR:     pR,
		pW:     pW,
	}

	go func() {
		_, err := io.Copy(pW, in)
		select {
		case <-ctx.Done():
			// If the context was closed, p.closeWithError
			// was already called. Calling it again would
			// change the error that Read returns.
		default:
			p.closeWithError(err)
		}
		in.Close()
	}()
	go func() {
		for {
			select {
			case <-ctx.Done():
				p.closeWithError(ctx.Err())
			case <-doneCtx.Done():
				return
			}
		}
	}()

	return p
}

// Read wraps the Read method of the pipe that provides data from the wrapped
// ReadCloser.
func (p *cancelReadCloser) Read(buf []byte) (n int, err error) {
	return p.pR.Read(buf)
}

// closeWithError closes the wrapper and its underlying reader. It will
// cause future calls to Read to return err.
func (p *cancelReadCloser) closeWithError(err error) {
	p.pW.CloseWithError(err)
	p.cancel()
}

// Close closes the wrapper its underlying reader. It will cause
// future calls to Read to return io.EOF.
func (p *cancelReadCloser) Close() error {
	p.closeWithError(io.EOF)
	return nil
}
