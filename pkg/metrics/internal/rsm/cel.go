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

package rsm

import (
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"sigs.k8s.io/controller-tools/pkg/metrics/internal/model"
)

// resolverCEL is the resolver of resource-state-metrics which evaluates CEL expressions.
const resolverCEL = "cel"

// toStore converts a resource to a resource-state-metrics store. All expressions are
// CEL expressions, as the store uses the CEL resolver. The object is available as `o`.
func toStore(resource model.Resource) Store {
	resourceLabels := labelsToCEL(resource.Labels, nil, model.PathKindScalar, "")
	store := Store{
		Group:    resource.GroupVersionKind.Group,
		Version:  resource.GroupVersionKind.Version,
		Kind:     resource.GroupVersionKind.Kind,
		Resource: resource.Plural(),
		Resolver: resolverCEL,
	}

	prefix := ""
	if resource.NamePrefix != nil && *resource.NamePrefix != "" {
		prefix = *resource.NamePrefix + "_"
	}
	for _, metric := range resource.Metrics {
		family := Family{Name: prefix + metric.Name, Help: metric.Help}
		switch metric.Type {
		case model.MetricTypeGauge:
			family.Metrics = []Metric{gaugeToCEL(metric)}
		case model.MetricTypeInfo:
			family.Metrics = []Metric{infoToCEL(metric)}
		case model.MetricTypeStateSet:
			family.Metrics = stateSetToCEL(metric)
		}
		// resource-state-metrics v0.1.0 mutates the labels of a family when inheriting the
		// labels of the store for every processed object. Put the labels of the resource
		// directly on each metric to avoid duplicate labels on the second and subsequent objects.
		for i := range family.Metrics {
			family.Metrics[i].Labels = append(family.Metrics[i].Labels, resourceLabels...)
		}
		store.Families = append(store.Families, family)
	}
	return store
}

// elements describes the elements of the list or map a metric is generated for. A list is
// iterated by its elements (`v`), a map by its keys (`k`) with the element being `base[k]`.
type elements struct {
	// guard checks the presence of the list or map.
	guard string
	// base is the expression of the list or map.
	base string
	kind model.PathKind
}

func newElements(path model.Path, kind model.PathKind) elements {
	guard, base := guardedPathCEL("o", path)
	return elements{guard: guard, base: base, kind: kind}
}

// isIterated returns true if the path is a list or map, for which a sample per element is generated.
func isIterated(path model.Path, kind model.PathKind) bool {
	return len(path) > 0 && (kind == model.PathKindArray || kind == model.PathKindMap)
}

func (e elements) variable() string {
	if e.kind == model.PathKindMap {
		return "k"
	}
	return "v"
}

// element returns the expression of an element, inside of a function using the variable.
func (e elements) element() string {
	if e.kind == model.PathKindMap {
		return e.base + "[k]"
	}
	return "v"
}

// mapCEL returns the expression evaluating expr for each element which fulfills the filter.
// The labels and the value of a metric have to use the same filter, as their lists are aligned.
func (e elements) mapCEL(filter, expr string) string {
	items := e.base
	if filter != "" {
		items = fmt.Sprintf("%s.filter(%s, %s)", e.base, e.variable(), filter)
	}
	return fmt.Sprintf("%s ? %s.map(%s, %s) : []", e.guard, items, e.variable(), expr)
}

// valueFilter returns the filter for the elements having the value. A list or map entry
// without the value would otherwise fail the evaluation of the whole metric.
func (e elements) valueFilter(value model.Path) string {
	if len(value) == 0 {
		return ""
	}
	guard, _ := guardedPathCEL(e.element(), value)
	return guard
}

func gaugeToCEL(m model.Generator) Metric {
	timestamp := m.ValueKind == model.ValueKindTimestamp

	if isIterated(m.Path, m.PathKind) {
		e := newElements(m.Path, m.PathKind)
		valueGuard, valueExpr := guardedPathCEL(e.element(), m.Value)
		valueExpr = timestampCEL(valueExpr, timestamp)

		// Elements without the value are exposed as zero or left out, if not set.
		filter := ""
		value := valueExpr
		if len(m.Value) > 0 {
			if m.MissingAsZero {
				value = fmt.Sprintf("%s ? %s : 0", valueGuard, valueExpr)
			} else {
				filter = valueGuard
			}
		}

		labels := labelsToCEL(m.Labels, m.Path, m.PathKind, filter)
		if m.KeyLabel != "" && m.PathKind == model.PathKindMap {
			labels = append(labels, Label{Name: m.KeyLabel, Value: e.mapCEL(filter, "k")})
		}
		return Metric{Labels: labels, Value: e.mapCEL(filter, value)}
	}

	valuePath := append(slices.Clone(m.Path), m.Value...)
	value := timestampCEL(pathCEL("o", valuePath), timestamp)
	guard, _ := guardedPathCEL("o", valuePath)
	if m.MissingAsZero {
		value = fmt.Sprintf("%s ? %s : 0", guard, value)
	} else if len(valuePath) > 0 {
		value = fmt.Sprintf("%s ? [%s] : []", guard, value)
	}
	return Metric{Labels: labelsToCEL(m.Labels, m.Path, m.PathKind, ""), Value: value}
}

