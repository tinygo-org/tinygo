// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package testing

import (
	"reflect"
)

func TestCleanup(t *T) {
	var cleanups []int
	t.Run("test", func(t *T) {
		t.Cleanup(func() { cleanups = append(cleanups, 1) })
		t.Cleanup(func() { cleanups = append(cleanups, 2) })
	})
	if got, want := cleanups, []int{2, 1}; !reflect.DeepEqual(got, want) {
		t.Errorf("unexpected cleanup record; got %v want %v", got, want)
	}
}

func TestRunCleanup(t *T) {
	outerCleanup := 0
	innerCleanup := 0
	t.Run("test", func(t *T) {
		t.Cleanup(func() { outerCleanup++ })
		t.Run("x", func(t *T) {
			t.Cleanup(func() { innerCleanup++ })
		})
	})
	if innerCleanup != 1 {
		t.Errorf("unexpected inner cleanup count; got %d want 1", innerCleanup)
	}
	if outerCleanup != 1 {
		t.Errorf("unexpected outer cleanup count; got %d want 0", outerCleanup)
	}
}

func TestCleanupParallelSubtests(t *T) {
	ranCleanup := 0
	t.Run("test", func(t *T) {
		t.Cleanup(func() { ranCleanup++ })
		t.Run("x", func(t *T) {
			t.Parallel()
			if ranCleanup > 0 {
				t.Error("outer cleanup ran before parallel subtest")
			}
		})
	})
	if ranCleanup != 1 {
		t.Errorf("unexpected cleanup count; got %d want 1", ranCleanup)
	}
}

func TestNestedCleanup(t *T) {
	ranCleanup := 0
	t.Run("test", func(t *T) {
		t.Cleanup(func() {
			if ranCleanup != 2 {
				t.Errorf("unexpected cleanup count in first cleanup: got %d want 2", ranCleanup)
			}
			ranCleanup++
		})
		t.Cleanup(func() {
			if ranCleanup != 0 {
				t.Errorf("unexpected cleanup count in second cleanup: got %d want 0", ranCleanup)
			}
			ranCleanup++
			t.Cleanup(func() {
				if ranCleanup != 1 {
					t.Errorf("unexpected cleanup count in nested cleanup: got %d want 1", ranCleanup)
				}
				ranCleanup++
			})
		})
	})
	if ranCleanup != 3 {
		t.Errorf("unexpected cleanup count: got %d want 3", ranCleanup)
	}
}
