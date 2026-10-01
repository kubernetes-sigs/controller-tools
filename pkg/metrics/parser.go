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
	"errors"
	"fmt"
	"go/ast"
	"go/types"
	"maps"
	"slices"
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

	// inProgress contains the types for which generators are currently getting
	// created to not recurse endlessly on self-referencing types.
	inProgress map[crd.TypeIdent]struct{}
}

func newParser(p *crd.Parser) *parser {
	return &parser{
		Parser:               p,
		CustomResourceStates: make(map[crd.TypeIdent]*config.Resource),
		inProgress:           make(map[crd.TypeIdent]struct{}),
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

	// Skip if the store marker is not set. This marker is the opt-in for creating metrics
	// for a custom resource.
	if m := typeInfo.Markers.Get(markers.StoreMarkerName); m == nil {
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
	for _, markerName := range sortedMarkerNames(typeInfo.Markers) {
		for _, val := range typeInfo.Markers[markerName] {
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

// NeedMetricsGeneratorFor creates the customresourcestate.Generator object for a
// Custom Resource.
func (p *parser) NeedMetricsGeneratorFor(typ crd.TypeIdent) ([]config.Generator, error) {
	// Types of other packages are only known after the package got indexed.
	p.NeedPackage(typ.Package)

	if _, recursing := p.inProgress[typ]; recursing {
		// Self-referencing types would result in an endless path.
		return nil, nil
	}
	p.inProgress[typ] = struct{}{}
	defer delete(p.inProgress, typ)

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
	// Note: This follows how the crd package traverses the fields of a struct, see structToSchema in pkg/crd/schema.go.
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

		// Fields without a name (embedded fields) and fields with the inline option are
		// inlined into the parent. They must not add an element to the path.
		var pathPrefix []string
		if jsonOpts[0] != "" && !slices.Contains(jsonOpts[1:], "inline") {
			pathPrefix = []string{jsonOpts[0]}
		}

		// Add metric markerGenerators defined by markers at the field.
		markerGenerators, err := generatorsFromMarkers(f.Markers, pathPrefix...)
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
			allGenerators = append(allGenerators, addPathPrefixOnGenerator(generator, pathPrefix))
		}
	}

	return allGenerators, nil
}

// sortedMarkerNames returns the marker names sorted, to iterate over the markers
// in a deterministic order. MarkerValues is a map and its iteration order is random.
func sortedMarkerNames(m ctrlmarkers.MarkerValues) []string {
	return slices.Sorted(maps.Keys(m))
}

func generatorsFromMarkers(m ctrlmarkers.MarkerValues, basePath ...string) ([]config.Generator, error) {
	generators := []config.Generator{}

	for _, markerName := range sortedMarkerNames(m) {
		for _, val := range m[markerName] {
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

// generatorsFor creates generators for the given AST type.
// Note: This follows how the crd package maps AST types to schemas, see typeToSchema in pkg/crd/schema.go.
func generatorsFor(ctx *generatorContext, rawType ast.Expr) ([]config.Generator, error) {
	switch expr := rawType.(type) {
	case *ast.Ident:
		return localNamedToGenerators(ctx, expr)
	case *ast.SelectorExpr:
		// Results in using transitive markers from external packages.
		return localNamedToGenerators(ctx, expr.Sel)
	case *ast.ArrayType:
		// The current configuration does not allow creating metric configurations inside arrays
		return nil, nil
	case *ast.MapType:
		// The current configuration does not allow creating metric configurations inside maps
		return nil, nil
	case *ast.StarExpr:
		return generatorsFor(ctx, expr.X)
	case *ast.StructType:
		// Markers on fields of inline struct types don't get collected, so metrics for them
		// would silently be missing. Report the same error as the crd generator.
		ctx.pkg.AddError(loader.ErrFromNode(errors.New("inline struct types are not supported, use a named type instead"), rawType))
		return nil, nil
	case *ast.InterfaceType:
		ctx.pkg.AddError(loader.ErrFromNode(errors.New("interface type is not supported in CRD schemas; consider using an explicit type or apiextensionsv1.JSON instead"), rawType))
		return nil, nil
	default:
		ctx.pkg.AddError(loader.ErrFromNode(fmt.Errorf("unsupported AST kind %T", expr), rawType))
		return nil, nil
	}
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

	namedType, isNamed := typeInfo.(*types.Named)
	if !isNamed {
		// There can't be markers for unnamed go types for this generator.
		return nil, nil
	}

	// NB(directxman12): if there are dot imports, this might be an external reference,
	// so use typechecking info to get the actual object
	typeNameInfo := namedType.Obj()
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
func addPathPrefixOnGenerator(generator config.Generator, pathPrefix []string) config.Generator {
	if len(pathPrefix) == 0 {
		return generator
	}

	switch generator.Each.Type {
	case config.MetricTypeGauge:
		generator.Each.Gauge.MetricMeta.Path = append(slices.Clone(pathPrefix), generator.Each.Gauge.MetricMeta.Path...)
	case config.MetricTypeStateSet:
		generator.Each.StateSet.MetricMeta.Path = append(slices.Clone(pathPrefix), generator.Each.StateSet.MetricMeta.Path...)
	case config.MetricTypeInfo:
		generator.Each.Info.MetricMeta.Path = append(slices.Clone(pathPrefix), generator.Each.Info.MetricMeta.Path...)
	}

	return generator
}
