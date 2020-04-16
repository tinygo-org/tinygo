// +build machine_external

package runtime

type timeUnit int64
var tickMicros = int64(1)
var asyncScheduler = false

/*
 * Use the interface external_* below to expose to the tinygo runtime your
 * external implementation of the runtime.  External here means outside the
 * source tree.
 *
 * Example (and the //export is mandatory)
 *
 * //export runtime.external_putchar
 * func putchar(c uint8) {
 *  //your implementation goes here
 * }
*/

//
// external interface
//

//go:extern external_ticks()
func external_ticks() timeUnit

//go:extern external_sleep_ticks
func external_sleep_ticks(timeUnit)

//go:extern external_abort()
func external_abort()

//go:extern putchar
func external_putchar(uint8)

//go:extern external_postinit
func external_postinit()

//
// connect the "internal" calls to the external
//

func ticks() timeUnit {
	return external_ticks()
}

func sleepTicks(d timeUnit) {
	external_sleep_ticks(timeUnit(d))
}

func postinit() {
	external_postinit()
}

func putchar(b uint8) {
	external_putchar(b)
}

func abort() {
	external_abort()
}

//so main() outside the tree can call internal run
func Run() {
	run()
}
