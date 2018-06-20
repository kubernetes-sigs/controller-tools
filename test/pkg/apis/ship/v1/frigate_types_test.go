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

package v1

import (
	"reflect"
	"testing"

	"golang.org/x/net/context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestStorage(t *testing.T) {
	instance := &Frigate{
		ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "default"},
	}
	if err := c.Create(context.TODO(), instance); err != nil {
		t.Logf("Could not create Frigate %v", err)
		t.FailNow()
	}

	read := &Frigate{}
	if err := c.Get(context.TODO(), types.NamespacedName{Name: "foo", Namespace: "default"}, read); err != nil {
		t.Logf("Could not get Frigate %v", err)
		t.FailNow()
	}
	if !reflect.DeepEqual(read, instance) {
		t.Logf("Created and Read do not match")
		t.FailNow()
	}

	new := read.DeepCopy()
	new.Labels = map[string]string{"hello": "world"}

	if err := c.Update(context.TODO(), new); err != nil {
		t.Logf("Could not create Frigate %v", err)
		t.FailNow()
	}

	if err := c.Get(context.TODO(), types.NamespacedName{Name: "foo", Namespace: "default"}, read); err != nil {
		t.Logf("Could not get Frigate %v", err)
		t.FailNow()
	}
	if !reflect.DeepEqual(read, new) {
		t.Logf("Updated and Read do not match")
		t.FailNow()
	}

	if err := c.Delete(context.TODO(), instance); err != nil {
		t.Logf("Could not get Frigate %v", err)
		t.FailNow()
	}

	if err := c.Get(context.TODO(), types.NamespacedName{Name: "foo", Namespace: "default"}, instance); err == nil {
		t.Logf("Found deleted Frigate")
		t.FailNow()
	}
}
