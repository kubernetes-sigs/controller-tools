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

// Package metrics contain libraries for generating custom resource metrics configurations
// for kube-state-metrics from metrics markers in Go source files.
package metrics

import (
	"errors"
	"slices"
	"strings"

	"sigs.k8s.io/controller-tools/pkg/crd"
	"sigs.k8s.io/controller-tools/pkg/genall"
	"sigs.k8s.io/controller-tools/pkg/loader"
	ctrlmarkers "sigs.k8s.io/controller-tools/pkg/markers"
	"sigs.k8s.io/controller-tools/pkg/metrics/internal/model"
	"sigs.k8s.io/controller-tools/pkg/metrics/internal/output"
	"sigs.k8s.io/controller-tools/pkg/metrics/markers"
)

// +controllertools:marker:generateHelp

// Generator generates custom resource metrics configuration files for resource-state-metrics
// (default) or kube-state-metrics.
//
// This generator is experimental. The markers, the flags and the generated output might change
// incompatibly in any release. It has to be enabled explicitly by setting the experimental flag.
type Generator struct {
	// Experimental enables the experimental metrics generator.
	// It has to be set to true to acknowledge that markers and output might change incompatibly.
	Experimental bool `marker:",optional"`

	// Target selects the metrics implementation to generate the configuration for.
	// Supported values are resource-state-metrics and kube-state-metrics.
	//
	// Left unspecified, the default is resource-state-metrics.
	Target string `marker:",optional"`

	// Name is the name of the ResourceMetricsMonitor which gets generated for resource-state-metrics.
	// It is only supported by the resource-state-metrics target.
	//
	// Left unspecified, the default is resource-metrics.
	Name string `marker:",optional"`

	// Namespace is the namespace of the ResourceMetricsMonitor which gets generated for resource-state-metrics.
	// It is only supported by the resource-state-metrics target.
	//
	// Left unspecified, the namespace is not set.
	Namespace string `marker:",optional"`
}

var _ genall.Generator = &Generator{}
var _ genall.NeedsTypeChecking = &Generator{}

// RegisterMarkers registers all markers needed by this Generator
// into the given registry.
func (g Generator) RegisterMarkers(into *ctrlmarkers.Registry) error {
	for _, m := range markers.MarkerDefinitions {
		if err := m.Register(into); err != nil {
			return err
		}
	}

	return nil
}

// Generate generates artifacts produced by this marker.
// It's called after RegisterMarkers has been called.
func (g Generator) Generate(ctx *genall.GenerationContext) error {
	if !g.Experimental {
		return errors.New("the metrics generator is experimental and might change incompatibly in future releases, set `metrics:experimental=true` to use it")
	}

	// Create the builder for the selected target first, to fail early on invalid options.
	builder, err := g.newOutputBuilder()
	if err != nil {
		return err
	}

	// Create the parser which is specific to the metric generator.
	parser := newParser(
		&crd.Parser{
			Collector: ctx.Collector,
			Checker:   ctx.Checker,
		},
	)

	// Loop over all passed packages.
	for _, pkg := range ctx.Roots {
		// skip packages which don't import metav1 because they can't define a CRD without meta v1.
		metav1 := pkg.Imports()["k8s.io/apimachinery/pkg/apis/meta/v1"]
		if metav1 == nil {
			continue
		}

		// parse the given package to feed crd.FindKubeKinds with Kubernetes Objects.
		parser.NeedPackage(pkg)

		kubeKinds := crd.FindKubeKinds(parser.Parser, metav1)
		if len(kubeKinds) == 0 {
			// no objects in this package
			continue
		}

		// Create metrics for all Custom Resources in this package.
		// This creates the customresourcestate.Resource object which contains all metric
		// definitions for the Custom Resource, if it is part of the package.
		for _, gv := range kubeKinds {
			if err := parser.NeedResourceFor(pkg, gv); err != nil {
				return err
			}
		}
	}

	// Add the resources to the output. Resources and their metrics get sorted to get a deterministic output.
	if err := addResources(builder, sortedResources(parser.resources)); err != nil {
		return err
	}

	return builder.Write(ctx)
}

// addResources adds the resources to the builder. The resources get validated before, to
// have all targets reject the same unsupported metrics.
func addResources(builder output.Builder, resources []model.Resource) error {
	for _, resource := range resources {
		if err := resource.Validate(); err != nil {
			return err
		}
		if err := builder.AddResource(resource); err != nil {
			return err
		}
	}
	return nil
}

// sortedResources returns the resources having metrics in a deterministic order.
func sortedResources(resourceStates map[crd.TypeIdent]*model.Resource) []model.Resource {
	resources := []model.Resource{}
	for _, resource := range resourceStates {
		if resource == nil || len(resource.Metrics) == 0 {
			continue
		}

		// Sort the metrics to get a deterministic output. The sort has to be stable
		// to keep the order of metrics with the same name.
		slices.SortStableFunc(resource.Metrics, func(a, b model.Generator) int {
			return strings.Compare(a.Name, b.Name)
		})
		resources = append(resources, *resource)
	}

	// Sort the resources by GVK to get a deterministic output.
	slices.SortFunc(resources, func(a, b model.Resource) int {
		return strings.Compare(a.GroupVersionKind.String(), b.GroupVersionKind.String())
	})
	return resources
}

// CheckFilter indicates the loader.NodeFilter (if any) that should be used
// to prune out unused types/packages when type-checking (nodes for which
// the filter returns true are considered "interesting").  This filter acts
// as a baseline -- all types the pass through this filter will be checked,
// but more than that may also be checked due to other generators' filters.
func (Generator) CheckFilter() loader.NodeFilter {
	// Re-use crd filter to filter out unrelated nodes that aren't used
	// in CRD generation, like interfaces and struct fields without JSON tag.
	return crd.Generator{}.CheckFilter()
}
