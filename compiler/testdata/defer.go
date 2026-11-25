package main

func external()

func deferSimple() {
	defer func() {
		print(3)
	}()
	external()
}

func noPanicCall() {
	print(1)
}

func deferNoPanicCall() {
	defer func() {
		print(2)
	}()
	noPanicCall()
}

func noPanicArithmetic(a, b int) bool {
	return a+b > 3
}

func deferNoPanicArithmetic(a, b int) bool {
	defer func() {
		print(4)
	}()
	return noPanicArithmetic(a, b)
}

func panicDivide(a, b int) int {
	return a / b
}

func deferPanicDivide(a, b int) int {
	defer func() {
		print(6)
	}()
	return panicDivide(a, b)
}

func deferMultiple() {
	defer func() {
		print(3)
	}()
	defer func() {
		print(5)
	}()
	external()
}

func deferInfiniteLoop() {
	for {
		defer print(8)
	}
}

func deferLoop() {
	for i := 0; i < 10; i++ {
		defer print(i)
	}
}

func deferBetweenLoops() {
	for i := 0; i < 10; i++ {
	}
	defer print(1)
	for i := 0; i < 10; i++ {
	}
}
