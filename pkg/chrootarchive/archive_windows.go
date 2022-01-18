package chrootarchive

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
	"io"

	longpath "github.com/bhojpur/cache/pkg/filepath"
	"github.com/bhojpur/ufs/pkg/archive"
)

// chroot is not supported by Windows
func chroot(path string) error {
	return nil
}

func invokeUnpack(decompressedArchive io.ReadCloser,
	dest string,
	options *archive.TarOptions, root string) error {
	// Windows is different to Linux here because Windows does not support
	// chroot. Hence there is no point sandboxing a chrooted process to
	// do the unpack. We call inline instead within the daemon process.
	return archive.Unpack(decompressedArchive, longpath.AddPrefix(dest), options)
}

func invokePack(srcPath string, options *archive.TarOptions, root string) (io.ReadCloser, error) {
	// Windows is different to Linux here because Windows does not support
	// chroot. Hence there is no point sandboxing a chrooted process to
	// do the pack. We call inline instead within the daemon process.
	return archive.TarWithOptions(srcPath, options)
}
