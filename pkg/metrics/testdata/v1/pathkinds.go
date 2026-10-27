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

package v1

import (
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// PathKinds is no custom resource. It has the fields used to test how the kinds of paths
// and values are derived from Go types in the tests of the metrics package.
type PathKinds struct {
	// Inner is inlined with the inline option.
	Inner `json:",inline"`
	// InnerNamed is inlined as the inline option is set, even though a name is set.
	InnerNamed `json:"x,inline"`
	// InnerNoTag is embedded without a json tag.
	InnerNoTag

	// NotInlined contains inline in its name only, so it must not be inlined.
	NotInlined Inner `json:"inlineFoo"`

	// Map is a map of structs.
	Map map[string]Inner `json:"map"`

	// Time is a timestamp.
	Time metav1.Time `json:"time"`
	// MicroTime is a timestamp with microseconds.
	MicroTime metav1.MicroTime `json:"microTime"`
	// TimePointer is a pointer to a timestamp.
	TimePointer *metav1.Time `json:"timePointer"`
	// TimeAlias is an alias of a timestamp.
	TimeAlias TimeAlias `json:"timeAlias"`
	// TimeAliasPointer is an alias of a pointer to a timestamp.
	TimeAliasPointer TimePointerAlias `json:"timeAliasPointer"`
	// PointerToTimeAlias is a pointer to an alias of a timestamp.
	PointerToTimeAlias *TimeAlias `json:"pointerToTimeAlias"`
	// OtherTime has the name of a metav1 timestamp, but is no timestamp.
	OtherTime Time `json:"otherTime"`
	// Duration is a type of metav1, which is no timestamp.
	Duration metav1.Duration `json:"duration"`

	// Any, Raw and JSON have no schema in Go, so no path can be resolved below them.
	Any  any                  `json:"any"`
	Raw  runtime.RawExtension `json:"raw"`
	JSON apiextensionsv1.JSON `json:"json"`
}

// Inner is a struct which gets inlined or used as a field.
type Inner struct {
	A string `json:"a"`
}

// InnerNamed is a struct which gets inlined with a name.
type InnerNamed struct {
	B string `json:"b"`
}

// InnerNoTag is a struct which gets embedded without a tag.
type InnerNoTag struct {
	C string `json:"c"`
}

// TimeAlias is an alias of metav1.Time.
type TimeAlias = metav1.Time

// TimePointerAlias is an alias of a pointer to metav1.Time.
type TimePointerAlias = *metav1.Time

// Time is not the timestamp of metav1.
type Time struct {
	Seconds int64 `json:"seconds"`
}
