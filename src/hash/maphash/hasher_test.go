//go:build go1.27

// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package maphash

import "testing"

func TestComparableHasher(t *testing.T) {
	var hasher Hasher[string] = ComparableHasher[string]{}
	seed := MakeSeed()

	sum := func(s string) uint64 {
		var h Hash
		h.SetSeed(seed)
		hasher.Hash(&h, s)
		return h.Sum64()
	}

	if sum("foo") != sum("foo") {
		t.Error("ComparableHasher.Hash is not deterministic")
	}
	if sum("foo") == sum("bar") {
		t.Error("ComparableHasher.Hash collides on foo and bar")
	}
	if !hasher.Equal("foo", "foo") || hasher.Equal("foo", "bar") {
		t.Error("ComparableHasher.Equal does not match ==")
	}
}
