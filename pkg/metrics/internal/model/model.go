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

// Package model contains the model the metrics markers get parsed to. It is independent
// of the configuration format of the targets, which create their configuration from it.
package model

import (
	"fmt"
	"strings"

	"github.com/gobuffalo/flect"
)

// Resource are the generators of the metrics of a custom resource.
type Resource struct {
	GroupVersionKind GroupVersionKind

	// NamePrefix is the prefix for the names of all metrics of the resource.
	// If nil, the target uses its default prefix. If set to "", no prefix will be added.
	// Example: If set to "foo", the name of the metric "bar" will be "foo_bar".
	NamePrefix *string

	// Labels are added to all metrics of the resource.
	// The path of the labels is relative to the custom resource.
	Labels []Label

	Metrics []Generator
}

// Plural returns the lowercase plural name of the resource.
func (r Resource) Plural() string {
	return strings.ToLower(flect.Pluralize(r.GroupVersionKind.Kind))
}

// GroupVersionKind is the Kubernetes group, version, and kind of a resource.
type GroupVersionKind struct {
	Group   string
	Version string
	Kind    string
}

func (gvk GroupVersionKind) String() string {
	return fmt.Sprintf("%s_%s_%s", gvk.Group, gvk.Version, gvk.Kind)
}

// Path is a path to a field in a custom resource, as a list of field names.
type Path []string

// Label is a label of a metric, with the value taken from a field.
type Label struct {
	Name string

	// Path is the path to the field containing the value of the label.
	// An empty path means the value of the field the label belongs to.
	Path Path
}

// MetricType is the type of a metric as defined by OpenMetrics.
type MetricType string

// Supported metric types.
const (
	// MetricTypeGauge is a metric with a numeric value.
	// Ref: https://github.com/OpenObservability/OpenMetrics/blob/main/specification/OpenMetrics.md#gauge
	MetricTypeGauge MetricType = "Gauge"

	// MetricTypeInfo is a metric which is used to expose textual information.
	// Ref: https://github.com/OpenObservability/OpenMetrics/blob/main/specification/OpenMetrics.md#info
	MetricTypeInfo MetricType = "Info"

	// MetricTypeStateSet is a metric which represents a series of related boolean values, also called a bitset.
	// Ref: https://github.com/OpenObservability/OpenMetrics/blob/main/specification/OpenMetrics.md#stateset
	MetricTypeStateSet MetricType = "StateSet"
)

// Generator describes a metric of a custom resource. The targets use it to
// generate the metric in the format of their configuration.
type Generator struct {
	Name string

	Help string

	Type MetricType

	// Path is the path to the field(s) to generate the metric for. The field may be a
	// single value, a list or a map. Lists and maps generate a metric per element.
	Path Path

	// Labels are additional labels of the metric. The path of the labels is relative to
	// Path, which means relative to the element if Path is a list or map.
	Labels []Label

	// KeyLabel is the name of a label which gets the key of the element as value, if Path is a map.
	// Only used by Gauge and Info.
	KeyLabel string

	// Value is the path to the field containing the value of the metric (Gauge) or
	// the field compared to the states (StateSet). It is relative to Path.
	// Not used by Info.
	Value Path

	// MissingAsZero exposes a missing value as zero instead of not exposing the metric.
	// Only used by Gauge.
	MissingAsZero bool

	// List are the states to expose a value for. Only used by StateSet.
	List []string

	// LabelName is the name of the label which contains the state. Only used by StateSet.
	LabelName string

	// PathKind is the kind of the field at Path. It is derived from the Go type and
	// is needed by targets which have to know how to iterate over the path.
	PathKind PathKind

	// ValueKind describes how the field at Value has to be converted. It is derived
	// from the Go type and is needed by targets which have to convert the value.
	ValueKind ValueKind
}

// PathKind describes the kind of the field a path points to.
type PathKind string

// Supported path kinds.
const (
	// PathKindScalar is a single value.
	PathKindScalar PathKind = "scalar"

	// PathKindObject is a struct.
	PathKindObject PathKind = "object"

	// PathKindMap is a map. A metric is generated for each entry.
	PathKindMap PathKind = "map"

	// PathKindArray is a list.
	PathKindArray PathKind = "array"
)

// ValueKind describes special conversion of a metric value.
type ValueKind string

// Supported value kinds.
const (
	// ValueKindDefault is a value which does not need a conversion.
	ValueKindDefault ValueKind = ""

	// ValueKindTimestamp is a timestamp which has to be converted to a number.
	ValueKindTimestamp ValueKind = "timestamp"
)
