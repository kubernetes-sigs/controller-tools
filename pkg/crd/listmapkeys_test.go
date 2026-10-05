/*
Copyright 2026 The Kubernetes Authors.

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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/controller-tools/pkg/crd"
)

// associativeList builds an associative list whose items carry the given
// properties and required fields, keyed by the given list map keys.
func associativeList(keys []string, required []string, props map[string]apiextensionsv1.JSONSchemaProps) *apiextensionsv1.JSONSchemaProps {
	listType := "map"
	return &apiextensionsv1.JSONSchemaProps{
		Type: "object",
		Properties: map[string]apiextensionsv1.JSONSchemaProps{
			"conditions": {
				Type:         "array",
				XListType:    &listType,
				XListMapKeys: keys,
				Items: &apiextensionsv1.JSONSchemaPropsOrArray{
					Schema: &apiextensionsv1.JSONSchemaProps{
						Type:       "object",
						Required:   required,
						Properties: props,
					},
				},
			},
		},
	}
}

var conditionProps = map[string]apiextensionsv1.JSONSchemaProps{
	"type":   {Type: "string"},
	"status": {Type: "string"},
	"reason": {Type: "string"},
}

var _ = Describe("RepairListMapKeys", func() {
	It("should mark an optional list map key as required", func() {
		schema := associativeList([]string{"type"}, nil, conditionProps)

		crd.RepairListMapKeys(schema)

		Expect(schema.Properties["conditions"].Items.Schema.Required).To(Equal([]string{"type"}))
	})

	It("should leave an already required list map key alone", func() {
		schema := associativeList([]string{"type"}, []string{"type"}, conditionProps)

		crd.RepairListMapKeys(schema)

		Expect(schema.Properties["conditions"].Items.Schema.Required).To(Equal([]string{"type"}))
	})

	It("should leave a defaulted list map key alone", func() {
		props := map[string]apiextensionsv1.JSONSchemaProps{
			"type": {Type: "string", Default: &apiextensionsv1.JSON{Raw: []byte(`"Ready"`)}},
		}
		schema := associativeList([]string{"type"}, nil, props)

		crd.RepairListMapKeys(schema)

		Expect(schema.Properties["conditions"].Items.Schema.Required).To(BeEmpty())
	})

	It("should repair every key of a composite key, keeping required sorted", func() {
		props := map[string]apiextensionsv1.JSONSchemaProps{
			"containerPort": {Type: "integer"},
			"protocol":      {Type: "string"},
		}
		schema := associativeList([]string{"protocol", "containerPort"}, nil, props)

		crd.RepairListMapKeys(schema)

		Expect(schema.Properties["conditions"].Items.Schema.Required).To(Equal([]string{"containerPort", "protocol"}))
	})

	It("should preserve unrelated required entries and stay sorted", func() {
		schema := associativeList([]string{"type"}, []string{"status"}, conditionProps)

		crd.RepairListMapKeys(schema)

		Expect(schema.Properties["conditions"].Items.Schema.Required).To(Equal([]string{"status", "type"}))
	})

	It("should ignore a key that names no property", func() {
		schema := associativeList([]string{"missing"}, nil, conditionProps)

		crd.RepairListMapKeys(schema)

		Expect(schema.Properties["conditions"].Items.Schema.Required).To(BeEmpty())
	})

	It("should ignore lists that are not associative", func() {
		schema := &apiextensionsv1.JSONSchemaProps{
			Type: "object",
			Properties: map[string]apiextensionsv1.JSONSchemaProps{
				"args": {
					Type:  "array",
					Items: &apiextensionsv1.JSONSchemaPropsOrArray{Schema: &apiextensionsv1.JSONSchemaProps{Type: "string"}},
				},
			},
		}

		crd.RepairListMapKeys(schema)

		Expect(schema.Properties["args"].Items.Schema.Required).To(BeEmpty())
	})

	It("should repair associative lists nested inside other schemata", func() {
		nested := associativeList([]string{"type"}, nil, conditionProps)
		schema := &apiextensionsv1.JSONSchemaProps{
			Type: "object",
			Properties: map[string]apiextensionsv1.JSONSchemaProps{
				"template": {
					Type: "object",
					Properties: map[string]apiextensionsv1.JSONSchemaProps{
						"status": *nested,
					},
				},
			},
		}

		crd.RepairListMapKeys(schema)

		status := schema.Properties["template"].Properties["status"]
		Expect(status.Properties["conditions"].Items.Schema.Required).To(Equal([]string{"type"}))
	})

	It("should be idempotent", func() {
		schema := associativeList([]string{"type"}, nil, conditionProps)

		crd.RepairListMapKeys(schema)
		crd.RepairListMapKeys(schema)

		Expect(schema.Properties["conditions"].Items.Schema.Required).To(Equal([]string{"type"}))
	})

})
