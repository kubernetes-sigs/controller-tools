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
type fakeT struct {
	ginkgo.GinkgoTInterface
	*testing.T
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
