package main

import (
	"os"
	"runtime"
	"time"
)

func main() {
	switch os.Args[1] {
	case "main":
		runtime.Goexit()
	case "deadlock":
		f := func() {
			for i := 0; i < 10; i++ {
			}
		}
		go f()
		go f()
		runtime.Goexit()
	case "exit":
		go func() {
			time.Sleep(time.Millisecond)
		}()
		i := 0
		runtime.SetFinalizer(&i, func(p *int) {})
		runtime.GC()
		runtime.Goexit()
	case "main-other":
		go func() {
			println("other goroutine ran")
		}()
		runtime.Goexit()
	case "in-panic":
		go func() {
			defer func() {
				runtime.Goexit()
			}()
			panic("panic before Goexit")
		}()
		runtime.Goexit()
	case "panic":
		defer func() {
			panic("panic after Goexit")
		}()
		runtime.Goexit()
	case "recovered-panic":
		defer func() {
			defer func() {
				recover()
			}()
			panic("recovered panic after Goexit")
		}()
		runtime.Goexit()
	case "recover-before-panic":
		defer func() {
			if recover() == nil {
				panic("bad recover")
			}
		}()
		defer func() {
			panic("panic during Goexit")
		}()
		runtime.Goexit()
	case "recover-before-panic-loop":
		for i := 0; i < 2; i++ {
			defer func() {
			}()
		}
		defer func() {
			if recover() == nil {
				panic("bad recover")
			}
		}()
		defer func() {
			panic("panic during Goexit")
		}()
		runtime.Goexit()
	}
}
