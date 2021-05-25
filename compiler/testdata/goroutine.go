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

func regularFunction(x int)
