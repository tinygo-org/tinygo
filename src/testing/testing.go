// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// This file has been modified for use by the TinyGo compiler.
// src: https://github.com/golang/go/blob/61bb56ad/src/testing/testing.go

// Package testing provides support for automated testing of Go packages.
package testing

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

// Testing flags.
var (
	flagVerbose   bool
	flagShort     bool
	flagRunRegexp string
)

var initRan bool

// Init registers testing flags. It has no effect if it has already run.
func Init() {
	if initRan {
		return
	}
	initRan = true

	flag.BoolVar(&flagVerbose, "test.v", false, "verbose: print additional output")
	flag.BoolVar(&flagShort, "test.short", false, "short: run smaller test suite to save time")
	flag.StringVar(&flagRunRegexp, "test.run", "", "run: regexp of tests to run")

	initBenchmarkFlags()
}

// common holds the elements common between T and B and
// captures common methods such as Errorf.
type common struct {
	output   bytes.Buffer
	indent   string
	ran      bool     // Test or benchmark (or one of its subtests) was executed.
	failed   bool     // Test or benchmark has failed.
	skipped  bool     // Test of benchmark has been skipped.
	cleanups []func() // optional functions to be called at the end of the test
	finished bool     // Test function has completed.

	hasSub bool // TODO: should be atomic

	parent   *common
	level    int       // Nesting depth of test or benchmark.
	name     string    // Name of test or benchmark.
	start    time.Time // Time test or benchmark started
	duration time.Duration

	tempDir    string
	tempDirErr error
	tempDirSeq int32
}

// Short reports whether the -test.short flag is set.
func Short() bool {
	return flagShort
}

// CoverMode reports what the test coverage mode is set to.
//
// Test coverage is not supported; this returns the empty string.
func CoverMode() string {
	return ""
}

// Verbose reports whether the -test.v flag is set.
func Verbose() bool {
	return flagVerbose
}

// flushToParent writes c.output to the parent after first writing the header
// with the given format and arguments.
func (c *common) flushToParent(testName, format string, args ...interface{}) {
	if c.parent == nil {
		// The fake top-level test doesn't want a FAIL or PASS banner.
		// Not quite sure how this works upstream.
		c.output.WriteTo(os.Stdout)
	} else {
		fmt.Fprintf(&c.parent.output, format, args...)
		c.output.WriteTo(&c.parent.output)
	}
}

// fmtDuration returns a string representing d in the form "87.00s".
func fmtDuration(d time.Duration) string {
	return fmt.Sprintf("%.2fs", d.Seconds())
}

// TB is the interface common to T and B.
type TB interface {
	Cleanup(func())
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Fail()
	FailNow()
	Failed() bool
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	Helper()
	Log(args ...interface{})
	Logf(format string, args ...interface{})
	Name() string
	Setenv(key, value string)
	Skip(args ...interface{})
	SkipNow()
	Skipf(format string, args ...interface{})
	Skipped() bool
	TempDir() string
}

var _ TB = (*T)(nil)
var _ TB = (*B)(nil)

// T is a type passed to Test functions to manage test state and support formatted test logs.
// Logs are accumulated during execution and dumped to standard output when done.
type T struct {
	common
	context *testContext // For running tests and subtests.
}

// Name returns the name of the running test or benchmark.
func (c *common) Name() string {
	return c.name
}

func (c *common) setRan() {
	if c.parent != nil {
		c.parent.setRan()
	}
	c.ran = true
}

// Fail marks the function as having failed but continues execution.
func (c *common) Fail() {
	c.failed = true
}

// Failed reports whether the function has failed.
func (c *common) Failed() bool {
	failed := c.failed
	return failed
}

// FailNow marks the function as having failed and stops its execution
// by calling runtime.Goexit (which then runs all deferred calls in the
// current goroutine).
func (c *common) FailNow() {
	c.Fail()

	c.finished = true
	c.Error("FailNow is incomplete, requires runtime.Goexit()")
}

// log generates the output.
func (c *common) log(s string) {
	// This doesn't print the same as in upstream go, but works for now.
	if len(s) != 0 && s[len(s)-1] == '\n' {
		s = s[:len(s)-1]
	}
	lines := strings.Split(s, "\n")
	// First line.
	c.output.WriteString(c.indent)
	c.output.WriteString("    ") // 4 spaces
	c.output.WriteString(lines[0])
	c.output.WriteByte('\n')
	// More lines.
	for _, line := range lines[1:] {
		c.output.WriteString(c.indent)
		c.output.WriteString("        ") // 8 spaces
		c.output.WriteString(line)
		c.output.WriteByte('\n')
	}
}

// Log formats its arguments using default formatting, analogous to Println,
// and records the text in the error log. For tests, the text will be printed only if
// the test fails or the -test.v flag is set. For benchmarks, the text is always
// printed to avoid having performance depend on the value of the -test.v flag.
func (c *common) Log(args ...interface{}) { c.log(fmt.Sprintln(args...)) }

