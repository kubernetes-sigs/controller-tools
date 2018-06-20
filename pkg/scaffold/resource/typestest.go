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

var _ scaffold.Template = &TypesTest{}
var _ scaffold.Name = &TypesTest{}

// TypesTest scaffolds the types_test.go file for testing APIsGo
type TypesTest struct {
	// Resource is the resource to scaffold the types_test.go file for
	*Resource
	// OutputPath is the output file to write
	OutputPath string
}

// Name implements scaffold.Name
func (TypesTest) Name() string {
	return "types-test-go"
}

// Path implements scaffold.Path.  Defaults to pkg/apis/group/version/kind_types_test.go
func (t TypesTest) Path() string {
	dir := filepath.Join("pkg", "apis", t.Group, t.Version)
	if t.OutputPath != "" {
		dir = t.OutputPath
	}

	return filepath.Join(dir, fmt.Sprintf("%s_types_test.go", strings.ToLower(t.Kind)))
}

// Execute writes the template file to wr.  b is the last value of the file.  temp is a template object.
func (t TypesTest) Execute(b []byte, temp *template.Template, wr func() io.WriteCloser) error {
	if len(b) > 0 {
		return fmt.Errorf("%s already exists", t.Path())
	}

	temp, err := temp.Parse(typesTestTemplate)
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

var typesTestTemplate = `{{ .Boilerplate }}

package {{ .Version }}

import (
	"reflect"
	"testing"

	"golang.org/x/net/context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestStorage(t *testing.T) {
	instance := &{{ .Kind }}{
		ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "default"},
	}
	if err := c.Create(context.TODO(), instance); err != nil {
		t.Logf("Could not create {{ .Kind }} %v", err)
		t.FailNow()
	}

	read := &{{ .Kind }}{}
	if err := c.Get(context.TODO(), types.NamespacedName{Name: "foo", Namespace: "default"}, read); err != nil {
		t.Logf("Could not get {{ .Kind }} %v", err)
		t.FailNow()
	}
	if !reflect.DeepEqual(read, instance) {
		t.Logf("Created and Read do not match")
		t.FailNow()
	}

	new := read.DeepCopy()
	new.Labels = map[string]string{"hello": "world"}

	if err := c.Update(context.TODO(), new); err != nil {
		t.Logf("Could not create {{ .Kind }} %v", err)
		t.FailNow()
	}

	if err := c.Get(context.TODO(), types.NamespacedName{Name: "foo", Namespace: "default"}, read); err != nil {
		t.Logf("Could not get {{ .Kind }} %v", err)
		t.FailNow()
	}
	if !reflect.DeepEqual(read, new) {
		t.Logf("Updated and Read do not match")
		t.FailNow()
	}

	if err := c.Delete(context.TODO(), instance); err != nil {
		t.Logf("Could not get {{ .Kind }} %v", err)
		t.FailNow()
	}

	if err := c.Get(context.TODO(), types.NamespacedName{Name: "foo", Namespace: "default"}, instance); err == nil {
		t.Logf("Found deleted {{ .Kind }}")
		t.FailNow()
	}
}
`
