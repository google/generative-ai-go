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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	"google.golang.org/api/googleapi"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

const (
	defaultModel   = "gemini-1.0-pro"
	modelForVision = "gemini-1.5-flash"
	modelForTools  = "gemini-1.5-pro-latest"
	imageFile      = "personWorkingOnComputer.jpg"
	audioFile      = "sample.mp3"
)

func TestLive(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if testing.Short() {
		t.Skip("skipping live test in -short mode")
	}

	if apiKey == "" {
		t.Skip("set a GEMINI_API_KEY env var to run live tests")
	}

	ctx := context.Background()
	client, err := NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	model := client.GenerativeModel(defaultModel)
	model.Temperature = Ptr[float32](0)

	t.Run("GenerateContent", func(t *testing.T) {
		resp, err := model.GenerateContent(ctx, Text("What is the average size of a swallow?"))
		if err != nil {
			t.Fatal(err)
		}
		got := responseString(resp)
		checkMatch(t, got, `[0-9]+ (cm|centimeters|inches)`)
	})

	t.Run("streaming", func(t *testing.T) {
		iter := model.GenerateContentStream(ctx, Text("Are you hungry?"))
		got := responsesString(t, iter)
		checkMatch(t, got, `(don't|do\s+not|not capable) (have|possess|experiencing) .*(a .* needs|body|sensations|the ability|living)`)
	})

	t.Run("streaming-counting", func(t *testing.T) {
		// Verify only that we don't crash. See #18.
		iter := model.GenerateContentStream(ctx, Text("count 1 to 10."))
		_ = responsesString(t, iter)
	})
	t.Run("streaming-error", func(t *testing.T) {
		iter := model.GenerateContentStream(ctx, ImageData("foo", []byte("bar")))
		_, err := iter.Next()
		if err == nil {
			t.Fatal("got nil, want error")
		}
		var gerr *googleapi.Error
		if !errors.As(err, &gerr) {
			t.Fatalf("does not wrap a googleapi.Error")
		}
	})
	t.Run("chat", func(t *testing.T) {
		session := model.StartChat()

		send := func(msg string, streaming bool) string {
			t.Helper()
			t.Logf("sending %q", msg)
			nh := len(session.History)
			if streaming {
				iter := session.SendMessageStream(ctx, Text(msg))
				for {
					_, err := iter.Next()
					if err == iterator.Done {
						break
					}
					if err != nil {
						t.Fatal(err)
					}
				}
			} else {
				if _, err := session.SendMessage(ctx, Text(msg)); err != nil {
					t.Fatal(err)
				}
			}
			// Check that two items, the sent message and the response) were
			// added to the history.
			if g, w := len(session.History), nh+2; g != w {
				t.Errorf("history length: got %d, want %d", g, w)
			}
			// Last history item is the one we just got from the model.
			return contentString(session.History[len(session.History)-1])
		}

		checkMatch(t,
			send("Name the 5 most popular puppy breeds.", false),
			"Retriever", "Poodle")

		checkMatch(t,
			send("Which is best?", true),
			"best", "depends", "([Cc]onsider|research|compare|preferences)")
	})

	t.Run("image", func(t *testing.T) {
		vmodel := client.GenerativeModel(modelForVision)
		vmodel.Temperature = Ptr[float32](0)

		data, err := os.ReadFile(filepath.Join("testdata", imageFile))
		if err != nil {
			t.Fatal(err)
		}
		resp, err := vmodel.GenerateContent(ctx,
			Text("What is in this picture?"),
			ImageData("jpeg", data))
		if err != nil {
			t.Fatal(err)
		}
		got := responseString(resp)
		checkMatch(t, got, "man|person", "computer|laptop")
	})
	t.Run("audio", func(t *testing.T) {
		vmodel := client.GenerativeModel(modelForVision)
		vmodel.Temperature = Ptr[float32](0)

		data, err := os.ReadFile(filepath.Join("testdata", audioFile))
		if err != nil {
			t.Fatal(err)
		}
		resp, err := vmodel.GenerateContent(ctx,
			Text("What is this sound?"),
			Blob{Data: data, MIMEType: http.DetectContentType(data)})
		if err != nil {
			t.Fatal(err)
		}
		got := responseString(resp)
		checkMatch(t, got, "sounds like")
	})

	t.Run("blocked", func(t *testing.T) {
		t.Skip("skipping until we find a prompt that is blocked")
		// Only happens with streaming at the moment.
		iter := model.GenerateContentStream(ctx, Text("???"))
		resps, err := all(iter)
		if err == nil {
			for _, r := range resps {
				fmt.Println(responseString(r))
			}
			t.Fatal("got nil, want error")
		}
		var berr *BlockedError
		if !errors.As(err, &berr) {
			t.Fatalf("got %v (%[1]T), want BlockedError", err)
		}
		if resps != nil {
			t.Errorf("got responses %v, want nil", resps)
		}
		if berr.PromptFeedback == nil || berr.PromptFeedback.BlockReason != BlockReasonSafety {
			t.Errorf("got PromptFeedback %v, want BlockReasonSafety", berr.PromptFeedback)
		}
		if berr.Candidate != nil {
			t.Fatal("got a candidate, expected nil")
		}
	})
	t.Run("max-tokens", func(t *testing.T) {
		// Verify that setting max output tokens truncates the response.
		// (It does not result in FinishReasonMaxTokens.)
		maxModel := client.GenerativeModel(defaultModel)
		maxModel.Temperature = Ptr(float32(0))
		maxModel.SetMaxOutputTokens(3)
		res, err := maxModel.GenerateContent(ctx, Text("What is a dog?"))
		if err != nil {
			t.Fatal(err)
		}
		if got, want := responseString(res), "A dog is"; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
		gotr := res.Candidates[0].FinishReason
		if gotr != FinishReasonStop && gotr != FinishReasonMaxTokens {
			t.Errorf("got %s, want 'stop' or 'max tokens'", gotr)
		}
	})
	t.Run("max-tokens-streaming", func(t *testing.T) {
		maxModel := client.GenerativeModel(defaultModel)
		maxModel.Temperature = Ptr[float32](0)
		maxModel.MaxOutputTokens = Ptr[int32](3)
		iter := maxModel.GenerateContentStream(ctx, Text("What is a dog?"))
		for {
			_, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				t.Fatal(err)
			}
		}
		res := iter.MergedResponse()
		if got, want := responseString(res), "A dog is"; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
		gotr := res.Candidates[0].FinishReason
		if gotr != FinishReasonStop && gotr != FinishReasonMaxTokens {
			t.Errorf("got %s, want 'stop' or 'max tokens'", gotr)
		}
	})
	t.Run("count-tokens", func(t *testing.T) {
		text := Text("The rain in Spain falls mainly on the plain.")
		res, err := model.CountTokens(ctx, text)
		if err != nil {
			t.Fatal(err)
		}
		if g, w := res.TotalTokens, int32(11); g != w {
			t.Errorf("got %d, want %d", g, w)
		}

		// Should count SystemInstruction tokens too.
		model2 := client.GenerativeModel(defaultModel)
		model2.SystemInstruction = &Content{Parts: []Part{Text("You are a swashbuckling pirate.")}}
		res, err = model2.CountTokens(ctx, text)
		if err != nil {
			t.Fatal(err)
		}
		if g, w := res.TotalTokens, int32(20); g != w {
			t.Errorf("got %d, want %d", g, w)
		}
	})

	t.Run("ReadUsageMetadata", func(t *testing.T) {
		resp, err := model.GenerateContent(ctx, Text("What is the average size of a swallow?"))
		if err != nil {
			t.Fatal(err)
		}
		um := resp.UsageMetadata
		if um.PromptTokenCount < 1 || um.CandidatesTokenCount < 1 || um.TotalTokenCount < 1 {
			t.Errorf("got UsageMetadata=%v, want counts > 0", um)
		}
	})

	t.Run("embed", func(t *testing.T) {
		em := client.EmbeddingModel("embedding-001")
		res, err := em.EmbedContent(ctx, Text("cheddar cheese"))
		if err != nil {
			t.Fatal(err)
		}
		if res == nil || res.Embedding == nil || len(res.Embedding.Values) < 10 {
			t.Errorf("bad result: %v\n", res)
		}

		res, err = em.EmbedContentWithTitle(ctx, "My Cheese Report", Text("I love cheddar cheese."))
		if err != nil {
			t.Fatal(err)
		}
		if res == nil || res.Embedding == nil || len(res.Embedding.Values) < 10 {
			t.Errorf("bad result: %v", res)
		}
	})
	t.Run("batch-embed", func(t *testing.T) {
		em := client.EmbeddingModel("embedding-001")
		b := em.NewBatch().
			AddContent(Text("cheddar cheese")).
			AddContentWithTitle("My Cheese Report", Text("I love cheddar cheese."))
		res, err := em.BatchEmbedContents(ctx, b)
		if err != nil {
			t.Fatal(err)
		}
		if res == nil || len(res.Embeddings) != 2 {
			t.Errorf("bad result: %v", res)
		}
	})

	t.Run("list-models", func(t *testing.T) {
		iter := client.ListModels(ctx)
		var got []*ModelInfo
		for {
			m, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				t.Fatal(err)
			}
			got = append(got, m)
		}

		for _, name := range []string{"gemini-1.0-pro", "embedding-001"} {
			has := false
			fullName := "models/" + name
			for _, m := range got {
				if m.Name == fullName {
					has = true
					break
				}
			}
			if !has {
				t.Errorf("missing model %q", name)
			}
		}
	})
	t.Run("get-model", func(t *testing.T) {
		modName := defaultModel
		got, err := client.GenerativeModel(modName).Info(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if w := "models/" + modName; got.Name != w {
			t.Errorf("got name %q, want %q", got.Name, w)
		}

		modName = "embedding-001"
		got, err = client.EmbeddingModel(modName).Info(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if w := "models/" + modName; got.Name != w {
			t.Errorf("got name %q, want %q", got.Name, w)
		}
	})

	t.Run("tools", func(t *testing.T) {
		movieChat := func(t *testing.T, s *Schema, fcm FunctionCallingMode) {
			movieTool := &Tool{
				FunctionDeclarations: []*FunctionDeclaration{{
					Name:        "find_theaters",
					Description: "find theaters based on location and optionally movie title which is currently playing in theaters",
					Parameters:  s,
				}},
			}
			model := client.GenerativeModel(modelForTools)
			model.SetTemperature(0)
			model.Tools = []*Tool{movieTool}
			model.ToolConfig = &ToolConfig{
				FunctionCallingConfig: &FunctionCallingConfig{
					Mode: fcm,
				},
			}
			session := model.StartChat()
			res, err := session.SendMessage(ctx, Text("Which theaters in Mountain View show Barbie movie?"))
			if err != nil {
				t.Fatal(err)
			}
			funcalls := res.Candidates[0].FunctionCalls()
			if fcm == FunctionCallingNone {
				if len(funcalls) != 0 {
					t.Fatalf("got %d FunctionCalls, want 0", len(funcalls))
				}
				return
			}
			if len(funcalls) != 1 {
				t.Fatalf("got %d FunctionCalls, want 1", len(funcalls))
			}
			funcall := funcalls[0]
			if g, w := funcall.Name, movieTool.FunctionDeclarations[0].Name; g != w {
				t.Errorf("FunctionCall.Name: got %q, want %q", g, w)
			}
			locArg, ok := funcall.Args["location"].(string)
			if !ok {
				t.Fatal(`funcall.Args["location"] is not a string`)
			}
			if c := "Mountain View"; !strings.Contains(locArg, c) {
				t.Errorf(`FunctionCall.Args["location"]: got %q, want string containing %q`, locArg, c)
			}
			res, err = session.SendMessage(ctx, FunctionResponse{
				Name: movieTool.FunctionDeclarations[0].Name,
				Response: map[string]any{
					"theater": "AMC16",
				},
			})
			if err != nil {
				t.Fatal(err)
			}
			checkMatch(t, responseString(res), "AMC")
		}
		schema := &Schema{
			Type: TypeObject,
			Properties: map[string]*Schema{
				"location": {
					Type:        TypeString,
					Description: "The city and state, e.g. San Francisco, CA or a zip code e.g. 95616",
				},
				"title": {
					Type:        TypeString,
					Description: "Any movie title",
				},
			},
			Required: []string{"location"},
		}
		t.Run("direct", func(t *testing.T) {
			movieChat(t, schema, FunctionCallingAuto)
		})
		t.Run("none", func(t *testing.T) {
			movieChat(t, schema, FunctionCallingNone)
		})
	})

	t.Run("files", func(t *testing.T) {
		file := uploadFile(t, ctx, client, filepath.Join("testdata", imageFile))
		// Sanity checks on the returned file.
		if !strings.HasPrefix(file.Name, "files/") {
			t.Fatalf("got %q, want file name beginning 'files/'", file.Name)
		}
		if got, want := file.SizeBytes, int64(9218); got != want {
			t.Errorf("got file size %d, want %d", got, want)
		}
		if got, want := file.MIMEType, "image/jpeg"; got != want {
			t.Errorf("got MIME type %q, want %q", got, want)
		}
		// The file should have just been created. Be generous, though.
		now := time.Now()
		if file.CreateTime.Before(now.Add(-time.Hour)) || file.CreateTime.After(now.Add(time.Hour)) {
			t.Errorf("got file time %s, wanted around now (%s)", file.CreateTime, now)
		}

		// Don't test GetFile, because UploadFile already calls GetFile.
		// ListFiles should return the file we just uploaded, and maybe other files too.
		iter := client.ListFiles(ctx)
		found := false
		for {
			ifile, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				t.Fatal(err)
			}
			if ifile.Name == file.Name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("ListFiles did not return the uploaded file %s", file.Name)
		}

		// Use the uploaded file to generate content.
		model := client.GenerativeModel("gemini-1.5-pro-latest")
		resp, err := model.GenerateContent(ctx, FileData{URI: file.URI})
		if err != nil {
			t.Fatal(err)
		}
		checkMatch(t, responseString(resp), "person", "computer|laptop")

		// Test MIME detection for mp3 files.
		file = uploadFile(t, ctx, client, filepath.Join("testdata", audioFile))
		if got, want := file.MIMEType, "audio/mpeg"; got != want {
			t.Errorf("got MIME type %q, want %q", got, want)
		}
		resp, err = model.GenerateContent(ctx, FileData{URI: file.URI})
		if err != nil {
			t.Fatal(err)
		}
		// We don't care about the response to the audio file, only that it doesn't fail.

		t.Run("metadata", func(t *testing.T) {
			f, err := os.Open(filepath.Join("testdata", "earth.mp4"))
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()
			file, err := client.UploadFile(ctx, "", f, nil)
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("uploaded %s, MIME type %q", file.Name, file.MIMEType)
			defer func() {
				// Delete the file when the test is done.
				if err := client.DeleteFile(ctx, file.Name); err != nil {
					t.Fatal(err)
				}
			}()
			// TODO(jba): verify metadata when it is populated.
			t.Logf("Metadata: %+v\n", file.Metadata)
		})
	})

	t.Run("JSON", func(t *testing.T) {
		model := client.GenerativeModel("gemini-1.5-pro-latest")
		model.SetTemperature(0)
		model.ResponseMIMEType = "application/json"
		res, err := model.GenerateContent(ctx, Text("List the primary colors."))
		if err != nil {
			t.Fatal(err)
		}
		got := responseString(res)
		t.Logf("got %s", got)
		// Accept any valid JSON
		var a any
		if err := json.Unmarshal([]byte(got), &a); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("JSON schema", func(t *testing.T) {
		model := client.GenerativeModel("gemini-1.5-pro-latest")
		model.SetTemperature(0)
		model.ResponseMIMEType = "application/json"
		model.ResponseSchema = &Schema{
			Type: TypeArray,
			Items: &Schema{
				Type: TypeObject,
				Properties: map[string]*Schema{
					"name": {
						Type:        TypeString,
						Description: "The name of the color",
					},
					"RGB": {
						Type:        TypeString,
						Description: "The RGB value of the color, in hex",
					},
				},
				Required: []string{"name", "RGB"},
			},
		}
		res, err := model.GenerateContent(ctx, Text("List the primary colors."))
		if err != nil {
			t.Fatal(err)
		}
		got := responseString(res)
		t.Logf("got %s", got)
		// Check that the format of the result matches the schema. The actual content
		// doesn't matter here.

		type color struct {
			Name, RGB string
		}
		var v []color
		if err := json.Unmarshal([]byte(got), &v); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("caching", func(t *testing.T) { testCaching(t, client) })
	t.Run("code-execution", func(t *testing.T) {
		model := client.GenerativeModel("gemini-1.5-flash")
		model.SetTemperature(0)
		model.Tools = []*Tool{{CodeExecution: &CodeExecution{}}}
		res, err := model.GenerateContent(ctx, Text(`
			788477675 * 778 = x.  Find x and also compute largest odd number smaller than this number.
			`))
		if err != nil {
			t.Fatal(err)
		}
		// Look for the right types of parts.
		if len(res.Candidates) == 0 {
			t.Fatal("no candidates")
		}
		seen := map[reflect.Type]bool{}
		for _, p := range res.Candidates[0].Content.Parts {
			seen[reflect.TypeOf(p)] = true
		}
		if !seen[reflect.TypeOf(&ExecutableCode{})] || !seen[reflect.TypeOf(&CodeExecutionResult{})] {
			t.Error("missing ExecutableCode or CodeExecutionResult")
		}
	})
}

func TestJoinResponses(t *testing.T) {
	r1 := &GenerateContentResponse{
		Candidates: []*Candidate{
			{
				Index:        2,
				Content:      &Content{Role: roleModel, Parts: []Part{Text("r1 i2")}},
				FinishReason: FinishReason(1),
			},
			{
				Index:        0,
				Content:      &Content{Role: roleModel, Parts: []Part{Text("r1 i0")}},
				FinishReason: FinishReason(2),
			},
		},
		PromptFeedback: &PromptFeedback{BlockReason: BlockReasonSafety},
	}
	r2 := &GenerateContentResponse{
		Candidates: []*Candidate{
			{
				Index:        0,
				Content:      &Content{Role: roleModel, Parts: []Part{Text(";r2 i0")}},
				FinishReason: FinishReason(3),
			},
			{
				// ignored
				Index:        1,
				Content:      &Content{Role: roleModel, Parts: []Part{Text(";r2 i1")}},
				FinishReason: FinishReason(4),
			},
		},

		PromptFeedback: &PromptFeedback{BlockReason: BlockReasonOther},
	}
	got := joinResponses(r1, r2)
	want := &GenerateContentResponse{
		Candidates: []*Candidate{
			{
				Index:        2,
				Content:      &Content{Role: roleModel, Parts: []Part{Text("r1 i2")}},
				FinishReason: FinishReason(1),
			},
			{
				Index:        0,
				Content:      &Content{Role: roleModel, Parts: []Part{Text("r1 i0;r2 i0")}},
				FinishReason: FinishReason(3),
			},
		},
		PromptFeedback: &PromptFeedback{BlockReason: BlockReasonSafety},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("\ngot %+v\nwant %+v", got, want)
	}
}

func TestMergeTexts(t *testing.T) {
	for _, test := range []struct {
		in   []Part
		want []Part
	}{
		{
			in:   []Part{Text("a")},
			want: []Part{Text("a")},
		},
		{
			in:   []Part{Text("a"), Text("b"), Text("c")},
			want: []Part{Text("abc")},
		},
		{
			in:   []Part{Blob{"b1", nil}, Text("a"), Text("b"), Blob{"b2", nil}, Text("c")},
			want: []Part{Blob{"b1", nil}, Text("ab"), Blob{"b2", nil}, Text("c")},
		},
		{
			in:   []Part{Text("a"), Text("b"), Blob{"b1", nil}, Text("c"), Text("d"), Blob{"b2", nil}},
			want: []Part{Text("ab"), Blob{"b1", nil}, Text("cd"), Blob{"b2", nil}},
		},
	} {
		got := mergeTexts(test.in)
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("%+v:\ngot  %+v\nwant %+v", test.in, got, test.want)
		}
	}
}

func checkMatch(t *testing.T, got string, wants ...string) {
	t.Helper()
	for _, want := range wants {
		re, err := regexp.Compile("(?i:" + want + ")")
		if err != nil {
			t.Fatal(err)
		}
		if !re.MatchString(got) {
			t.Errorf("\ngot %q\nwanted to match %q", got, want)
		}
	}
}

func responseString(resp *GenerateContentResponse) string {
	var b strings.Builder
	for i, cand := range resp.Candidates {
		if len(resp.Candidates) > 1 {
			fmt.Fprintf(&b, "%d:", i+1)
		}
		b.WriteString(contentString(cand.Content))
	}
	return b.String()
}

func contentString(c *Content) string {
	var b strings.Builder
	if c == nil || c.Parts == nil {
		return ""
	}
	for i, part := range c.Parts {
		if i > 0 {
			fmt.Fprintf(&b, ";")
		}
		fmt.Fprintf(&b, "%v", part)
	}
	return b.String()
}

func responsesString(t *testing.T, iter *GenerateContentResponseIterator) string {
	var lines []string
	for {
		resp, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		lines = append(lines, responseString(resp))
	}
	return strings.Join(lines, "\n")
}

func all(iter *GenerateContentResponseIterator) ([]*GenerateContentResponse, error) {
	var rs []*GenerateContentResponse
	for {
		r, err := iter.Next()
		if err == iterator.Done {
			return rs, nil
		}
		if err != nil {
			return nil, err
		}
		rs = append(rs, r)
	}
}

func dump(w io.Writer, x any) {
	var err error
	printf := func(format string, args ...any) {
		if err == nil {
			_, err = fmt.Fprintf(w, format, args...)
		}
	}
	printValue(reflect.ValueOf(x), "", "", printf)
	if err != nil {
		log.Fatal(err)
	}
}

func printValue(v reflect.Value, indent, first string, printf func(string, ...any)) {
	switch v.Kind() {
	case reflect.Slice, reflect.Array, reflect.Map, reflect.Struct:
		printf("%s%s%s{\n", indent, first, v.Type())
		indent1 := indent + "    "
		switch v.Kind() {
		case reflect.Slice, reflect.Array:
			for i := 0; i < v.Len(); i++ {
				printValue(v.Index(i), indent1, fmt.Sprintf("[%d]: ", i), printf)
			}
		case reflect.Map:
			iter := v.MapRange()
			for iter.Next() {
				printValue(iter.Value(), indent1, fmt.Sprintf("%q: ", iter.Key()), printf)
			}
		case reflect.Struct:
			for _, sf := range reflect.VisibleFields(v.Type()) {
				vf := v.FieldByName(sf.Name)
				if !vf.IsZero() {
					printValue(vf, indent1, sf.Name+": ", printf)
				}
			}
		default:
			panic("unhandled default case")
		}
		printf("%s}\n", indent)
	case reflect.Pointer, reflect.Interface:
		printValue(v.Elem(), indent, first, printf)
	case reflect.String:
		printf("%s%s%q\n", indent, first, v)
	default:
		printf("%s%s%v\n", indent, first, v)
	}
}

func TestMatchString(t *testing.T) {
	for _, test := range []struct {
		re, in string
	}{
		{"do not", "I do not have"},
		{"(don't|do not) have", "I do not have"},
		{"(don't|do not) have", "As an AI language model, I do not have physical needs"},
	} {
		re := regexp.MustCompile(test.re)
		if !re.MatchString(test.in) {
			t.Errorf("%q doesn't match %q", test.re, test.in)
		}
	}
}

func TestNoAPIKey(t *testing.T) {
	_, err := NewClient(context.Background())
	if err == nil {
		t.Fatal("got nil, want error")
	}
	_, err = NewClient(context.Background(), option.WithAPIKey(""))
	if err == nil {
		t.Fatal("got nil, want error")
	}
}

func TestRecoverPanic(t *testing.T) {
	// Verify that conversions that used to cause a panic now result in an error.
	fr := &FunctionResponse{
		Name:     "r",
		Response: map[string]any{"x": 1 + 2i}, // complex values are invalid
	}
	var m GenerativeModel
	_, err := m.newGenerateContentRequest(NewUserContent(fr))
	if err == nil {
		t.Fatal("got nil, want error")
	}
}

type customRT struct {
	APIKey string
}

func (t *customRT) RoundTrip(req *http.Request) (*http.Response, error) {
	transport := http.DefaultTransport.(*http.Transport).Clone()
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

func TestCustomHTTPClient(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if testing.Short() || apiKey == "" {
		t.Skip("skipping live test in -short mode, or when API key isn't provided")
	}
	c := &http.Client{
		Transport: &customRT{APIKey: apiKey},
	}

	ctx := context.Background()
	client, err := NewClient(ctx, option.WithHTTPClient(c))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	model := client.GenerativeModel(defaultModel)
	resp, err := model.GenerateContent(ctx, Text("What are some of the largest cities in the US?"))
	if err != nil {
		t.Fatal(err)
	}
	got := responseString(resp)
	checkMatch(t, got, `new york`)
}

func uploadFile(t *testing.T, ctx context.Context, client *Client, filename string) *File {
	osf, err := os.Open(filename)
	if err != nil {
		t.Fatal(err)
	}
	defer osf.Close()
	// Upload a file. Using the empty string as a name will generate a unique name.
	file, err := client.UploadFile(ctx, "", osf, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("uploaded %s, MIME type %q", file.Name, file.MIMEType)
	for file.State == FileStateProcessing {
		t.Logf("waiting for uploaded file to become active")
		time.Sleep(time.Second)
		var err error
		file, err = client.GetFile(ctx, file.Name)
		if err != nil {
			t.Fatal(err)
		}
	}
	if file.State != FileStateActive {
		t.Fatalf("uploaded file has state %s, not FileStateActive", file.State)
	}
	t.Cleanup(func() {
		// Delete the file when the test is done.
		if err := client.DeleteFile(ctx, file.Name); err != nil {
			t.Fatal(err)
		}
	})
	return file
}

func TestAttribution(t *testing.T) {
	a := WithClientInfo("k", "v")
	opts := []option.ClientOption{option.WithAPIKey("x"), a, option.WithEndpoint("e")}
	got, ok := optionOfType[*clientInfo](opts)
	if !ok {
		t.Fatal("not found")
	}
	want := a.(*clientInfo)
	if got.key != want.key || got.value != want.value {
		t.Errorf("got %q, %q, want %q, %q", got.key, got.value, want.key, want.value)
	}
}
