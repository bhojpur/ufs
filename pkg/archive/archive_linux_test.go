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
	"os"
	"path/filepath"
	"syscall"
	"testing"

	common "github.com/bhojpur/host/pkg/common"
	filesys "github.com/bhojpur/host/pkg/filesys"
	procsys "github.com/bhojpur/host/pkg/process"
	"github.com/containerd/containerd/pkg/userns"
	"golang.org/x/sys/unix"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/skip"
)

// setupOverlayTestDir creates files in a directory with overlay whiteouts
// Tree layout
// .
// ├── d1     # opaque, 0700
// │   └── f1 # empty file, 0600
// ├── d2     # opaque, 0750
// │   └── f1 # empty file, 0660
// └── d3     # 0700
//     └── f1 # whiteout, 0644
func setupOverlayTestDir(t *testing.T, src string) {
	skip.If(t, os.Getuid() != 0, "skipping test that requires root")
	skip.If(t, userns.RunningInUserNS(), "skipping test that requires initial userns (trusted.overlay.opaque xattr cannot be set in userns, even with Ubuntu kernel)")
	// Create opaque directory containing single file and permission 0700
	err := os.Mkdir(filepath.Join(src, "d1"), 0700)
	assert.NilError(t, err)

	err = common.Lsetxattr(filepath.Join(src, "d1"), "trusted.overlay.opaque", []byte("y"), 0)
	assert.NilError(t, err)

	err = os.WriteFile(filepath.Join(src, "d1", "f1"), []byte{}, 0600)
	assert.NilError(t, err)

	// Create another opaque directory containing single file but with permission 0750
	err = os.Mkdir(filepath.Join(src, "d2"), 0750)
	assert.NilError(t, err)

	err = common.Lsetxattr(filepath.Join(src, "d2"), "trusted.overlay.opaque", []byte("y"), 0)
	assert.NilError(t, err)

	err = os.WriteFile(filepath.Join(src, "d2", "f1"), []byte{}, 0660)
	assert.NilError(t, err)

	// Create regular directory with deleted file
	err = os.Mkdir(filepath.Join(src, "d3"), 0700)
	assert.NilError(t, err)

	err = filesys.Mknod(filepath.Join(src, "d3", "f1"), unix.S_IFCHR, 0)
	assert.NilError(t, err)
}

func checkOpaqueness(t *testing.T, path string, opaque string) {
	xattrOpaque, err := common.Lgetxattr(path, "trusted.overlay.opaque")
	assert.NilError(t, err)

	if string(xattrOpaque) != opaque {
		t.Fatalf("Unexpected opaque value: %q, expected %q", string(xattrOpaque), opaque)
	}

}

func checkOverlayWhiteout(t *testing.T, path string) {
	stat, err := os.Stat(path)
	assert.NilError(t, err)

	statT, ok := stat.Sys().(*syscall.Stat_t)
	if !ok {
		t.Fatalf("Unexpected type: %t, expected *syscall.Stat_t", stat.Sys())
	}
	if statT.Rdev != 0 {
		t.Fatalf("Non-zero device number for whiteout")
	}
}

func checkFileMode(t *testing.T, path string, perm os.FileMode) {
	stat, err := os.Stat(path)
	assert.NilError(t, err)

	if stat.Mode() != perm {
		t.Fatalf("Unexpected file mode for %s: %o, expected %o", path, stat.Mode(), perm)
	}
}

func TestOverlayTarUntar(t *testing.T) {
	oldmask, err := procsys.Umask(0)
	assert.NilError(t, err)
	defer procsys.Umask(oldmask)

	src, err := os.MkdirTemp("", "bhojpur-test-overlay-tar-src")
	assert.NilError(t, err)
	defer os.RemoveAll(src)

	setupOverlayTestDir(t, src)

	dst, err := os.MkdirTemp("", "bhojpur-test-overlay-tar-dst")
	assert.NilError(t, err)
	defer os.RemoveAll(dst)

	options := &TarOptions{
		Compression:    Uncompressed,
		WhiteoutFormat: OverlayWhiteoutFormat,
	}
	archive, err := TarWithOptions(src, options)
	assert.NilError(t, err)
	defer archive.Close()

	err = Untar(archive, dst, options)
	assert.NilError(t, err)

	checkFileMode(t, filepath.Join(dst, "d1"), 0700|os.ModeDir)
	checkFileMode(t, filepath.Join(dst, "d2"), 0750|os.ModeDir)
	checkFileMode(t, filepath.Join(dst, "d3"), 0700|os.ModeDir)
	checkFileMode(t, filepath.Join(dst, "d1", "f1"), 0600)
	checkFileMode(t, filepath.Join(dst, "d2", "f1"), 0660)
	checkFileMode(t, filepath.Join(dst, "d3", "f1"), os.ModeCharDevice|os.ModeDevice)

	checkOpaqueness(t, filepath.Join(dst, "d1"), "y")
	checkOpaqueness(t, filepath.Join(dst, "d2"), "y")
	checkOpaqueness(t, filepath.Join(dst, "d3"), "")
	checkOverlayWhiteout(t, filepath.Join(dst, "d3", "f1"))
}

func TestOverlayTarAUFSUntar(t *testing.T) {
	oldmask, err := procsys.Umask(0)
	assert.NilError(t, err)
	defer procsys.Umask(oldmask)

	src, err := os.MkdirTemp("", "bhojpur-test-overlay-tar-src")
	assert.NilError(t, err)
	defer os.RemoveAll(src)

	setupOverlayTestDir(t, src)

	dst, err := os.MkdirTemp("", "bhojpur-test-overlay-tar-dst")
	assert.NilError(t, err)
	defer os.RemoveAll(dst)

	archive, err := TarWithOptions(src, &TarOptions{
		Compression:    Uncompressed,
		WhiteoutFormat: OverlayWhiteoutFormat,
	})
	assert.NilError(t, err)
	defer archive.Close()

	err = Untar(archive, dst, &TarOptions{
		Compression:    Uncompressed,
		WhiteoutFormat: AUFSWhiteoutFormat,
	})
	assert.NilError(t, err)

	checkFileMode(t, filepath.Join(dst, "d1"), 0700|os.ModeDir)
	checkFileMode(t, filepath.Join(dst, "d1", WhiteoutOpaqueDir), 0700)
	checkFileMode(t, filepath.Join(dst, "d2"), 0750|os.ModeDir)
	checkFileMode(t, filepath.Join(dst, "d2", WhiteoutOpaqueDir), 0750)
	checkFileMode(t, filepath.Join(dst, "d3"), 0700|os.ModeDir)
	checkFileMode(t, filepath.Join(dst, "d1", "f1"), 0600)
	checkFileMode(t, filepath.Join(dst, "d2", "f1"), 0660)
	checkFileMode(t, filepath.Join(dst, "d3", WhiteoutPrefix+"f1"), 0600)
}
