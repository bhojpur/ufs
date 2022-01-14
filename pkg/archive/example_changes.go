//go:build ignore
// +build ignore

package main

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

// Simple tool to create an archive stream from an old and new directory
//
// By default it will stream the comparison of two temporary directories with junk files
import (
	"flag"
	"fmt"
	"io"
	"os"
	"path"

	archive "github.com/bhojpur/ufs/pkg/archive"
	"github.com/sirupsen/logrus"
)

var (
	flDebug  = flag.Bool("D", false, "debugging output")
	flNewDir = flag.String("newdir", "", "")
	flOldDir = flag.String("olddir", "", "")
	log      = logrus.New()
)

func main() {
	flag.Usage = func() {
		fmt.Println("Produce a tar from comparing two directory paths. By default a demo tar is created of around 200 files (including hardlinks)")
		fmt.Printf("%s [OPTIONS]\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	log.Out = os.Stderr
	if (len(os.Getenv("DEBUG")) > 0) || *flDebug {
		logrus.SetLevel(logrus.DebugLevel)
	}
	var newDir, oldDir string

	if len(*flNewDir) == 0 {
		var err error
		newDir, err = os.MkdirTemp("", "bhojpur-test-newDir")
		if err != nil {
			log.Fatal(err)
		}
		defer os.RemoveAll(newDir)
		if _, err := prepareUntarSourceDirectory(100, newDir, true); err != nil {
			log.Fatal(err)
		}
	} else {
		newDir = *flNewDir
	}

	if len(*flOldDir) == 0 {
		oldDir, err := os.MkdirTemp("", "bhojpur-test-oldDir")
		if err != nil {
			log.Fatal(err)
		}
		defer os.RemoveAll(oldDir)
	} else {
		oldDir = *flOldDir
	}

	changes, err := archive.ChangesDirs(newDir, oldDir)
	if err != nil {
		log.Fatal(err)
	}

	a, err := archive.ExportChanges(newDir, changes)
	if err != nil {
		log.Fatal(err)
	}
	defer a.Close()

	i, err := io.Copy(os.Stdout, a)
	if err != nil && err != io.EOF {
		log.Fatal(err)
	}
	fmt.Fprintf(os.Stderr, "wrote archive of %d bytes", i)
}

func prepareUntarSourceDirectory(numberOfFiles int, targetPath string, makeLinks bool) (int, error) {
	fileData := []byte("fooo")
	for n := 0; n < numberOfFiles; n++ {
		fileName := fmt.Sprintf("file-%d", n)
		if err := os.WriteFile(path.Join(targetPath, fileName), fileData, 0700); err != nil {
			return 0, err
		}
		if makeLinks {
			if err := os.Link(path.Join(targetPath, fileName), path.Join(targetPath, fileName+"-link")); err != nil {
				return 0, err
			}
		}
	}
	totalSize := numberOfFiles * len(fileData)
	return totalSize, nil
}
