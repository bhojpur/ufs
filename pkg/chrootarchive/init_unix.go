//go:build !windows
// +build !windows

package chrootarchive

import (
	"fmt"
	"io"
	"os"

	"github.com/bhojpur/ufs/pkg/reexec"
)

func init() {
	reexec.Register("bhojpur-applyLayer", applyLayer)
	reexec.Register("bhojpur-untar", untar)
	reexec.Register("bhojpur-tar", tar)
}

func fatal(err error) {
	fmt.Fprint(os.Stderr, err)
	os.Exit(1)
}

// flush consumes all the bytes from the reader discarding
// any errors
func flush(r io.Reader) (bytes int64, err error) {
	return io.Copy(io.Discard, r)
}
