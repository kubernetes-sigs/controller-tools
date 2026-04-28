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

package markers

import (
	"fmt"
	sc "text/scanner"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("parserScanner backticks handling", func() {
	// This tests the issue reported in https://github.com/kubernetes-sigs/controller-tools/issues/1084
	// When a marker contains backticks as literal characters (e.g. Pattern=^[`a-zA-Z]+$),
	// the scanner should gracefully handle them without "literal not terminated" errors.
	// The error recovery in Parse() retries without ScanRawStrings if needed.

	It("should handle pattern with single backtick without error (rawStrings=false)", func() {
		// Pattern with single backtick: ^[`a-zA-Z]+$
		// This is the real-world case from issue #1084
		raw := "^[`a-zA-Z]+$"
		errFunc := func(s *sc.Scanner, msg string) {
			Fail(fmt.Sprintf("scanner error: %s (at %s)", msg, s.Position))
		}

		scanner := parserScanner(raw, errFunc, false)

		// The scanner should be able to scan the pattern without errors
		tokens := []rune{}
		for {
			tok := scanner.Scan()
			if tok == sc.EOF {
				break
			}
			tokens = append(tokens, tok)
		}
		Expect(tokens).NotTo(BeEmpty(), "should have scanned some tokens")
	})

	It("should handle kubebuilder validation items pattern with single backtick (rawStrings=false)", func() {
		// Full marker as it would appear in source code
		raw := "+kubebuilder:validation:items:Pattern=^[`a-zA-Z]+$"
		errFunc := func(s *sc.Scanner, msg string) {
			Fail(fmt.Sprintf("scanner error: %s (at %s)", msg, s.Position))
		}

		scanner := parserScanner(raw, errFunc, false)

		tokens := []rune{}
		for {
			tok := scanner.Scan()
			if tok == sc.EOF {
				break
			}
			tokens = append(tokens, tok)
		}
		Expect(tokens).NotTo(BeEmpty(), "should have scanned some tokens")
	})

	It("should handle pattern with two backticks as raw string (rawStrings=true)", func() {
		// Two backticks form a valid raw string: ^[`a-zA-Z`]+$
		raw := "^[`a-zA-Z`]+$"
		errFunc := func(s *sc.Scanner, msg string) {
			Fail(fmt.Sprintf("scanner error: %s (at %s)", msg, s.Position))
		}

		scanner := parserScanner(raw, errFunc, true)

		Expect(scanner.Mode&sc.ScanRawStrings).To(Equal(uint(sc.ScanRawStrings)),
			"ScanRawStrings should be enabled when rawStrings=true")

		tokens := []rune{}
		for {
			tok := scanner.Scan()
			if tok == sc.EOF {
				break
			}
			tokens = append(tokens, tok)
		}
		Expect(tokens).NotTo(BeEmpty(), "should have scanned some tokens")
	})

	It("should handle pattern with three backticks without error (rawStrings=false)", func() {
		// Three backticks: odd count, fallback mode
		raw := "^[`a-zA-Z`a-z`]+$"
		errFunc := func(s *sc.Scanner, msg string) {
			Fail(fmt.Sprintf("scanner error: %s (at %s)", msg, s.Position))
		}

		scanner := parserScanner(raw, errFunc, false)

		tokens := []rune{}
		for {
			tok := scanner.Scan()
			if tok == sc.EOF {
				break
			}
			tokens = append(tokens, tok)
		}
		Expect(tokens).NotTo(BeEmpty(), "should have scanned some tokens")
	})

	It("should handle empty string without error", func() {
		raw := ""
		errFunc := func(s *sc.Scanner, msg string) {
			Fail(fmt.Sprintf("scanner error: %s (at %s)", msg, s.Position))
		}

		scanner := parserScanner(raw, errFunc, true)

		Expect(scanner.Mode&sc.ScanRawStrings).To(Equal(uint(sc.ScanRawStrings)),
			"ScanRawStrings should be enabled when rawStrings=true")
	})

	It("should handle two literal backticks inside regex character class (rawStrings=false)", func() {
		// Backticks as literal regex characters, not raw string delimiters:
		// ^[`a-zA-Z`]+$  -- these backticks are inside a character class
		raw := "^[`a-zA-Z`]+$"
		errFunc := func(s *sc.Scanner, msg string) {
			Fail(fmt.Sprintf("scanner error: %s (at %s)", msg, s.Position))
		}

		scanner := parserScanner(raw, errFunc, false)

		tokens := []rune{}
		for {
			tok := scanner.Scan()
			if tok == sc.EOF {
				break
			}
			tokens = append(tokens, tok)
		}
		Expect(tokens).NotTo(BeEmpty(), "should have scanned some tokens")
	})
})
