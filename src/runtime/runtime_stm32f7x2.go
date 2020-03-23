// +build stm32,stm32f7x2

package runtime

const tickMicros = 1000

var (
	// tick in milliseconds
	tickCount timeUnit
)

func init() {
	//initCLK()
	//initTIM3()
	//machine.UART2.Configure(machine.UARTConfig{})
	//initTIM7()
}

func putchar(c byte) {
	//machine.UART2.WriteByte(c)
}

const asyncScheduler = false

// sleepTicks should sleep for specific number of microseconds.
func sleepTicks(d timeUnit) {
	//timerSleep(uint32(d))
}

// number of ticks (microseconds) since start.
func ticks() timeUnit {
	// milliseconds to microseconds
	return tickCount * 1000
}
