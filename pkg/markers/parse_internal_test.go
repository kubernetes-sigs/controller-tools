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
	// When a marker contains an odd number of backticks (e.g., Pattern=^[`a-zA-Z]+$),
	// the scanner should NOT enable ScanRawStrings to avoid "literal not terminated" errors.

	It("should handle pattern with single backtick without error", func() {
		// Pattern with single backtick: ^[`a-zA-Z]+$
		// This is the real-world case from issue #1084
		raw := "^[`a-zA-Z]+$"
		errFunc := func(s *sc.Scanner, msg string) {
			Fail(fmt.Sprintf("scanner error: %s (at %s)", msg, s.Position))
		}

		scanner := parserScanner(raw, errFunc)

		// With odd number of backticks (1), ScanRawStrings should be disabled
		Expect(scanner.Mode&sc.ScanRawStrings).To(BeZero(), "ScanRawStrings should be disabled with odd backticks")

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

	It("should handle kubebuilder validation items pattern with single backtick", func() {
		// Full marker as it would appear in source code
		raw := "+kubebuilder:validation:items:Pattern=^[`a-zA-Z]+$"
		errFunc := func(s *sc.Scanner, msg string) {
			Fail(fmt.Sprintf("scanner error: %s (at %s)", msg, s.Position))
		}

		scanner := parserScanner(raw, errFunc)

		// With 1 backtick (odd), ScanRawStrings should be disabled
		Expect(scanner.Mode&sc.ScanRawStrings).To(BeZero(), "ScanRawStrings should be disabled with 1 backtick")
	})

	It("should handle pattern with two backticks (balanced)", func() {
		// Pattern with two backticks (balanced)
		raw := "^[`a-zA-Z`]+$"
		errFunc := func(s *sc.Scanner, msg string) {
			Fail(fmt.Sprintf("scanner error: %s (at %s)", msg, s.Position))
		}

		scanner := parserScanner(raw, errFunc)

		// With 2 backticks (even), ScanRawStrings should be enabled
		Expect(scanner.Mode&uint(sc.ScanRawStrings)).To(Equal(uint(sc.ScanRawStrings)), "ScanRawStrings should be enabled with even backticks")
	})

	It("should handle pattern with three backticks (unbalanced)", func() {
		// Pattern with three backticks (unbalanced)
		raw := "^[`a-zA-Z`a-z`]+$"
		errFunc := func(s *sc.Scanner, msg string) {
			Fail(fmt.Sprintf("scanner error: %s (at %s)", msg, s.Position))
		}

		scanner := parserScanner(raw, errFunc)

		// With 3 backticks (odd), ScanRawStrings should be disabled
		Expect(scanner.Mode&sc.ScanRawStrings).To(BeZero(), "ScanRawStrings should be disabled with 3 backticks")
	})

	It("should handle empty string without backticks", func() {
		raw := ""
		errFunc := func(s *sc.Scanner, msg string) {
			Fail(fmt.Sprintf("scanner error: %s (at %s)", msg, s.Position))
		}

		scanner := parserScanner(raw, errFunc)

		// With 0 backticks (even), ScanRawStrings should be enabled
		Expect(scanner.Mode&uint(sc.ScanRawStrings)).To(Equal(uint(sc.ScanRawStrings)), "ScanRawStrings should be enabled with 0 backticks")
	})
})
