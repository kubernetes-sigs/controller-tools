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

package crdschema

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"strings"

	"gopkg.in/yaml.v2"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	crdgen "sigs.k8s.io/controller-tools/pkg/crd"
	crdmarkers "sigs.k8s.io/controller-tools/pkg/crd/markers"
	crdschemayaml "sigs.k8s.io/controller-tools/pkg/crdschema/internal/yaml"
	"sigs.k8s.io/controller-tools/pkg/genall"
	"sigs.k8s.io/controller-tools/pkg/markers"
)

var (
	scheme = runtime.NewScheme()
	codecs = serializer.NewCodecFactory(scheme)
)

func init() {
	utilruntime.Must(apiextensionsv1beta1.AddToScheme(scheme))
}

// Generator is a genall.Generator that generates CRD schemas.
type Generator struct {
	// ManifestsPath is a path with CustomResourceDefinition manifest files
	ManifestsPath string `marker:"manifests,optional"`
	// If true, no file is written. The Generate func will return an
	// error if a file on disk differs from the generated manifest.
	VerifyOnly bool `marker:",optional"`
}

var _ genall.Generator = &Generator{}

func (Generator) RegisterMarkers(into *markers.Registry) error {
	return crdmarkers.Register(into)
}

func (g Generator) Generate(ctx *genall.GenerationContext) (result error) {
	parser := &crdgen.Parser{
		Collector: ctx.Collector,
		Checker:   ctx.Checker,
	}
	flattener := crdgen.Flattener{
		Parser: parser,
	}

	crdgen.AddKnownTypes(parser)
	for _, root := range ctx.Roots {
		parser.NeedPackage(root)
	}

	metav1Pkg := crdgen.FindMetav1(ctx.Roots)
	if metav1Pkg == nil {
		// no objects in the roots, since nothing imported metav1
		return fmt.Errorf("no metav1 import found in any of the API packages")
	}

	// load existing manifests from manifests/ dir
	existingCRDsByFileNames, err := crdsFromDirectory(ctx, g.ManifestsPath)
	if err != nil {
		return err
	}
	existingFileNamesByGroupKind := map[schema.GroupKind]string{}
	for fn, crd := range existingCRDsByFileNames {
		existingFileNamesByGroupKind[crd.GroupKind] = fn
	}

	changedFiles := []string{}
	for _, groupKind := range crdgen.FindKubeKinds(parser, metav1Pkg) {
		fileName, existsAsFile := existingFileNamesByGroupKind[groupKind]
		if !existsAsFile {
			continue
		}
		existing := existingCRDsByFileNames[fileName]

		// get existing versions
		existingVersions := make(map[string]bool)
		if existing.CRD.Spec.Version != "" {
			existingVersions[existing.CRD.Spec.Version] = true
		}
		for _, v := range existing.CRD.Spec.Versions {
			existingVersions[v.Name] = true
		}

		// go through source code versions
		versionSchemas := map[string]*apiextensionsv1beta1.JSONSchemaProps{}
		for pkg, gv := range parser.GroupVersions {
			if gv.Group != groupKind.Group {
				continue
			}
			if !existingVersions[gv.Version] {
				continue
			}
			versionSchemas[gv.Version] = nil

			// find parser schema
			typeIdent := crdgen.TypeIdent{Package: pkg, Name: groupKind.Kind}
			typeInfo := parser.Types[typeIdent]
			if typeInfo == nil {
				continue
			}

			// generate flattened schema
			parser.NeedSchemaFor(typeIdent)
			versionSchemas[gv.Version] = crdgen.FlattenEmbedded(flattener.FlattenType(typeIdent), pkg)
		}

		// edge cases: no schema or all nil
		allNil := true
		for _, s := range versionSchemas {
			if s != nil {
				allNil = false
				break
			}
		}
		if allNil {
			continue
		}

		// use global schema if there is just one, or all are equal
		allEqual := true
		var first *apiextensionsv1beta1.JSONSchemaProps
		for _, s := range versionSchemas {
			first = s
			break
		}
		for _, s := range versionSchemas {
			if !reflect.DeepEqual(s, first) {
				allEqual = false
				break
			}
		}

		// delete old and set new schemas
		if allEqual {
			if err := existing.setGlobalSchema(first); err != nil {
				return fmt.Errorf("failed to set global schema for %s in %s: %v", groupKind.Kind, groupKind.Group, err)
			}
		} else {
			if err := existing.setVersionedSchema(versionSchemas); err != nil {
				return fmt.Errorf("failed to set versioned schemas for %s in %s: %v", groupKind.Kind, groupKind.Group, err)
			}
		}

		// check whether the file changes
		if changed, err := existing.compareToExistingFile(ctx, fileName); err != nil {
			return fmt.Errorf("failed to compare against %q: %v", fileName, err)
		} else if changed {
			changedFiles = append(changedFiles, fileName)
		}
	}

	// only check for differences in verify-only mode
	if g.VerifyOnly {
		if len(changedFiles) > 0 {
			return fmt.Errorf("unexpected changes to: %s", strings.Join(changedFiles, ", "))
		}
		return nil
	}

	return g.writeChangedFiles(ctx, changedFiles, existingCRDsByFileNames)
}

