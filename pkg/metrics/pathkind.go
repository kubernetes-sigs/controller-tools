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

package metrics

import (
	"go/types"
	"reflect"
	"slices"
	"strings"

	"sigs.k8s.io/controller-tools/pkg/metrics/internal/model"
)

// annotatePathKinds records the kind of the Go type at the path of each metric and
// how its value has to be converted, which is needed by targets which have to know
// how to iterate over the path or convert the value.
func annotatePathKinds(metrics []model.Generator, root types.Type) {
	for i := range metrics {
		metric := &metrics[i]
		metric.PathKind = goPathKind(root, metric.Path)
		if metric.Type == model.MetricTypeGauge {
			valuePath := append(slices.Clone(metric.Path), metric.Value...)
			if isTimestampType(goTypeAtJSONPath(root, valuePath)) {
				metric.ValueKind = model.ValueKindTimestamp
			}
		}
	}
}

// goPathKind returns the kind of the Go type at the path. A path which cannot be
// resolved is treated as scalar.
func goPathKind(root types.Type, path []string) model.PathKind {
	current := goTypeAtJSONPath(root, path)
	if current == nil {
		return model.PathKindScalar
	}
	switch dereferenceType(current).Underlying().(type) {
	case *types.Slice, *types.Array:
		return model.PathKindArray
	case *types.Map:
		return model.PathKindMap
	case *types.Struct:
		return model.PathKindObject
	default:
		return model.PathKindScalar
	}
}

// goTypeAtJSONPath returns the Go type at the path, which is a list of json field names.
// Lists are walked through without consuming a path element, for maps the path element is the key.
// It returns nil if the path cannot be resolved.
func goTypeAtJSONPath(root types.Type, path []string) types.Type {
	current := root
	for _, element := range path {
		current = dereferenceType(current)
		switch typed := current.Underlying().(type) {
		case *types.Slice:
			current = typed.Elem()
		case *types.Array:
			current = typed.Elem()
		case *types.Map:
			current = typed.Elem()
			continue
		}

		fieldType, ok := jsonFieldType(dereferenceType(current), element)
		if !ok {
			return nil
		}
		current = fieldType
	}
	return current
}

// isTimestampType returns true for the metav1 time types, which are serialized as RFC3339 strings.
func isTimestampType(typ types.Type) bool {
	if typ == nil {
		return false
	}
	named, ok := dereferenceType(typ).(*types.Named)
	if !ok || named.Obj().Pkg() == nil {
		return false
	}
	return named.Obj().Pkg().Path() == "k8s.io/apimachinery/pkg/apis/meta/v1" &&
		(named.Obj().Name() == "Time" || named.Obj().Name() == "MicroTime")
}

// dereferenceType returns the type behind aliases and pointers.
func dereferenceType(typ types.Type) types.Type {
	for {
		typ = types.Unalias(typ)
		pointer, ok := typ.(*types.Pointer)
		if !ok {
			return typ
		}
		typ = pointer.Elem()
	}
}

// jsonFieldType returns the type of the field with the json name in the struct, including
// the fields of inlined structs.
func jsonFieldType(typ types.Type, name string) (types.Type, bool) {
	structure, ok := typ.Underlying().(*types.Struct)
	if !ok {
		return nil, false
	}
	for i := range structure.NumFields() {
		field := structure.Field(i)
		jsonOptions := strings.Split(reflect.StructTag(structure.Tag(i)).Get("json"), ",")
		jsonName := jsonOptions[0]
		if jsonName == name || (jsonName == "" && field.Name() == name) {
			return field.Type(), true
		}
		if field.Embedded() && (jsonName == "" || slices.Contains(jsonOptions[1:], "inline")) {
			if nested, found := jsonFieldType(dereferenceType(field.Type()), name); found {
				return nested, true
			}
		}
	}
	return nil, false
}
