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
	"fmt"
	"io"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"log"

	flag "github.com/spf13/pflag"
	"sigs.k8s.io/controller-tools/pkg/scaffold/resource"
)

// Controller scaffolds a Controller for a Resource
type Controller struct {
	// Resource is the Resource to make the Controller for
	*resource.Resource

	// OutputPath is the output location of the scaffold file
	OutputPath string

	// Boilerplate is the boilerplate header to apply
	Boilerplate string

	// ResourcePackage is the package of the Resource
	ResourcePackage string

	// Is the Group + "." + Domain for the Resource
	GroupDomain string
}

// Name is the name of the template
func (Controller) Name() string {
	return "controller-go"
}

// Path implements scaffold.Path.  Defaults to pkg/controller/kind/kind_controller
func (m *Controller) Path() string {
	dir := filepath.Join("pkg", "controller", strings.ToLower(m.Kind), strings.ToLower(m.Kind)+"_controller.go")
	if m.OutputPath != "" {
		dir = m.OutputPath
	}
	return dir
}

// SetBoilerplate implements scaffold.Boilerplate.
func (m *Controller) SetBoilerplate(b string) {
	m.Boilerplate = b
}

// Execute writes the template file to wr.  b is the last value of the file.  temp is a template object.
func (m *Controller) Execute(b []byte, t *template.Template, wr func() io.WriteCloser) error {
	if len(b) > 0 {
		return fmt.Errorf("pkg/controller/%s/controller.go already exists", strings.ToLower(m.Kind))
	}
	temp, err := t.Parse(managerTemplate)
	if err != nil {
		return err
	}

	// Use the k8s.io/api package for core AddResource
	coreGroups := map[string]string{
		"apps":                  "",
		"admissionregistration": "k8s.io",
		"apiextensions":         "k8s.io",
		"authentication":        "k8s.io",
		"autoscaling":           "",
		"batch":                 "",
		"certificates":          "k8s.io",
		"core":                  "",
		"extensions":            "",
		"metrics":               "k8s.io",
		"policy":                "",
		"rbac.authorization":    "k8s.io",
		"storage":               "k8s.io",
	}
	if domain, found := coreGroups[m.Group]; found {
		m.ResourcePackage = path.Join("k8s.io", "api")
		m.GroupDomain = m.Group
		if domain != "" {
			m.GroupDomain = m.GroupDomain + "." + domain
		}
	} else {
		m.ResourcePackage = path.Join(m.Project.Repo, "pkg", "apis")
		m.GroupDomain = m.Group + "." + m.Project.Domain
	}

	w := wr()
	defer func() {
		if err := w.Close(); err != nil {
			log.Fatal(err)
		}
	}()
	return temp.Execute(w, m)
}

// ForFlags registers flags for Controller fields and returns the Controller
func ForFlags(f *flag.FlagSet) *Controller {
	d := &Controller{
		Resource: resource.ForFlags(f),
	}

	return d
}

var managerTemplate = `{{ .Boilerplate }}

package {{ lower .Kind }}

import (
	"context"
	"log"
	"reflect"

	{{ .Group }}{{ .Version }} "{{ .ResourcePackage }}/{{ .Group }}/{{ .Version }}"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new {{ .Kind }} Controller and adds it to the Manager.  The Manager will set fields on the Controller
// and Start it when the Manager is Started.
// USER ACTION REQUIRED: update cmd/manager/main.go to call this {{ .Group }}.Add(mrg) to install this Controller
func Add(mrg manager.Manager) error {
	// Create a new controller
	c, err := controller.New("{{ lower .Kind }}-controller", mrg, controller.Options{
		Reconcile: &Reconcile{{ .Kind }}{
			client: mrg.GetClient(),
		},
	})
	if err != nil {
		return err
	}

	// Watch for changes to {{ .Kind }}
	err = c.Watch(&source.Kind{Type: &{{ .Group }}{{ .Version }}.{{ .Kind }}{}}, &handler.Enqueue{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create
	// Uncomment watch a Deployment created by {{ .Kind }} - change this for objects you create
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueOwner{
		IsController: true,
		OwnerType:    &{{ .Group }}{{ .Version }}.{{ .Kind }}{},
	})
	if err != nil {
		return err
	}

	return nil
}

// Reconcile{{ .Kind }} reconciles a {{ .Kind }} object
type Reconcile{{ .Kind }} struct {
	client client.Client
}

// Reconcile reads that state of the cluster for a {{ .Kind }} object and makes changes based on the state read
// and what is in the {{ .Kind }}.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  The scaffolding writes
// a Deployment as an example
func (r *Reconcile{{ .Kind }}) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the {{ .Kind }} instance
	instance := &{{ .Group }}{{ .Version }}.{{ .Kind }}{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		} else {
			// Error reading the object - requeue the request.
			return reconcile.Result{}, err
		}
	}

	// TODO(user): Change this to be the object type created by your controller
	// Define the desired Deployment object
	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-deployment",
			Namespace: instance.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				// TODO(user): Important to copy this line to any object you create so it can be tied back to
				// the {{ .Kind }} and cause a reconcile when it changes
				*metav1.NewControllerRef(instance, schema.GroupVersionKind{
					Group:   "{{ .GroupDomain }}",
					Version: "{{ .Version }}",
					Kind:    "{{ .Kind }}",
				}),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"deployment": instance.Name + "-deployment"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"deployment": instance.Name + "-deployment"}},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "nginx",
							Image: "nginx",
						},
					},
				},
			},
		},
	}

	// TODO(user): Change this for the object type created by your controller
	// Check if the Deployment already exists
	found := &appsv1.Deployment{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: deploy.Name, Namespace: deploy.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		log.Printf("Creating deployment %+v\n", deploy)
		err = r.client.Create(context.TODO(), deploy)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// TODO(user): Change this for the object type created by your controller
	// Update the found object and write the result back if there are any changes
	if !reflect.DeepEqual(deploy.Spec, found.Spec) {
		found.Spec = deploy.Spec
		log.Printf("Updating deployment %+v\n", found)
		err = r.client.Update(context.TODO(), found)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}
`
