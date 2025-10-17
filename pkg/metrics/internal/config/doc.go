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

// Package config contains a copy of the types from k8s.io/kube-state-metrics/pkg/customresourcestate.
// The following modifications got applied:
// For `config.go`:
// * Rename the package to `config`.
// * Drop `const customResourceState`.
// * Drop all functions, only preserve structs.
// * Use `int32` instead of `klog.Level`.
// * Use `MetricType` instead of `metric.Type`
// * Add `omitempty` to:
//   - `Labels.CommonLabels`
//   - `Labels.LabelsFromPath`
//   - `Generator.ErrorLogV`
//   - `Metric.Gauge`
//   - `Metric.StateSet`
//   - `Metric.Info`
//
// For `config_metrics_types.go`:
// * Rename the package to `config`.
// * Add `omitempty` to:
//   - `MetricMeta.LabelsFromPath
//   - `MetricGauge.LabelFromkey`
//   - `MetricInfo.LabelFromkey`
package config

// KubeStateMetricsVersion defines which version of kube-state-metrics these types
// are based on and the output file should be compatible to.
const KubeStateMetricsVersion = "v2.13.0"
