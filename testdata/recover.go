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

	println("\n# indirect recover")
	indirectRecover()

	println("\n# runtime.Goexit")
	runtimeGoexit()

	println("\n# repanic")
	recoverRepanic()

	println("\n# recover runtime errors")
	recoverRuntimeError()

	println("\n# recover from nil map and closed channel")
	recoverNilMapAndChan()

	println("\n# recover from hardware signals")
	recoverSignals()
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

// TODO: Go only allows recover() to succeed when called directly from a
// deferred function. Update this test once runtime recover can distinguish it.
func indirectRecover() {
	defer func() {
		if r := indirectRecoverHelper(); r == nil {
			println("indirect recover returned nil")
		} else {
			printitf("indirect recover returned:", r)
		}
	}()
	panic("indirect panic")
}

//go:noinline
func indirectRecoverHelper() interface{} {
	return recover()
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

// Test that a repanic inside a deferred function propagates correctly
// instead of re-running the same defer. This is a regression test for
// tinygo-org/tinygo issue 3449.
func recoverRepanic() {
	// Two defers: inner recovers and repanics, outer should catch it.
	defer func() {
		r := recover()
		if r != nil {
			printitf("outer recovered:", r)
		}
	}()
	defer func() {
		r := recover()
		if r != nil {
			println("inner, repanicking")
			panic(r)
		}
	}()
	panic("repanic value")
}

func printitf(msg string, itf interface{}) {
	switch itf := itf.(type) {
	case string:
		println(msg, itf)
	default:
		println(msg, itf)
	}
}

// Test recovering from runtime errors (bounds checks, type assertions, etc.)
func recoverRuntimeError() {
	recoverMustPanic("index", func() {
		s := make([]int, 5)
		_ = s[99]
	})
	recoverMustPanic("index from helper", func() {
		s := []byte{1}
		_ = readOutOfBounds(s)
	})
	recoverMustPanic("slice", func() {
		s := make([]int, 5)
		_ = s[3:99]
	})
	recoverMustPanic("type assert", func() {
		var x interface{} = 1
		_ = x.(string)
	})
	recoverEmptyInterfaceTypeAssert()
}

//go:noinline
func readOutOfBounds(s []byte) byte {
	return s[2]
}

func recoverEmptyInterfaceTypeAssert() {
	defer func() {
		r := recover()
		if r != nil {
			println("  failed empty interface type assert")
		} else {
			println("  recovered: empty interface type assert")
		}
	}()
	var intf interface{} = 3
	typed := intf.(interface{})
	useEmptyInterface(typed)
}

func useEmptyInterface(typed interface{}) {
	if typed.(int) != 3 {
		println("  failed empty interface value")
	}
}

func recoverMustPanic(name string, f func()) {
	defer func() {
		r := recover()
		if r != nil {
			println("  recovered:", name)
		} else {
			println("  failed to recover:", name)
		}
	}()
	f()
}

// Test recovering from nil map assignment and closed channel send.
func recoverNilMapAndChan() {
	recoverMustPanic("nil map", func() {
		var m map[string]int
		m["x"] = 1
	})
	recoverMustPanic("closed chan", func() {
		ch := make(chan int)
		close(ch)
		ch <- 1
	})
	recoverMustPanic("close nil chan", func() {
		var ch chan int
		close(ch)
	})
}

// Test recovering from hardware signals (SIGFPE, SIGSEGV).
func recoverSignals() {
	recoverMustPanic("divide by zero", func() {
		var x int
		println(1 / x)
	})
	recoverMustPanic("nil pointer dereference", func() {
		var p *int
		println(*p)
	})
}
