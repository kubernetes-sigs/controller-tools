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

// Package annotation represents annotation module of controller-tool and kubebuilder.
// It serves as a mechanism for parsing kubebuilder style annotations (e.g. starting with // +kubuilder)
// and handling coresponding process.
//
// K8s
// +k8s:deepcopy-gen=<VALUE>
// +k8s:defaulter-gen=<VALUE>
// +k8s:conversion-gen=<CONVERSION_TARGET_DIR>
// +k8s:openapi-gen=true
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
//
// Kubebuilder Annotation Examples:
//
// +kubebuilder:subresource:status
// +kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.replicas
// +kubebuilder:subresource:scale:specpath=<jsonpath>,statuspath=<jsonpath>,selectorpath=<jsonpath>
// +kubebuilder:webhook:admission:groups=apps,resources=deployments,verbs=CREATE;UPDATE,name=bar-webhook,path=/bar,type=mutating,failure-policy=Fail
// +kubebuilder:webhook:serveroption:port=7890,cert-dir=/tmp/test-cert,service=test-system|webhook-service,selector=app|webhook-server,secret=test-system|webhook-secret,mutating-webhook-config-name=test-mutating-webhook-cfg,validating-webhook-config-name=test-validating-webhook-cfg
//
// Annoation schema:
// <header> : <module> : <parameter key values>
// Generic Annotation Spec <https://github.com/kubernetes-sigs/kubebuilder/issues/554>
package annotation
