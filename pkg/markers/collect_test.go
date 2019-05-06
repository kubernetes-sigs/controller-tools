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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "sigs.k8s.io/controller-tools/pkg/markers"
)

type fieldPath struct {
	typ   string
	field string
}

var _ = Describe("Collecting", func() {
	var col *Collector
	var markersByType map[string]MarkerValues
	var markersByField map[fieldPath]MarkerValues

	JustBeforeEach(func() {
		By("setting up the registry and collector")
		reg := &Registry{}

		mustDefine(reg, "testing:pkglvl", DescribesPackage, "")
		mustDefine(reg, "testing:typelvl", DescribesType, "")
		mustDefine(reg, "testing:fieldlvl", DescribesField, "")
		mustDefine(reg, "testing:eitherlvl", DescribesType, "")
		mustDefine(reg, "testing:eitherlvl", DescribesPackage, "")

		col = &Collector{Registry: reg}

		By("gathering markers by type name")
		markersByType = make(map[string]MarkerValues)
		markersByField = make(map[fieldPath]MarkerValues)
		err := EachType(col, fakePkg, func(info *TypeInfo) {
			markersByType[info.Name] = info.Markers
			for _, field := range info.Fields {
				markersByField[fieldPath{typ: info.Name, field: field.Name}] = field.Markers
			}
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(fakePkg.Errors).To(HaveLen(0))
	})

	Context("of package-level markers", func() {

		It("should consider markers anywhere not obviously type- or field-level as package-level", func() {
			By("grabbing all package-level markers")
			pkgMarkers, err := PackageMarkers(col, fakePkg)
			Expect(err).NotTo(HaveOccurred())
			Expect(fakePkg.Errors).To(HaveLen(0))

			By("checking that it only contains package-level markers")
			Expect(pkgMarkers).NotTo(HaveKey("testing:typelvl"))

			By("checking that it contains the right package-level markers")
			Expect(pkgMarkers).To(HaveKeyWithValue("testing:pkglvl", ContainElement("here unattached")))
			Expect(pkgMarkers).To(HaveKeyWithValue("testing:pkglvl", ContainElement("here at end after last node")))

			By("checking that it doesn't contain any markers it's not supposed to")
			Expect(pkgMarkers).To(HaveKeyWithValue("testing:pkglvl", Not(ContainElement(ContainSubstring("not here")))))
		})
	})

	Context("of type-level markers", func() {
		It("should not have package-level markers associated", func() {
			Expect(markersByType).To(HaveKeyWithValue("Foo", Not(HaveKey("testing:pkglvl"))))
		})

		It("should not contain markers it's not supposed to", func() {
			Expect(markersByType).NotTo(ContainElement(HaveKeyWithValue("testing:typelvl", ContainElement(ContainSubstring("not here")))))
			Expect(markersByType).NotTo(ContainElement(HaveKeyWithValue("testing:eitherlvl", ContainElement(ContainSubstring("not here")))))
		})

		Context("with godoc", func() {
			It("should associate markers in godoc", func() {
				Expect(markersByType).To(HaveKeyWithValue("Foo",
					HaveKeyWithValue("testing:typelvl", ContainElement("here on type"))))
			})

			It("should associate markers in the closest non-godoc block", func() {
				Expect(markersByType).To(HaveKeyWithValue("Foo",
					HaveKeyWithValue("testing:typelvl", ContainElement("here before type"))))
			})

			It("should re-associate package-level markers from the closest non-godoc", func() {
				By("grabbing package markers")
				pkgMarkers, err := PackageMarkers(col, fakePkg)
				Expect(err).NotTo(HaveOccurred())
				Expect(fakePkg.Errors).To(HaveLen(0))

				By("checking that the marker got reassociated")
				Expect(pkgMarkers).To(HaveKeyWithValue("testing:pkglvl", ContainElement("here reassociated")))
			})
		})
		Context("without godoc", func() {
			It("should associate markers in the closest block", func() {
				Expect(markersByType).To(HaveKeyWithValue("Quux",
					HaveKeyWithValue("testing:typelvl", ContainElement("here without godoc"))))
			})

			It("should re-associate package-level markers from the closest block", func() {
				By("grabbing package markers")
				pkgMarkers, err := PackageMarkers(col, fakePkg)
				Expect(err).NotTo(HaveOccurred())
				Expect(fakePkg.Errors).To(HaveLen(0))

				By("checking that the marker got reassociated")
				Expect(pkgMarkers).To(HaveKeyWithValue("testing:pkglvl", ContainElement("here reassociated no godoc")))

			})
		})

		It("should not re-associate package-level markers in godoc", func() {
			By("grabbing package markers")
			pkgMarkers, err := PackageMarkers(col, fakePkg)
			Expect(err).NotTo(HaveOccurred())
			Expect(fakePkg.Errors).To(HaveLen(0))

			By("checking that the marker didn't get reassociated")
			Expect(pkgMarkers).To(HaveKeyWithValue("testing:pkglvl", Not(ContainElement("not here godoc"))))
		})

		It("should not re-associate package-level markers if there's also a type-level marker", func() {
			By("checking that the type has the marker")
			Expect(markersByType).To(HaveKeyWithValue("Foo",
				HaveKeyWithValue("testing:eitherlvl", ContainElement("here not reassociated"))))

			By("grabbing package markers")
			pkgMarkers, err := PackageMarkers(col, fakePkg)
			Expect(err).NotTo(HaveOccurred())
			Expect(fakePkg.Errors).To(HaveLen(0))

			By("checking that the marker didn't get reassociated to the package")
			Expect(pkgMarkers).To(SatisfyAny(
				Not(HaveKey("testing:eitherlvl")),
				HaveKeyWithValue("testing:eitherlvl", Not(ContainElement("here not reassociated")))))

		})

		It("should consider markers on the gendecl even if there are no more markers in the file", func() {
			Expect(markersByType).To(HaveKeyWithValue("Cheese",
				HaveKeyWithValue("testing:typelvl", ContainElement("here on typedecl with no more"))))

		})
	})

	Context("of field-level markers", func() {
		It("should not contain markers it's not supposed to", func() {
			Expect(markersByField).NotTo(ContainElement(HaveKeyWithValue("testing:fieldlvl", ContainElement(ContainSubstring("not here")))))
		})
		Context("with godoc", func() {
			It("should associate markers in godoc", func() {
				Expect(markersByField).To(HaveKeyWithValue(fieldPath{typ: "Foo", field: "WithGodoc"},
					HaveKeyWithValue("testing:fieldlvl", ContainElement("here in godoc"))))
			})

			It("should associate markers in the closest non-godoc block", func() {
				Expect(markersByField).To(HaveKeyWithValue(fieldPath{typ: "Foo", field: "WithGodoc"},
					HaveKeyWithValue("testing:fieldlvl", ContainElement("here before godoc"))))
			})
		})
		Context("without godoc", func() {
			It("should associate markers in the closest block", func() {
				Expect(markersByField).To(HaveKeyWithValue(fieldPath{typ: "Foo", field: "WithoutGodoc"},
					HaveKeyWithValue("testing:fieldlvl", ContainElement("here without godoc"))))
			})
		})

		It("should not consider markers at the end of a line", func() {
			Expect(markersByField).To(HaveKeyWithValue(fieldPath{typ: "Foo", field: "WithGodoc"},
				HaveKeyWithValue("testing:fieldlvl", Not(ContainElement("not here after field")))))

			Expect(markersByField).To(HaveKeyWithValue(fieldPath{typ: "Foo", field: "WithoutGodoc"},
				HaveKeyWithValue("testing:fieldlvl", Not(ContainElement("not here after field")))))
		})
	})
})
