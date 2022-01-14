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
	"bytes"
	"strings"
	"testing"
)

func TestWriteCloserWrapperClose(t *testing.T) {
	called := false
	writer := bytes.NewBuffer([]byte{})
	wrapper := NewWriteCloserWrapper(writer, func() error {
		called = true
		return nil
	})
	if err := wrapper.Close(); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatalf("writeCloserWrapper should have call the anonymous function.")
	}
}

func TestNopWriteCloser(t *testing.T) {
	writer := bytes.NewBuffer([]byte{})
	wrapper := NopWriteCloser(writer)
	if err := wrapper.Close(); err != nil {
		t.Fatal("NopWriteCloser always return nil on Close.")
	}

}

func TestNopWriter(t *testing.T) {
	nw := &NopWriter{}
	l, err := nw.Write([]byte{'c'})
	if err != nil {
		t.Fatal(err)
	}
	if l != 1 {
		t.Fatalf("Expected 1 got %d", l)
	}
}

func TestWriteCounter(t *testing.T) {
	dummy1 := "This is a dummy string."
	dummy2 := "This is another dummy string."
	totalLength := int64(len(dummy1) + len(dummy2))

	reader1 := strings.NewReader(dummy1)
	reader2 := strings.NewReader(dummy2)

	var buffer bytes.Buffer
	wc := NewWriteCounter(&buffer)

	reader1.WriteTo(wc)
	reader2.WriteTo(wc)

	if wc.Count != totalLength {
		t.Errorf("Wrong count: %d vs. %d", wc.Count, totalLength)
	}

	if buffer.String() != dummy1+dummy2 {
		t.Error("Wrong message written")
	}
}
