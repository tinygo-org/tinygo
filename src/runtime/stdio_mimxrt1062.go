//go:build mimxrt1062

package runtime

import "machine"

func putchar(c byte) {
	// Drop early output so a print from a fault handler before InitSerial
	// does not cause a second fault.
	if !serialReady() {
		return
	}
	machine.Serial.WriteByte(c)
}

func getchar() byte {
	for !serialReady() || machine.Serial.Buffered() == 0 {
		Gosched()
	}
	v, _ := machine.Serial.ReadByte()
	return v
}

func buffered() int {
	if !serialReady() {
		return 0
	}
	return machine.Serial.Buffered()
}

func initStdio() {
	machine.InitSerial()
}
