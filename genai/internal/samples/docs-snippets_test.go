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

//go:generate go run ../cmd/gen-examples/gen-examples.go -in $GOFILE -out ../../example_test.go

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
	"time"

	"github.com/google/generative-ai-go/genai"
	"github.com/google/generative-ai-go/genai/internal/testhelpers"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

var testDataDir = filepath.Join(testhelpers.ModuleRootDir(), "genai", "testdata")

func ExampleGenerativeModel_GenerateContent_textOnly() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// [START text_gen_text_only_prompt]
	model := client.GenerativeModel("gemini-1.5-flash")
	resp, err := model.GenerateContent(ctx, genai.Text("Write a story about a magic backpack."))
	if err != nil {
		log.Fatal(err)
	}

	printResponse(resp)
	// [END text_gen_text_only_prompt]
}

func ExampleGenerativeModel_GenerateContent_imagePrompt() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// [START text_gen_multimodal_one_image_prompt]
	model := client.GenerativeModel("gemini-1.5-flash")

	imgData, err := os.ReadFile(filepath.Join(testDataDir, "organ.jpg"))
	if err != nil {
		log.Fatal(err)
	}

	resp, err := model.GenerateContent(ctx,
		genai.Text("Tell me about this instrument"),
		genai.ImageData("jpeg", imgData))
	if err != nil {
		log.Fatal(err)
	}

	printResponse(resp)
	// [END text_gen_multimodal_one_image_prompt]
}

func ExampleGenerativeModel_GenerateContent_videoPrompt() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// [START text_gen_multimodal_video_prompt]
	model := client.GenerativeModel("gemini-1.5-flash")

	file, err := client.UploadFileFromPath(ctx, filepath.Join(testDataDir, "earth.mp4"), nil)
	if err != nil {
		log.Fatal(err)
	}
	defer client.DeleteFile(ctx, file.Name)

	// Videos need to be processed before you can use them.
	for file.State == genai.FileStateProcessing {
		log.Printf("processing %s", file.Name)
		time.Sleep(5 * time.Second)
		var err error
		if file, err = client.GetFile(ctx, file.Name); err != nil {
			log.Fatal(err)
		}
	}
	if file.State != genai.FileStateActive {
		log.Fatalf("uploaded file has state %s, not active", file.State)
	}

	resp, err := model.GenerateContent(ctx,
		genai.Text("Describe this video clip"),
		genai.FileData{URI: file.URI})
	if err != nil {
		log.Fatal(err)
	}

	printResponse(resp)
	// [END text_gen_multimodal_video_prompt]
}

func ExampleGenerativeModel_GenerateContent_pdfPrompt() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// [START text_gen_multimodal_pdf]
	model := client.GenerativeModel("gemini-1.5-flash")

	file, err := client.UploadFileFromPath(ctx, filepath.Join(testDataDir, "test.pdf"), nil)
	if err != nil {
		log.Fatal(err)
	}
	defer client.DeleteFile(ctx, file.Name)

	resp, err := model.GenerateContent(ctx,
		genai.Text("Give me a summary of this document:"),
		genai.FileData{URI: file.URI})
	if err != nil {
		log.Fatal(err)
	}

	printResponse(resp)
	// [END text_gen_multimodal_pdf_prompt]
}

func ExampleGenerativeModel_GenerateContent_multiImagePrompt() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// [START text_gen_multimodal_multi_image_prompt]
	model := client.GenerativeModel("gemini-1.5-flash")

	imgData1, err := os.ReadFile(filepath.Join(testDataDir, "Cajun_instruments.jpg"))
	if err != nil {
		log.Fatal(err)
	}
	imgData2, err := os.ReadFile(filepath.Join(testDataDir, "organ.jpg"))
	if err != nil {
		log.Fatal(err)
	}

	resp, err := model.GenerateContent(ctx,
		genai.Text("What is the difference between these instruments?"),
		genai.ImageData("jpeg", imgData1),
		genai.ImageData("jpeg", imgData2),
	)
	if err != nil {
		log.Fatal(err)
	}

	printResponse(resp)
	// [END text_gen_multimodal_multi_image_prompt]
}

