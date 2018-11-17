// +build wasm,!tinygo.arm,!avr

package runtime

type timeUnit int64

const tickMicros = 1

var timestamp timeUnit

// CommonWA: io_get_stdout
func _Cfunc_io_get_stdout() int32

// CommonWA: resource_write
func _Cfunc_resource_write(id int32, ptr *uint8, len int32) int32

var stdout int32

func init() {
	stdout = _Cfunc_io_get_stdout()
}

//go:export _start
func _start() {
	initAll()
}

//go:export cwa_main
func cwa_main() {
	initAll() // _start is not called by olin/cwa so has to be called here
	mainWrapper()
}

func putchar(c byte) {
	_Cfunc_resource_write(stdout, &c, 1)
}

func sleepTicks(d timeUnit) {
	// TODO: actually sleep here for the given time.
	timestamp += d
}

func ticks() timeUnit {
	return timestamp
}

// Abort executes the wasm 'unreachable' instruction.
func abort() {
	trap()
}
