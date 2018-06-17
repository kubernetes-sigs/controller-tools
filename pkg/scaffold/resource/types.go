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

package resource

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"text/template"

	"log"

	"sigs.k8s.io/controller-tools/pkg/scaffold"
)

var _ scaffold.Template = &Types{}
var _ scaffold.Name = &Types{}

// Types scaffolds the types.go file for defining APIsGo
type Types struct {
	// Resource is the resource to scaffold the types_test.go file for
	*Resource
	// OutputPath is the output file to write
	OutputPath string
}

// Name implements scaffold.Name
func (Types) Name() string {
	return "types-go"
}

// Path implements scaffold.Path.  Defaults to pkg/apis/group/version/kind_types.go
func (t *Types) Path() string {
	dir := filepath.Join("pkg", "apis", t.Group, t.Version)
	if t.OutputPath != "" {
		dir = t.OutputPath
	}

	return filepath.Join(dir, fmt.Sprintf("%s_types.go", strings.ToLower(t.Kind)))
}

// Execute writes the template file to wr.  b is the last value of the file.  temp is a template object.
func (t *Types) Execute(b []byte, temp *template.Template, wr func() io.WriteCloser) error {
	if len(b) > 0 {
		return fmt.Errorf("%s already exists", t.Path())
	}

	temp, err := temp.Parse(typesTemplate)
	if err != nil {
		return err
	}

	w := wr()
	defer func() {
		if err := w.Close(); err != nil {
			log.Fatal(err)
		}
	}()
	return temp.Execute(w, t)
}

var typesTemplate = `{{ .Boilerplate }}

package {{.Version}}

import (
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// {{.Kind}}Spec defines the desired state of {{.Kind}}
type {{.Kind}}Spec struct {
    // INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "kubebuilder generate" to regenerate code after modifying this file
}

// {{.Kind}}Status defines the observed state of {{.Kind}}
type {{.Kind}}Status struct {
    // INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "kubebuilder generate" to regenerate code after modifying this file
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
{{- if .Namespaced }} {{ else }}
// +genclient:nonNamespaced
{{- end }}

// {{.Kind}}
// +k8s:openapi-gen=true
type {{.Kind}} struct {
    metav1.TypeMeta   ` + "`" + `json:",inline"` + "`" + `
    metav1.ObjectMeta ` + "`" + `json:"metadata,omitempty"` + "`" + `

    Spec   {{.Kind}}Spec   ` + "`" + `json:"spec,omitempty"` + "`" + `
    Status {{.Kind}}Status ` + "`" + `json:"status,omitempty"` + "`" + `
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
{{- if .Namespaced }} {{ else }}
// +genclient:nonNamespaced
{{- end }}

type {{.Kind}}List struct {
	metav1.TypeMeta ` + "`" + `json:",inline"` + "`" + `
	metav1.ListMeta ` + "`" + `json:"metadata,omitempty"` + "`" + `
	Items           []{{ .Kind }} ` + "`" + `json:"items"` + "`" + `
}

func init() {
	KnownTypes = append(KnownTypes, &{{.Kind}}{}, &{{.Kind}}List{})
}
`
