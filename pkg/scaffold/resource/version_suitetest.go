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

	"log"

	"sigs.k8s.io/controller-tools/pkg/scaffold"
)

var _ scaffold.Template = &TypesTest{}
var _ scaffold.Name = &TypesTest{}

// VersionSuiteTest scaffolds the version_suite_test.go file
type VersionSuiteTest struct {
	// Resource is the resource to scaffold the types_test.go file for
	*Resource
	// OutputPath is the output file to write
	OutputPath string
}

// Name implements scaffold.Name
func (VersionSuiteTest) Name() string {
	return "version-suite-test-go"
}

// Path implements scaffold.Path.  Defaults to pkg/apis/group/version/kind_types_test.go
func (t VersionSuiteTest) Path() string {
	dir := filepath.Join("pkg", "apis", t.Group, t.Version)
	if t.OutputPath != "" {
		dir = t.OutputPath
	}

	return filepath.Join(dir, fmt.Sprintf("%s_suite_test.go", t.Version))
}

// Execute writes the template file to wr.  b is the last value of the file.  temp is a template object.
func (t VersionSuiteTest) Execute(b []byte, temp *template.Template, wr func() io.WriteCloser) error {
	if len(b) > 0 {
		return nil
	}

	temp, err := temp.Parse(veersionSuiteTestTemplate)
	if err != nil {
		return err
	}

	fmt.Println(t.Path())

	w := wr()
	defer func() {
		if err := w.Close(); err != nil {
			log.Fatal(err)
		}
	}()
	return temp.Execute(w, t)
}

var veersionSuiteTestTemplate = `{{ .Boilerplate }}

package {{ .Version }}

import (
	"log"
	"os"
	"path/filepath"
	"testing"

	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

var cfg *rest.Config
var c client.Client

func TestMain(m *testing.M) {
	t := &envtest.Environment{
		CRDDirectoryPaths: []string{filepath.Join("..", "..", "..", "..", "config", "crds")},
	}

	err := AddToScheme(scheme.Scheme)
	if err != nil {
		log.Fatal(err)
	}

	if cfg, err = t.Start(); err != nil {
		log.Fatal(err)
	}

	if c, err = client.New(cfg, client.Options{Scheme: scheme.Scheme}); err != nil {
		log.Fatal(err)
	}

	code := m.Run()
	t.Stop()
	os.Exit(code)
}
`
