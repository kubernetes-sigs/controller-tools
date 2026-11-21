# Contributing guidelines

## Sign the CLA

Kubernetes projects require that you sign a Contributor License Agreement (CLA) before we can accept your pull requests.  Please see https://git.k8s.io/community/CLA.md for more info

### Contributing A Patch

1. Submit an issue describing your proposed change to the repo in question.
1. The [repo owners](OWNERS) will respond to your issue promptly.
1. If your proposed change is accepted, and you haven't already done so, sign a Contributor License Agreement (see details above).
1. Fork the desired repo, develop and test your code changes.
1. Submit a pull request.

## Consistency

Markers and generators can be used together in the same project, so they should behave consistently:

- Options and markers supported by one generator should be supported by others where they apply.
- The same option or marker should behave the same way across generators. Users should not get different results from the same configuration.
- Use the same conventions for the same things across generators.
- Treat reported inconsistencies as bugs to fix.

## Feature lifecycle

- A feature stays supported until it is deprecated.
- A deprecation comes with a practical migration path. Users should not have to replace existing functionality by hand.
- A deprecated feature is removed no earlier than three minor releases later.
- Before deprecating or removing a feature, check how it is used across the community and ecosystem.

**Exception.** controller-tools cannot keep supporting a feature that its dependencies no longer support, such as the [compatible `k8s.io/*`, client-go and Go versions](README.md#compatibility), or the libraries it is built on (e.g. `golang.org/x/tools`, `k8s.io/code-generator` and `k8s.io/gengo`). In that case the feature can be deprecated or removed sooner, with the reason stated in the pull request.

**Experimental features** (documented as experimental in their marker help or release notes) are not supported yet, so they can change, be deprecated or be removed sooner.

**Example.** Deprecating `+kubebuilder:validation:Optional` in favour of `+optional`, while keeping both working, is a good migration path. Removing a widely adopted marker and asking users to "do it by hand" is not.

**Note** This adds a deprecation period on top of [VERSIONING.md](VERSIONING.md), which allows breaking changes in any 0.x minor release.

## Discussing changes that affect end users

Changes that affect end users, such as new generators, deprecations, removals or changes in behaviour, should be discussed in public and tracked in this repository, so that consumers and end users can give their input before the change is merged:

- Open a GitHub issue in this repository that describes the change, its impact and the migration path, and link it from the pull request.
- For larger changes, or changes with a large impact on consumers, write a design document, following the approach of [Kubernetes Enhancement Proposals (KEPs)][keps].
- Share the issue in the community channels listed in the [README](README.md).

## Marker documentation

controller-tools [generates marker help][marker-help] from Go doc comments. Write help text for end users, and explain how and when to use the marker, with examples.

The help can be checked via `controller-gen -wwww ...` and in the Go package documentation.

For example, the [RBAC marker][rbac-src] defines its help text in Go (shortened):

```go
// Rule specifies an RBAC rule to all access to some resources or non-resource URLs.
//
// RBAC markers are used to generate ClusterRole or Role manifests.
// Multiple markers can be combined to build comprehensive RBAC policies.
//
// Examples:
//
//	// Basic resource access
//	// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch
//
//	// Core API group (use empty string)
//	// +kubebuilder:rbac:groups="",resources=pods;services,verbs=get;list;watch
//
//	// ...
type Rule struct {
	// ...

	// Verbs specifies the (lowercase) kubernetes API verbs that this rule encompasses.
	// Common verbs: "get", "list", "watch", "create", "update", "patch", "delete".
	// Use "*" for all verbs.
	// Multiple verbs must be specified separated by semicolons.
	// Example: "get;list;watch".
	Verbs []string

	// ...
}
```

These comments appear in the [controller-tools Go documentation][rbac-go-docs] and are used to generate the marker help.

**Note.** [Kubebuilder][kubebuilder] also uses the generated help through its [`markerdocs`][markerdocs] tool to build the marker reference in the [Kubebuilder book][markers-book].

## Coordinating changes with Kubebuilder

controller-tools was created in 2018 as ["a subproject of the kubebuilder project"][ct-readme-2018], around the time [Kubebuilder][kubebuilder] was introduced as ["an SDK for building Kubernetes APIs"][kb-blog-2018]. Today, many users still get `controller-gen` through projects scaffolded with Kubebuilder, so changes with a real impact on them should be coordinated with Kubebuilder maintainers.

Deprecations, removals or changes in behaviour that can have a real impact on users (in markers, generator options or generated output) should involve [Kubebuilder][kubebuilder] maintainers in the PR.

For new generators, ping [Kubebuilder][kubebuilder] maintainers early. They can try the generator in a proof of concept with a Kubebuilder project, check how it fits the ecosystem and how the community uses it, and then document it or add it to the scaffold. Their review counts like any other reviewer's, and like any contributor they can ask for changes.

Use the correct icon in the PR title (see the [versioning guidelines](https://sigs.k8s.io/kubebuilder-release-tools/VERSIONING.md)) and describe the impact in the PR description.

[marker-help]: https://pkg.go.dev/sigs.k8s.io/controller-tools/pkg/markers#hdr-Help
[rbac-go-docs]: https://pkg.go.dev/sigs.k8s.io/controller-tools/pkg/rbac#Rule
[rbac-src]: https://github.com/kubernetes-sigs/controller-tools/blob/v0.23.0/pkg/rbac/parser.go#L42-L101
[markers-book]: https://book.kubebuilder.io/reference/markers.html
[markerdocs]: https://github.com/kubernetes-sigs/kubebuilder/blob/master/docs/book/utils/markerdocs/main.go
[kubebuilder]: https://github.com/kubernetes-sigs/kubebuilder
[keps]: https://github.com/kubernetes/enhancements/tree/master/keps
[ct-readme-2018]: https://github.com/kubernetes-sigs/controller-tools/blob/95af6dc11e7aeb7cf646221db54b3b2fe1e195d3/README.md
[kb-blog-2018]: https://kubernetes.io/blog/2018/08/10/introducing-kubebuilder-an-sdk-for-building-kubernetes-apis-using-crds/
