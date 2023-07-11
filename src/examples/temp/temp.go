// Read the internal temperature sensor of the chip.

package main

import (
	"fmt"
	"machine"
	"time"
)

type celsius float32

func (c celsius) String() string {
	return fmt.Sprintf("%4.1f℃", c)
}

func main() {
	for {
		temp := celsius(float32(machine.ReadTemperature()) / 1000)
		println("temperature:", temp.String())
		time.Sleep(time.Second)
	}
}
