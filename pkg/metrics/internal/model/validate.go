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

import "fmt"

// Validate returns an error if the resource has metrics which are not supported by the targets.
// It is checked for all targets, to have them behave the same way.
func (r Resource) Validate() error {
	for _, generator := range r.Metrics {
		if err := generator.validateMapLabels(r.GroupVersionKind); err != nil {
			return err
		}
	}
	return nil
}

// validateMapLabels rejects labels for the entries of a map which can't be aligned with each
// other and with the value. Some targets (e.g. resource-state-metrics) evaluate each label and
// the value separately and the iteration order of a map differs between evaluations, so the
// labels would get assigned to the wrong entries. This is not a problem for an Info metric
// having a single label, as the values of all entries are the same.
func (g Generator) validateMapLabels(gvk GroupVersionKind) error {
	if g.PathKind != PathKindMap {
		return nil
	}

	labels := len(g.Labels)
	if g.KeyLabel != "" {
		labels++
	}

	switch {
	case g.Type == MetricTypeInfo && labels > 1:
		return fmt.Errorf("metric %q of %s: an Info metric on a map supports at most one label (keyLabel or one of labels), got %d", g.Name, gvk, labels)
	case g.Type != MetricTypeInfo && labels > 0:
		return fmt.Errorf("metric %q of %s: a %s metric on a map doesn't support labels (keyLabel, labels)", g.Name, gvk, g.Type)
	}
	return nil
}