func ExampleGenerativeModel_GenerateContentStream_multiImagePrompt() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// [START text_gen_multimodal_multi_image_prompt_streaming]
	model := client.GenerativeModel("gemini-1.5-flash")

	imgData1, err := os.ReadFile(filepath.Join(testDataDir, "Cajun_instruments.jpg"))
	if err != nil {
		log.Fatal(err)
	}
	imgData2, err := os.ReadFile(filepath.Join(testDataDir, "organ.jpg"))
	if err != nil {
		log.Fatal(err)
	}

	iter := model.GenerateContentStream(ctx,
		genai.Text("What is the difference between these instruments?"),
		genai.ImageData("jpeg", imgData1),
		genai.ImageData("jpeg", imgData2),
	)
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
	// [END text_gen_multimodal_multi_image_prompt_streaming]
}

func ExampleGenerativeModel_GenerateContent_config() {
	// This example shows how to a configure a model. See [GenerationConfig]
	// for the complete set of configuration options.
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// [START configure_model_parameters]
	model := client.GenerativeModel("gemini-1.5-pro-latest")
	model.SetTemperature(0.9)
	model.SetTopP(0.5)
	model.SetTopK(20)
	model.SetMaxOutputTokens(100)
	model.SystemInstruction = genai.NewUserContent(genai.Text("You are Yoda from Star Wars."))
	model.ResponseMIMEType = "application/json"
	resp, err := model.GenerateContent(ctx, genai.Text("What is the average size of a swallow?"))
	if err != nil {
		log.Fatal(err)
	}
	printResponse(resp)
	// [END configure_model_parameters]
}

func ExampleGenerativeModel_GenerateContent_systemInstruction() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// [START system_instruction]
	model := client.GenerativeModel("gemini-1.5-flash")
	model.SystemInstruction = genai.NewUserContent(genai.Text("You are a cat. Your name is Neko."))
	resp, err := model.GenerateContent(ctx, genai.Text("Good morning! How are you?"))
	if err != nil {
		log.Fatal(err)
	}
	printResponse(resp)
	// [END system_instruction]
}

func ExampleGenerativeModel_GenerateContent_safetySetting() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// [START safety_settings]
	model := client.GenerativeModel("gemini-1.5-flash")
	model.SafetySettings = []*genai.SafetySetting{
		{
			Category:  genai.HarmCategoryHarassment,
			Threshold: genai.HarmBlockOnlyHigh,
		},
	}
	resp, err := model.GenerateContent(ctx, genai.Text("I support Martians Soccer Club and I think Jupiterians Football Club sucks! Write a ironic phrase about them."))
	if err != nil {
		log.Fatal(err)
	}
	printResponse(resp)
	// [END safety_settings]
}

func ExampleGenerativeModel_GenerateContent_safetySettingMulti() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// [START safety_settings_multi]
	model := client.GenerativeModel("gemini-1.5-flash")
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
	resp, err := model.GenerateContent(ctx, genai.Text("I support Martians Soccer Club and I think Jupiterians Football Club sucks! Write a ironic phrase about them."))
	if err != nil {
		log.Fatal(err)
	}
	printResponse(resp)
	// [END safety_settings_multi]
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

	// [START text_gen_text_only_prompt_streaming]
	model := client.GenerativeModel("gemini-1.5-flash")
	iter := model.GenerateContentStream(ctx, genai.Text("Write a story about a magic backpack."))
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
	// [END text_gen_text_only_prompt_streaming]
}

func ExampleGenerativeModel_GenerateContentStream_imagePrompt() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// [START text_gen_multimodal_one_image_prompt_streaming]
	model := client.GenerativeModel("gemini-1.5-flash")

	imgData, err := os.ReadFile(filepath.Join(testDataDir, "organ.jpg"))
	if err != nil {
		log.Fatal(err)
	}
	iter := model.GenerateContentStream(ctx,
		genai.Text("Tell me about this instrument"),
		genai.ImageData("jpeg", imgData))
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
	// [END text_gen_multimodal_one_image_prompt_streaming]
}

func ExampleGenerativeModel_GenerateContentStream_videoPrompt() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// [START text_gen_multimodal_video_prompt_streaming]
	model := client.GenerativeModel("gemini-1.5-flash")

	file, err := client.UploadFileFromPath(ctx, filepath.Join(testDataDir, "earth.mp4"), nil)
	if err != nil {
		log.Fatal(err)
	}
	defer client.DeleteFile(ctx, file.Name)

	iter := model.GenerateContentStream(ctx,
		genai.Text("Describe this video clip"),
		genai.FileData{URI: file.URI})
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
	// [END text_gen_multimodal_video_prompt_streaming]
}

