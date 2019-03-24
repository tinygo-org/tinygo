package testing

import (
	"fmt"
)

// T is a test helper.
type T struct {

	// flags the test as having failed when non-zero
	failed int
}

// TestToCall is a reference to a test that should be called during a test suite run.
type TestToCall struct {
	// Name of the test to call.
	Name string
	// Function reference to the test.
	Func func(*T)
}

// M is a test suite.
type M struct {
	// tests is a list of the test names to execute
	Tests []TestToCall
}

// Run the test suite.
func (m *M) Run() int {
	failures := 0
	for _, test := range m.Tests {
		t := &T{}
		test.Func(t)

		failures += t.failed
	}

	return failures
}

// Fatal is equivalent to Log followed by Fail
func (t *T) Error(args ...string) {
	// This doesn't print the same as in upstream go, but works good enough
	fmt.Print(args)
	t.Fail()
}

func (t *T) Fail() {
	t.failed = 1
}
