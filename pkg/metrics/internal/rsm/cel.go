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

const (
	// resolverCEL is the resolver of resource-state-metrics which evaluates CEL expressions.
	resolverCEL = "cel"

	// noGuard is the guard of a path which doesn't have to be checked for presence.
	noGuard = "true"
)

// toStore converts a resource to a resource-state-metrics store. All expressions are
// CEL expressions, as the store uses the CEL resolver. The object is available as `o`.
// The name of every family is prefixed with a prefix derived from the kind (see namePrefix).
func toStore(resource model.Resource) Store {
	// The labels of the resource usually point to optional fields (e.g. labels or annotations), which
	// would fail the evaluation of the label, so they are always guarded.
	resourceLabels := labelsToCEL(resource.Labels, nil, model.PathKindScalar, "", true)
	store := Store{
		Group:    resource.GroupVersionKind.Group,
		Version:  resource.GroupVersionKind.Version,
		Kind:     resource.GroupVersionKind.Kind,
		Resource: resource.Plural(),
		Resolver: resolverCEL,
	}

	prefix := namePrefix(resource)
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
		// Ref: https://github.com/kubernetes-sigs/resource-state-metrics/issues/81
		for i := range family.Metrics {
			family.Metrics[i].Labels = append(family.Metrics[i].Labels, resourceLabels...)
		}
		store.Families = append(store.Families, family)
	}
	return store
}

// namePrefix returns the prefix for the names of the families of the resource, including the
// separating underscore. It is the lowercased kind, e.g. machinedeployment_ for MachineDeployment,
// following the naming of the metrics of kube-state-metrics, e.g. kube_horizontalpodautoscaler_.
// resource-state-metrics prefixes the names with kube_customresource_ itself, which results in e.g.
// kube_customresource_machinedeployment_created. The prefix keeps the families of different resources
// apart, as resource-state-metrics exposes duplicate HELP and TYPE lines for a family name used by
// multiple stores. The name prefix of the resource is ignored, it is only used by kube-state-metrics.
func namePrefix(resource model.Resource) string {
	return strings.ToLower(resource.GroupVersionKind.Kind) + "_"
}

// elements describes the elements of the list or map a metric is generated for. A list is
// iterated by its elements (`v`), a map by its keys (`k`) with the element being `base[k]`.
type elements struct {
	// guard checks the presence of the list or map. It is `true` if the presence is not checked.
	guard string
	// base is the expression of the list or map.
	base string
	kind model.PathKind
}

func newElements(path model.Path, kind model.PathKind, ignoreMissing bool) elements {
	guard, base := guardedPathCEL("o", path, ignoreMissing)
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
	return ifGuard(e.guard, fmt.Sprintf("%s.map(%s, %s)", items, e.variable(), expr), "[]")
}

// valueFilter returns the filter for the elements having the value, which is empty if
// missing fields are not ignored. A list or map entry without the value would otherwise
// fail the evaluation of the whole metric.
func (e elements) valueFilter(value model.Path, ignoreMissing bool) string {
	guard, _ := guardedPathCEL(e.element(), value, ignoreMissing)
	if guard == noGuard {
		return ""
	}
	return guard
}

// ifGuard returns the conditional expression `guard ? then : otherwise`, or then if
// there is nothing to guard.
func ifGuard(guard, then, otherwise string) string {
	if guard == noGuard {
		return then
	}
	return fmt.Sprintf("%s ? %s : %s", guard, then, otherwise)
}

