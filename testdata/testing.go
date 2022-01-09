package main

// TODO: also test the verbose version.

import (
	"errors"
	"io"
	"testing"
)

func TestFoo(t *testing.T) {
	t.Log("log Foo.a")
	t.Log("log Foo.b")
}

func TestBar(t *testing.T) {
	t.Log("log Bar")
	t.Log("log g\nh\ni\n")
	t.Run("Bar1", func(t *testing.T) {})
	t.Run("Bar2", func(t *testing.T) {
		t.Log("log Bar2\na\nb\nc")
		t.Error("failed")
		t.Log("after failed")
	})
	t.Run("Bar3", func(t *testing.T) {})
	t.Log("log Bar end")
}

var tests = []testing.InternalTest{
	{"TestFoo", TestFoo},
	{"TestBar", TestBar},
}

var benchmarks = []testing.InternalBenchmark{}

var examples = []testing.InternalExample{}

var errMain = errors.New("testing: unexpected use of func Main")

// matchStringOnly is part of upstream, and is used below to provide a dummy deps to pass to MainStart
// so it can be run with go (tested with go 1.16) to provide a baseline for the regression test.
// See c56cc9b3b57276.  Unfortunately, testdeps is internal, so we can't just use &testdeps.TestDeps{}.
type matchStringOnly func(pat, str string) (bool, error)

func (f matchStringOnly) MatchString(pat, str string) (bool, error)   { return f(pat, str) }
func (f matchStringOnly) StartCPUProfile(w io.Writer) error           { return errMain }
func (f matchStringOnly) StopCPUProfile()                             {}
func (f matchStringOnly) WriteProfileTo(string, io.Writer, int) error { return errMain }
func (f matchStringOnly) ImportPath() string                          { return "" }
func (f matchStringOnly) StartTestLog(io.Writer)                      {}
func (f matchStringOnly) StopTestLog() error                          { return errMain }
func (f matchStringOnly) SetPanicOnExit0(bool)                        {}

func main() {
	m := testing.MainStart(matchStringOnly(nil), tests, benchmarks, examples)
	exitcode := m.Run()
	if exitcode != 0 {
		println("exitcode:", exitcode)
	}
}
