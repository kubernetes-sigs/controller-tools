package main

import (
	"log"
	"os"
	"strings"

	flag "github.com/spf13/pflag"

	"sigs.k8s.io/controller-tools/cmd/crd-linter/evaluator"
	"sigs.k8s.io/controller-tools/cmd/crd-linter/exceptions"
	"sigs.k8s.io/controller-tools/cmd/crd-linter/linters"
	"sigs.k8s.io/controller-tools/cmd/crd-linter/reader"
)

var (
	enabledLinters []string

	path string

	exceptionsFile   string
	outputExceptions bool

	errOnViolations bool
)

func init() {
	flag.StringSliceVar(&enabledLinters, "linters", linters.All().Names(), "Full list of linters to run against discovered CustomResourceDefinitions")
	flag.StringVar(&path, "path", "", "Path to recursively search for CustomResourceDefinition objects")
	flag.StringVar(&exceptionsFile, "exceptions-file", "", "Path to a list of exceptions for linter failures")
	flag.BoolVar(&outputExceptions, "output-exceptions", false, "If true, an exception list file will be written to the file denoted by '--exceptions-file'")
	flag.BoolVar(&errOnViolations, "error-on-violations", true, "If true, the linter will exit with a non-zero exit code if there are any unexpected violations")
}

func main() {
	flag.Parse()

	linters, err := linters.Named(enabledLinters)
	if err != nil {
		log.Fatalf("Failed to build linter list: %s", err.Error())
	}

	log.Printf("Enabled linters: %s", strings.Join(linters.Names(), ", "))

	log.Printf("Searching path %q for CustomResourceDefinition objects", path)
	crds, err := reader.DiscoverAllCRDs(path)
	if err != nil {
		log.Fatalf("Error reading CRDs from disk: %v", err)
	}

	if outputExceptions {
		results := evaluator.Evaluate(linters, crds, nil)
		if exceptionsFile == "" {
			log.Fatalf("--exceptions-file must be specified if --output-exceptions=true")
		}

		e := results.ToExceptionList()
		if err := e.WriteToFile(exceptionsFile); err != nil {
			log.Fatalf("Failed to write exceptions file: %v", err)
		}
		return
	}

	var excepted *exceptions.ExceptionList
	if exceptionsFile != "" {
		excepted, err = exceptions.LoadFromFile(exceptionsFile)
		if err != nil {
			log.Fatalf("Failed loading exceptions file %q: %v", exceptionsFile, err)
		}

		log.Printf("Loaded %d exceptions from file %q", excepted.Size(), exceptionsFile)
	}

	results := evaluator.Evaluate(linters, crds, excepted)
	for _, r := range results {
		if !r.HasViolations() {
			continue
		}

		log.Printf("CRD %q in file %q has linter failures:", r.Evaluated.CustomResourceDefinition.Name, r.Evaluated.OriginalFilename)
		for _, list := range r.ViolatedLinters() {
			log.Println()
			log.Printf("\tLinter %q failed:", list.Linter.Name())
			for _, err := range list.Violations {
				log.Printf("\t\t%s", err)
			}
			log.Println()
			log.Printf("\tDescription: %s", list.Linter.Description())
			log.Println()
		}
	}

	if errOnViolations && results.HasViolations() {
		log.Println()
		log.Printf("Unexpected violations found!")
		os.Exit(1)
	}
}