func ExampleGenerativeModel_GenerateContent_audioPrompt() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// [START text_gen_multimodal_audio]
	model := client.GenerativeModel("gemini-1.5-flash")

	file, err := client.UploadFileFromPath(ctx, filepath.Join(testDataDir, "sample.mp3"), nil)
	if err != nil {
		log.Fatal(err)
	}
	defer client.DeleteFile(ctx, file.Name)

	resp, err := model.GenerateContent(ctx,
		genai.Text("Give me a summary of this audio file."),
		genai.FileData{URI: file.URI})
	if err != nil {
		log.Fatal(err)
	}
	printResponse(resp)
	// [END text_gen_multimodal_audio]
}

func ExampleGenerativeModel_GenerateContentStream_audioPrompt() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// [START text_gen_multimodal_audio_streaming]
	model := client.GenerativeModel("gemini-1.5-flash")

	file, err := client.UploadFileFromPath(ctx, filepath.Join(testDataDir, "sample.mp3"), nil)
	if err != nil {
		log.Fatal(err)
	}
	defer client.DeleteFile(ctx, file.Name)

	iter := model.GenerateContentStream(ctx,
		genai.Text("Give me a summary of this audio file."),
		genai.FileData{URI: file.URI})
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
	// [END text_gen_multimodal_audio_streaming]
}

func ExampleGenerativeModel_GenerateContentStream_pdfPrompt() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// [START text_gen_multimodal_pdf_streaming]
	model := client.GenerativeModel("gemini-1.5-flash")

	file, err := client.UploadFileFromPath(ctx, filepath.Join(testDataDir, "test.pdf"), nil)
	if err != nil {
		log.Fatal(err)
	}
	defer client.DeleteFile(ctx, file.Name)

	iter := model.GenerateContentStream(ctx,
		genai.Text("Give me a summary of this document:"),
		genai.FileData{URI: file.URI})
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
	// [END text_gen_multimodal_pdf_streaming]
}

func ExampleGenerativeModel_CountTokens_contextWindow() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// [START tokens_context_window]
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
	// [END tokens_context_window]
}

func ExampleGenerativeModel_CountTokens_textOnly() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// [START tokens_text_only]
	model := client.GenerativeModel("gemini-1.5-flash")
	prompt := "The quick brown fox jumps over the lazy dog"

	// Call CountTokens to get the input token count (`total tokens`).
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

	// On the response for GenerateContent, use UsageMetadata to get
	// separate input and output token counts (PromptTokenCount and
	// CandidatesTokenCount, respectively), as well as the combined
	// token count (TotalTokenCount).
	fmt.Println("prompt_token_count:", resp.UsageMetadata.PromptTokenCount)
	fmt.Println("candidates_token_count:", resp.UsageMetadata.CandidatesTokenCount)
	fmt.Println("total_token_count:", resp.UsageMetadata.TotalTokenCount)
	// ( prompt_token_count: 10, candidates_token_count: 38, total_token_count: 48 )
	// [END tokens_text_only]
}

func ExampleGenerativeModel_CountTokens_tools() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// [START tokens_tools]
	model := client.GenerativeModel("gemini-1.5-flash-001")
	prompt := "I have 57 cats, each owns 44 mittens, how many mittens is that in total?"

	tokResp, err := model.CountTokens(ctx, genai.Text(prompt))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("total_tokens:", tokResp.TotalTokens)
	// ( total_tokens: 23 )

	tools := []*genai.Tool{
		&genai.Tool{FunctionDeclarations: []*genai.FunctionDeclaration{
			{Name: "add"},
			{Name: "subtract"},
			{Name: "multiply"},
			{Name: "divide"},
		}}}

	model.Tools = tools

	// The total token count includes everything sent to the GenerateContent
	// request. When you use tools (like function calling), the total
	// token count increases.
	tokResp, err = model.CountTokens(ctx, genai.Text(prompt))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("total_tokens:", tokResp.TotalTokens)
	// ( total_tokens: 99 )

	// [END tokens_tools]
}

func ExampleGenerativeModel_CountTokens_cachedContent() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// [START tokens_cached_content]
	txt := strings.Repeat("George Washington was the first president of the United States. ", 3000)
	argcc := &genai.CachedContent{
		Model:    "gemini-1.5-flash-001",
		Contents: []*genai.Content{genai.NewUserContent(genai.Text(txt))},
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
	// [END tokens_cached_content]
}

