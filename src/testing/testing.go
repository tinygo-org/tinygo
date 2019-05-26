package testing

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

// T is a test helper.
type T struct {
	name   string
	output io.Writer

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
			name:   test.Name,
			output: &bytes.Buffer{},
		}

		fmt.Printf("=== RUN   %s\n", test.Name)
		test.Func(t)

		if t.failed == 0 {
			fmt.Printf("--- PASS: %s\n", test.Name)
		} else {
			fmt.Printf("--- FAIL: %s\n", test.Name)
		}
		fmt.Println(t.output)

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

// Error is equivalent to Log followed by Fail
func (t *T) Error(args ...interface{}) {
	// This doesn't print the same as in upstream go, but works good enough
	// TODO: buffer test output like go does
	fmt.Fprintf(t.output, "\t")
	fmt.Fprintln(t.output, args...)
	t.Fail()
}

func (t *T) Fail() {
	t.failed = 1
}