func infoToCEL(m model.Generator) Metric {
	labels := labelsToCEL(m.Labels, m.Path, m.PathKind, "")

	if isIterated(m.Path, m.PathKind) {
		e := newElements(m.Path, m.PathKind)
		if m.KeyLabel != "" && m.PathKind == model.PathKindMap {
			labels = append(labels, Label{Name: m.KeyLabel, Value: e.mapCEL("", "k")})
		}
		return Metric{Labels: labels, Value: e.mapCEL("", "1")}
	}

	value := "1"
	if len(m.Path) > 0 {
		guard, _ := guardedPathCEL("o", m.Path)
		value = fmt.Sprintf("%s ? [1] : []", guard)
	}
	return Metric{Labels: labels, Value: value}
}

// stateSetToCEL returns a metric for each state, which has the value 1 if the field is
// equal to the state and 0 if not.
func stateSetToCEL(m model.Generator) []Metric {
	metrics := make([]Metric, 0, len(m.List))
	for _, state := range m.List {
		quoted := strconv.Quote(state)

		if isIterated(m.Path, m.PathKind) {
			e := newElements(m.Path, m.PathKind)
			filter := e.valueFilter(m.Value)
			_, actual := guardedPathCEL(e.element(), m.Value)

			labels := labelsToCEL(m.Labels, m.Path, m.PathKind, filter)
			labels = append(labels, Label{Name: m.LabelName, Value: quoted})
			metrics = append(metrics, Metric{
				Labels: labels,
				Value:  e.mapCEL(filter, fmt.Sprintf("int(%s == %s ? 1 : 0)", actual, quoted)),
			})
			continue
		}

		labels := labelsToCEL(m.Labels, m.Path, m.PathKind, "")
		labels = append(labels, Label{Name: m.LabelName, Value: quoted})
		actualPath := append(slices.Clone(m.Path), m.Value...)
		guard, actual := guardedPathCEL("o", actualPath)
		metrics = append(metrics, Metric{
			Labels: labels,
			Value:  fmt.Sprintf("%s ? [int(%s == %s ? 1 : 0)] : []", guard, actual, quoted),
		})
	}
	return metrics
}

// labelsToCEL converts the labels. The path of the labels is relative to the path, which
// has to be iterated if it is a list or map. The filter has to be the one of the value of the metric.
func labelsToCEL(labels []model.Label, path model.Path, kind model.PathKind, filter string) []Label {
	result := make([]Label, 0, len(labels))
	for _, label := range labels {
		// resource-state-metrics injects these labels into every sample.
		if isAutoLabel(label.Name) {
			continue
		}

		var value string
		if isIterated(path, kind) {
			e := newElements(path, kind)
			guard, resolved := guardedPathCEL(e.element(), label.Path)
			value = e.mapCEL(filter, fmt.Sprintf("%s ? string(%s) : %s", guard, resolved, strconv.Quote("")))
		} else {
			guard, resolved := guardedPathCEL("o", append(slices.Clone(path), label.Path...))
			value = fmt.Sprintf("%s ? string(%s) : %s", guard, resolved, strconv.Quote(""))
		}
		result = append(result, Label{Name: label.Name, Value: value})
	}
	return result
}

// isAutoLabel returns true for the labels resource-state-metrics adds to every sample.
func isAutoLabel(name string) bool {
	switch name {
	case "group", "version", "kind", "name", "namespace":
		return true
	default:
		return false
	}
}

func pathCEL(root string, path model.Path) string {
	_, value := guardedPathCEL(root, path)
	return value
}

// timestampCEL converts a timestamp to unix seconds. The conversion is not done based on the
// name of a field, as it would fail for numeric fields.
func timestampCEL(expr string, timestamp bool) string {
	if timestamp {
		return "unixSeconds(" + expr + ")"
	}
	return expr
}

var celIdentifier = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// celKeywords are the CEL keywords which are valid identifiers by their form but can't be used
// as field name in a selection like `o.in`.
var celKeywords = []string{"in", "true", "false", "null"}

// guardedPathCEL returns a presence guard and the CEL expression for the path.
func guardedPathCEL(root string, path model.Path) (string, string) {
	expr := root
	guards := make([]string, 0, len(path))
	for _, part := range path {
		parent := expr
		if celIdentifier.MatchString(part) && !slices.Contains(celKeywords, part) {
			expr += "." + part
			guards = append(guards, "has("+expr+")")
		} else {
			expr += "[" + strconv.Quote(part) + "]"
			guards = append(guards, fmt.Sprintf("%s in %s", strconv.Quote(part), parent))
		}
	}
	if len(guards) == 0 {
		return "true", expr
	}
	return strings.Join(guards, " && "), expr
}