func ExampleGenerativeModel_CountTokens_imageInline() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// [START tokens_multimodal_image_inline]
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
	// [END tokens_multimodal_image_inline]
}

func ExampleGenerativeModel_CountTokens_imageUploadFile() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// [START tokens_multimodal_image_file_api]
	model := client.GenerativeModel("gemini-1.5-flash")
	prompt := "Tell me about this image"
	file, err := client.UploadFileFromPath(ctx, filepath.Join(testDataDir, "personWorkingOnComputer.jpg"), nil)
	if err != nil {
		log.Fatal(err)
	}
	defer client.DeleteFile(ctx, file.Name)

	fd := genai.FileData{URI: file.URI}
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
	// [END tokens_multimodal_image_file_api]
}

func ExampleGenerativeModel_CountTokens_pdfUploadFile() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// [START tokens_multimodal_pdf_file_api]
	model := client.GenerativeModel("gemini-1.5-flash")
	prompt := "Give me a summary of this document."
	file, err := client.UploadFileFromPath(ctx, filepath.Join(testDataDir, "test.pdf"), nil)
	if err != nil {
		log.Fatal(err)
	}
	defer client.DeleteFile(ctx, file.Name)

	fd := genai.FileData{URI: file.URI}
	resp, err := model.GenerateContent(ctx, genai.Text(prompt), fd)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(resp.UsageMetadata)
	// [END tokens_multimodal_pdf_file_api]
}

func ExampleGenerativeModel_CountTokens_videoUploadFile() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// [START tokens_multimodal_video_audio_file_api]
	model := client.GenerativeModel("gemini-1.5-flash")
	prompt := "Tell me about this video"
	file, err := client.UploadFileFromPath(ctx, filepath.Join(testDataDir, "earth.mp4"), nil)
	if err != nil {
		log.Fatal(err)
	}
	defer client.DeleteFile(ctx, file.Name)

	fd := genai.FileData{URI: file.URI}
	// Call `CountTokens` to get the input token count
	// of the combined text and file (`total_tokens`).
	// A video or audio file is converted to tokens at a fixed rate of tokens per
	// second.
	// Optionally, you can call `count_tokens` for the text and file separately.
	tokResp, err := model.CountTokens(ctx, genai.Text(prompt), fd)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("total_tokens:", tokResp.TotalTokens)
	// ( total_tokens: 1481 )

	resp, err := model.GenerateContent(ctx, genai.Text(prompt), fd)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("prompt_token_count:", resp.UsageMetadata.PromptTokenCount)
	fmt.Println("candidates_token_count:", resp.UsageMetadata.CandidatesTokenCount)
	fmt.Println("total_token_count:", resp.UsageMetadata.TotalTokenCount)
	// ( prompt_token_count: 1481, candidates_token_count: 43, total_token_count: 1524 )

	// [END tokens_multimodal_video_audio_file_api]
}

func ExampleGenerativeModel_CountTokens_chat() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// [START tokens_chat]
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
	// [END tokens_chat]
}

func ExampleGenerativeModel_CountTokens_systemInstruction() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// [START tokens_system_instruction]
	model := client.GenerativeModel("gemini-1.5-flash")
	prompt := "The quick brown fox jumps over the lazy dog"

	respNoInstruction, err := model.CountTokens(ctx, genai.Text(prompt))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("total_tokens:", respNoInstruction.TotalTokens)
	// ( total_tokens: 10 )

	// The total token count includes everything sent to the GenerateContent
	// request. When you use system instructions, the total token
	// count increases.
	model.SystemInstruction = genai.NewUserContent(genai.Text("You are a cat. Your name is Neko."))
	respWithInstruction, err := model.CountTokens(ctx, genai.Text(prompt))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("total_tokens:", respWithInstruction.TotalTokens)
	// ( total_tokens: 21 )
	// [END tokens_system_instruction]
}

