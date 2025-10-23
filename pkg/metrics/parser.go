/*
Copyright 2024 The Kubernetes Authors All rights reserved.

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
	"fmt"
	"go/ast"
	"go/types"
	"strings"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-tools/pkg/crd"
	"sigs.k8s.io/controller-tools/pkg/loader"
	ctrlmarkers "sigs.k8s.io/controller-tools/pkg/markers"

	"sigs.k8s.io/controller-tools/pkg/metrics/internal/config"
	"sigs.k8s.io/controller-tools/pkg/metrics/markers"
)

type parser struct {
	*crd.Parser

	CustomResourceStates map[crd.TypeIdent]*config.Resource
}

func newParser(p *crd.Parser) *parser {
	return &parser{
		Parser:               p,
		CustomResourceStates: make(map[crd.TypeIdent]*config.Resource),
	}
}

// NeedResourceFor creates the customresourcestate.Resource object for the given
// GroupKind located at the package identified by packageID.
func (p *parser) NeedResourceFor(pkg *loader.Package, groupKind schema.GroupKind) error {
	typeIdent := crd.TypeIdent{Package: pkg, Name: groupKind.Kind}
	// Skip if type was already processed.
	if _, exists := p.CustomResourceStates[typeIdent]; exists {
		return nil
	}

	// Already mark the cacheID so the next time it enters NeedResourceFor it skips early.
	p.CustomResourceStates[typeIdent] = nil

	// Build the type identifier for the custom resource.
	typeInfo := p.Types[typeIdent]
	// typeInfo is nil if this GroupKind is not part of this package. In that case
	// we have nothing to process.
	if typeInfo == nil {
		return nil
	}

	// Skip if gvk marker is not set. This marker is the opt-in for creating metrics
	// for a custom resource.
	if m := typeInfo.Markers.Get(markers.GVKMarkerName); m == nil {
		return nil
	}

	metrics, err := p.NeedMetricsGeneratorFor(typeIdent)
	if err != nil {
		return err
	}

	// Initialize the Resource object.
	resource := config.Resource{
		GroupVersionKind: config.GroupVersionKind{
			Group:   groupKind.Group,
			Kind:    groupKind.Kind,
			Version: p.GroupVersions[pkg].Version,
		},
		// Create the metrics generators for the custom resource.
		Metrics: metrics,
	}

	// Iterate through all markers and run the ApplyToResource function of the ResourceMarkers.
	for _, markerVals := range typeInfo.Markers {
		for _, val := range markerVals {
			if resourceMarker, isResourceMarker := val.(markers.ResourceMarker); isResourceMarker {
				if err := resourceMarker.ApplyToResource(&resource); err != nil {
					pkg.AddError(loader.ErrFromNode(err /* an okay guess */, typeInfo.RawSpec))
				}
			}
		}
	}

	p.CustomResourceStates[typeIdent] = &resource
	return nil
}

type generatorRequester interface {
	NeedMetricsGeneratorFor(typ crd.TypeIdent) ([]config.Generator, error)
}

// generatorContext stores and provides information across a hierarchy of metric generators generation.
type generatorContext struct {
	pkg                *loader.Package
	generatorRequester generatorRequester

	PackageMarkers ctrlmarkers.MarkerValues
}

func newGeneratorContext(pkg *loader.Package, req generatorRequester) *generatorContext {
	pkg.NeedTypesInfo()
	return &generatorContext{
		pkg:                pkg,
		generatorRequester: req,
	}
}

func generatorsFromMarkers(m ctrlmarkers.MarkerValues, basePath ...string) ([]config.Generator, error) {
	generators := []config.Generator{}

	for _, markerVals := range m {
		for _, val := range markerVals {
			if generatorMarker, isGeneratorMarker := val.(markers.LocalGeneratorMarker); isGeneratorMarker {
				g, err := generatorMarker.ToGenerator(basePath...)
				if err != nil {
					return nil, err
				}
				if g != nil {
					generators = append(generators, *g)
				}
			}
		}
	}

	return generators, nil
}

