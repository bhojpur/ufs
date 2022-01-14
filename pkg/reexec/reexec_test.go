package reexec

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
	"os"
	"os/exec"
	"testing"

	"gotest.tools/v3/assert"
)

func init() {
	Register("reexec", func() {
		panic("Return Error")
	})
	Init()
}

func TestRegister(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			assert.Equal(t, `reexec func already registered under name "reexec"`, r)
		}
	}()
	Register("reexec", func() {})
}

func TestCommand(t *testing.T) {
	cmd := Command("reexec")
	w, err := cmd.StdinPipe()
	assert.NilError(t, err, "Error on pipe creation: %v", err)
	defer w.Close()

	err = cmd.Start()
	assert.NilError(t, err, "Error on re-exec cmd: %v", err)
	err = cmd.Wait()
	assert.Error(t, err, "exit status 2")
}

func TestNaiveSelf(t *testing.T) {
	if os.Getenv("TEST_CHECK") == "1" {
		os.Exit(2)
	}
	cmd := exec.Command(naiveSelf(), "-test.run=TestNaiveSelf")
	cmd.Env = append(os.Environ(), "TEST_CHECK=1")
	err := cmd.Start()
	assert.NilError(t, err, "Unable to start command")
	err = cmd.Wait()
	assert.Error(t, err, "exit status 2")

	os.Args[0] = "mkdir"
	assert.Check(t, naiveSelf() != os.Args[0])
}
