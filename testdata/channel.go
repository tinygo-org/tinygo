package main

import (
	"runtime"
	"sync"
	"time"
)

var wg sync.WaitGroup

type intchan chan int

func main() {
	ch := make(chan int, 2)
	ch <- 1
	println("len, cap of channel:", len(ch), cap(ch), ch == nil)

	ch = make(chan int)
	println("len, cap of channel:", len(ch), cap(ch), ch == nil)

	wg.Add(1)
	go sender(ch)

	n, ok := <-ch
	println("recv from open channel:", n, ok)

	for n := range ch {
		println("received num:", n)
	}

	wg.Wait()
	n, ok = <-ch
	println("recv from closed channel:", n, ok)

	// Test various channel size types.
	_ = make(chan int, int8(2))
	_ = make(chan int, int16(2))
	_ = make(chan int, int32(2))
	_ = make(chan int, int64(2))
	_ = make(chan int, uint8(2))
	_ = make(chan int, uint16(2))
	_ = make(chan int, uint32(2))
	_ = make(chan int, uint64(2))

	// Test that named channels don't crash the compiler.
	named := make(intchan, 1)
	named <- 3
	<-named
	select {
	case <-named:
	default:
	}

	// Test bigger values
	ch2 := make(chan complex128)
	wg.Add(1)
	go sendComplex(ch2)
	println("complex128:", <-ch2)
	wg.Wait()

	// Test multi-sender.
	ch = make(chan int)
	wg.Add(3)
	go fastsender(ch, 10)
	go fastsender(ch, 23)
	go fastsender(ch, 40)
	slowreceiver(ch)
	wg.Wait()

	// Test multi-receiver.
	ch = make(chan int)
	wg.Add(3)
	go fastreceiver(ch)
	go fastreceiver(ch)
	go fastreceiver(ch)
	slowsender(ch)
	wg.Wait()

	// Test iterator style channel.
	ch = make(chan int)
	wg.Add(1)
	go iterator(ch, 100)
	sum := 0
	for i := range ch {
		sum += i
	}
	wg.Wait()
	println("sum(100):", sum)

	// Test simple selects.
	go selectDeadlock() // cannot use waitGroup here - never terminates
	wg.Add(1)
	go selectNoOp()
	wg.Wait()

	// Test select with a single send operation (transformed into chan send).
	ch = make(chan int)
	wg.Add(1)
	go fastreceiver(ch)
	select {
	case ch <- 5:
	}
	close(ch)
	wg.Wait()
	println("did send one")

	// Test select with a single recv operation (transformed into chan recv).
	select {
	case n := <-ch:
		println("select one n:", n)
	}

	// Test select recv with channel that has one entry.
	ch = make(chan int)
	wg.Add(1)
	go func(ch chan int) {
		runtime.Gosched()
		ch <- 55
		wg.Done()
	}(ch)
	select {
	case make(chan int) <- 3:
		println("unreachable")
	case <-(chan int)(nil):
		println("unreachable")
	case n := <-ch:
		println("select n from chan:", n)
	case n := <-make(chan int):
		println("unreachable:", n)
	}
	wg.Wait()

	// Test select recv with closed channel.
	close(ch)
	select {
	case make(chan int) <- 3:
		println("unreachable")
	case n := <-ch:
		println("select n from closed chan:", n)
	case n := <-make(chan int):
		println("unreachable:", n)
	}

	// Test select send.
	ch = make(chan int)
	wg.Add(1)
	go fastreceiver(ch)
	select {
	case ch <- 235:
		println("select send")
	case n := <-make(chan int):
		println("unreachable:", n)
	}
	close(ch)
	wg.Wait()

	// test non-concurrent buffered channels
	ch = make(chan int, 2)
	ch <- 1
	ch <- 2
	println("non-concurrent channel receive:", <-ch)
	println("non-concurrent channel receive:", <-ch)

	// test closing channels with buffered data
	ch <- 3
	ch <- 4
	close(ch)
	println("closed buffered channel receive:", <-ch)
	println("closed buffered channel receive:", <-ch)
	println("closed buffered channel receive:", <-ch)

	// test using buffered channels as regular channels with special properties
	wg.Add(6)
	ch = make(chan int, 2)
	go send(ch)
	go send(ch)
	go send(ch)
	go send(ch)
	go receive(ch)
	go receive(ch)
	wg.Wait()
	close(ch)
	var count int
	for range ch {
		count++
	}
	println("hybrid buffered channel receive:", count)

	// test blocking selects
	ch = make(chan int)
	sch1 := make(chan int)
	sch2 := make(chan int)
	sch3 := make(chan int)
	wg.Add(3)
	go func() {
		defer wg.Done()
		time.Sleep(time.Millisecond)
		sch1 <- 1
	}()
	go func() {
		defer wg.Done()
		time.Sleep(time.Millisecond)
		sch2 <- 2
	}()
	go func() {
		defer wg.Done()
		// merge sch2 and sch3 into ch
		for i := 0; i < 2; i++ {
			var v int
			select {
			case v = <-sch1:
			case v = <-sch2:
			}
			select {
			case sch3 <- v:
				panic("sent to unused channel")
			case ch <- v:
			}
		}
	}()
	sum = 0
	for i := 0; i < 2; i++ {
		select {
		case sch3 <- sum:
			panic("sent to unused channel")
		case v := <-ch:
			sum += v
		}
	}
	wg.Wait()
	println("blocking select sum:", sum)
}

func send(ch chan<- int) {
	ch <- 1
	wg.Done()
}

func receive(ch <-chan int) {
	<-ch
	wg.Done()
}

func sender(ch chan int) {
	for i := 1; i <= 8; i++ {
		if i == 4 {
			time.Sleep(time.Microsecond)
			println("slept")
		}
		ch <- i
	}
	close(ch)
	wg.Done()
}

func sendComplex(ch chan complex128) {
	ch <- 7 + 10.5i
	wg.Done()
}

func fastsender(ch chan int, n int) {
	ch <- n
	ch <- n + 1
	wg.Done()
}

func slowreceiver(ch chan int) {
	sum := 0
	for i := 0; i < 6; i++ {
		sum += <-ch
		time.Sleep(time.Microsecond)
	}
	println("sum of n:", sum)
}

func slowsender(ch chan int) {
	for n := 0; n < 6; n++ {
		time.Sleep(time.Microsecond)
		ch <- 12 + n
	}
}

func fastreceiver(ch chan int) {
	sum := 0
	for i := 0; i < 2; i++ {
		n := <-ch
		sum += n
	}
	println("sum:", sum)
	wg.Done()
}

func iterator(ch chan int, top int) {
	for i := 0; i < top; i++ {
		ch <- i
	}
	close(ch)
	wg.Done()
}

func selectDeadlock() {
	println("deadlocking")
	select {}
	println("unreachable")
}

func selectNoOp() {
	println("select no-op")
	select {
	default:
	}
	println("after no-op")
	wg.Done()
}
