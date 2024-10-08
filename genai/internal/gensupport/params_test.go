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
	"net/http"
	"testing"

	"google.golang.org/api/googleapi"
)

func TestSetOptionsGetMulti(t *testing.T) {
	co := googleapi.QueryParameter("key", "foo", "bar")
	urlParams := make(URLParams)
	SetOptions(urlParams, co)
	if got, want := urlParams.Encode(), "key=foo&key=bar"; got != want {
		t.Fatalf("URLParams.Encode() = %q, want %q", got, want)
	}
}

func TestSetHeaders(t *testing.T) {
	userAgent := "google-api-go-client/123"
	contentType := "application/json"
	userHeaders := make(http.Header)
	userHeaders.Set("baz", "300")
	got := SetHeaders(userAgent, contentType, userHeaders, "foo", "100", "bar", "200")

	if len(got) != 6 {
		t.Fatalf("SetHeaders() = %q, want len(6)", got)
	}
}
