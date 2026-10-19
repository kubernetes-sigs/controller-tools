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

// Where filepath.Abs keeps a trailing "...", as it does on Linux and macOS,
// absRoot returns what the loader computed before it set the "..." aside:
// the cleaned root, joined to cfgDir or made absolute with filepath.Abs.
func TestAbsRootMatchesCleanThenAbs(t *testing.T) {
	if abs, err := filepath.Abs(filepath.Join("x", "...")); err != nil || filepath.Base(abs) != "..." {
		t.Skipf("filepath.Abs drops a trailing \"...\" on this platform (%q, %v)", abs, err)
	}
	sep := string(filepath.Separator)
	absDir, err := filepath.Abs("testmod")
	if err != nil {
		t.Fatalf("filepath.Abs: %v", err)
	}
	roots := []string{
		"./testmod/...",
		"testmod/...",
		"...",
		"./...",
		"../x/...",
		absDir + sep + "...",
		absDir,
		"./testmod",
		"testmod/./../testmod//.",
		"x/.../y",
		"x/....",
		"x/.../",
		"testmod/",
		".",
	}
	for _, cfgDir := range []string{"", absDir} {
		for _, root := range roots {
			want := filepath.Clean(root)
			if !filepath.IsAbs(want) {
				if cfgDir != "" {
					want = filepath.Join(cfgDir, want)
				} else if want, err = filepath.Abs(want); err != nil {
					t.Fatalf("filepath.Abs(%q): %v", root, err)
				}
			}
			got, err := absRoot(root, cfgDir)
			if err != nil {
				t.Fatalf("absRoot(%q, %q): %v", root, cfgDir, err)
			}
			if got != want {
				t.Errorf("absRoot(%q, %q) = %q, want %q", root, cfgDir, got, want)
			}
		}
	}
}

// A root that ends in "..." keeps it whatever filepath.Abs does to trailing
// dots, with and without cfgDir.
func TestAbsRootKeepsTrailingEllipsis(t *testing.T) {
	original := absPath
	absPath = func(path string) (string, error) {
		abs, err := filepath.Abs(path)
		return strings.TrimSuffix(abs, string(filepath.Separator)+"..."), err
	}
	defer func() { absPath = original }()

	base, err := filepath.Abs("testmod")
	if err != nil {
		t.Fatalf("filepath.Abs: %v", err)
	}
	for _, tc := range []struct{ root, cfgDir, want string }{
		{"./testmod/...", "", filepath.Join(base, "...")},
		{"testmod/./...", "", filepath.Join(base, "...")},
		{"./...", base, filepath.Join(base, "...")},
		{filepath.Join(base, "..."), "", filepath.Join(base, "...")},
	} {
		got, err := absRoot(tc.root, tc.cfgDir)
		if err != nil {
			t.Fatalf("absRoot(%q, %q): %v", tc.root, tc.cfgDir, err)
		}
		if got != tc.want {
			t.Errorf("absRoot(%q, %q) = %q, want %q", tc.root, tc.cfgDir, got, tc.want)
		}
	}
}
