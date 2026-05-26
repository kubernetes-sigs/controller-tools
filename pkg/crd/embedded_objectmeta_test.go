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

package crd_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-tools/pkg/crd"
	crdmarkers "sigs.k8s.io/controller-tools/pkg/crd/markers"
	"sigs.k8s.io/controller-tools/pkg/loader"
	"sigs.k8s.io/controller-tools/pkg/markers"
)

var _ = Describe("EmbeddedObjectMeta Type Registration", func() {
	var (
		reg    *markers.Registry
		parser *crd.Parser
		metav1 *loader.Package
	)

	BeforeEach(func() {
		reg = &markers.Registry{}
		Expect(crdmarkers.Register(reg)).To(Succeed())
	})

	Context("when GenerateEmbeddedObjectMeta is true", func() {
		BeforeEach(func() {
			By("setting up parser with GenerateEmbeddedObjectMeta enabled")
			parser = &crd.Parser{
				Collector:                  &markers.Collector{Registry: reg},
				Checker:                    &loader.TypeChecker{},
				GenerateEmbeddedObjectMeta: true,
			}
			crd.AddKnownTypes(parser)

			By("loading the metav1 package")
			pkgs, err := loader.LoadRoots("k8s.io/apimachinery/pkg/apis/meta/v1")
			Expect(err).NotTo(HaveOccurred())
			Expect(pkgs).To(HaveLen(1))
			metav1 = pkgs[0]

			By("processing the metav1 package")
			parser.NeedPackage(metav1)
		})

		It("should register EmbeddedObjectMeta type", func() {
			embeddedObjectMetaIdent := crd.TypeIdent{
				Name:    crd.EmbeddedObjectMetaTypeName,
				Package: metav1,
			}

			Expect(parser.Schemata).To(HaveKey(embeddedObjectMetaIdent))
		})

		It("should register ObjectMeta type", func() {
			objectMetaIdent := crd.TypeIdent{
				Name:    "ObjectMeta",
				Package: metav1,
			}

			Expect(parser.Schemata).To(HaveKey(objectMetaIdent))
		})

		It("should use the same schema for both types", func() {
			embeddedObjectMetaIdent := crd.TypeIdent{
				Name:    crd.EmbeddedObjectMetaTypeName,
				Package: metav1,
			}
			objectMetaIdent := crd.TypeIdent{
				Name:    "ObjectMeta",
				Package: metav1,
			}

			embeddedSchema := parser.Schemata[embeddedObjectMetaIdent]
			objectMetaSchema := parser.Schemata[objectMetaIdent]

			Expect(embeddedSchema).To(Equal(objectMetaSchema))
		})

		It("should have correct schema properties", func() {
			embeddedObjectMetaIdent := crd.TypeIdent{
				Name:    crd.EmbeddedObjectMetaTypeName,
				Package: metav1,
			}
			schema := parser.Schemata[embeddedObjectMetaIdent]

			By("verifying it is an object type")
			Expect(schema.Type).To(Equal("object"))

			By("verifying it has exactly 5 properties")
			Expect(schema.Properties).To(HaveLen(5))

			By("verifying name field")
			Expect(schema.Properties).To(HaveKey("name"))
			Expect(schema.Properties["name"].Type).To(Equal("string"))

			By("verifying namespace field")
			Expect(schema.Properties).To(HaveKey("namespace"))
			Expect(schema.Properties["namespace"].Type).To(Equal("string"))

			By("verifying labels field")
			Expect(schema.Properties).To(HaveKey("labels"))
			Expect(schema.Properties["labels"].Type).To(Equal("object"))
			Expect(schema.Properties["labels"].AdditionalProperties).NotTo(BeNil())
			Expect(schema.Properties["labels"].AdditionalProperties.Schema.Type).To(Equal("string"))

			By("verifying annotations field")
			Expect(schema.Properties).To(HaveKey("annotations"))
			Expect(schema.Properties["annotations"].Type).To(Equal("object"))
			Expect(schema.Properties["annotations"].AdditionalProperties).NotTo(BeNil())
			Expect(schema.Properties["annotations"].AdditionalProperties.Schema.Type).To(Equal("string"))

			By("verifying finalizers field")
			Expect(schema.Properties).To(HaveKey("finalizers"))
			Expect(schema.Properties["finalizers"].Type).To(Equal("array"))
			Expect(schema.Properties["finalizers"].Items).NotTo(BeNil())
			Expect(schema.Properties["finalizers"].Items.Schema.Type).To(Equal("string"))
		})
	})

	Context("when GenerateEmbeddedObjectMeta is false", func() {
		BeforeEach(func() {
			By("setting up parser with GenerateEmbeddedObjectMeta disabled")
			parser = &crd.Parser{
				Collector:                  &markers.Collector{Registry: reg},
				Checker:                    &loader.TypeChecker{},
				GenerateEmbeddedObjectMeta: false,
			}
			crd.AddKnownTypes(parser)

			By("loading the metav1 package")
			pkgs, err := loader.LoadRoots("k8s.io/apimachinery/pkg/apis/meta/v1")
			Expect(err).NotTo(HaveOccurred())
			Expect(pkgs).To(HaveLen(1))
			metav1 = pkgs[0]

			By("processing the metav1 package")
			parser.NeedPackage(metav1)
		})

		It("should not register EmbeddedObjectMeta type", func() {
			embeddedObjectMetaIdent := crd.TypeIdent{
				Name:    crd.EmbeddedObjectMetaTypeName,
				Package: metav1,
			}

			Expect(parser.Schemata).NotTo(HaveKey(embeddedObjectMetaIdent))
		})

		It("should register minimal ObjectMeta type", func() {
			objectMetaIdent := crd.TypeIdent{
				Name:    "ObjectMeta",
				Package: metav1,
			}

			Expect(parser.Schemata).To(HaveKey(objectMetaIdent))

			By("verifying it has minimal schema")
			schema := parser.Schemata[objectMetaIdent]
			Expect(schema.Type).To(Equal("object"))
			Expect(schema.Properties).To(BeNil())
		})
	})
})

var _ = Describe("EmbeddedObjectMetaTypeName Constant", func() {
	It("should have the correct value", func() {
		Expect(crd.EmbeddedObjectMetaTypeName).To(Equal("EmbeddedObjectMeta"))
	})
})
