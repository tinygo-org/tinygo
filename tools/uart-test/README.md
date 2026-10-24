# Serial test programs

Checks the default serial port of a connected board.

    go run ./tools/uart-test -target <target>

The tool flashes each program in the subdirectories with `tinygo flash`, then
talks to it over the board's serial port and checks the output. The tool
builds with the standard Go compiler. The programs it flashes build with
TinyGo. Tests:

- echo: bytes sent to the board come back unchanged
- burst: overflowing the RX buffer keeps the first bytes and receiving still works
- seq: 2000 numbered lines arrive in order
- bytes: all 256 byte values arrive intact
- conc: serial output and input work together with goroutines and allocation

Use `-port` if the board is not found, `-tinygo` for a different tinygo
binary and `-path-prepend` to put a flashing tool first on PATH. Names after
the flags select tests.
