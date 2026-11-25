//go:build py32

package runtime

import "machine"

//export Reset_Handler
func main() {
	preinit()

	configureHSI()

	initSysTick()
	machine.InitSerial()

	run()
	exit(0)
}

func sysTickFrequency() uint32 {
	return machine.CPUFrequency()
}

func putchar(c byte) {
	machine.Serial.WriteByte(c)
}

func getchar() byte {
	for machine.Serial.Buffered() == 0 {
		Gosched()
	}
	v, _ := machine.Serial.ReadByte()
	return v
}
