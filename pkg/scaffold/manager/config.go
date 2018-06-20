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
	"path/filepath"
	"text/template"

	"sigs.k8s.io/controller-tools/pkg/scaffold/util"
)

// Config scaffolds yaml config for the manager.
type Config struct {
	// OutputPath is the output file to write
	OutputPath string
}

// Name is the name of the template
func (Config) Name() string {
	return "manager-config-yaml"
}

// Path implements scaffold.Path.  Defaults to pkg/apis/apis.go
func (m *Config) Path() string {
	dir := filepath.Join("config", "manager", "manager.yaml")
	if m.OutputPath != "" {
		dir = m.OutputPath
	}
	return dir
}

// Execute writes the template file to wr.  b is the last value of the file.  temp is a template object.
func (m *Config) Execute(b []byte, t *template.Template, wr func() io.WriteCloser) error {
	if len(b) > 0 {
		// Do nothing if the file exists
		return nil
	}

	temp, err := t.Parse(configTemplate)
	if err != nil {
		return err
	}
	return util.WriteTemplate(temp, m, wr)
}

var configTemplate = `apiVersion: v1
kind: Namespace
metadata:
  labels:
      controller-tools.k8s.io: "1.0"
  name: system
---
apiVersion: v1
kind: Service
metadata:
  name: controller-manager-service
  labels:
    control-plane: controller-manager
    controller-tools.k8s.io: "1.0"
spec:
  selector:
    control-plane: controller-manager
    controller-tools.k8s.io: "1.0"
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: controller-manager
  labels:
      control-plane: controller-manager
      controller-tools.k8s.io: "1.0"
spec:
  selector:
    matchLabels:
      control-plane: controller-manager
      controller-tools.k8s.io: "1.0"
  serviceName: controller-manager-service
  template:
    metadata:
      labels:
        control-plane: controller-manager
        controller-tools.k8s.io: "1.0"
    spec:
      containers:
        command:
        - /root/manager
        name: manager
        resources:
          limits:
            cpu: 100m
            memory: 30Mi
          requests:
            cpu: 100m
            memory: 20Mi
      terminationGracePeriodSeconds: 10
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  labels:
    controller-tools.k8s.io: "1.0"
  name: rolebinding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: role
subjects:
- kind: ServiceAccount
  name: default
`
