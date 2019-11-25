# CRD Integration Test testdata

This contains a tiny module used for testdata for the CRD integration
test.  The directory should always be called testdata, so Go treats it
specially.

The `cronjob_types.go` file contains the input types, and is loosely based
on the CronJob tutorial from the [KubeBuilder
Book](https://book.kubebuilder.io/cronjob-tutorial/cronjob-tutorial.html), but with added
fields to test additional markers and generation behavior.

It's *highly* unlikely that the generated expected manifests will
ever change from these -- if they do, you've probably broken something.
Nonetheless, you can regenerate output using `make`.

Make sure you review the diff to ensure that it only contains the desired
changes!

If you didn't add a new marker and this output changes, make sure you have
a good explanation for why generated output needs to change!
