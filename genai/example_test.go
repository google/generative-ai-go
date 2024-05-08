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
	"net/http"
	"net/url"
	"os"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

func ExampleGenerativeModel_GenerateContent() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
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
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-1.5-pro-latest")
	model.SetTemperature(0.9)
	model.SetTopP(0.5)
	model.SetTopK(20)
	model.SetMaxOutputTokens(100)
	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{genai.Text("You are Yoda from Star Wars.")},
	}
	model.ResponseMIMEType = "application/json"
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
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
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
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
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
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
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
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
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

// This example shows how to set the History field on ChatSession explicitly.
func ExampleChatSession_history() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()
	model := client.GenerativeModel("gemini-1.0-pro")
	cs := model.StartChat()

	cs.History = []*genai.Content{
		&genai.Content{
			Parts: []genai.Part{
				genai.Text("Hello, I have 2 dogs in my house."),
			},
			Role: "user",
		},
		&genai.Content{
			Parts: []genai.Part{
				genai.Text("Great to meet you. What would you like to know?"),
			},
			Role: "model",
		},
	}

	res, err := cs.SendMessage(ctx, genai.Text("How many paws are in my house?"))
	if err != nil {
		log.Fatal(err)
	}
	printResponse(res)
}

func ExampleEmbeddingModel_EmbedContent() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()
	em := client.EmbeddingModel("embedding-001")
	res, err := em.EmbedContent(ctx, genai.Text("cheddar cheese"))

	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(res.Embedding.Values)
}

func ExampleEmbeddingBatch() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
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
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
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
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
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

func ExampleTool() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	currentWeather := func(city string) string {
		switch city {
		case "New York, NY":
			return "cold"
		case "Miami, FL":
			return "warm"
		default:
			return "unknown"
		}
	}

	// To use functions / tools, we have to first define a schema that describes
	// the function to the model. The schema is similar to OpenAPI 3.0.
	//
	// In this example, we create a single function that provides the model with
	// a weather forecast in a given location.
	schema := &genai.Schema{
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
	}

	weatherTool := &genai.Tool{
		FunctionDeclarations: []*genai.FunctionDeclaration{{
			Name:        "CurrentWeather",
			Description: "Get the current weather in a given location",
			Parameters:  schema,
		}},
	}

	model := client.GenerativeModel("gemini-1.0-pro")

	// Before initiating a conversation, we tell the model which tools it has
	// at its disposal.
	model.Tools = []*genai.Tool{weatherTool}

	// For using tools, the chat mode is useful because it provides the required
	// chat context. A model needs to have tools supplied to it in the chat
	// history so it can use them in subsequent conversations.
	//
	// The flow of message expected here is:
	//
	// 1. We send a question to the model
	// 2. The model recognizes that it needs to use a tool to answer the question,
	//    an returns a FunctionCall response asking to use the CurrentWeather
	//    tool.
	// 3. We send a FunctionResponse message, simulating the return value of
	//    CurrentWeather for the model's query.
	// 4. The model provides its text answer in response to this message.
	session := model.StartChat()

	res, err := session.SendMessage(ctx, genai.Text("What is the weather like in New York?"))
	if err != nil {
		log.Fatal(err)
	}

	part := res.Candidates[0].Content.Parts[0]
	funcall, ok := part.(genai.FunctionCall)
	if !ok {
		log.Fatalf("expected FunctionCall: %v", part)
	}

	if funcall.Name != "CurrentWeather" {
		log.Fatalf("expected CurrentWeather: %v", funcall.Name)
	}

	// Expect the model to pass a proper string "location" argument to the tool.
	locArg, ok := funcall.Args["location"].(string)
	if !ok {
		log.Fatalf("expected string: %v", funcall.Args["location"])
	}

	weatherData := currentWeather(locArg)
	res, err = session.SendMessage(ctx, genai.FunctionResponse{
		Name: weatherTool.FunctionDeclarations[0].Name,
		Response: map[string]any{
			"weather": weatherData,
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	printResponse(res)
}

func ExampleToolConifg() {
	// This example shows how to affect how the model uses the tools provided to it.
	// By setting the ToolConfig, you can disable function calling.

	// Assume we have created a Model and have set its Tools field with some functions.
	// See the Example for Tool for details.
	var model *genai.GenerativeModel

	// By default, the model will use the functions in its responses if it thinks they are
	// relevant, by returning FunctionCall parts.
	// Here we set the model's ToolConfig to disable function calling completely.
	model.ToolConfig = &genai.ToolConfig{
		FunctionCallingConfig: &genai.FunctionCallingConfig{
			Mode: genai.FunctionCallingNone,
		},
	}

	// Subsequent calls to ChatSession.SendMessage will not result in FunctionCall responses.
	session := model.StartChat()
	res, err := session.SendMessage(context.Background(), genai.Text("What is the weather like in New York?"))
	if err != nil {
		log.Fatal(err)
	}
	for _, part := range res.Candidates[0].Content.Parts {
		if _, ok := part.(genai.FunctionCall); ok {
			log.Fatal("did not expect FunctionCall")
		}
	}

	// It is also possible to force a function call by using FunctionCallingAny
	// instead of FunctionCallingNone. See the documentation for FunctionCallingMode
	// for details.
}

func ExampleClient_UploadFile() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// Use Client.UploadFile to Upload a file to the service.
	// Pass it an io.Reader.
	f, err := os.Open("path/to/file")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	// You can choose a name, or pass the empty string to generate a unique one.
	file, err := client.UploadFile(ctx, "", f, nil)
	if err != nil {
		log.Fatal(err)
	}
	// The return value's URI field should be passed to the model in a FileData part.
	model := client.GenerativeModel("gemini-1.5-pro")

	resp, err := model.GenerateContent(ctx, genai.FileData{URI: file.URI})
	if err != nil {
		log.Fatal(err)
	}
	_ = resp // Use resp as usual.
}

// ProxyRoundTripper is an implementation of http.RoundTripper that supports
// setting a proxy server URL for genai clients. This type should be used with
// a custom http.Client that's passed to WithHTTPClient. For such clients,
// WithAPIKey doesn't apply so the key has to be explicitly set here.
type ProxyRoundTripper struct {
	// APIKey is the API Key to set on requests.
	APIKey string

	// ProxyURL is the URL of the proxy server. If empty, no proxy is used.
	ProxyURL string
}

func (t *ProxyRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	transport := http.DefaultTransport.(*http.Transport).Clone()

	if t.ProxyURL != "" {
		proxyURL, err := url.Parse(t.ProxyURL)
		if err != nil {
			return nil, err
		}
		transport.Proxy = http.ProxyURL(proxyURL)
	}

	newReq := req.Clone(req.Context())
	vals := newReq.URL.Query()
	vals.Set("key", t.APIKey)
	newReq.URL.RawQuery = vals.Encode()

	resp, err := transport.RoundTrip(newReq)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func ExampleClient_setProxy() {
	c := &http.Client{Transport: &ProxyRoundTripper{
		APIKey:   os.Getenv("GEMINI_API_KEY"),
		ProxyURL: "http://<proxy-url>",
	}}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithHTTPClient(c))

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
