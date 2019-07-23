# DeepCopy Integration Test testdata

This contains a tiny module used for testdata for the DeepCopy integration
test. The directory should always be called testdata, so Go treats it
specially.

The `cronjob_types.go` file contains the input types, and is loosely based
on the CronJob tutorial from the [KubeBuilder
Book](https://book.kubebuilder.io/cronjob-tutorial.html), but with added
fields to test additional DeepCopy cases.

If you for some reason need to change deepcopy generation, you can
re-generate the golden output file,
`zz_generated.deepcopy.go`, with (if you have the latest
controller-gen on your path):

```bash
go generate
```

or, if you don't have the latest controller-gen on your path, use:

```bash
$ /path/to/current/build/of/controller-gen object paths=.
```


Make sure you review the diff to ensure that it only contains the desired
changes!

If you didn't add a new marker and this output changes, make sure you have
a good explanation for why generated output needs to change!
