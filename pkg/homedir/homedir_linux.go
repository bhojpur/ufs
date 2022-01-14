package homedir

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
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// GetRuntimeDir returns XDG_RUNTIME_DIR.
// XDG_RUNTIME_DIR is typically configured via pam_systemd.
// GetRuntimeDir returns non-nil error if XDG_RUNTIME_DIR is not set.
func GetRuntimeDir() (string, error) {
	if xdgRuntimeDir := os.Getenv("XDG_RUNTIME_DIR"); xdgRuntimeDir != "" {
		return xdgRuntimeDir, nil
	}
	return "", errors.New("could not get XDG_RUNTIME_DIR")
}

// StickRuntimeDirContents sets the sticky bit on files that are under
// XDG_RUNTIME_DIR, so that the files won't be periodically removed by the system.
//
// StickyRuntimeDir returns slice of sticked files.
// StickyRuntimeDir returns nil error if XDG_RUNTIME_DIR is not set.
func StickRuntimeDirContents(files []string) ([]string, error) {
	runtimeDir, err := GetRuntimeDir()
	if err != nil {
		// ignore error if runtimeDir is empty
		return nil, nil
	}
	runtimeDir, err = filepath.Abs(runtimeDir)
	if err != nil {
		return nil, err
	}
	var sticked []string
	for _, f := range files {
		f, err = filepath.Abs(f)
		if err != nil {
			return sticked, err
		}
		if strings.HasPrefix(f, runtimeDir+"/") {
			if err = stick(f); err != nil {
				return sticked, err
			}
			sticked = append(sticked, f)
		}
	}
	return sticked, nil
}

func stick(f string) error {
	st, err := os.Stat(f)
	if err != nil {
		return err
	}
	m := st.Mode()
	m |= os.ModeSticky
	return os.Chmod(f, m)
}

// GetDataHome returns XDG_DATA_HOME.
// GetDataHome returns $HOME/.local/share and nil error if XDG_DATA_HOME is not set.
func GetDataHome() (string, error) {
	if xdgDataHome := os.Getenv("XDG_DATA_HOME"); xdgDataHome != "" {
		return xdgDataHome, nil
	}
	home := os.Getenv("HOME")
	if home == "" {
		return "", errors.New("could not get either XDG_DATA_HOME or HOME")
	}
	return filepath.Join(home, ".local", "share"), nil
}

// GetConfigHome returns XDG_CONFIG_HOME.
// GetConfigHome returns $HOME/.config and nil error if XDG_CONFIG_HOME is not set.
func GetConfigHome() (string, error) {
	if xdgConfigHome := os.Getenv("XDG_CONFIG_HOME"); xdgConfigHome != "" {
		return xdgConfigHome, nil
	}
	home := os.Getenv("HOME")
	if home == "" {
		return "", errors.New("could not get either XDG_CONFIG_HOME or HOME")
	}
	return filepath.Join(home, ".config"), nil
}
