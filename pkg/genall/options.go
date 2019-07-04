/*
Copyright 2019 The Kubernetes Authors.

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

package genall

import (
	"fmt"
	"strings"

	"sigs.k8s.io/controller-tools/pkg/markers"
)

var (
	InputPathsMarker = markers.Must(markers.MakeDefinition("paths", markers.DescribesPackage, InputPaths(nil)))
)

// +controllertools:marker:generateHelp:category=""

// InputPaths represents paths and go-style path patterns to use as package roots.
type InputPaths []string

// RegisterOptionsMarkers registers "mandatory" options markers for FromOptions into the given registry.
// At this point, that's just InputPaths.
func RegisterOptionsMarkers(into *markers.Registry) error {
	if err := into.Register(InputPathsMarker); err != nil {
		return err
	}
	// NB(directxman12): we make this optional so we don't have a bootstrap problem with helpgen
	if helpGiver, hasHelp := ((interface{})(InputPaths(nil))).(HasHelp); hasHelp {
		into.AddHelp(InputPathsMarker, helpGiver.Help())
	}
	return nil
}

// FromOptions parses the options from markers stored in the given registry out into a runtime.
// The markers in the registry must be either
//
// a) Generators
// b) OutputRules
// c) InputPaths
//
// The paths specified in InputPaths are loaded as package roots, and the combined with
// the generators and the specified output rules to produce a runtime that can be run or
// further modified.  Not default generators are used if none are specified -- you can check
// the output and rerun for that.
func FromOptions(optionsRegistry *markers.Registry, options []string) (*Runtime, error) {
	var gens Generators
	rules := OutputRules{
		ByGenerator: make(map[Generator]OutputRule),
	}
	var paths []string

	// collect the generators first, so that we can key the output on the actual
	// generator, which matters if there's settings in the gen object and it's not a pointer.
	outputByGen := make(map[string]OutputRule)
	gensByName := make(map[string]Generator)

	for _, rawOpt := range options {
		if rawOpt[0] != '+' {
			rawOpt = "+" + rawOpt // add a `+` to make it acceptable for usage with the registry
		}
		defn := optionsRegistry.Lookup(rawOpt, markers.DescribesPackage)
		if defn == nil {
			return nil, fmt.Errorf("unknown option %q", rawOpt[1:])
		}

		val, err := defn.Parse(rawOpt)
		if err != nil {
			return nil, fmt.Errorf("unable to parse option %q: %v", rawOpt[1:], err)
		}

		switch val := val.(type) {
		case Generator:
			gens = append(gens, val)
			gensByName[defn.Name] = val
		case OutputRule:
			_, genName := splitOutputRuleOption(defn.Name)
			if genName == "" {
				// it's a default rule
				rules.Default = val
				continue
			}

			outputByGen[genName] = val
			continue
		case InputPaths:
			paths = append(paths, val...)
		default:
			return nil, fmt.Errorf("unknown option marker %q", defn.Name)
		}
	}

	// actually associate the rules now that we know the generators
	for genName, outputRule := range outputByGen {
		gen, knownGen := gensByName[genName]
		if !knownGen {
			return nil, fmt.Errorf("non-invoked generator %q", genName)
		}

		rules.ByGenerator[gen] = outputRule
	}

	// make the runtime
	genRuntime, err := Generators(gens).ForRoots(paths...)
	if err != nil {
		return nil, err
	}

	// attempt to figure out what the user wants without a lot of verbose specificity:
	// if the user specifies a default rule, assume that they probably want to fall back
	// to that.  Otherwise, assume that they just wanted to customize one option from the
	// set, and leave the rest in the standard configuration.
	if rules.Default != nil {
		genRuntime.OutputRules = rules
		return genRuntime, nil
	}

	outRules := DirectoryPerGenerator("config", gensByName)
	for gen, rule := range rules.ByGenerator {
		outRules.ByGenerator[gen] = rule
	}

	genRuntime.OutputRules = outRules
	return genRuntime, nil
}

// splitOutputRuleOption splits a marker name of "output:rule:gen" or "output:rule"
// into its compontent rule and generator name.
func splitOutputRuleOption(name string) (ruleName string, genName string) {
	parts := strings.SplitN(name, ":", 3)
	if len(parts) == 3 {
		// output:<generator>:<rule>
		return parts[2], parts[1]
	}
	// output:<rule>
	return parts[1], ""
}
