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
  * interface methods
  * standard library (but most packages won't work due to missing language
    features)
  * slices (partially)
  * maps (very rough, unfinished)

Not yet supported:

  * float, complex, etc.
  * garbage collection
  * defer
  * closures
  * channels
  * introspection (if it ever gets implemented)
  * ...

## Supported targets

Most targets that are supported by LLVM should be supported by this compiler.
This means amd64 (where most of the testing happens), ARM, and Cortex-M
microcontrollers.

The AVR platform (as used by the Arduino, for example) is also supported when
support for it is enabled in LLVM. However, because it is a Harvard style
architecture with different address spaces for code and data and because LLVM
turns globals into const for you (moving them to
[PROGMEM](https://www.nongnu.org/avr-libc/user-manual/pgmspace.html)) most real
programs don't work unfortunately. This can be fixed but that can be difficult
to do efficiently and hasn't been implemented yet.

## Analysis and optimizations

The goal is to reduce code size (and increase performance) by performing all
kinds of whole-program analysis passes. The official Go compiler doesn't do a
whole lot of analysis (except for escape analysis) becauses it needs to be fast,
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

## Building

Requirements:

  * Go 1.10+ (Go 1.9 has a
    [known bug](https://github.com/aykevl/tinygo/issues/2)).
  * LLVM dependencies, see the Software section in the
    [LLVM build guide](https://llvm.org/docs/GettingStarted.html#software).

First download the sources (this takes a while):

    go get -u github.com/aykevl/tinygo

You'll get an error like the following, this is expected:

    src/github.com/aykevl/llvm/bindings/go/llvm/analysis.go:17:10: fatal error: 'llvm-c/Analysis.h' file not found
    #include "llvm-c/Analysis.h" // If you are getting an error here read bindings/go/README.txt
             ^~~~~~~~~~~~~~~~~~~
    1 error generated.

To continue, you'll need to build LLVM for Go. This will take about an hour and
require a fair bit of RAM. In fact, I would recommend setting your `ld` binary
to `gold` to speed up linking, especially on systems with less than 16GB RAM.

Also, I would recommend editing build.sh and set `cmake_flags` to:

```sh
cmake_flags="../../../../.. $@ -DLLVM_EXPERIMENTAL_TARGETS_TO_BUILD=AVR -DLLVM_LINK_LLVM_DYLIB=ON"
```

This will enable the experimental AVR backend (for Arduino support) and will
make sure `tinygo` links to a shared library instead of a static library,
greatly improving link time on every rebuild. This is especially useful during
development.

After LLVM has been built, you can run an example with:

    make run-hello

For a blinky example on the PCA10040 development board, do this:

    make flash-blinky TARGET=pca10040

## License

This project is licensed under the BSD 3-clause license, just like the
[Go project](https://golang.org/LICENSE) itself.
