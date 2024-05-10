#!/bin/bash -e

# This script generates the "discovery" client for the GenerativeLanguage API.
# It is needed for file upload, which GAPIC clients don't support.

# The repo github.com/googleapis/google-api-go-client (corresponding to the Go import
# path google.golang.org/api) contains a program that generates a Go client from
# a discovery doc. It also contains all the clients generated from public discovery
# docs, but the generativelanguage doc isn't public. In fact, retrieving it requires
# an API key. We also don't want to put the discovery client in that repo, because
# we don't want it to be public either; that would only confuse users.

if [[ $GEMINI_API_KEY = '' ]]; then
  echo >&2 "need to set GEMINI_API_KEY"
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

