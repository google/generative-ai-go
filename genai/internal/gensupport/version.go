// Copyright 2020 Google LLC. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gensupport

import (
	"runtime"
	"strings"
	"unicode"
)

// GoVersion returns the Go runtime version. The returned string
// has no whitespace.
func GoVersion() string {
	return goVersion
}

var goVersion = goVer(runtime.Version())

const develPrefix = "devel +"

func goVer(s string) string {
	var builder strings.Builder

	if strings.HasPrefix(s, develPrefix) {
		s = s[len(develPrefix):]
		if p := strings.IndexFunc(s, unicode.IsSpace); p >= 0 {
			s = s[:p]
		}
		builder.WriteString(s)
		return builder.String()
	}

	if strings.HasPrefix(s, "go1") {
		s = s[2:]
		var prerelease string
		if p := strings.IndexFunc(s, notSemverRune); p >= 0 {
			s, prerelease = s[:p], s[p:]
		}
		builder.WriteString(s)
		if strings.HasSuffix(s, ".") {
			builder.WriteString("0")
		} else if strings.Count(s, ".") < 2 {
			builder.WriteString(".0")
		}
		if prerelease != "" {
			builder.WriteString("-")
			builder.WriteString(prerelease)
		}
		return builder.String()
	}
	return ""
}

func notSemverRune(r rune) bool {
	return !strings.ContainsRune("0123456789.", r)
}
