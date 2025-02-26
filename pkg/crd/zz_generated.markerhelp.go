//go:build !ignore_autogenerated

/*
Copyright2019 The Kubernetes Authors.

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

// Code generated by helpgen. DO NOT EDIT.

package crd

import (
	"sigs.k8s.io/controller-tools/pkg/markers"
)

func (Generator) Help() *markers.DefinitionHelp {
	return &markers.DefinitionHelp{
		Category: "",
		DetailedHelp: markers.DetailedHelp{
			Summary: "generates CustomResourceDefinition objects.",
			Details: "",
		},
		FieldHelp: map[string]markers.DetailedHelp{
			"IgnoreUnexportedFields": {
				Summary: "indicates that we should skip unexported fields.",
				Details: "Left unspecified, the default is false.",
			},
			"AllowDangerousTypes": {
				Summary: "allows types which are usually omitted from CRD generation",
				Details: "because they are not recommended.\n\nCurrently the following additional types are allowed when this is true:\nfloat32\nfloat64\n\nLeft unspecified, the default is false",
			},
			"MaxDescLen": {
				Summary: "specifies the maximum description length for fields in CRD's OpenAPI schema.",
				Details: "0 indicates drop the description for all fields completely.\nn indicates limit the description to at most n characters and truncate the description to\nclosest sentence boundary if it exceeds n characters.",
			},
			"CRDVersions": {
				Summary: "specifies the target API versions of the CRD type itself to",
				Details: "generate. Defaults to v1.\n\nCurrently, the only supported value is v1.\n\nThe first version listed will be assumed to be the \"default\" version and\nwill not get a version suffix in the output filename.\n\nYou'll need to use \"v1\" to get support for features like defaulting,\nalong with an API server that supports it (Kubernetes 1.16+).",
			},
			"GenerateEmbeddedObjectMeta": {
				Summary: "specifies if any embedded ObjectMeta in the CRD should be generated",
				Details: "",
			},
			"HeaderFile": {
				Summary: "specifies the header text (e.g. license) to prepend to generated files.",
				Details: "",
			},
			"Year": {
				Summary: "specifies the year to substitute for \" YEAR\" in the header file.",
				Details: "",
			},
			"DeprecatedV1beta1CompatibilityPreserveUnknownFields": {
				Summary: "indicates whether",
				Details: "or not we should turn off field pruning for this resource.\n\nSpecifies spec.preserveUnknownFields value that is false and omitted by default.\nThis value can only be specified for CustomResourceDefinitions that were created with\n`apiextensions.k8s.io/v1beta1`.\n\nThe field can be set for compatibility reasons, although strongly discouraged, resource\nauthors should move to a structural OpenAPI schema instead.\n\nSee https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/#field-pruning\nfor more information about field pruning and v1beta1 resources compatibility.",
			},
		},
	}
}
