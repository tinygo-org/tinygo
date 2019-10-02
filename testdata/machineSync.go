package main

import (
	"machine/sync"
	"time"
)

var cond sync.Cond

func delayedResume() {
	time.Sleep(time.Second)
	println("notifying")
	cond.Notify()
}

func main() {
	go delayedResume()
	cond.Wait()
	cond.Wait()
	cond.Clear()
	println("done")
}
