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
2. Run tests with `go test ./...`; the "live" tests will be skipped
   unless a valid API key is set with the `GEMINI_API_KEY` environment variable.

### Code Reviews

All submissions, including submissions by project members, require review. We
use [GitHub pull requests](https://docs.github.com/articles/about-pull-requests)
for this purpose.

## For Maintainers

### Preparation

Install the pre-push hook:
```
cp devtools/pre-push-hook.sh .git/hooks/pre-push
```

### Creating a new release

This repo consists of a single Go module.
To increase the minor or patch version of the module:

1. Determine the desired tag, using `git tag -l` to see existing tags
   and incrementing as appropriate. We will call the result TAG in
   these instructions. It should be of the form `vX.Y.Z`.
2. Update the version in genai/internal/version.go to match TAG.
3. Send a PR with that change. The pre-push hook should complain, so
   pass the `--no-verify` flag to `git push`.
4. Submit the PR when approved. _No other PRs should be submitted until
   the following steps have been completed._
5. Run `git pull` to get the submitted PR locally. You should be on main.
6. Run `git tag TAG` to tag the repo locally.
7. Run `git push origin TAG`. If the pre-push hook complains here, something
   is wrong; stop and review.
8. Use the [GitHub UI](https://github.com/google/generative-ai-go/releases) to
   create the release. Use TAG as the name.
   Provide release notes by summarizing the result of `git log PREVTAG..`,
   where PREVTAG is the previous release tag.
9. Visit https://pkg.go.dev/github.com/google/generative-ai-go@TAG and request
   that the version be processed.
