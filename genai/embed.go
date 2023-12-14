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

	pb "cloud.google.com/go/ai/generativelanguage/apiv1/generativelanguagepb"
)

// EmbeddingModel creates a new instance of the named embedding model.
// Example name: "embedding-001" or "models/embedding-001".
func (c *Client) EmbeddingModel(name string) *EmbeddingModel {
	return &EmbeddingModel{
		c:        c,
		name:     name,
		fullName: fullModelName(name),
	}
}

// EmbeddingModel is a model that computes embeddings.
// Create one with [Client.EmbeddingModel].
type EmbeddingModel struct {
	c        *Client
	name     string
	fullName string
	// TaskType describes how the embedding will be used.
	TaskType TaskType
}

// EmbedContent returns an embedding for the list of parts.
func (m *EmbeddingModel) EmbedContent(ctx context.Context, parts ...Part) (*EmbedContentResponse, error) {
	return m.EmbedContentWithTitle(ctx, "", parts...)
}

// EmbedContentWithTitle returns an embedding for the list of parts.
// If the given title is non-empty, it is passed to the model and
// the task type is set to TaskTypeRetrievalDocument.
func (m *EmbeddingModel) EmbedContentWithTitle(ctx context.Context, title string, parts ...Part) (*EmbedContentResponse, error) {
	req := &pb.EmbedContentRequest{
		Model:   m.fullName,
		Content: newUserContent(parts).toProto(),
	}
	// A non-empty title overrides the task type.
	tt := m.TaskType
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
