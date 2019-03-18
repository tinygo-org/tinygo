package testing

import (
	"fmt"
)

// T is a test helper.
type T struct {
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
	for _, test := range m.Tests {
		t := &T{}
		test.Func(t)
	}
	// TODO: detect failures and return a failing exit code
	// Right now we can't handle one anyway so it doesn't matter much
	return 0
}

// Fatal is equivalent to Log followed by FailNow
func (t *T) Fatal(args ...string) {
	// This doesn't print the same as in upstream go, but works good enough
	fmt.Print(args)
	//t.FailNow()
}

/*
func (t *T) FailNow() {
	// This fails with
Undefined symbols for architecture x86_64:
"_syscall.Exit", referenced from:
  _main in main.o
ld: symbol(s) not found for architecture x86_64
clang: error: linker command failed with exit code 1 (use -v to see invocation)
error: failed to link /var/folders/_t/fbw4wf_s42dfqdfjq7shdfm40000gn/T/tinygo382510652/main: exit status 1
	os.Exit(12)
}
*/
