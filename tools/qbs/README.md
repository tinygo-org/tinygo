# qbs

Quick, broad, shallow tests for a connected board. Each test is small and
runs on every target. The aim is to find obvious problems on many boards
quickly, not to test any one feature in depth.

    go run ./tools/qbs -target <target>

The tool flashes one program per class with `tinygo flash`, collects results
over the board's serial port and checks them. The tool builds with the
standard Go compiler. The programs it flashes build with TinyGo.

Classes:

- timing: sleep accuracy, a monotonic clock, loop speed, and the device clock
  against the host clock
- compute: the same computations (integers, bits, sorting, maps, hashes,
  encodings, strings, float32 and float64, math, strconv, fmt) run on the
  device and on the host and the results are compared
- gc: large blocks, linked lists, slice and string growth, map churn,
  fragmentation and collection cycles, sized from the free heap. The map and
  fragmentation tests are not built for AVR. A conservative collector can keep
  a whole linked structure alive through one stale pointer, so these tests
  fail only when memory use keeps growing over repeated runs.
- serial: the default serial port. Bytes sent to the board come back
  unchanged (echo), overflowing the RX buffer keeps the first bytes and
  receiving still works (burst), 2000 numbered lines arrive in order (seq), all
  256 byte values arrive intact (bytes), and output and input work together
  with goroutines and allocation (conc, not built on targets without a
  scheduler)

Use `-v` to see the values each test sent, `-port` if the board is not found,
`-tinygo` for a different tinygo binary and `-path-prepend` to put a flashing
tool first on PATH. Names after the flags select classes.

Use `-emulator` to run the programs in the emulator of the target instead of
flashing a board, for example `go run ./tools/qbs -target wasip1 -emulator`.
The runner then starts `tinygo run` for each class and talks to the program
through its standard input and output. The emulator, such as wasmtime or QEMU,
must be installed.

A class is skipped when its program does not fit on the target. A test is
skipped when the target cannot run it, for example when the garbage collector
does not reclaim memory.

## Adding a test

- A computation for the compute class goes in `cases`. It must give the same
  result on the host and the device. Use `ULP` for float results that may
  differ in the last bits.
- Any other test is a `dev.Test` in the program of its class. Use `t.Fail`,
  `t.Skip`, `t.Val` and `t.Raw`. A test that needs the host to judge the
  result sends values, ticks or raw bytes and gets a check in the host code.
  A test that needs the host to send data gets a driver, which runs on the
  host when the test begins (see `serial.go`).
- A new class is a new directory with a program that calls `dev.Main`, and an
  entry in the class list in `main.go`.