// gaugeToCEL returns the metric for a gauge. Example of the resulting value expressions:
//
//	scalar `spec.replicas`:                           o.spec.replicas
//	with ignoreMissing:                               has(o.spec) && has(o.spec.replicas) ? [o.spec.replicas] : []
//	with missingAsZero:                               has(o.spec) && has(o.spec.replicas) ? o.spec.replicas : 0
//	list `status.items`, value `size`:                o.status.items.map(v, v.size)
//	with ignoreMissing:                               has(o.status) && has(o.status.items) ? o.status.items.filter(v, has(v.size)).map(v, v.size) : []
//
// The expressions evaluate to a number or a list, with an element for each sample of the metric.
// Without ignoreMissing a missing field fails the evaluation, for which resource-state-metrics doesn't
// expose the sample. With ignoreMissing a scalar which is not set evaluates to an empty list instead.
func gaugeToCEL(m model.Generator) Metric {
	timestamp := m.ValueKind == model.ValueKindTimestamp

	if isIterated(m.Path, m.PathKind) {
		e := newElements(m.Path, m.PathKind, m.IgnoreMissing)
		// missingAsZero needs the guard independent of ignoreMissing.
		valueGuard, valueExpr := guardedPathCEL(e.element(), m.Value, m.IgnoreMissing || m.MissingAsZero)
		valueExpr = timestampCEL(valueExpr, timestamp)

		// Elements without the value are exposed as zero or left out, if not set.
		// The filter is empty if there is no value path (m.Value is empty). In that case the elements
		// themselves are the values, which are always present and don't have to be filtered.
		filter := ""
		value := valueExpr
		if len(m.Value) > 0 {
			if m.MissingAsZero {
				value = ifGuard(valueGuard, valueExpr, "0")
			} else if m.IgnoreMissing {
				filter = valueGuard
			}
		}

		labels := labelsToCEL(m.Labels, m.Path, m.PathKind, filter, m.IgnoreMissing)
		return Metric{Labels: labels, Value: e.mapCEL(filter, value)}
	}

	valuePath := append(slices.Clone(m.Path), m.Value...)
	// missingAsZero needs the guard independent of ignoreMissing.
	guard, valueExpr := guardedPathCEL("o", valuePath, m.IgnoreMissing || m.MissingAsZero)
	value := timestampCEL(valueExpr, timestamp)
	if m.MissingAsZero {
		value = ifGuard(guard, value, "0")
	} else if m.IgnoreMissing {
		value = ifGuard(guard, "["+value+"]", "[]")
	}
	return Metric{Labels: labelsToCEL(m.Labels, m.Path, m.PathKind, "", m.IgnoreMissing), Value: value}
}

// infoToCEL returns the metric for an info. An info metric always has the value 1, the information
// is exposed by the labels. Example of the resulting value expressions:
//
//	no path:                            1
//	object `status.nodeRef`:            size([o.status.nodeRef])
//	with ignoreMissing:                 has(o.status) && has(o.status.nodeRef) ? [1] : []
//	list `metadata.ownerReferences`:    o.metadata.ownerReferences.map(v, 1)
//	with ignoreMissing:                 has(o.metadata) && has(o.metadata.ownerReferences) ? o.metadata.ownerReferences.map(v, 1) : []
//
// For a list or map one sample with the value 1 is exposed for each element.
func infoToCEL(m model.Generator) Metric {
	labels := labelsToCEL(m.Labels, m.Path, m.PathKind, "", m.IgnoreMissing)

	if isIterated(m.Path, m.PathKind) {
		e := newElements(m.Path, m.PathKind, m.IgnoreMissing)
		return Metric{Labels: labels, Value: e.mapCEL("", "1")}
	}

	value := "1"
	if len(m.Path) > 0 {
		guard, expr := guardedPathCEL("o", m.Path, m.IgnoreMissing)
		// Without the guard the value has to access the path like the labels do, to fail
		// together with them. The size of a list with the single element is 1.
		value = ifGuard(guard, "[1]", "[]")
		if guard == noGuard {
			value = fmt.Sprintf("size([%s])", expr)
		}
	}
	return Metric{Labels: labels, Value: value}
}

