# TinyGo - Go compiler for small places

[![Linux](https://github.com/tinygo-org/tinygo/actions/workflows/linux.yml/badge.svg?branch=dev)](https://github.com/tinygo-org/tinygo/actions/workflows/linux.yml) [![macOS](https://github.com/tinygo-org/tinygo/actions/workflows/build-macos.yml/badge.svg?branch=dev)](https://github.com/tinygo-org/tinygo/actions/workflows/build-macos.yml) [![Windows](https://github.com/tinygo-org/tinygo/actions/workflows/windows.yml/badge.svg?branch=dev)](https://github.com/tinygo-org/tinygo/actions/workflows/windows.yml) [![Docker](https://github.com/tinygo-org/tinygo/actions/workflows/docker.yml/badge.svg?branch=dev)](https://github.com/tinygo-org/tinygo/actions/workflows/docker.yml) [![CircleCI](https://circleci.com/gh/tinygo-org/tinygo/tree/dev.svg?style=svg)](https://circleci.com/gh/tinygo-org/tinygo/tree/dev)

TinyGo is a Go compiler intended for use in small places such as microcontrollers, WebAssembly (wasm/wasi), and command-line tools.

It reuses libraries used by the [Go language tools](https://golang.org/pkg/go/) alongside [LLVM](http://llvm.org) to provide an alternative way to compile programs written in the Go programming language.

## Embedded

Here is an example program that blinks the built-in LED when run directly on any supported board with onboard LED:

```go
package main

import (
    "machine"
    "time"
)

func main() {
    led := machine.LED
    led.Configure(machine.PinConfig{Mode: machine.PinOutput})
    for {
        led.Low()
        time.Sleep(time.Millisecond * 1000)

        led.High()
        time.Sleep(time.Millisecond * 1000)
    }
}
```

The above program can be compiled and run without modification on an Arduino Uno, an Adafruit ItsyBitsy M0, or any of the supported boards that have a built-in LED, just by setting the correct TinyGo compiler target. For example, this compiles and flashes an Arduino Uno:

```shell
tinygo flash -target arduino examples/blinky1
```

## WebAssembly

TinyGo is very useful for compiling programs both for use in browsers (WASM) as well as for use on servers and other edge devices (WASI).

TinyGo programs can run in Fastly Compute@Edge (https://developer.fastly.com/learning/compute/go/), Fermyon Spin (https://developer.fermyon.com/spin/go-components), wazero (https://wazero.io/languages/tinygo/) and many other WebAssembly runtimes.

Here is a small TinyGo program for use by a WASI host application:

```go
package main

//go:wasm-module yourmodulename
//export add
func add(x, y uint32) uint32 {
	return x + y
}

// main is required for the `wasi` target, even if it isn't used.
func main() {}
```

This compiles the above TinyGo program for use on any WASI runtime:

```shell
tinygo build -o main.wasm -target=wasi main.go
```

## Installation

See the [getting started instructions](https://tinygo.org/getting-started/) for information on how to install TinyGo, as well as how to run the TinyGo compiler using our Docker container.

## Supported targets

### Embedded

You can compile TinyGo programs for over 94 different microcontroller boards.

For more information, please see https://tinygo.org/docs/reference/microcontrollers/

### WebAssembly

TinyGo programs can be compiled for both WASM and WASI targets.

For more information, see https://tinygo.org/docs/guides/webassembly/

### Operating Systems

You can also compile programs for Linux, macOS, and Windows targets.

For more information:

- Linux https://tinygo.org/docs/guides/linux/

- macOS https://tinygo.org/docs/guides/macos/

- Windows https://tinygo.org/docs/guides/windows/

## Currently supported features:

For a description of currently supported Go language features, please see [https://tinygo.org/lang-support/](https://tinygo.org/lang-support/).

## Documentation

Documentation is located on our web site at [https://tinygo.org/](https://tinygo.org/).

You can find the web site code at [https://github.com/tinygo-org/tinygo-site](https://github.com/tinygo-org/tinygo-site).

## Getting help

If you're looking for a more interactive way to discuss TinyGo usage or
development, we have a [#TinyGo channel](https://gophers.slack.com/messages/CDJD3SUP6/)
on the [Gophers Slack](https://gophers.slack.com).

If you need an invitation for the Gophers Slack, you can generate one here which
should arrive fairly quickly (under 1 min): https://invite.slack.golangbridge.org

## Contributing

Your contributions are welcome!

Please take a look at our [Contributing](https://tinygo.org/docs/guides/contributing/) page on our web site for details.

## Project Scope

Goals:

* Have very small binary sizes. Don't pay for what you don't use.
* Support for most common microcontroller boards.
* Be usable on the web using WebAssembly.
* Good CGo support, with no more overhead than a regular function call.
* Support most standard library packages and compile most Go code without modification.

Non-goals:

* Be efficient while using zillions of goroutines. However, good goroutine support is certainly a goal.
* Be as fast as `gc`. However, LLVM will probably be better at optimizing certain things so TinyGo might actually turn out to be faster for number crunching.
* Be able to compile every Go program out there.

## Why this project exists

> We never expected Go to be an embedded language and so its got serious problems...

-- Rob Pike, [GopherCon 2014 Opening Keynote](https://www.youtube.com/watch?v=VoS7DsT1rdM&feature=youtu.be&t=2799)

TinyGo is a project to bring Go to microcontrollers and small systems with a single processor core. It is similar to [emgo](https://github.com/ziutek/emgo) but a major difference is that we want to keep the Go memory model (which implies garbage collection of some sort). Another difference is that TinyGo uses LLVM internally instead of emitting C, which hopefully leads to smaller and more efficient code and certainly leads to more flexibility.

The original reasoning was: if [Python](https://micropython.org/) can run on microcontrollers, then certainly [Go](https://golang.org/) should be able to run on even lower level micros.

## License

This project is licensed under the BSD 3-clause license, just like the [Go project](https://golang.org/LICENSE) itself.

Some code has been copied from the LLVM project and is therefore licensed under [a variant of the Apache 2.0 license](http://releases.llvm.org/11.0.0/LICENSE.TXT). This has been clearly indicated in the header of these files.

Some code has been copied and/or ported from Paul Stoffregen's Teensy libraries and is therefore licensed under PJRC's license. This has been clearly indicated in the header of these files.
