//go:build py32 && py32_usart_split_data

package machine

import "device/py32"

func readUSARTData(bus *py32.USART_Type) uint32 {
	return bus.RDR.Get()
}

func writeUSARTData(bus *py32.USART_Type, value uint32) {
	bus.TDR.Set(value)
}

func usartTXReady(bus *py32.USART_Type) bool {
	return bus.SR.Get()&0x80 != 0
}

func usartTXComplete(bus *py32.USART_Type) bool {
	return bus.SR.Get()&0x40 != 0
}
