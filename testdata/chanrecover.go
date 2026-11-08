package main

import (
	"runtime"
	"sync"
)

func errString(r any) string {
	switch v := r.(type) {
	case nil:
		return "<nil>"
	case runtime.Error:
		return "runtime error: " + v.Error()
	case error:
		return "error: " + v.Error()
	case string:
		return "string: " + v
	case int:
		return "int"
	}
	return "other"
}

func try(name string, f func()) {
	defer func() {
		println(name+":", errString(recover()))
	}()
	f()
}

// Every recovered panic below must leave the channel machinery usable.
func checkSelectWorks(name string) {
	a := make(chan int, 1)
	b := make(chan int)
	select {
	case a <- 1:
	case <-b:
	}
	select {
	case v := <-a:
		_ = v
	case <-b:
	}
	println(name + ": select ok")
}

func closedBuffered() chan int {
	ch := make(chan int, 1)
	close(ch)
	return ch
}

func sendPlain() {
	try("plain send", func() { closedBuffered() <- 1 })
	checkSelectWorks("after plain send")
}

func reuseAfterRecover() {
	ch := closedBuffered()
	try("reuse plain send", func() { ch <- 1 })
	v, ok := <-ch
	println("reuse receive:", v, ok)
	try("reuse select default", func() {
		select {
		case ch <- 1:
		default:
		}
	})
	v, ok = <-ch
	println("reuse receive:", v, ok)
	try("reuse close", func() { close(ch) })
	runtime.GC()
	checkSelectWorks("after reuse")
}

func sendSelectDefault() {
	try("select default", func() {
		ch := closedBuffered()
		select {
		case ch <- 1:
		default:
		}
	})
	checkSelectWorks("after select default")
}

func sendSelectBlocking() {
	func() {
		defer func() { println("select blocking:", errString(recover())) }()
		ch := closedBuffered()
		block := make(chan int)
		select {
		case ch <- 1:
		case <-block:
		}
	}()
	checkSelectWorks("after select blocking")
}

func sendSelectLaterCase() {
	func() {
		defer func() { println("select later case:", errString(recover())) }()
		closed := closedBuffered()
		block := make(chan int)
		var nilch chan int
		select {
		case <-block:
		case nilch <- 1:
		case closed <- 1:
		}
	}()
	checkSelectWorks("after select later case")
}

func sendSelectLoop() {
	n := 0
	for i := 0; i < 200; i++ {
		func() {
			defer func() {
				if recover() != nil {
					n++
				}
			}()
			ch := closedBuffered()
			block := make(chan int)
			select {
			case ch <- i:
			case <-block:
			}
		}()
	}
	println("select loop recovered:", n)
	checkSelectWorks("after select loop")
}

func closeTwice() {
	try("close twice", func() {
		ch := make(chan int)
		close(ch)
		close(ch)
	})
	try("close nil", func() {
		var ch chan int
		close(ch)
	})
	checkSelectWorks("after close")
}

func nestedDefers() {
	func() {
		defer func() { println("outer:", errString(recover())) }()
		func() {
			defer func() { println("inner defer ran") }()
			ch := closedBuffered()
			block := make(chan int)
			select {
			case ch <- 1:
			case <-block:
			}
		}()
		println("not reached")
	}()
	checkSelectWorks("after nested")
}

func repanicInDefer() {
	func() {
		defer func() { println("final:", errString(recover())) }()
		defer func() {
			r := recover()
			println("first:", errString(r))
			panic("second")
		}()
		ch := closedBuffered()
		block := make(chan int)
		select {
		case ch <- 1:
		case <-block:
		}
	}()
	checkSelectWorks("after repanic")
}

func namedResult() (res int) {
	defer func() {
		if recover() != nil {
			res = 42
		}
	}()
	ch := closedBuffered()
	block := make(chan int)
	select {
	case ch <- 1:
	case <-block:
	}
	return 1
}

func selectInDeferredFunc() {
	func() {
		defer func() {
			recover()
			// Run a select while handling the panic.
			a := make(chan int, 1)
			block := make(chan int)
			select {
			case a <- 1:
			case <-block:
			}
			println("select in defer ok")
		}()
		var m map[int]int
		m[1] = 1
	}()
	checkSelectWorks("after select in defer")
}

// These two close a channel while a sender may be blocked on it, which the Go race detector reports.
func blockedSenderClosed() {
	for i := 0; i < 20; i++ {
		ch := make(chan int)
		res := make(chan string)
		go func() {
			defer func() { res <- errString(recover()) }()
			ch <- 1
		}()
		runtime.Gosched()
		close(ch)
		if s := <-res; i == 0 || s != "runtime error: send on closed channel" {
			println("blocked sender:", s)
		}
	}
	println("blocked sender done")
	checkSelectWorks("after blocked sender")
}

func blockedSelectSenderClosed() {
	for i := 0; i < 20; i++ {
		ch := make(chan int)
		block := make(chan int)
		res := make(chan string)
		go func() {
			defer func() { res <- errString(recover()) }()
			select {
			case ch <- 1:
			case <-block:
			}
		}()
		runtime.Gosched()
		close(ch)
		if s := <-res; i == 0 || s != "runtime error: send on closed channel" {
			println("blocked select sender:", s)
		}
	}
	println("blocked select sender done")
	checkSelectWorks("after blocked select sender")
}

func goroutineStress() {
	const workers = 8
	const iters = 200
	var wg sync.WaitGroup
	results := make(chan int, workers)
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			recovered := 0
			for i := 0; i < iters; i++ {
				func() {
					defer func() {
						if recover() != nil {
							recovered++
						}
					}()
					ch := closedBuffered()
					block := make(chan int)
					other := make(chan int, 1)
					select {
					case other <- i:
					case ch <- i:
					case <-block:
					}
					<-other
				}()
				runtime.Gosched()
			}
			results <- recovered
		}()
	}
	wg.Wait()
	close(results)
	total := 0
	for r := range results {
		total += r
	}
	// Each iteration either sends on other or panics on ch, depending on
	// which ready case is picked. Both outcomes are valid.
	println("goroutine stress ok:", total >= 0 && total <= workers*iters)
	checkSelectWorks("after goroutine stress")
}

func main() {
	sendPlain()
	reuseAfterRecover()
	sendSelectDefault()
	sendSelectBlocking()
	sendSelectLaterCase()
	sendSelectLoop()
	closeTwice()
	nestedDefers()
	repanicInDefer()
	println("named result:", namedResult())
	selectInDeferredFunc()
	blockedSenderClosed()
	blockedSelectSenderClosed()
	goroutineStress()
	println("done")
}
