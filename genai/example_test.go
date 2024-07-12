// This file was generated from internal/samples/docs-snippets_test.go. DO NOT EDIT.

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
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"github.com/google/generative-ai-go/genai/internal/testhelpers"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

var testDataDir = filepath.Join(testhelpers.ModuleRootDir(), "genai", "testdata")

func ExampleGenerativeModel_GenerateContent() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-1.5-pro")
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

	model := client.GenerativeModel("gemini-1.5-pro")
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

func ExampleGenerativeModel_GenerateContent_codeExecution() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-1.5-pro")
	// To enable code execution, set the CodeExecution tool.
	model.Tools = []*genai.Tool{{CodeExecution: &genai.CodeExecution{}}}
	resp, err := model.GenerateContent(ctx, genai.Text(`
		788477675 * 778 = x.  Find x and also compute largest odd number smaller than this number.
		`))
	if err != nil {
		log.Fatal(err)
	}
	// The model will generate code to solve the problem, which is returned in an ExecutableCode part.
	// It will also run that code and use the result, which is returned in a CodeExecutionResult part.
	printResponse(resp)
}

func ExampleGenerativeModel_GenerateContentStream() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-1.5-pro")

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

func ExampleGenerativeModel_CountTokens_contextWindow() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-1.0-pro-001")
	info, err := model.Info(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// Returns the "context window" for the model,
	// which is the combined input and output token limits.
	fmt.Printf("input_token_limit=%v\n", info.InputTokenLimit)
	fmt.Printf("output_token_limit=%v\n", info.OutputTokenLimit)
	// ( input_token_limit=30720, output_token_limit=2048 )

}

func ExampleGenerativeModel_CountTokens_textOnly() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-1.5-flash")
	prompt := "The quick brown fox jumps over the lazy dog"

	tokResp, err := model.CountTokens(ctx, genai.Text(prompt))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("total_tokens:", tokResp.TotalTokens)
	// ( total_tokens: 10 )

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("prompt_token_count:", resp.UsageMetadata.PromptTokenCount)
	fmt.Println("candidates_token_count:", resp.UsageMetadata.CandidatesTokenCount)
	fmt.Println("total_token_count:", resp.UsageMetadata.TotalTokenCount)
	// ( prompt_token_count: 10, candidates_token_count: 38, total_token_count: 48 )
}

func ExampleGenerativeModel_CountTokens_cachedContent() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	txt := strings.Repeat("George Washington was the first president of the United States. ", 3000)
	argcc := &genai.CachedContent{
		Model:    "gemini-1.5-flash-001",
		Contents: []*genai.Content{{Role: "user", Parts: []genai.Part{genai.Text(txt)}}},
	}
	cc, err := client.CreateCachedContent(ctx, argcc)
	if err != nil {
		log.Fatal(err)
	}
	defer client.DeleteCachedContent(ctx, cc.Name)

	modelWithCache := client.GenerativeModelFromCachedContent(cc)
	prompt := "Summarize this statement"
	tokResp, err := modelWithCache.CountTokens(ctx, genai.Text(prompt))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("total_tokens:", tokResp.TotalTokens)
	// ( total_tokens: 5 )

	resp, err := modelWithCache.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("prompt_token_count:", resp.UsageMetadata.PromptTokenCount)
	fmt.Println("candidates_token_count:", resp.UsageMetadata.CandidatesTokenCount)
	fmt.Println("cached_content_token_count:", resp.UsageMetadata.CachedContentTokenCount)
	fmt.Println("total_token_count:", resp.UsageMetadata.TotalTokenCount)
	// ( prompt_token_count: 33007,  candidates_token_count: 39, cached_content_token_count: 33002, total_token_count: 33046 )

}

