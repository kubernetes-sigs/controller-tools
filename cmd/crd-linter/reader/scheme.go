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
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	apiextensionsinstall "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/install"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/apimachinery/pkg/runtime/serializer/versioning"
)

var (
	// defaultScheme is used to decode CustomResourceDefinition objects into
	// the required 'v1' API version.
	defaultScheme = runtime.NewScheme()

	// codec is used to decode CRD resources into the desired GVK
	codec runtime.Serializer

	crdV1GVK      = schema.GroupVersionKind{Group: "apiextensions.k8s.io", Version: "v1", Kind: "CustomResourceDefinition"}
	crdV1beta1GVK = schema.GroupVersionKind{Group: "apiextensions.k8s.io", Version: "v1beta1", Kind: "CustomResourceDefinition"}
)

func init() {
	apiextensionsinstall.Install(defaultScheme)
	metav1.AddToGroupVersion(defaultScheme, schema.GroupVersion{Version: "v1"})

	serializer := json.NewSerializerWithOptions(json.DefaultMetaFactory, defaultScheme, defaultScheme, json.SerializerOptions{Yaml: true})
	codec = versioning.NewCodec(serializer, serializer, defaultScheme, defaultScheme, defaultScheme, nil, runtime.InternalGroupVersioner, runtime.InternalGroupVersioner, "")
}

func DecodeCustomResourceDefinition(data []byte, targetGV schema.GroupVersion) (runtime.Object, *schema.GroupVersionKind, error) {
	internalObj := &apiextensions.CustomResourceDefinition{}
	_, gvk, err := codec.Decode(data, nil, internalObj)
	if err != nil {
		return nil, nil, err
	}
	obj, err := defaultScheme.ConvertToVersion(internalObj, targetGV)
	if err != nil {
		return nil, nil, err
	}
	return obj, gvk, nil
}
