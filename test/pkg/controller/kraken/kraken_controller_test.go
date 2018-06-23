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
	"testing"
	"time"

	"github.com/onsi/gomega"
	"golang.org/x/net/context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	creaturesv2alpha1 "sigs.k8s.io/controller-tools/test/pkg/apis/creatures/v2alpha1"
)

var c client.Client

var expectedRequest = reconcile.Request{NamespacedName: types.NamespacedName{Name: "foo"}}

const timeout = time.Second * 5

func TestReconcile(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	instance := &creaturesv2alpha1.Kraken{ObjectMeta: metav1.ObjectMeta{Name: "foo"}}

	// Setup the Manager and Controller.  Wrap the Controller Reconcile function so it writes each request to a
	// channel when it is finished.
	mrg, err := manager.New(cfg, manager.Options{})
	g.Expect(err).NotTo(gomega.HaveOccurred())
	c = mrg.GetClient()

	recFn, requests := SetupTestReconcile(newReconcile(mrg))
	g.Expect(add(mrg, recFn)).NotTo(gomega.HaveOccurred())
	defer close(StartTestManager(mrg, g))

	// Create the Kraken object and expect the Reconcile
	g.Expect(c.Create(context.TODO(), instance)).To(gomega.Succeed())
	defer c.Delete(context.TODO(), instance)
	g.Eventually(requests, timeout).Should(gomega.Receive(gomega.Equal(expectedRequest)))

}
