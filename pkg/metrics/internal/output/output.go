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

// Package output contains what is shared by the builders of the targets.
package output

import (
	"fmt"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/controller-tools/pkg/genall"
	"sigs.k8s.io/controller-tools/pkg/metrics/internal/model"
	"sigs.k8s.io/controller-tools/pkg/rbac"
)

// ClusterRoleName is the name of the ClusterRole granting access to the custom resources.
const ClusterRoleName = "metrics-role"

// Builder builds the output files for one target.
// The builder gets created before the types get parsed, so the selected target and
// its options are validated early. Afterwards the resources get added one by one
// and finally the files get written.
type Builder interface {
	// AddResource adds the metrics of a custom resource. Resources are added in a
	// deterministic order. It returns an error if the target is not able to
	// represent the resource.
	AddResource(resource model.Resource) error

	// Write writes the output files for all added resources.
	Write(ctx *genall.GenerationContext) error
}

// RBAC collects the RBAC rules needed to read the custom resources of all added resources.
type RBAC struct {
	rules []*rbac.Rule
}

// Add adds the rule to read the resource.
func (r *RBAC) Add(resource model.Resource) {
	r.rules = append(r.rules, &rbac.Rule{
		Groups:    []string{resource.GroupVersionKind.Group},
		Resources: []string{resource.Plural()},
		Verbs:     []string{"get", "list", "watch"},
	})
}

// Write writes a ClusterRole with the collected rules and the given labels to rbac.yaml.
func (r *RBAC) Write(ctx *genall.GenerationContext, labels map[string]string) error {
	clusterRole := rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRole",
			APIVersion: rbacv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   ClusterRoleName,
			Labels: labels,
		},
		Rules: rbac.NormalizeRules(r.rules),
	}

	const virtualFilePath = "rbac.yaml"
	if err := ctx.WriteYAML(virtualFilePath, "", []any{clusterRole}); err != nil {
		return fmt.Errorf("failed to write YAML to %s: %w", virtualFilePath, err)
	}
	return nil
}
