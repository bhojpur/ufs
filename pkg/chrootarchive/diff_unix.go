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

	procsys "github.com/bhojpur/host/pkg/process"
	"github.com/bhojpur/ufs/pkg/archive"
	"github.com/bhojpur/ufs/pkg/reexec"
	"github.com/containerd/containerd/pkg/userns"
)

type applyLayerResponse struct {
	LayerSize int64 `json:"layerSize"`
}

// applyLayer is the entry-point for bhojpur-applylayer on re-exec. This is not
// used on Windows as it does not support chroot, hence no point sandboxing
// through chroot and rexec.
func applyLayer() {

	var (
		tmpDir  string
		err     error
		options *archive.TarOptions
	)
	runtime.LockOSThread()
	flag.Parse()

	inUserns := userns.RunningInUserNS()
	if err := chroot(flag.Arg(0)); err != nil {
		fatal(err)
	}

	// We need to be able to set any perms
	oldmask, err := procsys.Umask(0)
	defer procsys.Umask(oldmask)
	if err != nil {
		fatal(err)
	}

	if err := json.Unmarshal([]byte(os.Getenv("OPT")), &options); err != nil {
		fatal(err)
	}

	if inUserns {
		options.InUserNS = true
	}

	if tmpDir, err = os.MkdirTemp("/", "temp-bhojpur-extract"); err != nil {
		fatal(err)
	}

	os.Setenv("TMPDIR", tmpDir)
	size, err := archive.UnpackLayer("/", os.Stdin, options)
	os.RemoveAll(tmpDir)
	if err != nil {
		fatal(err)
	}

	encoder := json.NewEncoder(os.Stdout)
	if err := encoder.Encode(applyLayerResponse{size}); err != nil {
		fatal(fmt.Errorf("unable to encode layerSize JSON: %s", err))
	}

	if _, err := flush(os.Stdin); err != nil {
		fatal(err)
	}

	os.Exit(0)
}

// applyLayerHandler parses a diff in the standard layer format from `layer`, and
// applies it to the directory `dest`. Returns the size in bytes of the
// contents of the layer.
func applyLayerHandler(dest string, layer io.Reader, options *archive.TarOptions, decompress bool) (size int64, err error) {
	dest = filepath.Clean(dest)
	if decompress {
		decompressed, err := archive.DecompressStream(layer)
		if err != nil {
			return 0, err
		}
		defer decompressed.Close()

		layer = decompressed
	}
	if options == nil {
		options = &archive.TarOptions{}
		if userns.RunningInUserNS() {
			options.InUserNS = true
		}
	}
	if options.ExcludePatterns == nil {
		options.ExcludePatterns = []string{}
	}

	data, err := json.Marshal(options)
	if err != nil {
		return 0, fmt.Errorf("ApplyLayer json encode: %v", err)
	}

	cmd := reexec.Command("bhojpur-applyLayer", dest)
	cmd.Stdin = layer
	cmd.Env = append(cmd.Env, fmt.Sprintf("OPT=%s", data))

	outBuf, errBuf := new(bytes.Buffer), new(bytes.Buffer)
	cmd.Stdout, cmd.Stderr = outBuf, errBuf

	if err = cmd.Run(); err != nil {
		return 0, fmt.Errorf("ApplyLayer %s stdout: %s stderr: %s", err, outBuf, errBuf)
	}

	// Stdout should be a valid JSON struct representing an applyLayerResponse.
	response := applyLayerResponse{}
	decoder := json.NewDecoder(outBuf)
	if err = decoder.Decode(&response); err != nil {
		return 0, fmt.Errorf("unable to decode ApplyLayer JSON response: %s", err)
	}

	return response.LayerSize, nil
}
