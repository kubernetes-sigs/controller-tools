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

package crd

import (
	"fmt"
	"go/types"

	"k8s.io/apimachinery/pkg/runtime/schema"

	crdmarkers "sigs.k8s.io/controller-tools/pkg/crd/markers"
	"sigs.k8s.io/controller-tools/pkg/genall"
	"sigs.k8s.io/controller-tools/pkg/loader"
	"sigs.k8s.io/controller-tools/pkg/markers"
)

// Generator is a genall.Generator that generates CRDs.
type Generator struct{}

func (Generator) RegisterMarkers(into *markers.Registry) error {
	return crdmarkers.Register(into)
}
func (Generator) Generate(ctx *genall.GenerationContext) error {
	parser := &Parser{
		Collector: ctx.Collector,
		Checker:   ctx.Checker,
	}

	AddKnownTypes(parser)
	for _, root := range ctx.Roots {
		parser.NeedPackage(root)
	}

	metav1Pkg := findMetav1(ctx.Roots)
	if metav1Pkg == nil {
		// no objects in the roots, since nothing imported metav1
		return nil
	}

	// TODO: allow selecting a specific object
	kubeKinds := findKubeKinds(parser, metav1Pkg)
	if len(kubeKinds) == 0 {
		// no objects in the roots
		return nil
	}

	for _, groupKind := range kubeKinds {
		parser.NeedCRDFor(groupKind)
		crd := parser.CustomResourceDefinitions[groupKind]
		fileName := fmt.Sprintf("%s_%s.yaml", crd.Spec.Group, crd.Spec.Names.Plural)
		if err := ctx.WriteYAML(crd, fileName); err != nil {
			return err
		}
	}

	return nil
}

// findMetav1 locates the actual package representing metav1 amongst
// the imports of the roots.
func findMetav1(roots []*loader.Package) *loader.Package {
	for _, root := range roots {
		pkg := root.Imports()["k8s.io/apimachinery/pkg/apis/meta/v1"]
		if pkg != nil {
			return pkg
		}
	}
	return nil
}

// findKubeKinds locates all types that contain TypeMeta and ObjectMeta
// (and thus may be a Kubernetes object), and returns the corresponding
// group-kinds.
func findKubeKinds(parser *Parser, metav1Pkg *loader.Package) []schema.GroupKind {
	// TODO(directxman12): technically, we should be finding metav1 per-package
	var kubeKinds []schema.GroupKind
	for typeIdent, info := range parser.Types {
		hasObjectMeta := false
		hasTypeMeta := false

		pkg := typeIdent.Package
		pkg.NeedTypesInfo()
		typesInfo := pkg.TypesInfo

		for _, field := range info.Fields {
			if field.Name != "" {
				// type and object meta are embedded,
				// so they can't be this
				continue
			}

			fieldType := typesInfo.TypeOf(field.RawField.Type)
			namedField, isNamed := fieldType.(*types.Named)
			if !isNamed {
				// ObjectMeta and TypeMeta are named types
				continue
			}
			fieldPkgPath := loader.NonVendorPath(namedField.Obj().Pkg().Path())
			fieldPkg := pkg.Imports()[fieldPkgPath]
			if fieldPkg != metav1Pkg {
				fmt.Printf("field %q's package is %q (%v), not metav1", field.Name, namedField.Obj().Pkg().Path(), fieldPkg)
				continue
			}

			switch namedField.Obj().Name() {
			case "ObjectMeta":
				hasObjectMeta = true
			case "TypeMeta":
				hasTypeMeta = true
			}
		}

		if !hasObjectMeta || !hasTypeMeta {
			continue
		}

		groupKind := schema.GroupKind{
			Group: parser.GroupVersions[pkg].Group,
			Kind:  typeIdent.Name,
		}
		kubeKinds = append(kubeKinds, groupKind)
	}

	return kubeKinds
}
