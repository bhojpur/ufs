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
	"errors"
	"io"
)

var errBufferFull = errors.New("buffer is full")

type fixedBuffer struct {
	buf      []byte
	pos      int
	lastRead int
}

func (b *fixedBuffer) Write(p []byte) (int, error) {
	n := copy(b.buf[b.pos:cap(b.buf)], p)
	b.pos += n

	if n < len(p) {
		if b.pos == cap(b.buf) {
			return n, errBufferFull
		}
		return n, io.ErrShortWrite
	}
	return n, nil
}

func (b *fixedBuffer) Read(p []byte) (int, error) {
	n := copy(p, b.buf[b.lastRead:b.pos])
	b.lastRead += n
	return n, nil
}

func (b *fixedBuffer) Len() int {
	return b.pos - b.lastRead
}

func (b *fixedBuffer) Cap() int {
	return cap(b.buf)
}

func (b *fixedBuffer) Reset() {
	b.pos = 0
	b.lastRead = 0
	b.buf = b.buf[:0]
}

func (b *fixedBuffer) String() string {
	return string(b.buf[b.lastRead:b.pos])
}
