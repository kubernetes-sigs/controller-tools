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

	"log"

	flag "github.com/spf13/pflag"
	"sigs.k8s.io/controller-tools/pkg/scaffold/project"
)

// Dockerfile scaffolds a Dockerfile for building a main
type Dockerfile struct {
	// OutputPath is the output file to write
	OutputPath string

	// Project is the project
	Project project.Project
}

// Name is the name of the template
func (Dockerfile) Name() string {
	return "manager-dockerfile"
}

// Path implements scaffold.Path.  Defaults to hack/boilerplate.go.txt
func (d *Dockerfile) Path() string {
	dir := filepath.Join("Dockerfile")
	if d.OutputPath != "" {
		dir = d.OutputPath
	}
	return dir
}

// SetProject injects the Project
func (d *Dockerfile) SetProject(p project.Project) {
	d.Project = p
}

// Execute writes the template file to wr.  b is the last value of the file.  temp is a template object.
func (d *Dockerfile) Execute(b []byte, t *template.Template, wr func() io.WriteCloser) error {
	if len(b) > 0 {
		// Do nothing if the file exists
		return nil
	}
	t, err := t.Parse(dockerfileTemplate)
	if err != nil {
		return err
	}

	w := wr()
	defer func() {
		if err := w.Close(); err != nil {
			log.Fatal(err)
		}
	}()
	return t.Execute(w, d)
}

// DockerfileForFlags registers flags for Main fields and returns the Main
func DockerfileForFlags(_ *flag.FlagSet) *Dockerfile {
	return &Dockerfile{}
}

var dockerfileTemplate = `# Build and test the manager binary
FROM golang:1.9.3 as builder

# Copy in the go src
WORKDIR /go/src/{{ .Project.Repo }}
COPY pkg/    pkg/
COPY cmd/    cmd/
COPY vendor/ vendor/

# Run tests as a sanity check
ENV TEST_ASSET_DIR /usr/local/bin
ENV TEST_ASSET_KUBECTL $TEST_ASSET_DIR/kubectl
ENV TEST_ASSET_KUBE_APISERVER $TEST_ASSET_DIR/kube-apiserver
ENV TEST_ASSET_ETCD $TEST_ASSET_DIR/etcd
ENV TEST_ASSET_URL https://storage.googleapis.com/k8s-c10s-test-binaries
RUN curl ${TEST_ASSET_URL}/etcd-Linux-x86_64 --output $TEST_ASSET_ETCD
RUN curl ${TEST_ASSET_URL}/kube-apiserver-Linux-x86_64 --output $TEST_ASSET_KUBE_APISERVER
RUN curl https://storage.googleapis.com/kubernetes-release/release/v1.9.2/bin/linux/amd64/kubectl --output $TEST_ASSET_KUBECTL
RUN chmod +x $TEST_ASSET_ETCD
RUN chmod +x $TEST_ASSET_KUBE_APISERVER
RUN chmod +x $TEST_ASSET_KUBECTL
RUN go test ./pkg/... ./cmd/...

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager {{ .Project.Repo }}/cmd/manager

# Copy the controller-manager into a thin image
FROM ubuntu:latest
WORKDIR /root/
COPY --from=builder /go/src/{{ .Project.Repo }}/manager .
ENTRYPOINT ["./manager"]
`
