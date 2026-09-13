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
	"context"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode"
	"unicode/utf8"
	_ "unsafe"
)

// Testing flags.
var (
	flagVerbose    bool
	flagShort      bool
	flagRunRegexp  string
	flagSkipRegexp string
	flagShuffle    string
	flagCount      int
	flagParallel   int
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
	flag.StringVar(&flagSkipRegexp, "test.skip", "", "skip: regexp of tests to run")
	flag.StringVar(&flagShuffle, "test.shuffle", "off", "shuffle: off, on, <numeric-seed>")

	flag.IntVar(&flagCount, "test.count", 1, "run each test or benchmark `count` times")
	flag.IntVar(&flagParallel, "test.parallel", runtime.NumCPU(), "run at most `n` tests in parallel")

	initBenchmarkFlags()
}

// common holds the elements common between T and B and
// captures common methods such as Errorf.
type common struct {
	mu       sync.Mutex
	output   *logger
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

	cleanupStarted atomic.Bool

	tempDir    string
	tempDirErr error
	tempDirSeq int32

	ctx       context.Context
	cancelCtx context.CancelFunc

	barrier         chan struct{}
	signal          chan bool
	sub             []*T
	isParallel      bool
	parallelRunning bool
	denyParallel    string
}

var testOutputMu sync.Mutex
var errNilPanicOrGoexit = errors.New("test executed panic(nil) or runtime.Goexit")

type logger struct {
	mu          sync.Mutex
	logToStdout bool
	b           bytes.Buffer
}

func (l *logger) Write(p []byte) (int, error) {
	if l.logToStdout {
		testOutputMu.Lock()
		defer testOutputMu.Unlock()
		l.mu.Lock()
		defer l.mu.Unlock()
		return os.Stdout.Write(p)
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.b.Write(p)
}

func (l *logger) writeToStdout() (int64, error) {
	testOutputMu.Lock()
	defer testOutputMu.Unlock()
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.logToStdout {
		// We've already been logging to stdout; nothing to do.
		return 0, nil
	}
	return l.b.WriteTo(os.Stdout)
}

func (l *logger) Len() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.b.Len()
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

// String constant that is being set when running a test.
var testBinary string

// Testing returns whether the program was compiled as a test, using "tinygo
// test". It returns false when built using "tinygo build", "tinygo flash", etc.
func Testing() bool {
	return testBinary == "1"
}

// flushToParent writes c.output to the parent after first writing the header
// with the given format and arguments.
func (c *common) flushToParent(testName, format string, args ...any) {
	testOutputMu.Lock()
	defer testOutputMu.Unlock()

	c.output.mu.Lock()
	defer c.output.mu.Unlock()
	if c.parent == nil {
		// The fake top-level test doesn't want a FAIL or PASS banner.
		// Not quite sure how this works upstream.
		if !c.output.logToStdout {
			c.output.b.WriteTo(os.Stdout)
		}
	} else {
		c.parent.output.mu.Lock()
		defer c.parent.output.mu.Unlock()
		if c.parent.output.logToStdout {
			fmt.Fprintf(os.Stdout, format, args...)
			if !c.output.logToStdout {
				c.output.b.WriteTo(os.Stdout)
			}
		} else {
			fmt.Fprintf(&c.parent.output.b, format, args...)
			if !c.output.logToStdout {
				c.output.b.WriteTo(&c.parent.output.b)
			}
		}
	}
}

// fmtDuration returns a string representing d in the form "87.00s".
func fmtDuration(d time.Duration) string {
	return fmt.Sprintf("%.2fs", d.Seconds())
}

// TB is the interface common to T and B.
type TB interface {
	Cleanup(func())
	Context() context.Context
	Error(args ...any)
	Errorf(format string, args ...any)
	Fail()
	FailNow()
	Failed() bool
	Fatal(args ...any)
	Fatalf(format string, args ...any)
	Helper()
	Log(args ...any)
	Logf(format string, args ...any)
	Name() string
	Setenv(key, value string)
	Skip(args ...any)
	SkipNow()
	Skipf(format string, args ...any)
	Skipped() bool
	TempDir() string
}

var _ TB = (*T)(nil)
var _ TB = (*B)(nil)

// T is a type passed to Test functions to manage test state and support formatted test logs.
// Logs are accumulated during execution and dumped to standard output when done.
type T struct {
	common
	context    *testContext // For running tests and subtests.
	isSynctest bool
}

// Name returns the name of the running test or benchmark.
func (c *common) Name() string {
	return c.name
}

func (c *common) setRan() {
	if c.parent != nil {
		c.parent.setRan()
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ran = true
}

// Fail marks the function as having failed but continues execution.
func (c *common) Fail() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.failed = true
}

// Failed reports whether the function has failed.
func (c *common) Failed() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.failed
}

