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
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pmezard/go-difflib/difflib"
	gengo "k8s.io/gengo/v2"
	"sigs.k8s.io/controller-tools/pkg/loader"
	"sigs.k8s.io/controller-tools/pkg/version"
)

// nopCloser is a WriteCloser whose Close
// is a no-op.
type nopCloser struct {
	io.Writer
}

func (n nopCloser) Close() error {
	return nil
}

// DirectoryPerGenerator produces output rules mapping output to a different subdirectory
// of the given base directory for each generator (with each subdirectory specified as
// the key in the input map).
func DirectoryPerGenerator(base string, generators map[string]*Generator) OutputRules {
	rules := OutputRules{
		Default:     OutputArtifacts{Config: OutputToDirectory(base)},
		ByGenerator: make(map[*Generator]OutputRule, len(generators)),
	}

	for name, gen := range generators {
		rules.ByGenerator[gen] = OutputArtifacts{
			Config: OutputToDirectory(filepath.Join(base, name)),
		}
	}

	return rules
}

// OutputRules defines how to output artificats on a per-generator basis.
type OutputRules struct {
	// Default is the output rule used when no specific per-generator overrides match.
	Default OutputRule
	// ByGenerator contains specific per-generator overrides.
	// NB(directxman12): this is a pointer to avoid issues if a given Generator becomes unhashable
	// (interface values compare by "dereferencing" their internal pointer first, whereas pointers
	// compare by the actual pointer itself).
	ByGenerator map[*Generator]OutputRule
}

// ForGenerator returns the output rule that should be used
// by the given Generator.
func (o OutputRules) ForGenerator(gen *Generator) OutputRule {
	if forGen, specific := o.ByGenerator[gen]; specific {
		return forGen
	}
	return o.Default
}

// OutputRule defines how to output artifacts from a generator.
type OutputRule interface {
	// Open opens the given artifact path for writing.  If a package is passed,
	// the artifact is considered to be used as part of the package (e.g.
	// generated code), while a nil package indicates that the artifact is
	// config (or something else not involved in Go compilation).
	Open(pkg *loader.Package, path string) (io.WriteCloser, error)
}

// OutputToNothing skips outputting anything.
var OutputToNothing = outputToNothing{}

// +controllertools:marker:generateHelp:category=""

// outputToNothing skips outputting anything.
type outputToNothing struct{}

func (o outputToNothing) Open(_ *loader.Package, _ string) (io.WriteCloser, error) {
	return nopCloser{io.Discard}, nil
}

// +controllertools:marker:generateHelp:category=""

// OutputToDirectory outputs each artifact to the given directory, regardless
// of if it's package-associated or not.
type OutputToDirectory string

func (o OutputToDirectory) Open(_ *loader.Package, itemPath string) (io.WriteCloser, error) {
	// ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(filepath.Join(string(o), itemPath)), os.ModePerm); err != nil {
		return nil, err
	}
	path := filepath.Join(string(o), itemPath)
	return os.Create(path)
}

// OutputToStdout outputs everything to standard-out, with no separation.
//
// Generally useful for single-artifact outputs.
var OutputToStdout = outputToStdout{}

// +controllertools:marker:generateHelp:category=""

// outputToStdout outputs everything to standard-out, with no separation.
//
// Generally useful for single-artifact outputs.
type outputToStdout struct{}

func (o outputToStdout) Open(_ *loader.Package, _ string) (io.WriteCloser, error) {
	return nopCloser{os.Stdout}, nil
}

// +controllertools:marker:generateHelp:category=""

// OutputArtifacts outputs artifacts to different locations, depending on
// whether they're package-associated or not.
//
// Non-package associated artifacts
// are output to the Config directory, while package-associated ones are output
// to their package's source files' directory, unless an alternate path is
// specified in Code.
type OutputArtifacts struct {
	// Config points to the directory to which to write configuration.
	Config OutputToDirectory
	// Code overrides the directory in which to write new code (defaults to where the existing code lives).
	Code OutputToDirectory `marker:",optional"`
}

func (o OutputArtifacts) Open(pkg *loader.Package, itemPath string) (io.WriteCloser, error) {
	if pkg == nil {
		return o.Config.Open(pkg, itemPath)
	}

	if o.Code != "" {
		return o.Code.Open(pkg, itemPath)
	}

	if len(pkg.CompiledGoFiles) == 0 {
		return nil, fmt.Errorf("cannot output to a package with no path on disk")
	}
	outDir := filepath.Dir(pkg.CompiledGoFiles[0])
	if _, err := os.Stat(outDir); os.IsNotExist(err) {
		if err := os.MkdirAll(outDir, os.ModePerm); err != nil {
			return nil, err
		}
	}

	outPath := filepath.Join(outDir, itemPath)
	return os.Create(outPath)
}

