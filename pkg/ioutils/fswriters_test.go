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
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

var (
	testMode os.FileMode = 0640
)

func init() {
	// Windows does not support full Linux file mode
	if runtime.GOOS == "windows" {
		testMode = 0666
	}
}

func TestAtomicWriteToFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "atomic-writers-test")
	if err != nil {
		t.Fatalf("Error when creating temporary directory: %s", err)
	}
	defer os.RemoveAll(tmpDir)

	expected := []byte("barbaz")
	if err := AtomicWriteFile(filepath.Join(tmpDir, "foo"), expected, testMode); err != nil {
		t.Fatalf("Error writing to file: %v", err)
	}

	actual, err := os.ReadFile(filepath.Join(tmpDir, "foo"))
	if err != nil {
		t.Fatalf("Error reading from file: %v", err)
	}

	if !bytes.Equal(actual, expected) {
		t.Fatalf("Data mismatch, expected %q, got %q", expected, actual)
	}

	st, err := os.Stat(filepath.Join(tmpDir, "foo"))
	if err != nil {
		t.Fatalf("Error statting file: %v", err)
	}
	if expected := testMode; st.Mode() != expected {
		t.Fatalf("Mode mismatched, expected %o, got %o", expected, st.Mode())
	}
}

func TestAtomicWriteSetCommit(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "atomic-writerset-test")
	if err != nil {
		t.Fatalf("Error when creating temporary directory: %s", err)
	}
	defer os.RemoveAll(tmpDir)

	if err := os.Mkdir(filepath.Join(tmpDir, "tmp"), 0700); err != nil {
		t.Fatalf("Error creating tmp directory: %s", err)
	}

	targetDir := filepath.Join(tmpDir, "target")
	ws, err := NewAtomicWriteSet(filepath.Join(tmpDir, "tmp"))
	if err != nil {
		t.Fatalf("Error creating atomic write set: %s", err)
	}

	expected := []byte("barbaz")
	if err := ws.WriteFile("foo", expected, testMode); err != nil {
		t.Fatalf("Error writing to file: %v", err)
	}

	if _, err := os.ReadFile(filepath.Join(targetDir, "foo")); err == nil {
		t.Fatalf("Expected error reading file where should not exist")
	}

	if err := ws.Commit(targetDir); err != nil {
		t.Fatalf("Error committing file: %s", err)
	}

	actual, err := os.ReadFile(filepath.Join(targetDir, "foo"))
	if err != nil {
		t.Fatalf("Error reading from file: %v", err)
	}

	if !bytes.Equal(actual, expected) {
		t.Fatalf("Data mismatch, expected %q, got %q", expected, actual)
	}

	st, err := os.Stat(filepath.Join(targetDir, "foo"))
	if err != nil {
		t.Fatalf("Error statting file: %v", err)
	}
	if expected := testMode; st.Mode() != expected {
		t.Fatalf("Mode mismatched, expected %o, got %o", expected, st.Mode())
	}

}

func TestAtomicWriteSetCancel(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "atomic-writerset-test")
	if err != nil {
		t.Fatalf("Error when creating temporary directory: %s", err)
	}
	defer os.RemoveAll(tmpDir)

	if err := os.Mkdir(filepath.Join(tmpDir, "tmp"), 0700); err != nil {
		t.Fatalf("Error creating tmp directory: %s", err)
	}

	ws, err := NewAtomicWriteSet(filepath.Join(tmpDir, "tmp"))
	if err != nil {
		t.Fatalf("Error creating atomic write set: %s", err)
	}

	expected := []byte("barbaz")
	if err := ws.WriteFile("foo", expected, testMode); err != nil {
		t.Fatalf("Error writing to file: %v", err)
	}

	if err := ws.Cancel(); err != nil {
		t.Fatalf("Error committing file: %s", err)
	}

	if _, err := os.ReadFile(filepath.Join(tmpDir, "target", "foo")); err == nil {
		t.Fatalf("Expected error reading file where should not exist")
	} else if !os.IsNotExist(err) {
		t.Fatalf("Unexpected error reading file: %s", err)
	}
}
