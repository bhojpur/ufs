//go:build !linux
// +build !linux

package archive

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
	"path/filepath"
	"runtime"
	"strings"

	common "github.com/bhojpur/host/pkg/common"
	statistics "github.com/bhojpur/host/pkg/statistics"
)

func collectFileInfoForChanges(oldDir, newDir string) (*FileInfo, *FileInfo, error) {
	var (
		oldRoot, newRoot *FileInfo
		err1, err2       error
		errs             = make(chan error, 2)
	)
	go func() {
		oldRoot, err1 = collectFileInfo(oldDir)
		errs <- err1
	}()
	go func() {
		newRoot, err2 = collectFileInfo(newDir)
		errs <- err2
	}()

	// block until both routines have returned
	for i := 0; i < 2; i++ {
		if err := <-errs; err != nil {
			return nil, nil, err
		}
	}

	return oldRoot, newRoot, nil
}

func collectFileInfo(sourceDir string) (*FileInfo, error) {
	root := newRootFileInfo()

	err := filepath.Walk(sourceDir, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Rebase path
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		// As this runs on the daemon side, file paths are OS specific.
		relPath = filepath.Join(string(os.PathSeparator), relPath)

		// Temporary workaround. If the returned path starts with two backslashes,
		// trim it down to a single backslash. Only relevant on Windows.
		if runtime.GOOS == "windows" {
			if strings.HasPrefix(relPath, `\\`) {
				relPath = relPath[1:]
			}
		}

		if relPath == string(os.PathSeparator) {
			return nil
		}

		parent := root.LookUp(filepath.Dir(relPath))
		if parent == nil {
			return fmt.Errorf("collectFileInfo: Unexpectedly no parent for %s", relPath)
		}

		info := &FileInfo{
			name:     filepath.Base(relPath),
			children: make(map[string]*FileInfo),
			parent:   parent,
		}

		s, err := statistics.Lstat(path)
		if err != nil {
			return err
		}
		info.stat = s

		info.capability, _ = common.Lgetxattr(path, "security.capability")

		parent.children[info.name] = info

		return nil
	})
	if err != nil {
		return nil, err
	}
	return root, nil
}
