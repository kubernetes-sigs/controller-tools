/*
Copyright The Kubernetes Authors.

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

package applyconfiguration

import (
	"fmt"
	"strings"

	"golang.org/x/tools/go/packages"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/gengo/v2/types"
)

const (
	kubeAPIPkg    = "k8s.io/api/"
	clientGoACPkg = "k8s.io/client-go/applyconfigurations/"
	acSuffix      = "ApplyConfiguration"
)

// mapEmbeddedKubeTypes maps each k8s.io/api type embedded in a struct of pkg to its
// apply configuration in client-go. Without it, the generated code does not compile
// for types such as corev1.LocalObjectReference.
//
// Types already set in externalACs are not changed. A type is only mapped when
// client-go has its apply configuration.
func mapEmbeddedKubeTypes(dir string, pkg *types.Package, externalACs map[types.Name]string) error {
	toMap := map[types.Name]string{}
	for _, t := range pkg.Types {
		if t.Kind != types.Struct {
			continue
		}
		for _, m := range t.Members {
			if !m.Embedded {
				continue
			}
			embedded := m.Type
			if embedded.Kind == types.Pointer {
				embedded = embedded.Elem
			}
			if embedded.Kind != types.Struct || !strings.HasPrefix(embedded.Name.Package, kubeAPIPkg) {
				continue
			}
			if _, ok := externalACs[embedded.Name]; ok {
				continue
			}
			toMap[embedded.Name] = clientGoACPkg + strings.TrimPrefix(embedded.Name.Package, kubeAPIPkg)
		}
	}
	if len(toMap) == 0 {
		return nil
	}

	acPkgs := sets.New[string]()
	for _, acPkg := range toMap {
		acPkgs.Insert(acPkg)
	}
	loaded, err := packages.Load(&packages.Config{Mode: packages.NeedName | packages.NeedTypes, Dir: dir}, sets.List(acPkgs)...)
	if err != nil {
		return fmt.Errorf("failed to load client-go apply configurations: %w", err)
	}
	found := map[string]*packages.Package{}
	for _, p := range loaded {
		if len(p.Errors) == 0 && p.Types != nil {
			found[p.PkgPath] = p
		}
	}

	for name, acPkg := range toMap {
		p, ok := found[acPkg]
		if !ok || p.Types.Scope().Lookup(name.Name+acSuffix) == nil {
			continue
		}
		externalACs[name] = acPkg
	}
	return nil
}
