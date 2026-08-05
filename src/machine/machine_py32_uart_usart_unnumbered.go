//go:build py32 && py32_usart_unnumbered

package machine

import "device/py32"

func defaultUSART() *py32.USART_Type {
	return py32.USART
}
