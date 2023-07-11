// Example demonstrating I2C controller / target comms.
//
// To use this example, physically connect I2C0 and I2C1.
// I2C0 will be used as the controller and I2C1 used as
// the target.
//
// In this example, the target implements a simple memory
// map, with the controller able to read and write the
// memory.

package main

import (
	"machine"
	"time"
)

const (
	targetAddress = 0x11
	maxTxSize     = 16
)

func main() {
	// Delay to enable USB monitor time to attach
	time.Sleep(3 * time.Second)

	// Controller uses default I2C pins and controller
	// mode is default
	err := controller.Configure(machine.I2CConfig{})
	if err != nil {
		panic("failed to config I2C0 as controller")
	}

	// Target uses alternate pins and target mode is
	// explicit
	err = target.Configure(machine.I2CConfig{
		Mode: machine.I2CModeTarget,
		SCL:  TARGET_SCL,
		SDA:  TARGET_SDA,
	})
	if err != nil {
		panic("failed to config I2C1 as target")
	}

	// Put welcome message directly into target memory
	copy(mem[0:], []byte("Hello World!"))
	err = target.Listen(targetAddress)
	if err != nil {
		panic("failed to listen as I2C target")
	}

	// Start the target
	go targetMain()

	// Read welcome message from target over I2C
	buf := make([]byte, 12)
	err = controller.Tx(targetAddress, []byte{0}, buf)
	if err != nil {
		println("failed to read welcome message over I2C")
		panic(err)
	}
	println("message from target:", string(buf))

	// Write (1,2,3) to the target starting at memory address 3
	println("writing (1,2,3) @ 0x3")
	err = controller.Tx(targetAddress, []byte{3, 1, 2, 3}, nil)
	if err != nil {
		println("failed to write to I2C target")
		panic(err)
	}

	time.Sleep(100 * time.Millisecond) // Wait for target to process write
	println("mem:", mem[0], mem[1], mem[2], mem[3], mem[4], mem[5])

	// Read memory address 4 from target, which should be the value 2
	buf = make([]byte, 1)
	err = controller.Tx(targetAddress, []byte{4}, buf[:1])
	if err != nil {
		println("failed to read from I2C target")
		panic(err)
	}

	if buf[0] != 2 {
		panic("read incorrect value from I2C target")
	}

	println("all done!")
	for {
		time.Sleep(1 * time.Second)
	}
}

// -- target ---

var (
	mem [256]byte
)

// targetMain implements the 'main loop' for an I2C target
func targetMain() {
	buf := make([]byte, maxTxSize)
	var ptr uint8

	for {
		evt, n, err := target.WaitForEvent(buf)
		if err != nil {
			panic(err)
		}

		switch evt {
		case machine.I2CReceive:
			if n > 0 {
				ptr = buf[0]
			}

			for o := 1; o < n; o++ {
				mem[ptr] = buf[o]
				ptr++
			}

		case machine.I2CRequest:
			target.Reply(mem[ptr:256])

		case machine.I2CFinish:
			// nothing to do

		default:
		}
	}
}
