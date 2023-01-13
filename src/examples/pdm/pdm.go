package main

import (
	"fmt"
	"machine"
)

var (
	audio = make([]int16, 16)
	pdm   = machine.PDM{}
)

func main() {
	machine.BUTTONA.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
	err := pdm.Configure(machine.PDMConfig{CLK: machine.PDM_CLK_PIN, DIN: machine.PDM_DIN_PIN})
	if err != nil {
		panic(fmt.Sprintf("Failed to configure PDM:%v", err))
	}

	for {
		if machine.BUTTONA.Get() {
			println("Recording new audio clip into memory")
			pdm.Read(audio)
			println(fmt.Sprintf("Recorded new audio clip into memory: %v", audio))
		}
	}
}
