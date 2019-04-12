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
	"log"
	"path/filepath"
	"strconv"
	"strings"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

// methods under this file is mostly copied from controller-tools old generator's parsing logic.
// A lot of places are not consistent :(
// This should be cleanup

// parseCRDs populates the CRD field of each Group.Version.Resource,
// creating validations using the annotations on type fields.
func parseCRDs(comments []string) *v1beta1.CustomResourceDefinitionSpec {
	if !IsAPIResource(comments) {
		return nil
	}

	crdVersion := v1beta1.CustomResourceDefinitionVersion{
		Name:    parseVersion(comments),
		Served:  true,
		Storage: isStorageVersion(comments),
	}

	crdSpec := &v1beta1.CustomResourceDefinitionSpec{
		Group:    "",
		Names:    v1beta1.CustomResourceDefinitionNames{},
		Versions: []v1beta1.CustomResourceDefinitionVersion{crdVersion},
	}

	if IsNonNamespaced(comments) {
		crdSpec.Scope = "Cluster"
	} else {
		crdSpec.Scope = "Namespaced"
	}

	if hasCategories(comments) {
		categoriesTag := getCategoriesTag(comments)
		categories := strings.Split(categoriesTag, ",")
		crdSpec.Names.Categories = categories
	}

	if hasSingular(comments) {
		singularName := getSingularName(comments)
		crdSpec.Names.Singular = singularName
	}

	if hasStatusSubresource(comments) {
		if crdSpec.Subresources == nil {
			crdSpec.Subresources = &v1beta1.CustomResourceSubresources{}
		}
		crdSpec.Subresources.Status = &v1beta1.CustomResourceSubresourceStatus{}
	}

	if hasScaleSubresource(comments) {
		if crdSpec.Subresources == nil {
			crdSpec.Subresources = &v1beta1.CustomResourceSubresources{}
		}
		jsonPath, err := parseScaleParams(comments)
		if err != nil {
			log.Fatalf("failed in parsing CRD, error: %v", err.Error())
		}
		crdSpec.Subresources.Scale = &v1beta1.CustomResourceSubresourceScale{
			SpecReplicasPath:   jsonPath[specReplicasPath],
			StatusReplicasPath: jsonPath[statusReplicasPath],
		}
		labelSelctor, ok := jsonPath[labelSelectorPath]
		if ok && labelSelctor != "" {
			crdSpec.Subresources.Scale.LabelSelectorPath = &labelSelctor
		}
	}
	if hasPrintColumn(comments) {
		result, err := parsePrintColumnParams(comments)
		if err != nil {
			log.Fatalf("failed to parse printcolumn annotations, error: %v", err.Error())
		}
		crdSpec.Versions[0].AdditionalPrinterColumns = result
	}

	rt, err := parseResourceAnnotation(comments)
	if err != nil {
		log.Fatalf("failed to parse resource annotations, error: %v", err.Error())
	}
	crdSpec.Names.Plural = rt.Resource
	if len(rt.ShortName) > 0 {
		crdSpec.Names.ShortNames = strings.Split(rt.ShortName, ";")
	}

	return crdSpec
}

const (
	specReplicasPath   = "specpath"
	statusReplicasPath = "statuspath"
	labelSelectorPath  = "selectorpath"
	jsonPathError      = "invalid scale path. specpath, statuspath key-value pairs are required, only selectorpath key-value is optinal. For example: // +kubebuilder:subresource:scale:specpath=.spec.replica,statuspath=.status.replica,selectorpath=.spec.Label"
	printColumnName    = "name"
	printColumnType    = "type"
	printColumnDescr   = "description"
	printColumnPath    = "JSONPath"
	printColumnFormat  = "format"
	printColumnPri     = "priority"
	printColumnError   = "invalid printcolumn path. name,type, and JSONPath are required kye-value pairs and rest of the fields are optinal. For example: // +kubebuilder:printcolumn:name=abc,type=string,JSONPath=status"
)

// IsAPIResource returns true if either of the two conditions become true:
// 1. t has a +resource/+kubebuilder:resource comment tag
// 2. t has TypeMeta and ObjectMeta in its member list.
func IsAPIResource(comments []string) bool {
	for _, c := range comments {
		if strings.Contains(c, "+resource") || strings.Contains(c, "+kubebuilder:resource") {
			return true
		}
	}

	return false
}