func ExampleGenerativeModel_jSONSchema() {
	// This example shows how to get a JSON response that conforms to a schema.
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// [START json_controlled_generation]
	model := client.GenerativeModel("gemini-1.5-pro-latest")
	// Ask the model to respond with JSON.
	model.ResponseMIMEType = "application/json"
	// Specify the schema.
	model.ResponseSchema = &genai.Schema{
		Type:  genai.TypeArray,
		Items: &genai.Schema{Type: genai.TypeString},
	}
	resp, err := model.GenerateContent(ctx, genai.Text("List a few popular cookie recipes using this JSON schema."))
	if err != nil {
		log.Fatal(err)
	}
	for _, part := range resp.Candidates[0].Content.Parts {
		if txt, ok := part.(genai.Text); ok {
			var recipes []string
			if err := json.Unmarshal([]byte(txt), &recipes); err != nil {
				log.Fatal(err)
			}
			fmt.Println(recipes)
		}
	}
	// [END json_controlled_generation]
}

func ExampleGenerativeModel_jSONNoSchema() {
	// This example shows how to get a JSON response without requestin a specific
	// schema.
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// [START json_no_schema]
	model := client.GenerativeModel("gemini-1.5-pro-latest")
	// Ask the model to respond with JSON.
	model.ResponseMIMEType = "application/json"
	resp, err := model.GenerateContent(ctx, genai.Text("List a few popular cookie recipes."))
	if err != nil {
		log.Fatal(err)
	}

	printResponse(resp)
	// [END json_no_schema]
}

// This example shows how to set the History field on ChatSession explicitly.
func ExampleChatSession_history() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// [START chat]
	model := client.GenerativeModel("gemini-1.5-flash")
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
	// [END chat]
}

func ExampleChatSession_streaming() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// [START chat_streaming]
	model := client.GenerativeModel("gemini-1.5-flash")
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

	iter := cs.SendMessageStream(ctx, genai.Text("How many paws are in my house?"))
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
	// [END chat_streaming]
}

func ExampleChatSession_streamingWithImage() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// [START chat_streaming_with_images]
	model := client.GenerativeModel("gemini-1.5-flash")
	cs := model.StartChat()

	cs.SendMessage(ctx, genai.Text("Hello, I'm interested in learning about musical instruments. Can I show you one?"))

	imgData, err := os.ReadFile(filepath.Join(testDataDir, "organ.jpg"))
	if err != nil {
		log.Fatal(err)
	}

	iter := cs.SendMessageStream(ctx,
		genai.Text("What family of instruments does this instrument belong to?"),
		genai.ImageData("jpeg", imgData))
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
	// [END chat_streaming_with_images]
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

func ExampleClient_UploadFile_text() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// [START files_create_text]
	// Set MIME type explicitly for text files - the service may have difficulty
	// distingushing between different MIME types of text files automatically.
	file, err := client.UploadFileFromPath(ctx, filepath.Join(testDataDir, "poem.txt"), nil)
	if err != nil {
		log.Fatal(err)
	}
	defer client.DeleteFile(ctx, file.Name)

	model := client.GenerativeModel("gemini-1.5-flash")
	resp, err := model.GenerateContent(ctx,
		genai.FileData{URI: file.URI},
		genai.Text("Can you add a few more lines to this poem?"))
	if err != nil {
		log.Fatal(err)
	}

	printResponse(resp)
	// [END files_create_text]
}

func ExampleClient_UploadFile_image() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// [START files_create_image]
	file, err := client.UploadFileFromPath(ctx, filepath.Join(testDataDir, "Cajun_instruments.jpg"), nil)
	if err != nil {
		log.Fatal(err)
	}
	defer client.DeleteFile(ctx, file.Name)

	model := client.GenerativeModel("gemini-1.5-flash")
	resp, err := model.GenerateContent(ctx,
		genai.FileData{URI: file.URI},
		genai.Text("Can you tell me about the instruments in this photo?"))
	if err != nil {
		log.Fatal(err)
	}

	printResponse(resp)
	// [END files_create_image]
}

func ExampleClient_UploadFile_pdf() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// [START files_create_pdf]
	file, err := client.UploadFileFromPath(ctx, filepath.Join(testDataDir, "test.pdf"), nil)
	if err != nil {
		log.Fatal(err)
	}
	defer client.DeleteFile(ctx, file.Name)

	model := client.GenerativeModel("gemini-1.5-flash")
	resp, err := model.GenerateContent(ctx,
		genai.Text("Give me a summary of this pdf file."),
		genai.FileData{URI: file.URI})
	if err != nil {
		log.Fatal(err)
	}

	printResponse(resp)
	// [END files_create_pdf]
}

