# CRD Integration Test testdata

This contains a tiny module used for testdata for the CRD integration
test.  The directory should always be called testdata, so Go treats it
specially.

The `cronjob_types.go` file contains the input types, and is loosely based
on the CronJob tutorial from the [KubeBuilder
Book](https://book.kubebuilder.io/cronjob-tutorial.html), but with added
fields to test additional markers and generation behavior.

If you add a new marker, re-generate the golden output file,
`testdata.kubebuilder.io_cronjobs.yaml`, with (if you have the latest
controller-gen on your path):

```bash
go generate
```

or, if you don't have the latest controller-gen on your path, use:

```bash
$ /path/to/current/build/of/controller-gen crd paths=. output:dir=.
```

Make sure you review the diff to ensure that it only contains the desired
changes!

If you didn't add a new marker and this output changes, make sure you have
a good explanation for why generated output needs to change!
