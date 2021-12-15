// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package reflect_test

import (
	"math"
	. "reflect"
	"testing"
)

type Basic struct {
	x int
	y float32
}

type NotBasic Basic

type DeepEqualTest struct {
	a, b interface{}
	eq   bool
}

// Simple functions for DeepEqual tests.
var (
	fn1 func()             // nil.
	fn2 func()             // nil.
	fn3 = func() { fn1() } // Not nil.
)

type self struct{}

type Loopy interface{}

var loopy1, loopy2 Loopy
var cycleMap1, cycleMap2, cycleMap3 map[string]interface{}

type structWithSelfPtr struct {
	p *structWithSelfPtr
	s string
}

func init() {
	loopy1 = &loopy2
	loopy2 = &loopy1

	cycleMap1 = map[string]interface{}{}
	cycleMap1["cycle"] = cycleMap1
	cycleMap2 = map[string]interface{}{}
	cycleMap2["cycle"] = cycleMap2
	cycleMap3 = map[string]interface{}{}
	cycleMap3["different"] = cycleMap3
}

// Note: all tests involving maps have been commented out because they aren't
// supported yet.
var deepEqualTests = []DeepEqualTest{
	// Equalities
	{nil, nil, true},
	{1, 1, true},
	{int32(1), int32(1), true},
	{0.5, 0.5, true},
	{float32(0.5), float32(0.5), true},
	{"hello", "hello", true},
	{make([]int, 10), make([]int, 10), true},
	{&[3]int{1, 2, 3}, &[3]int{1, 2, 3}, true},
	{Basic{1, 0.5}, Basic{1, 0.5}, true},
	{error(nil), error(nil), true},
	//{map[int]string{1: "one", 2: "two"}, map[int]string{2: "two", 1: "one"}, true},
	{fn1, fn2, true},
	{[]byte{1, 2, 3}, []byte{1, 2, 3}, true},
	{[]MyByte{1, 2, 3}, []MyByte{1, 2, 3}, true},
	{MyBytes{1, 2, 3}, MyBytes{1, 2, 3}, true},

	// Inequalities
	{1, 2, false},
	{int32(1), int32(2), false},
	{0.5, 0.6, false},
	{float32(0.5), float32(0.6), false},
	{"hello", "hey", false},
	{make([]int, 10), make([]int, 11), false},
	{&[3]int{1, 2, 3}, &[3]int{1, 2, 4}, false},
	{Basic{1, 0.5}, Basic{1, 0.6}, false},
	{Basic{1, 0}, Basic{2, 0}, false},
	//{map[int]string{1: "one", 3: "two"}, map[int]string{2: "two", 1: "one"}, false},
	//{map[int]string{1: "one", 2: "txo"}, map[int]string{2: "two", 1: "one"}, false},
	//{map[int]string{1: "one"}, map[int]string{2: "two", 1: "one"}, false},
	//{map[int]string{2: "two", 1: "one"}, map[int]string{1: "one"}, false},
	{nil, 1, false},
	{1, nil, false},
	{fn1, fn3, false},
	{fn3, fn3, false},
	{[][]int{{1}}, [][]int{{2}}, false},
	{&structWithSelfPtr{p: &structWithSelfPtr{s: "a"}}, &structWithSelfPtr{p: &structWithSelfPtr{s: "b"}}, false},

	// Fun with floating point.
	{math.NaN(), math.NaN(), false},
	{&[1]float64{math.NaN()}, &[1]float64{math.NaN()}, false},
	{&[1]float64{math.NaN()}, self{}, true},
	{[]float64{math.NaN()}, []float64{math.NaN()}, false},
	{[]float64{math.NaN()}, self{}, true},
	//{map[float64]float64{math.NaN(): 1}, map[float64]float64{1: 2}, false},
	//{map[float64]float64{math.NaN(): 1}, self{}, true},

	// Nil vs empty: not the same.
	{[]int{}, []int(nil), false},
	{[]int{}, []int{}, true},
	{[]int(nil), []int(nil), true},
	//{map[int]int{}, map[int]int(nil), false},
	//{map[int]int{}, map[int]int{}, true},
	//{map[int]int(nil), map[int]int(nil), true},

	// Mismatched types
	{1, 1.0, false},
	{int32(1), int64(1), false},
	{0.5, "hello", false},
	{[]int{1, 2, 3}, [3]int{1, 2, 3}, false},
	{&[3]interface{}{1, 2, 4}, &[3]interface{}{1, 2, "s"}, false},
	{Basic{1, 0.5}, NotBasic{1, 0.5}, false},
	{map[uint]string{1: "one", 2: "two"}, map[int]string{2: "two", 1: "one"}, false},
	{[]byte{1, 2, 3}, []MyByte{1, 2, 3}, false},
	{[]MyByte{1, 2, 3}, MyBytes{1, 2, 3}, false},
	{[]byte{1, 2, 3}, MyBytes{1, 2, 3}, false},

	// Possible loops.
	{&loopy1, &loopy1, true},
	{&loopy1, &loopy2, true},
	//{&cycleMap1, &cycleMap2, true},
	//{&cycleMap1, &cycleMap3, false},
}

func TestDeepEqual(t *testing.T) {
	for _, test := range deepEqualTests {
		if test.b == (self{}) {
			test.b = test.a
		}
		if r := DeepEqual(test.a, test.b); r != test.eq {
			t.Errorf("DeepEqual(%#v, %#v) = %v, want %v", test.a, test.b, r, test.eq)
		}
	}
}

type Recursive struct {
	x int
	r *Recursive
}

func TestDeepEqualRecursiveStruct(t *testing.T) {
	a, b := new(Recursive), new(Recursive)
	*a = Recursive{12, a}
	*b = Recursive{12, b}
	if !DeepEqual(a, b) {
		t.Error("DeepEqual(recursive same) = false, want true")
	}
}

type _Complex struct {
	a int
	b [3]*_Complex
	c *string
	d map[float64]float64
}

func TestDeepEqualComplexStruct(t *testing.T) {
	m := make(map[float64]float64)
	stra, strb := "hello", "hello"
	a, b := new(_Complex), new(_Complex)
	*a = _Complex{5, [3]*_Complex{a, b, a}, &stra, m}
	*b = _Complex{5, [3]*_Complex{b, a, a}, &strb, m}
	if !DeepEqual(a, b) {
		t.Error("DeepEqual(complex same) = false, want true")
	}
}

func TestDeepEqualComplexStructInequality(t *testing.T) {
	m := make(map[float64]float64)
	stra, strb := "hello", "helloo" // Difference is here
	a, b := new(_Complex), new(_Complex)
	*a = _Complex{5, [3]*_Complex{a, b, a}, &stra, m}
	*b = _Complex{5, [3]*_Complex{b, a, a}, &strb, m}
	if DeepEqual(a, b) {
		t.Error("DeepEqual(complex different) = true, want false")
	}
}

type MyBytes []byte
type MyByte byte
