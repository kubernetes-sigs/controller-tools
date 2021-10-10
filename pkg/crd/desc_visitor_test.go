/*
Copyright 2019 The Kubernetes Authors.

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

package crd_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	"sigs.k8s.io/controller-tools/pkg/crd"
)

var _ = Describe("TruncateDescription", func() {

	It("should drop the description for all fields for zero value of maxLen", func() {
		schema := &apiext.JSONSchemaProps{
			Description: "top level description",
			Properties: map[string]apiext.JSONSchemaProps{
				"Spec": {
					Description: "specification for the API type",
				},
			},
		}
		crd.TruncateDescription(schema, 0)
		Expect(schema).To(Equal(&apiext.JSONSchemaProps{
			Properties: map[string]apiext.JSONSchemaProps{
				"Spec": {},
			},
		}))

	})

	It("should truncate the description only in cases where description exceeds maxLen", func() {
		schema := &apiext.JSONSchemaProps{
			Description: "top level description of the root object.",
			Properties: map[string]apiext.JSONSchemaProps{
				"Spec": {
					Description: "specification of a field",
				},
			},
		}
		original := schema.DeepCopy()
		crd.TruncateDescription(schema, len(schema.Description))
		Expect(schema).To(Equal(original))
	})

	It("should truncate the description at closest sentence boundary", func() {
		schema := &apiext.JSONSchemaProps{
			Description: `This is top level description. There is an empty schema. More to come`,
		}
		crd.TruncateDescription(schema, len(schema.Description)-5)
		Expect(schema).To(Equal(&apiext.JSONSchemaProps{
			Description: `This is top level description. There is an empty schema.`,
		}))
	})

	It("should truncate the description at maxLen in absence of sentence boundary", func() {
		schema := &apiext.JSONSchemaProps{
			Description: `This is top level description of the root object`,
		}
		crd.TruncateDescription(schema, len(schema.Description)-2)
		Expect(schema).To(Equal(&apiext.JSONSchemaProps{
			Description: `This is top level description of the root obje`,
		}))
	})
})
