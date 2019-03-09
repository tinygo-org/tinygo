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
