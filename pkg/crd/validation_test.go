package crd_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsvalidation "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/validation"
	apiserverschema "k8s.io/apiextensions-apiserver/pkg/apiserver/schema"
	"k8s.io/apiextensions-apiserver/pkg/apiserver/schema/cel"
	"k8s.io/apiextensions-apiserver/pkg/apiserver/validation"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/yaml"
	celconfig "k8s.io/apiserver/pkg/apis/cel"
)

func TestOneOfConstraints(t *testing.T) {
	testCases := []struct {
		name    string
		obj     string
		wantErr string
	}{
		{
			name: "satisfies AtMostOneOf and ExactlyOneOf constraints",
			obj: `---
kind: Oneof
apiVersion: testdata.kubebuilder.io/v1beta1
metadata:
  name: test
spec:
  firstTypeWithOneof:
    foo: "foo"
  secondTypeWithOneof:
    a: "a"
  firstTypeWithExactOneof:
    foo: "foo"
  secondTypeWithExactOneof:
    a: "a"
    c: "c"
`,
		},
		{
			name: "satisfies AtMostOneOf and ExactlyOneOf constraints by not specifying the Type containing the oneof fields",
			obj: `---
kind: Oneof
apiVersion: testdata.kubebuilder.io/v1beta1
metadata:
  name: test
spec: {}
`,
		},
		{
			name: "AtMostOneOf constraint violated by specifying both fields in firstTypeWithOneof",
			obj: `---
kind: Oneof
apiVersion: testdata.kubebuilder.io/v1beta1
metadata:
  name: test
spec:
  firstTypeWithOneof:
    foo: "foo"
    bar : "bar"
  firstTypeWithExactOneof:
    foo: "foo"
  secondTypeWithExactOneof:
    a: "a"
    c: "c"
`,
			wantErr: `spec.firstTypeWithOneof: Invalid value: at most one of the fields in [foo bar] may be set`,
		},
		{
			name: "ExactlyOneOf constraint violated by specifying both fields secondTypeWithExactOneof.c&d",
			obj: `---
kind: Oneof
apiVersion: testdata.kubebuilder.io/v1beta1
metadata:
  name: test
spec:
  firstTypeWithOneof:
    foo: "foo"
  firstTypeWithExactOneof:
    foo: "foo"
  secondTypeWithExactOneof:
    a: "a"
    c: "c"
    d: "d"
`,
			wantErr: `spec.secondTypeWithExactOneof: Invalid value: exactly one of the fields in [c d] must be set`,
		},
		{
			name: "ExactlyOneOf constraint violated by not specifying field secondTypeWithExactOneof.c|d",
			obj: `---
kind: Oneof
apiVersion: testdata.kubebuilder.io/v1beta1
metadata:
  name: test
spec:
  firstTypeWithExactOneof:
    foo: "foo"
  secondTypeWithExactOneof:
    a: "a"
`,
			wantErr: `spec.secondTypeWithExactOneof: Invalid value: exactly one of the fields in [c d] must be set`,
		},
		{
			name: "AtLeastOneOf constraint violated by not specifying field typeWithAllOneOf.e|f",
			obj: `---
kind: Oneof
apiVersion: testdata.kubebuilder.io/v1beta1
metadata:
  name: test
spec:
  typeWithAllOneOf:
    c: "c"
`,
			wantErr: `spec.typeWithAllOneOf: Invalid value: at least one of the fields in [e f] must be set`,
		},
	}

	validator, err := newValidator(t.Context(), "./testdata/testdata.kubebuilder.io_oneofs.yaml")
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validator.Validate(t.Context(), tc.obj)

			if tc.wantErr != "" && !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("expected error containing %q, got: %v", tc.wantErr, err)
			} else if tc.wantErr == "" && err != nil {
				t.Errorf("expected no error, got: %v", err)
			}
		})
	}
}

type validator struct {
	schemaValidator  map[schema.GroupVersionKind]validation.SchemaValidator
	structuralSchema map[schema.GroupVersionKind]*apiserverschema.Structural
	celValidator     map[schema.GroupVersionKind]*cel.Validator
}

