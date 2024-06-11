// Copyright 2023 Google LLC
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

package genai

import (
	"testing"
	"time"

	pb "cloud.google.com/go/ai/generativelanguage/apiv1beta/generativelanguagepb"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/types/known/durationpb"
	// google.golang.org/protobuf/proto
)

func TestPopulateFile(t *testing.T) {
	f1 := &File{}
	p1 := &pb.File{}
	f2 := &File{Metadata: &FileMetadata{
		Video: &VideoMetadata{Duration: time.Minute},
	}}
	p2 := &pb.File{
		Metadata: &pb.File_VideoMetadata{
			VideoMetadata: &pb.VideoMetadata{
				VideoDuration: durationpb.New(time.Minute),
			},
		},
	}

	for _, test := range []struct {
		f *File
		p *pb.File
	}{
		{f1, p1},
		{f2, p2},
	} {
		var pgot pb.File
		populateFileTo(&pgot, test.f)
		if !cmp.Equal(&pgot, test.p, cmpopts.IgnoreUnexported(pb.File{}, pb.VideoMetadata{}, durationpb.Duration{})) {
			t.Errorf("got %+v, want %+v", &pgot, test.p)
		}

		var fgot File
		populateFileFrom(&fgot, test.p)
		if !cmp.Equal(&fgot, test.f) {
			t.Errorf("got %+v, want %+v", &fgot, test.f)
		}
	}
}