// NeedMetricsGeneratorFor creates the customresourcestate.Generator object for a
// Custom Resource.
func (p *parser) NeedMetricsGeneratorFor(typ crd.TypeIdent) ([]config.Generator, error) {
	info, gotInfo := p.Types[typ]
	if !gotInfo {
		return nil, fmt.Errorf("type info for %v does not exist", typ)
	}

	// Add metric allGenerators defined by markers at the type.
	allGenerators, err := generatorsFromMarkers(info.Markers)
	if err != nil {
		return nil, err
	}

	// Traverse fields of the object and process markers.
	// Note: Partially inspired by controller-tools.
	// xref: https://github.com/kubernetes-sigs/controller-tools/blob/d89d6ae3df218a85f7cd9e477157cace704b37d1/pkg/crd/schema.go#L350
	for _, f := range info.Fields {
		// Only fields with the `json:"..."` tag are relevant. Others are not part of the Custom Resource.
		jsonTag, hasTag := f.Tag.Lookup("json")
		if !hasTag {
			// if the field doesn't have a JSON tag, it doesn't belong in output (and shouldn't exist in a serialized type)
			continue
		}
		jsonOpts := strings.Split(jsonTag, ",")
		if len(jsonOpts) == 1 && jsonOpts[0] == "-" {
			// skipped fields have the tag "-" (note that "-," means the field is named "-")
			continue
		}

		// Add metric markerGenerators defined by markers at the field.
		markerGenerators, err := generatorsFromMarkers(f.Markers, jsonOpts[0])
		if err != nil {
			return nil, err
		}
		allGenerators = append(allGenerators, markerGenerators...)

		// Create new generator context and recursively process the fields.
		generatorCtx := newGeneratorContext(typ.Package, p)
		generators, err := generatorsFor(generatorCtx, f.RawField.Type)
		if err != nil {
			return nil, err
		}
		for _, generator := range generators {
			allGenerators = append(allGenerators, addPathPrefixOnGenerator(generator, jsonOpts[0]))
		}
	}

	return allGenerators, nil
}

// generatorsFor creates generators for the given AST type.
// Note: Partially inspired by controller-tools.
// xref: https://github.com/kubernetes-sigs/controller-tools/blob/d89d6ae3df218a85f7cd9e477157cace704b37d1/pkg/crd/schema.go#L167-L193
func generatorsFor(ctx *generatorContext, rawType ast.Expr) ([]config.Generator, error) {
	switch expr := rawType.(type) {
	case *ast.Ident:
		return localNamedToGenerators(ctx, expr)
	case *ast.SelectorExpr:
		// Results in using transitive markers from external packages.
		return generatorsFor(ctx, expr.X)
	case *ast.ArrayType:
		// The current configuration does not allow creating metric configurations inside arrays
		return nil, nil
	case *ast.MapType:
		// The current configuration does not allow creating metric configurations inside maps
		return nil, nil
	case *ast.StarExpr:
		return generatorsFor(ctx, expr.X)
	case *ast.StructType:
		ctx.pkg.AddError(loader.ErrFromNode(fmt.Errorf("unsupported AST kind %T", expr), rawType))
	default:
		ctx.pkg.AddError(loader.ErrFromNode(fmt.Errorf("unsupported AST kind %T", expr), rawType))
		// NB(directxman12): we explicitly don't handle interfaces
		return nil, nil
	}

	return nil, nil
}

// localNamedToGenerators recurses back to NeedMetricsGeneratorFor for the type to
// get generators defined at the objects in a custom resource.
func localNamedToGenerators(ctx *generatorContext, ident *ast.Ident) ([]config.Generator, error) {
	typeInfo := ctx.pkg.TypesInfo.TypeOf(ident)
	if typeInfo == types.Typ[types.Invalid] {
		// It is expected to hit this error for types from not loaded transitive package dependencies.
		// This leads to ignoring markers defined on the transitive types. Otherwise
		// markers on transitive types would lead to additional metrics.
		return nil, nil
	}

	if _, isBasic := typeInfo.(*types.Basic); isBasic {
		// There can't be markers for basic go types for this generator.
		return nil, nil
	}

	// NB(directxman12): if there are dot imports, this might be an external reference,
	// so use typechecking info to get the actual object
	typeNameInfo := typeInfo.(*types.Named).Obj()
	pkg := typeNameInfo.Pkg()
	pkgPath := loader.NonVendorPath(pkg.Path())
	if pkg == ctx.pkg.Types {
		pkgPath = ""
	}
	return ctx.requestGenerator(pkgPath, typeNameInfo.Name())
}

// requestGenerator asks for the generator for a type in the package with the
// given import path.
func (c *generatorContext) requestGenerator(pkgPath, typeName string) ([]config.Generator, error) {
	pkg := c.pkg
	if pkgPath != "" {
		pkg = c.pkg.Imports()[pkgPath]
	}
	return c.generatorRequester.NeedMetricsGeneratorFor(crd.TypeIdent{
		Package: pkg,
		Name:    typeName,
	})
}

// addPathPrefixOnGenerator prefixes the path set at the generators MetricMeta object.
func addPathPrefixOnGenerator(generator config.Generator, pathPrefix string) config.Generator {
	switch generator.Each.Type {
	case config.MetricTypeGauge:
		generator.Each.Gauge.MetricMeta.Path = append([]string{pathPrefix}, generator.Each.Gauge.MetricMeta.Path...)
	case config.MetricTypeStateSet:
		generator.Each.StateSet.MetricMeta.Path = append([]string{pathPrefix}, generator.Each.StateSet.MetricMeta.Path...)
	case config.MetricTypeInfo:
		generator.Each.Info.MetricMeta.Path = append([]string{pathPrefix}, generator.Each.Info.MetricMeta.Path...)
	}

	return generator
}
