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

package crd

import (
	"slices"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

// RepairListMapKeys marks every x-kubernetes-list-map-keys property that is
// neither required nor defaulted as required.
//
// The API server rejects an associative list whose keys are not required or
// defaulted, because such an item cannot be addressed when merging:
//
//	spec.versions[0].schema.openAPIV3Schema.properties[...].items.properties[type].default:
//	Required value: this property is in x-kubernetes-list-map-keys, so it must
//	have a default or be a required property
//
// A list map key is by definition needed to identify an item, so promoting it
// to required reflects the semantics the markers already assert.
//
// This repair exists because a Go type can carry +optional on a field that is
// also a +listMapKey, which makes controller-gen emit a CRD the API server
// refuses. That combination is present in k8s.io/api itself as of v0.37.0:
// kubernetes/kubernetes#137103 added +optional to the Type and Status fields of
// the apps/v1 DaemonSetCondition, DeploymentCondition, ReplicaSetCondition and
// StatefulSetCondition structs, while Type remains their +listMapKey. Any CRD
// embedding one of those workload types is therefore invalid, and the author of
// that CRD cannot correct the upstream markers.
//
// It must run after flattening, since the item schema of a named element type
// is only a $ref until then.
func RepairListMapKeys(schema *apiextensionsv1.JSONSchemaProps) {
	EditSchema(schema, listMapKeyRepairer{})
}

// listMapKeyRepairer implements SchemaVisitor to repair every associative list
// reachable from the root schema.
type listMapKeyRepairer struct{}

func (v listMapKeyRepairer) Visit(schema *apiextensionsv1.JSONSchemaProps) SchemaVisitor {
	if schema == nil {
		return v
	}

	items := schema.Items
	if len(schema.XListMapKeys) == 0 || items == nil || items.Schema == nil {
		return v
	}

	for _, key := range schema.XListMapKeys {
		prop, found := items.Schema.Properties[key]
		if !found {
			// Keys that name no property are reported separately by
			// ValidateCustomResourceDefinitionOpenAPISchema, so leave them be.
			continue
		}
		if prop.Default != nil || slices.Contains(items.Schema.Required, key) {
			continue
		}
		items.Schema.Required = append(items.Schema.Required, key)
		// Required is kept sorted elsewhere in generation, so keep the
		// repaired output stable too.
		slices.Sort(items.Schema.Required)
	}

	return v
}
