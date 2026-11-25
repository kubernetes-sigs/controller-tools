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

package crd

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apiservercel "k8s.io/apiserver/pkg/cel"
)

// A generated rule only compiles if it names a property the way the API
// server exposes it to CEL, so the escaping has to match the API server's.
var _ = Describe("EscapeCELFieldName", func() {
	for _, name := range []string{
		"tokenFile",
		"token_file",
		"token-file",
		"token.file",
		"token/file",
		"token__file",
		"token___file",
		"a_-b",
		"a.b-c/d__e",
		// The CEL reserved words, which the API server escapes as a whole.
		"true", "false", "null", "in",
		"as", "break", "const", "continue", "else",
		"for", "function", "if", "import", "let",
		"loop", "package", "namespace", "return",
		"var", "void", "while",
		// Names that only contain a reserved word.
		"inside",
		"namespaces",
		"trueValue",
	} {
		It("escapes "+name+" like the API server", func() {
			want, ok := apiservercel.Escape(name)
			Expect(ok).To(BeTrue(), "apiservercel.Escape(%q) failed", name)
			Expect(EscapeCELFieldName(name)).To(Equal(want))
		})
	}
})