// IsNonNamespaced returns true if t has a +nonNamespaced comment tag
func IsNonNamespaced(comments []string) bool {
	if !IsAPIResource(comments) {
		return false
	}

	for _, c := range comments {
		if strings.Contains(c, "+genclient:nonNamespaced") {
			return true
		}
	}

	return false
}

// hasPrintColumn returns true if t has a +printcolumn or +kubebuilder:printcolumn annotation.
func hasPrintColumn(comments []string) bool {
	for _, c := range comments {
		if strings.Contains(c, "+printcolumn") || strings.Contains(c, "+kubebuilder:printcolumn") {
			return true
		}
	}
	return false
}

// IsAPISubresource returns true if t has a +subresource-request comment tag
func IsAPISubresource(comments []string) bool {
	for _, c := range comments {
		if strings.Contains(c, "+subresource-request") {
			return true
		}
	}
	return false
}

// HasSubresource returns true if t is an APIResource with one or more Subresources
func HasSubresource(comments []string) bool {
	if !IsAPIResource(comments) {
		return false
	}
	for _, c := range comments {
		if strings.Contains(c, "subresource") {
			return true
		}
	}
	return false
}

// hasStatusSubresource returns true if t is an APIResource annotated with
// +kubebuilder:subresource:status
func hasStatusSubresource(comments []string) bool {
	if !IsAPIResource(comments) {
		return false
	}
	for _, c := range comments {
		if strings.Contains(c, "+kubebuilder:subresource:status") {
			return true
		}
	}
	return false
}

// hasScaleSubresource returns true if t is an APIResource annotated with
// +kubebuilder:subresource:scale
func hasScaleSubresource(comments []string) bool {
	if !IsAPIResource(comments) {
		return false
	}
	for _, c := range comments {
		if strings.Contains(c, "+kubebuilder:subresource:scale") {
			return true
		}
	}
	return false
}

// hasCategories returns true if t is an APIResource annotated with
// +kubebuilder:categories
func hasCategories(comments []string) bool {
	if !IsAPIResource(comments) {
		return false
	}

	for _, c := range comments {
		if strings.Contains(c, "+kubebuilder:categories") {
			return true
		}
	}
	return false
}

// HasDocAnnotation returns true if t is an APIResource with doc annotation
// +kubebuilder:doc
func HasDocAnnotation(comments []string) bool {
	if !IsAPIResource(comments) {
		return false
	}
	for _, c := range comments {
		if strings.Contains(c, "+kubebuilder:doc") {
			return true
		}
	}
	return false
}

// hasSingular returns true if t is an APIResource annotated with
// +kubebuilder:singular
func hasSingular(comments []string) bool {
	if !IsAPIResource(comments) {
		return false
	}
	for _, c := range comments {
		if strings.Contains(c, "+kubebuilder:singular") {
			return true
		}
	}
	return false
}

// resourceTags contains the tags present in a "+resource=" comment
type resourceTags struct {
	Resource  string
	REST      string
	Strategy  string
	ShortName string
}

// ParseKV parses key-value string formatted as "foo=bar" and returns key and value.
func ParseKV(s string) (key, value string, err error) {
	kv := strings.Split(s, "=")
	if len(kv) != 2 {
		err = fmt.Errorf("invalid key value pair")
		return key, value, err
	}
	key, value = kv[0], kv[1]
	if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
		value = value[1 : len(value)-1]
	}
	return key, value, err
}

// resourceAnnotationValue is a helper function to extract resource annotation.
func resourceAnnotationValue(tag string) (resourceTags, error) {
	res := resourceTags{}
	for _, elem := range strings.Split(tag, ",") {
		key, value, err := ParseKV(elem)
		if err != nil {
			return resourceTags{}, fmt.Errorf("// +kubebuilder:resource: tags must be key value pairs.  Expected "+
				"keys [path=<resourcepath>] "+
				"Got string: [%s]", tag)
		}
		switch key {
		case "path":
			res.Resource = value
		case "shortName":
			res.ShortName = value
		default:
			return resourceTags{}, fmt.Errorf("The given input %s is invalid", value)
		}
	}
	return res, nil
}

// GetAnnotation extracts the annotation from comment text.
// It will return "foo" for comment "+kubebuilder:webhook:foo" .
func GetAnnotation(c, name string) string {
	prefix := fmt.Sprintf("+%s:", name)
	if strings.HasPrefix(c, prefix) {
		return strings.TrimPrefix(c, prefix)
	}
	return ""
}

