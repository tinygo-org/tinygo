// +build stm32

package machine

// Peripheral abstraction layer for the stm32.

type PinMode uint8

const (
	portA Pin = iota * 16
	portB
	portC
	portD
	portE
	portF
	portG
	portH
)
