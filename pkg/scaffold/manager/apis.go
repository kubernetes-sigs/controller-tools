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

package manager

import (
	"io"
	"log"
	"path/filepath"
	"text/template"
)

// APIs scaffolds a apis.go to register types with a Scheme
type APIs struct {
	// OutputPath is the output file to write
	OutputPath string

	// Boilerplate is the boilerplate header to write
	Boilerplate string

	// Comments is a list of comments to add to the apis.go
	Comments []string

	// BoilerplatePath is the path to the go boilerplate file
	BoilerplatePath string
}

// Name is the name of the template
func (APIs) Name() string {
	return "manager-apis-go"
}

// Path implements scaffold.Path.  Defaults to pkg/apis/apis.go
func (m *APIs) Path() string {
	dir := filepath.Join("pkg", "apis", "apis.go")
	if m.OutputPath != "" {
		dir = m.OutputPath
	}
	return dir
}

// SetBoilerplate implements scaffold.Boilerplate.
func (m *APIs) SetBoilerplate(b string) {
	m.Boilerplate = b
}

// SetBoilerplatePath implements scaffold.BoilerplatePath.
func (m *APIs) SetBoilerplatePath(p string) {
	m.BoilerplatePath = p
}

// Execute writes the template file to wr.  b is the last value of the file.  temp is a template object.
func (m *APIs) Execute(b []byte, t *template.Template, wr func() io.WriteCloser) error {
	if len(b) > 0 {
		// Do nothing if the file exists
		return nil
	}

	if len(m.Comments) == 0 {
		m.Comments = append(m.Comments,
			"// Generate deepcopy for apis",
			"//go:generate go run ../../vendor/k8s.io/code-generator/cmd/deepcopy-gen/main.go -i ./... -h ../../"+m.BoilerplatePath)
	}

	temp, err := t.Parse(setupTemplate)
	if err != nil {
		return err
	}

	w := wr()
	defer func() {
		if err := w.Close(); err != nil {
			log.Fatal(err)
		}
	}()
	return temp.Execute(w, m)
}

var setupTemplate = `{{ .Boilerplate }}

{{ range $line := .Comments }}{{ $line }}
{{ end }}
// Package apis contains Kubernetes API groups.
package apis

import (
	"k8s.io/apimachinery/pkg/runtime"
)

// AddToSchemes should be used to add new API groupversions to the Scheme
var AddToSchemes runtime.SchemeBuilder

// AddToScheme adds all Resources to the Scheme
func AddToScheme(s *runtime.Scheme) {
	AddToSchemes.AddToScheme(s)
}
`
