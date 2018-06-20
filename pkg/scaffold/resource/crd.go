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
	"text/template"

	"strings"

	"github.com/markbates/inflect"
	"sigs.k8s.io/controller-tools/pkg/scaffold"
	"sigs.k8s.io/controller-tools/pkg/scaffold/project"
	"sigs.k8s.io/controller-tools/pkg/scaffold/util"
)

var _ scaffold.Name = &AddResource{}
var _ scaffold.Template = &AddResource{}

// CRD scaffolds the manager init code.
type CRD struct {
	// OutputPath is the output file to write
	OutputPath string

	// Scope is Namespaced or Cluster
	Scope string

	// Plural is the plural lowercase of kind
	Plural string

	// Resource is a resource in the API group
	*Resource

	// Project is the project
	Project project.Project

	// PrintCreated will print the file it creates
	PrintCreated bool
}

// Name implements scaffold.Name
func (CRD) Name() string {
	return "crd-resource-yaml"
}

// Path implements scaffold.Path.  Defaults to cmd/manager/setup/group_version_kind_init
func (a CRD) Path() string {
	dir := filepath.Join("config", "crds", fmt.Sprintf(
		"%s_%s_%s.yaml", a.Group, a.Version, strings.ToLower(a.Kind)))
	if a.OutputPath != "" {
		dir = a.OutputPath
	}
	return dir
}

// SetProject injects the Project
func (a *CRD) SetProject(p project.Project) {
	a.Project = p
}

// Execute writes the template file to wr.  b is the last value of the file.  temp is a template object.
func (a CRD) Execute(b []byte, t *template.Template, wr func() io.WriteCloser) error {
	// Already exists, do nothing
	if len(b) > 0 {
		return nil
	}

	a.Scope = "Namespaced"
	if !a.Namespaced {
		a.Scope = "Cluster"
	}
	if a.Plural == "" {
		a.Plural = strings.ToLower(inflect.Pluralize(a.Kind))
	}
	temp, err := t.Parse(crdTemplate)
	if err != nil {
		return err
	}
	return util.WriteTemplate(temp, a, wr)
}

var crdTemplate = `apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  labels:
    controller-tools.k8s.io: "1.0"
  name: {{ .Plural }}.{{ .Group }}.{{ .Project.Domain }}
spec:
  group: {{ .Group }}.{{ .Project.Domain }}
  version: "{{ .Version }}"
  names:
    kind: {{ .Kind }}
    plural: {{ .Plural }}
  scope: {{ .Scope }}
`
