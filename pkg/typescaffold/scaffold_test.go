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

package typescaffold_test

import (
	"bytes"
	"testing"

	"sigs.k8s.io/controller-tools/pkg/typescaffold"
)

func TestScaffold(t *testing.T) {
	tests := []struct {
		name string
		opts typescaffold.ScaffoldOptions
	}{
		{
			name: "kind only",
			opts: typescaffold.ScaffoldOptions{
				Resource: typescaffold.Resource{
					Kind: "Foo",
				},
			},
		},
		{
			name: "kind and resource",
			opts: typescaffold.ScaffoldOptions{
				Resource: typescaffold.Resource{
					Kind:     "Foo",
					Resource: "foos",
				},
			},
		},
		{
			name: "namespaced",
			opts: typescaffold.ScaffoldOptions{
				Resource: typescaffold.Resource{
					Kind:       "Foo",
					Resource:   "foos",
					Namespaced: true,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.opts.Validate()
			if err != nil {
				t.Fatalf("unable to validate scaffold opts: %v", err)
			}
			var out bytes.Buffer
			if err := test.opts.Scaffold(&out); err != nil {
				t.Fatalf("unable to scaffold types: %v", err)
			}

			// TODO(directxman12): testing the direct output seems fragile
			// there must be a better way.
		})
	}
}

func TestInvalidScaffoldOpts(t *testing.T) {
	tests := []struct {
		name string
		opts typescaffold.ScaffoldOptions
	}{
		{
			name: "bad kind",
			opts: typescaffold.ScaffoldOptions{
				Resource: typescaffold.Resource{
					Kind: "Foo_bats",
				},
			},
		},
		{
			name: "no kind",
			opts: typescaffold.ScaffoldOptions{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.opts.Validate()
			if err == nil {
				t.Fatalf("expected error -- those options were invalid")
			}
		})
	}
}
