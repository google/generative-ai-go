#!/bin/bash -e
# Copyright 2024 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# This script generates the "discovery" client for the GenerativeLanguage API.
# It is needed for file upload, which GAPIC clients don't support.

# Run this tool from the `genai` directory of this repository.

# The repo github.com/googleapis/google-api-go-client (corresponding to the Go import
# path google.golang.org/api) contains a program that generates a Go client from
# a discovery doc. It also contains all the clients generated from public discovery
# docs, but the generativelanguage doc isn't public. In fact, retrieving it requires
# an API key. We also don't want to put the discovery client in that repo, because
# we don't want it to be public either; that would only confuse users.


if [[ $GEMINI_API_KEY = '' ]]; then
  echo >&2 "need to set GEMINI_API_KEY at https://aistudio.google.com"
  exit 1
fi

# Install the code generator for discovery clients.
go install google.golang.org/api/google-api-go-generator@latest

# Download the discovery document.
docfile=/tmp/gl.json
curl -s 'https://generativelanguage.googleapis.com/$discovery/rest?version=v1beta&key='$GEMINI_API_KEY > $docfile

# Generate the client. Write it to the internal directory to it is not exposed to users.
google-api-go-generator -api_json_file $docfile \
  -gendir internal \
  -internal_pkg github.com/google/generative-ai-go/genai/internal \
  -gensupport_pkg github.com/google/generative-ai-go/genai/internal/gensupport

# Replace license with the proper one for this repo.
file=internal/generativelanguage/v1beta/generativelanguage-gen.go
cat license.txt <(tail +5 $file) | sponge $file

#Send Notification
echo "Succesfully generated discovery client!"

