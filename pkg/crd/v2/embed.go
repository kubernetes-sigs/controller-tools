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
	"log"
	"strings"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

func embedSchema(defs map[string]v1beta1.JSONSchemaProps, startingTypes map[string]bool) map[string]v1beta1.JSONSchemaProps {
	newDefs := map[string]v1beta1.JSONSchemaProps{}
	for name := range startingTypes {
		def := defs[name]
		embedDefinition(&def, defs)
		newDefs[name] = def
	}
	return newDefs
}

func embedDefinition(def *v1beta1.JSONSchemaProps, refs map[string]v1beta1.JSONSchemaProps) {
	if def == nil {
		return
	}

	if def.Ref != nil && len(*def.Ref) > 0 {
		refName := strings.TrimPrefix(*def.Ref, defPrefix)
		ref, ok := refs[refName]
		if !ok {
			log.Panicf("can't find the definition of %q", refName)
		}
		def.Properties = ref.Properties
		def.Required = ref.Required
		def.Type = ref.Type
		def.Ref = nil
	}

	def.Definitions = embedDefinitionMap(def.Definitions, refs)
	def.Properties = embedDefinitionMap(def.Properties, refs)
	// TODO: decide if we want to do this.
	//def.AllOf = embedDefinitionArray(def.AllOf, refs)
	def.AnyOf = embedDefinitionArray(def.AnyOf, refs)
	def.OneOf = embedDefinitionArray(def.OneOf, refs)
	if def.AdditionalItems != nil {
		embedDefinition(def.AdditionalItems.Schema, refs)
	}
	if def.Items != nil {
		embedDefinition(def.Items.Schema, refs)
	}
	embedDefinition(def.Not, refs)
}

func embedDefinitionMap(defs map[string]v1beta1.JSONSchemaProps, refs map[string]v1beta1.JSONSchemaProps) map[string]v1beta1.JSONSchemaProps {
	newDefs := map[string]v1beta1.JSONSchemaProps{}
	for i := range defs {
		def := defs[i]
		embedDefinition(&def, refs)
		newDefs[i] = def
	}
	if len(newDefs) == 0 {
		return nil
	}
	return newDefs
}

func embedDefinitionArray(defs []v1beta1.JSONSchemaProps, refs map[string]v1beta1.JSONSchemaProps) []v1beta1.JSONSchemaProps {
	newDefs := make([]v1beta1.JSONSchemaProps, len(defs))
	for i := range defs {
		def := defs[i]
		embedDefinition(&def, refs)
		newDefs[i] = def
	}
	if len(newDefs) == 0 {
		return nil
	}
	return newDefs
}
