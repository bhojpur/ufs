package filenotify

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
	"fmt"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
)

func TestPollerAddRemove(t *testing.T) {
	w := NewPollingWatcher()

	if err := w.Add("no-such-file"); err == nil {
		t.Fatal("should have gotten error when adding a non-existent file")
	}
	if err := w.Remove("no-such-file"); err == nil {
		t.Fatal("should have gotten error when removing non-existent watch")
	}

	f, err := os.CreateTemp("", "asdf")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(f.Name())

	if err := w.Add(f.Name()); err != nil {
		t.Fatal(err)
	}

	if err := w.Remove(f.Name()); err != nil {
		t.Fatal(err)
	}
}

func TestPollerEvent(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("No chmod on Windows")
	}
	w := NewPollingWatcher()

	f, err := os.CreateTemp("", "test-poller")
	if err != nil {
		t.Fatal("error creating temp file")
	}
	defer os.RemoveAll(f.Name())
	f.Close()

	if err := w.Add(f.Name()); err != nil {
		t.Fatal(err)
	}

	select {
	case <-w.Events():
		t.Fatal("got event before anything happened")
	case <-w.Errors():
		t.Fatal("got error before anything happened")
	default:
	}

	if err := os.WriteFile(f.Name(), []byte("hello"), 0600); err != nil {
		t.Fatal(err)
	}
	assertFileMode(t, f.Name(), 0600)
	if err := assertEvent(w, fsnotify.Write); err != nil {
		t.Fatal(err)
	}

	if err := os.Chmod(f.Name(), 0644); err != nil {
		t.Fatal(err)
	}
	assertFileMode(t, f.Name(), 0644)
	if err := assertEvent(w, fsnotify.Chmod); err != nil {
		t.Fatal(err)
	}

	if err := os.Remove(f.Name()); err != nil {
		t.Fatal(err)
	}
	if err := assertEvent(w, fsnotify.Remove); err != nil {
		t.Fatal(err)
	}
}

func TestPollerClose(t *testing.T) {
	w := NewPollingWatcher()
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	// test double-close
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	f, err := os.CreateTemp("", "asdf")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(f.Name())
	if err := w.Add(f.Name()); err == nil {
		t.Fatal("should have gotten error adding watch for closed watcher")
	}
}

func assertFileMode(t *testing.T, fileName string, mode uint32) {
	t.Helper()
	f, err := os.Stat(fileName)
	if err != nil {
		t.Fatal(err)
	}
	if f.Mode() != os.FileMode(mode) {
		t.Fatalf("expected file %s to have mode %#o, but got %#o", fileName, mode, f.Mode())
	}
}

func assertEvent(w FileWatcher, eType fsnotify.Op) error {
	var err error
	select {
	case e := <-w.Events():
		if e.Op != eType {
			err = fmt.Errorf("got wrong event type, expected %q: %v", eType, e.Op)
		}
	case e := <-w.Errors():
		err = fmt.Errorf("got unexpected error waiting for events %v: %v", eType, e)
	case <-time.After(watchWaitTime * 3):
		err = fmt.Errorf("timeout waiting for event %v", eType)
	}
	return err
}
