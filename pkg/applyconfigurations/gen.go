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
	"bytes"
	"go/ast"
	"go/format"
	"go/types"
	"io"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
	crdmarkers "sigs.k8s.io/controller-tools/pkg/crd/markers"
	"sigs.k8s.io/controller-tools/pkg/genall"
	"sigs.k8s.io/controller-tools/pkg/loader"
	"sigs.k8s.io/controller-tools/pkg/markers"
	"sigs.k8s.io/controller-tools/pkg/util"
)

// Based on deepcopy gen but with legacy marker support removed.

var (
	groupNameMarker   = markers.Must(markers.MakeDefinition("groupName", markers.DescribesPackage, ""))
	versionNameMarker = markers.Must(markers.MakeDefinition("versionName", markers.DescribesPackage, ""))
	isCRDMarker       = markers.Must(markers.MakeDefinition("kubebuilder:resource", markers.DescribesType, crdmarkers.Resource{}))
	enablePkgMarker   = markers.Must(markers.MakeDefinition("kubebuilder:ac:generate", markers.DescribesPackage, false))
	enableTypeMarker  = markers.Must(markers.MakeDefinition("kubebuilder:ac:generate", markers.DescribesType, false))
)

var importMapping = map[string]string{
	"k8s.io/apimachinery/pkg/apis/": "k8s.io/client-go/applyconfigurations/",
	"k8s.io/api/":                   "k8s.io/client-go/applyconfigurations/",
}

const importPathSuffix = "ac"
const packageFileName = "zz_generated.applyconfigurations.go"

// +controllertools:marker:generateHelp

// Generator generates code containing apply configuration type implementations.
type Generator struct {
	// HeaderFile specifies the header text (e.g. license) to prepend to generated files.
	HeaderFile string `marker:",optional"`
	// Year specifies the year to substitute for " YEAR" in the header file.
	Year string `marker:",optional"`
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
		groupNameMarker, versionNameMarker, isCRDMarker, enablePkgMarker, enableTypeMarker); err != nil {
		return err
	}
	into.AddHelp(groupNameMarker,
		markers.SimpleHelp("apply", "specifies the API group name for this package."))

	into.AddHelp(versionNameMarker,
		markers.SimpleHelp("apply", "overrides the API group version for this package (defaults to the package name)."))
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
	if objectEnabled != nil {
		return true
	}
	return false
}

func isCRDClusterScope(info *markers.TypeInfo) bool {
	if o := info.Markers.Get(isCRDMarker.Name); o != nil {
		crd := o.(crdmarkers.Resource)
		return crd.Scope == "Cluster"
	}
	return false
}

func createApplyConfigPackage(pkg *loader.Package) *loader.Package {
	newPkg := &loader.Package{Package: &packages.Package{}}
	dir := filepath.Dir(pkg.CompiledGoFiles[0])
	newPkg.CompiledGoFiles = append(newPkg.CompiledGoFiles, dir+"/"+importPathSuffix+"/")
	return newPkg
}

func (d Generator) Generate(ctx *genall.GenerationContext) error {
	var headerText string

	if d.HeaderFile != "" {
		headerBytes, err := ctx.ReadFile(d.HeaderFile)
		if err != nil {
			return err
		}
		headerText = string(headerBytes)
	}
	headerText = strings.ReplaceAll(headerText, " YEAR", " "+d.Year)

	objGenCtx := ObjectGenCtx{
		Collector:  ctx.Collector,
		Checker:    ctx.Checker,
		HeaderText: headerText,
	}

	var pkgList []*loader.Package
	visited := make(map[string]*loader.Package)

	for _, root := range ctx.Roots {
		visited[root.PkgPath] = root
		pkgList = append(pkgList, root)
	}

	for _, pkg := range pkgList {
		for _, imp := range pkg.Imports() {
			if _, ok := visited[imp.PkgPath]; ok {
				continue
			}
			visited[imp.PkgPath] = imp
		}
	}

	universe := &Universe{typeMetadata: make(map[types.Type]*typeMetadata)}

	// Multiple traverses are required so that cross package imports are able
	// to be resolved.
	// generateEligibleTypes creates the universe for generateUsedTypes to perform a
	// breadth first search across all CRDs that need ac generation.
	// generateForPackage is final step and performs the code generation.
	for _, pkg := range visited {
		objGenCtx.generateEligibleTypes(pkg, universe)
	}
	for _, pkg := range visited {
		objGenCtx.generateUsedTypes(pkg, universe)
	}

	for _, pkg := range pkgList {
		outContents := objGenCtx.generateForPackage(universe, pkg)
		if outContents == nil {
			continue
		}
		newPkg := createApplyConfigPackage(pkg)
		writeOut(ctx, newPkg, outContents)
	}
	return nil
}

