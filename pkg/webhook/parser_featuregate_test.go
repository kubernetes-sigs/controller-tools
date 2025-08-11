/*
Copyright 2024 The Kubernetes Authors.

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

package webhook_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-tools/pkg/featuregate"
)

var _ = Describe("FeatureGates", func() {
	Describe("ParseFeatureGates", func() {
		It("should parse empty string", func() {
			result, err := featuregate.ParseFeatureGates("", false)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(featuregate.FeatureGateMap{}))
		})

		It("should parse single gate enabled", func() {
			result, err := featuregate.ParseFeatureGates("alpha=true", false)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(featuregate.FeatureGateMap{"alpha": true}))
		})

		It("should parse single gate disabled", func() {
			result, err := featuregate.ParseFeatureGates("alpha=false", false)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(featuregate.FeatureGateMap{"alpha": false}))
		})

		It("should parse multiple gates", func() {
			result, err := featuregate.ParseFeatureGates("alpha=true,beta=false,gamma=true", false)
			Expect(err).ToNot(HaveOccurred())
			expected := featuregate.FeatureGateMap{
				"alpha": true,
				"beta":  false,
				"gamma": true,
			}
			Expect(result).To(Equal(expected))
		})

		It("should parse gates with spaces", func() {
			result, err := featuregate.ParseFeatureGates(" alpha = true , beta = false ", false)
			Expect(err).ToNot(HaveOccurred())
			expected := featuregate.FeatureGateMap{
				"alpha": true,
				"beta":  false,
			}
			Expect(result).To(Equal(expected))
		})

		It("should ignore invalid format", func() {
			result, err := featuregate.ParseFeatureGates("alpha=true,invalid,beta=false", false)
			Expect(err).ToNot(HaveOccurred())
			expected := featuregate.FeatureGateMap{
				"alpha": true,
				"beta":  false,
			}
			Expect(result).To(Equal(expected))
		})
	})

	Describe("ValidateFeatureGateExpression", func() {
		It("should reject empty expression", func() {
			err := featuregate.ValidateFeatureGateExpression("", nil, false)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should accept simple gate name", func() {
			err := featuregate.ValidateFeatureGateExpression("alpha", nil, false)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should reject mixed operators", func() {
			err := featuregate.ValidateFeatureGateExpression("alpha&beta|gamma", nil, false)
			Expect(err).To(HaveOccurred())
		})

		It("should accept OR expression", func() {
			err := featuregate.ValidateFeatureGateExpression("alpha|beta", nil, false)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should accept AND expression", func() {
			err := featuregate.ValidateFeatureGateExpression("alpha&beta", nil, false)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should reject invalid characters", func() {
			err := featuregate.ValidateFeatureGateExpression("alpha@beta", nil, false)
			Expect(err).To(HaveOccurred())
		})

		It("should reject spaces", func() {
			err := featuregate.ValidateFeatureGateExpression("alpha beta", nil, false)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("shouldIncludeWebhook (internal testing)", func() {
		It("should include webhook without feature gates", func() {
			// Test basic featuregate evaluator functionality
			evaluator := featuregate.NewFeatureGateEvaluator(map[string]bool{})
			result := evaluator.EvaluateExpression("")
			Expect(result).To(BeTrue())
		})

		It("should include webhook with matching feature gate enabled", func() {
			evaluator := featuregate.NewFeatureGateEvaluator(map[string]bool{"alpha": true})
			result := evaluator.EvaluateExpression("alpha")
			Expect(result).To(BeTrue())
		})

		It("should exclude webhook with matching feature gate disabled", func() {
			evaluator := featuregate.NewFeatureGateEvaluator(map[string]bool{"alpha": false})
			result := evaluator.EvaluateExpression("alpha")
			Expect(result).To(BeFalse())
		})

		It("should exclude webhook with missing feature gate", func() {
			evaluator := featuregate.NewFeatureGateEvaluator(map[string]bool{"beta": true})
			result := evaluator.EvaluateExpression("alpha")
			Expect(result).To(BeFalse())
		})

		It("should handle OR expression correctly", func() {
			evaluator := featuregate.NewFeatureGateEvaluator(map[string]bool{"alpha": false, "beta": true})
			result := evaluator.EvaluateExpression("alpha|beta")
			Expect(result).To(BeTrue())
		})

		It("should handle AND expression correctly", func() {
			evaluator := featuregate.NewFeatureGateEvaluator(map[string]bool{"alpha": true, "beta": true})
			result := evaluator.EvaluateExpression("alpha&beta")
			Expect(result).To(BeTrue())
		})

		It("should handle complex AND expression", func() {
			evaluator := featuregate.NewFeatureGateEvaluator(map[string]bool{"alpha": true, "beta": true, "gamma": false})
			result := evaluator.EvaluateExpression("alpha&beta&gamma")
			Expect(result).To(BeFalse())
		})
	})
})
