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
	"os"
	"path/filepath"

	"github.com/bhojpur/drive/pkg/mount"
	"github.com/bhojpur/drive/pkg/mountinfo"
	"github.com/containerd/containerd/pkg/userns"
	"golang.org/x/sys/unix"
)

// chroot on linux uses pivot_root instead of chroot
// pivot_root takes a new root and an old root.
// Old root must be a sub-dir of new root, it is where the current rootfs will reside after the call to pivot_root.
// New root is where the new rootfs is set to.
// Old root is removed after the call to pivot_root so it is no longer available under the new root.
// This is similar to how libcontainer sets up a container's rootfs
func chroot(path string) (err error) {
	// if the engine is running in a user namespace we need to use actual chroot
	if userns.RunningInUserNS() {
		return realChroot(path)
	}
	if err := unix.Unshare(unix.CLONE_NEWNS); err != nil {
		return fmt.Errorf("Error creating mount namespace before pivot: %v", err)
	}

	// Make everything in new ns slave.
	// Don't use `private` here as this could race where the mountns gets a
	//   reference to a mount and an unmount from the host does not propagate,
	//   which could potentially cause transient errors for other operations,
	//   even though this should be relatively small window here `slave` should
	//   not cause any problems.
	if err := mount.MakeRSlave("/"); err != nil {
		return err
	}

	if mounted, _ := mountinfo.Mounted(path); !mounted {
		if err := mount.Mount(path, path, "bind", "rbind,rw"); err != nil {
			return realChroot(path)
		}
	}

	// setup oldRoot for pivot_root
	pivotDir, err := os.MkdirTemp(path, ".pivot_root")
	if err != nil {
		return fmt.Errorf("Error setting up pivot dir: %v", err)
	}

	var mounted bool
	defer func() {
		if mounted {
			// make sure pivotDir is not mounted before we try to remove it
			if errCleanup := unix.Unmount(pivotDir, unix.MNT_DETACH); errCleanup != nil {
				if err == nil {
					err = errCleanup
				}
				return
			}
		}

		errCleanup := os.Remove(pivotDir)
		// pivotDir doesn't exist if pivot_root failed and chroot+chdir was successful
		// because we already cleaned it up on failed pivot_root
		if errCleanup != nil && !os.IsNotExist(errCleanup) {
			errCleanup = fmt.Errorf("Error cleaning up after pivot: %v", errCleanup)
			if err == nil {
				err = errCleanup
			}
		}
	}()

	if err := unix.PivotRoot(path, pivotDir); err != nil {
		// If pivot fails, fall back to the normal chroot after cleaning up temp dir
		if err := os.Remove(pivotDir); err != nil {
			return fmt.Errorf("Error cleaning up after failed pivot: %v", err)
		}
		return realChroot(path)
	}
	mounted = true

	// This is the new path for where the old root (prior to the pivot) has been moved to
	// This dir contains the rootfs of the caller, which we need to remove so it is not visible during extraction
	pivotDir = filepath.Join("/", filepath.Base(pivotDir))

	if err := unix.Chdir("/"); err != nil {
		return fmt.Errorf("Error changing to new root: %v", err)
	}

	// Make the pivotDir (where the old root lives) private so it can be unmounted without propagating to the host
	if err := unix.Mount("", pivotDir, "", unix.MS_PRIVATE|unix.MS_REC, ""); err != nil {
		return fmt.Errorf("Error making old root private after pivot: %v", err)
	}

	// Now unmount the old root so it's no longer visible from the new root
	if err := unix.Unmount(pivotDir, unix.MNT_DETACH); err != nil {
		return fmt.Errorf("Error while unmounting old root after pivot: %v", err)
	}
	mounted = false

	return nil
}

func realChroot(path string) error {
	if err := unix.Chroot(path); err != nil {
		return fmt.Errorf("Error after fallback to chroot: %v", err)
	}
	if err := unix.Chdir("/"); err != nil {
		return fmt.Errorf("Error changing to new root after chroot: %v", err)
	}
	return nil
}
