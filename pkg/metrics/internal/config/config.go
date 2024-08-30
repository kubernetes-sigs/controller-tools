/*
Copyright 2021 The Kubernetes Authors All rights reserved.

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

package config

import (
	"fmt"
)

// Metrics is the top level configuration object.
type Metrics struct {
	Spec MetricsSpec `yaml:"spec" json:"spec"`
}

// MetricsSpec is the configuration describing the custom resource state metrics to generate.
type MetricsSpec struct {
	// Resources is the list of custom resources to be monitored. A resource with the same GroupVersionKind may appear
	// multiple times (e.g., to customize the namespace or subsystem,) but will incur additional overhead.
	Resources []Resource `yaml:"resources" json:"resources"`
}

// Resource configures a custom resource for metric generation.
type Resource struct {
	// MetricNamePrefix defines a prefix for all metrics of the resource.
	// If set to "", no prefix will be added.
	// Example: If set to "foo", MetricNamePrefix will be "foo_<metric>".
	MetricNamePrefix *string `yaml:"metricNamePrefix" json:"metricNamePrefix"`

	// GroupVersionKind of the custom resource to be monitored.
	GroupVersionKind GroupVersionKind `yaml:"groupVersionKind" json:"groupVersionKind"`

	// Labels are added to all metrics. If the same key is used in a metric, the value from the metric will overwrite the value here.
	Labels `yaml:",inline" json:",inline"`

	// Metrics are the custom resource fields to be collected.
	Metrics []Generator `yaml:"metrics" json:"metrics"`
	// ErrorLogV defines the verbosity threshold for errors logged for this resource.
	ErrorLogV int32 `yaml:"errorLogV" json:"errorLogV"`

	// ResourcePlural sets the plural name of the resource. Defaults to the plural version of the Kind according to flect.Pluralize.
	ResourcePlural string `yaml:"resourcePlural" json:"resourcePlural"`
}

// GroupVersionKind is the Kubernetes group, version, and kind of a resource.
type GroupVersionKind struct {
	Group   string `yaml:"group" json:"group"`
	Version string `yaml:"version" json:"version"`
	Kind    string `yaml:"kind" json:"kind"`
}

func (gvk GroupVersionKind) String() string {
	return fmt.Sprintf("%s_%s_%s", gvk.Group, gvk.Version, gvk.Kind)
}

// Labels is common configuration of labels to add to metrics.
type Labels struct {
	// CommonLabels are added to all metrics.
	CommonLabels map[string]string `yaml:"commonLabels,omitempty" json:"commonLabels,omitempty"`
	// LabelsFromPath adds additional labels where the value is taken from a field in the resource.
	LabelsFromPath map[string][]string `yaml:"labelsFromPath,omitempty" json:"labelsFromPath,omitempty"`
}

// Generator describes a unique metric name.
type Generator struct {
	// Name of the metric. Subject to prefixing based on the configuration of the Resource.
	Name string `yaml:"name" json:"name"`
	// Help text for the metric.
	Help string `yaml:"help" json:"help"`
	// Each targets a value or values from the resource.
	Each Metric `yaml:"each" json:"each"`

	// Labels are added to all metrics. Labels from Each will overwrite these if using the same key.
	Labels `yaml:",inline" json:",inline"` // json will inline because it is already tagged
	// ErrorLogV defines the verbosity threshold for errors logged for this metric. Must be non-zero to override the resource setting.
	ErrorLogV int32 `yaml:"errorLogV,omitempty" json:"errorLogV,omitempty"`
}

// Metric defines a metric to expose.
// +union
type Metric struct {
	// Type defines the type of the metric.
	// +unionDiscriminator
	Type MetricType `yaml:"type" json:"type"`

	// Gauge defines a gauge metric.
	// +optional
	Gauge *MetricGauge `yaml:"gauge,omitempty" json:"gauge,omitempty"`
	// StateSet defines a state set metric.
	// +optional
	StateSet *MetricStateSet `yaml:"stateSet,omitempty" json:"stateSet,omitempty"`
	// Info defines an info metric.
	// +optional
	Info *MetricInfo `yaml:"info,omitempty" json:"info,omitempty"`
}

// ConfigDecoder is for use with FromConfig.
type ConfigDecoder interface {
	Decode(v interface{}) (err error)
}
