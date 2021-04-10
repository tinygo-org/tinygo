package main

func foo(callback func(int)) {
	callback(3)
}

func bar() {
	foo(someFunc)
}

func someFunc(int) {
}
