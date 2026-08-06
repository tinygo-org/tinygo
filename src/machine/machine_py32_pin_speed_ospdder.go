//go:build py32 && py32_gpio_ospdder

package machine

import "device/py32"

const (
	gpioModeMask        = py32.GPIO_MODER_MODER0_Msk
	gpioOutputSpeedMask = py32.GPIO_OSPDDER_OSPEED0_Msk
)

func setPinOutputSpeed(port *py32.GPIO_Type, speed uint32, pos uint8) {
	port.OSPDDER.ReplaceBits(speed, gpioOutputSpeedMask, pos)
}
