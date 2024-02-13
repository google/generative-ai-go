# How to Contribute

We would love to accept your patches and contributions to this project.

## Before you begin

### Sign our Contributor License Agreement

Contributions to this project must be accompanied by a
[Contributor License Agreement](https://cla.developers.google.com/about) (CLA).
You (or your employer) retain the copyright to your contribution; this simply
gives us permission to use and redistribute your contributions as part of the
project.

If you or your current employer have already signed the Google CLA (even if it
was for a different project), you probably don't need to do it again.

Visit <https://cla.developers.google.com/> to see your current agreements or to
sign a new one.

### Review our Community Guidelines

This project follows [Google's Open Source Community
Guidelines](https://opensource.google/conduct/).

## Contribution process

1. Clone this repo
2. Run tests with `go test ./...`
3. You may need to run "live" tests that talk to a real endpoint; to do so, run
   `go test -v ./genai/...` passing it your API key with the `-apikey` flag
   and a model name flag like `-model gemini-1.0-pro`

### Code Reviews

All submissions, including submissions by project members, require review. We
use [GitHub pull requests](https://docs.github.com/articles/about-pull-requests)
for this purpose.
