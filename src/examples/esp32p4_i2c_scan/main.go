// I2C bus scan for ESP32-P4 (CrowPanel 10.1").
//
// Scans I2C0 on SDA=GPIO7 / SCL=GPIO8 at 100kHz and prints any
// responding device addresses. The GT911 touch controller should
// appear at 0x5D (or 0x14).
//
// Flash with:
//
//	tinygo flash -target crowpanel-esp32p4-10-1 ./src/examples/esp32p4_i2c_scan/

package main

import (
	"machine"
	"time"
)

func main() {
	time.Sleep(500 * time.Millisecond)
	println("I2C scan: SDA=GPIO7 SCL=GPIO8 100kHz")

	println("calling Configure...")
	err := machine.I2C0.Configure(machine.I2CConfig{
		Frequency: 100 * machine.KHz,
		SDA:       machine.SDA_PIN,
		SCL:       machine.SCL_PIN,
	})
	if err != nil {
		println("I2C configure error:", err.Error())
		for {
		}
	}
	println("Configure OK")

	// Probe addr 0x01 first as a diagnostic to confirm Tx() returns quickly.
	println("probing 0x01...")
	if err := machine.I2C0.Tx(1, nil, nil); err != nil {
		println("0x01 err:", err.Error())
		// Register dump to diagnose I2C hardware state after failed transaction.
		println("CTR      =", machine.I2C0.Bus.CTR.Get())
		println("CLK_CONF =", machine.I2C0.Bus.CLK_CONF.Get())
		println("INT_RAW  =", machine.I2C0.Bus.INT_RAW.Get())
		println("SR       =", machine.I2C0.Bus.SR.Get())
		println("COMD0    =", machine.I2C0.Bus.COMD0.Get())
		println("COMD1    =", machine.I2C0.Bus.COMD1.Get())
		println("SCL_SP   =", machine.I2C0.Bus.SCL_SP_CONF.Get())
		println("SCL_STR  =", machine.I2C0.Bus.SCL_STRETCH_CONF.Get())
	} else {
		println("0x01 ACK")
	}
	println("probe done, starting full scan")

	found := 0
	for addr := uint8(1); addr < 128; addr++ {
		if machine.I2C0.Tx(uint16(addr), nil, nil) == nil {
			print("found 0x")
			printHex(addr)
			println()
			found++
		}
	}

	if found == 0 {
		println("no devices found")
	} else {
		print("scan complete: ")
		println(found, "device(s)")
	}

	for {
		time.Sleep(time.Hour)
	}
}

func printHex(v uint8) {
	const digits = "0123456789abcdef"
	print(string([]byte{digits[v>>4], digits[v&0xf]}))
}
