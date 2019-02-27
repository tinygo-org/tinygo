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
