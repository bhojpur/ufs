package filesys

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
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows"
)

// GetLongPathName converts Windows short pathnames to full pathnames.
// For example C:\Users\ADMIN~1 --> C:\Users\Administrator.
// It is a no-op on non-Windows platforms
func GetLongPathName(path string) (string, error) {
	p, err := windows.UTF16FromString(path)
	if err != nil {
		return "", err
	}
	b := p // GetLongPathName says we can reuse buffer
	n, err := windows.GetLongPathName(&p[0], &b[0], uint32(len(b)))
	if err != nil {
		return "", err
	}
	if n > uint32(len(b)) {
		b = make([]uint16, n)
		_, err = windows.GetLongPathName(&p[0], &b[0], uint32(len(b)))
		if err != nil {
			return "", err
		}
	}
	return windows.UTF16ToString(b), nil
}

// checkSystemDriveAndRemoveDriveLetter is the Windows implementation
// of CheckSystemDriveAndRemoveDriveLetter
func checkSystemDriveAndRemoveDriveLetter(path string, driver PathVerifier) (string, error) {
	if len(path) == 2 && string(path[1]) == ":" {
		return "", fmt.Errorf("No relative path specified in %q", path)
	}
	if !driver.IsAbs(path) || len(path) < 2 {
		return filepath.FromSlash(path), nil
	}
	if string(path[1]) == ":" && !strings.EqualFold(string(path[0]), "c") {
		return "", fmt.Errorf("The specified path is not on the system drive (C:)")
	}
	return filepath.FromSlash(path[2:]), nil
}
