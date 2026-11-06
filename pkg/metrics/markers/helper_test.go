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

package markers

import (
	"reflect"
	"testing"

	"sigs.k8s.io/controller-tools/pkg/metrics/internal/model"
)

func Test_newPath(t *testing.T) {
	tests := []struct {
		name     string
		basePath []string
		j        jsonPath
		want     model.Path
		wantErr  bool
	}{
		{name: "with basePath and jsonpath", basePath: []string{"foo"}, j: ".bar", want: model.Path{"foo", "bar"}},
		{name: "without jsonpath", basePath: []string{"foo"}, j: "", want: model.Path{"foo"}},
		{name: "without basePath", basePath: nil, j: ".foo.bar", want: model.Path{"foo", "bar"}},
		{name: "empty", basePath: []string{}, j: "", want: model.Path{}},
		{name: "dot", basePath: nil, j: ".", want: model.Path{""}},
		{name: "invalid character ,", basePath: nil, j: ".foo,.bar", wantErr: true},
		{name: "invalid closure", basePath: nil, j: "{.bar}", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newPath(tt.basePath, tt.j)
			if (err != nil) != tt.wantErr {
				t.Fatalf("newPath() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newPath() = %#v, want %#v", got, tt.want)
			}
		})
	}

	t.Run("does not modify the basePath", func(t *testing.T) {
		basePath := make([]string, 1, 10)
		basePath[0] = "foo"
		if _, err := newPath(basePath, ".bar"); err != nil {
			t.Fatal(err)
		}
		if got := basePath[:2][1]; got != "" {
			t.Errorf("newPath() modified the backing array of basePath: %q", got)
		}
	})
}

func Test_newLabels(t *testing.T) {
	got, err := newLabels(map[string]jsonPath{
		"b": ".label.from.path",
		"a": ".",
		"c": ".c",
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []model.Label{
		{Name: "a", Path: model.Path{}},
		{Name: "b", Path: model.Path{"label", "from", "path"}},
		{Name: "c", Path: model.Path{"c"}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("newLabels() = %v, want %v", got, want)
	}

	if _, err := newLabels(map[string]jsonPath{"a": "{.bar}"}); err == nil {
		t.Error("expected an error for an invalid jsonpath")
	}
}
