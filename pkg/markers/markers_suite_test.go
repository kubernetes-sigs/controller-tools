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

package markers_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	pkgstest "golang.org/x/tools/go/packages/packagestest"
	"sigs.k8s.io/controller-tools/pkg/loader"
	testloader "sigs.k8s.io/controller-tools/pkg/loader/testutils"
)

func TestMarkers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Markers Suite")
}

// do this once because it creates files

var fakePkg *loader.Package
var exported *pkgstest.Exported

var _ = BeforeSuite(func() {
	modules := []pkgstest.Module{
		{
			Name: "sigs.k8s.io/controller-tools/pkg/markers/testdata",
			Files: map[string]interface{}{
				"file.go": `
					package testdata

					import (
						// nothing here should be parsed

						// +testing:pkglvl="not here import 1"

						// +testing:pkglvl="not here import 2"
						foo "fmt" // +testing:pkglvl="not here import 3"

						// +testing:pkglvl="not here import 4"
					) 

					// +testing:pkglvl="here unattached"

					// +testing:pkglvl="here reassociated"
					// +testing:typelvl="here before type"
					// +testing:eitherlvl="here not reassociated"
					// +testing:fieldlvl="not here not near field"

					// normal godoc
					// +testing:pkglvl="not here godoc"
					// +testing:typelvl="here on type"
					// normal godoc
					type Foo struct {
						// +testing:pkglvl="not here in struct"
						// +testing:fieldlvl="here before godoc"

						// normal godoc
						// +testing:fieldlvl="here in godoc"
						// +testing:pkglvl="not here godoc"
						// normal godoc
						WithGodoc string // +testing:fieldlvl="not here after field"

						// +testing:fieldlvl="here without godoc"

						WithoutGodoc int
					} // +testing:pkglvl="not here after type"

					// +testing:pkglvl="not here on var"
					var (
						// +testing:pkglvl="not here in var"
						Bar = "foo"
					)

					type Baz interface {
						// +testing:pkglvl="not here in interface"
					}

					// +testing:typelvl="not here beyond closest"

					// +testing:typelvl="here without godoc"
					// +testing:pkglvl="here reassociated no godoc"

					type Quux string

					// +testing:pkglvl="here at end after last node"

					/* +testing:pkglvl="not here in block" */

					// +testing:typelvl="here on typedecl with no more"
					type Cheese struct { }
				`,
			},
		},
	}

	By("setting up the fake packages")
	var pkgs []*loader.Package
	var err error
	pkgs, exported, err = testloader.LoadFakeRoots(pkgstest.Modules, modules, "sigs.k8s.io/controller-tools/pkg/markers/testdata")
	Expect(err).NotTo(HaveOccurred())
	Expect(pkgs).To(HaveLen(1))

	fakePkg = pkgs[0]
})

var _ = AfterSuite(func() {
	By("cleaning up the fake packages")
	exported.Cleanup()
})
