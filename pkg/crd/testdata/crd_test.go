/*
Copyright The Kubernetes Authors.

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
package cronjob

import (
	"context"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/yaml"
)

var _ = Describe("CronJob CRD", func() {
	It("should be successfully applied", func(ctx SpecContext) {
		data, err := os.ReadFile("testdata.kubebuilder.io_cronjobs.yaml")
		Expect(err).NotTo(HaveOccurred())

		crd := &apiextensionsv1.CustomResourceDefinition{}
		err = yaml.UnmarshalStrict(data, crd)
		Expect(err).NotTo(HaveOccurred())

		err = k8sClient.Create(ctx, crd)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("validating opaque markers", func() {
		applyCronJob := func(ctx context.Context, name, opaqueVal, nonOpaqueVal string) error {
			obj := &unstructured.Unstructured{}
			obj.SetGroupVersionKind(schema.GroupVersionKind{
				Group:   "testdata.kubebuilder.io",
				Version: "v1",
				Kind:    "CronJob",
			})
			obj.SetName(name)
			obj.SetNamespace("default")

			spec := map[string]interface{}{
				"schedule":                      "*/5 * * * *", // required field
				"foo":                           "bar",
				"baz":                           "baz",
				"binaryName":                    "YmluYXJ5", // base64 for "binary"
				"canBeNull":                     "ok",
				"defaultedEmptyMap":             map[string]interface{}{},
				"defaultedEmptyObject":          map[string]interface{}{},
				"defaultedEmptySlice":           []interface{}{},
				"defaultedObject":               []interface{}{},
				"defaultedSlice":                []interface{}{},
				"defaultedString":               "some string",
				"doubleDefaultedString":         "some string",
				"embeddedResource":              map[string]interface{}{"kind": "Pod", "apiVersion": "v1"},
				"explicitlyRequiredK8s":         "required",
				"explicitlyRequiredKubebuilder": "required",
				"explicitlyRequiredKubernetes":  "required",
				"float64WithValidations":        1.5,
				"floatWithValidations":          1.5,
				"int32WithValidations":          2,
				"intWithValidations":            2,
				"jobTemplate": map[string]interface{}{
					"template": map[string]interface{}{},
				},
				"kubernetesDefaultedEmptyMap":               map[string]interface{}{},
				"kubernetesDefaultedEmptyObject":            map[string]interface{}{},
				"kubernetesDefaultedEmptySlice":             []interface{}{},
				"kubernetesDefaultedObject":                 []interface{}{},
				"kubernetesDefaultedSlice":                  []interface{}{},
				"kubernetesDefaultedString":                 "string",
				"mapOfInfo":                                 map[string]interface{}{},
				"nestedMapOfInfo":                           map[string]interface{}{},
				"nestedStructWithSeveralFields":             map[string]interface{}{"foo": "str", "bar": true},
				"nestedStructWithSeveralFieldsDoubleMarked": map[string]interface{}{"foo": "str", "bar": true},
				"nestedassociativeList":                     []interface{}{},
				"patternObject":                             "https://example.com",
				"stringPair":                                []interface{}{"a", "b"},
				"structWithSeveralFields":                   map[string]interface{}{"foo": "str", "bar": true},
				"twoOfAKindPart0":                           "longenough",
				"twoOfAKindPart1":                           "longenough",
				"unprunedEmbeddedResource":                  map[string]interface{}{"kind": "Pod", "apiVersion": "v1"},
				"unprunedFomType":                           map[string]interface{}{},
				"unprunedFomTypeAndField":                   map[string]interface{}{},
				"unprunedJSON":                              map[string]interface{}{"foo": "str", "bar": true},
				"associativeList":                           []interface{}{},
			}
			if opaqueVal != "" {
				spec["opaqueField"] = opaqueVal
			}
			if nonOpaqueVal != "" {
				spec["nonOpaqueField"] = nonOpaqueVal
			}
			obj.Object["spec"] = spec

			return k8sClient.Create(ctx, obj)
		}

		It("should suppress type-level validation for fields with +k8s:opaque", func(ctx SpecContext) {
			// type-level validation is MinLength=4
			// field-level validation is MaxLength=5

			By("allowing opaqueField with length 3 (suppresses type-level MinLength=4)")
			Eventually(func() error {
				return applyCronJob(ctx, "test-opaque-short", "abc", "")
			}, 5*time.Second, 1*time.Second).Should(Succeed())

			By("rejecting nonOpaqueField with length 3 (inherits type-level MinLength=4)")
			err := applyCronJob(ctx, "test-non-opaque-short", "", "abc")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("should be at least 4 chars long"))

			By("rejecting opaqueField with length 6 (applies field-level MaxLength=5)")
			err = applyCronJob(ctx, "test-opaque-long", "abcdef", "")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Too long"))

			By("rejecting nonOpaqueField with length 6 (applies field-level MaxLength=5)")
			err = applyCronJob(ctx, "test-non-opaque-long", "", "abcdef")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Too long"))
		})
	})
})
