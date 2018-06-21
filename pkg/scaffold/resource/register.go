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

package resource

import (
	"path/filepath"

	"sigs.k8s.io/controller-tools/pkg/scaffold/input"
)

var _ input.File = &Register{}

// Register scaffolds the pkg/apis/group/version/register.go file
type Register struct {
	input.Input

	// Resource is the resource to scaffold the types_test.go file for
	Resource *Resource
}

// GetInput implements input.File
func (r *Register) GetInput() (input.Input, error) {
	if r.Path == "" {
		r.Path = filepath.Join("pkg", "apis", r.Resource.Group, r.Resource.Version, "register.go")
	}
	r.TemplateBody = registerTemplate
	return r.Input, nil
}

var registerTemplate = `{{ .Boilerplate }}

// NOTE: Boilerplate only.  Ignore this file.

// Package {{.Resource.Version}} contains API Schema definitions for the {{ .Resource.Group }} {{.Resource.Version}} API group
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen=package,register
// +k8s:conversion-gen={{ .Repo }}/pkg/apis/{{ .Resource.Group }}
// +k8s:defaulter-gen=TypeMeta
// +groupName={{ .Resource.Group }}.{{ .Domain }}
package {{.Resource.Version}}

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// KnownTypes is a collection of types to register with a Scheme
var KnownTypes = []runtime.Object{}

// SchemeGroupVersion is group version used to register these objects
var SchemeGroupVersion = schema.GroupVersion{Group: "{{ .Resource.Group }}.{{ .Domain }}", Version: "{{ .Resource.Version }}"}

// Kind takes an unqualified kind and returns back a Group qualified GroupKind
func Kind(kind string) schema.GroupKind {
	return SchemeGroupVersion.WithKind(kind).GroupKind()
}

// Resource takes an unqualified resource and returns a Group qualified GroupResource
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

var (
	// SchemeBuilder adds new types to a Scheme
	SchemeBuilder = runtime.NewSchemeBuilder(func(scheme *runtime.Scheme) error {
		scheme.AddKnownTypes(SchemeGroupVersion, KnownTypes...)
		metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
		return nil
	})
	// AddToScheme adds types to a Scheme
	AddToScheme = SchemeBuilder.AddToScheme
)
`