func ExampleGenerativeModel_CountTokens_imageInline() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-1.5-flash")
	prompt := "Tell me about this image"
	imageFile, err := os.ReadFile(filepath.Join(testDataDir, "personWorkingOnComputer.jpg"))
	if err != nil {
		log.Fatal(err)
	}
	// Call `CountTokens` to get the input token count
	// of the combined text and file (`total_tokens`).
	// An image's display or file size does not affect its token count.
	// Optionally, you can call `count_tokens` for the text and file separately.
	tokResp, err := model.CountTokens(ctx, genai.Text(prompt), genai.ImageData("jpeg", imageFile))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("total_tokens:", tokResp.TotalTokens)
	// ( total_tokens: 264 )

	resp, err := model.GenerateContent(ctx, genai.Text(prompt), genai.ImageData("jpeg", imageFile))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("prompt_token_count:", resp.UsageMetadata.PromptTokenCount)
	fmt.Println("candidates_token_count:", resp.UsageMetadata.CandidatesTokenCount)
	fmt.Println("total_token_count:", resp.UsageMetadata.TotalTokenCount)
	// ( prompt_token_count: 264, candidates_token_count: 100, total_token_count: 364 )

}

func ExampleGenerativeModel_CountTokens_imageUploadFile() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-1.5-flash")
	prompt := "Tell me about this image"
	imageFile, err := os.Open(filepath.Join(testDataDir, "personWorkingOnComputer.jpg"))
	if err != nil {
		log.Fatal(err)
	}
	defer imageFile.Close()

	uploadedFile, err := client.UploadFile(ctx, "", imageFile, nil)
	if err != nil {
		log.Fatal(err)
	}

	fd := genai.FileData{
		URI: uploadedFile.URI,
	}
	// Call `CountTokens` to get the input token count
	// of the combined text and file (`total_tokens`).
	// An image's display or file size does not affect its token count.
	// Optionally, you can call `count_tokens` for the text and file separately.
	tokResp, err := model.CountTokens(ctx, genai.Text(prompt), fd)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("total_tokens:", tokResp.TotalTokens)
	// ( total_tokens: 264 )

	resp, err := model.GenerateContent(ctx, genai.Text(prompt), fd)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("prompt_token_count:", resp.UsageMetadata.PromptTokenCount)
	fmt.Println("candidates_token_count:", resp.UsageMetadata.CandidatesTokenCount)
	fmt.Println("total_token_count:", resp.UsageMetadata.TotalTokenCount)
	// ( prompt_token_count: 264, candidates_token_count: 100, total_token_count: 364 )

}

func ExampleGenerativeModel_CountTokens_chat() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-1.5-flash")
	cs := model.StartChat()

	cs.History = []*genai.Content{
		{
			Parts: []genai.Part{
				genai.Text("Hi my name is Bob"),
			},
			Role: "user",
		},
		{
			Parts: []genai.Part{
				genai.Text("Hi Bob!"),
			},
			Role: "model",
		},
	}

	prompt := "Explain how a computer works to a young child."
	resp, err := cs.SendMessage(ctx, genai.Text(prompt))
	if err != nil {
		log.Fatal(err)
	}

	// On the response for SendMessage, use `UsageMetadata` to get
	// separate input and output token counts
	// (`prompt_token_count` and `candidates_token_count`, respectively),
	// as well as the combined token count (`total_token_count`).
	fmt.Println("prompt_token_count:", resp.UsageMetadata.PromptTokenCount)
	fmt.Println("candidates_token_count:", resp.UsageMetadata.CandidatesTokenCount)
	fmt.Println("total_token_count:", resp.UsageMetadata.TotalTokenCount)
	// ( prompt_token_count: 25, candidates_token_count: 21, total_token_count: 46 )

}

func ExampleGenerativeModel_CountTokens_systemInstruction() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-1.5-flash")
	prompt := "The quick brown fox jumps over the lazy dog"

	// Without system instruction
	respNoInstruction, err := model.CountTokens(ctx, genai.Text(prompt))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("total_tokens:", respNoInstruction.TotalTokens)
	// ( total_tokens: 10 )

	// Same prompt, this time with system instruction
	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{genai.Text("You are a cat. Your name is Neko.")},
	}
	respWithInstruction, err := model.CountTokens(ctx, genai.Text(prompt))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("total_tokens:", respWithInstruction.TotalTokens)
	// ( total_tokens: 21 )

}

