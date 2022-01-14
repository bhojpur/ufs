//go:build !linux
// +build !linux

package homedir

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
)

// GetRuntimeDir is unsupported on non-linux system.
func GetRuntimeDir() (string, error) {
	return "", errors.New("homedir.GetRuntimeDir() is not supported on this system")
}

// StickRuntimeDirContents is unsupported on non-linux system.
func StickRuntimeDirContents(files []string) ([]string, error) {
	return nil, errors.New("homedir.StickRuntimeDirContents() is not supported on this system")
}

// GetDataHome is unsupported on non-linux system.
func GetDataHome() (string, error) {
	return "", errors.New("homedir.GetDataHome() is not supported on this system")
}

// GetConfigHome is unsupported on non-linux system.
func GetConfigHome() (string, error) {
	return "", errors.New("homedir.GetConfigHome() is not supported on this system")
}
