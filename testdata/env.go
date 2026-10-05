package main

import (
	"os"
	"runtime"
)

func main() {
	if expected, ok := os.LookupEnv("EXPECT_CWD"); ok {
		cwd, err := os.Getwd()
		if err != nil || cwd != expected {
			println("cwd:", cwd, "expected:", expected, "PWD:", os.Getenv("PWD"))
			panic("unexpected initial working directory")
		}
		if err := os.Setenv("PWD", "/another"); err != nil {
			panic(err)
		}
		if changed, err := os.Getwd(); err != nil || changed != cwd {
			panic("changing PWD should not change the working directory")
		}
		runtime.GC()
		if file := os.Getenv("EXPECT_FILE"); file != "" {
			data, err := os.ReadFile(file)
			if err != nil {
				panic(err)
			}
			if len(data) == 0 {
				panic("relative file access returned an empty file")
			}
		}
		if file := os.Getenv("EXPECT_FILE_FAILURE"); file != "" {
			if _, err := os.ReadFile(file); err == nil {
				panic("unmapped working directory should not grant file access")
			}
		}
	}

	// Check for environment variables (set by the test runner).
	println("ENV1:", os.Getenv("ENV1"))
	v, ok := os.LookupEnv("ENV2")
	if !ok {
		println("ENV2 not found")
	}
	println("ENV2:", v)

	// A value that holds '=' must still be found.
	os.Setenv("ENV3", "a=b")
	v, ok = os.LookupEnv("ENV3")
	if !ok {
		println("ENV3 not found")
	}
	println("ENV3:", v)

	found := false
	expected := "ENV1=" + os.Getenv("ENV1")
	for _, envVar := range os.Environ() {
		if envVar == expected {
			found = true
		}
	}
	if !found {
		println("could not find " + expected + " in os.Environ()")
	}

	// Check for command line arguments.
	// Argument 0 is skipped because it is the program name, which varies by
	// test run.
	println()
	for _, arg := range os.Args[1:] {
		println("arg:", arg)
	}
}
