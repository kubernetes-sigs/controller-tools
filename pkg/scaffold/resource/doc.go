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

var _ scaffold.Name = &APIsDocGo{}
var _ scaffold.Template = &APIsDocGo{}

// APIsDocGo writes the doc.go file in the pkg/apis/group/version directory
type APIsDocGo struct {
	// Resource is a resource for the API version
	*Resource

	// OutputPath is the location to write to
	OutputPath string

	// Comments are additional lines to write to the doc.go file
	Comments []string
}

// Name is the template name
func (APIsDocGo) Name() string {
	return "apis-doc-go"
}

// Path is the default path to write to
func (r *APIsDocGo) Path() string {
	dir := filepath.Join("pkg", "apis", r.Group, r.Version)
	if r.OutputPath != "" {
		dir = r.OutputPath
	}
	return filepath.Join(dir, fmt.Sprintf("doc.go"))
}

// Execute implements Template
func (r *APIsDocGo) Execute(b []byte, t *template.Template, wr func() io.WriteCloser) error {
	if len(b) > 0 {
		return nil
	}

	temp, err := t.Parse(docGoTemplate)
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

var docGoTemplate = `{{ .Boilerplate }}

// +k8s:openapi-gen=true
// +k8s:deepcopy-gen=package,register
// +k8s:conversion-gen={{ .Project.Repo }}/pkg/apis/{{ .Group }}
// +k8s:defaulter-gen=TypeMeta
// +groupName={{ .Group }}.{{ .Project.Domain }}
package {{.Version}}
`
