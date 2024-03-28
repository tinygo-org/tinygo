//go:build wasip2

package runtime

import (
	"github.com/ydnar/wasm-tools-go/cm"
	"github.com/ydnar/wasm-tools-go/wasi/cli/exit"
	"github.com/ydnar/wasm-tools-go/wasi/cli/stdout"
	monotonicclock "github.com/ydnar/wasm-tools-go/wasi/clocks/monotonic-clock"
	wallclock "github.com/ydnar/wasm-tools-go/wasi/clocks/wall-clock"
	"github.com/ydnar/wasm-tools-go/wasi/random/random"
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
		result := putcharStdout.BlockingWriteAndFlush(list)
		if err := result.Err(); err != nil {
			// TODO(ydnar): handle error case
			panic(*err)
		}
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
