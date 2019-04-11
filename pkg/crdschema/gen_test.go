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

package crdschema

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/sergi/go-diff/diffmatchpatch"

	"sigs.k8s.io/controller-tools/pkg/genall"

	// avoid go lint issue of not imported package
	_ "sigs.k8s.io/controller-tools/pkg/crdschema/fixtures/pkg/apis/kubebuilder/v1"
	_ "sigs.k8s.io/controller-tools/pkg/crdschema/fixtures/pkg/apis/legacy/v1"
)

func TestGen(t *testing.T) {
	crdSchemaGen := &Generator{
		ManifestsPath: "./fixtures/manifests",
		VerifyOnly:    false,
	}
	rt, err := genall.Generators{crdSchemaGen}.ForRoots("./fixtures/pkg/apis/...")
	if err != nil {
		t.Fatal(err)
	}

	outputDir, err := ioutil.TempDir("", "controller-tools-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(outputDir)

	rt.OutputRules.Default = genall.OutputToDirectory(outputDir)

	if someError := rt.Run(); someError {
		t.Fatal("unexpected Run() error")
	}

	t.Log("Checking for files to be written")
	expectedFiles, err := ioutil.ReadDir(filepath.Join("fixtures", "expected"))
	if err != nil {
		t.Fatal(err)
	}
	for _, fi := range expectedFiles {
		writtenFn := filepath.Join(outputDir, fi.Name())
		if _, err := os.Stat(writtenFn); err != nil {
			t.Errorf("expected %s to be written, but got: %v", writtenFn, err)
			continue
		}

		written, err := ioutil.ReadFile(writtenFn)
		if err != nil {
			t.Error(err)
			continue
		}

		expectedFn := filepath.Join("fixtures", "expected", fi.Name())
		expected, err := ioutil.ReadFile(expectedFn)
		if err != nil {
			t.Error(err)
			continue
		}

		if string(expected) != string(written) {
			dmp := diffmatchpatch.New()
			patch := dmp.DiffMain(string(expected), string(written), false)
			t.Errorf("expected file %s differs from written file %s: %s", expectedFn, writtenFn, dmp.DiffPrettyText(patch))
		}
	}

	t.Log("Checking whether unexpected files were written")
	writtenFiles, err := ioutil.ReadDir(filepath.Join(outputDir))
	if err != nil {
		t.Fatal(err)
	}
	for _, fi := range writtenFiles {
		writtenFn := filepath.Join(outputDir, fi.Name())
		expectedFn := filepath.Join("fixtures", "expected", fi.Name())
		if _, err := os.Stat(expectedFn); err != nil {
			written, err := ioutil.ReadFile(writtenFn)
			if err != nil {
				t.Error(err)
				continue
			}

			originalFn := filepath.Join("fixtures", "manifests", fi.Name())
			original, err := ioutil.ReadFile(originalFn)
			if err != nil {
				t.Error(err)
				continue
			}

			dmp := diffmatchpatch.New()
			patch := dmp.DiffMain(string(original), string(written), false)
			t.Errorf("unexpected file written: %v\n\n%s", writtenFn, dmp.DiffPrettyText(patch))
		}
	}
}
