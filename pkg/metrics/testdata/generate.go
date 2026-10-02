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

// Package testdata contains the types used to test and demonstrate the metrics generator.
//
// Changes to the packages below this directory may require to regenerate the
// `metrics.yaml`, `rbac.yaml` and `bar.example.com_foos.yaml` files. Otherwise the tests in
// ../generate_integration_test.go may fail.
// The below marker can be used to regenerate the files by running the following command:
// $ go generate ./pkg/metrics/testdata
//
//go:generate sh -c "go run ../../../cmd/controller-gen crd metrics:experimental=true paths=./... output:dir=."
package testdata
