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
	"encoding/json"
	"slices"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiserverschema "k8s.io/apiextensions-apiserver/pkg/apiserver/schema"
	"k8s.io/apiextensions-apiserver/pkg/apiserver/schema/defaulting"
	"k8s.io/apiextensions-apiserver/pkg/apiserver/validation"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"sigs.k8s.io/controller-tools/pkg/crd"
	crdmarkers "sigs.k8s.io/controller-tools/pkg/crd/markers"
	"sigs.k8s.io/controller-tools/pkg/loader"
	"sigs.k8s.io/controller-tools/pkg/markers"
)

// These specs cover +kubebuilder:default={} on a pointer field whose nested
// fields have their own defaults. They check that the parent renders as
// default: {} (not default: null), that the API server fills in the nested
// defaults at runtime, and that a required child makes the API server reject
// the empty default.
var _ = Describe("Empty object defaulting (+kubebuilder:default={})", func() {
	Context("when a pointer field defaults to {} and its nested fields have their own defaults", func() {
		It("generates default: {} for the parent, not null, with the nested defaults intact", func() {
			parent := specProperty("EmptyObjectDefault", "policyAuditConfig")

			Expect(parent.Default).NotTo(BeNil(), "parent default must not regress to null")
			Expect(string(parent.Default.Raw)).To(Equal("{}"))

			for field, want := range map[string]string{
				"rateLimit":      "20",
				"maxFileSize":    "50",
				"syslogFacility": `"local0"`,
				"destination":    `"null"`,
			} {
				prop, ok := parent.Properties[field]
				Expect(ok).To(BeTrue(), "nested field %q should be present", field)
				Expect(prop.Default).NotTo(BeNil(), "nested field %q should keep its default", field)
				Expect(string(prop.Default.Raw)).To(Equal(want), "nested field %q default", field)
			}
		})

		It("populates the nested defaults at runtime via the API server, with no pre-merging in the CRD", func() {
			parent := specProperty("EmptyObjectDefault", "policyAuditConfig")
			structural := newStructuralFromV1Schema(parent)

			// An object that omits the field still comes back fully populated.
			obj := map[string]any{}
			defaulting.Default(obj, structural)

			got, err := json.Marshal(obj)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(got)).To(Equal(`{"destination":"null","maxFileSize":50,"rateLimit":20,"syslogFacility":"local0"}`))
		})
	})

	Context("when the parent defaults to {} but a child field is required", func() {
		It("still generates default: {}, but that empty default is invalid because the required child is missing", func() {
			status := specProperty("RequiredChildDefault", "status")

			// controller-gen still emits default: {} with the required child.
			Expect(status.Default).NotTo(BeNil())
			Expect(string(status.Default.Raw)).To(Equal("{}"))
			Expect(slices.Contains(status.Required, "stage")).To(BeTrue())

			// The {} default does not satisfy the schema, so the API server rejects
			// such a CRD. This is API server behavior, not a controller-gen bug.
			validator, _, err := validation.NewSchemaValidator(toInternalSchema(status))
			Expect(err).NotTo(HaveOccurred())

			errs := validation.ValidateCustomResource(nil, map[string]any{}, validator).ToAggregate()
			Expect(errs).To(HaveOccurred())
			Expect(errs.Error()).To(ContainSubstring("stage"))
		})
	})
})

// generateCRD runs controller-gen over the testdata package and returns the CRD
// for the given kind.
func generateCRD(kind string) apiextensionsv1.CustomResourceDefinition {
	reg := &markers.Registry{}
	Expect(crdmarkers.Register(reg)).To(Succeed())

	parser := &crd.Parser{
		Collector:           &markers.Collector{Registry: reg},
		Checker:             &loader.TypeChecker{},
		AllowDangerousTypes: true, // the scenario uses uint32
	}
	crd.AddKnownTypes(parser)

	pkgs, err := loader.LoadRoots("./testdata/emptyobjectdefault/...")
	Expect(err).NotTo(HaveOccurred())
	for _, p := range pkgs {
		parser.NeedPackage(p)
	}

	gk := schema.GroupKind{Group: "testdata.kubebuilder.io", Kind: kind}
	parser.NeedCRDFor(gk, nil)

	crdObj, ok := parser.CustomResourceDefinitions[gk]
	Expect(ok).To(BeTrue(), "a CRD should be generated for %v", gk)
	crd.FixTopLevelMetadata(crdObj)
	return crdObj
}

// specProperty returns the generated schema for a named property of spec.
func specProperty(kind, property string) apiextensionsv1.JSONSchemaProps {
	crdObj := generateCRD(kind)
	spec := crdObj.Spec.Versions[0].Schema.OpenAPIV3Schema.Properties["spec"]
	return spec.Properties[property]
}

// toInternalSchema converts a generated v1 schema into the internal type the
// API server validation and defaulting packages work on.
func toInternalSchema(in apiextensionsv1.JSONSchemaProps) *apiextensions.JSONSchemaProps {
	out := &apiextensions.JSONSchemaProps{}
	Expect(apiextensionsv1.Convert_v1_JSONSchemaProps_To_apiextensions_JSONSchemaProps(&in, out, nil)).To(Succeed())
	return out
}

// newStructuralFromV1Schema builds the structural schema API server defaulting
// works on from a generated v1 schema.
func newStructuralFromV1Schema(in apiextensionsv1.JSONSchemaProps) *apiserverschema.Structural {
	structural, err := apiserverschema.NewStructural(toInternalSchema(in))
	Expect(err).NotTo(HaveOccurred())
	return structural
}
