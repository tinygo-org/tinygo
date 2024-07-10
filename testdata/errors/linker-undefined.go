package main

func foo()

func main() {
	foo()
	foo()
}

// ERROR: linker-undefined.go:6: linker could not find symbol {{_?}}main.foo
// ERROR: linker-undefined.go:7: linker could not find symbol {{_?}}main.foo
