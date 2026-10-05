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
	"go/token"
	"go/types"
	"os"
	"path"
	"testing"

	"sigs.k8s.io/controller-tools/pkg/loader"
	"sigs.k8s.io/controller-tools/pkg/metrics/internal/model"
)

func Test_annotatePathKinds(t *testing.T) {
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
	foo := roots[0].Types.Scope().Lookup("Foo")
	if foo == nil {
		t.Fatal("type Foo not found")
	}

	tests := []struct {
		name          string
		metric        model.Generator
		wantPathKind  model.PathKind
		wantValueKind model.ValueKind
	}{
		{name: "scalar", metric: model.Generator{Type: model.MetricTypeGauge, Path: model.Path{"spec", "someString"}}, wantPathKind: model.PathKindScalar},
		{name: "struct", metric: model.Generator{Type: model.MetricTypeInfo, Path: model.Path{"spec"}}, wantPathKind: model.PathKindObject},
		{name: "map of an inlined field", metric: model.Generator{Type: model.MetricTypeInfo, Path: model.Path{"metadata", "labels"}}, wantPathKind: model.PathKindMap},
		{name: "list", metric: model.Generator{Type: model.MetricTypeStateSet, Path: model.Path{"status", "conditions"}}, wantPathKind: model.PathKindArray},
		{name: "list of an inlined field", metric: model.Generator{Type: model.MetricTypeInfo, Path: model.Path{"metadata", "ownerReferences"}}, wantPathKind: model.PathKindArray},
		{
			name:         "timestamp type",
			metric:       model.Generator{Type: model.MetricTypeGauge, Path: model.Path{"metadata", "creationTimestamp"}},
			wantPathKind: model.PathKindObject, wantValueKind: model.ValueKindTimestamp,
		},
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
		{name: "unknown path is scalar", metric: model.Generator{Type: model.MetricTypeGauge, Path: model.Path{"spec", "doesNotExist"}}, wantPathKind: model.PathKindScalar},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := []model.Generator{tt.metric}
			annotatePathKinds(metrics, foo.Type())
			if metrics[0].PathKind != tt.wantPathKind {
				t.Errorf("PathKind = %q, want %q", metrics[0].PathKind, tt.wantPathKind)
			}
			if metrics[0].ValueKind != tt.wantValueKind {
				t.Errorf("ValueKind = %q, want %q", metrics[0].ValueKind, tt.wantValueKind)
			}
		})
	}
}

func Test_goTypeAtJSONPath_inlineAndMaps(t *testing.T) {
	pkg := types.NewPackage("example.io/api", "api")
	str := types.Typ[types.String]
	field := func(name string, typ types.Type, embedded bool) *types.Var {
		return types.NewField(token.NoPos, pkg, name, typ, embedded)
	}

	inner := types.NewNamed(types.NewTypeName(token.NoPos, pkg, "Inner", nil), types.NewStruct([]*types.Var{field("A", str, false)}, []string{`json:"a"`}), nil)

	tests := []struct {
		name     string
		root     types.Type
		path     []string
		wantType types.Type
	}{
		{
			name:     "inlined embedded field",
			root:     types.NewStruct([]*types.Var{field("Inner", inner, true)}, []string{`json:",inline"`}),
			path:     []string{"a"},
			wantType: str,
		},
		{
			name:     "inlined embedded field with a name and the inline option",
			root:     types.NewStruct([]*types.Var{field("Inner", inner, true)}, []string{`json:"x,inline"`}),
			path:     []string{"a"},
			wantType: str,
		},
		{
			name:     "embedded field without a json name",
			root:     types.NewStruct([]*types.Var{field("Inner", inner, true)}, []string{""}),
			path:     []string{"a"},
			wantType: str,
		},
		{
			// The name only contains the word inline, which must not inline the field.
			name: "embedded field with a name containing inline is not inlined",
			root: types.NewStruct([]*types.Var{field("Inner", inner, true)}, []string{`json:"inlineFoo"`}),
			path: []string{"a"},
		},
		{
			name:     "embedded field with a name containing inline is accessible by its name",
			root:     types.NewStruct([]*types.Var{field("Inner", inner, true)}, []string{`json:"inlineFoo"`}),
			path:     []string{"inlineFoo", "a"},
			wantType: str,
		},
		{
			name:     "key of a map is consumed",
			root:     types.NewStruct([]*types.Var{field("M", types.NewMap(str, inner), false)}, []string{`json:"m"`}),
			path:     []string{"m", "someKey", "a"},
			wantType: str,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := goTypeAtJSONPath(tt.root, tt.path)
			if got != tt.wantType {
				t.Errorf("goTypeAtJSONPath() = %v, want %v", got, tt.wantType)
			}
		})
	}

	t.Run("kind of a map", func(t *testing.T) {
		root := types.NewStruct([]*types.Var{field("M", types.NewMap(str, inner), false)}, []string{`json:"m"`})
		if got := goPathKind(root, []string{"m"}); got != model.PathKindMap {
			t.Errorf("goPathKind() = %q, want %q", got, model.PathKindMap)
		}
		if got := goPathKind(root, []string{"m", "key"}); got != model.PathKindObject {
			t.Errorf("goPathKind() of an entry of a map of structs = %q, want %q", got, model.PathKindObject)
		}
	})
}

func Test_isTimestampType_aliases(t *testing.T) {
	metav1 := types.NewPackage("k8s.io/apimachinery/pkg/apis/meta/v1", "v1")
	other := types.NewPackage("example.io/api", "api")
	newNamed := func(pkg *types.Package, name string) *types.Named {
		return types.NewNamed(types.NewTypeName(token.NoPos, pkg, name, nil), types.NewStruct(nil, nil), nil)
	}
	newAlias := func(name string, rhs types.Type) types.Type {
		return types.NewAlias(types.NewTypeName(token.NoPos, other, name, nil), rhs)
	}

	time := newNamed(metav1, "Time")
	tests := []struct {
		name string
		typ  types.Type
		want bool
	}{
		{name: "time", typ: time, want: true},
		{name: "micro time", typ: newNamed(metav1, "MicroTime"), want: true},
		{name: "pointer to time", typ: types.NewPointer(time), want: true},
		{name: "alias of time", typ: newAlias("MyTime", time), want: true},
		{name: "alias of a pointer to time", typ: newAlias("MyTimePtr", types.NewPointer(time)), want: true},
		{name: "pointer to an alias of time", typ: types.NewPointer(newAlias("MyTime", time)), want: true},
		{name: "other type named Time", typ: newNamed(other, "Time"), want: false},
		{name: "other type of metav1", typ: newNamed(metav1, "Duration"), want: false},
		{name: "nil", typ: nil, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isTimestampType(tt.typ); got != tt.want {
				t.Errorf("isTimestampType() = %v, want %v", got, tt.want)
			}
		})
	}
}
