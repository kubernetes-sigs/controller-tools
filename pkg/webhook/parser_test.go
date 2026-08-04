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

package webhook

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	admissionregv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func patchTestConfig(mutating bool, path string) Config {
	return Config{
		Mutating:                mutating,
		Name:                    "test.webhook.io",
		FailurePolicy:           "fail",
		SideEffects:             "None",
		Path:                    path,
		Groups:                  []string{"apps"},
		Resources:               []string{"deployments"},
		Versions:                []string{"v1"},
		Verbs:                   []string{"create", "update"},
		AdmissionReviewVersions: []string{"v1"},
	}
}

func expectSingleRuleWithOnlyAPIGroup(rules []admissionregv1.RuleWithOperations, group string) {
	GinkgoHelper()
	Expect(rules).To(HaveLen(1))
	Expect(rules[0].APIGroups).To(ConsistOf(group))
	Expect(rules[0].APIVersions).To(BeNil())
	Expect(rules[0].Resources).To(BeNil())
	Expect(rules[0].Operations).To(BeNil())
}

var _ = Describe("applyPatch", func() {
	var webhook *admissionregv1.MutatingWebhook

	BeforeEach(func() {
		webhook = &admissionregv1.MutatingWebhook{Name: "test-webhook"}
	})

	It("should leave the webhook unchanged for an empty patch", func() {
		By("applying an empty patch")
		Expect(applyPatch(webhook, "")).To(Succeed())

		By("verifying the webhook is unchanged")
		Expect(webhook.Name).To(Equal("test-webhook"))
		Expect(webhook.NamespaceSelector).To(BeNil())
	})

	It("should set namespaceSelector with matchLabels", func() {
		By("applying a namespaceSelector patch")
		Expect(applyPatch(webhook, `{"namespaceSelector":{"matchLabels":{"webhook-enabled":"true"}}}`)).To(Succeed())

		By("verifying the namespaceSelector")
		Expect(webhook.NamespaceSelector).NotTo(BeNil())
		Expect(webhook.NamespaceSelector.MatchLabels).To(Equal(map[string]string{"webhook-enabled": "true"}))
	})

	It("should set objectSelector with matchExpressions", func() {
		By("applying an objectSelector patch")
		Expect(applyPatch(webhook, `{"objectSelector":{"matchExpressions":[{"key":"tier","operator":"In","values":["frontend","backend"]}]}}`)).To(Succeed())

		By("verifying the objectSelector")
		Expect(webhook.ObjectSelector).NotTo(BeNil())
		Expect(webhook.ObjectSelector.MatchExpressions).To(ConsistOf(metav1.LabelSelectorRequirement{
			Key:      "tier",
			Operator: metav1.LabelSelectorOpIn,
			Values:   []string{"frontend", "backend"},
		}))
	})

	It("should apply multiple fields including timeoutSeconds", func() {
		By("applying a patch with a selector and timeoutSeconds")
		Expect(applyPatch(webhook, `{"namespaceSelector":{"matchLabels":{"env":"prod"}},"timeoutSeconds":15}`)).To(Succeed())

		By("verifying both fields")
		Expect(webhook.NamespaceSelector).NotTo(BeNil())
		Expect(webhook.NamespaceSelector.MatchLabels).To(HaveKeyWithValue("env", "prod"))
		Expect(webhook.TimeoutSeconds).To(HaveValue(Equal(int32(15))))
	})

	It("should set both selectors with matchLabels and matchExpressions", func() {
		By("applying a patch with both selectors")
		Expect(applyPatch(webhook, `{"namespaceSelector":{"matchLabels":{"environment":"staging"}},"objectSelector":{"matchLabels":{"managed-by":"my-operator"},"matchExpressions":[{"key":"app","operator":"NotIn","values":["legacy-app"]}]}}`)).To(Succeed())

		By("verifying both selectors")
		Expect(webhook.NamespaceSelector).NotTo(BeNil())
		Expect(webhook.NamespaceSelector.MatchLabels).To(HaveKeyWithValue("environment", "staging"))
		Expect(webhook.ObjectSelector).NotTo(BeNil())
		Expect(webhook.ObjectSelector.MatchLabels).To(HaveKeyWithValue("managed-by", "my-operator"))
		Expect(webhook.ObjectSelector.MatchExpressions).To(HaveLen(1))
	})

	It("should error on invalid JSON", func() {
		Expect(applyPatch(webhook, `{invalid`)).NotTo(Succeed())
	})

	It("should handle keys with special characters", func() {
		By("applying a patch with a label key containing dots and slashes")
		Expect(applyPatch(webhook, `{"namespaceSelector":{"matchLabels":{"app.kubernetes.io/name":"myapp"}}}`)).To(Succeed())

		By("verifying the label key")
		Expect(webhook.NamespaceSelector).NotTo(BeNil())
		Expect(webhook.NamespaceSelector.MatchLabels).To(HaveKeyWithValue("app.kubernetes.io/name", "myapp"))
	})

	Context("strategic merge semantics", func() {
		It("should replace the whole rules list", func() {
			By("populating a rule from marker fields")
			webhook.Rules = []admissionregv1.RuleWithOperations{{
				Operations: []admissionregv1.OperationType{admissionregv1.Create, admissionregv1.Update},
				Rule: admissionregv1.Rule{
					APIGroups:   []string{"apps"},
					APIVersions: []string{"v1"},
					Resources:   []string{"widgets"},
				},
			}}

			By("applying a patch with a partial rule")
			Expect(applyPatch(webhook, `{"rules":[{"apiGroups":["batch"]}]}`)).To(Succeed())

			By("verifying the whole list was replaced")
			expectSingleRuleWithOnlyAPIGroup(webhook.Rules, "batch")
		})

		It("should remove timeoutSeconds set to null", func() {
			By("populating timeoutSeconds")
			webhook.TimeoutSeconds = new(int32(10))

			By("applying a null patch")
			Expect(applyPatch(webhook, `{"timeoutSeconds":null}`)).To(Succeed())

			By("verifying the field was removed")
			Expect(webhook.TimeoutSeconds).To(BeNil())
		})

		It("should remove namespaceSelector set to null", func() {
			By("populating the namespaceSelector")
			webhook.NamespaceSelector = &metav1.LabelSelector{
				MatchLabels: map[string]string{"env": "prod"},
			}

			By("applying a null patch")
			Expect(applyPatch(webhook, `{"namespaceSelector":null}`)).To(Succeed())

			By("verifying the selector was removed")
			Expect(webhook.NamespaceSelector).To(BeNil())
		})

		It("should treat a valid empty JSON object as a no-op", func() {
			By("populating timeoutSeconds")
			webhook.TimeoutSeconds = new(int32(10))

			By("applying an empty JSON object")
			Expect(applyPatch(webhook, `{}`)).To(Succeed())

			By("verifying the webhook is unchanged")
			Expect(webhook.Name).To(Equal("test-webhook"))
			Expect(webhook.TimeoutSeconds).To(HaveValue(Equal(int32(10))))
		})

		It("should error on a whitespace-only patch", func() {
			By("applying a whitespace-only patch")
			Expect(applyPatch(webhook, " \n\t ")).NotTo(Succeed())
		})

		It("should treat null on an unset field as a no-op", func() {
			By("applying a null patch to an unset field")
			Expect(applyPatch(webhook, `{"timeoutSeconds":null}`)).To(Succeed())

			By("verifying the field stays unset")
			Expect(webhook.TimeoutSeconds).To(BeNil())
		})

		It("should error on unknown fields", func() {
			Expect(applyPatch(webhook, `{"notAField":"value","timeoutSeconds":5}`)).NotTo(Succeed())
		})

		It("should error on unknown fields set to null", func() {
			Expect(applyPatch(webhook, `{"notAField":null}`)).NotTo(Succeed())
		})

		It("should error on unknown fields with self-removing directives", func() {
			Expect(applyPatch(webhook, `{"notAField":{"$patch":"delete"}}`)).NotTo(Succeed())
		})

		It("should error on miscased fields", func() {
			Expect(applyPatch(webhook, `{"TimeoutSeconds":25}`)).NotTo(Succeed())
		})

		It("should error on miscased fields set to null", func() {
			Expect(applyPatch(webhook, `{"TimeoutSeconds":null}`)).NotTo(Succeed())
		})

		It("should error on miscased nested fields", func() {
			Expect(applyPatch(webhook, `{"namespaceSelector":{"matchlabels":{"env":"prod"}}}`)).NotTo(Succeed())
		})

		It("should error on misspelt patch directives", func() {
			Expect(applyPatch(webhook, `{"$patc":"replace"}`)).NotTo(Succeed())
		})

		It("should error on unknown patch directives", func() {
			Expect(applyPatch(webhook, `{"$unknown":"value"}`)).NotTo(Succeed())
		})

		It("should error on misspelt directives in list elements", func() {
			Expect(applyPatch(webhook, `{"matchConditions":[{"name":"exclude-kube-system","$patc":"delete"}]}`)).NotTo(Succeed())
		})

		It("should error on directive suffixes that are not fields", func() {
			Expect(applyPatch(webhook, `{"$setElementOrder/notAField":[{"name":"exclude-kube-system"}]}`)).NotTo(Succeed())
		})

		It("should delete from a primitive list via $deleteFromPrimitiveList", func() {
			By("populating admissionReviewVersions")
			webhook.AdmissionReviewVersions = []string{"v1", "v1beta1"}

			By("applying a $deleteFromPrimitiveList directive")
			Expect(applyPatch(webhook, `{"$deleteFromPrimitiveList/admissionReviewVersions":["v1beta1"]}`)).To(Succeed())

			By("verifying the entry was removed")
			Expect(webhook.AdmissionReviewVersions).To(Equal([]string{"v1"}))
		})

		It("should reorder merge-keyed entries via $setElementOrder", func() {
			By("populating matchConditions")
			webhook.MatchConditions = []admissionregv1.MatchCondition{
				{Name: "exclude-kube-system", Expression: `object.metadata.namespace != "kube-system"`},
				{Name: "exclude-leases", Expression: `request.resource.resource != "leases"`},
			}

			By("applying a $setElementOrder directive")
			Expect(applyPatch(webhook, `{"$setElementOrder/matchConditions":[{"name":"exclude-leases"},{"name":"exclude-kube-system"}]}`)).To(Succeed())

			By("verifying the conditions were reordered")
			Expect(webhook.MatchConditions).To(Equal([]admissionregv1.MatchCondition{
				{Name: "exclude-leases", Expression: `request.resource.resource != "leases"`},
				{Name: "exclude-kube-system", Expression: `object.metadata.namespace != "kube-system"`},
			}))
		})

		It("should merge matchConditions by name", func() {
			By("populating matchConditions")
			webhook.MatchConditions = []admissionregv1.MatchCondition{
				{Name: "updated", Expression: "old-expression"},
				{Name: "kept", Expression: "kept-expression"},
			}

			By("applying a patch that updates one condition and adds another")
			Expect(applyPatch(webhook, `{"matchConditions":[{"name":"updated","expression":"new-expression"},{"name":"added","expression":"added-expression"}]}`)).To(Succeed())

			By("verifying the conditions were merged by name")
			Expect(webhook.MatchConditions).To(ConsistOf(
				admissionregv1.MatchCondition{Name: "updated", Expression: "new-expression"},
				admissionregv1.MatchCondition{Name: "kept", Expression: "kept-expression"},
				admissionregv1.MatchCondition{Name: "added", Expression: "added-expression"},
			))
		})

		It("should delete a merge-keyed entry via $patch delete", func() {
			By("populating matchConditions")
			webhook.MatchConditions = []admissionregv1.MatchCondition{
				{Name: "removed", Expression: "removed-expression"},
				{Name: "kept", Expression: "kept-expression"},
			}

			By("applying a $patch delete for one condition")
			Expect(applyPatch(webhook, `{"matchConditions":[{"name":"removed","$patch":"delete"}]}`)).To(Succeed())

			By("verifying only the other condition remains")
			Expect(webhook.MatchConditions).To(ConsistOf(
				admissionregv1.MatchCondition{Name: "kept", Expression: "kept-expression"},
			))
		})
	})
})

