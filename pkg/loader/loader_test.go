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
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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

	Context("with roots=[./testmod/dummy.go]", func() {
		It("should load one package from individual file", func() {
			By("loading individual file path")
			pkgs, err := loader.LoadRoots("./testmod/dummy.go")
			Expect(err).ToNot(HaveOccurred())
			Expect(pkgs).To(HaveLen(1))
			By("verifying package ID is command-line-arguments for individual files")
			Expect(pkgs[0].ID).To(Equal("command-line-arguments"))
			Expect(pkgs[0].Name).To(Equal("dummy"))
		})
	})

	Context("with roots=[./testmod/submod1/dummy.go]", func() {
		It("should load one package from individual file", func() {
			By("loading individual file from subdirectory")
			pkgs, err := loader.LoadRoots("./testmod/submod1/dummy.go")
			Expect(err).ToNot(HaveOccurred())
			Expect(pkgs).To(HaveLen(1))
			By("verifying package ID is command-line-arguments for individual files")
			Expect(pkgs[0].ID).To(Equal("command-line-arguments"))
			Expect(pkgs[0].Name).To(Equal("dummy"))
		})
	})

	Context("with roots=[./testmod/subdir1/dummy.go]", func() {
		It("should load one package from individual file in subdirectory", func() {
			By("loading individual file from nested subdirectory")
			pkgs, err := loader.LoadRoots("./testmod/subdir1/dummy.go")
			Expect(err).ToNot(HaveOccurred())
			Expect(pkgs).To(HaveLen(1))
			Expect(pkgs[0].ID).To(Equal("command-line-arguments"))
			Expect(pkgs[0].Name).To(Equal("dummy"))
		})
	})

	Context("with roots=[./testmod/dummy.go, ./testmod/submod1]", func() {
		It("should load packages from both file and directory", func() {
			By("loading mixed file and directory paths")
			pkgs, err := loader.LoadRoots("./testmod/dummy.go", "./testmod/submod1")
			Expect(err).ToNot(HaveOccurred())
			Expect(pkgs).To(HaveLen(2))
			By("verifying first package from individual file")
			Expect(pkgs[0].ID).To(Equal("command-line-arguments"))
			Expect(pkgs[0].Name).To(Equal("dummy"))
			By("verifying second package from directory")
			assertPkgExists(testmodPkg+"/submod1", pkgs)
		})
	})

	Context("with roots=[./testmod/subdir1/../dummy.go]", func() {
		It("should load one package from individual file with relative path", func() {
			pkgs, err := loader.LoadRoots("./testmod/subdir1/../dummy.go")
			Expect(err).ToNot(HaveOccurred())
			Expect(pkgs).To(HaveLen(1))
			Expect(pkgs[0].ID).To(Equal("command-line-arguments"))
			Expect(pkgs[0].Name).To(Equal("dummy"))
		})
	})

	Context("with roots=[./testmod/nonexistent.go]", func() {
		It("should return error for non-existent file", func() {
			_, err := loader.LoadRoots("./testmod/nonexistent.go")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unable to stat path"))
		})
	})
})