func ExampleClient_UploadFile_video() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// [START files_create_video]
	file, err := client.UploadFileFromPath(ctx, filepath.Join(testDataDir, "earth.mp4"), nil)
	if err != nil {
		log.Fatal(err)
	}
	defer client.DeleteFile(ctx, file.Name)

	// Videos need to be processed before you can use them.
	for file.State == genai.FileStateProcessing {
		log.Printf("processing %s", file.Name)
		time.Sleep(5 * time.Second)
		var err error
		if file, err = client.GetFile(ctx, file.Name); err != nil {
			log.Fatal(err)
		}
	}
	if file.State != genai.FileStateActive {
		log.Fatalf("uploaded file has state %s, not active", file.State)
	}

	model := client.GenerativeModel("gemini-1.5-flash")
	resp, err := model.GenerateContent(ctx,
		genai.FileData{URI: file.URI},
		genai.Text("Describe this video clip"))
	if err != nil {
		log.Fatal(err)
	}

	printResponse(resp)
	// [END files_create_video]
}

func ExampleClient_UploadFile_audio() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// [START files_create_audio]
	file, err := client.UploadFileFromPath(ctx, filepath.Join(testDataDir, "sample.mp3"), nil)
	if err != nil {
		log.Fatal(err)
	}
	defer client.DeleteFile(ctx, file.Name)

	model := client.GenerativeModel("gemini-1.5-flash")
	resp, err := model.GenerateContent(ctx,
		genai.FileData{URI: file.URI},
		genai.Text("Describe this audio clip"))
	if err != nil {
		log.Fatal(err)
	}

	printResponse(resp)
	// [END files_create_audio]
}

func ExampleClient_GetFile() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// [START files_get]
	// [START files_delete]
	file, err := client.UploadFileFromPath(ctx, filepath.Join(testDataDir, "personWorkingOnComputer.jpg"), nil)
	if err != nil {
		log.Fatal(err)
	}
	defer client.DeleteFile(ctx, file.Name)

	gotFile, err := client.GetFile(ctx, file.Name)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Got file:", gotFile.Name)

	model := client.GenerativeModel("gemini-1.5-flash")
	resp, err := model.GenerateContent(ctx,
		genai.FileData{URI: file.URI},
		genai.Text("Describe this image"))
	if err != nil {
		log.Fatal(err)
	}

	printResponse(resp)
	// [END files_get]
	// [END files_delete]
}

func ExampleClient_ListFiles() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// [START files_list]
	iter := client.ListFiles(ctx)
	for {
		ifile, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(ifile.Name)
	}
	// [END files_list]
}

func ExampleCachedContent_create() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// [START cache_create]
	// [START cache_delete]
	file, err := client.UploadFileFromPath(ctx,
		filepath.Join(testDataDir, "a11.txt"),
		&genai.UploadFileOptions{MIMEType: "text/plain"})
	if err != nil {
		log.Fatal(err)
	}
	defer client.DeleteFile(ctx, file.Name)
	// [END cache_delete]
	fd := genai.FileData{URI: file.URI}

	argcc := &genai.CachedContent{
		Model:             "gemini-1.5-flash-001",
		SystemInstruction: genai.NewUserContent(genai.Text("You are an expert analyzing transcripts.")),
		Contents:          []*genai.Content{genai.NewUserContent(fd)},
	}
	cc, err := client.CreateCachedContent(ctx, argcc)
	if err != nil {
		log.Fatal(err)
	}
	defer client.DeleteCachedContent(ctx, cc.Name)

	modelWithCache := client.GenerativeModelFromCachedContent(cc)
	prompt := "Please summarize this transcript"
	resp, err := modelWithCache.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		log.Fatal(err)
	}

	printResponse(resp)
	// [END cache_create]
}