// This example shows how to get a JSON response that conforms to a schema.
func ExampleGenerativeModel_jSONSchema() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-1.5-pro-latest")
	// Ask the model to respond with JSON.
	model.ResponseMIMEType = "application/json"
	// Specify the format of the JSON.
	model.ResponseSchema = &genai.Schema{
		Type:  genai.TypeArray,
		Items: &genai.Schema{Type: genai.TypeString},
	}
	res, err := model.GenerateContent(ctx, genai.Text("List the primary colors."))
	if err != nil {
		log.Fatal(err)
	}
	for _, part := range res.Candidates[0].Content.Parts {
		if txt, ok := part.(genai.Text); ok {
			var colors []string
			if err := json.Unmarshal([]byte(txt), &colors); err != nil {
				log.Fatal(err)
			}
			fmt.Println(colors)
		}
	}
}

func ExampleChatSession() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()
	model := client.GenerativeModel("gemini-1.5-pro")
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
	model := client.GenerativeModel("gemini-1.5-pro")
	cs := model.StartChat()

	cs.History = []*genai.Content{
		{
			Parts: []genai.Part{
				genai.Text("Hello, I have 2 dogs in my house."),
			},
			Role: "user",
		},
		{
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

	model := client.GenerativeModel("gemini-1.5-pro")

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

	// To use functions / tools, we have to first define a schema that describes
	// the function to the model. The schema is similar to OpenAPI 3.0.
	schema := &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"location": {
				Type:        genai.TypeString,
				Description: "The city and state, e.g. San Francisco, CA or a zip code e.g. 95616",
			},
			"title": {
				Type:        genai.TypeString,
				Description: "Any movie title",
			},
		},
		Required: []string{"location"},
	}

	movieTool := &genai.Tool{
		FunctionDeclarations: []*genai.FunctionDeclaration{{
			Name:        "find_theaters",
			Description: "find theaters based on location and optionally movie title which is currently playing in theaters",
			Parameters:  schema,
		}},
	}

	model := client.GenerativeModel("gemini-1.5-pro-latest")

	// Before initiating a conversation, we tell the model which tools it has
	// at its disposal.
	model.Tools = []*genai.Tool{movieTool}

	// For using tools, the chat mode is useful because it provides the required
	// chat context. A model needs to have tools supplied to it in the chat
	// history so it can use them in subsequent conversations.
	//
	// The flow of message expected here is:
	//
	// 1. We send a question to the model
	// 2. The model recognizes that it needs to use a tool to answer the question,
	//    an returns a FunctionCall response asking to use the tool.
	// 3. We send a FunctionResponse message, simulating the return value of
	//    the tool for the model's query.
	// 4. The model provides its text answer in response to this message.
	session := model.StartChat()

	res, err := session.SendMessage(ctx, genai.Text("Which theaters in Mountain View show Barbie movie?"))
	if err != nil {
		log.Fatalf("session.SendMessage: %v", err)
	}

	part := res.Candidates[0].Content.Parts[0]
	funcall, ok := part.(genai.FunctionCall)
	if !ok || funcall.Name != "find_theaters" {
		log.Fatalf("expected FunctionCall to find_theaters: %v", part)
	}

	// Expect the model to pass a proper string "location" argument to the tool.
	if _, ok := funcall.Args["location"].(string); !ok {
		log.Fatalf("expected string: %v", funcall.Args["location"])
	}

	// Provide the model with a hard-coded reply.
	res, err = session.SendMessage(ctx, genai.FunctionResponse{
		Name: movieTool.FunctionDeclarations[0].Name,
		Response: map[string]any{
			"theater": "AMC16",
		},
	})
	printResponse(res)
}

func ExampleToolConfig() {
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

	model := client.GenerativeModel("gemini-1.5-pro")
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
