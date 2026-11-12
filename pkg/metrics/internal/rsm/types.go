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

// Package rsm contains the types of the ResourceMetricsMonitor of resource-state-metrics,
// the translation of the metrics model to it and the output builder creating it.
package rsm

// Version is the version of resource-state-metrics whose API is represented by the types below.
// They are a subset of
// https://github.com/kubernetes-sigs/resource-state-metrics/blob/v0.1.0/pkg/apis/resourcestatemetrics/v1alpha1/types.go
const Version = "v0.1.0"

// ResourceMetricsMonitor is the resource-state-metrics configuration resource.
type ResourceMetricsMonitor struct {
	APIVersion string `json:"apiVersion"`

	Kind string `json:"kind"`

	Metadata ResourceMetricsMonitorMeta `json:"metadata"`

	Spec ResourceMetricsMonitorSpec `json:"spec"`
}

// ResourceMetricsMonitorMeta is the metadata of a ResourceMetricsMonitor.
type ResourceMetricsMonitorMeta struct {
	Name string `json:"name"`

	Namespace string `json:"namespace,omitempty"`
}

// ResourceMetricsMonitorSpec is the spec of a ResourceMetricsMonitor.
type ResourceMetricsMonitorSpec struct {
	Configuration Configuration `json:"configuration"`
}

// Configuration is the configuration of the stores to generate metrics for.
type Configuration struct {
	Stores []Store `json:"stores"`
}

// Store configures the metrics generated for a custom resource.
type Store struct {
	Group string `json:"group"`

	Version string `json:"version"`

	Kind string `json:"kind"`

	// Resource is the plural name of the resource.
	Resource string `json:"resource"`

	// Resolver is the resolver used to evaluate the expressions of the store.
	Resolver string `json:"resolver"`

	// Labels are added to all metrics of the store.
	Labels []Label `json:"labels,omitempty"`

	Families []Family `json:"families"`
}

// Family is a metric family.
type Family struct {
	Name string `json:"name"`

	Help string `json:"help"`

	Metrics []Metric `json:"metrics"`
}

// Metric is a single time series, or a list of them, within a family.
type Metric struct {
	Labels []Label `json:"labels,omitempty"`

	// Value is the expression evaluating to the value(s) of the metric.
	Value string `json:"value"`
}

// Label is a label with an expression evaluating to its value.
type Label struct {
	Name string `json:"name"`

	// Value is the expression evaluating to the value(s) of the label.
	Value string `json:"value"`
}
