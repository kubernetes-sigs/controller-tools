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

// Package testutils defines utilities for using loader.Packages
// in tests.
//
// The main utility is LoadFakeRoots, which allows writing of fake
// modules to the local filesystem, then loading those from within
// a Ginkgo test.  If not using Ginkgo, you can reproduce similar
// logic with loader.LoadWithConfig and go/packages/packagestest.
package testutils

import (
	"testing"

	"github.com/onsi/ginkgo"
	pkgstest "golang.org/x/tools/go/packages/packagestest"

	"sigs.k8s.io/controller-tools/pkg/loader"
)

// fakeT wraps Ginkgo's fake testing.T with a real testing.T
// to get the few missing methods.
// NOTE: Go 1.17 doesn't allow ambiguous embedded funcs anymore.
// We have to proxy ginkgo.GinkgoTInterface instead of testing.T
// because we have to embed testing.T's private() func.
type fakeT struct {
	g ginkgo.GinkgoTInterface
	*testing.T
}

func (t fakeT) Cleanup(f func()) {
	t.g.Cleanup(f)
}

func (t fakeT) Error(args ...interface{}) {
	t.g.Error(args)
}

func (t fakeT) Errorf(format string, args ...interface{}) {
	t.g.Errorf(format, args)
}

func (t fakeT) Fail() {
	t.g.Fail()
}

func (t fakeT) FailNow() {
	t.g.FailNow()
}

func (t fakeT) Failed() bool {
	return t.g.Failed()
}

func (t fakeT) Fatal(args ...interface{}) {
	t.g.Fatal(args)
}

func (t fakeT) Fatalf(format string, args ...interface{}) {
	t.g.Fatalf(format, args)
}

func (t fakeT) Helper() {
	t.g.Helper()
}

func (t fakeT) Log(args ...interface{}) {
	t.g.Log(args)
}

func (t fakeT) Logf(format string, args ...interface{}) {
	t.g.Logf(format, args)
}

func (t fakeT) Name() string {
	return t.g.Name()
}

func (t fakeT) Setenv(key, value string) {
	t.g.Setenv(key, value)
}

func (t fakeT) Skip(args ...interface{}) {
	t.g.Skip(args)
}

func (t fakeT) SkipNow() {
	t.g.SkipNow()
}

func (t fakeT) Skipf(format string, args ...interface{}) {
	t.g.Skipf(format, args)
}

func (t fakeT) Skipped() bool {
	return t.g.Skipped()
}

func (t fakeT) TempDir() string {
	return t.g.TempDir()
}

// LoadFakeRoots loads the given "root" packages by fake module,
// transitively loading and all imports as well.
//
// Loaded packages will have type size, imports, and exports file information
// populated.  Additional information, like ASTs and type-checking information,
// can be accessed via methods on individual packages.
func LoadFakeRoots(exporter pkgstest.Exporter, modules []pkgstest.Module, roots ...string) ([]*loader.Package, *pkgstest.Exported, error) {
	exported := pkgstest.Export(fakeT{ginkgo.GinkgoT(), &testing.T{}}, exporter, modules)

	pkgs, err := loader.LoadRootsWithConfig(exported.Config, roots...)
	return pkgs, exported, err
}