// skipUnchangedRule wraps an OutputRule to skip writing files whose content
// has not meaningfully changed. It defers the actual file creation to Close()
// time so that unchanged files are never truncated.
type skipUnchangedRule struct {
	inner OutputRule
}

func (s *skipUnchangedRule) Open(pkg *loader.Package, itemPath string) (io.WriteCloser, error) {
	filePath := resolveOutputPath(s.inner, pkg, itemPath)
	if filePath == "" {
		return s.inner.Open(pkg, itemPath)
	}

	var oldContent []byte
	oldContent, _ = os.ReadFile(filePath)

	return &skipUnchangedWriter{
		inner:      s.inner,
		pkg:        pkg,
		itemPath:   itemPath,
		oldContent: oldContent,
	}, nil
}

// resolveOutputPath determines the filesystem path that the given OutputRule
// would write to, without actually opening the file.
func resolveOutputPath(rule OutputRule, pkg *loader.Package, itemPath string) string {
	switch r := rule.(type) {
	case OutputToDirectory:
		return filepath.Join(string(r), itemPath)
	case OutputArtifacts:
		if pkg == nil {
			return filepath.Join(string(r.Config), itemPath)
		}
		if r.Code != "" {
			return filepath.Join(string(r.Code), itemPath)
		}
		if len(pkg.CompiledGoFiles) > 0 {
			return filepath.Join(filepath.Dir(pkg.CompiledGoFiles[0]), itemPath)
		}
	}
	return ""
}

// skipUnchangedWriter buffers all writes and only creates/writes the target
// file on Close() if the content has meaningfully changed.
type skipUnchangedWriter struct {
	inner      OutputRule
	pkg        *loader.Package
	itemPath   string
	buf        bytes.Buffer
	oldContent []byte
}

func (w *skipUnchangedWriter) Write(p []byte) (int, error) {
	return w.buf.Write(p)
}

func (w *skipUnchangedWriter) Close() error {
	newContent := w.buf.Bytes()

	_, err := WriteFileIfChanged(w.oldContent, newContent, func(content []byte) error {
		innerWriter, err := w.inner.Open(w.pkg, w.itemPath)
		if err != nil {
			return err
		}
		if _, err := innerWriter.Write(content); err != nil {
			innerWriter.Close()
			return err
		}
		return innerWriter.Close()
	})
	return err
}

// WriteFileIfChanged compares newContent against oldContent and only calls
// writeFn if the content has meaningfully changed (ignoring version
// annotations and generated-by comments). Returns true if writeFn was called.
func WriteFileIfChanged(oldContent, newContent []byte, writeFn func([]byte) error) (bool, error) {
	if oldContent != nil && bytes.Equal(oldContent, newContent) {
		return false, nil
	}
	if len(oldContent) > 0 && ContentEqualIgnoringGenVersion(oldContent, newContent) {
		return false, nil
	}
	return true, writeFn(newContent)
}

// codeGeneratedByPattern matches the "Code generated by" header that gengo
// and controller-gen emit, derived from gengo.StdGeneratedBy by replacing
// the GENERATOR_NAME placeholder with a wildcard.
var codeGeneratedByPattern = regexp.MustCompile(
	"^" + strings.Replace(
		regexp.QuoteMeta(strings.Replace(gengo.StdGeneratedBy, "GENERATOR_NAME", "PLACEHOLDER", 1)),
		"PLACEHOLDER", ".+", 1,
	) + "$",
)

// ContentEqualIgnoringGenVersion compares two file contents and returns true
// if they are equal after stripping auto-generated version markers. It uses
// line-level diffing to identify changes, then checks whether all changed
// lines are version annotations or generated-by comments.
func ContentEqualIgnoringGenVersion(a, b []byte) bool {
	aLines := strings.Split(string(a), "\n")
	bLines := strings.Split(string(b), "\n")

	matcher := difflib.NewMatcher(aLines, bLines)
	for _, op := range matcher.GetOpCodes() {
		if op.Tag == 'e' {
			continue
		}
		for i := op.I1; i < op.I2; i++ {
			if !isAutoGeneratedLine(aLines[i]) {
				return false
			}
		}
		for j := op.J1; j < op.J2; j++ {
			if !isAutoGeneratedLine(bLines[j]) {
				return false
			}
		}
	}
	return true
}

func isAutoGeneratedLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	if strings.HasPrefix(trimmed, version.VersionAnnotationKey+":") {
		return true
	}
	if codeGeneratedByPattern.MatchString(trimmed) {
		return true
	}
	if trimmed == "//go:build !"+gengo.StdBuildTag || trimmed == "// +build !"+gengo.StdBuildTag {
		return true
	}
	return false
}