func (v *validator) Validate(ctx context.Context, rawObj string) error {
	u, err := parseObjToUnstructured([]byte(rawObj))
	if err != nil {
		return fmt.Errorf("failed to parse object: %w", err)
	}

	gvk := u.GroupVersionKind()
	schemaValidator := v.schemaValidator[gvk]
	structuralSchema := v.structuralSchema[gvk]
	celValidator := v.celValidator[gvk]

	if err := validation.ValidateCustomResource(nil, u.Object, schemaValidator).ToAggregate(); err != nil {
		return fmt.Errorf("schema validation failed: %w", err)
	}

	errs, _ := celValidator.Validate(ctx, nil, structuralSchema, u.Object, nil /*UPDATE not handled*/, celconfig.RuntimeCELCostBudget)
	if errs.ToAggregate() != nil {
		return fmt.Errorf("CEL validation failed: %w", errs.ToAggregate())
	}

	return nil
}

func newValidator(ctx context.Context, crdFile string) (*validator, error) {
	crds, err := parseCRDs(crdFile)
	if err != nil {
		return nil, err
	}

	v := &validator{
		schemaValidator:  make(map[schema.GroupVersionKind]validation.SchemaValidator),
		structuralSchema: make(map[schema.GroupVersionKind]*apiserverschema.Structural),
		celValidator:     make(map[schema.GroupVersionKind]*cel.Validator),
	}

	for _, crd := range crds {
		versions := crd.Spec.Versions
		if len(versions) == 0 {
			return nil, fmt.Errorf("spec.versions not set for CRD %s.%s", crd.Kind, crd.Spec.Group)
		}

		for _, ver := range versions {
			crd.Status.StoredVersions = append(crd.Status.StoredVersions, ver.Name) // HACK

			gvk := schema.GroupVersionKind{
				Group:   crd.Spec.Group,
				Version: ver.Name,
				Kind:    crd.Spec.Names.Kind,
			}
			validationSchema, err := apiextensions.GetSchemaForVersion(crd, ver.Name)
			if err != nil {
				return nil, err
			}
			schemaValidator, _, err := validation.NewSchemaValidator(validationSchema.OpenAPIV3Schema)
			if err != nil {
				return nil, err
			}
			v.schemaValidator[gvk] = schemaValidator
			structuralSchema, err := apiserverschema.NewStructural(validationSchema.OpenAPIV3Schema)
			if err != nil {
				return nil, err
			}
			v.structuralSchema[gvk] = structuralSchema
			celValidator := cel.NewValidator(structuralSchema, true /*resource root*/, celconfig.PerCallLimit)
			v.celValidator[gvk] = celValidator

			// Statically validate the CRD. This also validates the total cost of CEL rules in the CRD
			if err := apiextensionsvalidation.ValidateCustomResourceDefinition(ctx, crd).ToAggregate(); err != nil {
				return nil, err
			}
		}
	}

	return v, nil
}

func parseCRDs(path string) ([]*apiextensions.CustomResourceDefinition, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	decoder := yaml.NewYAMLOrJSONDecoder(f, 4096)

	// There could be multiple CRDs per file (e.g., for testing)
	var crds []*apiextensions.CustomResourceDefinition
	for {
		raw := new(unstructured.Unstructured)
		err := decoder.Decode(raw)
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return nil, err
		}

		// Assume all our CRDs are apiextensions.k8s.io/v1
		crd := new(apiextensions.CustomResourceDefinition)
		crdv1 := new(apiextensionsv1.CustomResourceDefinition)
		if err := runtime.DefaultUnstructuredConverter.
			FromUnstructured(raw.UnstructuredContent(), crdv1); err != nil {
			return nil, err
		}
		if err := apiextensionsv1.Convert_v1_CustomResourceDefinition_To_apiextensions_CustomResourceDefinition(crdv1, crd, nil); err != nil {
			return nil, err
		}

		crds = append(crds, crd)
	}

	return crds, nil
}

func parseObjToUnstructured(data []byte) (unstructured.Unstructured, error) {
	var u unstructured.Unstructured

	// Split the YAML manifest into separate documents
	decoder := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(data), 4096)
	if err := decoder.Decode(&u); err != nil {
		if errors.Is(err, io.EOF) {
			return u, err
		}
		return u, err
	}
	return u, nil
}
