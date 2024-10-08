// Copyright 2024 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bufio"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"unicode"
)

var wantFile = filepath.Join("genai", "license.txt")

func TestLicense(t *testing.T) {
	lic, err := os.ReadFile(wantFile)
	if err != nil {
		t.Fatal(err)
	}
	want := string(lic)
	want = eraseYear(want)
	want = removeCommentPrefix(want, "//")
	// Remove final blank line(s).
	want = strings.TrimRightFunc(want, unicode.IsSpace)

	check := func(t *testing.T, file, prefix, contents string) {
		t.Helper()
		got := removeCommentPrefix(contents, prefix)
		got = eraseYear(got)
		if got != want {
			t.Errorf("%s: bad license: does not match contents of %s", file, wantFile)
			t.Logf("got  %q", got)
			t.Logf("want %q", want)
		}
	}

	t.Run("scripts", func(t *testing.T) {
		shellScripts, err := globTree(".", "*.sh")
		if err != nil {
			t.Fatal(err)
		}
		for _, f := range shellScripts {
			got, err := topComment(f, "#")
			if err != nil {
				t.Fatal(err)
			}
			// Remove shbang line.
			if strings.HasPrefix(got, "#!") {
				if i := strings.IndexByte(got, '\n'); i > 0 {
					got = got[i+1:]
				}
			}
			check(t, f, "#", got)
		}
	})
	t.Run("go source", func(t *testing.T) {
		goFiles, err := globTree(".", "*.go")
		if err != nil {
			t.Fatal(err)
		}
		for _, f := range goFiles {
			got, err := topComment(f, "//")
			if err != nil {
				t.Fatal(err)
			}
			check(t, f, "//", got)
		}
	})
}

var yearRegexp = regexp.MustCompile(`[Cc]opyright \d\d\d\d`)

func eraseYear(s string) string {
	return yearRegexp.ReplaceAllLiteralString(s, "Copyright YYYY")
}

func removeCommentPrefix(s, prefix string) string {
	var lines []string
	for _, line := range strings.Split(s, "\n") {
		lines = append(lines, strings.TrimPrefix(line, prefix))
	}
	return strings.Join(lines, "\n")
}

// topComment returns the comment at the top of the file, up to the first blank or non-comment line.
// Exception: the first comment contains "generated", in which case we take the second one.
func topComment(file, commentPrefix string) (string, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer f.Close()
	scan := bufio.NewScanner(f)
	var lines []string
	gen := false
	n := 0
	for scan.Scan() {
		line := scan.Text()
		n++
		if n == 1 && strings.Contains(line, "generated") {
			gen = true
			continue
		}
		if gen && line == "" {
			gen = false
			continue
		}
		if !gen {
			if strings.HasPrefix(line, commentPrefix) {
				lines = append(lines, line)
			} else {
				break
			}
		}
	}
	if scan.Err() != nil {
		return "", scan.Err()
	}
	return strings.Join(lines, "\n"), nil
}

// globTree runs filepath.Glob on dir and all its  subdirectories, recursively.
// The filenames it returns begin with dir.
// The pattern must not contain path separators.
func globTree(dir, pattern string) ([]string, error) {
	if strings.ContainsRune(pattern, filepath.Separator) {
		return nil, errors.New("pattern contains path separator")
	}

	// Check for bad pattern.
	if _, err := filepath.Match(pattern, ""); err != nil {
		return nil, err
	}
	var paths []string
	err := filepath.WalkDir(dir, func(path string, _ fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if ok, _ := filepath.Match(pattern, filepath.Base(path)); ok {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return paths, nil
}
