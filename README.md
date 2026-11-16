[![Go Reference](https://pkg.go.dev/badge/sigs.k8s.io/controller-tools.svg)](https://pkg.go.dev/sigs.k8s.io/controller-tools)
[![Go Report Card](https://goreportcard.com/badge/sigs.k8s.io/controller-tools)](https://goreportcard.com/report/sigs.k8s.io/controller-tools)

# Kubernetes controller-tools Project

The Kubernetes controller-tools Project is a set of go libraries for building Controllers.

## Development

Clone this project, and iterate on changes by running `./test.sh`.

## Releasing and Versioning

See [VERSIONING.md](VERSIONING.md).

## Compatibility

Every minor version of controller-tools (CT) has been tested with a specific minor version of client-go. A controller-tools minor version *may* be compatible with
other client-go minor versions, but this is by chance and neither supported nor tested. In general, we create one minor version of controller-tools
for each minor version of client-go and other k8s.io/* dependencies.

The minimum Go version of controller-tools is the highest minimum Go version of our Go dependencies. Usually, this will
be identical to the minimum Go version of the corresponding k8s.io/* dependencies.

Compatible k8s.io/*, client-go and minimum Go versions can be looked up in our [go.mod](go.mod) file.

|          | k8s.io/*, client-go | minimum Go version |
|----------|:-------------------:|:------------------:|
| CT v0.21 |        v0.36        |        1.26        |
| CT v0.20 |        v0.35        |        1.25        |
| CT v0.19 |        v0.34        |        1.24        |

## Community, discussion, contribution, and support

Learn how to engage with the Kubernetes community on the [community page](http://kubernetes.io/community/).

controller-tools is a subproject of [sig apimachinery](https://github.com/kubernetes/community/tree/main/sig-api-machinery#server-sdk).

You can reach the maintainers of this project at:

- Slack channel: [#controller-runtime](http://slack.k8s.io/#controller-runtime)

### Code of conduct

Participation in the Kubernetes community is governed by the [Kubernetes Code of Conduct](code-of-conduct.md).

### AI Guidance

This project follows the [Kubernetes AI guidance][Kubernetes AI guidance]:

> Using AI tools to help write your PR is acceptable, but as the author, you are responsible for understanding every change. If you used AI tools in preparing your PR, you must disclose this in the description of your PR. For example, including “This PR was written in part with the assistance of generative AI,” in the PR description is sufficient.
> 
> Listing AI tooling as a co-author, co-signing commits using an AI tool, or using the assisted-by, co-developed or similar commit trailer is not allowed.
>
> All contributions must follow the [contributions policies] and use commit messages that align with [the policy].
>
> [Large AI generated PRs] and AI generated commit messages are not allowed.
>
> Do not leave the first review of AI generated changes to the reviewers. Verify the changes (code review, testing, etc.) before submitting your PR. Reviewers may ask questions about your AI-assisted code, and if you cannot explain why a change was made, the PR will be closed.
>
> When responding to review comments, you must do so without relying on AI tools. Reviewers want to engage directly with you, not with generated responses. If you do not engage directly with reviewers, the PR will be closed.

[Kubernetes AI guidance]: https://www.kubernetes.dev/docs/guide/pull-requests/#ai-guidance
[contributions policies]: https://www.kubernetes.dev/docs/guide/contributing/
[the policy]: https://www.kubernetes.dev/docs/guide/pull-requests/#commit-message-guidelines
[Large AI generated PRs]: https://www.kubernetes.dev/docs/guide/pull-requests/#large-or-automatic-edits
