//go:build wasip2

package runtime

import (
	exit "internal/wasi/cli/v0.2.0/exit"
	stdout "internal/wasi/cli/v0.2.0/stdout"
	monotonicclock "internal/wasi/clocks/v0.2.0/monotonic-clock"
	wallclock "internal/wasi/clocks/v0.2.0/wall-clock"
	random "internal/wasi/random/v0.2.0/random"

	"github.com/ydnar/wasm-tools-go/cm"
)

const putcharBufferSize = 120

// Using global variables to avoid heap allocation.
var (
	putcharStdout        = stdout.GetStdout()
	putcharBuffer        = [putcharBufferSize]byte{}
	putcharPosition uint = 0
)

func putchar(c byte) {
	putcharBuffer[putcharPosition] = c
	putcharPosition++
	if c == '\n' || putcharPosition >= putcharBufferSize {
		list := cm.NewList(&putcharBuffer[0], putcharPosition)
		putcharStdout.BlockingWriteAndFlush(list) // error return ignored; can't do anything anyways
		putcharPosition = 0
	}
}

func getchar() byte {
	// dummy, TODO
	return 0
}

func buffered() int {
	// dummy, TODO
	return 0
}

//go:linkname now time.now
func now() (sec int64, nsec int32, mono int64) {
	now := wallclock.Now()
	sec = int64(now.Seconds)
	nsec = int32(now.Nanoseconds)
	mono = int64(monotonicclock.Now())
	return
}

// Abort executes the wasm 'unreachable' instruction.
func abort() {
	trap()
}

//go:linkname syscall_Exit syscall.Exit
func syscall_Exit(code int) {
	exit.Exit(code != 0)
}

// TinyGo does not yet support any form of parallelism on WebAssembly, so these
// can be left empty.

//go:linkname procPin sync/atomic.runtime_procPin
func procPin() {
}

//go:linkname procUnpin sync/atomic.runtime_procUnpin
func procUnpin() {
}

func hardwareRand() (n uint64, ok bool) {
	return random.GetRandomU64(), true
}
