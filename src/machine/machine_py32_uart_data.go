//go:build py32 && !py32_uart_type && !py32_usart_split_data && !py32_usart_txe_txf

package machine

import "device/py32"

func readUSARTData(bus *py32.USART_Type) uint32 {
	return bus.DR.Get()
}

func writeUSARTData(bus *py32.USART_Type, value uint32) {
	bus.DR.Set(value)
}

func usartTXReady(bus *py32.USART_Type) bool {
	return bus.SR.Get()&py32.USART_SR_TXE != 0
}

func usartTXComplete(bus *py32.USART_Type) bool {
	return bus.SR.Get()&py32.USART_SR_TC != 0
}