var _ = Describe("Config.ToMutatingWebhook with a patch", func() {
	var config Config

	BeforeEach(func() {
		config = patchTestConfig(true, "/mutate")
	})

	It("should apply a namespaceSelector patch", func() {
		By("generating the webhook with a namespaceSelector patch")
		config.Patch = `{"namespaceSelector":{"matchLabels":{"webhook":"enabled"}}}`
		webhook, err := config.ToMutatingWebhook()
		Expect(err).NotTo(HaveOccurred())

		By("verifying the namespaceSelector")
		Expect(webhook.NamespaceSelector).NotTo(BeNil())
		Expect(webhook.NamespaceSelector.MatchLabels).To(HaveKeyWithValue("webhook", "enabled"))
	})

	It("should apply an objectSelector patch", func() {
		By("generating the webhook with an objectSelector patch")
		config.Patch = `{"objectSelector":{"matchExpressions":[{"key":"managed-by","operator":"In","values":["controller"]}]}}`
		webhook, err := config.ToMutatingWebhook()
		Expect(err).NotTo(HaveOccurred())

		By("verifying the objectSelector")
		Expect(webhook.ObjectSelector).NotTo(BeNil())
		Expect(webhook.ObjectSelector.MatchExpressions).To(HaveLen(1))
	})

	It("should override the marker timeoutSeconds", func() {
		By("generating the webhook with a timeoutSeconds override patch")
		config.TimeoutSeconds = 10
		config.Patch = `{"timeoutSeconds":25}`
		webhook, err := config.ToMutatingWebhook()
		Expect(err).NotTo(HaveOccurred())

		By("verifying the timeout")
		Expect(webhook.TimeoutSeconds).To(HaveValue(Equal(int32(25))))
	})

	It("should error on invalid patch JSON", func() {
		config.Patch = `{invalid json`
		_, err := config.ToMutatingWebhook()
		Expect(err).To(HaveOccurred())
	})
})

