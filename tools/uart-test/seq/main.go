//go:build tinygo

package main

import (
	"time"

	"github.com/tinygo-org/tinygo/tools/uart-test/common"
)

func main() {
	common.WaitGo("seq")
	println("@@GO seq")
	for i := 0; i < 2000; i++ {
		println("S", i, i*7919%10007)
	}
	println("@@END seq")
	for {
		time.Sleep(time.Second)
	}
}
