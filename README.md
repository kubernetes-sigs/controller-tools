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

controller-tools is a subproject of the [kubebuilder](https://sigs.k8s.io/kubebuilder) project
in sig apimachinery.

You can reach the maintainers of this project at:

- Slack channel: [#kubebuilder](http://slack.k8s.io/#kubebuilder)
- Google Group: [kubebuilder@googlegroups.com](https://groups.google.com/forum/#!forum/kubebuilder)

### Code of conduct

Participation in the Kubernetes community is governed by the [Kubernetes Code of Conduct](code-of-conduct.md).
