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
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ghodss/yaml"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type WriterOptions struct {
	// OutputPath is the path that the schema will be written to.
	OutputPath string
	// OutputFormat should be either json or yaml. Default to json
	OutputFormat string

	defs     v1beta1.JSONSchemaDefinitions
	crdSpecs crdSpecByKind
}

func (op *WriterOptions) write(outputCRD bool, types []string) {
	var toSerilizeList []interface{}
	if outputCRD {
		for gk, spec := range op.crdSpecs {
			crd := &v1beta1.CustomResourceDefinition{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "apiextensions.k8s.io/v1beta1",
					Kind:       "CustomResourceDefinition",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:   strings.ToLower(gk.Kind),
					Labels: map[string]string{"controller-tools.k8s.io": "1.0"},
				},
				Spec: *spec,
			}
			toSerilizeList = append(toSerilizeList, crd)
		}
	} else {
		schema := v1beta1.JSONSchemaProps{Definitions: op.defs}
		schema.Type = "object"
		schema.AnyOf = []v1beta1.JSONSchemaProps{}
		for _, typeName := range types {
			schema.AnyOf = append(schema.AnyOf, v1beta1.JSONSchemaProps{Ref: getDefLink(typeName)})
		}
		toSerilizeList = []interface{}{schema}
	}

	dir := filepath.Dir(op.OutputPath)
	err := os.MkdirAll(dir, 0666)
	if err != nil {
		log.Fatal(err)
	}
	out, err := os.Create(op.OutputPath)
	if err != nil {
		log.Fatal(err)
	}

	for i := range toSerilizeList {
		switch strings.ToLower(op.OutputFormat) {
		// default to json
		case "json", "":
			enc := json.NewEncoder(out)
			enc.SetIndent("", "  ")
			err = enc.Encode(toSerilizeList[i])
			if err2 := out.Close(); err == nil {
				err = err2
			}
			if err != nil {
				log.Panic(err)
			}
		case "yaml":
			m, err := yaml.Marshal(toSerilizeList[i])
			if err != nil {
				log.Panic(err)
			}
			err = ioutil.WriteFile(op.OutputPath, m, 0644)
			if err != nil {
				log.Panic(err)
			}
		}
	}
}
