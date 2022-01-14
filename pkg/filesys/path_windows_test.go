//go:build windows
// +build windows

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
	"testing"

	"github.com/containerd/continuity/pathdriver"
)

// TestCheckSystemDriveAndRemoveDriveLetter tests CheckSystemDriveAndRemoveDriveLetter
func TestCheckSystemDriveAndRemoveDriveLetter(t *testing.T) {
	// Fails if not C drive.
	_, err := CheckSystemDriveAndRemoveDriveLetter(`d:\`, pathdriver.LocalPathDriver)
	if err == nil || err.Error() != "The specified path is not on the system drive (C:)" {
		t.Fatalf("Expected error for d:")
	}

	// Single character is unchanged
	var path string
	if path, err = CheckSystemDriveAndRemoveDriveLetter("z", pathdriver.LocalPathDriver); err != nil {
		t.Fatalf("Single character should pass")
	}
	if path != "z" {
		t.Fatalf("Single character should be unchanged")
	}

	// Two characters without colon is unchanged
	if path, err = CheckSystemDriveAndRemoveDriveLetter("AB", pathdriver.LocalPathDriver); err != nil {
		t.Fatalf("2 characters without colon should pass")
	}
	if path != "AB" {
		t.Fatalf("2 characters without colon should be unchanged")
	}

	// Abs path without drive letter
	if path, err = CheckSystemDriveAndRemoveDriveLetter(`\l`, pathdriver.LocalPathDriver); err != nil {
		t.Fatalf("abs path no drive letter should pass")
	}
	if path != `\l` {
		t.Fatalf("abs path without drive letter should be unchanged")
	}

	// Abs path without drive letter, linux style
	if path, err = CheckSystemDriveAndRemoveDriveLetter(`/l`, pathdriver.LocalPathDriver); err != nil {
		t.Fatalf("abs path no drive letter linux style should pass")
	}
	if path != `\l` {
		t.Fatalf("abs path without drive letter linux failed %s", path)
	}

	// Drive-colon should be stripped
	if path, err = CheckSystemDriveAndRemoveDriveLetter(`c:\`, pathdriver.LocalPathDriver); err != nil {
		t.Fatalf("An absolute path should pass")
	}
	if path != `\` {
		t.Fatalf(`An absolute path should have been shortened to \ %s`, path)
	}

	// Verify with a linux-style path
	if path, err = CheckSystemDriveAndRemoveDriveLetter(`c:/`, pathdriver.LocalPathDriver); err != nil {
		t.Fatalf("An absolute path should pass")
	}
	if path != `\` {
		t.Fatalf(`A linux style absolute path should have been shortened to \ %s`, path)
	}

	// Failure on c:
	if path, err = CheckSystemDriveAndRemoveDriveLetter(`c:`, pathdriver.LocalPathDriver); err == nil {
		t.Fatalf("c: should fail")
	}
	if err.Error() != `No relative path specified in "c:"` {
		t.Fatalf(path, err)
	}

	// Failure on d:
	if path, err = CheckSystemDriveAndRemoveDriveLetter(`d:`, pathdriver.LocalPathDriver); err == nil {
		t.Fatalf("c: should fail")
	}
	if err.Error() != `No relative path specified in "d:"` {
		t.Fatalf(path, err)
	}
}
