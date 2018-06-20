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

	"sigs.k8s.io/controller-tools/pkg/scaffold"
	"sigs.k8s.io/controller-tools/pkg/scaffold/project"
	"sigs.k8s.io/controller-tools/pkg/scaffold/util"
)

var _ scaffold.Name = &AddResource{}
var _ scaffold.Template = &AddResource{}

// RoleBinding scaffolds the a role binding for RBAC permissions to a CRD
type RoleBinding struct {
	// OutputPath is the output file to write
	OutputPath string

	// Resource is a resource in the API group
	*Resource

	// Project is the project
	Project project.Project
}

// Name implements scaffold.Name
func (RoleBinding) Name() string {
	return "role-resource-yaml"
}

// Path implements scaffold.Path.  Defaults to cmd/manager/setup/group_version_kind_init
func (r RoleBinding) Path() string {
	dir := filepath.Join("config", "manager", fmt.Sprintf(
		"%s_rolebinding_rbac.yaml", r.Group))
	if r.OutputPath != "" {
		dir = r.OutputPath
	}
	return dir
}

// SetProject injects the Project
func (r *RoleBinding) SetProject(p project.Project) {
	r.Project = p
}

// Execute writes the template file to wr.  b is the last value of the file.  temp is a template object.
func (r RoleBinding) Execute(b []byte, t *template.Template, wr func() io.WriteCloser) error {
	// Already exists, do nothing
	if len(b) > 0 {
		return nil
	}

	temp, err := t.Parse(roleBindingTemplate)
	if err != nil {
		return err
	}
	return util.WriteTemplate(temp, r, wr)
}

var roleBindingTemplate = `apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  labels:
    controller-tools.k8s.io: "1.0"
  name: {{ .Group }}-rolebinding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: {{ .Group }}-role
subjects:
- kind: ServiceAccount
  name: default
`
