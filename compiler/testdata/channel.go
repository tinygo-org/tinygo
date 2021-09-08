package main

func chanIntSend(ch chan int) {
	ch <- 3
}

func chanIntRecv(ch chan int) {
	<-ch
}

func chanZeroSend(ch chan struct{}) {
	ch <- struct{}{}
}

func chanZeroRecv(ch chan struct{}) {
	<-ch
}

func selectZeroRecv(ch1 chan int, ch2 chan struct{}) {
	select {
	case ch1 <- 1:
	case <-ch2:
	default:
	}
}
