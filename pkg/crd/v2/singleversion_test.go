/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v2

import (
	"reflect"
	"testing"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/ghodss/yaml"
	"github.com/spf13/afero"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

type singleVersionTestcase struct {
	inputPackage string
	types        []string
	flatten      bool

	listFilesFn listFilesFn
	// map of path to file content.
	inputFiles map[string][]byte

	expectedDefs     []byte
	expectedCrdSpecs map[schema.GroupKind][]byte
}

func TestSingleVerGenerate(t *testing.T) {
	testcases := []singleVersionTestcase{
		{
			inputPackage: "github.com/myorg/myapi",
			types:        []string{"Toy"},
			flatten:      false,
			listFilesFn: func(pkgPath string) (s string, strings []string, e error) {
				return "github.com/myorg/myapi", []string{"types.go"}, nil
			},
			inputFiles: map[string][]byte{
				"github.com/myorg/myapi/types.go": []byte(`
package myapi

// Toy is a toy struct
type Toy struct {
	// Replicas is a number
	Replicas int32 ` + "`" + `json:"replicas"` + "`" + `
}
`),
			},
			expectedDefs: []byte(`Toy:
  description: Toy is a toy struct
  properties:
    replicas:
      description: Replicas is a number
      type: integer
  required:
  - replicas
  type: object
`),
		},
	}
	for _, tc := range testcases {
		fs, err := prepareTestFs(tc.inputFiles)
		if err != nil {
			t.Errorf("unable to prepare the in-memory fs for testing: %v", err)
			continue
		}

		op := &SingleVersionOptions{
			InputPackage: tc.inputPackage,
			Types:        tc.types,
			Flatten:      tc.flatten,
			listFilesFn:  tc.listFilesFn,
			fs:           fs,
		}

		defs, crdSpecs := op.parse()

		if len(tc.expectedDefs) > 0 {
			var expectedDefs v1beta1.JSONSchemaDefinitions
			err = yaml.Unmarshal(tc.expectedDefs, &expectedDefs)
			if err != nil {
				t.Errorf("unable to unmarshal the expected definitions: %v", err)
				continue
			}
			if !reflect.DeepEqual(defs, expectedDefs) {
				defsYaml, err := yaml.Marshal(defs)
				if err != nil {
					t.Errorf("unable to marshal the actual definitions: %v", err)
				}
				t.Errorf("expected: %s, but got: %s", tc.expectedDefs, defsYaml)
				continue
			}
		}
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

			if !reflect.DeepEqual(crdSpecs, expectedSpecsByKind) {
				t.Errorf("expected: %s, but got: %s", expectedSpecsByKind, crdSpecs)
				continue
			}
		}
	}
}

func prepareTestFs(files map[string][]byte) (afero.Fs, error) {
	fs := afero.NewMemMapFs()
	for filename, content := range files {
		if err := afero.WriteFile(fs, filename, content, 0666); err != nil {
			return nil, err
		}
	}
	return fs, nil
}