// parseResourceAnnotation parses the tags in a "+resource=" comment into a resourceTags struct.
func parseResourceAnnotation(comments []string) (resourceTags, error) {
	finalResult := resourceTags{}
	var resourceAnnotationFound bool
	for _, comment := range comments {
		anno := GetAnnotation(comment, "kubebuilder:resource")
		if len(anno) == 0 {
			continue
		}
		result, err := resourceAnnotationValue(anno)
		if err != nil {
			return resourceTags{}, err
		}
		if resourceAnnotationFound {
			return resourceTags{}, fmt.Errorf("resource annotation should only exists once per type")
		}
		resourceAnnotationFound = true
		finalResult = result
	}
	return finalResult, nil
}

// GetVersion returns version of t.
func GetVersion(pwd string) string {
	return filepath.Base(pwd)
}

// Comments is a structure for using comment tags on go structs and fields
type Comments []string

// GetTags returns the value for the first comment with a prefix matching "+name="
// e.g. "+name=foo\n+name=bar" would return "foo"
func (c Comments) getTag(name, sep string) string { // nolint: unparam
	for _, c := range c {
		prefix := fmt.Sprintf("+%s%s", name, sep)
		if strings.HasPrefix(c, prefix) {
			return strings.Replace(c, prefix, "", 1)
		}
	}
	return ""
}

// hasTag returns true if the Comments has a tag with the given name
func (c Comments) hasTag(name string) bool {
	for _, c := range c {
		prefix := fmt.Sprintf("+%s", name)
		if strings.HasPrefix(c, prefix) {
			return true
		}
	}
	return false
}

// GetTags returns the value for all comments with a prefix and separator.  E.g. for "name" and "="
// "+name=foo\n+name=bar" would return []string{"foo", "bar"}
func (c Comments) getTags(name, sep string) []string {
	tags := []string{}
	for _, c := range c {
		prefix := fmt.Sprintf("+%s%s", name, sep)
		if strings.HasPrefix(c, prefix) {
			tags = append(tags, strings.Replace(c, prefix, "", 1))
		}
	}
	return tags
}

// getKVTags returns the value for all comments with a prefix and separator.  E.g. for "name" and "="
// "+name=foo\n+name=bar" would return []string{"foo", "bar"}
func (c Comments) getKVTags(prefix, sep string) []string { // nolint: unparam
	tags := []string{}
	for _, c := range c {
		if strings.HasPrefix(c, prefix) {
			rawKVs := strings.Replace(c, prefix, "", 1)
			for _, rawKV := range strings.Split(rawKVs, ",") {
				strings.Split(rawKV, "=")
			}
		}
	}
	return tags
}

// getCategoriesTag returns the value of the +kubebuilder:categories tags
func getCategoriesTag(comments []string) string {
	cs := Comments(comments)
	resource := cs.getTag("kubebuilder:categories", "=")
	if len(resource) == 0 {
		panic(fmt.Errorf("must specify +kubebuilder:categories comment"))
	}
	return resource
}

// getSingularName returns the value of the +kubebuilder:singular tag
func getSingularName(comments []string) string {
	cs := Comments(comments)
	singular := cs.getTag("kubebuilder:singular", "=")
	if len(singular) == 0 {
		panic(fmt.Errorf("must specify a value to use with +kubebuilder:singular comment"))
	}
	return singular
}

// Scale subresource requires specpath, statuspath, selectorpath key values, represents for JSONPath of
// SpecReplicasPath, StatusReplicasPath, LabelSelectorPath separately. e.g.
// +kubebuilder:subresource:scale:specpath=.spec.replica,statuspath=.status.replica,selectorpath=
func parseScaleParams(comments []string) (map[string]string, error) {
	jsonPath := make(map[string]string)
	for _, c := range comments {
		if strings.Contains(c, "+kubebuilder:subresource:scale") {
			paths := strings.Replace(c, "+kubebuilder:subresource:scale:", "", -1)
			path := strings.Split(paths, ",")
			if len(path) < 2 {
				return nil, fmt.Errorf(jsonPathError)
			}
			for _, s := range path {
				kv := strings.Split(s, "=")
				if kv[0] == specReplicasPath || kv[0] == statusReplicasPath || kv[0] == labelSelectorPath {
					jsonPath[kv[0]] = kv[1]
				} else {
					return nil, fmt.Errorf(jsonPathError)
				}
			}
			var ok bool
			_, ok = jsonPath[specReplicasPath]
			if !ok {
				return nil, fmt.Errorf(jsonPathError)
			}
			_, ok = jsonPath[statusReplicasPath]
			if !ok {
				return nil, fmt.Errorf(jsonPathError)
			}
			return jsonPath, nil
		}
	}
	return nil, fmt.Errorf(jsonPathError)
}

