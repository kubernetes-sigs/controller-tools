# RBAC Integration Test testdata

This contains a tiny module used for testdata for the RBAC integration
test.  The directory should always be called testdata, so Go treats it
specially.

If you add a new marker, re-generate the golden output file,
`role.yaml`, with:

```bash
$ /path/to/current/build/of/controller-gen rbac:roleName=manager-role paths=. output:dir=.
```

Make sure you review the diff to ensure that it only contains the desired
changes!

If you didn't add a new marker and this output changes, make sure you have
a good explanation for why generated output needs to change!
