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
	"log"
	"path/filepath"
	"text/template"

	"sigs.k8s.io/controller-tools/pkg/scaffold"
)

var _ scaffold.Name = &RegisterGo{}
var _ scaffold.Template = &RegisterGo{}

// RegisterGo scaffolds the register.go file for defining APIsGo
type RegisterGo struct {
	// Resource is the resource to scaffold the types_test.go file for
	*Resource

	// OutputPath is the output file to write
	OutputPath string
}

// Name implements scaffold.Name
func (RegisterGo) Name() string {
	return "register-go"
}

// Path implements scaffold.Path.  Defaults to pkg/apis/group/version/register.go
func (r *RegisterGo) Path() string {
	dir := filepath.Join("pkg", "apis", r.Group, r.Version)
	if r.OutputPath != "" {
		dir = r.OutputPath
	}
	return filepath.Join(dir, fmt.Sprintf("register.go"))
}

// Execute writes the template file to wr.  b is the last value of the file.  temp is a template object.
func (r *RegisterGo) Execute(b []byte, t *template.Template, wr func() io.WriteCloser) error {
	// Don't overwrite
	if len(b) > 0 {
		return nil
	}

	temp, err := t.Parse(registerTemplate)
	if err != nil {
		return err
	}

	w := wr()
	defer func() {
		if err := w.Close(); err != nil {
			log.Fatal(err)
		}
	}()
	return temp.Execute(w, r)
}

var registerTemplate = `{{ .Boilerplate }}

// NOTE: Boilerplate only.  Ignore this file.

// +k8s:openapi-gen=true
// +k8s:deepcopy-gen=package,register
// +k8s:conversion-gen={{ .Project.Repo }}/pkg/apis/{{ .Group }}
// +k8s:defaulter-gen=TypeMeta
// +groupName={{ .Group }}.{{ .Project.Domain }}
package {{.Version}}

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var KnownTypes = []runtime.Object{}

// SchemeGroupVersion is group version used to register these objects
var SchemeGroupVersion = schema.GroupVersion{Group: "{{ .Group }}.{{ .Project.Domain }}", Version: "{{ .Version }}"}

// Kind takes an unqualified kind and returns back a Group qualified GroupKind
func Kind(kind string) schema.GroupKind {
	return SchemeGroupVersion.WithKind(kind).GroupKind()
}

// Resource takes an unqualified resource and returns a Group qualified GroupResource
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

var (
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	AddToScheme   = SchemeBuilder.AddToScheme
)

// Adds the list of known types to Scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion, KnownTypes...)
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}
`
