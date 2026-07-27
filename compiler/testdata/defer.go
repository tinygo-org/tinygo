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
