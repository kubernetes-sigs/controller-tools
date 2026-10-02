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
	"bytes"
	"io"
	"maps"
	"os"
	"path"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
	"sigs.k8s.io/controller-tools/pkg/genall"
	"sigs.k8s.io/controller-tools/pkg/loader"
	"sigs.k8s.io/controller-tools/pkg/markers"
)

func Test_Generate(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	rsmFiles := []string{"resource-metrics-monitor.yaml", "rbac.yaml"}
	rsmGenerator := Generator{Experimental: true, Name: "foo-metrics", Namespace: "default"}

	tests := []struct {
		name      string
		generator Generator
		// goldenDir is the directory in testdata containing the expected files.
		goldenDir string
		// files are the names of the files the generator is expected to write.
		files []string
	}{
		{
			name:      "default",
			generator: rsmGenerator,
			goldenDir: "resource-state-metrics",
			files:     rsmFiles,
		},
		{
			name: "resource-state-metrics",
			generator: Generator{
				Experimental: true,
				Target:       targetResourceStateMetrics,
				Name:         rsmGenerator.Name,
				Namespace:    rsmGenerator.Namespace,
			},
			goldenDir: "resource-state-metrics",
			files:     rsmFiles,
		},
		{
			name:      "kube-state-metrics",
			generator: Generator{Experimental: true, Target: targetKubeStateMetrics},
			goldenDir: "kube-state-metrics",
			files:     []string{"metrics.yaml", "rbac.yaml"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			optionsRegistry := &markers.Registry{}
			if err := tt.generator.RegisterMarkers(optionsRegistry); err != nil {
				t.Fatal(err)
			}

			// Load the passed packages as roots.
			roots, err := loader.LoadRoots(path.Join(cwd, "testdata", "..."))
			if err != nil {
				t.Fatalf("loading packages %v", err)
			}

			out := &outputRule{files: map[string]*bytes.Buffer{}}
			generationContext := &genall.GenerationContext{
				Collector:  &markers.Collector{Registry: optionsRegistry},
				Roots:      roots,
				Checker:    &loader.TypeChecker{},
				OutputRule: out,
			}

			t.Log("Trying to generate a custom resource configuration from the loaded packages")

			if err := tt.generator.Generate(generationContext); err != nil {
				t.Fatal(err)
			}

			written := slices.Sorted(maps.Keys(out.files))
			if diff := cmp.Diff(slices.Sorted(slices.Values(tt.files)), written); diff != "" {
				t.Fatalf("The generator wrote an unexpected set of files (-want,+got): %s", diff)
			}

			t.Log("Comparing output to testdata to check for regressions")

			for _, golden := range tt.files {
				expected, err := os.ReadFile(path.Clean(path.Join(cwd, "testdata", tt.goldenDir, golden)))
				if err != nil {
					t.Fatal(err)
				}

				generated := out.files[golden].String()
				if diff := cmp.Diff(string(expected), generated); diff != "" {
					t.Log("generated:")
					t.Log(generated)
					t.Log("diff:")
					t.Log(diff)
					t.Logf("Expected output to match file `testdata/%s/%s` but it does not.", tt.goldenDir, golden)
					t.Log("If the change is intended, use `go generate ./pkg/metrics/testdata` to regenerate the files in `testdata`.")
					t.Errorf("Detected a diff between the output of the integration test and the file `testdata/%s/%s`.", tt.goldenDir, golden)
				}
			}
		})
	}
}

// outputRule stores the written files in memory.
type outputRule struct {
	files map[string]*bytes.Buffer
}

func (o *outputRule) Open(_ *loader.Package, itemPath string) (io.WriteCloser, error) {
	buf := &bytes.Buffer{}
	o.files[itemPath] = buf
	return nopCloser{buf}, nil
}

type nopCloser struct {
	io.Writer
}

func (n nopCloser) Close() error {
	return nil
}
