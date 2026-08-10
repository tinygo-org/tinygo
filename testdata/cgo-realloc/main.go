package main

/*
int reallocPreservesContents(void);
*/
import "C"

func main() {
	println("realloc preserves contents:", C.reallocPreservesContents())
}
