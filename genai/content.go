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
	"fmt"

	pb "cloud.google.com/go/ai/generativelanguage/apiv1/generativelanguagepb"
)

const (
	roleUser  = "user"
	roleModel = "model"
)

// A Part is either a Text, a Blob, or a FileData.
type Part interface {
	toPart() *pb.Part
}

func partToProto(p Part) *pb.Part {
	if p == nil {
		return nil
	}
	return p.toPart()
}

func partFromProto(p *pb.Part) Part {
	switch d := p.Data.(type) {
	case *pb.Part_Text:
		return Text(d.Text)
	case *pb.Part_InlineData:
		return Blob{
			MIMEType: d.InlineData.MimeType,
			Data:     d.InlineData.Data,
		}
	default:
		panic(fmt.Errorf("unknown Part.Data type %T", p.Data))
	}
}

// A Text is a piece of text, like a question or phrase.
type Text string

func (t Text) toPart() *pb.Part {
	return &pb.Part{
		Data: &pb.Part_Text{Text: string(t)},
	}
}

func (b Blob) toPart() *pb.Part {
	return &pb.Part{
		Data: &pb.Part_InlineData{
			InlineData: b.toProto(),
		},
	}
}

// ImageData is a convenience function for creating an image
// Blob for input to a model.
// The format should be the second part of the MIME type, after "image/".
// For example, for a PNG image, pass "png".
func ImageData(format string, data []byte) Blob {
	return Blob{
		MIMEType: "image/" + format,
		Data:     data,
	}
}

// SetTemperature sets the temperature on the config.
func (gc *GenerationConfig) SetTemperature(t float32) {
	gc.Temperature = &t
}
