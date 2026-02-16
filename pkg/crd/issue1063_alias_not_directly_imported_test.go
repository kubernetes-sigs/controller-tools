package crd_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// repoRoot walks up from this test file until it finds the repository go.mod.
func repoRoot(t *testing.T) string {
	t.Helper()

	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	dir := filepath.Dir(thisFile)

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("could not find go.mod walking up from %s", thisFile)
		}
		dir = parent
	}
}

func buildControllerGen(t *testing.T, root string) string {
	t.Helper()

	bin := filepath.Join(t.TempDir(), "controller-gen")
	cmd := exec.Command("go", "build", "-o", bin, "./cmd/controller-gen")
	cmd.Dir = root

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to build controller-gen: %v\nstderr:\n%s", err, stderr.String())
	}
	return bin
}

func TestIssue1063_AliasToPackageNotDirectlyImported(t *testing.T) {
	root := repoRoot(t)
	bin := buildControllerGen(t, root)

	fixtureRoot := filepath.Join(root, "pkg", "crd", "testdata")
	repro1Rel := "./issue1063/repro1"

	cmd := exec.Command(bin,
		"crd",
		"paths="+repro1Rel,
		"output:crd:stdout",
	)
	// Run within the testdata module (module testdata.kubebuilder.io/cronjob).
	cmd.Dir = fixtureRoot

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		t.Fatalf("controller-gen failed: %v\nstderr:\n%s\nstdout:\n%s", err, stderr.String(), stdout.String())
	}

	out := stdout.String()
	errOut := stderr.String()

	if strings.Contains(errOut, "panic:") {
		t.Fatalf("unexpected panic output:\n%s", errOut)
	}

	if !strings.Contains(out, "kind: CustomResourceDefinition") {
		t.Fatalf("expected CRD output, got:\n%s", out)
	}

	// The struct field is `Repro string `json:"repro"``, so YAML schema should mention `repro:`.
	if !strings.Contains(out, "\n                  repro:\n") &&
		!strings.Contains(out, "\n                repro:\n") &&
		!strings.Contains(out, "\n              repro:\n") {
		t.Fatalf("expected schema to include 'repro' property, got:\n%s", out)
	}
}
