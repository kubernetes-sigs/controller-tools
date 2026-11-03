/*
Copyright 2022 The Kubernetes Authors.

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

package loader_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/tools/go/packages"
	"sigs.k8s.io/controller-tools/pkg/loader"
)

var _ = Describe("Loader parsing root module", func() {
	const (
		rootPkg    = "sigs.k8s.io/controller-tools"
		pkgPkg     = rootPkg + "/pkg"
		loaderPkg  = pkgPkg + "/loader"
		testmodPkg = loaderPkg + "/testmod"
	)

	var indexOfPackage = func(pkgID string, pkgs []*loader.Package) int {
		for i := range pkgs {
			if pkgs[i].ID == pkgID {
				return i
			}
		}
		return -1
	}

	var assertPkgExists = func(pkgID string, pkgs []*loader.Package) {
		ExpectWithOffset(1, indexOfPackage(pkgID, pkgs)).Should(BeNumerically(">", -1))
	}

	Context("with named packages/modules", func() {

		// we need to change into the directory of ./testmod in order to execute
		// tests due to the inability to place the replace statement in the
		// project's root go.mod file (this breaks "go install")
		var previousWorkingDir string
		BeforeEach(func() {
			cwd, err := os.Getwd()
			Expect(err).ToNot(HaveOccurred())
			Expect(cwd).ToNot(BeEmpty())
			previousWorkingDir = cwd
			Expect(os.Chdir("./testmod")).To(Succeed())
		})
		AfterEach(func() {
			Expect(os.Chdir(previousWorkingDir)).To(Succeed())
		})

		Context("with roots=[sigs.k8s.io/controller-tools/pkg/loader/testmod/submod1]", func() {
			It("should load one package", func() {
				pkgs, err := loader.LoadRoots("sigs.k8s.io/controller-tools/pkg/loader/testmod/submod1")
				Expect(err).ToNot(HaveOccurred())
				Expect(pkgs).To(HaveLen(1))
				assertPkgExists(testmodPkg+"/submod1", pkgs)
			})
		})

		Context("with roots=[sigs.k8s.io/controller-tools/pkg/loader/testmod/...]", func() {
			It("should load six packages", func() {
				pkgs, err := loader.LoadRoots("sigs.k8s.io/controller-tools/pkg/loader/testmod/...")
				Expect(err).ToNot(HaveOccurred())
				Expect(pkgs).To(HaveLen(6))
				assertPkgExists(testmodPkg, pkgs)
				assertPkgExists(testmodPkg+"/subdir1", pkgs)
				assertPkgExists(testmodPkg+"/subdir1/subdir1", pkgs)
				assertPkgExists(testmodPkg+"/subdir1/subdir2", pkgs)
				assertPkgExists(testmodPkg+"/submod1", pkgs)
				assertPkgExists(testmodPkg+"/submod1/subdir1", pkgs)
			})
		})

		Context("with roots=[sigs.k8s.io/controller-tools/pkg/loader/testmod/..., ./...]", func() {
			It("should load seven packages", func() {
				pkgs, err := loader.LoadRoots("sigs.k8s.io/controller-tools/pkg/loader/testmod/...", "./...")
				Expect(err).ToNot(HaveOccurred())
				Expect(pkgs).To(HaveLen(7))
				assertPkgExists(testmodPkg, pkgs)
				assertPkgExists(testmodPkg+"/subdir1", pkgs)
				assertPkgExists(testmodPkg+"/subdir1/subdir1", pkgs)
				assertPkgExists(testmodPkg+"/subdir1/subdir2", pkgs)
				assertPkgExists(testmodPkg+"/subdir1/submod1", pkgs)
				assertPkgExists(testmodPkg+"/submod1", pkgs)
				assertPkgExists(testmodPkg+"/submod1/subdir1", pkgs)
			})
		})
	})

	Context("with roots=[../crd/.]", func() {
		It("should load one package", func() {
			pkgs, err := loader.LoadRoots("../crd/.")
			Expect(err).ToNot(HaveOccurred())
			Expect(pkgs).To(HaveLen(1))
			assertPkgExists(pkgPkg+"/crd", pkgs)
		})
	})

	Context("with roots=[./]", func() {
		It("should load one package", func() {
			pkgs, err := loader.LoadRoots("./")
			Expect(err).ToNot(HaveOccurred())
			Expect(pkgs).To(HaveLen(1))
			assertPkgExists(loaderPkg, pkgs)
		})
	})

	Context("with roots=[../../pkg/loader]", func() {
		It("should load one package", func() {
			pkgs, err := loader.LoadRoots("../../pkg/loader")
			Expect(err).ToNot(HaveOccurred())
			Expect(pkgs).To(HaveLen(1))
			assertPkgExists(loaderPkg, pkgs)
		})
	})

	Context("with roots=[../../pkg/loader/../loader/testmod/..., ./testmod/./../testmod//.]", func() {
		It("should load seven packages", func() {
			pkgs, err := loader.LoadRoots(
				"../../pkg/loader/../loader/testmod/...",
				"./testmod/./../testmod//.")
			Expect(err).ToNot(HaveOccurred())
			Expect(pkgs).To(HaveLen(7))
			assertPkgExists(testmodPkg, pkgs)
			assertPkgExists(testmodPkg+"/subdir1", pkgs)
			assertPkgExists(testmodPkg+"/subdir1/subdir1", pkgs)
			assertPkgExists(testmodPkg+"/subdir1/subdir2", pkgs)
			assertPkgExists(testmodPkg+"/subdir1/submod1", pkgs)
			assertPkgExists(testmodPkg+"/submod1", pkgs)
			assertPkgExists(testmodPkg+"/submod1/subdir1", pkgs)
		})
	})

	Context("with roots=[./testmod/...]", func() {
		It("should load seven packages", func() {
			pkgs, err := loader.LoadRoots("./testmod/...")
			Expect(err).ToNot(HaveOccurred())
			Expect(pkgs).To(HaveLen(7))
			assertPkgExists(testmodPkg, pkgs)
			assertPkgExists(testmodPkg+"/subdir1", pkgs)
			assertPkgExists(testmodPkg+"/subdir1/subdir1", pkgs)
			assertPkgExists(testmodPkg+"/subdir1/subdir2", pkgs)
			assertPkgExists(testmodPkg+"/subdir1/submod1", pkgs)
			assertPkgExists(testmodPkg+"/submod1", pkgs)
			assertPkgExists(testmodPkg+"/submod1/subdir1", pkgs)
		})
	})

	Context("with roots=[./testmod/subdir1/submod1/...]", func() {
		It("should load one package", func() {
			pkgs, err := loader.LoadRoots("./testmod/subdir1/submod1/...")
			Expect(err).ToNot(HaveOccurred())
			Expect(pkgs).To(HaveLen(1))
			assertPkgExists(testmodPkg+"/subdir1/submod1", pkgs)
		})
	})

	Context("with roots=[./testmod, ./testmod/submod1]", func() {
		It("should load two packages", func() {
			pkgs, err := loader.LoadRoots("./testmod", "./testmod/submod1")
			Expect(err).ToNot(HaveOccurred())
			Expect(pkgs).To(HaveLen(2))
			assertPkgExists(testmodPkg, pkgs)
			assertPkgExists(testmodPkg+"/submod1", pkgs)
		})
	})

	Context("with roots=[./testmod/submod1/subdir1/]", func() {
		It("should load one package", func() {
			pkgs, err := loader.LoadRoots("./testmod/submod1/subdir1/")
			Expect(err).ToNot(HaveOccurred())
			Expect(pkgs).To(HaveLen(1))
			assertPkgExists(testmodPkg+"/submod1/subdir1", pkgs)
		})
	})

	Context("with an absolute root ending in ...", func() {
		It("should load seven packages", func() {
			root, err := filepath.Abs("testmod")
			Expect(err).ToNot(HaveOccurred())
			pkgs, err := loader.LoadRoots(filepath.Join(root, "..."))
			Expect(err).ToNot(HaveOccurred())
			Expect(pkgs).To(HaveLen(7))
			assertPkgExists(testmodPkg, pkgs)
			assertPkgExists(testmodPkg+"/subdir1/submod1", pkgs)
			assertPkgExists(testmodPkg+"/submod1/subdir1", pkgs)
		})
	})

	Context("with cfg.Dir=testmod and roots=[./...]", func() {
		It("should load seven packages", func() {
			dir, err := filepath.Abs("testmod")
			Expect(err).ToNot(HaveOccurred())
			pkgs, err := loader.LoadRootsWithConfig(&packages.Config{Dir: dir}, "./...")
			Expect(err).ToNot(HaveOccurred())
			Expect(pkgs).To(HaveLen(7))
			assertPkgExists(testmodPkg, pkgs)
			assertPkgExists(testmodPkg+"/subdir1/submod1", pkgs)
			assertPkgExists(testmodPkg+"/submod1/subdir1", pkgs)
		})
	})

	Context("when filepath.Abs drops trailing dots (Windows GetFullPathName)", func() {
		// On Windows, filepath.Abs resolves a path with GetFullPathName, which drops
		// trailing dots: Abs(`testmod\...`) is the directory testmod itself. A
		// filesystem root ending in "..." must keep its nested traversal anyway.
		It("should keep the trailing ellipsis on a filesystem root when filepath.Abs drops trailing dots", func() {
			restore := loader.SetAbsPath(func(path string) (string, error) {
				abs, err := filepath.Abs(path)
				return strings.TrimSuffix(abs, string(filepath.Separator)+"..."), err
			})
			DeferCleanup(restore)

			pkgs, err := loader.LoadRoots("./testmod/...")
			Expect(err).ToNot(HaveOccurred())
			Expect(pkgs).To(HaveLen(7))
			assertPkgExists(testmodPkg, pkgs)
			assertPkgExists(testmodPkg+"/subdir1", pkgs)
			assertPkgExists(testmodPkg+"/subdir1/subdir1", pkgs)
			assertPkgExists(testmodPkg+"/subdir1/subdir2", pkgs)
			assertPkgExists(testmodPkg+"/subdir1/submod1", pkgs)
			assertPkgExists(testmodPkg+"/submod1", pkgs)
			assertPkgExists(testmodPkg+"/submod1/subdir1", pkgs)
		})
	})

	Context("absRoot normalisation", func() {
		// Where filepath.Abs keeps a trailing "...", as it does on Linux and macOS,
		// absRoot returns what the loader computed before it set the "..." aside:
		// the cleaned root, joined to cfgDir or made absolute with filepath.Abs.
		It("should match clean then abs where filepath.Abs keeps trailing ellipsis", func() {
			if abs, err := filepath.Abs(filepath.Join("x", "...")); err != nil || filepath.Base(abs) != "..." {
				Skip(fmt.Sprintf("filepath.Abs drops a trailing \"...\" on this platform (%q, %v)", abs, err))
			}
			sep := string(filepath.Separator)
			absDir, err := filepath.Abs("testmod")
			Expect(err).ToNot(HaveOccurred())
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
						} else {
							var err error
							want, err = filepath.Abs(want)
							Expect(err).ToNot(HaveOccurred())
						}
					}
					got, err := loader.AbsRoot(root, cfgDir)
					Expect(err).ToNot(HaveOccurred(), "AbsRoot(%q, %q)", root, cfgDir)
					Expect(got).To(Equal(want), "AbsRoot(%q, %q)", root, cfgDir)
				}
			}
		})

		// A root that ends in "..." keeps it whatever filepath.Abs does to trailing
		// dots, with and without cfgDir.
		It("should keep trailing ellipsis with and without cfgDir", func() {
			restore := loader.SetAbsPath(func(path string) (string, error) {
				abs, err := filepath.Abs(path)
				return strings.TrimSuffix(abs, string(filepath.Separator)+"..."), err
			})
			DeferCleanup(restore)

			base, err := filepath.Abs("testmod")
			Expect(err).ToNot(HaveOccurred())
			for _, tc := range []struct{ root, cfgDir, want string }{
				{"./testmod/...", "", filepath.Join(base, "...")},
				{"testmod/./...", "", filepath.Join(base, "...")},
				{"./...", base, filepath.Join(base, "...")},
				{filepath.Join(base, "..."), "", filepath.Join(base, "...")},
			} {
				got, err := loader.AbsRoot(tc.root, tc.cfgDir)
				Expect(err).ToNot(HaveOccurred(), "AbsRoot(%q, %q)", tc.root, tc.cfgDir)
				Expect(got).To(Equal(tc.want), "AbsRoot(%q, %q)", tc.root, tc.cfgDir)
			}
		})
	})
})
