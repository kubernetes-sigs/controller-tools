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

package crd_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-tools/pkg/crd"
)

var _ = Describe("CRD Generation", func() {
	Describe("check storage version", func() {

		var (
			groupKind schema.GroupKind
		)

		BeforeEach(func() {
			groupKind = schema.GroupKind{
				Group: "TestKind",
				Kind:  "test.example.com",
			}
		})

		It("should return an error when multiple storage versions are present", func() {
			obj := &apiext.CustomResourceDefinition{
				Spec: apiext.CustomResourceDefinitionSpec{
					Versions: []apiext.CustomResourceDefinitionVersion{
						{
							Name:    "v1",
							Storage: true,
							Served:  true,
						},
						{
							Name:    "v2",
							Storage: true,
						},
					},
				},
			}
			listOferrors := crd.ValidateStorageVersionAndServe(*obj, groupKind)
			Expect(listOferrors).Should(ContainElement(fmt.Errorf("CRD for %s has more than one storage version", groupKind)))
		})
		It("should not return any errors when only one version is present", func() {
			obj := &apiext.CustomResourceDefinition{
				Spec: apiext.CustomResourceDefinitionSpec{
					Versions: []apiext.CustomResourceDefinitionVersion{
						{
							Name:   "v1",
							Served: true,
						},
					},
				},
			}
			listOferrors := crd.ValidateStorageVersionAndServe(*obj, groupKind)
			Expect(listOferrors).To(Equal([]error{}))
		})
		It("should return an error when no storage version is present", func() {
			obj := &apiext.CustomResourceDefinition{
				Spec: apiext.CustomResourceDefinitionSpec{
					Versions: []apiext.CustomResourceDefinitionVersion{
						{
							Name:    "v1",
							Storage: false,
							Served:  true,
						},
						{
							Name:    "v2",
							Storage: false,
						},
					},
				},
			}
			listOferrors := crd.ValidateStorageVersionAndServe(*obj, groupKind)
			Expect(listOferrors).Should(ContainElement(fmt.Errorf("CRD for %s has no storage version", groupKind)))
		})
		It("should return an error when no served versions are present", func() {
			obj := &apiext.CustomResourceDefinition{
				Spec: apiext.CustomResourceDefinitionSpec{
					Versions: []apiext.CustomResourceDefinitionVersion{
						{
							Name:    "v1",
							Storage: true,
						},
						{
							Name:    "v2",
							Storage: false,
						},
					},
				},
			}
			listOferrors := crd.ValidateStorageVersionAndServe(*obj, groupKind)
			Expect(listOferrors).Should(ContainElement(fmt.Errorf("CRD for %s with version(s) %v does not serve any version", groupKind, obj.Spec.Versions)))
		})
		It("should not return any errors when one served storage version is present", func() {
			obj := &apiext.CustomResourceDefinition{
				Spec: apiext.CustomResourceDefinitionSpec{
					Versions: []apiext.CustomResourceDefinitionVersion{
						{
							Name:    "v1",
							Storage: false,
						},
						{
							Name:    "v2",
							Storage: true,
							Served:  true,
						},
					},
				},
			}
			listOferrors := crd.ValidateStorageVersionAndServe(*obj, groupKind)
			Expect(listOferrors).To(Equal([]error{}))
		})
		It("should not return any errors when multiple served storage version is present", func() {
			obj := &apiext.CustomResourceDefinition{
				Spec: apiext.CustomResourceDefinitionSpec{
					Versions: []apiext.CustomResourceDefinitionVersion{
						{
							Name:    "v1",
							Storage: false,
							Served:  true,
						},
						{
							Name:    "v2",
							Storage: true,
							Served:  true,
						},
					},
				},
			}
			listOferrors := crd.ValidateStorageVersionAndServe(*obj, groupKind)
			Expect(listOferrors).To(Equal([]error{}))
		})
	})
})
