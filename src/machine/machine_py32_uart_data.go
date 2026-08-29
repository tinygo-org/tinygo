//go:build py32 && !py32_uart_type && !py32_usart_split_data

package machine

import "device/py32"

func readUSARTData(bus *py32.USART_Type) uint32 {
	return bus.DR.Get()
}

func writeUSARTData(bus *py32.USART_Type, value uint32) {
	bus.DR.Set(value)
}

func usartTXReady(bus *py32.USART_Type) bool {
	const txEmpty = 1 << 7 // TXE or TXE_TXFNF, depending on the family.
	return bus.SR.Get()&txEmpty != 0
}

func usartTXComplete(bus *py32.USART_Type) bool {
	return bus.SR.Get()&py32.USART_SR_TC != 0
}
