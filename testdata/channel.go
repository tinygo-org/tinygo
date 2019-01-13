package main

import "time"

func main() {
	ch := make(chan int)
	println("len, cap of channel:", len(ch), cap(ch))
	go sender(ch)

	n, ok := <-ch
	println("recv from open channel:", n, ok)

	for n := range ch {
		if n == 6 {
			time.Sleep(time.Microsecond)
		}
		println("received num:", n)
	}

	n, ok = <-ch
	println("recv from closed channel:", n, ok)

	// Test multi-sender.
	ch = make(chan int)
	go fastsender(ch)
	go fastsender(ch)
	go fastsender(ch)
	slowreceiver(ch)

	// Test multi-receiver.
	ch = make(chan int)
	go fastreceiver(ch)
	go fastreceiver(ch)
	go fastreceiver(ch)
	slowsender(ch)

	// Test iterator style channel.
	ch = make(chan int)
	go iterator(ch, 100)
	sum := 0
	for i := range ch {
		sum += i
	}
	println("sum(100):", sum)

	// Allow goroutines to exit.
	time.Sleep(time.Microsecond)
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
}

func fastsender(ch chan int) {
	ch <- 10
	ch <- 11
}

func slowreceiver(ch chan int) {
	for i := 0; i < 6; i++ {
		n := <-ch
		println("got n:", n)
		time.Sleep(time.Microsecond)
	}
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
}

func iterator(ch chan int, top int) {
	for i := 0; i < top; i++ {
		ch <- i
	}
	close(ch)
}
