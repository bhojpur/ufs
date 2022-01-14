package fileutils

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
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/scanner"
	"unicode/utf8"
)

// escapeBytes is a bitmap used to check whether a character should be escaped when creating the regex.
var escapeBytes [8]byte

// shouldEscape reports whether a rune should be escaped as part of the regex.
//
// This only includes characters that require escaping in regex but are also NOT valid filepath pattern characters.
// Additionally, '\' is not excluded because there is specific logic to properly handle this, as it's a path separator
// on Windows.
//
// Adapted from regexp::QuoteMeta in go stdlib.
func shouldEscape(b rune) bool {
	return b < utf8.RuneSelf && escapeBytes[b%8]&(1<<(b/8)) != 0
}

func init() {
	for _, b := range []byte(`.+()|{}$`) {
		escapeBytes[b%8] |= 1 << (b / 8)
	}
}

// PatternMatcher allows checking paths against a list of patterns
type PatternMatcher struct {
	patterns   []*Pattern
	exclusions bool
}

// NewPatternMatcher creates a new matcher object for specific patterns that can
// be used later to match against patterns against paths
func NewPatternMatcher(patterns []string) (*PatternMatcher, error) {
	pm := &PatternMatcher{
		patterns: make([]*Pattern, 0, len(patterns)),
	}
	for _, p := range patterns {
		// Eliminate leading and trailing whitespace.
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		p = filepath.Clean(p)
		newp := &Pattern{}
		if p[0] == '!' {
			if len(p) == 1 {
				return nil, errors.New("illegal exclusion pattern: \"!\"")
			}
			newp.exclusion = true
			p = p[1:]
			pm.exclusions = true
		}
		// Do some syntax checking on the pattern.
		// filepath's Match() has some really weird rules that are inconsistent
		// so instead of trying to dup their logic, just call Match() for its
		// error state and if there is an error in the pattern return it.
		// If this becomes an issue we can remove this since its really only
		// needed in the error (syntax) case - which isn't really critical.
		if _, err := filepath.Match(p, "."); err != nil {
			return nil, err
		}
		newp.cleanedPattern = p
		newp.dirs = strings.Split(p, string(os.PathSeparator))
		pm.patterns = append(pm.patterns, newp)
	}
	return pm, nil
}

// Matches returns true if "file" matches any of the patterns
// and isn't excluded by any of the subsequent patterns.
//
// The "file" argument should be a slash-delimited path.
//
// Matches is not safe to call concurrently.
//
// Deprecated: This implementation is buggy (it only checks a single parent dir
// against the pattern) and will be removed soon. Use either
// MatchesOrParentMatches or MatchesUsingParentResults instead.
func (pm *PatternMatcher) Matches(file string) (bool, error) {
	matched := false
	file = filepath.FromSlash(file)
	parentPath := filepath.Dir(file)
	parentPathDirs := strings.Split(parentPath, string(os.PathSeparator))

	for _, pattern := range pm.patterns {
		// Skip evaluation if this is an inclusion and the filename
		// already matched the pattern, or it's an exclusion and it has
		// not matched the pattern yet.
		if pattern.exclusion != matched {
			continue
		}

		match, err := pattern.match(file)
		if err != nil {
			return false, err
		}

		if !match && parentPath != "." {
			// Check to see if the pattern matches one of our parent dirs.
			if len(pattern.dirs) <= len(parentPathDirs) {
				match, _ = pattern.match(strings.Join(parentPathDirs[:len(pattern.dirs)], string(os.PathSeparator)))
			}
		}

		if match {
			matched = !pattern.exclusion
		}
	}

	return matched, nil
}

// MatchesOrParentMatches returns true if "file" matches any of the patterns
// and isn't excluded by any of the subsequent patterns.
//
// The "file" argument should be a slash-delimited path.
//
// Matches is not safe to call concurrently.
func (pm *PatternMatcher) MatchesOrParentMatches(file string) (bool, error) {
	matched := false
	file = filepath.FromSlash(file)
	parentPath := filepath.Dir(file)
	parentPathDirs := strings.Split(parentPath, string(os.PathSeparator))

	for _, pattern := range pm.patterns {
		// Skip evaluation if this is an inclusion and the filename
		// already matched the pattern, or it's an exclusion and it has
		// not matched the pattern yet.
		if pattern.exclusion != matched {
			continue
		}

		match, err := pattern.match(file)
		if err != nil {
			return false, err
		}

		if !match && parentPath != "." {
			// Check to see if the pattern matches one of our parent dirs.
			for i := range parentPathDirs {
				match, _ = pattern.match(strings.Join(parentPathDirs[:i+1], string(os.PathSeparator)))
				if match {
					break
				}
			}
		}

		if match {
			matched = !pattern.exclusion
		}
	}

	return matched, nil
}

