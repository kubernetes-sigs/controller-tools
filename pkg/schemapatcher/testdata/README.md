# CRD Schema Patcher Integration Test testdata

This contains a tiny module used for testdata for the CRD Schema Patcher
integration test.  The directory should always be called testdata, so Go
treats it specially.

The types in `apis/<somepkg>` represent different CRD schema versions that
are either expected to change or not change.

The `kubebuilder` API group contains types that look like modern
KubeBuilder-generated types, while the `legacy` API group contains types
that look like core k8s/legacy kubebuilder types.

The `manifests` directory contains input manifests that will be patched,
while `expected` contains expected output from the patching process.

It's *highly* unlikely that the generated expected manifests will
ever change from these -- if they do, you've probably broken something.
Nonetheless, you can regenerate output using `make`.

Make sure you review the diff to ensure that it only contains the desired
changes!

If you didn't add a new marker and this output changes, make sure you have
a good explanation for why generated output needs to change!
