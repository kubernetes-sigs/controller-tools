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
	"path"
	"path/filepath"
	"strings"

	"sigs.k8s.io/controller-tools/pkg/scaffold/input"
	"sigs.k8s.io/controller-tools/pkg/scaffold/resource"
)

// Controller scaffolds a Controller for a Resource
type Controller struct {
	input.Input

	// Resource is the Resource to make the Controller for
	Resource *resource.Resource

	// ResourcePackage is the package of the Resource
	ResourcePackage string

	// Is the Group + "." + Domain for the Resource
	GroupDomain string
}

// GetInput implements input.File
func (a *Controller) GetInput() (input.Input, error) {
	// Use the k8s.io/api package for core resources
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
	if domain, found := coreGroups[a.Resource.Group]; found {
		a.ResourcePackage = path.Join("k8s.io", "api")
		a.GroupDomain = a.Resource.Group
		if domain != "" {
			a.GroupDomain = a.Resource.Group + "." + domain
		}
	} else {
		a.ResourcePackage = path.Join(a.Repo, "pkg", "apis")
		a.GroupDomain = a.Resource.Group + "." + a.Domain
	}

	if a.Path == "" {
		a.Path = filepath.Join("pkg", "controller",
			strings.ToLower(a.Resource.Kind),
			strings.ToLower(a.Resource.Kind)+"_controller.go")
	}
	a.TemplateBody = controllerTemplate
	a.Input.IfExistsAction = input.Error
	return a.Input, nil
}

var controllerTemplate = `{{ .Boilerplate }}

package {{ lower .Resource.Kind }}

import (
{{ if .Resource.CreateExampleReconcileBody }}	"context"
	"log"
	"reflect"

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
	{{ .Resource.Group}}{{ .Resource.Version }} "{{ .ResourcePackage }}/{{ .Resource.Group}}/{{ .Resource.Version }}"
{{ else }}	"context"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	{{ .Resource.Group}}{{ .Resource.Version }} "{{ .ResourcePackage }}/{{ .Resource.Group}}/{{ .Resource.Version }}"
{{ end -}}
)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new {{ .Resource.Kind }} Controller and adds it to the Manager.  The Manager will set fields on the Controller
// and Start it when the Manager is Started.
// USER ACTION REQUIRED: update cmd/manager/main.go to call this {{ .Resource.Group}}.Add(mrg) to install this Controller
func Add(mrg manager.Manager) error {
	return add(mrg, newReconcile(mrg))
}

// newReconcile returns a new reconcile.Reconcile
func newReconcile(mrg manager.Manager) reconcile.Reconciler {
	return &Reconcile{{ .Resource.Kind }}{client: mrg.GetClient()}
}

// add adds a new Controller to mrg with r as the reconcile.Reconcile
func add(mrg manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("{{ lower .Resource.Kind }}-controller", mrg, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to {{ .Resource.Kind }}
	err = c.Watch(&source.Kind{Type: &{{ .Resource.Group}}{{ .Resource.Version }}.{{ .Resource.Kind }}{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create
	// Uncomment watch a Deployment created by {{ .Resource.Kind }} - change this for objects you create
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &{{ .Resource.Group}}{{ .Resource.Version }}.{{ .Resource.Kind }}{},
	})
	if err != nil {
		return err
	}

	return nil
}

// Reconcile{{ .Resource.Kind }} reconciles a {{ .Resource.Kind }} object
type Reconcile{{ .Resource.Kind }} struct {
	client client.Client
}

// Reconcile reads that state of the cluster for a {{ .Resource.Kind }} object and makes changes based on the state read
// and what is in the {{ .Resource.Kind }}.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  The scaffolding writes
// a Deployment as an example
func (r *Reconcile{{ .Resource.Kind }}) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the {{ .Resource.Kind }} instance
	instance := &{{ .Resource.Group}}{{ .Resource.Version }}.{{ .Resource.Kind }}{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	{{ if .Resource.CreateExampleReconcileBody -}}
	// TODO(user): Change this to be the object type created by your controller
	// Define the desired Deployment object
	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-deployment",
			Namespace: {{ if .Resource.Namespaced}}instance.Namespace{{ else }}"default"{{ end }},
			OwnerReferences: []metav1.OwnerReference{
				// TODO(user): Important to copy this line to any object you create so it can be tied back to
				// the {{ .Resource.Kind }} and cause a reconcile when it changes
				*metav1.NewControllerRef(instance, schema.GroupVersionKind{
					Group:   {{ .Resource.Group}}{{ .Resource.Version }}.SchemeGroupVersion.Group,
					Version: {{ .Resource.Group}}{{ .Resource.Version }}.SchemeGroupVersion.Version,
					Kind:    "{{ .Resource.Kind }}",
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
		log.Printf("Creating Deployment %s/%s\n", deploy.Namespace, deploy.Name)
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
		log.Printf("Updating Deployment %s/%s\n", deploy.Namespace, deploy.Name)
		err = r.client.Update(context.TODO(), found)
		if err != nil {
			return reconcile.Result{}, err
		}
	}
	{{ end -}}

	return reconcile.Result{}, nil
}
`
