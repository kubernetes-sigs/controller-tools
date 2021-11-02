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

package linters

import (
	"fmt"

	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

func recurseAllSchemas(versions []v1.CustomResourceDefinitionVersion, fn func(props v1.JSONSchemaProps, path string) []Warning) []Warning {
	var errs []Warning
	for i, vers := range versions {
		// Skip versions that do not specify a schema
		if vers.Schema == nil || vers.Schema.OpenAPIV3Schema == nil {
			continue
		}

		errs = append(errs, recurseAllProps(*vers.Schema.OpenAPIV3Schema, fmt.Sprintf("spec.versions[%d].schema.openAPIV3Schema", i), fn)...)
	}
	return errs
}

func recurseAllProps(props v1.JSONSchemaProps, path string, fn func(props v1.JSONSchemaProps, path string) []Warning) []Warning {
	var errs []Warning
	errs = append(errs, fn(props, path)...)
	for i, val := range props.AnyOf {
		errs = append(errs, recurseAllProps(val, fmt.Sprintf("%s.anyOf[%d]", path, i), fn)...)
	}
	for i, val := range props.OneOf {
		errs = append(errs, recurseAllProps(val, fmt.Sprintf("%s.oneOf[%d]", path, i), fn)...)
	}
	for i, val := range props.AllOf {
		errs = append(errs, recurseAllProps(val, fmt.Sprintf("%s.allOf[%d]", path, i), fn)...)
	}
	for name, val := range props.Properties {
		errs = append(errs, recurseAllProps(val, fmt.Sprintf("%s.properties.%s", path, name), fn)...)
	}
	if props.Items != nil {
		if props.Items.Schema != nil {
			errs = append(errs, recurseAllProps(*props.Items.Schema, fmt.Sprintf("%s.items", path), fn)...)
		}
		for i, v := range props.Items.JSONSchemas {
			errs = append(errs, recurseAllProps(v, fmt.Sprintf("%s.items[%d]", path, i), fn)...)
		}
	}
	return errs
}
