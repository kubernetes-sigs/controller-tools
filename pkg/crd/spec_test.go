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
package crd

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

func TestApplyFieldScopes(t *testing.T) {
	testCases := []struct {
		desc     string
		scope    apiext.ResourceScope
		props    *apiext.JSONSchemaProps
		expected *apiext.JSONSchemaProps
	}{
		{
			desc:  "nested-cluster-scoped",
			scope: apiext.ClusterScoped,
			props: &apiext.JSONSchemaProps{
				Type: "object",
				Properties: map[string]apiext.JSONSchemaProps{
					"nonScopedRequired": {
						Type: "string",
					},
					"nonScopedOptional": {
						Type: "object",
						Properties: map[string]apiext.JSONSchemaProps{
							"innerCluster": {
								Type: "string",
							},
							"innerNamespaced": {
								Type: "string",
							},
							fieldScopePropertyName: {
								Type: "object",
								Properties: map[string]apiext.JSONSchemaProps{
									string(apiext.ClusterScoped): {
										Type:     "object",
										Required: []string{"innerCluster"},
									},
									string(apiext.NamespaceScoped): {
										Type:     "object",
										Required: []string{"innerNamespaced"}},
								},
							},
						},
						Required: []string{"innerCluster", "innerNamespaced"},
					},
					"clusterScopedRequired": {
						Type: "string",
					},
					"clusterScopedOptional": {
						Type: "string",
					},
					"namespaceScopedRequired": {
						Type: "string",
					},
					"namespaceScopedOptional": {
						Type: "string",
					},
					fieldScopePropertyName: {
						Type: "object",
						Properties: map[string]apiext.JSONSchemaProps{
							string(apiext.ClusterScoped): {
								Type:     "object",
								Required: []string{"clusterScopedRequired", "clusterScopedOptional"},
							},
							string(apiext.NamespaceScoped): {
								Type:     "object",
								Required: []string{"namespaceScopedRequired", "namespaceScopedOptional"}},
						},
					},
				},
				Required: []string{
					"nonScopedRequired",
					"clusterScopedRequired",
					"namespaceScopedRequired",
				},
			},
			expected: &apiext.JSONSchemaProps{
				Type: "object",
				Properties: map[string]apiext.JSONSchemaProps{
					"nonScopedRequired": {
						Type: "string",
					},
					"nonScopedOptional": {
						Type: "object",
						Properties: map[string]apiext.JSONSchemaProps{
							"innerCluster": {
								Type: "string",
							},
						},
						Required: []string{"innerCluster"},
					},
					"clusterScopedRequired": {
						Type: "string",
					},
					"clusterScopedOptional": {
						Type: "string",
					},
				},
				Required: []string{
					"nonScopedRequired",
					"clusterScopedRequired",
				},
			},
		},
		{
			desc:  "nested-namespaced",
			scope: apiext.NamespaceScoped,
			props: &apiext.JSONSchemaProps{
				Type: "object",
				Properties: map[string]apiext.JSONSchemaProps{
					"nonScopedRequired": {
						Type: "string",
					},
					"nonScopedOptional": {
						Type: "object",
						Properties: map[string]apiext.JSONSchemaProps{
							"innerCluster": {
								Type: "string",
							},
							"innerNamespaced": {
								Type: "string",
							},
							fieldScopePropertyName: {
								Type: "object",
								Properties: map[string]apiext.JSONSchemaProps{
									string(apiext.ClusterScoped): {
										Type:     "object",
										Required: []string{"innerCluster"},
									},
									string(apiext.NamespaceScoped): {
										Type:     "object",
										Required: []string{"innerNamespaced"}},
								},
							},
						},
						Required: []string{"innerCluster", "innerNamespaced"},
					},
					"clusterScopedRequired": {
						Type: "string",
					},
					"clusterScopedOptional": {
						Type: "string",
					},
					"namespaceScopedRequired": {
						Type: "string",
					},
					"namespaceScopedOptional": {
						Type: "string",
					},
					fieldScopePropertyName: {
						Type: "object",
						Properties: map[string]apiext.JSONSchemaProps{
							string(apiext.ClusterScoped): {
								Type:     "object",
								Required: []string{"clusterScopedRequired", "clusterScopedOptional"},
							},
							string(apiext.NamespaceScoped): {
								Type:     "object",
								Required: []string{"namespaceScopedRequired", "namespaceScopedOptional"}},
						},
					},
				},
				Required: []string{
					"nonScopedRequired",
					"clusterScopedRequired",
					"namespaceScopedRequired",
				},
			},
			expected: &apiext.JSONSchemaProps{
				Type: "object",
				Properties: map[string]apiext.JSONSchemaProps{
					"nonScopedRequired": {
						Type: "string",
					},
					"nonScopedOptional": {
						Type: "object",
						Properties: map[string]apiext.JSONSchemaProps{
							"innerNamespaced": {
								Type: "string",
							},
						},
						Required: []string{"innerNamespaced"},
					},
					"namespaceScopedRequired": {
						Type: "string",
					},
					"namespaceScopedOptional": {
						Type: "string",
					},
				},
				Required: []string{
					"nonScopedRequired",
					"namespaceScopedRequired",
				},
			},
		},
		{
			desc:  "cluster-scoped-array",
			scope: apiext.ClusterScoped,
			props: &apiext.JSONSchemaProps{
				Type: "array",
				Items: &apiext.JSONSchemaPropsOrArray{
					Schema: &apiext.JSONSchemaProps{
						Type: "object",
						Properties: map[string]apiext.JSONSchemaProps{
							"nonScopedRequired": {
								Type: "string",
							},
							"nonScopedOptional": {
								Type: "object",
								Properties: map[string]apiext.JSONSchemaProps{
									"innerCluster": {
										Type: "string",
									},
									"innerNamespaced": {
										Type: "string",
									},
									fieldScopePropertyName: {
										Type: "object",
										Properties: map[string]apiext.JSONSchemaProps{
											string(apiext.ClusterScoped): {
												Type:     "object",
												Required: []string{"innerCluster"},
											},
											string(apiext.NamespaceScoped): {
												Type:     "object",
												Required: []string{"innerNamespaced"}},
										},
									},
								},
								Required: []string{"innerCluster", "innerNamespaced"},
							},
							"clusterScopedRequired": {
								Type: "string",
							},
							"clusterScopedOptional": {
								Type: "string",
							},
							"namespaceScopedRequired": {
								Type: "string",
							},
							"namespaceScopedOptional": {
								Type: "string",
							},
							fieldScopePropertyName: {
								Type: "object",
								Properties: map[string]apiext.JSONSchemaProps{
									string(apiext.ClusterScoped): {
										Type:     "object",
										Required: []string{"clusterScopedRequired", "clusterScopedOptional"},
									},
									string(apiext.NamespaceScoped): {
										Type:     "object",
										Required: []string{"namespaceScopedRequired", "namespaceScopedOptional"}},
								},
							},
						},
						Required: []string{
							"nonScopedRequired",
							"clusterScopedRequired",
							"namespaceScopedRequired",
						},
					},
				},
			},
			expected: &apiext.JSONSchemaProps{
				Type: "array",
				Items: &apiext.JSONSchemaPropsOrArray{
					Schema: &apiext.JSONSchemaProps{
						Type: "object",
						Properties: map[string]apiext.JSONSchemaProps{
							"nonScopedRequired": {
								Type: "string",
							},
							"nonScopedOptional": {
								Type: "object",
								Properties: map[string]apiext.JSONSchemaProps{
									"innerCluster": {
										Type: "string",
									},
								},
								Required: []string{"innerCluster"},
							},
							"clusterScopedRequired": {
								Type: "string",
							},
							"clusterScopedOptional": {
								Type: "string",
							},
						},
						Required: []string{
							"nonScopedRequired",
							"clusterScopedRequired",
						},
					},
				},
			},
		},
		{
			desc:  "namespaced-array",
			scope: apiext.NamespaceScoped,
			props: &apiext.JSONSchemaProps{
				Type: "array",
				Items: &apiext.JSONSchemaPropsOrArray{
					Schema: &apiext.JSONSchemaProps{
						Type: "object",
						Properties: map[string]apiext.JSONSchemaProps{
							"nonScopedRequired": {
								Type: "string",
							},
							"nonScopedOptional": {
								Type: "object",
								Properties: map[string]apiext.JSONSchemaProps{
									"innerCluster": {
										Type: "string",
									},
									"innerNamespaced": {
										Type: "string",
									},
									fieldScopePropertyName: {
										Type: "object",
										Properties: map[string]apiext.JSONSchemaProps{
											string(apiext.ClusterScoped): {
												Type:     "object",
												Required: []string{"innerCluster"},
											},
											string(apiext.NamespaceScoped): {
												Type:     "object",
												Required: []string{"innerNamespaced"}},
										},
									},
								},
								Required: []string{"innerCluster", "innerNamespaced"},
							},
							"clusterScopedRequired": {
								Type: "string",
							},
							"clusterScopedOptional": {
								Type: "string",
							},
							"namespaceScopedRequired": {
								Type: "string",
							},
							"namespaceScopedOptional": {
								Type: "string",
							},
							fieldScopePropertyName: {
								Type: "object",
								Properties: map[string]apiext.JSONSchemaProps{
									string(apiext.ClusterScoped): {
										Type:     "object",
										Required: []string{"clusterScopedRequired", "clusterScopedOptional"},
									},
									string(apiext.NamespaceScoped): {
										Type:     "object",
										Required: []string{"namespaceScopedRequired", "namespaceScopedOptional"}},
								},
							},
						},
						Required: []string{
							"nonScopedRequired",
							"clusterScopedRequired",
							"namespaceScopedRequired",
						},
					},
				},
			},
			expected: &apiext.JSONSchemaProps{
				Type: "array",
				Items: &apiext.JSONSchemaPropsOrArray{
					Schema: &apiext.JSONSchemaProps{
						Type: "object",
						Properties: map[string]apiext.JSONSchemaProps{
							"nonScopedRequired": {
								Type: "string",
							},
							"nonScopedOptional": {
								Type: "object",
								Properties: map[string]apiext.JSONSchemaProps{
									"innerNamespaced": {
										Type: "string",
									},
								},
								Required: []string{"innerNamespaced"},
							},
							"namespaceScopedRequired": {
								Type: "string",
							},
							"namespaceScopedOptional": {
								Type: "string",
							},
						},
						Required: []string{
							"nonScopedRequired",
							"namespaceScopedRequired",
						},
					},
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			applyFieldScopes(tc.props, tc.scope)
			if diff := cmp.Diff(tc.props, tc.expected); diff != "" {
				t.Errorf("invalid field scopes (-want +got):\n%s", diff)
			}
		})
	}
}
