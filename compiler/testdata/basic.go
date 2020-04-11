package main

func add(x, y int) int {
	return x + y
}

func stringEqual(s string) bool {
	return s == "s"
}

func closeChan(ch chan int) {
	close(ch)
}
