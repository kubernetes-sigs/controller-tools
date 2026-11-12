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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/gengo/v2/types"
)

var _ = Describe("mapEmbeddedKubeTypes", func() {
	localObjRef := types.Name{Package: "k8s.io/api/core/v1", Name: "LocalObjectReference"}
	clientGoCoreV1 := "k8s.io/client-go/applyconfigurations/core/v1"

	structOf := func(name types.Name) *types.Type {
		return &types.Type{Name: name, Kind: types.Struct}
	}
	embed := func(t *types.Type) types.Member {
		return types.Member{Name: "Embedded", Embedded: true, Type: t}
	}

	// run calls mapEmbeddedKubeTypes for a Spec struct with the given members.
	// The testdata module is used because it depends on client-go.
	run := func(externalACs map[types.Name]string, members ...types.Member) map[types.Name]string {
		spec := &types.Type{Name: types.Name{Package: "example.com/api/v1", Name: "Spec"}, Kind: types.Struct, Members: members}
		pkg := &types.Package{Path: "example.com/api/v1", Types: map[string]*types.Type{"Spec": spec}}
		Expect(mapEmbeddedKubeTypes(cronjobDir, pkg, externalACs)).To(Succeed())
		return externalACs
	}

	It("maps an embedded k8s.io/api type to client-go", func() {
		Expect(run(map[types.Name]string{}, embed(structOf(localObjRef)))).
			To(Equal(map[types.Name]string{localObjRef: clientGoCoreV1}))
	})

	It("maps an embedded pointer to a k8s.io/api type", func() {
		ptr := &types.Type{Kind: types.Pointer, Elem: structOf(localObjRef)}
		Expect(run(map[types.Name]string{}, embed(ptr))).
			To(Equal(map[types.Name]string{localObjRef: clientGoCoreV1}))
	})

	It("keeps the mapping set by the user", func() {
		userMapping := map[types.Name]string{localObjRef: "example.com/ac/core/v1"}
		Expect(run(userMapping, embed(structOf(localObjRef)))).
			To(Equal(map[types.Name]string{localObjRef: "example.com/ac/core/v1"}))
	})

	It("does not map a k8s.io/api type that is not embedded", func() {
		field := types.Member{Name: "Ref", Type: structOf(localObjRef)}
		Expect(run(map[types.Name]string{}, field)).To(BeEmpty())
	})

	It("does not map a type that client-go does not have", func() {
		missing := types.Name{Package: "k8s.io/api/core/v1", Name: "DoesNotExist"}
		Expect(run(map[types.Name]string{}, embed(structOf(missing)))).To(BeEmpty())
	})

	It("does not map a type that is not from k8s.io/api", func() {
		other := types.Name{Package: "example.com/shared", Name: "Selector"}
		Expect(run(map[types.Name]string{}, embed(structOf(other)))).To(BeEmpty())
	})
})
