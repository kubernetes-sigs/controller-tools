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
	"testing"

	admissionregv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestApplyPatch verifies that strategic merge patches are correctly applied to webhook configurations.
func TestApplyPatch(t *testing.T) {
	tests := []struct {
		name        string
		patch       string
		expectError bool
		validate    func(*testing.T, *admissionregv1.MutatingWebhook)
	}{
		{
			name:  "empty patch",
			patch: "",
			validate: func(t *testing.T, wh *admissionregv1.MutatingWebhook) {
				if wh.NamespaceSelector != nil {
					t.Error("expected nil namespaceSelector")
				}
			},
		},
		{
			name:  "namespaceSelector with matchLabels",
			patch: `{"namespaceSelector":{"matchLabels":{"webhook-enabled":"true"}}}`,
			validate: func(t *testing.T, wh *admissionregv1.MutatingWebhook) {
				if wh.NamespaceSelector == nil {
					t.Fatal("expected namespaceSelector to be set")
				}
				if len(wh.NamespaceSelector.MatchLabels) != 1 {
					t.Errorf("expected 1 matchLabel, got %d", len(wh.NamespaceSelector.MatchLabels))
				}
				if wh.NamespaceSelector.MatchLabels["webhook-enabled"] != "true" {
					t.Errorf("expected webhook-enabled=true, got %q", wh.NamespaceSelector.MatchLabels["webhook-enabled"])
				}
			},
		},
		{
			name:  "objectSelector with matchExpressions",
			patch: `{"objectSelector":{"matchExpressions":[{"key":"tier","operator":"In","values":["frontend","backend"]}]}}`,
			validate: func(t *testing.T, wh *admissionregv1.MutatingWebhook) {
				if wh.ObjectSelector == nil {
					t.Fatal("expected objectSelector to be set")
				}
				if len(wh.ObjectSelector.MatchExpressions) != 1 {
					t.Fatalf("expected 1 matchExpression, got %d", len(wh.ObjectSelector.MatchExpressions))
				}
				expr := wh.ObjectSelector.MatchExpressions[0]
				if expr.Key != "tier" {
					t.Errorf("expected key=tier, got %q", expr.Key)
				}
				if expr.Operator != metav1.LabelSelectorOpIn {
					t.Errorf("expected operator=In, got %q", expr.Operator)
				}
				if len(expr.Values) != 2 {
					t.Errorf("expected 2 values, got %d", len(expr.Values))
				}
			},
		},
		{
			name:  "multiple fields including timeoutSeconds",
			patch: `{"namespaceSelector":{"matchLabels":{"env":"prod"}},"timeoutSeconds":15}`,
			validate: func(t *testing.T, wh *admissionregv1.MutatingWebhook) {
				if wh.NamespaceSelector == nil {
					t.Fatal("expected namespaceSelector to be set")
				}
				if wh.NamespaceSelector.MatchLabels["env"] != "prod" {
					t.Error("expected env=prod")
				}
				if wh.TimeoutSeconds == nil || *wh.TimeoutSeconds != 15 {
					t.Errorf("expected timeoutSeconds=15, got %v", wh.TimeoutSeconds)
				}
			},
		},
		{
			name:  "both selectors with matchLabels and matchExpressions",
			patch: `{"namespaceSelector":{"matchLabels":{"ns-label":"value"}},"objectSelector":{"matchLabels":{"obj-label":"value"},"matchExpressions":[{"key":"app","operator":"NotIn","values":["excluded"]}]}}`,
			validate: func(t *testing.T, wh *admissionregv1.MutatingWebhook) {
				if wh.NamespaceSelector == nil {
					t.Fatal("expected namespaceSelector to be set")
				}
				if wh.ObjectSelector == nil {
					t.Fatal("expected objectSelector to be set")
				}
				if wh.NamespaceSelector.MatchLabels["ns-label"] != "value" {
					t.Error("expected ns-label=value")
				}
				if wh.ObjectSelector.MatchLabels["obj-label"] != "value" {
					t.Error("expected obj-label=value")
				}
				if len(wh.ObjectSelector.MatchExpressions) != 1 {
					t.Errorf("expected 1 matchExpression, got %d", len(wh.ObjectSelector.MatchExpressions))
				}
			},
		},
		{
			name:        "invalid JSON",
			patch:       `{invalid`,
			expectError: true,
		},
		{
			name:  "patch with special characters in keys",
			patch: `{"namespaceSelector":{"matchLabels":{"app.kubernetes.io/name":"myapp"}}}`,
			validate: func(t *testing.T, wh *admissionregv1.MutatingWebhook) {
				if wh.NamespaceSelector == nil {
					t.Fatal("expected namespaceSelector to be set")
				}
				if wh.NamespaceSelector.MatchLabels["app.kubernetes.io/name"] != "myapp" {
					t.Error("expected app.kubernetes.io/name=myapp")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a basic webhook
			webhook := &admissionregv1.MutatingWebhook{
				Name: "test-webhook",
			}

			err := applyPatch(webhook, tt.patch)
			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.validate != nil {
				tt.validate(t, webhook)
			}
		})
	}
}

// TestConfigToMutatingWebhookWithPatch verifies that the Config.ToMutatingWebhook method
// correctly applies patches to the generated webhook.
func TestConfigToMutatingWebhookWithPatch(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		expectError bool
		validate    func(*testing.T, admissionregv1.MutatingWebhook)
	}{
		{
			name: "mutating webhook with namespaceSelector patch",
			config: Config{
				Mutating:                true,
				Name:                    "test.webhook.io",
				FailurePolicy:           "fail",
				SideEffects:             "None",
				Path:                    "/mutate",
				Groups:                  []string{"apps"},
				Resources:               []string{"deployments"},
				Versions:                []string{"v1"},
				Verbs:                   []string{"create", "update"},
				AdmissionReviewVersions: []string{"v1"},
				Patch:                   `{"namespaceSelector":{"matchLabels":{"webhook":"enabled"}}}`,
			},
			validate: func(t *testing.T, wh admissionregv1.MutatingWebhook) {
				if wh.NamespaceSelector == nil {
					t.Fatal("expected namespaceSelector to be set")
				}
				if wh.NamespaceSelector.MatchLabels["webhook"] != "enabled" {
					t.Error("expected webhook=enabled")
				}
			},
		},
		{
			name: "mutating webhook with objectSelector patch",
			config: Config{
				Mutating:                true,
				Name:                    "test.webhook.io",
				FailurePolicy:           "fail",
				SideEffects:             "None",
				Path:                    "/mutate",
				Groups:                  []string{"apps"},
				Resources:               []string{"deployments"},
				Versions:                []string{"v1"},
				Verbs:                   []string{"create"},
				AdmissionReviewVersions: []string{"v1"},
				Patch:                   `{"objectSelector":{"matchExpressions":[{"key":"managed-by","operator":"In","values":["controller"]}]}}`,
			},
			validate: func(t *testing.T, wh admissionregv1.MutatingWebhook) {
				if wh.ObjectSelector == nil {
					t.Fatal("expected objectSelector to be set")
				}
				if len(wh.ObjectSelector.MatchExpressions) != 1 {
					t.Fatalf("expected 1 matchExpression, got %d", len(wh.ObjectSelector.MatchExpressions))
				}
			},
		},
		{
			name: "mutating webhook with timeout override",
			config: Config{
				Mutating:                true,
				Name:                    "test.webhook.io",
				FailurePolicy:           "fail",
				SideEffects:             "None",
				Path:                    "/mutate",
				Groups:                  []string{"apps"},
				Resources:               []string{"deployments"},
				Versions:                []string{"v1"},
				Verbs:                   []string{"create"},
				AdmissionReviewVersions: []string{"v1"},
				TimeoutSeconds:          10,
				Patch:                   `{"timeoutSeconds":25}`,
			},
			validate: func(t *testing.T, wh admissionregv1.MutatingWebhook) {
				if wh.TimeoutSeconds == nil || *wh.TimeoutSeconds != 25 {
					t.Errorf("expected timeoutSeconds to be overridden to 25, got %v", wh.TimeoutSeconds)
				}
			},
		},
		{
			name: "invalid patch JSON",
			config: Config{
				Mutating:                true,
				Name:                    "test.webhook.io",
				FailurePolicy:           "fail",
				SideEffects:             "None",
				Path:                    "/mutate",
				Groups:                  []string{"apps"},
				Resources:               []string{"deployments"},
				Versions:                []string{"v1"},
				Verbs:                   []string{"create"},
				AdmissionReviewVersions: []string{"v1"},
				Patch:                   `{invalid json`,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			webhook, err := tt.config.ToMutatingWebhook()
			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.validate != nil {
				tt.validate(t, webhook)
			}
		})
	}
}

// TestConfigToValidatingWebhookWithPatch verifies that the Config.ToValidatingWebhook method
// correctly applies patches to the generated webhook.
func TestConfigToValidatingWebhookWithPatch(t *testing.T) {
	config := Config{
		Mutating:                false,
		Name:                    "test.webhook.io",
		FailurePolicy:           "fail",
		SideEffects:             "None",
		Path:                    "/validate",
		Groups:                  []string{"apps"},
		Resources:               []string{"deployments"},
		Versions:                []string{"v1"},
		Verbs:                   []string{"create", "update"},
		AdmissionReviewVersions: []string{"v1"},
		Patch:                   `{"namespaceSelector":{"matchLabels":{"env":"production"}}}`,
	}

	webhook, err := config.ToValidatingWebhook()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if webhook.NamespaceSelector == nil {
		t.Fatal("expected namespaceSelector to be set")
	}
	if webhook.NamespaceSelector.MatchLabels["env"] != "production" {
		t.Errorf("expected env=production, got %q", webhook.NamespaceSelector.MatchLabels["env"])
	}
}
