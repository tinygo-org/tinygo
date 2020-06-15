// +build k210

package machine

import (
	"device/kendryte"
	"runtime/interrupt"
)

func CPUFrequency() uint32 {
	return 390000000
}

type PinMode uint8

const (
	PinInput PinMode = iota
	PinOutput
	PinPWM
	PinSPI
	PinI2C = PinSPI
)

// Configure this pin with the given configuration.
func (p Pin) Configure(config PinConfig) {
}

// Set the pin to high or low.
func (p Pin) Set(high bool) {
}

// Get returns the current value of a GPIO pin.
func (p Pin) Get() bool {
	return true
}

type FPIOA struct {
	Bus *kendryte.FPIOA_Type
}

var (
	FPIOA0 = FPIOA{Bus: kendryte.FPIOA}
)

func (fpioa FPIOA) Init() {
	// Enable APB0 clock.
	kendryte.SYSCTL.CLK_EN_CENT.Set(kendryte.SYSCTL_CLK_EN_CENT_APB0_CLK_EN)

	// Enable FPIOA peripheral.
	kendryte.SYSCTL.CLK_EN_PERI.Set(kendryte.SYSCTL_CLK_EN_PERI_FPIOA_CLK_EN)
}

type UART struct {
	Bus    *kendryte.UARTHS_Type
	Buffer *RingBuffer
}

var (
	UART0 = UART{Bus: kendryte.UARTHS, Buffer: NewRingBuffer()}
)

func (uart UART) Configure(config UARTConfig) {

	div := CPUFrequency()/115200 - 1

	uart.Bus.DIV.Set(div)
	uart.Bus.TXCTRL.Set(kendryte.UARTHS_TXCTRL_TXEN)
	uart.Bus.RXCTRL.Set(kendryte.UARTHS_RXCTRL_RXEN)
	//uart.Bus.IP.Set(kendryte.UARTHS_IP_TXWM | kendryte.UARTHS_IP_RXWM)
	//uart.Bus.IE.Set(kendryte.UARTHS_IE_RXWM)

	/*intr := interrupt.New(kendryte.IRQ_UARTHS, UART0.handleInterrupt)
	intr.SetPriority(5)
	intr.Enable()*/

}

func (uart UART) handleInterrupt(interrupt.Interrupt) {
	rxdata := uart.Bus.RXDATA.Get()
	c := byte(rxdata)
	if uint32(c) != rxdata {
		// The rxdata has other bits set than just the low 8 bits. This probably
		// means that the 'empty' flag is set, which indicates there is no data
		// to be read and the byte is garbage. Ignore this byte.
		return
	}
	uart.Receive(c)
}

func (uart UART) WriteByte(c byte) {
	for uart.Bus.TXDATA.Get()&kendryte.UARTHS_TXDATA_FULL != 0 {
	}

	uart.Bus.TXDATA.Set(uint32(c))
}
