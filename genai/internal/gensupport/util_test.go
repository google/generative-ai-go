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

package gensupport

import (
	"io"
	"time"
)

// errReader reads out of a buffer until it is empty, then returns the specified error.
type errReader struct {
	buf []byte
	err error
}

func (er *errReader) Read(p []byte) (int, error) {
	if len(er.buf) == 0 {
		if er.err == nil {
			return 0, io.EOF
		}
		return 0, er.err
	}
	n := copy(p, er.buf)
	er.buf = er.buf[n:]
	return n, nil
}

// NoPauseBackoff implements backoff with infinite 0-length pauses.
type NoPauseBackoff struct{}

func (bo *NoPauseBackoff) Pause() time.Duration { return 0 }

// PauseOneSecond implements backoff with infinite 1s pauses.
type PauseOneSecond struct{}

func (bo *PauseOneSecond) Pause() time.Duration { return time.Second }