// printColumnKV parses key-value string formatted as "foo=bar" and returns key and value.
func printColumnKV(s string) (key, value string, err error) {
	kv := strings.SplitN(s, "=", 2)
	if len(kv) != 2 {
		err = fmt.Errorf("invalid key value pair")
		return key, value, err
	}
	key, value = kv[0], kv[1]
	if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
		value = value[1 : len(value)-1]
	}
	return key, value, err
}

// helperPrintColumn is a helper function for the parsePrintColumnParams to compute printer columns.
// TODO: reduce the cyclomatic complexity and remove next line
// nolint: gocyclo,goconst
func helperPrintColumn(parts string) (v1beta1.CustomResourceColumnDefinition, error) {
	config := v1beta1.CustomResourceColumnDefinition{}
	var count int
	part := strings.Split(parts, ",")
	if len(part) < 3 {
		return v1beta1.CustomResourceColumnDefinition{}, fmt.Errorf(printColumnError)
	}

	for _, elem := range strings.Split(parts, ",") {
		key, value, err := printColumnKV(elem)
		if err != nil {
			return v1beta1.CustomResourceColumnDefinition{},
				fmt.Errorf("//+kubebuilder:printcolumn: tags must be key value pairs.Expected "+
					"keys [name=<name>,type=<type>,description=<descr>,format=<format>] "+
					"Got string: [%s]", parts)
		}
		if key == printColumnName || key == printColumnType || key == printColumnPath {
			count++
		}
		switch key {
		case printColumnName:
			config.Name = value
		case printColumnType:
			if value == "integer" || value == "number" || value == "string" || value == "boolean" || value == "date" {
				config.Type = value
			} else {
				return v1beta1.CustomResourceColumnDefinition{}, fmt.Errorf("invalid value for %s printcolumn", printColumnType)
			}
		case printColumnFormat:
			if config.Type == "integer" && (value == "int32" || value == "int64") {
				config.Format = value
			} else if config.Type == "number" && (value == "float" || value == "double") {
				config.Format = value
			} else if config.Type == "string" && (value == "byte" || value == "date" || value == "date-time" || value == "password") {
				config.Format = value
			} else {
				return v1beta1.CustomResourceColumnDefinition{}, fmt.Errorf("invalid value for %s printcolumn", printColumnFormat)
			}
		case printColumnPath:
			config.JSONPath = value
		case printColumnPri:
			i, err := strconv.Atoi(value)
			v := int32(i)
			if err != nil {
				return v1beta1.CustomResourceColumnDefinition{}, fmt.Errorf("invalid value for %s printcolumn", printColumnPri)
			}
			config.Priority = v
		case printColumnDescr:
			config.Description = value
		default:
			return v1beta1.CustomResourceColumnDefinition{}, fmt.Errorf(printColumnError)
		}
	}
	if count != 3 {
		return v1beta1.CustomResourceColumnDefinition{}, fmt.Errorf(printColumnError)
	}
	return config, nil
}

// printcolumn requires name,type,JSONPath fields and rest of the field are optional
// +kubebuilder:printcolumn:name=<name>,type=<type>,description=<desc>,JSONPath:<.spec.Name>,priority=<int32>,format=<format>
func parsePrintColumnParams(comments []string) ([]v1beta1.CustomResourceColumnDefinition, error) {
	result := []v1beta1.CustomResourceColumnDefinition{}
	for _, comment := range comments {
		if strings.Contains(comment, "+kubebuilder:printcolumn") {
			parts := strings.Replace(comment, "+kubebuilder:printcolumn:", "", -1)
			res, err := helperPrintColumn(parts)
			if err != nil {
				return []v1beta1.CustomResourceColumnDefinition{}, err
			}
			result = append(result, res)
		}
	}
	return result, nil
}

func parseVersion(comments []string) string {
	return Comments(comments).getTag("kubebuilder:crd:version", "=")
}

func isStorageVersion(comments []string) bool {
	storage := strings.ToLower(Comments(comments).getTag("kubebuilder:crd:storage", "="))
	if len(storage) > 0 {
		switch storage {
		case "true":
			return true
		case "false":
			return false
		default:
			log.Fatalf("the value associated with kubebuilder:crd:storage should either true or false")
		}
	}
	return false
}
