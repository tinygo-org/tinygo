package testing

import (
	"fmt"
	"os"
)

// T is a test helper.
type T struct {
	name string

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
		t := &T{
			name: test.Name,
		}
		test.Func(t)

		failures += t.failed
	}

	if failures > 0 {
		fmt.Printf("exit status %d\n", failures)
		fmt.Println("FAIL")
	}
	return failures
}

func TestMain(m *M) {
	os.Exit(m.Run())
}

// Fatal is equivalent to Log followed by Fail
func (t *T) Error(args ...interface{}) {
	// This doesn't print the same as in upstream go, but works good enough
	// TODO: buffer test output like go does
	fmt.Printf("--- FAIL: %s\n", t.name)
	fmt.Printf("\t")
	fmt.Println(args...)
	t.Fail()
}

func (t *T) Fail() {
	t.failed = 1
}
