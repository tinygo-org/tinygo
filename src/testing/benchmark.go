// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// This file has been modified for use by the TinyGo compiler.

package testing

import (
	"time"
)

var (
	benchTime = benchTimeFlag{d: 1 * time.Second} // changed during test of testing package
)

type benchTimeFlag struct {
	d time.Duration
}

// B is a type passed to Benchmark functions to manage benchmark timing and to
// specify the number of iterations to run.
type B struct {
	common
	hasSub    bool          // TODO: should be in common, and atomic
	start     time.Time     // TODO: should be in common
	duration  time.Duration // TODO: should be in common
	N         int
	benchFunc func(b *B)
	benchTime benchTimeFlag
	timerOn   bool
	result    BenchmarkResult
}

// InternalBenchmark is an internal type but exported because it is cross-package;
// it is part of the implementation of the "go test" command.
type InternalBenchmark struct {
	Name string
	F    func(b *B)
}

// BenchmarkResult contains the results of a benchmark run.
type BenchmarkResult struct {
	N int           // The number of iterations.
	T time.Duration // The total time taken.
}

// NsPerOp returns the "ns/op" metric.
func (r BenchmarkResult) NsPerOp() int64 {
	if r.N <= 0 {
		return 0
	}
	return r.T.Nanoseconds() / int64(r.N)
}

// AllocsPerOp returns the "allocs/op" metric,
// which is calculated as r.MemAllocs / r.N.
func (r BenchmarkResult) AllocsPerOp() int64 {
	return 0 // Dummy version to allow running e.g. golang.org/test/fibo.go
}

// AllocedBytesPerOp returns the "B/op" metric,
// which is calculated as r.MemBytes / r.N.
func (r BenchmarkResult) AllocedBytesPerOp() int64 {
	return 0 // Dummy version to allow running e.g. golang.org/test/fibo.go
}

func (b *B) SetBytes(n int64) {
	panic("testing: unimplemented: B.SetBytes")
}

// ReportAllocs enables malloc statistics for this benchmark.
func (b *B) ReportAllocs() {
	// Dummy version to allow building e.g. golang.org/crypto/...
	panic("testing: unimplemented: B.ReportAllocs")
}

// StartTimer starts timing a test. This function is called automatically
// before a benchmark starts, but it can also be used to resume timing after
// a call to StopTimer.
func (b *B) StartTimer() {
	if !b.timerOn {
		b.start = time.Now()
		b.timerOn = true
	}
}

// StopTimer stops timing a test. This can be used to pause the timer
// while performing complex initialization that you don't
// want to measure.
func (b *B) StopTimer() {
	if b.timerOn {
		b.duration += time.Since(b.start)
		b.timerOn = false
	}
}

// ResetTimer zeroes the elapsed benchmark time.
// It does not affect whether the timer is running.
func (b *B) ResetTimer() {
	if b.timerOn {
		b.start = time.Now()
	}
	b.duration = 0
}

// runN runs a single benchmark for the specified number of iterations.
func (b *B) runN(n int) {
	b.N = n
	b.ResetTimer()
	b.StartTimer()
	b.benchFunc(b)
	b.StopTimer()
}

func min(x, y int64) int64 {
	if x > y {
		return y
	}
	return x
}

func max(x, y int64) int64 {
	if x < y {
		return y
	}
	return x
}

// run1 runs the first iteration of benchFunc. It reports whether more
// iterations of this benchmarks should be run.
func (b *B) run1() bool {
	b.runN(1)
	return !b.hasSub
}

// run executes the benchmark.
func (b *B) run() {
	b.launch()
}

// launch launches the benchmark function. It gradually increases the number
// of benchmark iterations until the benchmark runs for the requested benchtime.
// run1 must have been called on b.
func (b *B) launch() {
	d := b.benchTime.d
	for n := int64(1); !b.failed && b.duration < d && n < 1e9; {
		last := n
		// Predict required iterations.
		goalns := d.Nanoseconds()
		prevIters := int64(b.N)
		prevns := b.duration.Nanoseconds()
		if prevns <= 0 {
			// Round up, to avoid div by zero.
			prevns = 1
		}
		// Order of operations matters.
		// For very fast benchmarks, prevIters ~= prevns.
		// If you divide first, you get 0 or 1,
		// which can hide an order of magnitude in execution time.
		// So multiply first, then divide.
		n = goalns * prevIters / prevns
		// Run more iterations than we think we'll need (1.2x).
		n += n / 5
		// Don't grow too fast in case we had timing errors previously.
		n = min(n, 100*last)
		// Be sure to run at least one more than last time.
		n = max(n, last+1)
		// Don't run more than 1e9 times. (This also keeps n in int range on 32 bit platforms.)
		n = min(n, 1e9)
		b.runN(int(n))
	}
	b.result = BenchmarkResult{b.N, b.duration}
}

// Run benchmarks f as a subbenchmark with the given name. It reports
// true if the subbenchmark succeeded.
//
// A subbenchmark is like any other benchmark. A benchmark that calls Run at
// least once will not be measured itself and will be called once with N=1.
func (b *B) Run(name string, f func(b *B)) bool {
	b.hasSub = true
	sub := &B{
		common:    common{name: name},
		benchFunc: f,
		benchTime: b.benchTime,
	}
	if sub.run1() {
		sub.run()
	}
	b.add(sub.result)
	return !sub.failed
}

// Benchmark benchmarks a single function. It is useful for creating
// custom benchmarks that do not use the "go test" command.
//
// If f calls Run, the result will be an estimate of running all its
// subbenchmarks that don't call Run in sequence in a single benchmark.
func Benchmark(f func(b *B)) BenchmarkResult {
	b := &B{
		benchFunc: f,
		benchTime: benchTime,
	}
	if b.run1() {
		b.run()
	}
	return b.result
}

// add simulates running benchmarks in sequence in a single iteration. It is
// used to give some meaningful results in case func Benchmark is used in
// combination with Run.
func (b *B) add(other BenchmarkResult) {
	r := &b.result
	// The aggregated BenchmarkResults resemble running all subbenchmarks as
	// in sequence in a single benchmark.
	r.N = 1
	r.T += time.Duration(other.NsPerOp())
}

// A PB is used by RunParallel for running parallel benchmarks.
type PB struct {
}

// Next reports whether there are more iterations to execute.
func (pb *PB) Next() bool {
	return false
}

// RunParallel runs a benchmark in parallel.
//
// Not implemented
func (b *B) RunParallel(body func(*PB)) {
	return
}
