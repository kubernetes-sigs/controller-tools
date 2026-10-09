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

package loader

import (
	"path/filepath"
	"strings"
	"testing"
)

// On Windows, filepath.Abs resolves a path with GetFullPathName, which drops
// trailing dots: Abs(`testmod\...`) is the directory testmod itself. A
// filesystem root ending in "..." must keep its nested traversal anyway.
func TestFilesystemRootKeepsTrailingEllipsisWhenAbsDropsTrailingDots(t *testing.T) {
	original := absPath
	absPath = func(path string) (string, error) {
		abs, err := filepath.Abs(path)
		return strings.TrimSuffix(abs, string(filepath.Separator)+"..."), err
	}
	defer func() { absPath = original }()

	pkgs, err := LoadRoots("./testmod/...")
	if err != nil {
		t.Fatalf("LoadRoots: %v", err)
	}
	if len(pkgs) != 7 {
		ids := make([]string, 0, len(pkgs))
		for _, pkg := range pkgs {
			ids = append(ids, pkg.ID)
		}
		t.Fatalf("LoadRoots(./testmod/...) loaded %d packages %v, want the 7 the test suite expects", len(pkgs), ids)
	}
}
