package main

import (
	"device/rp"
	"machine"
)

func main() {
	useBoardHardware()

	i2c := &machine.I2C{Bus: rp.I2C0}
	if err := i2c.Configure(machine.I2CConfig{
		SDA: machine.GPIO4,
		SCL: machine.GPIO5,
	}); err != nil {
		println(err.Error())
	}

	spi := &machine.SPI{Bus: rp.SPI0}
	if err := spi.Configure(machine.SPIConfig{
		SCK: machine.GPIO2,
		SDO: machine.GPIO3,
		SDI: machine.GPIO0,
	}); err != nil {
		println(err.Error())
	}

	machine.InitADC()
	adc := machine.ADC{Pin: machine.GPIO26}
	_ = adc.Configure(machine.ADCConfig{})
	_ = adc.Get()

	rng, _ := machine.GetRNG()
	_ = rng
	_ = machine.Watchdog
}
