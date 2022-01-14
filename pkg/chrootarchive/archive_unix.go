//go:build !windows
// +build !windows

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
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/bhojpur/ufs/pkg/archive"
	"github.com/bhojpur/ufs/pkg/reexec"
	"github.com/pkg/errors"
)

// untar is the entry-point for bhojpur-untar on re-exec. This is not used on
// Windows as it does not support chroot, hence no point sandboxing through
// chroot and rexec.
func untar() {
	runtime.LockOSThread()
	flag.Parse()

	var options archive.TarOptions

	// read the options from the pipe "ExtraFiles"
	if err := json.NewDecoder(os.NewFile(3, "options")).Decode(&options); err != nil {
		fatal(err)
	}

	dst := flag.Arg(0)
	var root string
	if len(flag.Args()) > 1 {
		root = flag.Arg(1)
	}

	if root == "" {
		root = dst
	}

	if err := chroot(root); err != nil {
		fatal(err)
	}

	if err := archive.Unpack(os.Stdin, dst, &options); err != nil {
		fatal(err)
	}
	// fully consume stdin in case it is zero padded
	if _, err := flush(os.Stdin); err != nil {
		fatal(err)
	}

	os.Exit(0)
}

func invokeUnpack(decompressedArchive io.Reader, dest string, options *archive.TarOptions, root string) error {
	if root == "" {
		return errors.New("must specify a root to chroot to")
	}

	// We can't pass a potentially large exclude list directly via cmd line
	// because we easily overrun the kernel's max argument/environment size
	// when the full image list is passed (e.g. when this is used by
	// `bhojpur load`). We will marshall the options via a pipe to the
	// child
	r, w, err := os.Pipe()
	if err != nil {
		return fmt.Errorf("Untar pipe failure: %v", err)
	}

	if root != "" {
		relDest, err := filepath.Rel(root, dest)
		if err != nil {
			return err
		}
		if relDest == "." {
			relDest = "/"
		}
		if relDest[0] != '/' {
			relDest = "/" + relDest
		}
		dest = relDest
	}

	cmd := reexec.Command("bhojpur-untar", dest, root)
	cmd.Stdin = decompressedArchive

	cmd.ExtraFiles = append(cmd.ExtraFiles, r)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	cmd.Stderr = output

	if err := cmd.Start(); err != nil {
		w.Close()
		return fmt.Errorf("Untar error on re-exec cmd: %v", err)
	}

	// write the options to the pipe for the untar exec to read
	if err := json.NewEncoder(w).Encode(options); err != nil {
		w.Close()
		return fmt.Errorf("Untar json encode to pipe failed: %v", err)
	}
	w.Close()

	if err := cmd.Wait(); err != nil {
		// when `xz -d -c -q | bhojpur-untar ...` failed on bhojpur-untar side,
		// we need to exhaust `xz`'s output, otherwise the `xz` side will be
		// pending on write pipe forever
		io.Copy(io.Discard, decompressedArchive)

		return fmt.Errorf("Error processing tar file(%v): %s", err, output)
	}
	return nil
}

func tar() {
	runtime.LockOSThread()
	flag.Parse()

	src := flag.Arg(0)
	var root string
	if len(flag.Args()) > 1 {
		root = flag.Arg(1)
	}

	if root == "" {
		root = src
	}

	if err := realChroot(root); err != nil {
		fatal(err)
	}

	var options archive.TarOptions
	if err := json.NewDecoder(os.Stdin).Decode(&options); err != nil {
		fatal(err)
	}

	rdr, err := archive.TarWithOptions(src, &options)
	if err != nil {
		fatal(err)
	}
	defer rdr.Close()

	if _, err := io.Copy(os.Stdout, rdr); err != nil {
		fatal(err)
	}

	os.Exit(0)
}

func invokePack(srcPath string, options *archive.TarOptions, root string) (io.ReadCloser, error) {
	if root == "" {
		return nil, errors.New("root path must not be empty")
	}

	relSrc, err := filepath.Rel(root, srcPath)
	if err != nil {
		return nil, err
	}
	if relSrc == "." {
		relSrc = "/"
	}
	if relSrc[0] != '/' {
		relSrc = "/" + relSrc
	}

	// make sure we didn't trim a trailing slash with the call to `Rel`
	if strings.HasSuffix(srcPath, "/") && !strings.HasSuffix(relSrc, "/") {
		relSrc += "/"
	}

	cmd := reexec.Command("bhojpur-tar", relSrc, root)

	errBuff := bytes.NewBuffer(nil)
	cmd.Stderr = errBuff

	tarR, tarW := io.Pipe()
	cmd.Stdout = tarW

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, errors.Wrap(err, "error getting options pipe for tar process")
	}

	if err := cmd.Start(); err != nil {
		return nil, errors.Wrap(err, "tar error on re-exec cmd")
	}

	go func() {
		err := cmd.Wait()
		err = errors.Wrapf(err, "error processing tar file: %s", errBuff)
		tarW.CloseWithError(err)
	}()

	if err := json.NewEncoder(stdin).Encode(options); err != nil {
		stdin.Close()
		return nil, errors.Wrap(err, "tar json encode to pipe failed")
	}
	stdin.Close()

	return tarR, nil
}
