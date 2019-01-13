// +build atsamd21g18

package runtime

type timeUnit int64

//go:export Reset_Handler
func main() {
	preinit()
	initAll()
	mainWrapper()
	abort()
}

func init() {
	//machine.UART0.Configure(machine.UARTConfig{})
}

func putchar(c byte) {
	//machine.UART0.WriteByte(c)
}

const tickMicros = 1024 * 32

var (
	timestamp        timeUnit // microseconds since boottime
	timerLastCounter uint64
)

//go:volatile
type isrFlag bool

var timerWakeup isrFlag

// sleepTicks should sleep for specific number of microseconds.
func sleepTicks(d timeUnit) {
}

// number of ticks (microseconds) since start.
func ticks() timeUnit {
	return tickMicros
}

// ticks are in microseconds
func timerSleep(ticks uint32) {
	timerWakeup = false
}