// MatchesUsingParentResult returns true if "file" matches any of the patterns
// and isn't excluded by any of the subsequent patterns. The functionality is
// the same as Matches, but as an optimization, the caller keeps track of
// whether the parent directory matched.
//
// The "file" argument should be a slash-delimited path.
//
// MatchesUsingParentResult is not safe to call concurrently.
//
// Use MatchesUsingParentResults instead.
func (pm *PatternMatcher) MatchesUsingParentResult(file string, parentMatched bool) (bool, error) {
	matched := parentMatched
	file = filepath.FromSlash(file)

	for _, pattern := range pm.patterns {
		// Skip evaluation if this is an inclusion and the filename
		// already matched the pattern, or it's an exclusion and it has
		// not matched the pattern yet.
		if pattern.exclusion != matched {
			continue
		}

		match, err := pattern.match(file)
		if err != nil {
			return false, err
		}

		if match {
			matched = !pattern.exclusion
		}
	}
	return matched, nil
}

// MatchInfo tracks information about parent dir matches while traversing a
// filesystem.
type MatchInfo struct {
	parentMatched []bool
}

// MatchesUsingParentResults returns true if "file" matches any of the patterns
// and isn't excluded by any of the subsequent patterns. The functionality is
// the same as Matches, but as an optimization, the caller passes in
// intermediate results from matching the parent directory.
//
// The "file" argument should be a slash-delimited path.
//
// MatchesUsingParentResults is not safe to call concurrently.
func (pm *PatternMatcher) MatchesUsingParentResults(file string, parentMatchInfo MatchInfo) (bool, MatchInfo, error) {
	parentMatched := parentMatchInfo.parentMatched
	if len(parentMatched) != 0 && len(parentMatched) != len(pm.patterns) {
		return false, MatchInfo{}, errors.New("wrong number of values in parentMatched")
	}

	file = filepath.FromSlash(file)
	matched := false

	matchInfo := MatchInfo{
		parentMatched: make([]bool, len(pm.patterns)),
	}
	for i, pattern := range pm.patterns {
		match := false
		// If the parent matched this pattern, we don't need to recheck.
		if len(parentMatched) != 0 {
			match = parentMatched[i]
		}

		if !match {
			// Skip evaluation if this is an inclusion and the filename
			// already matched the pattern, or it's an exclusion and it has
			// not matched the pattern yet.
			if pattern.exclusion != matched {
				continue
			}

			var err error
			match, err = pattern.match(file)
			if err != nil {
				return false, matchInfo, err
			}

			// If the zero value of MatchInfo was passed in, we don't have
			// any information about the parent dir's match results, and we
			// apply the same logic as MatchesOrParentMatches.
			if !match && len(parentMatched) == 0 {
				if parentPath := filepath.Dir(file); parentPath != "." {
					parentPathDirs := strings.Split(parentPath, string(os.PathSeparator))
					// Check to see if the pattern matches one of our parent dirs.
					for i := range parentPathDirs {
						match, _ = pattern.match(strings.Join(parentPathDirs[:i+1], string(os.PathSeparator)))
						if match {
							break
						}
					}
				}
			}
		}
		matchInfo.parentMatched[i] = match

		if match {
			matched = !pattern.exclusion
		}
	}
	return matched, matchInfo, nil
}

// Exclusions returns true if any of the patterns define exclusions
func (pm *PatternMatcher) Exclusions() bool {
	return pm.exclusions
}

// Patterns returns array of active patterns
func (pm *PatternMatcher) Patterns() []*Pattern {
	return pm.patterns
}

// Pattern defines a single regexp used to filter file paths.
type Pattern struct {
	cleanedPattern string
	dirs           []string
	regexp         *regexp.Regexp
	exclusion      bool
}

func (p *Pattern) String() string {
	return p.cleanedPattern
}

// Exclusion returns true if this pattern defines exclusion
func (p *Pattern) Exclusion() bool {
	return p.exclusion
}

func (p *Pattern) match(path string) (bool, error) {
	if p.regexp == nil {
		if err := p.compile(); err != nil {
			return false, filepath.ErrBadPattern
		}
	}

	b := p.regexp.MatchString(path)

	return b, nil
}