// FailNow marks the function as having failed and stops its execution
// by calling runtime.Goexit (which then runs all deferred calls in the
// current goroutine).
func (c *common) FailNow() {
	c.Fail()
	c.mu.Lock()
	c.finished = true
	c.mu.Unlock()
	if supportsRecover() {
		runtime.Goexit()
	}
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
	fmt.Fprintf(c.output, "%s    %s\n", c.indent, lines[0])
	// More lines.
	for _, line := range lines[1:] {
		fmt.Fprintf(c.output, "%s        %s\n", c.indent, line)
	}
}

// Log formats its arguments using default formatting, analogous to Println,
// and records the text in the error log. For tests, the text will be printed only if
// the test fails or the -test.v flag is set. For benchmarks, the text is always
// printed to avoid having performance depend on the value of the -test.v flag.
func (c *common) Log(args ...any) { c.log(fmt.Sprintln(args...)) }

// Logf formats its arguments according to the format, analogous to Printf, and
// records the text in the error log. A final newline is added if not provided. For
// tests, the text will be printed only if the test fails or the -test.v flag is
// set. For benchmarks, the text is always printed to avoid having performance
// depend on the value of the -test.v flag.
func (c *common) Logf(format string, args ...any) { c.log(fmt.Sprintf(format, args...)) }

// Error is equivalent to Log followed by Fail.
func (c *common) Error(args ...any) {
	c.log(fmt.Sprintln(args...))
	c.Fail()
}

// Errorf is equivalent to Logf followed by Fail.
func (c *common) Errorf(format string, args ...any) {
	c.log(fmt.Sprintf(format, args...))
	c.Fail()
}

// Fatal is equivalent to Log followed by FailNow.
func (c *common) Fatal(args ...any) {
	c.log(fmt.Sprintln(args...))
	c.FailNow()
}

// Fatalf is equivalent to Logf followed by FailNow.
func (c *common) Fatalf(format string, args ...any) {
	c.log(fmt.Sprintf(format, args...))
	c.FailNow()
}

// Skip is equivalent to Log followed by SkipNow.
func (c *common) Skip(args ...any) {
	c.log(fmt.Sprintln(args...))
	c.SkipNow()
}

// Skipf is equivalent to Logf followed by SkipNow.
func (c *common) Skipf(format string, args ...any) {
	c.log(fmt.Sprintf(format, args...))
	c.SkipNow()
}

// SkipNow marks the test as having been skipped and stops its execution
// by calling runtime.Goexit.
func (c *common) SkipNow() {
	c.skip()
	c.mu.Lock()
	c.finished = true
	c.mu.Unlock()
	if supportsRecover() {
		runtime.Goexit()
	}
	c.Error("SkipNow is incomplete, requires runtime.Goexit()")
}

func (c *common) skip() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.skipped = true
}

//go:linkname supportsRecover runtime.supportsRecover
func supportsRecover() bool

// Skipped reports whether the test was skipped.
func (c *common) Skipped() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
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

