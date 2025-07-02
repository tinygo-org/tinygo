// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file is a copy of src/unique/handle_test.go in upstream Go, but with
// some parts removed that rely on Go runtime implementation details.

package unique

import (
	"fmt"
	"reflect"
	"testing"
)

// Set up special types. Because the internal maps are sharded by type,
// this will ensure that we're not overlapping with other tests.
type testString string
type testIntArray [4]int
type testEface any
type testStringArray [3]string
type testStringStruct struct {
	a string
}
type testStringStructArrayStruct struct {
	s [2]testStringStruct
}
type testStruct struct {
	z float64
	b string
}

func TestHandle(t *testing.T) {
	testHandle[testString](t, "foo")
	testHandle[testString](t, "bar")
	testHandle[testString](t, "")
	testHandle[testIntArray](t, [4]int{7, 77, 777, 7777})
	//testHandle[testEface](t, nil) // requires Go 1.20
	testHandle[testStringArray](t, [3]string{"a", "b", "c"})
	testHandle[testStringStruct](t, testStringStruct{"x"})
	testHandle[testStringStructArrayStruct](t, testStringStructArrayStruct{
		s: [2]testStringStruct{testStringStruct{"y"}, testStringStruct{"z"}},
	})
	testHandle[testStruct](t, testStruct{0.5, "184"})
}

func testHandle[T comparable](t *testing.T, value T) {
	name := reflect.TypeFor[T]().Name()
	t.Run(fmt.Sprintf("%s/%#v", name, value), func(t *testing.T) {
		t.Parallel()

		v0 := Make(value)
		v1 := Make(value)

		if v0.Value() != v1.Value() {
			t.Error("v0.Value != v1.Value")
		}
		if v0.Value() != value {
			t.Errorf("v0.Value not %#v", value)
		}
		if v0 != v1 {
			t.Error("v0 != v1")
		}

		drainMaps(t)
	})
}

// drainMaps ensures that the internal maps are drained.
func drainMaps(t *testing.T) {
	t.Helper()

	globalMapMutex.Lock()
	globalMap = nil
	globalMapMutex.Unlock()
}
