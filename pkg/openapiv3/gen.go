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

package openapiv3

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	kyaml "sigs.k8s.io/yaml"

	crdmarkers "sigs.k8s.io/controller-tools/pkg/crd/markers"
	"sigs.k8s.io/controller-tools/pkg/genall"
	"sigs.k8s.io/controller-tools/pkg/markers"
)

// +controllertools:marker:generateHelp

// Generator outputs OpenAPI validation spec for the CRDs given as input.
type Generator struct {
	// ManifestsPath contains the CustomResourceDefinition YAML files.
	ManifestsPath string `marker:"manifests"`

	// MaxDescLen specifies the maximum description length for fields in CRD's OpenAPI schema.
	//
	// 0 indicates drop the description for all fields completely.
	// n indicates limit the description to at most n characters and truncate the description to
	// closest sentence boundary if it exceeds n characters.
	MaxDescLen *int `marker:",optional"`
}

var _ genall.Generator = &Generator{}

func (Generator) RegisterMarkers(into *markers.Registry) error {
	return crdmarkers.Register(into)
}

func (g Generator) Generate(ctx *genall.GenerationContext) (result error) {
	// load existing CRD manifests with group-kind and versions
	crdList, err := renderCRDs(ctx, g.ManifestsPath)
	if err != nil {
		return err
	}

	for _, crd := range crdList {
		for _, version := range crd.Spec.Versions {
			outWriter, err := ctx.OutputRule.Open(nil, outputfileName(crd, version))
			if err != nil {
				return err
			}
			defer outWriter.Close()

			out := map[string]interface{}{
				"components": map[string]interface{}{
					"schemas": map[string]*apiextensionsv1.JSONSchemaProps{
						crd.Spec.Names.Kind: version.Schema.OpenAPIV3Schema,
					},
				},
			}

			b, err := kyaml.Marshal(out)
			if err != nil {
				return err
			}
			outWriter.Write(b)
		}
	}

	return nil
}

func outputfileName(crd *apiextensionsv1.CustomResourceDefinition, version apiextensionsv1.CustomResourceDefinitionVersion) string {
	return fmt.Sprintf("%s_%s_%s.openapi.yaml", crd.Spec.Group, crd.Spec.Names.Plural, version.Name)
}

// renderCRDs iterate through options.Paths and extract all CRD files.
func renderCRDs(ctx *genall.GenerationContext, path string) ([]*apiextensionsv1.CustomResourceDefinition, error) {
	var (
		err   error
		info  os.FileInfo
		files []os.FileInfo
	)

	type GVKN struct {
		GVK  schema.GroupVersionKind
		Name string
	}

	var filePath = path

	// Return the error if ErrorIfPathMissing exists
	if info, err = os.Stat(path); os.IsNotExist(err) {
		return nil, err
	}

	if !info.IsDir() {
		filePath, files = filepath.Dir(path), []os.FileInfo{info}
	} else {
		if files, err = ioutil.ReadDir(path); err != nil {
			return nil, err
		}
	}

	return readCRDs(ctx, filePath, files)
}

// readCRDs reads the CRDs from files and Unmarshals them into structs
func readCRDs(ctx *genall.GenerationContext, dir string, files []os.FileInfo) ([]*apiextensionsv1.CustomResourceDefinition, error) {
	var crds []*apiextensionsv1.CustomResourceDefinition

	// Allowlist of file extensions that may contain CRDs.
	crdExts := sets.NewString(".json", ".yaml", ".yml")

	for _, file := range files {
		// Only parse allowlisted file types
		if !crdExts.Has(filepath.Ext(file.Name())) {
			continue
		}

		rawContent, err := ctx.ReadFile(filepath.Join(dir, file.Name()))
		if err != nil {
			return nil, err
		}

		// Unmarshal CRDs from file into structs
		crd := &apiextensionsv1.CustomResourceDefinition{}
		if err := kyaml.Unmarshal(rawContent, crd); err != nil {
			return nil, err
		}
		crds = append(crds, crd)
	}
	return crds, nil
}