func (g Generator) writeChangedFiles(ctx *genall.GenerationContext, changedFiles []string, existingCRDsByFileNames map[string]*existingCRD) (result error) {
	// write changes
	for _, fn := range changedFiles {
		bs, err := yaml.Marshal(existingCRDsByFileNames[fn].Yaml)
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(g.ManifestsPath, fn)
		if err != nil {
			return err
		}

		w, err := ctx.OutputRule.Open(nil, rel)
		if err != nil {
			return err
		}
		if _, err := w.Write(bs); err != nil {
			return err
		}
		if err := w.Close(); err != nil {
			return err
		}
	}

	return nil
}

type existingCRD struct {
	GroupKind schema.GroupKind
	Yaml      interface{}
	CRD       *apiextensionsv1beta1.CustomResourceDefinition
}

func (e *existingCRD) compareToExistingFile(ctx *genall.GenerationContext, fn string) (bool, error) {
	bs, err := ctx.ReadFile(fn)
	if err != nil {
		return false, err
	}
	var old yaml.MapSlice
	if err := yaml.Unmarshal(bs, &old); err != nil {
		return false, err
	}
	y := e.Yaml.(yaml.MapSlice)
	return !reflect.DeepEqual(old, y), nil
}

func (e *existingCRD) setGlobalSchema(schema *apiextensionsv1beta1.JSONSchemaProps) error {
	y, err := crdschemayaml.ToYAML(schema)
	if err != nil {
		return err
	}

	if e.Yaml, err = crdschemayaml.SetNestedField(e.Yaml, y, "spec", "validation", "openAPIV3Schema"); err != nil {
		return err
	}

	versions, found, err := crdschemayaml.GetNestedFieldSliceNoCopy(e.Yaml, "spec", "versions")
	if err != nil {
		return err
	}
	if found {
		for i := range versions {
			versions[i], err = crdschemayaml.DeleteNestedField(versions[i], "schema")
			if err != nil {
				return fmt.Errorf("spec.versions[%d].%s", i, err)
			}
		}
		e.Yaml, err = crdschemayaml.SetNestedField(e.Yaml, versions, "spec", "versions")
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *existingCRD) setVersionedSchema(versionSchemas map[string]*apiextensionsv1beta1.JSONSchemaProps) error {
	var err error
	e.Yaml, err = crdschemayaml.DeleteNestedField(e.Yaml, "spec", "validation")
	if err != nil {
		return err
	}
	versions, found, err := crdschemayaml.GetNestedFieldSliceNoCopy("spec", "versions")
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("unexpected missing versions")
	}
	for i := range versions {
		name, _, _ := crdschemayaml.GetNestedFieldString(versions[i], "name")
		if name == "" {
			return fmt.Errorf("unexpected empty name at spec.versions[%d]", i)
		}
		y, err := crdschemayaml.ToYAML(versionSchemas[name])
		if err != nil {
			return fmt.Errorf("failed to convert schema to YAML: %v", err)
		}

		if y == nil {
			versions[i], err = crdschemayaml.DeleteNestedField(versions[i], "schema")
			if err != nil {
				return fmt.Errorf("spec.versions[%s].%s", name, err)
			}
		} else {
			versions[i], err = crdschemayaml.SetNestedField(versions[i], y, "schema", "openAPIV3Schema")
			if err != nil {
				return fmt.Errorf("spec.versions[%s].%s", name, err)
			}
		}
	}
	e.Yaml, err = crdschemayaml.SetNestedField(e.Yaml, versions, "spec", "versions")
	if err != nil {
		return err
	}
	return nil
}

// crdsFromDirectory returns CRDs by file path
func crdsFromDirectory(ctx *genall.GenerationContext, dir string) (map[string]*existingCRD, error) {
	ret := map[string]*existingCRD{}
	infos, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, info := range infos {
		if info.IsDir() {
			continue
		}
		if !strings.HasSuffix(info.Name(), ".yaml") {
			continue
		}
		bs, err := ctx.ReadFile(filepath.Join(dir, info.Name()))
		if err != nil {
			return nil, err
		}

		obj, _, err := codecs.UniversalDeserializer().Decode(bs, nil, nil)
		if err != nil {
			continue
		}
		crd, ok := obj.(*apiextensionsv1beta1.CustomResourceDefinition)
		if !ok {
			continue
		}

		var y yaml.MapSlice
		if err := yaml.Unmarshal(bs, &y); err != nil {
			continue
		}
		gk := schema.GroupKind{Group: crd.Spec.Group, Kind: crd.Spec.Names.Kind}
		ret[filepath.Join(dir, info.Name())] = &existingCRD{gk, y, crd}
	}
	return ret, nil
}
