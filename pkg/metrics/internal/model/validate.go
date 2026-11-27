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

package model

import (
	"fmt"
	"regexp"
	"slices"
)

// reservedLabelNames are the names of the labels which the targets add to every metric. resource-state-metrics
// adds all of them, kube-state-metrics adds name and namespace. They can't be used as label names, to have
// the same markers work for all targets.
var reservedLabelNames = []string{"group", "version", "kind", "name", "namespace"}

// metricNamePattern is the pattern for the name of a metric. The targets add a prefix to the name,
// which is why it may start with a digit. resource-state-metrics rejects other names and
// they are not valid for Prometheus either.
var metricNamePattern = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

// Validate returns an error if the resource has metrics which are not supported by the targets.
// All checks are done here and for all targets, also the ones only required by one of them, to
// have the same markers either work for all targets or for none.
func (r Resource) Validate() error {
	for _, label := range r.Labels {
		if err := validateLabelName(r.GroupVersionKind, "", "label", label.Name); err != nil {
			return err
		}
	}
	for _, generator := range r.Metrics {
		if err := generator.validateNameAndHelp(r.GroupVersionKind); err != nil {
			return err
		}
		if err := generator.validateReservedLabels(r.GroupVersionKind); err != nil {
			return err
		}
		if err := generator.validateUniqueLabels(r.GroupVersionKind, r.Labels); err != nil {
			return err
		}
		if err := generator.validateMapLabels(r.GroupVersionKind); err != nil {
			return err
		}
		if generator.MissingAsZero && generator.IgnoreMissing {
			return fmt.Errorf("metric %q of %s: missingAsZero and ignoreMissing are mutually exclusive", generator.Name, r.GroupVersionKind)
		}
	}
	return nil
}

// validateLabelName rejects the names of the labels which are added to every metric by the targets.
// option is the name of the marker option the label is created from. metric is empty for the labels of the resource.
func validateLabelName(gvk GroupVersionKind, metric, option, name string) error {
	if !slices.Contains(reservedLabelNames, name) {
		return nil
	}
	if metric == "" {
		return fmt.Errorf("%s: the label %q of the option %s is added to every metric by the targets, use a different name", gvk, name, option)
	}
	return fmt.Errorf("metric %q of %s: the label %q of the option %s is added to every metric by the targets, use a different name", metric, gvk, name, option)
}

// validateNameAndHelp rejects metrics with a name which is not a valid metric name and metrics without
// help. kube-state-metrics accepts both, resource-state-metrics rejects the configuration.
func (g Generator) validateNameAndHelp(gvk GroupVersionKind) error {
	if !metricNamePattern.MatchString(g.Name) {
		return fmt.Errorf("metric %q of %s: the name has to match %s", g.Name, gvk, metricNamePattern)
	}
	if g.Help == "" {
		return fmt.Errorf("metric %q of %s: the help has to be set", g.Name, gvk)
	}
	return nil
}

// validateUniqueLabels rejects empty and duplicate label names of a metric. The labels of the
// resource get exposed on all of its metrics, so they have to be unique together with the labels
// of the metric. resource-state-metrics rejects duplicates, kube-state-metrics silently
// exposes only one of them.
func (g Generator) validateUniqueLabels(gvk GroupVersionKind, resourceLabels []Label) error {
	names := map[string]struct{}{}
	// option is the name of the marker option the label is created from.
	addUnique := func(option, name string) error {
		if name == "" {
			return fmt.Errorf("metric %q of %s: the name of the label of the option %s must not be empty", g.Name, gvk, option)
		}
		if _, exists := names[name]; exists {
			return fmt.Errorf("metric %q of %s: the label %q is set multiple times, the label names have to be unique", g.Name, gvk, name)
		}
		names[name] = struct{}{}
		return nil
	}

	for _, label := range resourceLabels {
		if err := addUnique("label", label.Name); err != nil {
			return err
		}
	}
	for _, label := range g.Labels {
		if err := addUnique("labels", label.Name); err != nil {
			return err
		}
	}
	if g.Type == MetricTypeStateSet {
		return addUnique("labelName", g.LabelName)
	}
	return nil
}

// validateReservedLabels rejects labels having the name of a label which is added by the targets.
func (g Generator) validateReservedLabels(gvk GroupVersionKind) error {
	for _, label := range g.Labels {
		if err := validateLabelName(gvk, g.Name, "labels", label.Name); err != nil {
			return err
		}
	}
	if g.Type == MetricTypeStateSet {
		return validateLabelName(gvk, g.Name, "labelName", g.LabelName)
	}
	return nil
}

// validateMapLabels rejects labels for the entries of a map which can't be aligned with each
// other and with the value. Some targets (e.g. resource-state-metrics) evaluate each label and
// the value separately and the iteration order of a map differs between evaluations, so the
// labels would get assigned to the wrong entries. This is not a problem for an Info metric
// having a single label, as the values of all entries are the same.
// This is a limitation of the CEL resolver of resource-state-metrics.
func (g Generator) validateMapLabels(gvk GroupVersionKind) error {
	if g.PathKind != PathKindMap {
		return nil
	}

	labels := len(g.Labels)

	switch {
	case g.Type == MetricTypeInfo && labels > 1:
		return fmt.Errorf("metric %q of %s: an Info metric on a map supports at most one label, got %d", g.Name, gvk, labels)
	case g.Type != MetricTypeInfo && labels > 0:
		return fmt.Errorf("metric %q of %s: a %s metric on a map doesn't support labels", g.Name, gvk, g.Type)
	}
	return nil
}
