package version

import (
	"fmt"
	"runtime/debug"
)

// version returns the version of the main module
func version() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		// binary has not been built with module support
		return "(unknown)"
	}
	return info.Main.Version
}

// Print prints the main module version on stdout.
//
// Print will display either:
//
// - "Version: v0.2.1" when the program has been compiled with:
//
//   $ go get github.com/controller-tools/cmd/controller-gen@v0.2.1
//
//   Note: go modules requires the usage of semver compatible tags starting with
//        'v' to have nice human-readable versions.
//
// - "Version: (devel)" when the program is compiled from a local git checkout.
//
// - "Version: (unknown)" when not using go modules.
func Print() {
	fmt.Printf("Version: %s\n", version())
}
