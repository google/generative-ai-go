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

package genai_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

func ExampleGenerativeModel_GenerateContent() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-1.0-pro")
	resp, err := model.GenerateContent(ctx, genai.Text("What is the average size of a swallow?"))
	if err != nil {
		log.Fatal(err)
	}

	printResponse(resp)
}

// This example shows how to a configure a model. See [GenerationConfig]
// for the complete set of configuration options.
func ExampleGenerativeModel_GenerateContent_config() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-1.0-pro")
	model.SetTemperature(0.9)
	model.SetTopP(0.5)
	model.SetTopK(20)
	model.SetMaxOutputTokens(100)
	resp, err := model.GenerateContent(ctx, genai.Text("What is the average size of a swallow?"))
	if err != nil {
		log.Fatal(err)
	}
	printResponse(resp)
}

// This example shows how to use SafetySettings to change the threshold
// for unsafe responses.
func ExampleGenerativeModel_GenerateContent_safetySetting() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-1.0-pro")
	model.SafetySettings = []*genai.SafetySetting{
		{
			Category:  genai.HarmCategoryDangerousContent,
			Threshold: genai.HarmBlockLowAndAbove,
		},
		{
			Category:  genai.HarmCategoryHarassment,
			Threshold: genai.HarmBlockMediumAndAbove,
		},
	}
	resp, err := model.GenerateContent(ctx, genai.Text("I want to be bad. Please help."))
	if err != nil {
		log.Fatal(err)
	}
	printResponse(resp)
}

func ExampleGenerativeModel_GenerateContentStream() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-1.0-pro")

	iter := model.GenerateContentStream(ctx, genai.Text("Tell me a story about a lumberjack and his giant ox. Keep it very short."))
	for {
		resp, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		printResponse(resp)
	}
}

func ExampleGenerativeModel_CountTokens() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-1.0-pro")

	resp, err := model.CountTokens(ctx, genai.Text("What kind of fish is this?"))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Num tokens:", resp.TotalTokens)
}

func ExampleChatSession() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()
	model := client.GenerativeModel("gemini-1.0-pro")
	cs := model.StartChat()

	send := func(msg string) *genai.GenerateContentResponse {
		fmt.Printf("== Me: %s\n== Model:\n", msg)
		res, err := cs.SendMessage(ctx, genai.Text(msg))
		if err != nil {
			log.Fatal(err)
		}
		return res
	}

	res := send("Can you name some brands of air fryer?")
	printResponse(res)
	iter := cs.SendMessageStream(ctx, genai.Text("Which one of those do you recommend?"))
	for {
		res, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		printResponse(res)
	}

	for i, c := range cs.History {
		log.Printf("    %d: %+v", i, c)
	}
	res = send("Why do you like the Philips?")
	if err != nil {
		log.Fatal(err)
	}
	printResponse(res)
}

func ExampleEmbeddingModel_EmbedContent() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()
	em := client.EmbeddingModel("embedding-001")
	res, err := em.EmbedContent(ctx, genai.Text("cheddar cheese"))

	if err != nil {
		panic(err)
	}
	fmt.Println(res.Embedding.Values)
}

func ExampleEmbeddingBatch() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()
	em := client.EmbeddingModel("embedding-001")
	b := em.NewBatch().
		AddContent(genai.Text("cheddar cheese")).
		AddContentWithTitle("My Cheese Report", genai.Text("I love cheddar cheese."))
	res, err := em.BatchEmbedContents(ctx, b)
	if err != nil {
		panic(err)
	}
	for _, e := range res.Embeddings {
		fmt.Println(e.Values)
	}
}

// This example shows how to get more information from an error.
func ExampleGenerativeModel_GenerateContentStream_errors() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("API_KEY")))
	if err != nil {
		log.Fatal(err)
	}

	model := client.GenerativeModel("gemini-1.0-pro")

	iter := model.GenerateContentStream(ctx, genai.ImageData("foo", []byte("bar")))
	res, err := iter.Next()
	if err != nil {
		var gerr *googleapi.Error
		if !errors.As(err, &gerr) {
			log.Fatalf("error: %s\n", err)
		} else {
			log.Fatalf("error details: %s\n", gerr)
		}
	}
	_ = res
}

func ExampleClient_ListModels() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	iter := client.ListModels(ctx)
	for {
		m, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			panic(err)
		}
		fmt.Println(m.Name, m.Description)
	}
}

func printResponse(resp *genai.GenerateContentResponse) {
	for _, cand := range resp.Candidates {
		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				fmt.Println(part)
			}
		}
	}
	fmt.Println("---")
}
