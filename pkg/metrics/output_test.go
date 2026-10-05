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
	"strings"
	"testing"

	"sigs.k8s.io/controller-tools/pkg/metrics/internal/ksm"
	"sigs.k8s.io/controller-tools/pkg/metrics/internal/output"
	"sigs.k8s.io/controller-tools/pkg/metrics/internal/rsm"
)

func Test_newOutputBuilder(t *testing.T) {
	tests := []struct {
		name        string
		generator   Generator
		wantRSM     bool
		wantKSM     bool
		wantErrPart string
	}{
		{name: "default is resource-state-metrics", generator: Generator{}, wantRSM: true},
		{name: "default with name and namespace", generator: Generator{Name: "foo", Namespace: "bar"}, wantRSM: true},
		{name: "resource-state-metrics", generator: Generator{Target: targetResourceStateMetrics}, wantRSM: true},
		{name: "resource-state-metrics with name and namespace", generator: Generator{Target: targetResourceStateMetrics, Name: "foo", Namespace: "bar"}, wantRSM: true},
		{name: "invalid name", generator: Generator{Name: "Foo_Bar"}, wantErrPart: `invalid name "Foo_Bar"`},
		{name: "invalid namespace", generator: Generator{Namespace: "foo.bar"}, wantErrPart: `invalid namespace "foo.bar"`},
		{name: "kube-state-metrics", generator: Generator{Target: targetKubeStateMetrics}, wantKSM: true},
		{name: "name is not supported for kube-state-metrics", generator: Generator{Target: targetKubeStateMetrics, Name: "foo"}, wantErrPart: "only supported for the resource-state-metrics target"},
		{name: "namespace is not supported for kube-state-metrics", generator: Generator{Target: targetKubeStateMetrics, Namespace: "foo"}, wantErrPart: "only supported for the resource-state-metrics target"},
		{name: "unsupported target", generator: Generator{Target: "foo"}, wantErrPart: `unsupported metrics target "foo", supported targets are kube-state-metrics, resource-state-metrics`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.generator.newOutputBuilder()
			if tt.wantErrPart != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErrPart) {
					t.Fatalf("newOutputBuilder() error = %v, want error containing %q", err, tt.wantErrPart)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			assertBuilderType(t, got, tt.wantRSM, tt.wantKSM)
		})
	}
}

func assertBuilderType(t *testing.T, got output.Builder, wantRSM, wantKSM bool) {
	t.Helper()
	_, isRSM := got.(*rsm.Builder)
	_, isKSM := got.(*ksm.Builder)
	if isRSM != wantRSM || isKSM != wantKSM {
		t.Errorf("newOutputBuilder() = %T, want rsm=%v ksm=%v", got, wantRSM, wantKSM)
	}
}
