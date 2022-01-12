// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// This file has been modified for use by the TinyGo compiler.

package testing

import (
	"flag"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"time"
)

func initBenchmarkFlags() {
	matchBenchmarks = flag.String("test.bench", "", "run only benchmarks matching `regexp`")
	flag.Var(&benchTime, "test.benchtime", "run each benchmark for duration `d`")
}

var (
	matchBenchmarks *string
	benchTime       = benchTimeFlag{d: 1 * time.Second} // changed during test of testing package
)

type benchTimeFlag struct {
	d time.Duration
	n int
}

func (f *benchTimeFlag) String() string {
	if f.n > 0 {
		return fmt.Sprintf("%dx", f.n)
	}
	return time.Duration(f.d).String()
}

func (f *benchTimeFlag) Set(s string) error {
	if strings.HasSuffix(s, "x") {
		n, err := strconv.ParseInt(s[:len(s)-1], 10, 0)
		if err != nil || n <= 0 {
			return fmt.Errorf("invalid count")
		}
		*f = benchTimeFlag{n: int(n)}
		return nil
	}
	d, err := time.ParseDuration(s)
	if err != nil || d <= 0 {
		return fmt.Errorf("invalid duration")
	}
	*f = benchTimeFlag{d: d}
	return nil
}

// InternalBenchmark is an internal type but exported because it is cross-package;
// it is part of the implementation of the "go test" command.
type InternalBenchmark struct {
	Name string
	F    func(b *B)
}

