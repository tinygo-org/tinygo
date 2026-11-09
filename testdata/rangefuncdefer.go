package main

import "runtime"

var sequence int

func record(n int) {
	sequence = sequence*10 + n
}

func numbers(yield func(int) bool) {
	for n := 1; n <= 3; n++ {
		if !yield(n) {
			return
		}
	}
}

func collecting(yield func(int) bool) {
	var reuse [][]int
	for n := 1; n <= 3; n++ {
		runtime.GC()
		reuse = reuse[:0]
		for size := 1; size <= 64; size++ {
			s := make([]int, size%8+1)
			for i := range s {
				s[i] = -1
			}
			reuse = append(reuse, s)
		}
		if !yield(n) {
			return
		}
	}
}

type recorder interface {
	record(n int)
}

type adder struct {
	offset int
}

func (a adder) record(n int) {
	record(n + a.offset)
}

type big struct {
	a, b, c, d, e int
}

func (b big) sum() int {
	return b.a + b.b + b.c + b.d + b.e
}

func main() {
	run("direct", direct)
	run("closure", closure)
	run("closure without free variables", closureNoFreeVars)
	run("function value", funcValue)
	run("interface", interfaceCall)
	run("method", method)
	run("builtin", builtin)
	run("aggregate", aggregate)
	run("nested", nested)
	run("break", earlyBreak)
	run("return", earlyReturn)
	run("named result", func() { record(namedResult()) })
	run("bare return", func() { record(bareReturn()) })
	run("return in loop body", func() { record(returnInBody()) })
	run("return local", func() { record(returnLocal()) })
	run("own defers", ownDefers)
	run("generic", generic[int])
	run("function literal", func() {
		for n := range numbers {
			defer record(n)
		}
		record(4)
	})
	run("gc", gc)
	run("panic", recoverPanic)
	run("recover in loop body", recoverInBody)
	run("panic in deferred call", panicInDefer)
}

func run(name string, f func()) {
	sequence = 0
	f()
	println(name+":", sequence)
}

func direct() {
	for n := range numbers {
		defer record(n)
	}
	record(4)
}

func closure() {
	k := 2
	for n := range numbers {
		defer func() {
			record(n * k)
		}()
	}
	record(1)
}

func closureNoFreeVars() {
	for range numbers {
		defer func() {
			record(7)
		}()
	}
}

func funcValue() {
	f := record
	for n := range numbers {
		defer f(n)
	}
}

func interfaceCall() {
	var r recorder = adder{offset: 3}
	for n := range numbers {
		defer r.record(n)
	}
}

func method() {
	a := adder{offset: 5}
	for n := range numbers {
		defer a.record(n)
	}
}

func builtin() {
	m := map[int]bool{1: true, 2: true, 3: true}
	defer func() {
		record(len(m))
	}()
	for n := range numbers {
		defer delete(m, n)
	}
	record(len(m))
}

func aggregate() {
	for n := range numbers {
		b := big{n, n, n, n, n}
		defer func(b big) {
			record(b.sum() % 10)
		}(b)
	}
}

func nested() {
	for a := range numbers {
		for b := range numbers {
			if b > a {
				break
			}
			defer record(b)
		}
		defer record(a)
	}
}

func earlyBreak() {
	for n := range numbers {
		defer record(n)
		if n == 2 {
			break
		}
	}
	record(9)
}

func earlyReturn() {
	for n := range numbers {
		defer record(n)
		if n == 2 {
			record(8)
			return
		}
	}
	record(9)
}

func namedResult() (result int) {
	for n := range numbers {
		defer func() {
			result = result*10 + n
		}()
	}
	return 4
}

func bareReturn() (result int) {
	for n := range numbers {
		defer func() {
			result = result*10 + n
		}()
	}
	result = 4
	return
}

func returnInBody() (result int) {
	for n := range numbers {
		defer func() {
			result = result*10 + n
		}()
		if n == 2 {
			return 5
		}
	}
	return 4
}

func returnLocal() int {
	local := 4
	for n := range numbers {
		defer func() {
			local = local*10 + n
		}()
	}
	return local
}

func ownDefers() {
	defer record(1)
	for n := range numbers {
		defer record(n + 1)
	}
	defer record(5)
}

func generic[T ~int]() {
	for n := range numbers {
		defer record(int(T(n)))
	}
}

func gc() {
	for n := range collecting {
		s := make([]int, 16)
		s[15] = n
		defer func() {
			record(s[15])
		}()
	}
	runtime.GC()
}

func recoverPanic() {
	defer func() {
		record(len(recover().(string)))
	}()
	for n := range numbers {
		defer record(n)
		if n == 2 {
			panic("panic")
		}
	}
}

func recoverInBody() {
	for n := range numbers {
		defer func() {
			if recover() != nil {
				record(9)
			}
		}()
		defer record(n)
		if n == 3 {
			panic("panic")
		}
	}
}

func panicInDefer() {
	defer func() {
		record(len(recover().(string)))
	}()
	for n := range numbers {
		if n == 2 {
			defer func() {
				panic("deferred")
			}()
		}
		defer record(n)
	}
}
