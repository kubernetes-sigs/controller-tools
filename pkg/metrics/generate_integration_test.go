/*
Copyright 2023 The Kubernetes Authors All rights reserved.

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
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"sigs.k8s.io/controller-tools/pkg/genall"
	"sigs.k8s.io/controller-tools/pkg/loader"
	"sigs.k8s.io/controller-tools/pkg/markers"
	"sigs.k8s.io/controller-tools/pkg/metrics/internal/config"
)

func Test_Generate(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}

	optionsRegistry := &markers.Registry{}

	metricGenerator := Generator{}
	if err := metricGenerator.RegisterMarkers(optionsRegistry); err != nil {
		t.Error(err)
	}

	out := &outputRule{
		buf: &bytes.Buffer{},
	}

	// Load the passed packages as roots.
	roots, err := loader.LoadRoots(path.Join(cwd, "testdata", "..."))
	if err != nil {
		t.Errorf("loading packages %v", err)
	}

	gen := Generator{}

	generationContext := &genall.GenerationContext{
		Collector:  &markers.Collector{Registry: optionsRegistry},
		Roots:      roots,
		Checker:    &loader.TypeChecker{},
		OutputRule: out,
	}

	t.Log("Trying to generate a custom resource configuration from the loaded packages")

	if err := gen.Generate(generationContext); err != nil {
		t.Error(err)
	}

	output := strings.Split(out.buf.String(), "\n---\n")

	header := fmt.Sprintf(headerText, "(devel)", config.KubeStateMetricsVersion)

	if len(output) != 3 {
		t.Error("Expected two output files, metrics configuration followed by rbac.")
		return
	}

	generatedData := map[string]string{
		"metrics.yaml": header + "---\n" + string(output[1]),
		"rbac.yaml":    "---\n" + string(output[2]),
	}

	t.Log("Comparing output to testdata to check for regressions")

	for _, golden := range []string{"metrics.yaml", "rbac.yaml"} {
		// generatedRaw := strings.TrimSpace(output[i])

		expectedRaw, err := os.ReadFile(path.Clean(path.Join(cwd, "testdata", golden)))
		if err != nil {
			t.Error(err)
			return
		}

		// Remove leading `---` and trim newlines
		generated := strings.TrimSpace(strings.TrimPrefix(generatedData[golden], "---"))
		expected := strings.TrimSpace(strings.TrimPrefix(string(expectedRaw), "---"))

		diff := cmp.Diff(expected, generated)
		if diff != "" {
			t.Log("generated:")
			t.Log(generated)
			t.Log("diff:")
			t.Log(diff)
			t.Logf("Expected output to match file `testdata/%s` but it does not.", golden)
			t.Logf("If the change is intended, use `go generate ./pkg/metrics/testdata` to regenerate the `testdata/%s` file.", golden)
			t.Errorf("Detected a diff between the output of the integration test and the file `testdata/%s`.", golden)
			return
		}
	}
}

type outputRule struct {
	buf *bytes.Buffer
}

func (o *outputRule) Open(_ *loader.Package, _ string) (io.WriteCloser, error) {
	return nopCloser{o.buf}, nil
}

type nopCloser struct {
	io.Writer
}

func (n nopCloser) Close() error {
	return nil
}
