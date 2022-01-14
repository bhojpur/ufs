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

	"github.com/bhojpur/ufs/pkg/archive"
)

// ApplyLayer parses a diff in the standard layer format from `layer`,
// and applies it to the directory `dest`. The stream `layer` can only be
// uncompressed.
// Returns the size in bytes of the contents of the layer.
func ApplyLayer(dest string, layer io.Reader) (size int64, err error) {
	return applyLayerHandler(dest, layer, &archive.TarOptions{}, true)
}

// ApplyUncompressedLayer parses a diff in the standard layer format from
// `layer`, and applies it to the directory `dest`. The stream `layer`
// can only be uncompressed.
// Returns the size in bytes of the contents of the layer.
func ApplyUncompressedLayer(dest string, layer io.Reader, options *archive.TarOptions) (int64, error) {
	return applyLayerHandler(dest, layer, options, false)
}
