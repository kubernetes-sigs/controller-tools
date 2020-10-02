# Versioning and Releasing in controller-tools

We follow the [common KubeBuilder versioning guidelines][guidelines], and
use the corresponding tooling.

For the purposes of the aforementioned guidelines, controller-tools counts
as a "library project", but otherwise follows the guidelines closely.

[guidelines]: https://sigs.k8s.io/kubebuilder-release-tools/VERSIONING.md

## Compatibility and Release Support

For release branches, we generally do not support older branches.  This
may change in the future.

Compability-wise, remember that changes to generation defaults are
breaking changes.

## Updates to Other Projects on Release

When you release, you'll need to perform updates in other KubeBuilder
projects:

- Update the controller-tools version used to generate the KubeBuilder
  book at [docs/book/install-and-build.sh][book-script]

[book-script]: https://sigs.k8s.io/kubebuilder/docs/book/install-and-build.sh
