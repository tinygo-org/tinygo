package main

import (
	"runtime"
	"sync"
)

var wg sync.WaitGroup

func main() {
	println("# simple recover")
	recoverSimple()

	println("\n# recover with result")
	result := recoverWithResult()
	println("result:", result)

	println("\n# nested defer frame")
	nestedDefer()

	println("\n# nested panic: panic inside recover")
	nestedPanic()

	println("\n# panic inside defer")
	panicInsideDefer()

	println("\n# panic replace")
	panicReplace()

	println("\n# defer panic")
	deferPanic()

	println("\n# runtime.Goexit")
	runtimeGoexit()
}

func recoverSimple() {
	defer func() {
		println("recovering...")
		printitf("recovered:", recover())
	}()
	println("running panic...")
	panic("panic")
}

func recoverWithResult() (result int) {
	defer func() {
		printitf("recovered:", recover())
	}()
	result = 3
	println("running panic...")
	panic("panic")
}

func nestedDefer() {
	defer func() {
		printitf("recovered:", recover())
	}()

	func() {
		// The defer here doesn't catch the panic using recover(), so the outer
		// panic should do that.
		defer func() {
			println("deferred nested function")
		}()
		panic("panic")
	}()
	println("unreachable")
}

func nestedPanic() {
	defer func() {
		printitf("recovered 1:", recover())

		defer func() {
			printitf("recovered 2:", recover())
		}()

		panic("foo")
	}()
	panic("panic")
}

func panicInsideDefer() {
	defer func() {
		printitf("recovered:", recover())
	}()
	defer func() {
		panic("panic")
	}()
}

func panicReplace() {
	defer func() {
		printitf("recovered:", recover())
	}()
	defer func() {
		println("panic 2")
		panic("panic 2")
	}()
	println("panic 1")
	panic("panic 1")
}

func deferPanic() {
	defer func() {
		printitf("recovered from deferred call:", recover())
	}()

	// This recover should not do anything.
	defer recover()

	defer panic("deferred panic")
	println("defer panic")
}

func runtimeGoexit() {
	wg.Add(1)
	go func() {
		defer func() {
			println("Goexit deferred function, recover is nil:", recover() == nil)
			wg.Done()
		}()

		runtime.Goexit()
	}()
	wg.Wait()
}

func printitf(msg string, itf interface{}) {
	switch itf := itf.(type) {
	case string:
		println(msg, itf)
	default:
		println(msg, itf)
	}
}
