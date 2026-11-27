/*
Copyright 2026 The Kubernetes Authors.

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

package rsm

import (
	"testing"

	"sigs.k8s.io/controller-tools/pkg/metrics/internal/model"
)

func TestBuilder_AddResource(t *testing.T) {
	resource := model.Resource{
		GroupVersionKind: model.GroupVersionKind{Group: "example.io", Version: "v1", Kind: "Policy"},
		Metrics: []model.Generator{{
			Name: "foo_info", Help: "Foo.", Type: model.MetricTypeInfo,
			Path: model.Path{"spec"}, PathKind: model.PathKindObject,
		}},
	}

	// The metrics are validated for all targets by model.Resource.Validate before they get added.
	b := NewBuilder("", "", "")
	if err := b.AddResource(resource); err != nil {
		t.Fatal(err)
	}
	if len(b.stores) != 1 {
		t.Errorf("expected one store, got %d", len(b.stores))
	}
}
