/*
Copyright 2018 The Kubernetes authors.

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

package firstmate

import (
	"context"
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
	crewv1 "sigs.k8s.io/controller-tools/test/pkg/apis/crew/v1"
)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new FirstMate Controller and adds it to the Manager.  The Manager will set fields on the Controller
// and Start it when the Manager is Started.
// USER ACTION REQUIRED: update cmd/manager/main.go to call this crew.Add(mrg) to install this Controller
func Add(mrg manager.Manager) error {
	return add(mrg, newReconcile(mrg))
}

// newReconcile returns a new reconcile.Reconcile
func newReconcile(mrg manager.Manager) reconcile.Reconcile {
	return &ReconcileFirstMate{client: mrg.GetClient()}
}

// add adds a new Controller to mrg with r as the reconcile.Reconcile
func add(mrg manager.Manager, r reconcile.Reconcile) error {
	// Create a new controller
	c, err := controller.New("firstmate-controller", mrg, controller.Options{Reconcile: r})
	if err != nil {
		return err
	}

	// Watch for changes to FirstMate
	err = c.Watch(&source.Kind{Type: &crewv1.FirstMate{}}, &handler.Enqueue{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create
	// Uncomment watch a Deployment created by FirstMate - change this for objects you create
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueOwner{
		IsController: true,
		OwnerType:    &crewv1.FirstMate{},
	})
	if err != nil {
		return err
	}

	return nil
}

// ReconcileFirstMate reconciles a FirstMate object
type ReconcileFirstMate struct {
	client client.Client
}

// Reconcile reads that state of the cluster for a FirstMate object and makes changes based on the state read
// and what is in the FirstMate.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  The scaffolding writes
// a Deployment as an example
func (r *ReconcileFirstMate) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the FirstMate instance
	instance := &crewv1.FirstMate{}
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

	// TODO(user): Change this to be the object type created by your controller
	// Define the desired Deployment object
	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-deployment",
			Namespace: instance.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				// TODO(user): Important to copy this line to any object you create so it can be tied back to
				// the FirstMate and cause a reconcile when it changes
				*metav1.NewControllerRef(instance, schema.GroupVersionKind{
					Group:   crewv1.SchemeGroupVersion.Group,
					Version: crewv1.SchemeGroupVersion.Version,
					Kind:    "FirstMate",
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

	return reconcile.Result{}, nil
}
