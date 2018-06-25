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

package kraken

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	creaturesv2alpha1 "sigs.k8s.io/controller-tools/test/pkg/apis/creatures/v2alpha1"
)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Kraken Controller and adds it to the Manager.  The Manager will set fields on the Controller
// and Start it when the Manager is Started.
// USER ACTION REQUIRED: update cmd/manager/main.go to call this creatures.Add(mrg) to install this Controller
func Add(mrg manager.Manager) error {
	return add(mrg, newReconcile(mrg))
}

// newReconcile returns a new reconcile.Reconcile
func newReconcile(mrg manager.Manager) reconcile.Reconcile {
	return &ReconcileKraken{client: mrg.GetClient()}
}

// add adds a new Controller to mrg with r as the reconcile.Reconcile
func add(mrg manager.Manager, r reconcile.Reconcile) error {
	// Create a new controller
	c, err := controller.New("kraken-controller", mrg, controller.Options{Reconcile: r})
	if err != nil {
		return err
	}

	// Watch for changes to Kraken
	err = c.Watch(&source.Kind{Type: &creaturesv2alpha1.Kraken{}}, &handler.Enqueue{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create
	// Uncomment watch a Deployment created by Kraken - change this for objects you create
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueOwner{
		IsController: true,
		OwnerType:    &creaturesv2alpha1.Kraken{},
	})
	if err != nil {
		return err
	}

	return nil
}

// ReconcileKraken reconciles a Kraken object
type ReconcileKraken struct {
	client client.Client
}

// Reconcile reads that state of the cluster for a Kraken object and makes changes based on the state read
// and what is in the Kraken.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  The scaffolding writes
// a Deployment as an example
func (r *ReconcileKraken) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the Kraken instance
	instance := &creaturesv2alpha1.Kraken{}
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

	return reconcile.Result{}, nil
}
