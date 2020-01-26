0.11.0
---
* **command line**
  - add support for QEMU in `gdb` subcommand
  - use builtin Clang when building statically, dropping the clang-9 dependency
  - search for default serial port on both macOS and Linux
  - windows: support `tinygo flash` directly by using win32 wmi
* **compiler**
  - add location information to the IR checker
  - make reflection sidetables constant globals
  - improve error locations in goroutine lowering
  - interp: improve support for maps with string keys
  - interp: add runtime fallback for mapassign operations
* **standard library**
  - `machine`: add support for `SPI.Tx()` on play.tinygo.org
  - `machine`: rename `CPU_FREQUENCY` to `CPUFrequency()`
* **targets**
  - `adafruit-pybadge`: add Adafruit Pybadge
  - `arduino-nano33`: allow simulation on play.tinygo.org
  - `arduino-nano33`: fix default SPI pin numbers to be D13/D11/D12
  - `circuitplay-express`: allow simulation on play.tinygo.org
  - `hifive1-qemu`: add target for testing RISC-V bare metal in QEMU
  - `riscv`: fix heap corruption due to changes in LLVM 9
  - `riscv`: add support for compiler-rt
  - `qemu`: rename to `cortex-m-qemu`

0.10.0
---
* **command line**
  - halt GDB after flashing with `gdb` subcommand
  - fix a crash when using `-ocd-output`
  - add `info` subcommand
  - add `-programmer` flag
* **builder**
  - macos: use llvm@8 instead of just llvm in paths
  - add `linkerscript` key to target JSON files
  - write a symbol table when writing out the compiler-rt lib
  - make Clang header detection more robust
  - switch to LLVM 9
* **compiler**
  - fix interface miscompilation with reflect
  - fix miscompile of static goroutine calls to closures
  - fix `todo: store` panic
  - fix incorrect starting value for optimized allocations in a loop
  - optimize coroutines on non-Cortex-M targets
  - fix crash for programs which have heap allocations but never hit the GC
  - add support for async interface calls
  - fix inserting non-const values in a const global
  - interp: improve error reporting
  - interp: implement comparing ptrtoint to 0
* **cgo**
  - improve diagnostics
  - implement the constant parser (for `#define`) as a real parser
  - rename reserved field names such as `type`
  - avoid `"unsafe" imported but not used` error
  - include all enums in the CGo Go AST
  - add support for nested structs and unions
  - implement `#cgo CFLAGS`
* **standard library**
  - `reflect`: add implementation of array alignment
  - `runtime`: improve scheduler performance when no goroutines are queued
  - `runtime`: add blocking select
  - `runtime`: implement interface equality in non-trivial cases
  - `runtime`: add AdjustTimeOffset to update current time
  - `runtime`: only implement CountString for required platforms
  - `runtime`: use MSP/PSP registers for scheduling on Cortex-M
* **targets**
  - `arm`: add system timer registers
  - `atmega`: add port C GPIO support
  - `atsamd21`: correct handling of pins >= 32
  - `atsamd21`: i2s initialization fixes
  - `atsamd51`: fix clock init code
  - `atsamd51`: correct initialization for RTC
  - `atsamd51`: fix pin function selection
  - `atsamd51`: pin method cleanup
  - `atsamd51`: allow setting pin mode for each of the SPI pins
  - `atsamd51`: correct channel init and pin map for ADC based on ItsyBitsy-M4
  - `feather-m4`: add Adafruit Feather M4 board
  - `hifive1b`: add support for SPI1
  - `hifive1b`: fix compiling in simulation
  - `linux`: fix time on arm32
  - `metro-m4`: add support for Adafruit Metro M4 Express Airlift board
  - `metro-m4`: fixes for UART2
  - `pinetime-devkit0`: add support for the PineTime dev kit
  - `x9pro`: add support for this smartwatch
  - `pca10040-s132v6`: add support for SoftDevice
  - `pca10056-s140v7`: add support for SoftDevice
  - `arduino-nano33`: added SPI1 connected to NINA-W102 chip on Arduino Nano 33 IOT

0.9.0
---
* **command line**
  - implement 1200-baud UART bootloader reset when flashing boards that support
    it
  - flash using mass-storage device for boards that support it
  - implement `tinygo env`
  - add support for Windows (but not yet producing Windows binaries)
  - add Go version to `tinygo env`
  - update SVD files for up-to-date peripheral interfaces
* **compiler**
  - add `//go:align` pragma
  - fix bug related to type aliases
  - add support for buffered channels
  - remove incorrect reflect optimization
  - implement copying slices in init interpretation
  - add support for constant indices with a named type
  - add support for recursive types like linked lists
  - fix miscompile of function nil panics
  - fix bug related to goroutines
* **standard library**
  - `machine`: do not check for nil slices in `SPI.Tx`
  - `reflectlite`: add support for Go 1.13
  - `runtime`: implement `internal/bytealg.CountString`
  - `sync`: properly handle nil `New` func in `sync.Pool`
