package v2

import (
	"reflect"
	"testing"

	"github.com/ghodss/yaml"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type multiVersionTestcase struct {
	inputPackage string
	types        []string

	listFilesFn listFilesFn
	listDirsFn  listDirsFn
	// map of path to file content.
	inputFiles map[string][]byte

	expectedCrdSpecs map[schema.GroupKind][]byte
}

func TestMultiVerGenerate(t *testing.T) {
	testcases := []multiVersionTestcase{
		{
			inputPackage: "github.com/myorg/myapi",
			types:        []string{"Toy"},
			listDirsFn: func(pkgPath string) (strings []string, e error) {
				return []string{"v1", "v1alpha1"}, nil
			},
			listFilesFn: func(pkgPath string) (s string, strings []string, e error) {
				return pkgPath, []string{"types.go"}, nil
			},
			inputFiles: map[string][]byte{
				"github.com/myorg/myapi/v1/types.go": []byte(`
package v1

// +groupName=foo.bar.com
// +versionName=v1

// +kubebuilder:resource:path=toys,shortName=to;ty
// +kubebuilder:singular=toy

// Toy is a toy struct
type Toy struct {
	// +kubebuilder:validation:Maximum=90
	// +kubebuilder:validation:Minimum=1

	// Replicas is a number
	Replicas int32 ` + "`" + `json:"replicas"` + "`" + `
}
`),
				"github.com/myorg/myapi/v1alpha1/types.go": []byte(`
package v1alpha1

// +groupName=foo.bar.com
// +versionName=v1alpha1

// +kubebuilder:resource:path=toys,shortName=to;ty
// +kubebuilder:singular=toy

// Toy is a toy struct
type Toy struct {
	// +kubebuilder:validation:MaxLength=15
	// +kubebuilder:validation:MinLength=1

	// Name is a string
	Name string ` + "`" + `json:"name,omitempty"` + "`" + `

	// +kubebuilder:validation:Maximum=100
	// +kubebuilder:validation:Minimum=1

	// Replicas is a number
	Replicas int32 ` + "`" + `json:"replicas"` + "`" + `
}
`),
			},
			expectedCrdSpecs: map[schema.GroupKind][]byte{
				schema.GroupKind{Group: "foo.bar.com", Kind: "Toy"}: []byte(`group: foo.bar.com
names:
  kind: Toy
  plural: toys
  shortNames:
  - to
  - ty
  singular: toy
scope: Namespaced
versions:
- name: v1
  schema:
    openAPIV3Schema:
      description: Toy is a toy struct
      properties:
        replicas:
          description: Replicas is a number
          maximum: 90
          minimum: 1
          type: integer
      required:
      - replicas
      type: object
  served: true
  storage: false
- name: v1alpha1
  schema:
    openAPIV3Schema:
      description: Toy is a toy struct
      properties:
        name:
          description: Name is a string
          maxLength: 15
          minLength: 1
          type: string
        replicas:
          description: Replicas is a number
          maximum: 100
          minimum: 1
          type: integer
      required:
      - replicas
      type: object
  served: true
  storage: false
`),
			},
		},
	}
	for _, tc := range testcases {
		fs, err := prepareTestFs(tc.inputFiles)
		if err != nil {
			t.Errorf("unable to prepare the in-memory fs for testing: %v", err)
			continue
		}

		op := &MultiVersionOptions{
			InputPackage: tc.inputPackage,
			Types:        tc.types,
			listDirsFn:   tc.listDirsFn,
			listFilesFn:  tc.listFilesFn,
			fs:           fs,
		}

		crdSpecs := op.parse()

		if len(tc.expectedCrdSpecs) > 0 {
			expectedSpecsByKind := map[schema.GroupKind]*v1beta1.CustomResourceDefinitionSpec{}
			for gk := range tc.expectedCrdSpecs {
				var spec v1beta1.CustomResourceDefinitionSpec
				err = yaml.Unmarshal(tc.expectedCrdSpecs[gk], &spec)
				if err != nil {
					t.Errorf("unable to unmarshal the expected crd spec: %v", err)
					continue
				}
				expectedSpecsByKind[gk] = &spec
			}

			if !reflect.DeepEqual(crdSpecs, crdSpecByKind(expectedSpecsByKind)) {
				t.Errorf("expected:\n%+v,\nbut got:\n%+v\n", expectedSpecsByKind, crdSpecs)
				continue
			}
		}
	}
}
