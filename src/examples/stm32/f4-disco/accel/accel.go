package main

// test the stm32f4 discovery board accelerometer, LIS302DL

import (
	"encoding/hex"
	"fmt"
	"machine"
	"time"
)

var ctrl1, status byte
var x, y, z int8

func initAccel() {
	var SPI1Conf = machine.SPIConfig{Frequency: 1312500}
	machine.MEMS_ACCEL_CS.Configure(machine.PinConfig{Mode: machine.PinOutput})
	// Set the /CS line high in idle
	machine.MEMS_ACCEL_CS.Set(true)
	machine.SPI1.Configure(SPI1Conf)
	// LIS302DL needs a short period to boot up
	time.Sleep(time.Millisecond * 5)
}

func sendSPIAccel(cmd, resp []byte) error {
	// Pull /CS low to enable accelerometer
	machine.MEMS_ACCEL_CS.Set(false)
	err := machine.SPI1.Tx(cmd, resp)
	// Pull /CS high when done
	machine.MEMS_ACCEL_CS.Set(true)

	return err
}

// Write a 1-byte value to a register in the accelerometer
func setAccelReg(cmd, val byte) (byte, error) {
	cmdArr := []byte{cmd, val}
	respArr := []byte{0x00, 0x00}
	err := sendSPIAccel(cmdArr, respArr)
	if err != nil {
		return 0xff, err
	}
	return respArr[1], nil
}

// Read a 1-byte register from the accelerometer
func getAccelReg(cmd byte) (byte, error) {
	// just use the setAccelReg call with the read | /write flag set
	cmd = cmd | 0x80
	return setAccelReg(cmd, 0x00)
}

// Get accelerometer ID
func getAccelID() (byte, error) {
	return getAccelReg(0x0F)
}

// Get accelerometer status
func getAccelStatus() (byte, error) {
	return getAccelReg(0x27)
}

// Get accelerometer control registers
func getAccelCtrl(reg byte) (byte, error) {
	return getAccelReg(0x20 + (reg%3 - 1))
}

// Set accelerometer control registers
func setAccelCtrl(reg, val byte) (byte, error) {
	return setAccelReg(0x20+(reg%3-1), val)
}

// Get the current accelerometer X Y and Z values
func getAccelValues() (int8, int8, int8, error) {
	// Get accelerometer status, read+incr command
	cmd := []byte{0x29 | 0x80, 0x00, 0x00, 0x00, 0x00, 0x00}
	resp := []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	err := sendSPIAccel(cmd, resp)
	if err != nil {
		return 0, 0, 0, err
	}
	if 1 == 0 {
		println(hex.Dump(resp))
	}
	return int8(resp[1]), int8(resp[3]), int8(resp[5]), nil
}

func listenForAccel(timing time.Duration) {
	// Wake up the accelerometer from power-down mode
	_, err := setAccelReg(0x20, 0x47)
	if err != nil {
		println("Unable to wake up accelerometer")
		return
	}
	// Scan for accelation values
	for {
		// Read status register till there's data, then read the data, then sleep
		ctrl1, err = getAccelCtrl(1)
		if err != nil {
			println("Error rx-ing ctrl-1")
		}
		status, err = getAccelStatus()
		if err != nil {
			println("Error rx-ing status")
		} else {
			//status |= 0x3
			if status&0x3 != 0 || 1 > 0 {
				// data available
				x, y, z, err = getAccelValues()
				if err != nil {
					println("Error rx-ing X/Y/Z")
				}
			}
		}
		time.Sleep(time.Millisecond * timing)
	}
}

func displayAccel(timing time.Duration) {
	i := 0
	for {
		// Show the output. append spaces to values so they are cleared on screen
		fmt.Printf("%03d : ctrl1=%08b\tstatus=%08b\tx=% 3d  \ty=% 3d  \tz=% 3d  \r",
			i%1000, ctrl1, status, x, y, z)
		time.Sleep(time.Millisecond * timing)
		i = i + 1
	}
}

func programID(msg string, timing time.Duration) {
	for {
		println("\n" + msg)
		time.Sleep(time.Millisecond * timing)
	}
}

func main() {
	go programID("accel", 5000)
	initAccel()
	// See if accelerometer is responding on SPI
	id, err := getAccelID()
	if err != nil {
		println("Unable to read accelerometer ID; accelerometer setup error")
	} else {
		fmt.Printf("Accelerometer ID : %02x (should be 0x3b for LIS302DL)\n", id)
		go listenForAccel(10) // check every ~200th of second
		displayAccel(200)
	}
}
