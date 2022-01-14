package labanifs

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
	"path/filepath"
	"runtime"

	"github.com/bhojpur/drive/pkg/symlink"
	"github.com/containerd/continuity/driver"
	"github.com/containerd/continuity/pathdriver"
)

// LabaniFS is that represents a root file system
type LabaniFS interface {
	// Path returns the path to the root. Note that this may not exist
	// on the local system, so the continuity operations must be used
	Path() string

	// ResolveScopedPath evaluates the given path scoped to the root.
	// For example, if root=/a, and path=/b/c, then this function would return /a/b/c.
	// If rawPath is true, then the function will not preform any modifications
	// before path resolution. Otherwise, the function will clean the given path
	// by making it an absolute path.
	ResolveScopedPath(path string, rawPath bool) (string, error)

	Driver
}

// Driver combines both continuity's Driver and PathDriver interfaces with a Platform
// field to determine the OS.
type Driver interface {
	// OS returns the OS where the rootfs is located. Essentially, runtime.GOOS.
	OS() string

	// Architecture returns the hardware architecture where the
	// container is located.
	Architecture() string

	// Driver & PathDriver provide methods to manipulate files & paths
	driver.Driver
	pathdriver.PathDriver
}

// NewLocalLabaniFS is a helper function to implement daemon's Mount interface
// when the graphdriver mount point is a local path on the machine.
func NewLocalLabaniFS(path string) LabaniFS {
	return &local{
		path:       path,
		Driver:     driver.LocalDriver,
		PathDriver: pathdriver.LocalPathDriver,
	}
}

// NewLocalDriver provides file and path drivers for a local file system. They are
// essentially a wrapper around the `os` and `filepath` functions.
func NewLocalDriver() Driver {
	return &local{
		Driver:     driver.LocalDriver,
		PathDriver: pathdriver.LocalPathDriver,
	}
}

type local struct {
	path string
	driver.Driver
	pathdriver.PathDriver
}

func (l *local) Path() string {
	return l.path
}

func (l *local) ResolveScopedPath(path string, rawPath bool) (string, error) {
	cleanedPath := path
	if !rawPath {
		cleanedPath = cleanScopedPath(path)
	}
	return symlink.FollowSymlinkInScope(filepath.Join(l.path, cleanedPath), l.path)
}

func (l *local) OS() string {
	return runtime.GOOS
}

func (l *local) Architecture() string {
	return runtime.GOARCH
}
