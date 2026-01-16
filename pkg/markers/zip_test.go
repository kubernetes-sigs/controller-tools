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

package markers

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

func TestExtractDoc(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "single line comment",
			input: `package test
// Single line doc
type Foo struct{}`,
			expected: "Single line doc",
		},
		{
			name: "multi-line comment joined into paragraph",
			input: `package test
// This is a multi-line comment that spans
// across multiple lines and should be joined
// into a single paragraph.
type Foo struct{}`,
			expected: "This is a multi-line comment that spans across multiple lines and should be joined into a single paragraph.",
		},
		{
			name: "paragraph separation with blank line",
			input: `package test
// First paragraph of documentation.
//
// Second paragraph after blank line.
type Foo struct{}`,
			expected: "First paragraph of documentation.\n\nSecond paragraph after blank line.",
		},
		{
			name: "multiple paragraphs",
			input: `package test
// First paragraph line one
// continues here.
//
// Second paragraph line one
// continues here.
//
// Third paragraph.
type Foo struct{}`,
			expected: "First paragraph line one continues here.\n\nSecond paragraph line one continues here.\n\nThird paragraph.",
		},
		{
			name: "code block preserved",
			input: `package test
// Example usage:
// ~~~yaml
// foo:
//   bar: baz
// ~~~
type Foo struct{}`,
			expected: "Example usage:\n~~~yaml\nfoo:\n  bar: baz\n~~~",
		},
		{
			name: "code block with description before and after (no blank line)",
			input: `package test
// This is a description
// with multiple lines.
// ~~~yaml
// foo: bar
// ~~~
// More description here.
type Foo struct{}`,
			expected: "This is a description with multiple lines.\n~~~yaml\nfoo: bar\n~~~ More description here.",
		},
		{
			name: "code block with blank line after (recommended)",
			input: `package test
// This is a description.
//
// ~~~yaml
// foo: bar
// ~~~
//
// More description here.
type Foo struct{}`,
			expected: "This is a description.\n\n~~~yaml\nfoo: bar\n~~~\n\nMore description here.",
		},
		{
			name: "triple backticks code block",
			input: `package test
// Example:
// ` + "```" + `go
// func main() {
//   fmt.Println("hello")
// }
// ` + "```" + `
type Foo struct{}`,
			expected: "Example:\n```go\nfunc main() {\n  fmt.Println(\"hello\")\n}\n```",
		},
		{
			name: "TODO comment filtered out",
			input: `package test
// This is documentation.
// TODO: fix this later
// More documentation.
type Foo struct{}`,
			expected: "This is documentation. More documentation.",
		},
		{
			name: "--- cutoff",
			input: `package test
// This is documentation.
// ---
// This should not appear.
type Foo struct{}`,
			expected: "This is documentation.",
		},
		{
			name: "block comment with asterisks",
			input: `package test
/*
   This is a block comment
   with multiple lines.
*/
type Foo struct{}`,
			expected: "This is a block comment with multiple lines.",
		},
		{
			name: "block comment with extra whitespace",
			input: `package test
/*
  This has extra spaces
  that should be trimmed.
*/
type Foo struct{}`,
			expected: "This has extra spaces that should be trimmed.",
		},
		{
			name: "marker comments filtered out",
			input: `package test
// Normal documentation
// +kubebuilder:validation:Optional
// More normal documentation
type Foo struct{}`,
			expected: "Normal documentation More normal documentation",
		},
		{
			name: "empty comment",
			input: `package test
type Foo struct{}`,
			expected: "",
		},
		{
			name: "URL preserved across lines",
			input: `package test
// See documentation at
// https://example.com/very/long/url/path
// for more details.
type Foo struct{}`,
			expected: "See documentation at https://example.com/very/long/url/path for more details.",
		},
		{
			name: "list-like structure joined (current behavior)",
			input: `package test
// Valid values:
// - Option A does this
// - Option B does that
type Foo struct{}`,
			expected: "Valid values: - Option A does this - Option B does that",
		},
		{
			name: "mixed TODO and regular content",
			input: `package test
// This is documentation.
// TODO(user): implement this
// TODO: another todo
// Still more docs.
type Foo struct{}`,
			expected: "This is documentation. Still more docs.",
		},
		{
			name: "code block in middle of paragraphs",
			input: `package test
// Before code block.
//
// ` + "```" + `
// code here
// ` + "```" + `
//
// After code block.
type Foo struct{}`,
			expected: "Before code block.\n\n```\ncode here\n```\n\nAfter code block.",
		},
		// Edge cases
		{
			name: "only whitespace comment",
			input: `package test
//
//
type Foo struct{}`,
			expected: "",
		},
		{
			name: "only markers (all filtered)",
			input: `package test
// +kubebuilder:validation:Required
// +kubebuilder:validation:Optional
type Foo struct{}`,
			expected: "",
		},
		{
			name: "only TODO lines (all filtered)",
			input: `package test
// TODO: implement this
// TODO(user): fix that
type Foo struct{}`,
			expected: "",
		},
		{
			name: "multiple consecutive blank lines (collapsed)",
			input: `package test
// First paragraph.
//
//
//
// Second paragraph.
type Foo struct{}`,
			expected: "First paragraph.\n\nSecond paragraph.",
		},
		{
			name: "very long single line",
			input: `package test
// ` + strings.Repeat("This is a very long line. ", 100) + `
type Foo struct{}`,
			expected: strings.TrimSpace(strings.Repeat("This is a very long line. ", 100)),
		},
		{
			name: "unicode and emoji",
			input: `package test
// This field supports unicode: 日本語 🚀
// and emojis work fine too! 🎉
type Foo struct{}`,
			expected: "This field supports unicode: 日本語 🚀 and emojis work fine too! 🎉",
		},
		{
			name: "tabs and mixed whitespace (preserved for // comments)",
			input: `package test
//		Tab at start gets trimmed
//    Multiple spaces
type Foo struct{}`,
			expected: "\t\tTab at start gets trimmed    Multiple spaces",
		},
		{
			name: "unclosed code block (toggles forever)",
			input: `package test
// Start of doc.
// ` + "```" + `
// code here
// more code
// No closing delimiter!
type Foo struct{}`,
			expected: "Start of doc.\n```\ncode here\nmore code\nNo closing delimiter!",
		},
		{
			name: "multiple code block delimiters on same line",
			input: `package test
// Text before ` + "```code```" + ` inline code
type Foo struct{}`,
			expected: "Text before ```code``` inline code",
		},
		{
			name: "mixing ~~~ and ``` delimiters (code block toggle doesn't discriminate)",
			input: `package test
// Start with ~~~
// ~~~yaml
// foo: bar
// ` + "```" + `
// Still in first block because different delimiter
type Foo struct{}`,
			expected: "Start with ~~~\n~~~yaml\nfoo: bar\n``` Still in first block because different delimiter",
		},
		{
			name: "--- in middle of line (not filtered)",
			input: `package test
// This is a horizontal rule ---
// More text here.
type Foo struct{}`,
			expected: "This is a horizontal rule --- More text here.",
		},
		{
			name: "--- at start (stops processing)",
			input: `package test
// Valid documentation.
// ---
// This should not appear.
type Foo struct{}`,
			expected: "Valid documentation.",
		},
		{
			name: "TODO in middle of line (not filtered)",
			input: `package test
// This has a TODO in the middle
// More text here.
type Foo struct{}`,
			expected: "This has a TODO in the middle More text here.",
		},
		{
			name: "TODO and --- inside code block (preserved)",
			input: `package test
// Example code:
// ` + "```" + `
// TODO: implement this
// ---
// More code
// ` + "```" + `
type Foo struct{}`,
			expected: "Example code:\n```\nTODO: implement this\n---\nMore code\n```",
		},
		{
			name: "code block at start of doc",
			input: `package test
// ` + "```" + `
// code here
// ` + "```" + `
// Text after.
type Foo struct{}`,
			expected: "```\ncode here\n``` Text after.",
		},
		{
			name: "code block at end of doc",
			input: `package test
// Text before.
// ` + "```" + `
// code here
// ` + "```" + `
type Foo struct{}`,
			expected: "Text before.\n```\ncode here\n```",
		},
		{
			name: "empty code block",
			input: `package test
// Empty code block:
// ` + "```" + `
// ` + "```" + `
// After.
type Foo struct{}`,
			expected: "Empty code block:\n```\n``` After.",
		},
		{
			name: "only blank lines between paragraphs (consecutive blanks collapsed)",
			input: `package test
// First
//
//
// Second
type Foo struct{}`,
			expected: "First\n\nSecond",
		},
		{
			name: "trailing blank lines (removed)",
			input: `package test
// Documentation here.
//
//
type Foo struct{}`,
			expected: "Documentation here.",
		},
		{
			name: "marker-like text that isn't a marker",
			input: `package test
// Use +1 to increment
// The +kubebuilder annotation is useful
type Foo struct{}`,
			expected: "Use +1 to increment The +kubebuilder annotation is useful",
		},
		{
			name: "special markdown characters preserved",
			input: `package test
// This uses *asterisks* and _underscores_
// and [links](http://example.com)
type Foo struct{}`,
			expected: "This uses *asterisks* and _underscores_ and [links](http://example.com)",
		},
		{
			name: "alternating blank and non-blank lines",
			input: `package test
// Line 1
//
// Line 2
//
// Line 3
type Foo struct{}`,
			expected: "Line 1\n\nLine 2\n\nLine 3",
		},
		{
			name: "document ending with ---",
			input: `package test
// Valid docs.
// More docs.
// ---
type Foo struct{}`,
			expected: "Valid docs. More docs.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			f, err := parser.ParseFile(fset, "test.go", tt.input, parser.ParseComments)
			if err != nil {
				t.Fatalf("failed to parse test input: %v", err)
			}

			// Find the type spec
			var typeSpec *ast.TypeSpec
			var genDecl *ast.GenDecl
			for _, decl := range f.Decls {
				if gd, ok := decl.(*ast.GenDecl); ok {
					for _, spec := range gd.Specs {
						if ts, ok := spec.(*ast.TypeSpec); ok {
							typeSpec = ts
							genDecl = gd
							break
						}
					}
				}
			}

			if typeSpec == nil {
				t.Fatal("no type spec found in test input")
			}

			result := extractDoc(typeSpec, genDecl)
			if result != tt.expected {
				t.Errorf("extractDoc() result mismatch\nExpected: %q\nGot:      %q", tt.expected, result)
			}
		})
	}
}
