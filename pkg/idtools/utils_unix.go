//go:build !windows
// +build !windows

package idtools

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
	"os/exec"
	"path/filepath"
)

func resolveBinary(binname string) (string, error) {
	binaryPath, err := exec.LookPath(binname)
	if err != nil {
		return "", err
	}
	resolvedPath, err := filepath.EvalSymlinks(binaryPath)
	if err != nil {
		return "", err
	}
	// only return no error if the final resolved binary basename
	// matches what was searched for
	if filepath.Base(resolvedPath) == binname {
		return resolvedPath, nil
	}
	return "", fmt.Errorf("Binary %q does not resolve to a binary of that name in $PATH (%q)", binname, resolvedPath)
}

func execCmd(cmd string, arg ...string) ([]byte, error) {
	execCmd := exec.Command(cmd, arg...)
	return execCmd.CombinedOutput()
}