// ObjectGenCtx contains the common info for generating apply configuration implementations.
// It mostly exists so that generating for a package can be easily tested without
// requiring a full set of output rules, etc.
type ObjectGenCtx struct {
	Collector  *markers.Collector
	Checker    *loader.TypeChecker
	HeaderText string
}

// generateEligibleTypes generates a universe of all possible ApplyConfiguration types.
// The function also scans all imported packages for types that are eligible to be ApplyConfigurations.
// This first pass is necessary because the loader package is not able to follow references between packages
// and this universe constructs the necessary references.
func (ctx *ObjectGenCtx) generateEligibleTypes(root *loader.Package, universe *Universe) {
	ctx.Checker.Check(root)
	root.NeedTypesInfo()

	if err := markers.EachType(ctx.Collector, root, func(info *markers.TypeInfo) {
		// not all types required a generate apply configuration. For example, no apply configuration
		// type is needed for Quantity, IntOrString, RawExtension or Unknown.

		if shouldBeApplyConfiguration(root, info) {
			typeInfo := root.TypesInfo.TypeOf(info.RawSpec.Name)
			universe.typeMetadata[typeInfo] = &typeMetadata{
				info:     info,
				root:     root,
				eligible: true,
				used:     false,
			}
		}

	}); err != nil {
		root.AddError(err)
		return
	}
	return
}

// generateUsedTypes does a breadth first search from each top level root object
// to find all ApplyConfiguration types that must be generated based on the fields
// that the object references.
func (ctx *ObjectGenCtx) generateUsedTypes(root *loader.Package, universe *Universe) {
	ctx.Checker.Check(root)
	root.NeedTypesInfo()

	if err := markers.EachType(ctx.Collector, root, func(info *markers.TypeInfo) {
		if !enabledOnType(info) {
			return
		}

		var q []types.Type
		q = append(q, root.TypesInfo.TypeOf(info.RawSpec.Name))

		for len(q) > 0 {
			node := universe.typeMetadata[q[0]]
			q = q[1:]
			if node.used {
				continue
			}
			node.used = true
			if len(node.info.Fields) > 0 {
				for _, field := range node.info.Fields {
					fieldType := node.root.TypesInfo.TypeOf(field.RawField.Type)
					resolved := false
					// TODO: Are these all the types that need to be resolved?
					for !resolved {
						resolved = true
						switch typeInfo := fieldType.(type) {
						case *types.Pointer:
							fieldType = typeInfo.Elem()
							resolved = false
						case *types.Slice:
							fieldType = typeInfo.Elem()
							resolved = false
						}
					}

					if _, ok := universe.typeMetadata[fieldType]; ok {
						q = append(q, fieldType)
					}
				}
			}
		}
	}); err != nil {
		root.AddError(err)
		return
	}
	return
}

type Universe struct {
	typeMetadata map[types.Type]*typeMetadata
}

