// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// This file has been modified for use by the TinyGo compiler.

package testing

// B is a type passed to Benchmark functions to manage benchmark timing and to
// specify the number of iterations to run.
//
// TODO: Implement benchmarks. This struct allows test files containing
// benchmarks to compile and run, but will not run the benchmarks themselves.
type B struct {
	common
	N int
}

type InternalBenchmark struct {
	Name string
	F    func(b *B)
}
