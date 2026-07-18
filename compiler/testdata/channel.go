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

func selectNonBlockingSend(ch chan int, value int) bool {
	select {
	case ch <- value:
		return true
	default:
		return false
	}
}

func selectNonBlockingRecv(ch chan int) (int, bool, bool) {
	select {
	case value, ok := <-ch:
		return value, ok, true
	default:
		return 0, false, false
	}
}

func selectNonBlockingZeroSend(ch chan struct{}) bool {
	select {
	case ch <- struct{}{}:
		return true
	default:
		return false
	}
}

func selectNonBlockingZeroRecv(ch chan struct{}) (bool, bool) {
	select {
	case _, ok := <-ch:
		return ok, true
	default:
		return false, false
	}
}

func selectBlocking(ch1, ch2 chan int) (int, bool) {
	select {
	case value, ok := <-ch1:
		return value, ok
	case value, ok := <-ch2:
		return value, ok
	}
}
