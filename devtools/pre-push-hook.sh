#!/bin/sh -e
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

# This script performs some checks.
# Install as a pre-push hook from the repo root with:
#   cp devtools/pre-push-hook.sh .git/hooks/pre-push

go test -short ./...
go vet ./...

# Check that the version in the code matches the latest version tag.
version_file=genai/internal/version.go
latest_tag=$(git tag -l 'v*' | sort -V | tail -1)
code_version=v$(awk '/^const Version/ {print substr($4, 2, length($4)-2)}' $version_file)

if [[ $latest_tag == $code_version ]]; then
  exit 0
fi

echo "version $code_version in $version_file does not match latest tag $latest_tag."
exit 1
