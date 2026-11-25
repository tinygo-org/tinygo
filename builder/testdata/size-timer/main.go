package main

import "time"

func main() {
	ticker := time.NewTicker(2 * time.Millisecond)
	<-ticker.C
	ticker.Stop()
	<-time.After(10 * time.Millisecond)
}
