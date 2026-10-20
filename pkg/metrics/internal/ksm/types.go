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

// Package ksm contains the output builder for kube-state-metrics and a copy of the types
// describing the custom resource state configuration of kube-state-metrics:
// https://github.com/kubernetes/kube-state-metrics/blob/v2.20.0/pkg/customresourcestate/config.go
// https://github.com/kubernetes/kube-state-metrics/blob/v2.20.0/pkg/customresourcestate/config_metrics_types.go
//
// The following modifications got applied:
//   - Rename the package to `ksm` and merge both files into this one.
//   - Rename `Metrics` to `CustomResourceStateMetrics` (the kind used in the configuration file)
//     and `MetricsSpec` to `CustomResourceStateMetricsSpec`.
//   - Drop `const customResourceState`, `ConfigDecoder` and all functions, only preserve structs.
//   - Use `int32` instead of `klog.Level` and `MetricType` instead of `metric.Type`.
//   - Drop the `yaml` struct tags, the output gets written using the `json` struct tags.
//   - Add `omitempty` to all fields which don't have to be set, to not render zero values.
//   - Replace the `+union` markers with a comment.
package ksm

import (
	"fmt"
)

// CustomResourceStateMetrics is the top level configuration object.
type CustomResourceStateMetrics struct {
	Spec CustomResourceStateMetricsSpec `json:"spec"`
}

// CustomResourceStateMetricsSpec is the configuration describing the custom resource state metrics to generate.
type CustomResourceStateMetricsSpec struct {
	// Resources is the list of custom resources to be monitored. A resource with the same GroupVersionKind may appear
	// multiple times (e.g. to use different metric name prefixes) but will incur additional overhead.
	Resources []Resource `json:"resources"`
}

// Resource configures a custom resource for metric generation.
type Resource struct {
	// MetricNamePrefix defines a prefix for all metrics of the resource.
	// If set to "", no prefix will be added.
	// Example: If set to "foo", MetricNamePrefix will be "foo_<metric>".
	MetricNamePrefix *string `json:"metricNamePrefix,omitempty"`

	// GroupVersionKind of the custom resource to be monitored.
	GroupVersionKind GroupVersionKind `json:"groupVersionKind"`

	// Labels are added to all metrics. If the same key is used in a metric, the value from the metric will overwrite the value here.
	Labels

	// Metrics are the custom resource fields to be collected.
	Metrics []Generator `json:"metrics"`

	// ErrorLogV defines the verbosity threshold for errors logged for this resource.
	ErrorLogV int32 `json:"errorLogV,omitempty"`

	// ResourcePlural sets the plural name of the resource. Defaults to the plural version of the Kind according to flect.Pluralize.
	ResourcePlural string `json:"resourcePlural,omitempty"`
}

// GroupVersionKind is the Kubernetes group, version, and kind of a resource.
type GroupVersionKind struct {
	Group   string `json:"group"`
	Version string `json:"version"`
	Kind    string `json:"kind"`
}

func (gvk GroupVersionKind) String() string {
	return fmt.Sprintf("%s_%s_%s", gvk.Group, gvk.Version, gvk.Kind)
}

// Labels is common configuration of labels to add to metrics.
type Labels struct {
	// CommonLabels are added to all metrics.
	CommonLabels map[string]string `json:"commonLabels,omitempty"`

	// LabelsFromPath adds additional labels where the value is taken from a field in the resource.
	LabelsFromPath map[string][]string `json:"labelsFromPath,omitempty"`
}

// Generator describes a unique metric name.
type Generator struct {
	// Name of the metric. Subject to prefixing based on the configuration of the Resource.
	Name string `json:"name"`

	// Help text for the metric.
	Help string `json:"help"`

	// Each targets a value or values from the resource.
	Each Metric `json:"each"`

	// Labels are added to all metrics. Labels from Each will overwrite these if using the same key.
	Labels

	// ErrorLogV defines the verbosity threshold for errors logged for this metric. Must be non-zero to override the resource setting.
	ErrorLogV int32 `json:"errorLogV,omitempty"`
}

// Metric defines a metric to expose. Exactly one of Gauge, StateSet and Info has to be set, matching Type.
type Metric struct {
	// Type defines the type of the metric.
	Type MetricType `json:"type"`

	// Gauge defines a gauge metric.
	Gauge *MetricGauge `json:"gauge,omitempty"`

	// StateSet defines a state set metric.
	StateSet *MetricStateSet `json:"stateSet,omitempty"`

	// Info defines an info metric.
	Info *MetricInfo `json:"info,omitempty"`
}

// Version defines which version of kube-state-metrics these types
// are based on and the output file should be compatible to.
const Version = "v2.20.0"

// MetricType is the type of a metric.
type MetricType string

// Supported metric types.
const (
	MetricTypeGauge    MetricType = "Gauge"
	MetricTypeStateSet MetricType = "StateSet"
	MetricTypeInfo     MetricType = "Info"
)

// MetricMeta are variables which may used for any metric type.
type MetricMeta struct {
	// LabelsFromPath adds additional labels where the value of the label is taken from a field under Path.
	LabelsFromPath map[string][]string `json:"labelsFromPath,omitempty"`

	// Path is the path to to generate metric(s) for.
	Path []string `json:"path"`
}

// MetricGauge targets a Path that may be a single value, array, or object. Arrays and objects will generate a metric per element.
// Ref: https://github.com/prometheus/OpenMetrics/blob/v1.0.0/specification/OpenMetrics.md#gauge
type MetricGauge struct {
	MetricMeta

	// ValueFrom is the path to a numeric field under Path that will be the metric value.
	ValueFrom []string `json:"valueFrom,omitempty"`

	// LabelFromKey adds a label with the given name if Path is an object. The label value will be the object key.
	LabelFromKey string `json:"labelFromKey,omitempty"`

	// NilIsZero indicates that if a value is nil it will be treated as zero value.
	NilIsZero bool `json:"nilIsZero,omitempty"`
}

// MetricInfo is a metric which is used to expose textual information.
// Ref: https://github.com/prometheus/OpenMetrics/blob/v1.0.0/specification/OpenMetrics.md#info
type MetricInfo struct {
	MetricMeta

	// LabelFromKey adds a label with the given name if Path is an object. The label value will be the object key.
	LabelFromKey string `json:"labelFromKey,omitempty"`
}

// MetricStateSet is a metric which represent a series of related boolean values, also called a bitset.
// Ref: https://github.com/prometheus/OpenMetrics/blob/v1.0.0/specification/OpenMetrics.md#stateset
type MetricStateSet struct {
	MetricMeta

	// List is the list of values to expose a value for.
	List []string `json:"list,omitempty"`

	// LabelName is the key of the label which is used for each entry in List to expose the value.
	LabelName string `json:"labelName,omitempty"`

	// ValueFrom is the subpath to compare the list to.
	ValueFrom []string `json:"valueFrom,omitempty"`
}
