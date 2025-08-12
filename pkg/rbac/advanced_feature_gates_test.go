package rbac

import (
	"strings"
	"testing"

	"github.com/onsi/gomega"
	"sigs.k8s.io/controller-tools/pkg/genall"
	"sigs.k8s.io/controller-tools/pkg/loader"
	"sigs.k8s.io/controller-tools/pkg/markers"
	rbacv1 "k8s.io/api/rbac/v1"
)

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
		name           string
		featureGates   string
		expectedRules  int
		shouldContain  []string
		shouldNotContain []string
	}{
		{
			name:         "OR logic - alpha enabled",
			featureGates: "alpha=true,beta=false",
			expectedRules: 3, // always-on + OR rule (alpha|beta)
			shouldContain: []string{"pods", "configmaps", "secrets"},
			shouldNotContain: []string{"services"},
		},
		{
			name:         "OR logic - beta enabled",
			featureGates: "alpha=false,beta=true",
			expectedRules: 3, // always-on + OR rule (alpha|beta)
			shouldContain: []string{"pods", "configmaps", "secrets"},
			shouldNotContain: []string{"services"},
		},
		{
			name:         "OR logic - both enabled",
			featureGates: "alpha=true,beta=true",
			expectedRules: 4, // always-on + OR rule + AND rule
			shouldContain: []string{"pods", "configmaps", "secrets", "services"},
			shouldNotContain: []string{},
		},
		{
			name:         "OR logic - neither enabled",
			featureGates: "alpha=false,beta=false",
			expectedRules: 2, // only always-on
			shouldContain: []string{"pods", "configmaps"},
			shouldNotContain: []string{"secrets", "services"},
		},
		{
			name:         "AND logic - only alpha enabled",
			featureGates: "alpha=true,beta=false",
			expectedRules: 3, // always-on + OR rule (alpha|beta)
			shouldContain: []string{"pods", "configmaps", "secrets"},
			shouldNotContain: []string{"services"},
		},
		{
			name:         "AND logic - both enabled",
			featureGates: "alpha=true,beta=true",
			expectedRules: 4, // always-on + OR rule + AND rule
			shouldContain: []string{"pods", "configmaps", "secrets", "services"},
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
		{name: "mixed operators", expression: "alpha&beta|gamma", shouldError: true},
		{name: "invalid character", expression: "alpha@beta", shouldError: true},
		{name: "hyphenated gate", expression: "feature-alpha", shouldError: false},
		{name: "underscore gate", expression: "feature_alpha", shouldError: false},
		{name: "numeric gate", expression: "v1beta1", shouldError: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFeatureGateExpression(tt.expression)
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
