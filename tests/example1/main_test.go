package main

import (
	"fmt"
	"os"
	"testing" // This is the tinygo testing package
)

func TestFoo(t *testing.T) {
	t.Error("test failed\n")
}

// TODO: change signature to accept a prepopulated test suite
func TestMain() {
	fmt.Println("running tests...")
	m := testing.M{
		Tests: []testing.TestToCall{
			{Name: "TestFoo", Func: TestFoo},
		},
	}
	os.Exit(m.Run())
}