func (p *Pattern) compile() error {
	regStr := "^"
	pattern := p.cleanedPattern
	// Go through the pattern and convert it to a regexp.
	// We use a scanner so we can support utf-8 chars.
	var scan scanner.Scanner
	scan.Init(strings.NewReader(pattern))

	sl := string(os.PathSeparator)
	escSL := sl
	if sl == `\` {
		escSL += `\`
	}

	for scan.Peek() != scanner.EOF {
		ch := scan.Next()

		if ch == '*' {
			if scan.Peek() == '*' {
				// is some flavor of "**"
				scan.Next()

				// Treat **/ as ** so eat the "/"
				if string(scan.Peek()) == sl {
					scan.Next()
				}

				if scan.Peek() == scanner.EOF {
					// is "**EOF" - to align with .gitignore just accept all
					regStr += ".*"
				} else {
					// is "**"
					// Note that this allows for any # of /'s (even 0) because
					// the .* will eat everything, even /'s
					regStr += "(.*" + escSL + ")?"
				}
			} else {
				// is "*" so map it to anything but "/"
				regStr += "[^" + escSL + "]*"
			}
		} else if ch == '?' {
			// "?" is any char except "/"
			regStr += "[^" + escSL + "]"
		} else if shouldEscape(ch) {
			// Escape some regexp special chars that have no meaning
			// in golang's filepath.Match
			regStr += `\` + string(ch)
		} else if ch == '\\' {
			// escape next char. Note that a trailing \ in the pattern
			// will be left alone (but need to escape it)
			if sl == `\` {
				// On windows map "\" to "\\", meaning an escaped backslash,
				// and then just continue because filepath.Match on
				// Windows doesn't allow escaping at all
				regStr += escSL
				continue
			}
			if scan.Peek() != scanner.EOF {
				regStr += `\` + string(scan.Next())
			} else {
				regStr += `\`
			}
		} else {
			regStr += string(ch)
		}
	}

	regStr += "$"

	re, err := regexp.Compile(regStr)
	if err != nil {
		return err
	}

	p.regexp = re
	return nil
}

// Matches returns true if file matches any of the patterns
// and isn't excluded by any of the subsequent patterns.
//
// This implementation is buggy (it only checks a single parent dir against the
// pattern) and will be removed soon. Use MatchesOrParentMatches instead.
func Matches(file string, patterns []string) (bool, error) {
	pm, err := NewPatternMatcher(patterns)
	if err != nil {
		return false, err
	}
	file = filepath.Clean(file)

	if file == "." {
		// Don't let them exclude everything, kind of silly.
		return false, nil
	}

	return pm.Matches(file)
}

// MatchesOrParentMatches returns true if file matches any of the patterns
// and isn't excluded by any of the subsequent patterns.
func MatchesOrParentMatches(file string, patterns []string) (bool, error) {
	pm, err := NewPatternMatcher(patterns)
	if err != nil {
		return false, err
	}
	file = filepath.Clean(file)

	if file == "." {
		// Don't let them exclude everything, kind of silly.
		return false, nil
	}

	return pm.MatchesOrParentMatches(file)
}

// CopyFile copies from src to dst until either EOF is reached
// on src or an error occurs. It verifies src exists and removes
// the dst if it exists.
func CopyFile(src, dst string) (int64, error) {
	cleanSrc := filepath.Clean(src)
	cleanDst := filepath.Clean(dst)
	if cleanSrc == cleanDst {
		return 0, nil
	}
	sf, err := os.Open(cleanSrc)
	if err != nil {
		return 0, err
	}
	defer sf.Close()
	if err := os.Remove(cleanDst); err != nil && !os.IsNotExist(err) {
		return 0, err
	}
	df, err := os.Create(cleanDst)
	if err != nil {
		return 0, err
	}
	defer df.Close()
	return io.Copy(df, sf)
}

// ReadSymlinkedDirectory returns the target directory of a symlink.
// The target of the symbolic link may not be a file.
func ReadSymlinkedDirectory(path string) (string, error) {
	var realPath string
	var err error
	if realPath, err = filepath.Abs(path); err != nil {
		return "", fmt.Errorf("unable to get absolute path for %s: %s", path, err)
	}
	if realPath, err = filepath.EvalSymlinks(realPath); err != nil {
		return "", fmt.Errorf("failed to canonicalise path for %s: %s", path, err)
	}
	realPathInfo, err := os.Stat(realPath)
	if err != nil {
		return "", fmt.Errorf("failed to stat target '%s' of '%s': %s", realPath, path, err)
	}
	if !realPathInfo.Mode().IsDir() {
		return "", fmt.Errorf("canonical path points to a file '%s'", realPath)
	}
	return realPath, nil
}

// CreateIfNotExists creates a file or a directory only if it does not already exist.
func CreateIfNotExists(path string, isDir bool) error {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			if isDir {
				return os.MkdirAll(path, 0755)
			}
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				return err
			}
			f, err := os.OpenFile(path, os.O_CREATE, 0755)
			if err != nil {
				return err
			}
			f.Close()
		}
	}
	return nil
}
