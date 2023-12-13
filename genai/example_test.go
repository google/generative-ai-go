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
	"fmt"
	"log"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

const projectID = "your-project"
const model = "some-model"
const location = "some-location"

func ExampleGenerativeModel_GenerateContent() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey("your-API-key"))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	model := client.GenerativeModel(model)
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
	client, err := genai.NewClient(ctx, option.WithAPIKey("your-api-key"))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	model := client.GenerativeModel(model)
	model.Temperature = 0.9
	model.TopP = 0.5
	model.TopK = 20
	model.MaxOutputTokens = 100
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
	client, err := genai.NewClient(ctx, option.WithAPIKey("your-API-key"))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	model := client.GenerativeModel(model)
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

// This example shows how to get the model to call functions that you provide.
func ExampleGenerativeModel_GenerateContent_tool() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey("your-API-key"))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	model := client.GenerativeModel(model)
	weatherTool := &genai.Tool{
		FunctionDeclarations: []*genai.FunctionDeclaration{{
			Name:        "CurrentWeather",
			Description: "Get the current weather in a given location",
			Parameters: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"location": {
						Type:        genai.TypeString,
						Description: "The city and state, e.g. San Francisco, CA",
					},
					"unit": {
						Type: genai.TypeString,
						Enum: []string{"celsius", "fahrenheit"},
					},
				},
				Required: []string{"location"},
			},
		}},
	}

	model.Tools = []*genai.Tool{weatherTool}
	session := model.StartChat()
	res, err := session.SendMessage(ctx, genai.Text("What is the weather like in New York?"))
	if err != nil {
		log.Fatal(err)
	}
	part := res.Candidates[0].Content.Parts[0]
	funcall, ok := part.(genai.FunctionCall)
	if !ok {
		log.Fatalf("want FunctionCall, got %T", part)
	}
	if funcall.Name != "CurrentWeather" {
		log.Fatalf("unknown function %q", funcall.Name)
	}
	w := getCurrentWeather(funcall.Args["location"])
	resp, err := session.SendMessage(ctx, genai.FunctionResponse{
		Name:     "CurrentWeather",
		Response: map[string]any{"weather_there": w},
	})
	if err != nil {
		log.Fatal(err)
	}
	printResponse(resp)
}

func getCurrentWeather(any) string {
	return "cold"
}

func ExampleGenerativeModel_GenerateContentStream() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey("your-API-key"))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	model := client.GenerativeModel(model)

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
	client, err := genai.NewClient(ctx, option.WithAPIKey("your-API-key"))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	model := client.GenerativeModel(model)

	resp, err := model.CountTokens(ctx, genai.Text("What kind of fish is this?"))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Num tokens:", resp.TotalTokens)
}

func ExampleChatSession() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey("your-API-key"))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()
	model := client.GenerativeModel(model)
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
	client, err := genai.NewClient(ctx, option.WithAPIKey("your-API-key"))
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

func ExampleClient_ListModels() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey("your-API-key"))
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
		for _, part := range cand.Content.Parts {
			fmt.Println(part)
		}
	}
	fmt.Println("---")
}