type typeMetadata struct {
	info     *markers.TypeInfo
	root     *loader.Package
	eligible bool
	used     bool
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

func (u *Universe) IsApplyConfigGenerated(typeInfo *types.Named, pkgPath string) bool {
	if t, ok := u.typeMetadata[typeInfo]; ok {
		return t.used
	}
	return false
}

func (u *Universe) GetApplyConfigPath(typeInfo *types.Named, pkgPath string) (string, bool) {
	isApplyConfigGenerated := u.IsApplyConfigGenerated(typeInfo, pkgPath)
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
func (ctx *ObjectGenCtx) generateForPackage(universe *Universe, root *loader.Package) []byte {
	pkgMarkers, err := markers.PackageMarkers(ctx.Collector, root)
	if err != nil {
		root.AddError(err)
	}
	group := ""
	if val := pkgMarkers.Get("groupName"); val != nil {
		group = val.(string)
	}

	version := root.Name
	if val := pkgMarkers.Get("versionName"); val != nil {
		version = val.(string)
	}

	byType := make(map[string][]byte)
	imports := util.NewImportsList(root)

	enabled, _ := enabledOnPackage(ctx.Collector, root)
	if !enabled {
		return nil
	}

	if err := markers.EachType(ctx.Collector, root, func(info *markers.TypeInfo) {
		outContent := new(bytes.Buffer)

		if t, ok := universe.typeMetadata[root.TypesInfo.TypeOf(info.RawSpec.Name)]; ok {
			if !t.used {
				return
			}
		}

		if !shouldBeApplyConfiguration(root, info) {
			return
		}

		copyCtx := &applyConfigurationMaker{
			pkg:         root,
			ImportsList: imports,
			codeWriter:  &codeWriter{out: outContent},
		}

		copyCtx.GenerateTypesFor(universe, root, info)
		for _, field := range info.Fields {
			if field.Name != "" {
				switch root.TypesInfo.TypeOf(field.RawField.Type).(type) {
				case *types.Slice:
					copyCtx.GenerateMemberSetForSlice(universe, field, root, info)
				case *types.Map:
					copyCtx.GenerateMemberSetForMap(universe, field, root, info)
				default:
					copyCtx.GenerateMemberSet(universe, field, root, info)
				}
			}
		}

		if enabledOnType(info) {
			copyCtx.GenerateRootStructConstructor(root, info, isCRDClusterScope(info), group, version)
		} else {
			copyCtx.GenerateStructConstructor(root, info)
		}

		if enabledOnType(info) {
			copyCtx.GenerateRootFunctions(universe, root, info)
		}

		outBytes := outContent.Bytes()
		if len(outBytes) > 0 {
			byType[info.Name] = outBytes
		}
	}); err != nil {
		root.AddError(err)
		return nil
	}

	if len(byType) == 0 {
		return nil
	}

	outContent := new(bytes.Buffer)
	util.WriteHeader(root, outContent, root.Name, imports, ctx.HeaderText)
	writeTypes(root, outContent, byType)

	outBytes := outContent.Bytes()
	formattedBytes, err := format.Source(outBytes)
	if err != nil {
		root.AddError(err)
		// we still write the invalid source to disk to figure out what went wrong
	} else {
		outBytes = formattedBytes
	}

	return outBytes
}

// writeTypes writes each method to the file, sorted by type name.
func writeTypes(pkg *loader.Package, out io.Writer, byType map[string][]byte) {
	sortedNames := make([]string, 0, len(byType))
	for name := range byType {
		sortedNames = append(sortedNames, name)
	}
	sort.Strings(sortedNames)

	for _, name := range sortedNames {
		_, err := out.Write(byType[name])
		if err != nil {
			pkg.AddError(err)
		}
	}
}

// writeFormatted outputs the given code, after gofmt-ing it.  If we couldn't gofmt,
// we write the unformatted code for debugging purposes.
func writeOut(ctx *genall.GenerationContext, root *loader.Package, outBytes []byte) {
	outputFile, err := ctx.Open(root, packageFileName)
	if err != nil {
		root.AddError(err)
		return
	}
	defer outputFile.Close()
	n, err := outputFile.Write(outBytes)
	if err != nil {
		root.AddError(err)
		return
	}
	if n < len(outBytes) {
		root.AddError(io.ErrShortWrite)
	}
}
