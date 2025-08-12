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
		name           string
		featureGates   string
		expectedRules  int
		shouldContain  []string
		shouldNotContain []string
	}{
		{
			name:         "no feature gates",
			featureGates: "",
			expectedRules: 2, // only always-on rules
			shouldContain: []string{"pods", "configmaps"},
			shouldNotContain: []string{"deployments", "ingresses"},
		},
		{
			name:         "alpha enabled",
			featureGates: "alpha=true",
			expectedRules: 3, // always-on + alpha
			shouldContain: []string{"pods", "configmaps", "deployments"},
			shouldNotContain: []string{"ingresses"},
		},
		{
			name:         "beta enabled", 
			featureGates: "beta=true",
			expectedRules: 3, // always-on + beta
			shouldContain: []string{"pods", "configmaps", "ingresses"},
			shouldNotContain: []string{"deployments"},
		},
		{
			name:         "both enabled",
			featureGates: "alpha=true,beta=true",
			expectedRules: 4, // all rules
			shouldContain: []string{"pods", "configmaps", "deployments", "ingresses"},
			shouldNotContain: []string{},
		},
		{
			name:         "alpha enabled beta disabled",
			featureGates: "alpha=true,beta=false",
			expectedRules: 3, // always-on + alpha
			shouldContain: []string{"pods", "configmaps", "deployments"},
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
