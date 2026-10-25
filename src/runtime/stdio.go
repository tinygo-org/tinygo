//go:build baremetal && !(cortexm && qemu) && !attiny && !avrtiny && !gameboyadvance && !mk20d7 && !mk66f18 && !mimxrt1062 && !virt && !uefi

package runtime

// The targets excluded by the build tag define their own console functions.

import "machine"

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

func buffered() int {
	return machine.Serial.Buffered()
}
