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

package v2

import (
	"fmt"
	"go/ast"
	"log"
	"strconv"
	"strings"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

// filterDescription parse comments above each field in the type definition.
func filterDescription(res string) string {
	var temp strings.Builder
	var desc string
	for _, comment := range strings.Split(res, "\n") {
		comment = strings.Trim(comment, " ")
		if !(strings.Contains(comment, "+kubebuilder") || strings.HasPrefix(comment, "+")) {
			temp.WriteString(comment)
			temp.WriteString(" ")
			desc = strings.TrimRight(temp.String(), " ")
		}
	}
	return desc
}

func processMarkersInComments(def *v1beta1.JSONSchemaProps, commentGroups ...*ast.CommentGroup) {
	for _, commentGroup := range commentGroups {
		for _, comment := range strings.Split(commentGroup.Text(), "\n") {
			getValidation(comment, def)
		}
	}
}

// This method is ported from controller-tools, it can removed when things are moved back.
// getValidation parses the validation tags from the comment and sets the
// validation rules on the given JSONSchemaProps.
// TODO: reduce the cyclomatic complexity and remove next line
//// nolint: gocyclo
func getValidation(comment string, props *v1beta1.JSONSchemaProps) {
	const arrayType = "array"
	comment = strings.TrimLeft(comment, " ")
	if !strings.HasPrefix(comment, "+kubebuilder:validation:") {
		return
	}
	c := strings.Replace(comment, "+kubebuilder:validation:", "", -1)
	parts := strings.Split(c, "=")
	if len(parts) != 2 {
		log.Fatalf("Expected +kubebuilder:validation:<key>=<value> actual: %s", comment)
		return
	}
	switch parts[0] {
	case "Maximum":
		f, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			log.Fatalf("Could not parse float from %s: %v", comment, err)
			return
		}
		props.Maximum = &f
	case "ExclusiveMaximum":
		b, err := strconv.ParseBool(parts[1])
		if err != nil {
			log.Fatalf("Could not parse bool from %s: %v", comment, err)
			return
		}
		props.ExclusiveMaximum = b
	case "Minimum":
		f, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			log.Fatalf("Could not parse float from %s: %v", comment, err)
			return
		}
		props.Minimum = &f
	case "ExclusiveMinimum":
		b, err := strconv.ParseBool(parts[1])
		if err != nil {
			log.Fatalf("Could not parse bool from %s: %v", comment, err)
			return
		}
		props.ExclusiveMinimum = b
	case "MaxLength":
		i, err := strconv.Atoi(parts[1])
		v := int64(i)
		if err != nil {
			log.Fatalf("Could not parse int from %s: %v", comment, err)
			return
		}
		props.MaxLength = &v
	case "MinLength":
		i, err := strconv.Atoi(parts[1])
		v := int64(i)
		if err != nil {
			log.Fatalf("Could not parse int from %s: %v", comment, err)
			return
		}
		props.MinLength = &v
	case "Pattern":
		props.Pattern = parts[1]
	case "MaxItems":
		if props.Type == arrayType {
			i, err := strconv.Atoi(parts[1])
			v := int64(i)
			if err != nil {
				log.Fatalf("Could not parse int from %s: %v", comment, err)
				return
			}
			props.MaxItems = &v
		}
	case "MinItems":
		if props.Type == arrayType {
			i, err := strconv.Atoi(parts[1])
			v := int64(i)
			if err != nil {
				log.Fatalf("Could not parse int from %s: %v", comment, err)
				return
			}
			props.MinItems = &v
		}
	case "UniqueItems":
		if props.Type == arrayType {
			b, err := strconv.ParseBool(parts[1])
			if err != nil {
				log.Fatalf("Could not parse bool from %s: %v", comment, err)
				return
			}
			props.UniqueItems = b
		}
	case "MultipleOf":
		f, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			log.Fatalf("Could not parse float from %s: %v", comment, err)
			return
		}
		props.MultipleOf = &f
	case "Enum":
		if props.Type != arrayType {
			value := strings.Split(parts[1], ",")
			enums := []v1beta1.JSON{}
			for _, s := range value {
				checkType(props, s, &enums)
			}
			props.Enum = enums
		}
	case "Format":
		props.Format = parts[1]
	default:
		log.Fatalf("Unsupport validation: %s", comment)
	}
}

// check type of enum element value to match type of field
func checkType(props *v1beta1.JSONSchemaProps, s string, enums *[]v1beta1.JSON) {
	switch props.Type {
	case "integer":
		if _, err := strconv.ParseInt(s, 0, 64); err != nil {
			log.Fatalf("Invalid integer value [%v] for a field of integer type", s)
		}
		*enums = append(*enums, v1beta1.JSON{Raw: []byte(fmt.Sprintf("%v", s))})
	case "float", "number":
		if _, err := strconv.ParseFloat(s, 64); err != nil {
			log.Fatalf("Invalid float value [%v] for a field of float type", s)
		}
		*enums = append(*enums, v1beta1.JSON{Raw: []byte(fmt.Sprintf("%v", s))})
	case "string":
		*enums = append(*enums, v1beta1.JSON{Raw: []byte(`"` + s + `"`)})
	}
}
