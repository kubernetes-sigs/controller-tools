/*
Copyright 2017 The Kubernetes Authors.

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

package generator

import (
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"
	"text/template"

	"k8s.io/gengo/generator"
	"k8s.io/gengo/types"
	"sigs.k8s.io/controller-tools/pkg/internal/codegen"
)

type crdGenerator struct {
	generator.DefaultGen
	apiversion *codegen.APIVersion
	apigroup   *codegen.APIGroup
}

var _ generator.Generator = &crdGenerator{}

func (d *crdGenerator) Imports(c *generator.Context) []string {
	imports := []string{
		"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1",
		"metav1 \"k8s.io/apimachinery/pkg/apis/meta/v1\"",
		"k8s.io/apimachinery/pkg/runtime",
		"k8s.io/apimachinery/pkg/runtime/schema",
	}
	if d.apigroup.Pkg != nil {
		imports = append(imports, d.apigroup.Pkg.Path)
	}
	return imports
}

func (d *crdGenerator) Finalize(context *generator.Context, w io.Writer) error {
	temp := template.Must(template.New("crd-template").Parse(crdTemplate))
	return temp.Execute(w, d.apiversion)
}

// generatorToPackage creates a new package from a generator and package name
func generatorToPackage(pkg string, gen generator.Generator) generator.Package {
	name := strings.Split(filepath.Base(pkg), ".")[0]
	return &generator.DefaultPackage{
		PackageName: name,
		PackagePath: pkg,
		HeaderText:  append(generatedGoHeader(), []byte("\n\n")...),
		GeneratorFunc: func(c *generator.Context) (generators []generator.Generator) {
			return []generator.Generator{gen}
		},
		FilterFunc: func(c *generator.Context, t *types.Type) bool {
			// Generators only see Types in the same package as the generator
			return t.Name.Package == pkg
		},
	}
}

// generatedGoHeader returns the header to preprend to generated go files
func generatedGoHeader() []byte {
	cr, err := ioutil.ReadFile(filepath.Join("hack", "boilerplate.go.txt"))
	if err != nil {
		return []byte{}
	}
	return cr
}

// packages wraps a collection of generator.Packages
type packages struct {
	value generator.Packages
}

// add creates a new generator.Package from gen and adds it to the collection
func (g *packages) add(pkg string, gen generator.Generator) {
	g.value = append(g.value, generatorToPackage(pkg, gen))
}

var crdTemplate = `
// CRD Generation
func getFloat(f float64) *float64 {
    return &f
}

func getInt(i int64) *int64 {
    return &i
}

var (
    {{ range $api := .Resources -}}
    // {{.Kind}}CRD resource
    {{.Kind}}CRD = v1beta1.CustomResourceDefinition{
        ObjectMeta: metav1.ObjectMeta{
            Name: "{{.Resource}}.{{.Group}}.{{.Domain}}",
        },
        Spec: v1beta1.CustomResourceDefinitionSpec {
            Group: "{{.Group}}.{{.Domain}}",
            Version: "{{.Version}}",
            Names: v1beta1.CustomResourceDefinitionNames{
                Kind: "{{.Kind}}",
                Plural: "{{.Resource}}",
                {{ if .ShortName -}}
                ShortNames: []string{"{{.ShortName}}"},
                {{ end -}}
                {{ if .Categories -}}
                Categories: []string{
                {{ range .Categories -}}
                    "{{ . }}",
                {{ end -}}
                },
                {{ end -}}
            },
            {{ if .NonNamespaced -}}
            Scope: "Cluster",
            {{ else -}}
            Scope: "Namespaced",
            {{ end -}}
            Validation: &v1beta1.CustomResourceValidation{
                OpenAPIV3Schema: &{{.Validation}},
            },
            Subresources: &v1beta1.CustomResourceSubresources{
                {{ if .HasStatusSubresource -}}
                Status: &v1beta1.CustomResourceSubresourceStatus{},
                {{ end -}}
                {{ if .HasScaleSubresource -}}
                Scale: &v1beta1.CustomResourceSubresourceScale{
                    SpecReplicasPath: "{{ .CRD.Spec.Subresources.Scale.SpecReplicasPath }}",
                    StatusReplicasPath: "{{ .CRD.Spec.Subresources.Scale.StatusReplicasPath }}",
                    {{ if .CRD.Spec.Subresources.Scale.LabelSelectorPath -}}
                    LabelSelectorPath: {{ .CRD.Spec.Subresources.Scale.LabelSelectorPath }},
                    {{ end -}}
                },
                {{ end -}}
            },
        },
    }
    {{ end -}}
)
`
