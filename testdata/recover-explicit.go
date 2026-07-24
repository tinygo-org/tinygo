package main

type pair struct {
	first  int
	second int
}

var panicMap = map[any]int{}

func main() {
	println("# direct panic")
	direct()

	println("\n# results")
	println("scalar result:", scalarResult())
	result := aggregateResult()
	println("aggregate result:", result.first, result.second)

	println("\n# nested panics")
	nestedDefer()
	nestedPanic()
	panicReplace()
	deferPanic()
	repanic()

	println("\n# runtime panics")
	mustRecover("index", func() {
		values := []int{1}
		println(values[2])
	})
	mustRecover("index from helper", func() {
		println(readOutOfBounds([]byte{1}))
	})
	mustRecover("slice", func() {
		values := []int{1}
		_ = values[:2]
	})
	mustRecover("type assertion", func() {
		var value any = "string"
		println(value.(int))
	})
	mustRecover("interface comparison", func() {
		var value any = []int{}
		println(value == value)
	})
	mustRecover("map assignment", func() {
		panicMap[[]int{}] = 1
	})
	mustRecover("map lookup", func() {
		_ = panicMap[[]int{}]
	})
	mustRecover("map delete", func() {
		delete(panicMap, []int{})
	})
	mustRecover("nil map", func() {
		var values map[string]int
		values["key"] = 1
	})
	mustRecover("divide by zero", func() {
		var divisor int
		println(1 / divisor)
	})
	mustRecover("nil pointer", func() {
		var pointer *int
		println(*pointer)
	})
}

func direct() {
	defer func() {
		println("recovered direct:", recover() == "direct panic")
	}()
	panicHelper("direct panic")
	println("unreachable after direct panic")
}

//go:noinline
func panicHelper(value any) {
	panic(value)
}

func scalarResult() (result int) {
	defer func() {
		recover()
	}()
	result = 3
	panicHelper("scalar result panic")
	return
}

func aggregateResult() (result pair) {
	defer func() {
		recover()
	}()
	result = pair{1, 2}
	panicHelper("aggregate result panic")
	return
}

func nestedDefer() {
	defer func() {
		println("recovered nested:", recover() == "nested panic")
	}()
	func() {
		defer println("nested defer ran")
		panicHelper("nested panic")
	}()
}

func nestedPanic() {
	defer func() {
		println("recovered outer:", recover() == "outer panic")
	}()
	defer func() {
		println("recovered inner:", recover() == "inner panic")
		panicHelper("outer panic")
	}()
	panicHelper("inner panic")
}

func panicReplace() {
	defer func() {
		println("recovered replacement:", recover() == "replacement panic")
	}()
	defer func() {
		panicHelper("replacement panic")
	}()
	panicHelper("original panic")
}

func deferPanic() {
	defer func() {
		println("recovered deferred:", recover() == "deferred panic")
	}()
	defer panicHelper("deferred panic")
}

func repanic() {
	defer func() {
		println("recovered repanic:", recover() == "repanic")
	}()
	defer func() {
		value := recover()
		panicHelper(value)
	}()
	panicHelper("repanic")
}

func mustRecover(name string, fn func()) {
	defer func() {
		if recover() == nil {
			println("failed to recover:", name)
		} else {
			println("recovered:", name)
		}
	}()
	fn()
	println("unreachable after:", name)
}

//go:noinline
func readOutOfBounds(values []byte) byte {
	return values[2]
}
