// +build avr nrf sam sifive stm32 k210

package machine

type UARTConfig struct {
	BaudRate uint32
	TX       Pin
	RX       Pin
}
