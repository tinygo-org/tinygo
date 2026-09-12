package pass_test

import "testing"

func TestPass(t *testing.T) {
	// This test passes.
}

func TestDeferredSuspend(t *testing.T) {
	ready := make(chan struct{})
	release := make(chan struct{})
	finished := make(chan struct{})
	go func() {
		defer func() {
			close(ready)
			<-release
			close(finished)
		}()
	}()
	<-ready
	close(release)
	<-finished
}

func TestBlockingSendWithDefer(t *testing.T) {
	ready := make(chan struct{})
	values := make(chan int)
	finished := make(chan struct{})
	go func() {
		defer close(finished)
		close(ready)
		values <- 1
	}()
	<-ready
	if value := <-values; value != 1 {
		t.Fatalf("unexpected value: %d", value)
	}
	<-finished
}

func TestConcurrentAggregateSuspend(t *testing.T) {
	ready := make(chan int)
	release := make(chan int)
	results := make(chan [2]int)
	for id := 1; id <= 2; id++ {
		go func() {
			defer func() {}()
			first, second := suspendedPair(id, ready, release)
			results <- [2]int{first, second}
		}()
	}
	<-ready
	<-ready
	release <- 20
	release <- 10
	for range 2 {
		result := <-results
		if result[0] != 1 && result[0] != 2 {
			t.Fatalf("unexpected first result: %d", result[0])
		}
		if result[1] != 10 && result[1] != 20 {
			t.Fatalf("unexpected second result: %d", result[1])
		}
	}
}

func suspendedPair(id int, ready chan<- int, release <-chan int) (int, int) {
	ready <- id
	return id, <-release
}