// Logf formats its arguments according to the format, analogous to Printf, and
// records the text in the error log. A final newline is added if not provided. For
// tests, the text will be printed only if the test fails or the -test.v flag is
// set. For benchmarks, the text is always printed to avoid having performance
// depend on the value of the -test.v flag.
func (c *common) Logf(format string, args ...interface{}) { c.log(fmt.Sprintf(format, args...)) }

// Error is equivalent to Log followed by Fail.
func (c *common) Error(args ...interface{}) {
	c.log(fmt.Sprintln(args...))
	c.Fail()
}

// Errorf is equivalent to Logf followed by Fail.
func (c *common) Errorf(format string, args ...interface{}) {
	c.log(fmt.Sprintf(format, args...))
	c.Fail()
}

// Fatal is equivalent to Log followed by FailNow.
func (c *common) Fatal(args ...interface{}) {
	c.log(fmt.Sprintln(args...))
	c.FailNow()
}

// Fatalf is equivalent to Logf followed by FailNow.
func (c *common) Fatalf(format string, args ...interface{}) {
	c.log(fmt.Sprintf(format, args...))
	c.FailNow()
}

// Skip is equivalent to Log followed by SkipNow.
func (c *common) Skip(args ...interface{}) {
	c.log(fmt.Sprintln(args...))
	c.SkipNow()
}

// Skipf is equivalent to Logf followed by SkipNow.
func (c *common) Skipf(format string, args ...interface{}) {
	c.log(fmt.Sprintf(format, args...))
	c.SkipNow()
}

// SkipNow marks the test as having been skipped and stops its execution
// by calling runtime.Goexit.
func (c *common) SkipNow() {
	c.skip()
	c.finished = true
	c.Error("SkipNow is incomplete, requires runtime.Goexit()")
}

func (c *common) skip() {
	c.skipped = true
}

// Skipped reports whether the test was skipped.
func (c *common) Skipped() bool {
	return c.skipped
}

// Helper is not implemented, it is only provided for compatibility.
func (c *common) Helper() {
	// Unimplemented.
}

// Cleanup registers a function to be called when the test (or subtest) and all its
// subtests complete. Cleanup functions will be called in last added,
// first called order.
func (c *common) Cleanup(f func()) {
	c.cleanups = append(c.cleanups, f)
}

// TempDir returns a temporary directory for the test to use.
// The directory is automatically removed by Cleanup when the test and
// all its subtests complete.
// Each subsequent call to t.TempDir returns a unique directory;
// if the directory creation fails, TempDir terminates the test by calling Fatal.
func (c *common) TempDir() string {
	// Use a single parent directory for all the temporary directories
	// created by a test, each numbered sequentially.
	var nonExistent bool
	if c.tempDir == "" { // Usually the case with js/wasm
		nonExistent = true
	} else {
		_, err := os.Stat(c.tempDir)
		nonExistent = errors.Is(err, fs.ErrNotExist)
		if err != nil && !nonExistent {
			c.Fatalf("TempDir: %v", err)
		}
	}

	if nonExistent {
		c.Helper()

		// Drop unusual characters (such as path separators or
		// characters interacting with globs) from the directory name to
		// avoid surprising os.MkdirTemp behavior.
		mapper := func(r rune) rune {
			if r < utf8.RuneSelf {
				const allowed = "!#$%&()+,-.=@^_{}~ "
				if '0' <= r && r <= '9' ||
					'a' <= r && r <= 'z' ||
					'A' <= r && r <= 'Z' {
					return r
				}
				if strings.ContainsRune(allowed, r) {
					return r
				}
			} else if unicode.IsLetter(r) || unicode.IsNumber(r) {
				return r
			}
			return -1
		}
		pattern := strings.Map(mapper, c.Name())
		c.tempDir, c.tempDirErr = os.MkdirTemp("", pattern)
		if c.tempDirErr == nil {
			c.Cleanup(func() {
				if err := os.RemoveAll(c.tempDir); err != nil {
					c.Errorf("TempDir RemoveAll cleanup: %v", err)
				}
			})
		}
	}

	if c.tempDirErr != nil {
		c.Fatalf("TempDir: %v", c.tempDirErr)
	}
	seq := c.tempDirSeq
	c.tempDirSeq++
	dir := fmt.Sprintf("%s%c%03d", c.tempDir, os.PathSeparator, seq)
	if err := os.Mkdir(dir, 0777); err != nil {
		c.Fatalf("TempDir: %v", err)
	}
	return dir
}

// Setenv calls os.Setenv(key, value) and uses Cleanup to
// restore the environment variable to its original value
// after the test.
func (c *common) Setenv(key, value string) {
	prevValue, ok := os.LookupEnv(key)

	if err := os.Setenv(key, value); err != nil {
		c.Fatalf("cannot set environment variable: %v", err)
	}

	if ok {
		c.Cleanup(func() {
			os.Setenv(key, prevValue)
		})
	} else {
		c.Cleanup(func() {
			os.Unsetenv(key)
		})
	}
}

