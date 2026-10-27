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

package metrics

import (
	"os"
	"path"
	"strings"
	"testing"

	"sigs.k8s.io/controller-tools/pkg/loader"
	"sigs.k8s.io/controller-tools/pkg/metrics/internal/model"
)

func Test_annotatePathAndValueKinds(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	roots, err := loader.LoadRoots(path.Join(cwd, "testdata", "v1"))
	if err != nil || len(roots) != 1 {
		t.Fatalf("loading the package: %v (%d packages)", err, len(roots))
	}
	// The type checker makes sure the imported packages are checked first. Without it
	// the types of the imported packages are invalid, like it is in the generator.
	(&loader.TypeChecker{}).Check(roots[0])
	roots[0].NeedTypesInfo()

	gauge := func(path ...string) model.Generator {
		return model.Generator{Type: model.MetricTypeGauge, Path: path}
	}

	tests := []struct {
		name string
		// root is the name of the type in testdata/v1 the path starts at. Defaults to Foo.
		root          string
		metric        model.Generator
		wantPathKind  model.PathKind
		wantValueKind model.ValueKind
		wantErrPart   string
	}{
		// Path kinds.
		{name: "scalar", metric: gauge("spec", "someString"), wantPathKind: model.PathKindScalar},
		{name: "struct", metric: model.Generator{Type: model.MetricTypeInfo, Path: model.Path{"spec"}}, wantPathKind: model.PathKindObject},
		{name: "map of an inlined field", metric: model.Generator{Type: model.MetricTypeInfo, Path: model.Path{"metadata", "labels"}}, wantPathKind: model.PathKindMap},
		{name: "list", metric: model.Generator{Type: model.MetricTypeStateSet, Path: model.Path{"status", "conditions"}}, wantPathKind: model.PathKindArray},
		{name: "list of an inlined field", metric: model.Generator{Type: model.MetricTypeInfo, Path: model.Path{"metadata", "ownerReferences"}}, wantPathKind: model.PathKindArray},
		{name: "map", root: "PathKinds", metric: gauge("map"), wantPathKind: model.PathKindMap},
		{name: "entry of a map of structs", root: "PathKinds", metric: gauge("map", "someKey"), wantPathKind: model.PathKindObject},
		{name: "field of an entry of a map, the key is consumed", root: "PathKinds", metric: gauge("map", "someKey", "a"), wantPathKind: model.PathKindScalar},

		// Inlined fields.
		{name: "inlined embedded field", root: "PathKinds", metric: gauge("a"), wantPathKind: model.PathKindScalar},
		{name: "inlined embedded field with a name and the inline option", root: "PathKinds", metric: gauge("b"), wantPathKind: model.PathKindScalar},
		{name: "embedded field without a json name", root: "PathKinds", metric: gauge("c"), wantPathKind: model.PathKindScalar},
		{name: "field with a name containing inline is accessible by its name", root: "PathKinds", metric: gauge("inlineFoo", "a"), wantPathKind: model.PathKindScalar},

		// Value kinds.
		{name: "timestamp type", metric: gauge("metadata", "creationTimestamp"), wantPathKind: model.PathKindObject, wantValueKind: model.ValueKindTimestamp},
		{
			name:         "timestamp type in the elements of a list",
			metric:       model.Generator{Type: model.MetricTypeGauge, Path: model.Path{"status", "conditions"}, Value: model.Path{"lastTransitionTime"}},
			wantPathKind: model.PathKindArray, wantValueKind: model.ValueKindTimestamp,
		},
		{
			name:         "no timestamp type",
			metric:       model.Generator{Type: model.MetricTypeGauge, Path: model.Path{"status", "conditions"}, Value: model.Path{"status"}},
			wantPathKind: model.PathKindArray,
		},
		{
			name:         "only gauges have a value kind",
			metric:       model.Generator{Type: model.MetricTypeStateSet, Path: model.Path{"status", "conditions"}, Value: model.Path{"lastTransitionTime"}},
			wantPathKind: model.PathKindArray,
		},
		{name: "micro time", root: "PathKinds", metric: gauge("microTime"), wantPathKind: model.PathKindObject, wantValueKind: model.ValueKindTimestamp},
		{name: "pointer to time", root: "PathKinds", metric: gauge("timePointer"), wantPathKind: model.PathKindObject, wantValueKind: model.ValueKindTimestamp},
		{name: "alias of time", root: "PathKinds", metric: gauge("timeAlias"), wantPathKind: model.PathKindObject, wantValueKind: model.ValueKindTimestamp},
		{name: "alias of a pointer to time", root: "PathKinds", metric: gauge("timeAliasPointer"), wantPathKind: model.PathKindObject, wantValueKind: model.ValueKindTimestamp},
		{name: "pointer to an alias of time", root: "PathKinds", metric: gauge("pointerToTimeAlias"), wantPathKind: model.PathKindObject, wantValueKind: model.ValueKindTimestamp},
		{name: "other type named Time", root: "PathKinds", metric: gauge("otherTime"), wantPathKind: model.PathKindObject},
		{name: "other type of metav1", root: "PathKinds", metric: gauge("duration"), wantPathKind: model.PathKindObject},

		// Paths which don't exist in the Go types are configuration errors.
		{name: "unknown path", metric: gauge("spec", "doesNotExist"), wantErrPart: `the path "spec.doesNotExist" does not exist`},
		{
			name:        "unknown value path",
			metric:      model.Generator{Type: model.MetricTypeGauge, Path: model.Path{"status", "conditions"}, Value: model.Path{"doesNotExist"}},
			wantErrPart: `the value path "status.conditions.doesNotExist" does not exist`,
		},
		{
			name:        "unknown value path of a stateset",
			metric:      model.Generator{Type: model.MetricTypeStateSet, Path: model.Path{"status", "conditions"}, Value: model.Path{"doesNotExist"}},
			wantErrPart: `the value path "status.conditions.doesNotExist" does not exist`,
		},
		{name: "field with a name containing inline is not inlined", root: "PathKinds", metric: gauge("inlineFoo", "x"), wantErrPart: "does not exist"},
		{name: "path below any", root: "PathKinds", metric: gauge("any", "foo"), wantErrPart: `"any.foo" does not exist`},
		{name: "path below runtime.RawExtension", root: "PathKinds", metric: gauge("raw", "foo"), wantErrPart: `"raw.foo" does not exist`},
		{name: "path below apiextensionsv1.JSON", root: "PathKinds", metric: gauge("json", "foo"), wantErrPart: `"json.foo" does not exist`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rootName := tt.root
			if rootName == "" {
				rootName = "Foo"
			}
			object := roots[0].Types.Scope().Lookup(rootName)
			if object == nil {
				t.Fatalf("type %s not found", rootName)
			}

			metrics := []model.Generator{tt.metric}
			err := annotatePathAndValueKinds(metrics, object.Type())
			if tt.wantErrPart != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErrPart) {
					t.Fatalf("annotatePathAndValueKinds() error = %v, want error containing %q", err, tt.wantErrPart)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if metrics[0].PathKind != tt.wantPathKind {
				t.Errorf("PathKind = %q, want %q", metrics[0].PathKind, tt.wantPathKind)
			}
			if metrics[0].ValueKind != tt.wantValueKind {
				t.Errorf("ValueKind = %q, want %q", metrics[0].ValueKind, tt.wantValueKind)
			}
		})
	}
}
