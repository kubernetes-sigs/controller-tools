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
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
)

// celReservedWords are the reserved words of the CEL language. The API server
// exposes a property with one of these names to CEL as __<name>__.
var celReservedWords = sets.New(
	"true", "false", "null", "in",
	"as", "break", "const", "continue", "else",
	"for", "function", "if", "import", "let",
	"loop", "package", "namespace", "return",
	"var", "void", "while",
)

var celFieldNameReplacer = strings.NewReplacer(
	"__", "__underscores__",
	".", "__dot__",
	"-", "__dash__",
	"/", "__slash__",
)

// EscapeCELFieldName returns the name a CEL validation rule has to use to
// select the property with the given JSON name, as in has(self.<name>).
// Use it whenever a generated rule refers to a property by its JSON name:
// the API server escapes "__", ".", "-" and "/" in the property names it
// exposes to CEL, and a name that is a CEL reserved word, such as in or
// namespace, becomes __<name>__, so a rule with the raw name fails to compile.
func EscapeCELFieldName(name string) string {
	if celReservedWords.Has(name) {
		return "__" + name + "__"
	}
	return celFieldNameReplacer.Replace(name)
}
