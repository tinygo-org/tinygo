package main

/*
#define foo 3
#define bar foo
*/
import "C"

const (
	Foo = C.foo
	Bar = C.bar
)
