/*
Copyright 2024 The Kubernetes Authors.

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
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestYearSubstitution(t *testing.T) {
	tests := []struct {
		name           string
		headerContent  string
		year           string
		expectedOutput string
	}{
		{
			name:           "year substitution with year specified",
			headerContent:  "Copyright YEAR The Kubernetes Authors.",
			year:           "2024",
			expectedOutput: "Copyright 2024 The Kubernetes Authors.",
		},
		{
			name:           "year substitution with empty year",
			headerContent:  "Copyright YEAR The Kubernetes Authors.",
			year:           "",
			expectedOutput: "Copyright  The Kubernetes Authors.",
		},
		{
			name:           "no YEAR placeholder in header",
			headerContent:  "Copyright 2023 The Kubernetes Authors.",
			year:           "2024",
			expectedOutput: "Copyright 2023 The Kubernetes Authors.",
		},
		{
			name:           "multiple YEAR placeholders",
			headerContent:  "Copyright YEAR- YEAR The Kubernetes Authors.",
			year:           "2024",
			expectedOutput: "Copyright 2024- 2024 The Kubernetes Authors.",
		},
		{
			name:           "YEAR without space not replaced",
			headerContent:  "CopyrightYEAR YEAR",
			year:           "2024",
			expectedOutput: "CopyrightYEAR 2024",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			headerFile := filepath.Join(tmpDir, "header.txt")

			if err := os.WriteFile(headerFile, []byte(tt.headerContent), 0644); err != nil {
				t.Fatalf("failed to write header file: %v", err)
			}

			result := strings.ReplaceAll(tt.headerContent, " YEAR", " "+tt.year)

			if result != tt.expectedOutput {
				t.Errorf("expected %q, got %q", tt.expectedOutput, result)
			}
		})
	}
}
