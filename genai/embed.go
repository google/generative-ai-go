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
	"context"
	"fmt"

	pb "cloud.google.com/go/ai/generativelanguage/apiv1beta/generativelanguagepb"
)

func (c *Client) EmbeddingModel(name string) *EmbeddingModel {
	return &EmbeddingModel{
		c:        c,
		name:     name,
		fullName: fmt.Sprintf("models/%s", name),
	}
}

// EmbeddingModel is a model that computes embeddings.
// Create one with [Client.EmbeddingModel].
type EmbeddingModel struct {
	c        *Client
	name     string
	fullName string
	TaskType TaskType
}

func (m *EmbeddingModel) EmbedContent(ctx context.Context, parts ...Part) (*EmbedContentResponse, error) {
	return m.EmbedContentTitle(ctx, "", parts...)
}

func (m *EmbeddingModel) EmbedContentTitle(ctx context.Context, title string, parts ...Part) (*EmbedContentResponse, error) {
	req := &pb.EmbedContentRequest{
		Model:   m.fullName,
		Content: newUserContent(parts).toProto(),
	}
	// A non-empty title overrides the task type.
	var tt TaskType
	if title != "" {
		req.Title = &title
		tt = TaskTypeRetrievalDocument
	}
	if tt != TaskTypeUnspecified {
		taskType := pb.TaskType(tt)
		req.TaskType = &taskType
	}
	res, err := m.c.c.EmbedContent(ctx, req)
	if err != nil {
		return nil, err
	}
	return (EmbedContentResponse{}).fromProto(res), nil
}

// 	req := m.newEmbedContentRequest(parts,

// 	req := &pb.EmbedContentRequest{
// 		Model:   m.fullName,
// 		Content: newUserContent(parts).toProto(),
// 		Title:   &title,
// 	}
// 	taskType := TaskTypeRetrievalDocument
// 	req.TaskType = &taskType

// }

// func (m *EmbeddingModel) newEmbedContentRequest(parts []Part, tt TaskType, title string) *pb.EmbedContentRequest {
// 	return req
// }

// 	req := &pb.EmbedContentRequest{
// 		Model:   m.fullName,
// 		Content: newUserContent(parts).toProto(),
// 	}
// 	if m.TaskType != TaskTypeUnspecified {
// 		req.TaskType = (*pb.TaskType)(&m.TaskType)
// 	}
// 	res, err := m.c.c.EmbedContent(ctx, req)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return (EmbedContentResponse{}).fromProto(res), nil
// }
