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
	"archive/tar"
	"bytes"
	"io"
)

// Generate generates a new archive from the content provided as input.
//
// `files` is a sequence of path/content pairs. A new file is added to the
// archive for each pair.
// If the last pair is incomplete, the file is created with an empty content.
// For example:
//
// Generate("foo.txt", "hello world", "emptyfile")
//
// The above call will return an archive with 2 files:
//  * ./foo.txt with content "hello world"
//  * ./empty with empty content
//
// FIXME: stream content instead of buffering
// FIXME: specify permissions and other archive metadata
func Generate(input ...string) (io.Reader, error) {
	files := parseStringPairs(input...)
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	for _, file := range files {
		name, content := file[0], file[1]
		hdr := &tar.Header{
			Name: name,
			Size: int64(len(content)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return nil, err
		}
		if _, err := tw.Write([]byte(content)); err != nil {
			return nil, err
		}
	}
	if err := tw.Close(); err != nil {
		return nil, err
	}
	return buf, nil
}

func parseStringPairs(input ...string) (output [][2]string) {
	output = make([][2]string, 0, len(input)/2+1)
	for i := 0; i < len(input); i += 2 {
		var pair [2]string
		pair[0] = input[i]
		if i+1 < len(input) {
			pair[1] = input[i+1]
		}
		output = append(output, pair)
	}
	return
}
