package main

func regularFunctionGoroutine() {
	go regularFunction(5)
}

func inlineFunctionGoroutine() {
	go func(x int) {
	}(5)
}

func closureFunctionGoroutine() {
	n := 3
	go func(x int) {
		n = 7
	}(5)
	print(n) // note: this is racy (but good enough for this test)
}

func funcGoroutine(fn func(x int)) {
	go fn(5)
}

func recoverBuiltinGoroutine() {
	// This is a no-op.
	go recover()
}

func copyBuiltinGoroutine(dst, src []byte) {
	// This is not run in a goroutine. While this copy operation can indeed take
	// some time (if there is a lot of data to copy), there is no race-free way
	// to make use of the result so it's unlikely applications will make use of
	// it. And doing it this way should be just within the Go specification.
	go copy(dst, src)
}

func closeBuiltinGoroutine(ch chan int) {
	// This builtin is executed directly, not in a goroutine.
	// The observed behavior is the same.
	go close(ch)
}

func regularFunction(x int)

type simpleInterface interface {
	Print(string)
}

func startInterfaceMethod(itf simpleInterface) {
	go itf.Print("test")
}
