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

package crd

import (
	"fmt"
	"strings"

	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/controller-tools/pkg/loader"
	"sigs.k8s.io/controller-tools/pkg/markers"
)

// PostFlattenMarker is a marker that gets applied after the schema is flattened, where all nested
// properties are available for the marker to use.
type PostFlattenMarker interface {
	PostFlatten()
}

// postFlattenSchema updates an existing schema that has already been flattened, where all nested
// properties are available for markers to use.
func postFlattenSchema(ctx *schemaContext, props *apiext.JSONSchemaProps) {
	applySchemaMarkers(ctx, ctx.info.RawSpec, getPostFlattenSchemaMarkers(ctx.PackageMarkers), props)
	applySchemaMarkers(ctx, ctx.info.RawSpec, getPostFlattenSchemaMarkers(ctx.info.Markers), props)

	for _, field := range ctx.info.Fields {
		fieldMarkers := getPostFlattenSchemaMarkers(field.Markers)
		if len(fieldMarkers) == 0 {
			continue
		}

		jsonTag, hasTag := field.Tag.Lookup("json")
		if !hasTag {
			// if the field doesn't have a JSON tag, it doesn't belong in output (and shouldn't exist in a serialized type)
			ctx.pkg.AddError(loader.ErrFromNode(fmt.Errorf("encountered struct field %q without JSON tag in type %q", field.Name, ctx.info.Name), field.RawField))
			continue
		}

		fieldName := strings.SplitN(jsonTag, ",", 2)[0]

		fieldProps, hasFieldProps := props.Properties[fieldName]
		if !hasFieldProps {
			var names []string
			for name := range fieldMarkers {
				names = append(names, name)
			}

			encountered := "a marker"
			if len(names) > 1 {
				encountered = "markers"
			}

			err := fmt.Errorf("encountered %s which may not be used on inline JSON struct fields, on field %q in type %q: %v", encountered, field.Name, ctx.info.Name, names)
			ctx.pkg.AddError(loader.ErrFromNode(err, field.RawField))
		}

		applySchemaMarkers(ctx, field.RawField, fieldMarkers, &fieldProps)

		props.Properties[fieldName] = fieldProps
	}
}

// getPostFlattenSchemaMarkers returns a filtered slice of markers that are both PostFlattenMarker
// and SchemaMarker types.
func getPostFlattenSchemaMarkers(markerSet markers.MarkerValues) map[string][]SchemaMarker {
	result := make(map[string][]SchemaMarker)

	for name, markerValues := range markerSet {
		for _, markerValue := range markerValues {
			if _, isPostFlattenMarker := markerValue.(PostFlattenMarker); !isPostFlattenMarker {
				continue
			}

			schemaMarker, isSchemaMarker := markerValue.(SchemaMarker)
			if !isSchemaMarker {
				continue
			}

			result[name] = append(result[name], schemaMarker)
		}
	}

	return result
}

// applySchemaMarkers applies SchemaMarker schema markers to the given schema.
func applySchemaMarkers(ctx *schemaContext, node loader.Node, markerSet map[string][]SchemaMarker, props *apiext.JSONSchemaProps) {
	for _, markerValues := range markerSet {
		for _, markerValue := range markerValues {
			if err := markerValue.ApplyToSchema(props); err != nil {
				ctx.pkg.AddError(loader.ErrFromNode(err, node))
			}
		}
	}
}
