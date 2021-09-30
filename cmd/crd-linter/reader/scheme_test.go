/*
Copyright 2021 The Kubernetes Authors.

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

package reader

import (
	"reflect"
	"testing"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

func TestDecodeCustomResourceDefinitionV1Beta1(t *testing.T) {
	v1beta1Bytes := []byte(`apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: test-resource`)

	obj, gvk, err := DecodeCustomResourceDefinition(v1beta1Bytes, crdV1GVK.GroupVersion())
	if err != nil {
		t.Fatalf("Error decoding bytes: %v", err)
	}
	if !reflect.DeepEqual(*gvk, crdV1beta1GVK) {
		t.Fatalf("Unexpected GVK: %v", *gvk)
	}

	_, ok := obj.(*v1.CustomResourceDefinition)
	if !ok {
		t.Fatalf("Invalid object type: %T", obj)
	}
}

func TestDecodeCustomResourceDefinitionV1(t *testing.T) {
	v1beta1Bytes := []byte(`apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: test-resource`)

	obj, gvk, err := DecodeCustomResourceDefinition(v1beta1Bytes, crdV1GVK.GroupVersion())
	if err != nil {
		t.Fatalf("Error decoding bytes: %v", err)
	}
	if !reflect.DeepEqual(*gvk, crdV1GVK) {
		t.Fatalf("Unexpected GVK: %v", *gvk)
	}

	_, ok := obj.(*v1.CustomResourceDefinition)
	if !ok {
		t.Fatalf("Invalid object type: %T", obj)
	}
}

func TestDecodeCustomResourceDefinitionV1ToV1beta1(t *testing.T) {
	v1beta1Bytes := []byte(`apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: test-resource`)

	obj, gvk, err := DecodeCustomResourceDefinition(v1beta1Bytes, crdV1beta1GVK.GroupVersion())
	if err != nil {
		t.Fatalf("Error decoding bytes: %v", err)
	}
	if !reflect.DeepEqual(*gvk, crdV1GVK) {
		t.Fatalf("Unexpected GVK: %v", *gvk)
	}

	_, ok := obj.(*v1beta1.CustomResourceDefinition)
	if !ok {
		t.Fatalf("Invalid object type: %T", obj)
	}
}

func TestDecodeCustomResourceDefinitionV1beta1ToV1beta1(t *testing.T) {
	v1beta1Bytes := []byte(`apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: test-resource`)

	obj, gvk, err := DecodeCustomResourceDefinition(v1beta1Bytes, crdV1beta1GVK.GroupVersion())
	if err != nil {
		t.Fatalf("Error decoding bytes: %v", err)
	}
	if !reflect.DeepEqual(*gvk, crdV1beta1GVK) {
		t.Fatalf("Unexpected GVK: %v", *gvk)
	}

	_, ok := obj.(*v1beta1.CustomResourceDefinition)
	if !ok {
		t.Fatalf("Invalid object type: %T", obj)
	}
}
