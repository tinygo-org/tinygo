package main

import _ "unsafe"

func init() {
	foo()
}

func foo() {
	interp_test_error()
}

// This is a function that always causes an error in interp, for testing.
//
//go:linkname interp_test_error __tinygo_interp_raise_test_error
func interp_test_error()

func main() {
}

// ERROR: # main
// ERROR: {{.*testdata[\\/]errors[\\/]interp\.go}}:10:19: test error
// ERROR:   call void @__tinygo_interp_raise_test_error{{.*}}
// ERROR: {{}}
// ERROR: traceback:
// ERROR: {{.*testdata[\\/]errors[\\/]interp\.go}}:10:19:
// ERROR:   call void @__tinygo_interp_raise_test_error{{.*}}
// ERROR: {{.*testdata[\\/]errors[\\/]interp\.go}}:6:5:
// ERROR:   call void @main.foo{{.*}}
// ERROR: {{.*testdata[\\/]errors}}:
// ERROR:   call void @"main.init#1"{{.*}}
