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

package metrics

import (
	"errors"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/util/validation"

	"sigs.k8s.io/controller-tools/pkg/metrics/internal/ksm"
	"sigs.k8s.io/controller-tools/pkg/metrics/internal/output"
	"sigs.k8s.io/controller-tools/pkg/metrics/internal/rsm"
)

// Supported targets, which are the metrics implementations to generate configuration for.
const (
	targetResourceStateMetrics = "resource-state-metrics"
	targetKubeStateMetrics     = "kube-state-metrics"
)

func newRSMBuilder(g Generator) (output.Builder, error) {
	if g.ResourceMetricsMonitorName != "" {
		if errs := validation.IsDNS1123Subdomain(g.ResourceMetricsMonitorName); len(errs) > 0 {
			return nil, fmt.Errorf("invalid resourceMetricsMonitorName %q: %s", g.ResourceMetricsMonitorName, strings.Join(errs, ", "))
		}
	}
	if g.ResourceMetricsMonitorNamespace != "" {
		if errs := validation.IsDNS1123Label(g.ResourceMetricsMonitorNamespace); len(errs) > 0 {
			return nil, fmt.Errorf("invalid resourceMetricsMonitorNamespace %q: %s", g.ResourceMetricsMonitorNamespace, strings.Join(errs, ", "))
		}
	}
	return rsm.NewBuilder(g.ResourceMetricsMonitorName, g.ResourceMetricsMonitorNamespace), nil
}

func newKSMBuilder(g Generator) (output.Builder, error) {
	if g.ResourceMetricsMonitorName != "" || g.ResourceMetricsMonitorNamespace != "" {
		return nil, errors.New("the resourceMetricsMonitorName and resourceMetricsMonitorNamespace options are only supported for the " + targetResourceStateMetrics + " target")
	}
	return ksm.NewBuilder(), nil
}

// newOutputBuilder returns the output.Builder for the selected target.
// The builder validates the options of the Generator which are relevant for the target.
// Adding a new target requires an output.Builder implementation and a case here.
func (g Generator) newOutputBuilder() (output.Builder, error) {
	switch g.Target {
	case "", targetResourceStateMetrics:
		return newRSMBuilder(g)
	case targetKubeStateMetrics:
		return newKSMBuilder(g)
	default:
		return nil, fmt.Errorf("unsupported metrics target %q, supported targets are %s and %s", g.Target, targetKubeStateMetrics, targetResourceStateMetrics)
	}
}