* **targets**
  - `arduino`: fix .bss section initialization
  - `fe310`: implement `Pin.Get`
  - `gameboy-advance`: support directly outputting .gba files
  - `samd`: reduce code size by avoiding reflection
  - `samd21`: do not hardcode pin numbers for peripherals
  - `stm32f103`: avoid issue with `time.Sleep` less than 200µs

0.8.0
---
* **command line**
  - fix parsing of beta Go versions
  - check the major/minor installed version of Go before compiling
  - validate `-target` flag better to not panic on an invalid target
* **compiler**
  - implement full slice expression: `s[:2:4]`
  - fix a crash when storing a linked list in an interface
  - fix comparing struct types by making type IDs more unique
  - fix some bugs in IR generation
  - add support for linked lists in reflect data
  - implement `[]rune` to string conversion
  - implement support for `go` on func values
* **standard library**
  - `reflect`: add support for named types
  - `reflect`: add support for `t.Bits()`
  - `reflect`: add basic support for `t.AssignableTo()`
  - `reflect`: implement `t.Align()`
  - `reflect`: add support for struct types
  - `reflect`: fix bug in `v.IsNil` and `v.Pointer` for addressable values
  - `reflect`: implement support for array types
  - `reflect`: implement `t.Comparable()`
  - `runtime`: implement stack-based scheduler
  - `runtime`: fix bug in the sleep queue of the scheduler
  - `runtime`: implement `memcpy` for Cortex-M
  - `testing`: implement stub `testing.B` struct
  - `testing`: add common test logging methods such as Errorf/Fatalf/Printf
* **targets**
  - `386`: add support for linux/386 syscalls
  - `atsamd21`: make SPI pins configurable so that multiple SPI ports can be
    used
  - `atsamd21`: correct issue with invalid first reading coming from ADC
  - `atsamd21`: add support for reset-to-bootloader using 1200baud over USB-CDC
  - `atsamd21`: make pin selection more flexible for peripherals
  - `atsamd21`: fix minimum delay in `time.Sleep`
  - `atsamd51`: fix minimum delay in `time.Sleep`
  - `nrf`: improve SPI write-only speed, by making use of double buffering
  - `stm32f103`: fix SPI frequency selection
  - `stm32f103`: add machine.Pin.Get method for reading GPIO values
  - `stm32f103`: allow board specific UART usage
  - `nucleo-f103rb`: add support for NUCLEO-F103RB board
  - `itsybitsy-m4`: add support for this board with a SAMD51 family chip
  - `cortex-m`: add support for `arm.SystemReset()`
  - `gameboy-advance`: add initial support for the GameBoy Advance
  - `wasm`: add `//go:wasm-module` magic comment to set the wasm module name
  - `wasm`: add syscall/js.valueSetIndex support
  - `wasm`: add syscall/js.valueInvoke support

0.7.1
---
* **targets**
  - `atsamd21`: add support for the `-port` flag in the flash subcommand

0.7.0
---
* **command line**
  - try more locations to find Clang built-in headers
  - add support for `tinygo test`
  - build current directory if no package is specified
  - support custom .json target spec with `-target` flag
  - use zversion.go to detect version of GOROOT version
  - make initial heap size configurable for some targets (currently WebAssembly
    only)
* **cgo**
  - add support for bitfields using generated getters and setters
  - add support for anonymous structs
* **compiler**
  - show an error instead of panicking on duplicate function definitions
  - allow packages like github.com/tinygo-org/tinygo/src/\* by aliasing it
  - remove `//go:volatile` support  
    It has been replaced with the runtime/volatile package.
  - allow poiners in map keys
  - support non-constant syscall numbers
  - implement non-blocking selects
  - add support for the `-tags` flag
  - add support for `string` to `[]rune` conversion
  - implement a portable conservative garbage collector (with support for wasm)
  - add the `//go:noinline` pragma
* **standard library**
  - `os`: add `os.Exit` and `syscall.Exit`
  - `os`: add several stubs
  - `runtime`: fix heap corruption in conservative GC
  - `runtime`: add support for math intrinsics where supported, massively
    speeding up some benchmarks
  - `testing`: add basic support for testing
* **targets**
  - add support for a generic target that calls `__tinygo_*` functions for
    peripheral access
  - `arduino-nano33`: add support for this board
  - `hifive1`: add support for this RISC-V board
  - `reelboard`: add e-paper pins
  - `reelboard`: add `PowerSupplyActive` to enable voltage for on-board devices
  - `wasm`: put the stack at the start of linear memory, to detect stack
    overflows

0.6.0
---
* **command line**
  - some portability improvements
  - make `$GOROOT` more robust and configurable
  - check for Clang at the Homebrew install location as fallback
* **compiler driver**
  - support multiple variations of LLVM commands, for non-Debian distributions
* **compiler**
  - improve code quality in multiple ways
  - make panic configurable, adding trap on panic
  - refactor many internal parts of the compiler
  - print all errors encountered during compilation
  - implement calling function values of a named type
  - implement returning values from blocking functions
  - allow larger-than-int values to be sent across a channel
  - implement complex arithmetic
  - improve hashmap support
  - add debuginfo for function arguments
  - insert nil checks on stores (increasing code size)
  - implement volatile operations as compiler builtins
  - add `//go:inline` pragma
  - add build tags for the Go stdlib version
