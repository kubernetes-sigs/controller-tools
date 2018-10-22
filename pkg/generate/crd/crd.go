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
	"log"
	"os"
	"path"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/spf13/afero"
	extensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"

	"k8s.io/gengo/types"
	"sigs.k8s.io/controller-tools/pkg/internal/codegen"
	"sigs.k8s.io/controller-tools/pkg/internal/codegen/parse"
	util "sigs.k8s.io/controller-tools/pkg/util"
)

type Options struct {
	RootPath          string
	OutputDir         string
	Domain            string
	Namespace         string
	SkipMapValidation bool
	apisPkg           string
	// OutFs is filesystem to be used for writing out the result
	OutFs afero.Fs
}

// ValidateAndInitFields validate and init generator fields.
func (c *Options) ValidateAndInitFields() (err error) {
	if c.OutFs == nil {
		c.OutFs = afero.NewOsFs()
	}

	// RootPath usually is Project path, default is current path if not specified.
	if len(c.RootPath) == 0 {
		c.RootPath, err = os.Getwd()
		if err != nil {
			return err
		}
	}

	// Validate root path is under go src path
	if !util.IsUnderGoSrcPath(c.RootPath) {
		return fmt.Errorf("command must be run from path under $GOPATH/src/")
	}

	// current PROJECT
	if len(c.Domain) == 0 {
		if !util.PathHasProjectFile(c.RootPath) {
			return fmt.Errorf("PROJECT file missing in dir %s", c.RootPath)
		}
		c.Domain = util.GetDomainFromProject(c.RootPath)
	}

	// Validate apis directory exists under working path
	apisPath := path.Join(c.RootPath, "pkg/apis")
	if _, err := os.Stat(apisPath); err != nil {
		return fmt.Errorf("error validating apis path %s: %v", apisPath, err)
	}

	c.apisPkg, err = util.DirToGoPkg(apisPath)
	if err != nil {
		return err
	}

	if c.OutputDir == "" {
		c.OutputDir = path.Join(c.RootPath, "config/crds")
	}

	return nil
}

// Run generates CRD and stores CRD into output dir.
func (c *Options) Run() error {

	apiParser, err := c.ParseAPI()
	if err != nil {
		return err
	}
	return c.writeCRDs(c.GetCRDs(apiParser))
}

// ParseAPI returns information of collection of APIs
func (c *Options) ParseAPI() (*parse.APIs, error) {
	if err := os.Chdir(c.RootPath); err != nil {
		return nil, fmt.Errorf("failed switching working dir to %s: %v", c.RootPath, err)
	}

	apiGenerator, err := codegen.DefaultGenerator()
	if err != nil {
		return nil, fmt.Errorf("failed initializing default generator : %v", err)
	}

	builder := apiGenerator.Builder
	arguments := apiGenerator.Argument
	arguments.CustomArgs = &parse.Options{SkipMapValidation: c.SkipMapValidation}

	if err := builder.AddDirRecursive("./pkg/apis"); err != nil {
		return nil, fmt.Errorf("failed making a parser: %v", err)
	}

	ctx, err := parse.NewContext(builder)
	if err != nil {
		return nil, fmt.Errorf("failed making a context: %v", err)
	}

	apiParser := parse.NewAPIs(ctx, arguments, c.Domain, c.apisPkg)
	return apiParser, nil
}

// GetCRDs returns all crds mapped by api resource name from parsed apis package
func (c *Options) GetCRDs(p *parse.APIs) map[string][]byte {
	crds := map[string]extensionsv1beta1.CustomResourceDefinition{}
	for _, g := range p.APIs.Groups {
		for _, v := range g.Versions {
			for _, r := range v.Resources {
				crd := r.CRD
				// ignore types which do not belong to this project
				if !c.belongsToAPIsPkg(r.Type) {
					continue
				}
				if len(c.Namespace) > 0 {
					crd.Namespace = c.Namespace
				}
				fileName := getCRDFileName(r)
				crds[fileName] = crd
			}
		}
	}

	result := map[string][]byte{}
	for file, crd := range crds {
		b, err := yaml.Marshal(crd)
		if err != nil {
			log.Fatalf("Error: %v", err)
		}
		result[file] = b
	}

	return result
}

func (c *Options) writeCRDs(crds map[string][]byte) error {
	// Ensure output dir exists.
	if err := c.OutFs.MkdirAll(c.OutputDir, os.FileMode(0700)); err != nil {
		return err
	}

	for file, crd := range crds {
		outFile := path.Join(c.OutputDir, file)
		if err := (&util.FileWriter{Fs: c.OutFs}).WriteFile(outFile, crd); err != nil {
			return err
		}
	}
	return nil
}

// belongsToAPIsPkg returns true if type t is defined under pkg/apis pkg of
// current project.
func (c *Options) belongsToAPIsPkg(t *types.Type) bool {
	return strings.HasPrefix(t.Name.Package, c.apisPkg)
}

func getCRDFileName(resource *codegen.APIResource) string {
	elems := []string{resource.Group, resource.Version, strings.ToLower(resource.Kind)}
	return strings.Join(elems, "_") + ".yaml"
}
