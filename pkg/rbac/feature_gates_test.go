package rbac

import (
	"strings"
	"testing"

	"github.com/onsi/gomega"
	rbacv1 "k8s.io/api/rbac/v1"
	"sigs.k8s.io/controller-tools/pkg/featuregate"
	"sigs.k8s.io/controller-tools/pkg/genall"
	"sigs.k8s.io/controller-tools/pkg/loader"
	"sigs.k8s.io/controller-tools/pkg/markers"
)

func TestFeatureGates(t *testing.T) {
	g := gomega.NewWithT(t)

	// Load test packages
	pkgs, err := loader.LoadRoots("./testdata/feature_gates")
	g.Expect(err).NotTo(gomega.HaveOccurred())

	// Set up generation context
	reg := &markers.Registry{}
	g.Expect(reg.Register(RuleDefinition)).To(gomega.Succeed())

	ctx := &genall.GenerationContext{
		Collector: &markers.Collector{Registry: reg},
		Roots:     pkgs,
	}

	tests := []struct {
		name             string
		featureGates     string
		expectedRules    int
		shouldContain    []string
		shouldNotContain []string
	}{
		{
			name:             "no feature gates",
			featureGates:     "",
			expectedRules:    2, // only always-on rules
			shouldContain:    []string{"pods", "configmaps"},
			shouldNotContain: []string{"deployments", "ingresses"},
		},
		{
			name:             "alpha enabled",
			featureGates:     "alpha=true",
			expectedRules:    3, // always-on + alpha
			shouldContain:    []string{"pods", "configmaps", "deployments"},
			shouldNotContain: []string{"ingresses"},
		},
		{
			name:             "beta enabled",
			featureGates:     "beta=true",
			expectedRules:    3, // always-on + beta
			shouldContain:    []string{"pods", "configmaps", "ingresses"},
			shouldNotContain: []string{"deployments"},
		},
		{
			name:             "both enabled",
			featureGates:     "alpha=true,beta=true",
			expectedRules:    4, // all rules
			shouldContain:    []string{"pods", "configmaps", "deployments", "ingresses"},
			shouldNotContain: []string{},
		},
		{
			name:             "alpha enabled beta disabled",
			featureGates:     "alpha=true,beta=false",
			expectedRules:    3, // always-on + alpha
			shouldContain:    []string{"pods", "configmaps", "deployments"},
			shouldNotContain: []string{"ingresses"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewWithT(t)

			objs, err := GenerateRoles(ctx, "test-role", tt.featureGates)
			g.Expect(err).NotTo(gomega.HaveOccurred())
			g.Expect(objs).To(gomega.HaveLen(1))

			role, ok := objs[0].(rbacv1.ClusterRole)
			g.Expect(ok).To(gomega.BeTrue())
			g.Expect(role.Rules).To(gomega.HaveLen(tt.expectedRules))

			// Convert rules to string for easier checking
			rulesStr := ""
			for _, rule := range role.Rules {
				rulesStr += strings.Join(rule.Resources, ",") + " "
			}

			for _, resource := range tt.shouldContain {
				g.Expect(rulesStr).To(gomega.ContainSubstring(resource),
					"Expected resource %s to be present", resource)
			}

			for _, resource := range tt.shouldNotContain {
				g.Expect(rulesStr).NotTo(gomega.ContainSubstring(resource),
					"Expected resource %s to be absent", resource)
			}
		})
	}
}

