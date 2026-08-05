//go:build py32 && !py32_gpio_ospdder

package machine

import "device/py32"

func setPinOutputSpeed(port *py32.GPIO_Type, speed uint32, pos uint8) {
	port.OSPEEDR.ReplaceBits(speed, gpioOutputSpeedMask, pos)
}