// runCleanup is called at the end of the test.
func (c *common) runCleanup() {
	for {
		var cleanup func()
		if len(c.cleanups) > 0 {
			last := len(c.cleanups) - 1
			cleanup = c.cleanups[last]
			c.cleanups = c.cleanups[:last]
		}
		if cleanup == nil {
			return
		}
		cleanup()
	}
}

// Parallel is not implemented, it is only provided for compatibility.
func (t *T) Parallel() {
	// Unimplemented.
}

// InternalTest is a reference to a test that should be called during a test suite run.
type InternalTest struct {
	Name string
	F    func(*T)
}

func tRunner(t *T, fn func(t *T)) {
	defer func() {
		t.runCleanup()
	}()

	// Run the test.
	t.start = time.Now()
	fn(t)
	t.duration += time.Since(t.start) // TODO: capture cleanup time, too.

	t.report() // Report after all subtests have finished.
	if t.parent != nil && !t.hasSub {
		t.setRan()
	}
}

// Run runs f as a subtest of t called name. It waits until the subtest is finished
// and returns whether the subtest succeeded.
func (t *T) Run(name string, f func(t *T)) bool {
	t.hasSub = true
	testName, ok, _ := t.context.match.fullName(&t.common, name)
	if !ok {
		return true
	}

	// Create a subtest.
	sub := T{
		common: common{
			name:   testName,
			parent: &t.common,
			level:  t.level + 1,
		},
		context: t.context,
	}
	if t.level > 0 {
		sub.indent = sub.indent + "    "
	}
	if flagVerbose {
		fmt.Fprintf(&t.output, "=== RUN   %s\n", sub.name)
	}

	tRunner(&sub, f)
	return !sub.failed
}

// testContext holds all fields that are common to all tests. This includes
// synchronization primitives to run at most *parallel tests.
type testContext struct {
	match *matcher
}

func newTestContext(m *matcher) *testContext {
	return &testContext{
		match: m,
	}
}

// M is a test suite.
type M struct {
	// tests is a list of the test names to execute
	Tests      []InternalTest
	Benchmarks []InternalBenchmark

	deps testDeps

	// value to pass to os.Exit, the outer test func main
	// harness calls os.Exit with this code. See #34129.
	exitCode int
}

type testDeps interface {
	MatchString(pat, str string) (bool, error)
}

// Run runs the tests. It returns an exit code to pass to os.Exit.
func (m *M) Run() (code int) {
	defer func() {
		code = m.exitCode
	}()

	if !flag.Parsed() {
		flag.Parse()
	}

	testRan, testOk := runTests(m.deps.MatchString, m.Tests)
	if !testRan && *matchBenchmarks == "" {
		fmt.Fprintln(os.Stderr, "testing: warning: no tests to run")
	}
	if !testOk || !runBenchmarks(m.deps.MatchString, m.Benchmarks) {
		fmt.Println("FAIL")
		m.exitCode = 1
	} else {
		fmt.Println("PASS")
		m.exitCode = 0
	}
	return
}

func runTests(matchString func(pat, str string) (bool, error), tests []InternalTest) (ran, ok bool) {
	ok = true

	ctx := newTestContext(newMatcher(matchString, flagRunRegexp, "-test.run"))
	t := &T{
		context: ctx,
	}

	tRunner(t, func(t *T) {
		for _, test := range tests {
			t.Run(test.Name, test.F)
			ok = ok && !t.Failed()
		}
	})

	return t.ran, ok
}

func (t *T) report() {
	dstr := fmtDuration(t.duration)
	format := t.indent + "--- %s: %s (%s)\n"
	if t.Failed() {
		if t.parent != nil {
			t.parent.failed = true
		}
		t.flushToParent(t.name, format, "FAIL", t.name, dstr)
	} else if flagVerbose {
		if t.Skipped() {
			t.flushToParent(t.name, format, "SKIP", t.name, dstr)
		} else {
			t.flushToParent(t.name, format, "PASS", t.name, dstr)
		}
	}
}

// AllocsPerRun returns the average number of allocations during calls to f.
// Although the return value has type float64, it will always be an integral
// value.
//
// Not implemented.
func AllocsPerRun(runs int, f func()) (avg float64) {
	f()
	for i := 0; i < runs; i++ {
		f()
	}
	return 0
}

type InternalExample struct {
	Name      string
	F         func()
	Output    string
	Unordered bool
}

// MainStart is meant for use by tests generated by 'go test'.
// It is not meant to be called directly and is not subject to the Go 1 compatibility document.
// It may change signature from release to release.
func MainStart(deps interface{}, tests []InternalTest, benchmarks []InternalBenchmark, fuzzTargets []InternalFuzzTarget, examples []InternalExample) *M {
	Init()
	return &M{
		Tests:      tests,
		Benchmarks: benchmarks,
		deps:       deps.(testDeps),
	}
}