var _ = Describe("Config.ToValidatingWebhook with a patch", func() {
	var config Config

	BeforeEach(func() {
		config = patchTestConfig(false, "/validate")
	})

	It("should apply a namespaceSelector patch", func() {
		By("generating the webhook with a namespaceSelector patch")
		config.Patch = `{"namespaceSelector":{"matchLabels":{"env":"production"}}}`
		webhook, err := config.ToValidatingWebhook()
		Expect(err).NotTo(HaveOccurred())

		By("verifying the namespaceSelector")
		Expect(webhook.NamespaceSelector).NotTo(BeNil())
		Expect(webhook.NamespaceSelector.MatchLabels).To(HaveKeyWithValue("env", "production"))
	})

	It("should remove timeoutSeconds and replace rules in one patch", func() {
		By("generating the webhook with a removal and replacement patch")
		config.TimeoutSeconds = 10
		config.Patch = `{"timeoutSeconds":null,"rules":[{"apiGroups":["batch"]}]}`
		webhook, err := config.ToValidatingWebhook()
		Expect(err).NotTo(HaveOccurred())

		By("verifying the timeout was removed")
		Expect(webhook.TimeoutSeconds).To(BeNil())

		By("verifying the rules were replaced")
		expectSingleRuleWithOnlyAPIGroup(webhook.Rules, "batch")
	})
})