// stateSetToCEL returns a metric for each state, which has the value 1 if the field is
// equal to the state and 0 if not. The state is exposed as label (labelName). Example for the
// states "True" and "False" and the labelName "status", which results in two metrics:
//
//	foo{status="True"}  1
//	foo{status="False"} 0
//
// for an object having `status.phase: "True"`. The value expression of the first one is:
//
//	int(o.status.phase == "True" ? 1 : 0)
//
// With ignoreMissing it is guarded against the missing field:
//
//	has(o.status) && has(o.status.phase) ? [int(o.status.phase == "True" ? 1 : 0)] : []
func stateSetToCEL(m model.Generator) []Metric {
	metrics := make([]Metric, 0, len(m.List))
	for _, state := range m.List {
		quoted := strconv.Quote(state)

		if isIterated(m.Path, m.PathKind) {
			e := newElements(m.Path, m.PathKind, m.IgnoreMissing)
			filter := e.valueFilter(m.Value, m.IgnoreMissing)
			_, actual := guardedPathCEL(e.element(), m.Value, false)

			labels := labelsToCEL(m.Labels, m.Path, m.PathKind, filter, m.IgnoreMissing)
			labels = append(labels, Label{Name: m.LabelName, Value: quoted})
			metrics = append(metrics, Metric{
				Labels: labels,
				Value:  e.mapCEL(filter, fmt.Sprintf("int(%s == %s ? 1 : 0)", actual, quoted)),
			})
			continue
		}

		labels := labelsToCEL(m.Labels, m.Path, m.PathKind, "", m.IgnoreMissing)
		labels = append(labels, Label{Name: m.LabelName, Value: quoted})
		actualPath := append(slices.Clone(m.Path), m.Value...)
		guard, actual := guardedPathCEL("o", actualPath, m.IgnoreMissing)
		value := fmt.Sprintf("int(%s == %s ? 1 : 0)", actual, quoted)
		if guard != noGuard {
			value = fmt.Sprintf("%s ? [%s] : []", guard, value)
		}
		metrics = append(metrics, Metric{Labels: labels, Value: value})
	}
	return metrics
}

// labelsToCEL converts the labels. The path of the labels is relative to the path, which
// has to be iterated if it is a list or map. The filter has to be the one of the value of the metric.
// If missing fields are not ignored, a missing field fails the evaluation of the label, so the value of the
// metric has to fail with it. Otherwise resource-state-metrics exposes the expression as value of the label.
func labelsToCEL(labels []model.Label, path model.Path, kind model.PathKind, filter string, ignoreMissing bool) []Label {
	result := make([]Label, 0, len(labels))
	for _, label := range labels {
		var value string
		if isIterated(path, kind) {
			e := newElements(path, kind, ignoreMissing)
			guard, resolved := guardedPathCEL(e.element(), label.Path, ignoreMissing)
			value = e.mapCEL(filter, ifGuard(guard, labelValueCEL(label, resolved), strconv.Quote("")))
		} else {
			guard, resolved := guardedPathCEL("o", append(slices.Clone(path), label.Path...), ignoreMissing)
			value = ifGuard(guard, labelValueCEL(label, resolved), strconv.Quote(""))
		}
		result = append(result, Label{Name: label.Name, Value: value})
	}
	return result
}

// labelValueCEL returns the expression of the value of a label, which has to be a string. A field
// which is not a string, e.g. a bool or a number, is converted.
func labelValueCEL(label model.Label, expr string) string {
	if label.IsString {
		return expr
	}
	return "string(" + expr + ")"
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

// guardedPathCEL returns a presence guard and the CEL expression for the path. The guard
// is `true` for an empty path or if guard is false.
// Example: for the root `o` and the path "spec", "replicas" it returns the guard
// `has(o.spec) && has(o.spec.replicas)` and the expression `o.spec.replicas`.
func guardedPathCEL(root string, path model.Path, guard bool) (string, string) {
	expr := root
	guards := make([]string, 0, len(path))
	for _, part := range path {
		parent := expr
		if celIdentifier.MatchString(part) && !slices.Contains(celKeywords, part) {
			expr += "." + part
			guards = append(guards, "has("+expr+")")
		} else {
			// Fields which are not valid CEL identifiers or are CEL keywords can't be selected
			// with `.`, so they are accessed by index and checked with the `in` operator, as has()
			// only supports field selections.
			// Example: for the path "metadata", "labels", "cluster.x-k8s.io/name":
			//   has(o.metadata) && has(o.metadata.labels) && "cluster.x-k8s.io/name" in o.metadata.labels
			// resulting in the expression o.metadata.labels["cluster.x-k8s.io/name"].
			expr += "[" + strconv.Quote(part) + "]"
			guards = append(guards, fmt.Sprintf("%s in %s", strconv.Quote(part), parent))
		}
	}
	if !guard || len(guards) == 0 {
		return noGuard, expr
	}
	return strings.Join(guards, " && "), expr
}
