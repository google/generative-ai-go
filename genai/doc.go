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

// Package genai is a client for the Google AI generative models.
//
// NOTE: This client uses the v1beta version of the API.
//
// # Getting started
//
// Reading the [examples] is the best way to learn how to use this package.
//
// # Authorization
//
// You will need an API key to use the service.
// See the [setup tutorial] for details.
//
// # Tools
//
// Gemini can call functions if you tell it about them.
// Create FunctionDeclarations, add them to a Tool, and install the Tool in a Model.
// When used in a ChatSession, the content returned from a model may include FunctionCall
// parts. Your code performs the requested call and sends back a FunctionResponse.
// See The example for Tool
//
// To have the SDK call a Go function for you, assign it to the FunctionDeclaration.Function.
// field. A ChatSession will look for FunctionCalls, invoke the function you supply, and reply
// with a FunctionResponse. Your code will see only the final result.
//
// The NewCallableFunctionDeclaration function will infer the schema for a function you supply,
// and create a FunctionDeclaration that exposes that function for automatic calling.
// See the example for NewCallableFunctionDeclaration.
//
// # Errors
//
// [examples]: https://pkg.go.dev/github.com/google/generative-ai-go/genai#pkg-examples
// [setup tutorial]: https://ai.google.dev/tutorials/setup
package genai
