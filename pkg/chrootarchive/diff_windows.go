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
	"fmt"
	"io"
	"os"
	"path/filepath"

	longpath "github.com/bhojpur/cache/pkg/filepath"
	"github.com/bhojpur/ufs/pkg/archive"
)

// applyLayerHandler parses a diff in the standard layer format from `layer`, and
// applies it to the directory `dest`. Returns the size in bytes of the
// contents of the layer.
func applyLayerHandler(dest string, layer io.Reader, options *archive.TarOptions, decompress bool) (size int64, err error) {
	dest = filepath.Clean(dest)

	// Ensure it is a Windows-style volume path
	dest = longpath.AddPrefix(dest)

	if decompress {
		decompressed, err := archive.DecompressStream(layer)
		if err != nil {
			return 0, err
		}
		defer decompressed.Close()

		layer = decompressed
	}

	tmpDir, err := os.MkdirTemp(os.Getenv("temp"), "temp-bhojpur-extract")
	if err != nil {
		return 0, fmt.Errorf("ApplyLayer failed to create temp-bhojpur-extract under %s. %s", dest, err)
	}

	s, err := archive.UnpackLayer(dest, layer, nil)
	os.RemoveAll(tmpDir)
	if err != nil {
		return 0, fmt.Errorf("ApplyLayer %s failed UnpackLayer to %s: %s", layer, dest, err)
	}

	return s, nil
}
