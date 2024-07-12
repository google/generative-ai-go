// Copyright 2024 Google LLC
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

// This code generator takes examples from the internal/samples directory
// and copies them to "official" examples in genai/example_test.go, while
// removing snippet comments (between [START...] and [END...]) that are used
// for website documentation purposes.
// It's invoked with a go:generate directive in the source file.

package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strings"
)

func main() {
	inPath := flag.String("in", "", "input file path")
	outPath := flag.String("out", "", "output file path")
	flag.Parse()

	if len(*inPath) == 0 || len(*outPath) == 0 {
		log.Fatalf("got empty -in (%v) or -out (%v)", *inPath, *outPath)
	}

	inFile, err := os.Open(*inPath)
	if err != nil {
		log.Fatal(err)
	}
	defer inFile.Close()

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, *inPath, inFile, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	for _, cgroup := range file.Comments {
		sanitizeCommentGroup(cgroup)
	}

	outFile, err := os.Create(*outPath)
	if err != nil {
		log.Fatal(err)
	}
	defer outFile.Close()

	fmt.Fprintln(outFile, strings.TrimLeft(preamble, "\r\n"))
	format.Node(outFile, fset, file)
}

const preamble = `
// This file was generated from internal/samples/docs-snippets_test.go. DO NOT EDIT.
`

func printCommentGroup(cg *ast.CommentGroup) {
	fmt.Printf("-- comment group %p\n", cg)
	for _, c := range cg.List {
		fmt.Println(c.Slash, c.Text)
	}
}

// sanitizeCommentGroup removes comment blocks between [START... and [END...
// (including these lines), and also any go:generate directives - it modifies cg.
func sanitizeCommentGroup(cg *ast.CommentGroup) {
	var nl []*ast.Comment
	excludeBlock := false
	for _, commentLine := range cg.List {
		if strings.Contains(commentLine.Text, "[START") {
			excludeBlock = true
		} else if strings.Contains(commentLine.Text, "[END") {
			excludeBlock = false
		} else if !excludeBlock {

			if !strings.Contains(commentLine.Text, "go:generate") {
				nl = append(nl, commentLine)
			}
		}
	}
	cg.List = nl
}