// Context returns a context that is canceled just before
// Cleanup-registered functions are called.
//
// Cleanup functions can wait for any resources
// that shut down on [context.Context.Done] before the test or benchmark completes.
func (c *common) Context() context.Context {
	return c.ctx
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

func (t *T) Setenv(key, value string) {
	t.checkParallel("t.Setenv")
	t.common.Setenv(key, value)
}

// Chdir calls os.Chdir(dir) and uses Cleanup to restore the current
// working directory to its original value after the test. On Unix, it
// also sets PWD environment variable for the duration of the test.
//
// Because Chdir affects the whole process, it cannot be used
// in parallel tests or tests with parallel ancestors.
func (c *common) Chdir(dir string) {
	// Note: function copied from the Go 1.24.0 source tree.

	oldwd, err := os.Open(".")
	if err != nil {
		c.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		c.Fatal(err)
	}
	// On POSIX platforms, PWD represents “an absolute pathname of the
	// current working directory.” Since we are changing the working
	// directory, we should also set or update PWD to reflect that.
	switch runtime.GOOS {
	case "windows", "plan9":
		// Windows and Plan 9 do not use the PWD variable.
	default:
		if !filepath.IsAbs(dir) {
			dir, err = os.Getwd()
			if err != nil {
				c.Fatal(err)
			}
		}
		c.Setenv("PWD", dir)
	}
	c.Cleanup(func() {
		err := oldwd.Chdir()
		oldwd.Close()
		if err != nil {
			// It's not safe to continue with tests if we can't
			// get back to the original working directory. Since
			// we are holding a dirfd, this is highly unlikely.
			panic("testing.Chdir: " + err.Error())
		}
	})
}

func (t *T) Chdir(dir string) {
	t.checkParallel("t.Chdir")
	t.common.Chdir(dir)
}

func parallelConflict(op string) string {
	return "testing: test using " + op + " can not use t.Parallel"
}

// checkParallel is called by testing/cryptotest.SetGlobalRandom.
func checkParallel(t *T) {
	t.checkParallel("cryptotest.SetGlobalRandom")
}

func (t *T) checkParallel(op string) {
	for common := &t.common; common != nil; common = common.parent {
		common.mu.Lock()
		parallel := common.isParallel
		common.mu.Unlock()
		if parallel {
			panic(parallelConflict(op))
		}
	}
	t.mu.Lock()
	t.denyParallel = op
	t.mu.Unlock()
}

// runCleanup is called at the end of the test.
func (c *common) runCleanup() {
	c.cleanupStarted.Store(true)
	if c.cancelCtx != nil {
		c.cancelCtx()
		c.cancelCtx = nil
	}

	// If a cleanup exits early, continue with the remaining cleanups.
	defer func() {
		if len(c.cleanups) != 0 {
			c.runCleanup()
		}
	}()

	for len(c.cleanups) != 0 {
		last := len(c.cleanups) - 1
		cleanup := c.cleanups[last]
		c.cleanups = c.cleanups[:last]
		cleanup()
	}
}

// Parallel signals that this test can run with other parallel tests.
func (t *T) Parallel() {
	if t.isSynctest {
		panic("testing: t.Parallel called inside synctest bubble")
	}
	t.mu.Lock()
	if t.isParallel {
		t.mu.Unlock()
		panic("testing: t.Parallel called multiple times")
	}
	if t.denyParallel != "" {
		op := t.denyParallel
		t.mu.Unlock()
		panic(parallelConflict(op))
	}
	if t.parent == nil {
		t.mu.Unlock()
		return
	}

	t.isParallel = true
	t.parallelRunning = false
	t.mu.Unlock()
	t.duration += time.Since(t.start)
	t.parent.mu.Lock()
	t.parent.sub = append(t.parent.sub, t)
	t.parent.mu.Unlock()
	t.signal <- true
	<-t.parent.barrier
	t.context.waitParallel()
	t.mu.Lock()
	t.parallelRunning = true
	t.mu.Unlock()
	t.start = time.Now()
}

// InternalTest is a reference to a test that should be called during a test suite run.
type InternalTest struct {
	Name string
	F    func(*T)
}

func tRunner(t *T, fn func(t *T)) {
	t.start = time.Now()
	defer func() {
		err := recover()
		signal := true
		t.mu.Lock()
		finished := t.finished
		t.mu.Unlock()
		if !finished && err == nil {
			for parent := t.parent; parent != nil; parent = parent.parent {
				parent.mu.Lock()
				parentFinished := parent.finished
				parent.mu.Unlock()
				if parentFinished {
					if !t.isParallel {
						t.Errorf("subtest may have called FailNow on a parent test")
					}
					signal = false
					break
				}
			}
			if signal {
				err = errNilPanicOrGoexit
			}
		}
		reported := false
		defer func() {
			if !reported {
				t.report()
				if t.parent != nil {
					t.mu.Lock()
					hasSub := t.hasSub
					t.mu.Unlock()
					if !hasSub {
						t.setRan()
					}
				}
			}
			t.mu.Lock()
			release := t.isParallel && t.parallelRunning
			if release {
				t.parallelRunning = false
			}
			t.mu.Unlock()
			if release {
				t.context.releaseParallel()
			}
			if t.parent != nil && !t.isSynctest && err == nil {
				t.signal <- signal
			}
			if err != nil {
				panic(err)
			}
		}()

		t.duration += time.Since(t.start) // TODO: capture cleanup time, too.
		if err != nil {
			t.Fail()
			if cleanupErr := recoverCleanup(t); cleanupErr != nil {
				t.Logf("cleanup panicked with %v", cleanupErr)
			}
			return
		}

		t.mu.Lock()
		sub := append([]*T(nil), t.sub...)
		wasRunning := t.parallelRunning
		if len(sub) != 0 && wasRunning {
			t.parallelRunning = false
		}
		t.mu.Unlock()
		if len(sub) != 0 {
			if wasRunning {
				t.context.releaseParallel()
			}
			close(t.barrier)
			for _, sub := range sub {
				<-sub.signal
			}
			if wasRunning {
				t.context.waitParallel()
				t.mu.Lock()
				t.parallelRunning = true
				t.mu.Unlock()
			}
		}
		if cleanupErr := recoverCleanup(t); cleanupErr != nil {
			err = cleanupErr
			t.Fail()
			return
		}
		t.report() // Report after all subtests have finished.
		reported = true
		if t.parent != nil {
			t.mu.Lock()
			hasSub := t.hasSub
			t.mu.Unlock()
			if !hasSub {
				t.setRan()
			}
		}
	}()

	// Run the test.
	fn(t)
	t.mu.Lock()
	t.finished = true
	t.mu.Unlock()
}

func recoverCleanup(t *T) (err any) {
	finished := false
	defer func() {
		err = recover()
		if !finished && err == nil {
			err = errNilPanicOrGoexit
		}
	}()
	t.runCleanup()
	finished = true
	return nil
}

//go:linkname testingSynctestTest testing/synctest.testingSynctestTest
func testingSynctestTest(t *T, f func(*T)) bool {
	if t.cleanupStarted.Load() {
		panic("testing: synctest.Run called during t.Cleanup")
	}

	ctx, cancelCtx := context.WithCancel(context.Background())
	synctestT := T{
		common: common{
			output:    &logger{logToStdout: flagVerbose},
			name:      t.name,
			parent:    &t.common,
			level:     t.level + 1,
			ctx:       ctx,
			cancelCtx: cancelCtx,
		},
		context:    t.context,
		isSynctest: true,
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		tRunner(&synctestT, f)
	}()
	<-done
	return !synctestT.failed
}

// Run runs f as a subtest of t called name. It blocks until the subtest
// finishes or calls Parallel.
func (t *T) Run(name string, f func(t *T)) bool {
	if t.isSynctest {
		panic("testing: t.Run called inside synctest bubble")
	}
	t.mu.Lock()
	t.hasSub = true
	parallelRunning := t.parallelRunning
	t.mu.Unlock()
	testName, ok, _ := t.context.match.fullName(&t.common, name)
	if !ok {
		return true
	}

	// Create a subtest.
	ctx, cancelCtx := context.WithCancel(context.Background())
	sub := T{
		common: common{
			output:          &logger{logToStdout: flagVerbose},
			name:            testName,
			parent:          &t.common,
			level:           t.level + 1,
			ctx:             ctx,
			cancelCtx:       cancelCtx,
			barrier:         make(chan struct{}),
			signal:          make(chan bool, 1),
			parallelRunning: parallelRunning,
		},
		context: t.context,
	}
	if t.level > 0 {
		sub.indent = sub.indent + "    "
	}
	if flagVerbose {
		fmt.Fprintf(t.output, "=== RUN   %s\n", sub.name)
	}

	go func() {
		tRunner(&sub, f)
	}()
	if !<-sub.signal {
		runtime.Goexit()
	}
	return !sub.Failed()
}

// Deadline reports the time at which the test binary will have
// exceeded the timeout specified by the -timeout flag.
//
// The ok result is false if the -timeout flag indicates “no timeout” (0).
// For now tinygo always return 0, false.
//
// Not Implemented.
func (t *T) Deadline() (deadline time.Time, ok bool) {
	if t.isSynctest {
		panic("testing: t.Deadline called inside synctest bubble")
	}
	deadline = t.context.deadline
	return deadline, !deadline.IsZero()
}

// testContext holds all fields that are common to all tests. This includes
// synchronization primitives to run at most *parallel tests.
type testContext struct {
	match       *matcher
	deadline    time.Time
	parallelSem chan struct{}
}

func newTestContext(m *matcher) *testContext {
	return &testContext{
		match:       m,
		parallelSem: make(chan struct{}, flagParallel),
	}
}

func (c *testContext) waitParallel() {
	c.parallelSem <- struct{}{}
}

func (c *testContext) releaseParallel() {
	<-c.parallelSem
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

func (m *M) shuffle() error {
	var n int64

	if flagShuffle == "on" {
		n = time.Now().UnixNano()
	} else {
		var err error
		n, err = strconv.ParseInt(flagShuffle, 10, 64)
		if err != nil {
			m.exitCode = 2
			return fmt.Errorf(`testing: -shuffle should be "off", "on", or a valid integer: %v`, err)
		}
	}

	fmt.Println("-test.shuffle", n)
	rng := rand.New(rand.NewSource(n))
	rng.Shuffle(len(m.Tests), func(i, j int) { m.Tests[i], m.Tests[j] = m.Tests[j], m.Tests[i] })
	rng.Shuffle(len(m.Benchmarks), func(i, j int) { m.Benchmarks[i], m.Benchmarks[j] = m.Benchmarks[j], m.Benchmarks[i] })
	return nil
}

// Run runs the tests. It returns an exit code to pass to os.Exit.
func (m *M) Run() (code int) {
	defer func() {
		code = m.exitCode
	}()

	if !flag.Parsed() {
		flag.Parse()
	}
	if flagParallel < 1 {
		fmt.Fprintln(os.Stderr, "testing: -test.parallel must be at least 1")
		m.exitCode = 2
		return
	}

	if flagShuffle != "off" {
		if err := m.shuffle(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
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

	for i := 0; i < flagCount; i++ {
		ctx := newTestContext(newMatcher(matchString, flagRunRegexp, "-test.run", flagSkipRegexp))
		runCtx, cancelCtx := context.WithCancel(context.Background())
		t := &T{
			common: common{
				output:    &logger{logToStdout: flagVerbose},
				ctx:       runCtx,
				cancelCtx: cancelCtx,
				barrier:   make(chan struct{}),
				signal:    make(chan bool, 1),
			},
			context: ctx,
		}
		tRunner(t, func(t *T) {
			for _, test := range tests {
				t.Run(test.Name, test.F)
			}
		})
		ran = ran || t.ran
		ok = ok && !t.Failed()
	}

	return ran, ok
}

func (t *T) report() {
	dstr := fmtDuration(t.duration)
	format := t.indent + "--- %s: %s (%s)\n"
	if t.Failed() {
		if t.parent != nil {
			t.parent.Fail()
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
	for range runs {
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
func MainStart(deps any, tests []InternalTest, benchmarks []InternalBenchmark, fuzzTargets []InternalFuzzTarget, examples []InternalExample) *M {
	Init()
	return &M{
		Tests:      tests,
		Benchmarks: benchmarks,
		deps:       deps.(testDeps),
	}
}

// A fake regexp matcher.
// Inflexible, but saves 50KB of flash and 50KB of RAM per -size full,
// and lets tests pass on cortex-m.
func fakeMatchString(pat, str string) (bool, error) {
	if pat == ".*" {
		return true, nil
	}
	matched := strings.Contains(str, pat)
	return matched, nil
}
