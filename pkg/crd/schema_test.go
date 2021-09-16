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

package crd

import (
	"reflect"
	"testing"
)

type alwaysApplyBefore string

func (a alwaysApplyBefore) ApplyFirst() {}

func (a alwaysApplyBefore) ApplyBefore(v interface{}) bool {
	return true
}

type applyBefore string

func (a applyBefore) ApplyFirst() {}

func (a applyBefore) ApplyBefore(v interface{}) bool {
	switch v.(type) {
	case alwaysApplyBefore:
		return false
	default:
		return true
	}
}

type neverApplyBefore string

func (a neverApplyBefore) ApplyFirst() {}

func (a neverApplyBefore) ApplyBefore(v interface{}) bool {
	return false
}

type applyFirst string

func (a applyFirst) ApplyFirst() {}

func TestSortByApplyBefore(t *testing.T) {
	values := []interface{}{
		neverApplyBefore("baz"),
		applyFirst("test"),
		applyBefore("bar"),
		alwaysApplyBefore("foo"),
	}
	expected := []interface{}{
		alwaysApplyBefore("foo"),
		applyBefore("bar"),
		neverApplyBefore("baz"),
		applyFirst("test"),
	}
	sorted := sortByApplyBefore(values)
	if !reflect.DeepEqual(expected, sorted) {
		t.Fatalf("expected values to equal %+v, got: %+v", expected, sorted)
	}
}
