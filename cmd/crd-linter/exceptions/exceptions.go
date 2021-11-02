/*
Copyright 2021 The Kubernetes Authors.

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

package exceptions

import (
	"fmt"
	"io/ioutil"
	"strings"
)

type ExceptionList struct {
	Exceptions []Exception
}

type Exception struct {
	filename, crdName, linter, fieldRef string
}

func (e Exception) String() string {
	return fmt.Sprintf("%s|%s|%s|%s", e.filename, e.crdName, e.linter, e.fieldRef)
}

func (l *ExceptionList) IsExcepted(filename, crdName, linter, fieldRef string) bool {
	toFind := Exception{
		filename: filename,
		crdName:  crdName,
		linter:   linter,
		fieldRef: fieldRef,
	}
	for _, e := range l.Exceptions {
		if e == toFind {
			return true
		}
	}
	return false
}

func (l *ExceptionList) Add(filename, crdName, linter, fieldRef string) {
	if l.IsExcepted(filename, crdName, linter, fieldRef) {
		return
	}
	l.Exceptions = append(l.Exceptions, Exception{
		filename: filename,
		crdName:  crdName,
		linter:   linter,
		fieldRef: fieldRef,
	})
}

func (l *ExceptionList) String() string {
	buf := &strings.Builder{}
	for _, e := range l.Exceptions {
		buf.WriteString(fmt.Sprintf("%s\n", e.String()))
	}
	return buf.String()
}

func (l *ExceptionList) Size() int {
	return len(l.Exceptions)
}

func (l *ExceptionList) WriteToFile(path string) error {
	// #nosec G306
	return ioutil.WriteFile(path, []byte(l.String()), 0644)
}

func NewExceptionList() *ExceptionList {
	return &ExceptionList{Exceptions: make([]Exception, 0)}
}

// LoadFromFile will load a pipe (|) separated list of Exceptions from the
// given file.
func LoadFromFile(path string) (*ExceptionList, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	list := NewExceptionList()
	entries := strings.Split(string(data), "\n")
	for lineNo, e := range entries {
		// skip empty lines
		if len(e) == 0 {
			continue
		}
		e := strings.Split(e, "|")
		if len(e) != 4 {
			return nil, fmt.Errorf("invalid Exception entry at line %d", lineNo+1)
		}
		list.Add(e[0], e[1], e[2], e[3])
	}
	return list, nil
}
