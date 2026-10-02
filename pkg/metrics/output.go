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
	"maps"
	"slices"
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

// defaultTarget is the target used if none is selected.
const defaultTarget = targetResourceStateMetrics

// targets maps the supported targets to the function creating their output.Builder.
// The function validates the options of the Generator which are relevant for the target.
// Adding a new target requires an output.Builder implementation and an entry here.
var targets = map[string]func(g Generator) (output.Builder, error){
	targetResourceStateMetrics: newRSMBuilder,
	targetKubeStateMetrics:     newKSMBuilder,
}

func newRSMBuilder(g Generator) (output.Builder, error) {
	if g.Name != "" {
		if errs := validation.IsDNS1123Subdomain(g.Name); len(errs) > 0 {
			return nil, fmt.Errorf("invalid name %q: %s", g.Name, strings.Join(errs, ", "))
		}
	}
	if g.Namespace != "" {
		if errs := validation.IsDNS1123Label(g.Namespace); len(errs) > 0 {
			return nil, fmt.Errorf("invalid namespace %q: %s", g.Namespace, strings.Join(errs, ", "))
		}
	}
	return rsm.NewBuilder(g.Name, g.Namespace), nil
}

func newKSMBuilder(g Generator) (output.Builder, error) {
	if g.Name != "" || g.Namespace != "" {
		return nil, errors.New("the name and namespace options are only supported for the " + targetResourceStateMetrics + " target")
	}
	return ksm.NewBuilder(), nil
}

// newOutputBuilder returns the output.Builder for the selected target.
func (g Generator) newOutputBuilder() (output.Builder, error) {
	target := g.Target
	if target == "" {
		target = defaultTarget
	}

	newBuilder, ok := targets[target]
	if !ok {
		return nil, fmt.Errorf("unsupported metrics target %q, supported targets are %s", target, strings.Join(slices.Sorted(maps.Keys(targets)), ", "))
	}
	return newBuilder(g)
}
