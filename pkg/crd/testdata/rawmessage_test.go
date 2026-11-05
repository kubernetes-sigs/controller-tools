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
	"encoding/base64"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// The Plugin CRD has json.RawMessage and named byte slice fields. They are
// base64-encoded strings, so the API server accepts those and rejects other
// values. A byte slice typed as an object by its markers accepts a JSON
// object. Fields that share a named list each keep their own list type, keys
// and size, so the API server accepts the CRD and enforces them.
var _ = Describe("Plugin CRD", func() {
	config := base64.StdEncoding.EncodeToString([]byte(`{"logLevel":"debug"}`))
	certificate := base64.StdEncoding.EncodeToString([]byte("-----BEGIN CERTIFICATE-----\nMIIB\n-----END CERTIFICATE-----\n"))
	token := base64.StdEncoding.EncodeToString([]byte("my-secret-token"))

	port := func(name, protocol string, number int64) map[string]any {
		return map[string]any{"name": name, "protocol": protocol, "port": number}
	}

	newPlugin := func(name string, spec map[string]any) *unstructured.Unstructured {
		return &unstructured.Unstructured{
			Object: map[string]any{
				"apiVersion": "testdata.kubebuilder.io/v1",
				"kind":       "Plugin",
				"metadata": map[string]any{
					"name":      name,
					"namespace": "default",
				},
				"spec": spec,
			},
		}
	}

	It("should accept base64-encoded strings for byte slice fields and store them unchanged", func(ctx SpecContext) {
		plugin := newPlugin("logging-plugin", map[string]any{
			"config":            config,
			"caBundle":          certificate,
			"clientCertificate": certificate,
			"authToken":         token,
			"specDescriptors": []any{
				map[string]any{"path": "spec.replicas", "value": base64.StdEncoding.EncodeToString([]byte("3"))},
			},
		})
		Expect(k8sClient.Create(ctx, plugin)).To(Succeed())

		got := plugin.DeepCopy()
		Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(got), got)).To(Succeed())
		value, found, err := unstructured.NestedString(got.Object, "spec", "config")
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeTrue())
		Expect(value).To(Equal(config))
	})

	It("should accept a JSON object for a byte slice typed as an object", func(ctx SpecContext) {
		plugin := newPlugin("metrics-plugin", map[string]any{
			"values": map[string]any{"replicaCount": int64(2), "image": map[string]any{"tag": "v1.2.3"}},
		})
		Expect(k8sClient.Create(ctx, plugin)).To(Succeed())
	})

	It("should accept arrays and maps for named list and map fields", func(ctx SpecContext) {
		plugin := newPlugin("tracing-plugin", map[string]any{
			"tags":        []any{"observability", "tracing"},
			"keywords":    []any{"otel", "jaeger"},
			"topics":      []any{"traces", "spans"},
			"scopes":      []any{"read", "write"},
			"labels":      map[string]any{"app.kubernetes.io/name": "tracing"},
			"headers":     map[string]any{"X-Request-ID": "abc123"},
			"requests":    map[string]any{"cpu": "500m", "memory": "128Mi"},
			"nameservers": []any{"8.8.8.8", "8.8.4.4", "1.1.1.1"},
		})
		Expect(k8sClient.Create(ctx, plugin)).To(Succeed())
	})

	It("should keep each field's own list type for fields that share a named list", func(ctx SpecContext) {
		plugin := newPlugin("gateway-plugin", map[string]any{
			// A map list keyed by name: names are unique.
			"ports": []any{
				port("http", "TCP", 80),
				port("https", "TCP", 443),
			},
			// Atomic lists: two ports can have the same name.
			"exposedPorts": []any{
				port("http", "TCP", 80),
				port("http", "TCP", 8080),
			},
			"metricsPorts": []any{
				port("metrics", "TCP", 9090),
				port("metrics", "TCP", 9091),
			},
			"sidecarPorts": []any{
				port("proxy", "TCP", 15001),
				port("proxy", "TCP", 15006),
			},
			// Map lists keyed by protocol: two ports can have the same name.
			"protocolPorts": []any{
				port("dns", "TCP", 53),
				port("dns", "UDP", 53),
			},
			"healthPorts": []any{
				port("health", "TCP", 8081),
				port("health", "UDP", 8081),
			},
		})
		Expect(k8sClient.Create(ctx, plugin)).To(Succeed())
	})

	It("should accept a plugin without the optional fields", func(ctx SpecContext) {
		plugin := newPlugin("minimal-plugin", map[string]any{
			"specDescriptors": []any{
				map[string]any{"path": "spec.replicas"},
			},
		})
		Expect(k8sClient.Create(ctx, plugin)).To(Succeed())
	})

	DescribeTable("should reject values of the wrong type",
		func(ctx SpecContext, name string, spec map[string]any, message string) {
			plugin := newPlugin(name, spec)
			Expect(k8sClient.Create(ctx, plugin)).To(MatchError(ContainSubstring(message)))
		},
		Entry("object on json.RawMessage", "invalid-config-object",
			map[string]any{"config": map[string]any{"logLevel": "debug"}}, "must be of type byte"),
		Entry("integer on json.RawMessage", "invalid-config-integer",
			map[string]any{"config": int64(42)}, "must be of type byte"),
		Entry("string that is not base64-encoded on json.RawMessage", "invalid-config-string",
			map[string]any{"config": `{"logLevel":"debug"}`}, "must be of type byte"),
		Entry("object on json.RawMessage in a list", "invalid-descriptor-value",
			map[string]any{"specDescriptors": []any{map[string]any{"path": "spec.replicas", "value": map[string]any{"default": 3}}}}, "must be of type byte"),
		Entry("object on a named byte slice", "invalid-ca-bundle",
			map[string]any{"caBundle": map[string]any{"pem": "MIIB"}}, "must be of type byte"),
		Entry("object on a named byte slice from another package", "invalid-auth-token",
			map[string]any{"authToken": map[string]any{"token": "my-secret-token"}}, "must be of type byte"),
		Entry("string on a byte slice typed as an object", "invalid-values",
			map[string]any{"values": config}, "must be of type object"),
	)

	DescribeTable("should enforce the list type, keys and size of each named list field",
		func(ctx SpecContext, name string, spec map[string]any, message string) {
			plugin := newPlugin(name, spec)
			Expect(k8sClient.Create(ctx, plugin)).To(MatchError(ContainSubstring(message)))
		},
		Entry("two ports with the same name on a list keyed by name", "duplicate-port-name",
			map[string]any{"ports": []any{
				port("http", "TCP", 80),
				port("http", "TCP", 8080),
			}}, "Duplicate value"),
		Entry("two ports with the same protocol on a field keyed by protocol", "duplicate-port-protocol",
			map[string]any{"protocolPorts": []any{
				port("http", "TCP", 80),
				port("https", "TCP", 443),
			}}, "Duplicate value"),
		Entry("two ports with the same protocol on a field keyed by protocol with the k8s markers", "duplicate-health-port-protocol",
			map[string]any{"healthPorts": []any{
				port("liveness", "TCP", 8081),
				port("readiness", "TCP", 8082),
			}}, "Duplicate value"),
		Entry("no ports on a field that needs at least one", "no-service-ports",
			map[string]any{"servicePorts": []any{}}, "should have at least 1 items"),
		Entry("more ports than allowed on a field with a maximum", "too-many-service-ports",
			map[string]any{"servicePorts": []any{
				port("p1", "TCP", 8001), port("p2", "TCP", 8002), port("p3", "TCP", 8003),
				port("p4", "TCP", 8004), port("p5", "TCP", 8005), port("p6", "TCP", 8006),
				port("p7", "TCP", 8007), port("p8", "TCP", 8008), port("p9", "TCP", 8009),
				port("p10", "TCP", 8010), port("p11", "TCP", 8011),
			}}, "must have at most 10 items"),
	)
})