func ExampleCachedContent_createFromChat() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// [START cache_create_from_chat]
	file, err := client.UploadFileFromPath(ctx, filepath.Join(testDataDir, "a11.txt"), nil)
	if err != nil {
		log.Fatal(err)
	}
	defer client.DeleteFile(ctx, file.Name)
	fd := genai.FileData{URI: file.URI}

	modelName := "gemini-1.5-flash-001"
	model := client.GenerativeModel(modelName)
	model.SystemInstruction = genai.NewUserContent(genai.Text("You are an expert analyzing transcripts."))

	cs := model.StartChat()
	resp, err := cs.SendMessage(ctx, genai.Text("Hi, could you summarize this transcript?"), fd)
	if err != nil {
		log.Fatal(err)
	}

	resp, err = cs.SendMessage(ctx, genai.Text("Okay, could you tell me more about the trans-lunar injection"))
	if err != nil {
		log.Fatal(err)
	}

	// To cache the conversation so far, pass the chat history as the list of
	// contents.

	argcc := &genai.CachedContent{
		Model:             modelName,
		SystemInstruction: model.SystemInstruction,
		Contents:          cs.History,
	}
	cc, err := client.CreateCachedContent(ctx, argcc)
	if err != nil {
		log.Fatal(err)
	}
	defer client.DeleteCachedContent(ctx, cc.Name)

	modelWithCache := client.GenerativeModelFromCachedContent(cc)
	cs = modelWithCache.StartChat()
	resp, err = cs.SendMessage(ctx, genai.Text("I didn't understand that last part, could you please explain it in simpler language?"))
	if err != nil {
		log.Fatal(err)
	}
	printResponse(resp)

	// [END cache_create_from_chat]
}

func ExampleClient_GetCachedContent() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// [START cache_create_from_name]
	// [START cache_get]
	file, err := client.UploadFileFromPath(ctx, filepath.Join(testDataDir, "a11.txt"), nil)
	if err != nil {
		log.Fatal(err)
	}
	defer client.DeleteFile(ctx, file.Name)
	fd := genai.FileData{URI: file.URI}

	argcc := &genai.CachedContent{
		Model:             "gemini-1.5-flash-001",
		SystemInstruction: genai.NewUserContent(genai.Text("You are an expert analyzing transcripts.")),
		Contents:          []*genai.Content{genai.NewUserContent(fd)},
	}
	cc, err := client.CreateCachedContent(ctx, argcc)
	if err != nil {
		log.Fatal(err)
	}
	defer client.DeleteCachedContent(ctx, cc.Name)

	// Save the name for later
	cacheName := cc.Name

	// ... Later
	cc2, err := client.GetCachedContent(ctx, cacheName)
	if err != nil {
		log.Fatal(err)
	}
	modelWithCache := client.GenerativeModelFromCachedContent(cc2)
	prompt := "Find a lighthearted moment from this transcript"
	resp, err := modelWithCache.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		log.Fatal(err)
	}

	printResponse(resp)
	// [END cache_create_from_name]
	// [END cache_get]
}

func ExampleClient_ListCachedContents() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// [START cache_list]
	file, err := client.UploadFileFromPath(ctx, filepath.Join(testDataDir, "a11.txt"), nil)
	if err != nil {
		log.Fatal(err)
	}
	defer client.DeleteFile(ctx, file.Name)
	fd := genai.FileData{URI: file.URI}

	argcc := &genai.CachedContent{
		Model:             "gemini-1.5-flash-001",
		SystemInstruction: genai.NewUserContent(genai.Text("You are an expert analyzing transcripts.")),
		Contents:          []*genai.Content{genai.NewUserContent(fd)},
	}
	cc, err := client.CreateCachedContent(ctx, argcc)
	if err != nil {
		log.Fatal(err)
	}
	defer client.DeleteCachedContent(ctx, cc.Name)

	fmt.Println("My caches:")
	iter := client.ListCachedContents(ctx)
	for {
		cc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("   ", cc.Name)
	}
	// [END cache_list]
}

func ExampleClient_UpdateCachedContent() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// [START cache_update]
	file, err := client.UploadFileFromPath(ctx, filepath.Join(testDataDir, "a11.txt"), nil)
	if err != nil {
		log.Fatal(err)
	}
	defer client.DeleteFile(ctx, file.Name)
	fd := genai.FileData{URI: file.URI}

	argcc := &genai.CachedContent{
		Model:             "gemini-1.5-flash-001",
		SystemInstruction: genai.NewUserContent(genai.Text("You are an expert analyzing transcripts.")),
		Contents:          []*genai.Content{genai.NewUserContent(fd)},
	}
	cc, err := client.CreateCachedContent(ctx, argcc)
	if err != nil {
		log.Fatal(err)
	}
	defer client.DeleteCachedContent(ctx, cc.Name)

	// You can update the TTL
	newExpireTime := cc.Expiration.ExpireTime.Add(2 * time.Hour)
	_, err = client.UpdateCachedContent(ctx, cc, &genai.CachedContentToUpdate{
		Expiration: &genai.ExpireTimeOrTTL{ExpireTime: newExpireTime}})
	if err != nil {
		log.Fatal(err)
	}
	// [END cache_update]
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
