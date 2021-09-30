# Linter for Kubernetes CustomResourceDefinition resources

The crd-linter encompasses a number of linters that can be run against
Kubernetes CustomResourceDefinition (CRD) objects to help ensure best-practices
are followed when CRDs are authored.

## Usage

```
$ crd-linter --help
Usage of crd-linter:
      --exceptions-file string   Path to a list of exceptions for linter failures
      --linters strings          Full list of linters to run against discovered CustomResourceDefinitions (default [SchemaProvided,PreserveUnknownFields,MaxLengthStrings,MaxItemsArrays])
      --output-exceptions        If true, an exception list file will be written to the file denoted by '--exceptions-file'
      --path string              Path to recursively search for CustomResourceDefinition objects
```

The crd-linter must be pointed at a directory containing YAML files with CRD
object(s) defined in them using the `--path` flag.

For example:

```
crd-linter --path ./path/to/crds/
```

### Linters

The following linters are supported:

| Name                  | Description                                                         |
|-----------------------|---------------------------------------------------------------------|
| MaxLengthStrings      | Ensures that all 'string' type fields have a `maxLength` option set |
| MaxItemsArrays        | Ensures that all 'array' type fields have a `maxItems` option set   |
| SchemaProvided        | Ensures that all versions of all CRDs provide an OpenAPI schema     |
| PreserveUnknownFields | Ensures that no fields within a CRD schema permit unknown fields    |

### Exceptions

An exceptions file may be provided which will permit certain rule violations to
be skipped. This is useful when gradually improving/fixing issues within a
repository of CRDs, whilst not permitting new violations.

You can generate an exceptions file by setting `--output-exceptions=true`
whilst also providing a `--exceptions-file`.
This will output all linter failures to the named exceptions file, so that
future runs of the linter will skip these failures.

Use this in combination with the `--linters` flag to generate an exceptions
file that only contains a subset of linter failures (e.g. only the nosiest or
least problematic failures).
