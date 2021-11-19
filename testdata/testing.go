package main

// TODO: also test the verbose version.

import (
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

func main() {
	m := testing.MainStart(testdeps{}, tests, benchmarks, examples)
	exitcode := m.Run()
	if exitcode != 0 {
		println("exitcode:", exitcode)
	}
}

type testdeps struct{}

func (testdeps) MatchString(pat, str string) (bool, error) { return true, nil }
