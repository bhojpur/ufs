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
	"testing"

	"gotest.tools/v3/assert"
)

func TestGenerateEmptyFile(t *testing.T) {
	archive, err := Generate("emptyFile")
	assert.NilError(t, err)
	if archive == nil {
		t.Fatal("The generated archive should not be nil.")
	}

	expectedFiles := [][]string{
		{"emptyFile", ""},
	}

	tr := tar.NewReader(archive)
	actualFiles := make([][]string, 0, 10)
	i := 0
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		assert.NilError(t, err)
		buf := new(bytes.Buffer)
		buf.ReadFrom(tr)
		content := buf.String()
		actualFiles = append(actualFiles, []string{hdr.Name, content})
		i++
	}
	if len(actualFiles) != len(expectedFiles) {
		t.Fatalf("Number of expected file %d, got %d.", len(expectedFiles), len(actualFiles))
	}
	for i := 0; i < len(expectedFiles); i++ {
		actual := actualFiles[i]
		expected := expectedFiles[i]
		if actual[0] != expected[0] {
			t.Fatalf("Expected name '%s', Actual name '%s'", expected[0], actual[0])
		}
		if actual[1] != expected[1] {
			t.Fatalf("Expected content '%s', Actual content '%s'", expected[1], actual[1])
		}
	}
}

func TestGenerateWithContent(t *testing.T) {
	archive, err := Generate("file", "content")
	assert.NilError(t, err)
	if archive == nil {
		t.Fatal("The generated archive should not be nil.")
	}

	expectedFiles := [][]string{
		{"file", "content"},
	}

	tr := tar.NewReader(archive)
	actualFiles := make([][]string, 0, 10)
	i := 0
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		assert.NilError(t, err)
		buf := new(bytes.Buffer)
		buf.ReadFrom(tr)
		content := buf.String()
		actualFiles = append(actualFiles, []string{hdr.Name, content})
		i++
	}
	if len(actualFiles) != len(expectedFiles) {
		t.Fatalf("Number of expected file %d, got %d.", len(expectedFiles), len(actualFiles))
	}
	for i := 0; i < len(expectedFiles); i++ {
		actual := actualFiles[i]
		expected := expectedFiles[i]
		if actual[0] != expected[0] {
			t.Fatalf("Expected name '%s', Actual name '%s'", expected[0], actual[0])
		}
		if actual[1] != expected[1] {
			t.Fatalf("Expected content '%s', Actual content '%s'", expected[1], actual[1])
		}
	}
}
