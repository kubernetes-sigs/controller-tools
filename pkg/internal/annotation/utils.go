/*
Copyright 2018 The Kubernetes Authors.

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

package annotation

import (
	"fmt"
	"os"
	"strings"
)

func prefixName(name string) string {
	return "+" + name
}

// isGoFile filters files from parsing.
func isGoFile(f os.FileInfo) bool {
	// ignore non-Go or Go test files
	name := f.Name()
	return !f.IsDir() &&
		!strings.HasPrefix(name, ".") &&
		!strings.HasSuffix(name, "_test.go") &&
		strings.HasSuffix(name, ".go")
}

// ParseKV parses key-value string formatted as "foo=bar" and returns key and value.
func ParseKV(s string) (key, value string, err error) {
	kv := strings.SplitN(s, "=", 2)
	if len(kv) != 2 {
		err = fmt.Errorf("invalid key value pair")
		return key, value, err
	}
	key, value = kv[0], kv[1]
	if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
		value = value[1 : len(value)-1]
	}
	return key, value, err
}
