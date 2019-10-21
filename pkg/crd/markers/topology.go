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

package markers

import (
	"fmt"

	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/controller-tools/pkg/markers"
)

// ToplogyMarkers list topology markers (i.e. markers that specify if a
// list is an associative-list or a set, or if a map is atomic or not).
var TopologyMarkers = []*definitionWithHelp{
	must(markers.MakeDefinition("listMapKey", markers.DescribesField, ListMapKey(""))).
		WithHelp(ListMapKey("").Help()),
	must(markers.MakeDefinition("listType", markers.DescribesField, ListType(""))).
		WithHelp(ListType("").Help()),
}

func init() {
	AllDefinitions = append(AllDefinitions, TopologyMarkers...)
}

// +controllertools:marker:generateHelp:category="CRD topology"

// ListType specifies the type of data-structure that the list
// represents (map, set, atomic).
//
// Possible data-structure types of a list are:
//
// - "map": it needs to have a key field, which will be used to build an
//   associative list. A typical example is a the pod container list,
//   which is indexed by the container name.
//
// - "set": Fields need to be "scalar", and there can be only one
//   occurrence of each.
//
// - "atomic": All the fields in the list are treated as a single value,
//   are typically manipulated together by the same actor.
type ListType string

// +controllertools:marker:generateHelp:category="CRD topology"

// ListMapKey specifies the keys to map listTypes.
//
// It indicates the index of a map list. They can be repeated if multiple keys
// must be used. It can only be used when ListType is set to map, and the keys
// should be scalar types.
type ListMapKey string

func (l ListType) ApplyToSchema(schema *apiext.JSONSchemaProps) error {
	if schema.Type != "array" {
		return fmt.Errorf("must apply listType to an array")
	}
	if l != "map" && l != "atomic" && l != "set" {
		return fmt.Errorf(`ListType must be either "map", "set" or "atomic"`)
	}
	p := string(l)
	schema.XListType = &p
	return nil
}

func (l ListType) ApplyFirst() {}

func (l ListMapKey) ApplyToSchema(schema *apiext.JSONSchemaProps) error {
	if schema.Type != "array" {
		return fmt.Errorf("must apply listMapKey to an array")
	}
	if schema.XListType == nil || *schema.XListType != "map" {
		return fmt.Errorf("must apply listMapKey to an associative-list")
	}
	schema.XListMapKeys = append(schema.XListMapKeys, string(l))
	return nil
}
