/*
Copyright 2019 The Kubernetes Authors.

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

package crd_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"sigs.k8s.io/controller-tools/pkg/crd"
	crdmarkers "sigs.k8s.io/controller-tools/pkg/crd/markers"
	"sigs.k8s.io/controller-tools/pkg/genall"
	"sigs.k8s.io/controller-tools/pkg/loader"
	"sigs.k8s.io/controller-tools/pkg/markers"
)

// Harness encapsulates common functionality for tests.
type Harness struct {
	*testing.T
}

// NewTempDir creates a temporary directory.
// The directory (and contents) will be deleted as part of test cleanup.
func (t *Harness) NewTempDir() string {
	tempDir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatalf("error from MkdirTemp: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Errorf("error cleaning up temp directory %q: %v", tempDir, err)
		}
	})

	return tempDir
}

// MustReadFile is a wrapper around os.ReadFile that fails the test on error.
func (t *Harness) MustReadFile(p string) []byte {
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("error from ReadFile(%q): %v", p, err)
	}
	return b
}

// CompareGoldenFile compares the contents of the file at p with the actual results we got.
// It prints the diff and calls t.Errorf on an error.
// If the WRITE_GOLDEN_OUTPUT env var is set, it will write the contents to the file
// (updating the expected output).
func (t *Harness) CompareGoldenFile(p string, got string) {
	if os.Getenv("WRITE_GOLDEN_OUTPUT") != "" {
		// Short-circuit when the output is correct
		b, err := os.ReadFile(p)
		if err == nil && bytes.Equal(b, []byte(got)) {
			return
		}

		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			t.Fatalf("failed to make directories %q for writing golden output: %v", filepath.Dir(p), err)
		}

		// #nosec G306 -- expected test results are not secret
		if err := os.WriteFile(p, []byte(got), 0644); err != nil {
			t.Fatalf("failed to write golden output %s: %v", p, err)
		}
		t.Errorf("wrote output to %s", p)
	} else {
		want := string(t.MustReadFile(p))
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("unexpected diff in %s: %s", p, diff)
		}
	}
}

// CompareGoldenDirectories compares the contents of a directory with the contents of a golden directory.
// It will typically be used to compare a temp directory of output to a testdata subdirectory.
func (t *Harness) CompareGoldenDirectories(gotDir string, wantDir string) {
	gotFiles, err := os.ReadDir(gotDir)
	if err != nil {
		t.Fatalf("error from ReadDir(%q): %v", gotDir, err)
	}

	for _, gotFile := range gotFiles {
		name := gotFile.Name()

		gotFilePath := filepath.Join(gotDir, name)
		gotContents := t.MustReadFile(gotFilePath)

		wantFilePath := filepath.Join(wantDir, name)
		if _, err := os.Stat(wantFilePath); err != nil {
			if os.IsNotExist(err) {
				// Extra file in gotDir
				if os.Getenv("WRITE_GOLDEN_OUTPUT") != "" {
					t.CompareGoldenFile(wantFilePath, string(gotContents))
				} else {
					t.Errorf("extra file %q was generated", name)
				}
			} else {
				t.Errorf("error reading file %q: %v", wantFilePath, err)
			}
			continue
		}

		t.CompareGoldenFile(wantFilePath, string(gotContents))
	}

	wantFiles, err := os.ReadDir(wantDir)
	if err != nil {
		t.Fatalf("error from ReadDir(%q): %v", wantDir, err)
	}

	for _, wantFile := range wantFiles {
		name := wantFile.Name()

		wantFilePath := filepath.Join(wantDir, name)

		gotFilePath := filepath.Join(gotDir, name)
		if _, err := os.Stat(gotFilePath); err != nil {
			if os.IsNotExist(err) {
				// Extra file in wantDir
				if os.Getenv("WRITE_GOLDEN_OUTPUT") != "" {
					if err := os.Remove(wantFilePath); err != nil {
						t.Errorf("error removing file from golden output %q: %v", wantFilePath, err)
					}
				}
				t.Errorf("extra file %q in golden output directory", name)
			} else {
				t.Errorf("error reading file %q: %v", gotFilePath, err)
			}
			continue
		}
	}
}

// LoadPackages loads go packages for generation from baseDir, with loadSpec specifying the packages to load.
// loadSpec = "./..." is a typical example, and loads the packages in baseDir and subdirectories.
func (t *Harness) LoadPackages(baseDir string, loadSpec string) []*loader.Package {
	t.Logf("switching into testdata to appease go modules")
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("error from os.Getwd: %v", err)
	}
	if err := os.Chdir(baseDir); err != nil {
		t.Fatalf("error from os.Chdir(%q): %v", baseDir, err)
	}
	defer func() {
		if err := os.Chdir(cwd); err != nil {
			t.Errorf("error returning to original directory %q: %v", cwd, err)
		}
	}()

	t.Logf("loading the roots")
	pkgs, err := loader.LoadRoots(loadSpec)
	if err != nil {
		t.Errorf("error from loader.LoadRoots: %v", err)
	}

	return pkgs
}

// RunCRDGenerator will run the crd.Generator against the given roots.
// Output is sent to a temporary directory, the path to that temporary directory is returned.
func (t *Harness) RunCRDGenerator(roots []*loader.Package, gen *crd.Generator) string {
	reg := &markers.Registry{}
	if err := crdmarkers.Register(reg); err != nil {
		t.Fatalf("error from crdmarkers.Register: %v", err)
	}

	outputDir := t.NewTempDir()

	out := genall.OutputToDirectory(outputDir)

	ctx := &genall.GenerationContext{
		Collector:  &markers.Collector{Registry: reg},
		Roots:      roots,
		Checker:    &loader.TypeChecker{},
		OutputRule: out,
		InputRule:  genall.InputFromFileSystem,
	}

	if err := gen.Generate(ctx); err != nil {
		t.Fatalf("error from crd Generate: %v", err)
	}

	return outputDir
}

// TestCRDGeneration does some simple golden testing of CRD generation.
func TestGeneration(t *testing.T) {
	grid := []struct {
		name         string
		crdGenerator *crd.Generator
	}{
		{
			name: "simple",
			crdGenerator: &crd.Generator{
				CRDVersions: []string{"v1"},
			},
		},
		{
			name: "with-header-file",
			crdGenerator: &crd.Generator{
				CRDVersions: []string{"v1"},
				HeaderFile:  filepath.Join("testdata", "headers", "boilerplate.yaml"),
				Year:        "2023",
			},
		},
	}

	for _, g := range grid {
		g := g
		t.Run(g.name, func(t *testing.T) {
			h := Harness{t}

			genDir := filepath.Join("testdata", "gen")

			roots := h.LoadPackages(genDir, "./...")
			outputDir := h.RunCRDGenerator(roots, g.crdGenerator)

			wantDir := filepath.Join("testdata", "golden", g.name)
			h.CompareGoldenDirectories(outputDir, wantDir)
		})
	}
}
