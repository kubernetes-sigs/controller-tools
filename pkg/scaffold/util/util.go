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

package util

import (
	"io"
	"log"
	"text/template"
)

// WriteTemplate writes the template to the WriteCloser provided by the function
func WriteTemplate(t *template.Template, i interface{}, wr func() io.WriteCloser) error {
	w := wr()
	defer func() {
		if err := w.Close(); err != nil {
			log.Fatal(err)
		}
	}()
	return t.Execute(w, i)
}

// WriteBytes writes the bytes to the WriteCloser provided by the function
func WriteBytes(b []byte, wr func() io.WriteCloser) error {
	w := wr()
	defer func() {
		if err := w.Close(); err != nil {
			log.Fatal(err)
		}
	}()
	_, err := w.Write(b)
	return err
}