func TestAdvancedFeatureGates(t *testing.T) {
	g := gomega.NewWithT(t)

	// Load test packages
	pkgs, err := loader.LoadRoots("./testdata/advanced_feature_gates")
	g.Expect(err).NotTo(gomega.HaveOccurred())

	// Set up generation context
	reg := &markers.Registry{}
	g.Expect(reg.Register(RuleDefinition)).To(gomega.Succeed())

	ctx := &genall.GenerationContext{
		Collector: &markers.Collector{Registry: reg},
		Roots:     pkgs,
	}

	tests := []struct {
		name             string
		featureGates     string
		expectedRules    int
		shouldContain    []string
		shouldNotContain []string
	}{
		{
			name:             "OR logic - alpha enabled",
			featureGates:     "alpha=true,beta=false,gamma=false",
			expectedRules:    3, // always-on + OR rule (alpha|beta)
			shouldContain:    []string{"pods", "configmaps", "secrets"},
			shouldNotContain: []string{"services", "jobs", "replicasets"},
		},
		{
			name:             "OR logic - beta enabled",
			featureGates:     "alpha=false,beta=true,gamma=false",
			expectedRules:    3, // always-on + OR rule (alpha|beta)
			shouldContain:    []string{"pods", "configmaps", "secrets"},
			shouldNotContain: []string{"services", "jobs", "replicasets"},
		},
		{
			name:             "AND logic - both alpha and beta enabled",
			featureGates:     "alpha=true,beta=true,gamma=false",
			expectedRules:    5, // always-on + OR rule + AND rule + complex OR (alpha&beta)|gamma
			shouldContain:    []string{"pods", "configmaps", "secrets", "services", "jobs"},
			shouldNotContain: []string{"replicasets"},
		},
		{
			name:             "OR logic - neither enabled",
			featureGates:     "alpha=false,beta=false,gamma=false",
			expectedRules:    2, // only always-on
			shouldContain:    []string{"pods", "configmaps"},
			shouldNotContain: []string{"secrets", "services", "jobs", "replicasets"},
		},
		{
			name:             "Complex precedence - gamma enabled",
			featureGates:     "alpha=false,beta=false,gamma=true",
			expectedRules:    3, // always-on + complex OR (alpha&beta)|gamma (satisfied by gamma)
			shouldContain:    []string{"pods", "configmaps", "jobs"},
			shouldNotContain: []string{"secrets", "services", "replicasets"},
		},
		{
			name:             "Complex precedence - alpha and gamma enabled",
			featureGates:     "alpha=true,beta=false,gamma=true",
			expectedRules:    5, // always-on + OR + complex OR + complex AND
			shouldContain:    []string{"pods", "configmaps", "secrets", "jobs", "replicasets"},
			shouldNotContain: []string{"services"},
		},
		{
			name:             "Complex precedence - all enabled",
			featureGates:     "alpha=true,beta=true,gamma=true",
			expectedRules:    6, // all rules enabled
			shouldContain:    []string{"pods", "configmaps", "secrets", "services", "jobs", "replicasets"},
			shouldNotContain: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewWithT(t)

			objs, err := GenerateRoles(ctx, "test-role", tt.featureGates)
			g.Expect(err).NotTo(gomega.HaveOccurred())
			g.Expect(objs).To(gomega.HaveLen(1))

			role, ok := objs[0].(rbacv1.ClusterRole)
			g.Expect(ok).To(gomega.BeTrue())
			g.Expect(role.Rules).To(gomega.HaveLen(tt.expectedRules))

			// Convert rules to string for easier checking
			rulesStr := ""
			for _, rule := range role.Rules {
				rulesStr += strings.Join(rule.Resources, ",") + " "
			}

			for _, resource := range tt.shouldContain {
				g.Expect(rulesStr).To(gomega.ContainSubstring(resource),
					"Expected resource %s to be present", resource)
			}

			for _, resource := range tt.shouldNotContain {
				g.Expect(rulesStr).NotTo(gomega.ContainSubstring(resource),
					"Expected resource %s to be absent", resource)
			}
		})
	}
}

func TestFeatureGateValidation(t *testing.T) {
	tests := []struct {
		name        string
		expression  string
		shouldError bool
	}{
		{name: "empty expression", expression: "", shouldError: false},
		{name: "single gate", expression: "alpha", shouldError: false},
		{name: "OR expression", expression: "alpha|beta", shouldError: false},
		{name: "AND expression", expression: "alpha&beta", shouldError: false},
		{name: "mixed operators", expression: "alpha&beta|gamma", shouldError: false}, // AND has higher precedence than OR
		{name: "invalid character", expression: "alpha@beta", shouldError: true},
		{name: "hyphenated gate", expression: "feature-alpha", shouldError: false},
		{name: "underscore gate", expression: "feature_alpha", shouldError: false},
		{name: "numeric gate", expression: "v1beta1", shouldError: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := featuregate.ValidateFeatureGateExpression(tt.expression, nil, false)
			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error for expression %s, but got none", tt.expression)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for expression %s, but got: %v", tt.expression, err)
				}
			}
		})
	}
}
