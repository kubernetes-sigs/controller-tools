/*
Copyright 2021 The Kubernetes Authors.

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

package applyconfigurations

import (
	"fmt"
	"go/ast"
	"go/types"
	"os"
	"path/filepath"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
	generatorargs "k8s.io/code-generator/cmd/applyconfiguration-gen/args"
	applygenerator "k8s.io/code-generator/cmd/applyconfiguration-gen/generators"
	"k8s.io/gengo/generator"
	crdmarkers "sigs.k8s.io/controller-tools/pkg/crd/markers"
	"sigs.k8s.io/controller-tools/pkg/genall"
	"sigs.k8s.io/controller-tools/pkg/loader"
	"sigs.k8s.io/controller-tools/pkg/markers"
)

// Based on deepcopy gen but with legacy marker support removed.

var (
	isCRDMarker      = markers.Must(markers.MakeDefinition("kubebuilder:resource", markers.DescribesType, crdmarkers.Resource{}))
	enablePkgMarker  = markers.Must(markers.MakeDefinition("kubebuilder:ac:generate", markers.DescribesPackage, false))
	enableTypeMarker = markers.Must(markers.MakeDefinition("kubebuilder:ac:generate", markers.DescribesType, false))
)

var importMapping = map[string]string{
	"k8s.io/apimachinery/pkg/apis/": "k8s.io/client-go/applyconfigurations/",
	"k8s.io/api/":                   "k8s.io/client-go/applyconfigurations/",
}

const importPathSuffix = "applyconfiguration"

// +controllertools:marker:generateHelp

// Generator generates code containing apply configuration type implementations.
type Generator struct {
	// HeaderFile specifies the header text (e.g. license) to prepend to generated files.
	HeaderFile string `marker:",optional"`
}

func (Generator) CheckFilter() loader.NodeFilter {
	return func(node ast.Node) bool {
		// ignore interfaces
		_, isIface := node.(*ast.InterfaceType)
		return !isIface
	}
}

func (Generator) RegisterMarkers(into *markers.Registry) error {
	if err := markers.RegisterAll(into,
		isCRDMarker, enablePkgMarker, enableTypeMarker); err != nil {
		return err
	}

	into.AddHelp(isCRDMarker,
		markers.SimpleHelp("apply", "enables apply configuration generation for this type"))
	into.AddHelp(
		enableTypeMarker, markers.SimpleHelp("apply", "overrides enabling or disabling applyconfigurations generation for the type"))

	into.AddHelp(
		enablePkgMarker, markers.SimpleHelp("apply", "overrides enabling or disabling applyconfigurations generation for the package"))
	return nil

}

func enabledOnPackage(col *markers.Collector, pkg *loader.Package) (bool, error) {
	pkgMarkers, err := markers.PackageMarkers(col, pkg)
	if err != nil {
		return false, err
	}
	pkgMarker := pkgMarkers.Get(enablePkgMarker.Name)
	if pkgMarker != nil {
		return pkgMarker.(bool), nil
	}
	return false, nil
}

// enableOnType marks whether applyconfiguration generation is enabled for the type.
func enabledOnType(info *markers.TypeInfo) bool {
	if typeMarker := info.Markers.Get(enableTypeMarker.Name); typeMarker != nil {
		return typeMarker.(bool)
	}
	return isCRD(info)
}

// isCRD marks whether the type is a CRD based on the +kubebuilder:resource marker.
func isCRD(info *markers.TypeInfo) bool {
	objectEnabled := info.Markers.Get(isCRDMarker.Name)
	return objectEnabled != nil
}

func (d Generator) Generate(ctx *genall.GenerationContext) error {
	headerFilePath := d.HeaderFile

	if headerFilePath == "" {
		tmpFile, err := os.CreateTemp("", "applyconfig-header-*.txt")
		if err != nil {
			return fmt.Errorf("failed to create temporary file: %w", err)
		}
		tmpFile.Close()

		defer os.Remove(tmpFile.Name())

		headerFilePath = tmpFile.Name()
	}

	objGenCtx := ObjectGenCtx{
		Collector:      ctx.Collector,
		Checker:        ctx.Checker,
		HeaderFilePath: headerFilePath,
	}

	for _, pkg := range ctx.Roots {
		if err := objGenCtx.generateForPackage(pkg); err != nil {
			return err
		}
	}
	return nil
}

// ObjectGenCtx contains the common info for generating apply configuration implementations.
// It mostly exists so that generating for a package can be easily tested without
// requiring a full set of output rules, etc.
type ObjectGenCtx struct {
	Collector      *markers.Collector
	Checker        *loader.TypeChecker
	HeaderFilePath string
}

type Universe struct {
	typeMetadata map[types.Type]*typeMetadata
}

type typeMetadata struct {
	used bool
}

func (u *Universe) existingApplyConfigPath(_ *types.Named, pkgPath string) (string, bool) {
	for prefix, replacePath := range importMapping {
		if strings.HasPrefix(pkgPath, prefix) {
			path := replacePath + strings.TrimPrefix(pkgPath, prefix)
			return path, true
		}
	}
	return "", false
}

func (u *Universe) IsApplyConfigGenerated(typeInfo *types.Named) bool {
	if t, ok := u.typeMetadata[typeInfo]; ok {
		return t.used
	}
	return false
}

func (u *Universe) GetApplyConfigPath(typeInfo *types.Named, pkgPath string) (string, bool) {
	isApplyConfigGenerated := u.IsApplyConfigGenerated(typeInfo)
	if path, ok := u.existingApplyConfigPath(typeInfo, pkgPath); ok {
		if isApplyConfigGenerated {
			return path, true
		}
		return pkgPath, false
	}
	// ApplyConfig is necessary but location is not explicitly specified. Assume the ApplyConfig exists at the below directory
	if isApplyConfigGenerated {
		return pkgPath + "/" + importPathSuffix, true
	}
	return pkgPath, false
}

// generateForPackage generates apply configuration implementations for
// types in the given package, writing the formatted result to given writer.
// May return nil if source could not be generated.
func (ctx *ObjectGenCtx) generateForPackage(root *loader.Package) error {
	enabled, _ := enabledOnPackage(ctx.Collector, root)
	if !enabled {
		return nil
	}

	genericArgs, _ := generatorargs.NewDefaults()
	genericArgs.InputDirs = []string{root.PkgPath}
	genericArgs.OutputPackagePath = filepath.Join(root.PkgPath, importPathSuffix)
	genericArgs.GoHeaderFilePath = ctx.HeaderFilePath

	// Make the generated header static so that it doesn't rely on the compiled binary name.
	genericArgs.GeneratedByCommentTemplate = "// Code generated by applyconfiguration-gen. DO NOT EDIT.\n"

	if err := generatorargs.Validate(genericArgs); err != nil {
		return err
	}

	b, err := genericArgs.NewBuilder()
	if err != nil {
		return err
	}

	c, err := generator.NewContext(b, applygenerator.NameSystems(), applygenerator.DefaultNameSystem())
	if err != nil {
		return err
	}

	// This allows the correct output location when GOPATH is unset.
	c.TrimPathPrefix = root.PkgPath + "/"

	pkg, ok := c.Universe[root.PkgPath]
	if !ok {
		return fmt.Errorf("package %q not found in universe", root.Name)
	}

	// For each type we think should be generated, make sure it has a genclient
	// marker else the apply generator will not generate it.
	if err := markers.EachType(ctx.Collector, root, func(info *markers.TypeInfo) {
		if !enabledOnType(info) {
			return
		}

		typ, ok := pkg.Types[info.Name]
		if !ok {
			return
		}

		comments := sets.NewString(typ.CommentLines...)
		comments.Insert(typ.SecondClosestCommentLines...)

		if !comments.Has("// +genclient") {
			typ.CommentLines = append(typ.CommentLines, "+genclient")
		}
	}); err != nil {
		return err
	}

	packages := applygenerator.Packages(c, genericArgs)
	if err := c.ExecutePackages(genericArgs.OutputBase, packages); err != nil {
		return fmt.Errorf("error executing packages: %w", err)
	}

	return nil
}
