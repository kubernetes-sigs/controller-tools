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

package featuregate_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"sigs.k8s.io/controller-tools/pkg/featuregate"
)

var _ = Describe("FeatureGate Evaluator", func() {
	var evaluator *featuregate.FeatureGateEvaluator
	var gates featuregate.FeatureGateMap

	BeforeEach(func() {
		gates = featuregate.FeatureGateMap{
			"alpha": true,
			"beta":  false,
			"gamma": true,
			"delta": false,
		}
		evaluator = featuregate.NewFeatureGateEvaluator(gates)
	})

	Describe("NewFeatureGateEvaluator", func() {
		It("should create evaluator with provided gates", func() {
			Expect(evaluator).NotTo(BeNil())
		})
	})

	Describe("FeatureGateMap.IsEnabled", func() {
		It("should return true for enabled gates", func() {
			Expect(gates.IsEnabled("alpha")).To(BeTrue())
			Expect(gates.IsEnabled("gamma")).To(BeTrue())
		})

		It("should return false for disabled gates", func() {
			Expect(gates.IsEnabled("beta")).To(BeFalse())
			Expect(gates.IsEnabled("delta")).To(BeFalse())
		})

		It("should return false for unknown gates", func() {
			Expect(gates.IsEnabled("unknown")).To(BeFalse())
		})
	})

	Describe("EvaluateExpression", func() {
		Context("with empty expressions", func() {
			It("should return true for empty string", func() {
				Expect(evaluator.EvaluateExpression("")).To(BeTrue())
			})
		})

		Context("with single gate expressions", func() {
			It("should evaluate enabled gates correctly", func() {
				Expect(evaluator.EvaluateExpression("alpha")).To(BeTrue())
				Expect(evaluator.EvaluateExpression("gamma")).To(BeTrue())
			})

			It("should evaluate disabled gates correctly", func() {
				Expect(evaluator.EvaluateExpression("beta")).To(BeFalse())
				Expect(evaluator.EvaluateExpression("delta")).To(BeFalse())
			})

			It("should evaluate unknown gates as false", func() {
				Expect(evaluator.EvaluateExpression("unknown")).To(BeFalse())
			})
		})

		Context("with OR expressions", func() {
			It("should return true when any gate is enabled", func() {
				Expect(evaluator.EvaluateExpression("alpha|beta")).To(BeTrue())
				Expect(evaluator.EvaluateExpression("beta|gamma")).To(BeTrue())
				Expect(evaluator.EvaluateExpression("alpha|gamma")).To(BeTrue())
			})

			It("should return false when all gates are disabled", func() {
				Expect(evaluator.EvaluateExpression("beta|delta")).To(BeFalse())
			})

			It("should handle multiple OR gates", func() {
				Expect(evaluator.EvaluateExpression("beta|delta|alpha")).To(BeTrue())
				Expect(evaluator.EvaluateExpression("beta|delta|unknown")).To(BeFalse())
			})
		})

		Context("with AND expressions", func() {
			It("should return true when all gates are enabled", func() {
				Expect(evaluator.EvaluateExpression("alpha&gamma")).To(BeTrue())
			})

			It("should return false when any gate is disabled", func() {
				Expect(evaluator.EvaluateExpression("alpha&beta")).To(BeFalse())
				Expect(evaluator.EvaluateExpression("beta&gamma")).To(BeFalse())
				Expect(evaluator.EvaluateExpression("beta&delta")).To(BeFalse())
			})

			It("should handle multiple AND gates", func() {
				Expect(evaluator.EvaluateExpression("alpha&gamma&beta")).To(BeFalse())
				Expect(evaluator.EvaluateExpression("alpha&gamma&unknown")).To(BeFalse())
			})
		})

		Context("with complex expressions", func() {
			It("should handle gates with special characters", func() {
				gatesWithSpecial := featuregate.FeatureGateMap{
					"my-feature":  true,
					"under_score": false,
					"v1beta1":     true,
				}
				specialEvaluator := featuregate.NewFeatureGateEvaluator(gatesWithSpecial)

				Expect(specialEvaluator.EvaluateExpression("my-feature")).To(BeTrue())
				Expect(specialEvaluator.EvaluateExpression("under_score")).To(BeFalse())
				Expect(specialEvaluator.EvaluateExpression("v1beta1")).To(BeTrue())
				Expect(specialEvaluator.EvaluateExpression("my-feature|under_score")).To(BeTrue())
				Expect(specialEvaluator.EvaluateExpression("my-feature&v1beta1")).To(BeTrue())
			})
		})

		Context("with complex precedence rules", func() {
			It("should handle simple parentheses", func() {
				Expect(evaluator.EvaluateExpression("(alpha)")).To(BeTrue())
				Expect(evaluator.EvaluateExpression("(beta)")).To(BeFalse())
			})

			It("should handle parentheses with AND", func() {
				Expect(evaluator.EvaluateExpression("(alpha&gamma)")).To(BeTrue())
				Expect(evaluator.EvaluateExpression("(alpha&beta)")).To(BeFalse())
			})

			It("should handle parentheses with OR", func() {
				Expect(evaluator.EvaluateExpression("(alpha|beta)")).To(BeTrue())
				Expect(evaluator.EvaluateExpression("(beta|delta)")).To(BeFalse())
			})

			It("should handle (AND) OR combinations", func() {
				// (alpha&gamma)|beta = (true&true)|false = true|false = true
				Expect(evaluator.EvaluateExpression("(alpha&gamma)|beta")).To(BeTrue())
				// (alpha&beta)|delta = (true&false)|false = false|false = false
				Expect(evaluator.EvaluateExpression("(alpha&beta)|delta")).To(BeFalse())
				// (beta&delta)|alpha = (false&false)|true = false|true = true
				Expect(evaluator.EvaluateExpression("(beta&delta)|alpha")).To(BeTrue())
			})

			It("should handle (OR) AND combinations", func() {
				// (alpha|beta)&gamma = (true|false)&true = true&true = true
				Expect(evaluator.EvaluateExpression("(alpha|beta)&gamma")).To(BeTrue())
				// (beta|delta)&alpha = (false|false)&true = false&true = false
				Expect(evaluator.EvaluateExpression("(beta|delta)&alpha")).To(BeFalse())
				// (alpha|gamma)&beta = (true|true)&false = true&false = false
				Expect(evaluator.EvaluateExpression("(alpha|gamma)&beta")).To(BeFalse())
			})

			It("should handle nested parentheses", func() {
				// ((alpha&gamma)|beta)&delta = ((true&true)|false)&false = (true|false)&false = true&false = false
				Expect(evaluator.EvaluateExpression("((alpha&gamma)|beta)&delta")).To(BeFalse())
				// ((alpha&gamma)|beta)|delta = ((true&true)|false)|false = (true|false)|false = true|false = true
				Expect(evaluator.EvaluateExpression("((alpha&gamma)|beta)|delta")).To(BeTrue())
			})

			It("should handle multiple grouped expressions", func() {
				// (alpha&gamma)|(beta&delta) = (true&true)|(false&false) = true|false = true
				Expect(evaluator.EvaluateExpression("(alpha&gamma)|(beta&delta)")).To(BeTrue())
				// (alpha&beta)|(delta&unknown) = (true&false)|(false&false) = false|false = false
				Expect(evaluator.EvaluateExpression("(alpha&beta)|(delta&unknown)")).To(BeFalse())
			})

			It("should handle complex expressions with spaces", func() {
				Expect(evaluator.EvaluateExpression("( alpha & gamma ) | beta")).To(BeTrue())
				Expect(evaluator.EvaluateExpression("( alpha | beta ) & gamma")).To(BeTrue())
			})
		})
	})
})
