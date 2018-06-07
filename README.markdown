# TinyGo - Go compiler for microcontrollers

> We never expected Go to be an embedded language and so it's got serious
> problems [...].

-- Rob Pike, [GopherCon 2014 Opening Keynote](https://www.youtube.com/watch?v=VoS7DsT1rdM&feature=youtu.be&t=2799)

TinyGo is a project to bring Go to microcontrollers and small systems with a
single processor core. It is similar to [emgo](https://github.com/ziutek/emgo)
but a major difference is that I want to keep the Go memory model (which implies
garbage collection of some sort). Another difference is that TinyGo uses LLVM
internally instead of emitting C, which hopefully leads to smaller and more
efficient code and certainly leads to more flexibility.

My original reasoning was: if [Python](https://micropython.org/) can run on
microcontrollers, then certainly [Go](https://golang.org/) should be able to and
run on even lower level micros.

Example program (blinky):

```go
import "machine"

func main() {
	led := machine.GPIO{machine.LED}
	led.Configure(machine.GPIOConfig{Mode: machine.GPIO_OUTPUT})
	for {
		led.Low()
		runtime.Sleep(runtime.Millisecond * 1000)

		led.High()
		runtime.Sleep(runtime.Millisecond * 1000)
	}
}
```

Currently supported features:

  * control flow
  * many (but not all) basic types: most ints, strings, structs
  * function calling
  * interfaces for basic types (with type switches and asserts)
  * goroutines (very initial support)
  * function pointers (non-blocking)

Not yet supported:

  * float, complex, etc.
  * garbage collection
  * interface methods
  * channels
  * introspection (if it ever gets implemented)
  * standard library (needs more language support)
  * `defer`
  * closures
  * ...

## Analysis

The goal is to reduce code size (and increase performance) by performing all
kinds of whole-program analysis passes. The official Go compiler doesn't do a
whole lot of analysis (except for escape analysis) becauses it needs to be fast,
but embedded programs are necessarily smaller so it becomes practical. And I
think especially program size can be reduced by a large margin when actually
trying to optimize for it.
