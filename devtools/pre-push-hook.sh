#!/bin/sh

# This script checks that the version in the code matches the latest version
# tag.
#
# Install as a pre-push hook from the repo root with:
#   cp devtools/pre-push-hook.sh .git/hooks/pre-push

version_file=genai/internal/version.go
latest_tag=$(git tag -l 'v*' | sort -V | tail -1)
code_version=v$(awk '/^const Version/ {print substr($4, 2, length($4)-2)}' $version_file)

if [[ $latest_tag == $code_version ]]; then
  exit 0
fi

echo "version $code_version in $version_file does not match latest tag $latest_tag."
exit 1
