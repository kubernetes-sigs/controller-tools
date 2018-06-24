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

package controller

import (
	"path/filepath"
	"strings"

	"sigs.k8s.io/controller-tools/pkg/scaffold/input"
	"sigs.k8s.io/controller-tools/pkg/scaffold/resource"
)

// Test scaffolds a Controller Test
type Test struct {
	input.Input

	// Resource is the Resource to make the Controller for
	Resource *resource.Resource
}

// GetInput implements input.File
func (a *Test) GetInput() (input.Input, error) {
	if a.Path == "" {
		a.Path = filepath.Join("pkg", "controller",
			strings.ToLower(a.Resource.Kind), strings.ToLower(a.Resource.Kind)+"_controller_test.go")
	}
	a.TemplateBody = controllerTestTemplate
	return a.Input, nil
}

var controllerTestTemplate = `{{ .Boilerplate }}

package {{ lower .Resource.Kind }}

import (
	"testing"
	"time"

	{{ if .Resource.CreateExampleReconcileBody }}
	"github.com/onsi/gomega"
	"golang.org/x/net/context"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	{{ .Resource.Group }}{{ .Resource.Version }} "{{ .Repo }}/pkg/apis/{{ .Resource.Group }}/{{ .Resource.Version }}"
	{{ else }}
	"github.com/onsi/gomega"
	"golang.org/x/net/context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	{{ .Resource.Group }}{{ .Resource.Version }} "{{ .Repo }}/pkg/apis/{{ .Resource.Group }}/{{ .Resource.Version }}"
	{{ end }}
)

var c client.Client

var expectedRequest = reconcile.Request{NamespacedName: types.NamespacedName{Name: "foo"{{ if .Resource.Namespaced }}, Namespace: "default"{{end}}}}
{{ if .Resource.CreateExampleReconcileBody }}var depKey = types.NamespacedName{Name: "foo-deployment"{{ if .Resource.Namespaced }}, Namespace: "default"{{end}}}{{ end }}

const timeout = time.Second * 5

func TestReconcile(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	instance := &{{ .Resource.Group }}{{ .Resource.Version }}.{{ .Resource.Kind }}{ObjectMeta: metav1.ObjectMeta{Name: "foo"{{ if .Resource.Namespaced }}, Namespace: "default"{{end}}}}

	// Setup the Manager and Controller.  Wrap the Controller Reconcile function so it writes each request to a
	// channel when it is finished.
	mrg, err := manager.New(cfg, manager.Options{})
	g.Expect(err).NotTo(gomega.HaveOccurred())
	c = mrg.GetClient()

	recFn, requests := SetupTestReconcile(newReconcile(mrg))
	g.Expect(add(mrg, recFn)).NotTo(gomega.HaveOccurred())
	defer close(StartTestManager(mrg, g))

	// Create the {{ .Resource.Kind }} object and expect the Reconcile {{ if .Resource.CreateExampleReconcileBody }}and Deployment to be created{{ end }}
	g.Expect(c.Create(context.TODO(), instance)).To(gomega.Succeed())
	defer c.Delete(context.TODO(), instance)
	g.Eventually(requests, timeout).Should(gomega.Receive(gomega.Equal(expectedRequest)))
	{{ if .Resource.CreateExampleReconcileBody }}
	deploy := &appsv1.Deployment{}
	g.Eventually(func() error { return c.Get(context.TODO(), depKey, deploy) }, timeout).
		Should(gomega.Succeed())

	// Delete the Deployment and expect Reconcile to be called for Deployment deletion
	g.Expect(c.Delete(context.TODO(), deploy)).NotTo(gomega.HaveOccurred())
	g.Eventually(requests, timeout).Should(gomega.Receive(gomega.Equal(expectedRequest)))
	g.Eventually(func() error { return c.Get(context.TODO(), depKey, deploy) }, timeout).
		Should(gomega.Succeed())

	// Manually delete Deployment since GC isn't enabled in the test control plane
	g.Expect(c.Delete(context.TODO(), deploy)).To(gomega.Succeed())
	{{ end }}
}
`