* **cgo**
  - implement `char`, `enum` and `void*` types
  - support `#include` for builtin headers
  - improve typedef/struct/enum support
  - only include symbols that are necessary, for broader support
  - mark external function args as `nocapture`
  - implement support for some `#define` constants
  - implement support for multiple CGo files in a single package
- **standard library**
  - `machine`: remove microbit matrix (moved to drivers repository)
  - `machine`: refactor pins to use `Pin` type instead of `GPIO`
  - `runtime`: print more interface types on panic, including `error`
* **targets**
  - `arm`: print an error on HardFault (including stack overflows)
  - `atsamd21`: fix a bug in the ADC peripheral
  - `atsamd21`: add support for I2S
  - `feather-m0`: add support for this board
  - `nrf51`: fix a bug in I2C
  - `stm32f103xx`: fix a bug in I2C
  - `syscall`: implement `Exit` on unix
  - `trinket-m0`: add support for this board
  - `wasm`: make _main_ example smaller
  - `wasm`: don't cache wasm file in the server, for ease of debugging
  - `wasm`: work around bug #41508 that caused a deadlock while linking
  - `wasm`: add support for `js.FuncOf`

0.5.0
---
- **compiler driver**
  - use `wasm-ld` instead of `wasm-ld-8` on macOS
  - drop dependency on `llvm-ar`
  - fix linker script includes when running outside `TINYGOROOT`
- **compiler**
  - switch to LLVM 8
  - add support for the Go 1.12 standard library (Go 1.11 is still supported)
  - work around lack of escape analysis due to nil checks
  - implement casting named structs and pointers to them
  - fix int casting to use the source signedness
  - fix some bugs around `make([]T, …)` with uncommon index types
  - some other optimizations
  - support interface asserts in interp for "math/rand" support
  - resolve all func value targets at compile time (wasm-only at the moment)
- **cgo**
  - improve diagnostics
  - implement C `struct`, `union`, and arrays
  - fix CGo-related crash in libclang
  - implement `C.struct_` types
- **targets**
  - all baremetal: pretend to be linux/arm instead of js/wasm
  - `avr`: improve `uintptr` support
  - `cortexm`: implement memmove intrinsic generated by LLVM
  - `cortexm`: use the lld linker instead of `arm-none-eabi-ld`
  - `darwin`: use custom syscall package that links to libSystem.dylib
  - `microbit`: add blink example
  - `samd21`: support I2C1
  - `samd21`: machine/atsamd21: correct pad/pin handling when using both UART
     and USBCDC interfaces at same time
  - `stm32f4discovery`: add support for this board
  - `wasm`: support async func values
  - `wasm`: improve documentation and add extra example

0.4.1
---
- **compiler**
  - fix `objcopy` replacement to include the .data section in the firmware image
  - use `llvm-ar-7` on Linux to fix the Docker image

0.4.0
---
- **compiler**
  - switch to the hardfloat ABI on ARM, which is more widely used
  - avoid a dependency on `objcopy` (`arm-none-eabi-objcopy` etc.)
  - fix a bug in `make([]T, n)` where `n` is 64-bits on a 32-bit platform
  - adapt to a change in the AVR backend in LLVM 8
  - directly support the .uf2 firmware format as used on Adafruit boards
  - fix a bug when calling `panic()` at init time outside of the main package
  - implement nil checks, which results in a ~5% increase in code size
  - inline slice bounds checking, which results in a ~1% decrease in code size
- **targets**
  - `samd21`: fix a bug in port B pins
  - `samd21`: implement SPI peripheral
  - `samd21`: implement ADC peripheral
  - `stm32`: fix a bug in timekeeping
  - `wasm`: fix a bug in `wasm_exec.js` that caused corruption in linear memory
     when running on Node.js.

0.3.0
---
- **compiler**
  - remove old `-initinterp` flag
  - add support for macOS
- **cgo**
  - add support for bool/float/complex types
- **standard library**
  - `device/arm`: add support to disable/enable hardware interrupts
  - `machine`: add CPU frequency for nrf-based boards
  - `syscall`: add support for darwin/amd64
- **targets**
  - `circuitplay_express`: add support for this board
  - `microbit`: add regular pin constants
  - `samd21`: fix time function for goroutine support
  - `samd21`: add support for USB-CDC (serial over USB)
  - `samd21`: add support for pins in port B
  - `samd21`: add support for pullup and pulldown pins
  - `wasm`: add support for Safari in example


0.2.0
---
- **command line**
  - add version subcommand
- **compiler**
  - fix a bug in floating point comparisons with NaN values
  - fix a bug when calling `panic` in package initialization code
  - add support for comparing `complex64` and `complex128`
- **cgo**
  - add support for external globals
  - add support for pointers and function pointers
- **standard library**
  - `fmt`: initial support, `fmt.Println` works
  - `math`: support for most/all functions
  - `os`: initial support (only stdin/stdout/stderr)
  - `reflect`: initial support
  - `syscall`: add support for amd64, arm, and arm64
