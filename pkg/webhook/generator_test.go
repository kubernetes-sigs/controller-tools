/*
Copyright 2018 The Kubernetes Authors.

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
	"reflect"
	"testing"

	"github.com/spf13/afero"
)

var expected = map[string]string{
	"config/webhook/webhookmanifests.yaml": `apiVersion: admissionregistration.k8s.io/v1beta1
kind: MutatingWebhookConfiguration
metadata:
  annotations:
    alpha.admissionwebhook.cert-manager.io: "true"
  creationTimestamp: null
  name: test-mutating-webhook-cfg
webhooks:
- clientConfig:
    caBundle: XG4=
    service:
      name: webhook-service
      namespace: test-system
      path: /bar
  failurePolicy: Fail
  name: bar-webhook
  namespaceSelector:
    matchExpressions:
    - key: control-plane
      operator: DoesNotExist
  rules:
  - apiGroups:
    - apps
    operations:
    - CREATE
    - UPDATE
    resources:
    - deployments
---
apiVersion: admissionregistration.k8s.io/v1beta1
kind: ValidatingWebhookConfiguration
metadata:
  annotations:
    alpha.admissionwebhook.cert-manager.io: "true"
  creationTimestamp: null
  name: test-validating-webhook-cfg
webhooks:
- clientConfig:
    caBundle: XG4=
    service:
      name: webhook-service
      namespace: test-system
      path: /baz
  failurePolicy: Ignore
  name: baz-webhook
  namespaceSelector:
    matchExpressions:
    - key: control-plane
      operator: DoesNotExist
  rules:
  - apiGroups:
    - crew
    apiVersions:
    - v1
    operations:
    - DELETE
    resources:
    - firstmates
---
apiVersion: v1
kind: Service
metadata:
  annotations:
    alpha.service.cert-manager.io/serving-cert-secret-name: webhook-secret
  creationTimestamp: null
  name: webhook-service
  namespace: test-system
spec:
  ports:
  - port: 443
    targetPort: 7890
  selector:
    app: webhook-server
status:
  loadBalancer: {}
`,
	"config/default/manager_patch.yaml": `apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: controller-manager
spec:
  template:
    metadata:
      labels:
        app: webhook-server
    spec:
      containers:
      - name: manager
        ports:
        - containerPort: 7890
          name: webhook-server
          protocol: TCP
        volumeMounts:
        - mountPath: /tmp/test-cert
          name: cert
          readOnly: true
      volumes:
      - name: cert
        secret:
          defaultMode: 420
          secretName: webhook-secret
`,
}

func TestGenerator(t *testing.T) {
	fs := afero.NewMemMapFs()

	o := &Options{
		WriterOptions: WriterOptions{
			InputDir: "./testdata/input",
			outFs:    fs,
		},
	}
	o.SetDefaults()

	if err := Generate(o); err != nil {
		t.Fatalf("error when generating the files: %v", err)
	}

	for name, content := range expected {
		got := make([]byte, 2048)
		f1, err := fs.Open(name)
		if err != nil {
			t.Fatalf("error when opening generated file %s: %v", name, err)
		}
		n, err := f1.Read(got)
		if err != nil {
			t.Fatalf("error when reading from generated file %s: %v", name, err)
		}
		if !reflect.DeepEqual([]byte(content), got[:n]) {
			t.Fatalf("expected: %v, but got: %v", content, string(got[:n]))
		}
	}
}
