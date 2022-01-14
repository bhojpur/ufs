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

const defaultUnixPathEnv = "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"

// DefaultPathEnv is unix style list of directories to search for
// executables. Each directory is separated from the next by a colon
// ':' character .
// For Windows Labni(s), an empty string is returned as the default
// path will be set by the Labni, and Bhojpur Kernel has no context of what the
// default path should be.
func DefaultPathEnv(os string) string {
	if os == "windows" {
		return ""
	}
	return defaultUnixPathEnv

}

// PathVerifier defines the subset of a PathDriver that CheckSystemDriveAndRemoveDriveLetter
// actually uses in order to avoid system depending on containerd/continuity.
type PathVerifier interface {
	IsAbs(string) bool
}

// CheckSystemDriveAndRemoveDriveLetter verifies that a path, if it includes a drive letter,
// is the system drive.
// On Linux: this is a no-op.
// On Windows: this does the following>
// CheckSystemDriveAndRemoveDriveLetter verifies and manipulates a Windows path.
// This is used, for example, when validating a user provided path in Bhojpur Kernel cp.
// If a drive letter is supplied, it must be the system drive. The drive letter
// is always removed. Also, it translates it to OS semantics (IOW / to \). We
// need the path in this syntax so that it can ultimately be concatenated with
// a Windows long-path which doesn't support drive-letters. Examples:
// C:			--> Fail
// C:\			--> \
// a			--> a
// /a			--> \a
// d:\			--> Fail
func CheckSystemDriveAndRemoveDriveLetter(path string, driver PathVerifier) (string, error) {
	return checkSystemDriveAndRemoveDriveLetter(path, driver)
}