// B is a type passed to Benchmark functions to manage benchmark
// timing and to specify the number of iterations to run.
//
// A benchmark ends when its Benchmark function returns or calls any of the methods
// FailNow, Fatal, Fatalf, SkipNow, Skip, or Skipf. Those methods must be called
// only from the goroutine running the Benchmark function.
// The other reporting methods, such as the variations of Log and Error,
// may be called simultaneously from multiple goroutines.
//
// Like in tests, benchmark logs are accumulated during execution
// and dumped to standard output when done. Unlike in tests, benchmark logs
// are always printed, so as not to hide output whose existence may be
// affecting benchmark results.
type B struct {
	common
	hasSub       bool          // TODO: should be in common, and atomic
	start        time.Time     // TODO: should be in common
	duration     time.Duration // TODO: should be in common
	context      *benchContext
	N            int
	benchFunc    func(b *B)
	bytes        int64
	missingBytes bool // one of the subbenchmarks does not have bytes set.
	benchTime    benchTimeFlag
	timerOn      bool
	result       BenchmarkResult
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

// ResetTimer zeroes the elapsed benchmark time and memory allocation counters
// and deletes user-reported metrics.
func (b *B) ResetTimer() {
	if b.timerOn {
		b.start = time.Now()
	}
	b.duration = 0
}

// SetBytes records the number of bytes processed in a single operation.
// If this is called, the benchmark will report ns/op and MB/s.
func (b *B) SetBytes(n int64) { b.bytes = n }

// ReportAllocs enables malloc statistics for this benchmark.
// It is equivalent to setting -test.benchmem, but it only affects the
// benchmark function that calls ReportAllocs.
func (b *B) ReportAllocs() {
	return // TODO: implement
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
	if ctx := b.context; ctx != nil {
		// Extend maxLen, if needed.
		if n := len(b.name); n > ctx.maxLen {
			ctx.maxLen = n + 8 // Add additional slack to avoid too many jumps in size.
		}
	}
	b.runN(1)
	return !b.hasSub
}

// run executes the benchmark.
func (b *B) run() {
	if b.context != nil {
		// Running go test --test.bench
		b.processBench(b.context) // calls doBench and prints results
	} else {
		// Running func Benchmark.
		b.doBench()
	}
}

func (b *B) doBench() BenchmarkResult {
	// in upstream, this uses a goroutine
	b.launch()
	return b.result
}

// launch launches the benchmark function. It gradually increases the number
// of benchmark iterations until the benchmark runs for the requested benchtime.
// run1 must have been called on b.
func (b *B) launch() {
	// Run the benchmark for at least the specified amount of time.
	if b.benchTime.n > 0 {
		b.runN(b.benchTime.n)
	} else {
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
	}
	b.result = BenchmarkResult{b.N, b.duration, b.bytes}
}

// BenchmarkResult contains the results of a benchmark run.
type BenchmarkResult struct {
	N     int           // The number of iterations.
	T     time.Duration // The total time taken.
	Bytes int64         // Bytes processed in one iteration.
}

// NsPerOp returns the "ns/op" metric.
func (r BenchmarkResult) NsPerOp() int64 {
	if r.N <= 0 {
		return 0
	}
	return r.T.Nanoseconds() / int64(r.N)
}

// mbPerSec returns the "MB/s" metric.
func (r BenchmarkResult) mbPerSec() float64 {
	if r.Bytes <= 0 || r.T <= 0 || r.N <= 0 {
		return 0
	}
	return (float64(r.Bytes) * float64(r.N) / 1e6) / r.T.Seconds()
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

// String returns a summary of the benchmark results.
// It follows the benchmark result line format from
// https://golang.org/design/14313-benchmark-format, not including the
// benchmark name.
// Extra metrics override built-in metrics of the same name.
// String does not include allocs/op or B/op, since those are reported
// by MemString.
func (r BenchmarkResult) String() string {
	buf := new(strings.Builder)
	fmt.Fprintf(buf, "%8d", r.N)

	// Get ns/op as a float.
	ns := float64(r.T.Nanoseconds()) / float64(r.N)
	if ns != 0 {
		buf.WriteByte('\t')
		prettyPrint(buf, ns, "ns/op")
	}

	if mbs := r.mbPerSec(); mbs != 0 {
		fmt.Fprintf(buf, "\t%7.2f MB/s", mbs)
	}
	return buf.String()
}

func prettyPrint(w io.Writer, x float64, unit string) {
	// Print all numbers with 10 places before the decimal point
	// and small numbers with four sig figs. Field widths are
	// chosen to fit the whole part in 10 places while aligning
	// the decimal point of all fractional formats.
	var format string
	switch y := math.Abs(x); {
	case y == 0 || y >= 999.95:
		format = "%10.0f %s"
	case y >= 99.995:
		format = "%12.1f %s"
	case y >= 9.9995:
		format = "%13.2f %s"
	case y >= 0.99995:
		format = "%14.3f %s"
	case y >= 0.099995:
		format = "%15.4f %s"
	case y >= 0.0099995:
		format = "%16.5f %s"
	case y >= 0.00099995:
		format = "%17.6f %s"
	default:
		format = "%18.7f %s"
	}
	fmt.Fprintf(w, format, x, unit)
}

type benchContext struct {
	match *matcher

	maxLen int // The largest recorded benchmark name.
}

func runBenchmarks(matchString func(pat, str string) (bool, error), benchmarks []InternalBenchmark) bool {
	// If no flag was specified, don't run benchmarks.
	if len(*matchBenchmarks) == 0 {
		return true
	}
	ctx := &benchContext{
		match: newMatcher(matchString, *matchBenchmarks, "-test.bench"),
	}
	var bs []InternalBenchmark
	for _, Benchmark := range benchmarks {
		if _, matched, _ := ctx.match.fullName(nil, Benchmark.Name); matched {
			bs = append(bs, Benchmark)
			benchName := Benchmark.Name
			if l := len(benchName); l > ctx.maxLen {
				ctx.maxLen = l
			}
		}
	}
	main := &B{
		common: common{
			name: "Main",
		},
		benchTime: benchTime,
		benchFunc: func(b *B) {
			for _, Benchmark := range bs {
				b.Run(Benchmark.Name, Benchmark.F)
			}
		},
		context: ctx,
	}

	main.runN(1)
	return true
}

// processBench runs bench b and prints the results.
func (b *B) processBench(ctx *benchContext) {
	benchName := b.name

	if ctx != nil {
		fmt.Printf("%-*s\t", ctx.maxLen, benchName)
	}
	r := b.doBench()
	if b.failed {
		// The output could be very long here, but probably isn't.
		// We print it all, regardless, because we don't want to trim the reason
		// the benchmark failed.
		fmt.Printf("--- FAIL: %s\n%s", benchName, "") // b.output)
		return
	}
	if ctx != nil {
		results := r.String()
		fmt.Println(results)
	}
}

// Run benchmarks f as a subbenchmark with the given name. It reports
// true if the subbenchmark succeeded.
//
// A subbenchmark is like any other benchmark. A benchmark that calls Run at
// least once will not be measured itself and will be called once with N=1.
func (b *B) Run(name string, f func(b *B)) bool {
	benchName, ok, partial := b.name, true, false
	if b.context != nil {
		benchName, ok, partial = b.context.match.fullName(&b.common, name)
	}
	if !ok {
		return true
	}
	b.hasSub = true
	sub := &B{
		common: common{
			name:  benchName,
			level: b.level + 1,
		},
		benchFunc: f,
		benchTime: b.benchTime,
		context:   b.context,
	}
	if partial {
		// Partial name match, like -bench=X/Y matching BenchmarkX.
		// Only process sub-benchmarks, if any.
		sub.hasSub = true
	}
	if sub.run1() {
		sub.run()
	}
	b.add(sub.result)
	return !sub.failed
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
	if other.Bytes == 0 {
		// Summing Bytes is meaningless in aggregate if not all subbenchmarks
		// set it.
		b.missingBytes = true
		r.Bytes = 0
	}
	if !b.missingBytes {
		r.Bytes += other.Bytes
	}
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
