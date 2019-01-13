# TinyGo - Go compiler for microcontrollers

[![Build Status](https://travis-ci.com/aykevl/tinygo.svg?branch=master)](https://travis-ci.com/aykevl/tinygo)

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
import (
	"machine"
	"time"
)

func main() {
	led := machine.GPIO{machine.LED}
	led.Configure(machine.GPIOConfig{Mode: machine.GPIO_OUTPUT})
	for {
		led.Low()
		time.Sleep(time.Millisecond * 1000)

		led.High()
		time.Sleep(time.Millisecond * 1000)
	}
}
```

Currently supported features:

  * control flow
  * many (but not all) basic types: most ints, floats, strings, structs
  * function calling
  * interfaces for basic types (with type switches and asserts)
  * goroutines (very initial support)
  * function pointers (non-blocking)
  * interface methods
  * standard library (but most packages won't work due to missing language
    features)
  * slices (partially)
  * maps (very rough, unfinished)
  * defer
  * closures
  * bound methods
  * complex numbers (except for arithmetic)
  * channels (with some limitations)

Not yet supported:

  * select
  * complex arithmetic
  * garbage collection
  * recover
  * introspection (if it ever gets implemented)
  * ...

## Installation

See the [getting started instructions](https://tinygo.org/getting-started/).

### Running with Docker

A docker container exists for easy access to the `tinygo` CLI:

```sh
$ docker run --rm -v $(pwd):/src tinygo/tinygo tinygo build -o /src/wasm.wasm -target wasm examples/wasm
```

Note that you cannot run `tinygo flash` from inside the docker container,
so it is less useful for microcontroller development.

## Supported targets

The following architectures/systems are currently supported:

  * ARM (Cortex-M)
  * AVR (Arduino Uno)
  * Linux
  * WebAssembly

For more information, see [this list of targets and
boards](https://tinygo.org/targets/). Pull requests for
broader support are welcome!

## Analysis and optimizations

The goal is to reduce code size (and increase performance) by performing all
kinds of whole-program analysis passes. The official Go compiler doesn't do a
whole lot of analysis (except for escape analysis) because it needs to be fast,
but embedded programs are necessarily smaller so it becomes practical. And I
think especially program size can be reduced by a large margin when actually
trying to optimize for it.

Implemented compiler passes:

  * Analyse which functions are blocking. Blocking functions are functions that
    call sleep, chan send, etc. Its parents are also blocking.
  * Analyse whether the scheduler is needed. It is only needed when there are
    `go` statements for blocking functions.
  * Analyse whether a given type switch or type assert is possible with
    [type-based alias analysis](https://en.wikipedia.org/wiki/Alias_analysis#Type-based_alias_analysis).
    I would like to use flow-based alias analysis in the future, if feasible.
  * Do basic dead code elimination of functions. This pass makes later passes
    better and probably improves compile time as well.

## Scope

Goals:

  * Have very small binary sizes. Don't pay for what you don't use.
  * Support for most common microcontroller boards.
  * Be usable on the web using WebAssembly.
  * Good CGo support, with no more overhead than a regular function call.
  * Support most standard library packages and compile most Go code without
    modification.

Non-goals:

  * Using more than one core.
  * Be efficient while using zillions of goroutines. However, good goroutine
    support is certainly a goal.
  * Be as fast as `gc`. However, LLVM will probably be better at optimizing
    certain things so TinyGo might actually turn out to be faster for number
    crunching.
  * Be able to compile every Go program out there.

## Documentation

Documentation is currently maintained on a dedicated web site located at [https://tinygo.org/](https://tinygo.org/).

You can find the web site code at [https://github.com/tinygo-org/tinygo-site](https://github.com/tinygo-org/tinygo-site).

## Getting help

If you're looking for a more interactive way to discuss TinyGo usage or
development, we have a [#TinyGo channel](https://gophers.slack.com/messages/CDJD3SUP6/)
on the [Gophers Slack](https://gophers.slack.com).

If you need an invitation for the Gophers Slack, you can generate one here which
should arrive fairly quickly (under 1 min): https://invite.slack.golangbridge.org

## Contributing

Patches are welcome!

If you want to contribute, here are some suggestions:

  * A long tail of small (and large) language features haven't been implemented
    yet. In almost all cases, the compiler will show a `todo:` error from
    `compiler/compiler.go` when you try to use it. You can try implementing it,
    or open a bug report with a small code sample that fails to compile.
  * Lots of targets/boards are still unsupported. Adding an architecture often
    requires a few compiler changes, but if the architecture is supported you
    can try implementing support for a new chip or board in `src/runtime`. For
    details, see [this wiki entry on adding
    archs/chips/boards](https://github.com/aykevl/tinygo/wiki/Adding-a-new-board).
  * Microcontrollers have lots of peripherals and many don't have an
    implementation yet in the `machine` package. Adding support for new
    peripherals is very useful.
  * Just raising bugs for things you'd like to see implemented is also a form of
    contributing! It helps prioritization.

## License

This project is licensed under the BSD 3-clause license, just like the
[Go project](https://golang.org/LICENSE) itself.
