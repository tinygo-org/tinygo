//go:build nrf
// +build nrf

package machine

import (
	"device/nrf"
	"runtime/interrupt"
	"unsafe"
)

// EasyDMA UART on the NRF.
type UARTE struct {
	*nrf.UARTE_Type
	Buffer *RingBuffer
	interrupt.Interrupt
	rxbuffer [1]uint8
}

//go:linkname gosched runtime.Gosched
func gosched()

func (uart *UARTE) setPins(tx, rx Pin) {
	tx.Configure(PinConfig{Mode: PinOutput})
	rx.Configure(PinConfig{Mode: PinInput})
	uart.UARTE_Type.PSEL.TXD.Set(uint32(tx))
	uart.UARTE_Type.PSEL.RXD.Set(uint32(rx))
}

func (uart *UARTE) setRtsCtsPins(rts, cts Pin) {
	rts.Configure(PinConfig{Mode: PinOutput})
	cts.Configure(PinConfig{Mode: PinInput})
	uart.UARTE_Type.PSEL.RTS.Set(uint32(rts))
	uart.UARTE_Type.PSEL.CTS.Set(uint32(cts))
}

// Configure the UART.
func (uart *UARTE) Configure(config UARTConfig) {
//	uart.UARTE_Type.ENABLE.Set(nrf.UARTE_ENABLE_ENABLE_Disabled)

	// Default baud rate to 115200.
	if config.BaudRate == 0 {
		config.BaudRate = 115200
	}

	uart.SetBaudRate(config.BaudRate)
	uart.UARTE_Type.SetCONFIG_HWFC(nrf.UARTE_CONFIG_HWFC_Disabled)

	// Set TX and RX pins
	if config.TX == 0 && config.RX == 0 {
		// Use default pins
		uart.setPins(UART_TX_PIN, UART_RX_PIN)
	} else {
		uart.setPins(config.TX, config.RX)
	}

	// apply_anomaly:
	// *(*uint32)(unsafe.Pointer(uintptr(0x00806020) + 4))
	// (*volatile.Register32)(unsafe.Pointer(uintptr(0x4006EC00))).Set(0x00009375)

	// Enable RX IRQ.
	uart.setupInterrupts()

	uart.Interrupt.Enable()

	uart.UARTE_Type.ENABLE.Set(nrf.UARTE_ENABLE_ENABLE_Enabled)

	uart.UARTE_Type.RXD.SetPTR(uint32(uintptr(unsafe.Pointer(&uart.rxbuffer))))
	uart.UARTE_Type.RXD.SetMAXCNT(1)
	uart.UARTE_Type.TASKS_STARTRX.Set(1)
}

func (uart *UARTE) setupInterrupts() {
	uart.UARTE_Type.EVENTS_ENDRX.Set(0)
	uart.UARTE_Type.EVENTS_ENDTX.Set(0)
	uart.UARTE_Type.EVENTS_ERROR.Set(0)
	uart.UARTE_Type.EVENTS_RXTO.Set(0)
	uart.UARTE_Type.EVENTS_TXSTOPPED.Set(0)

//	uart.UARTE_Type.INTENSET.Set(nrf.UARTE_INTENSET_ENDRX_Msk |
//		nrf.UARTE_INTENSET_ENDTX_Msk |
//		nrf.UARTE_INTENSET_ERROR_Msk |
//		nrf.UARTE_INTENSET_RXTO_Msk |
//		nrf.UARTE_INTENSET_TXSTOPPED_Msk)

	uart.UARTE_Type.INTENSET.Set(nrf.UARTE_INTENSET_ENDRX_Msk)

	uart.Interrupt.SetPriority(0xc0) // low priority
}

// SetBaudRate sets the communication speed for the UART.
func (uart *UARTE) SetBaudRate(br uint32) {
	// Magic: calculate 'baudrate' register from the input number.
	// Every value listed in the datasheet will be converted to the
	// correct register value, except for 192600. I suspect the value
	// listed in the nrf52 datasheet (0x0EBED000) is incorrectly rounded
	// and should be 0x0EBEE000, as the nrf51 datasheet lists the
	// nonrounded value 0x0EBEDFA4.
	// Some background:
	// https://devzone.nordicsemi.com/f/nordic-q-a/391/uart-baudrate-register-values/2046#2046
	rate := uint32((uint64(br/400)*uint64(400*0xffffffff/16000000) + 0x800) & 0xffffff000)

	uart.UARTE_Type.BAUDRATE.Set(rate)
}

func (uart *UARTE) handleInterrupt(interrupt.Interrupt) {
	if uart.UARTE_Type.EVENTS_ERROR.Get() != 0 {
		uart.UARTE_Type.EVENTS_ERROR.Set(0)
	} else if uart.UARTE_Type.EVENTS_ENDRX.Get() != 0 {
		uart.UARTE_Type.EVENTS_ENDRX.Set(0)
		uart.Receive(byte(uart.rxbuffer[0]))
		uart.UARTE_Type.TASKS_STARTRX.Set(1)
	} else if uart.UARTE_Type.EVENTS_RXTO.Get() != 0 {
		uart.UARTE_Type.EVENTS_RXTO.Set(0)
	} else if uart.UARTE_Type.EVENTS_ENDTX.Get() != 0 {
		uart.UARTE_Type.EVENTS_ENDTX.Set(0)
		uart.UARTE_Type.TASKS_STOPTX.Set(1)
	} else if uart.UARTE_Type.EVENTS_TXSTOPPED.Get() != 0 {
		uart.UARTE_Type.EVENTS_TXSTOPPED.Set(0)
	}
}

// Read from the RX buffer.
func (uart *UARTE) Read(data []byte) (n int, err error) {
	uart.UARTE_Type.INTENCLR.Set(nrf.UARTE_INTENSET_ENDRX_Msk)

	uart.UARTE_Type.EVENTS_ENDRX.Set(0)
	uart.UARTE_Type.EVENTS_RXTO.Set(0)

	var buf [64]uint8
	size := len(data)
	if size > len(buf) {
		size = len(buf)
	}

	uart.UARTE_Type.RXD.SetPTR(uint32(uintptr(unsafe.Pointer(&buf))))
	uart.UARTE_Type.RXD.SetMAXCNT(uint32(size))

	uart.UARTE_Type.TASKS_STARTRX.Set(1)

	for uart.UARTE_Type.EVENTS_ENDRX.Get() == 0 && uart.UARTE_Type.EVENTS_RXTO.Get() == 0 && uart.UARTE_Type.EVENTS_ERROR.Get() == 0 {
		gosched()
	}

	if uart.UARTE_Type.EVENTS_ERROR.Get() != 0 {
		return 0, nil
	}

	if uart.UARTE_Type.EVENTS_RXTO.Get() != 0 {
		return 0, nil
	}

	size = int(uart.UARTE_Type.RXD.GetAMOUNT())

	copy(data[:], buf[:size])

	uart.UARTE_Type.RXD.SetPTR(uint32(uintptr(unsafe.Pointer(&uart.rxbuffer))))
	uart.UARTE_Type.RXD.SetMAXCNT(1)
	uart.UARTE_Type.TASKS_STARTRX.Set(1)

	uart.UARTE_Type.INTENSET.Set(nrf.UARTE_INTENSET_ENDRX_Msk)

	return size, nil
}

func (uart *UARTE) WriteByte(c byte) error {
	_, err := uart.Write([]byte{c})
	return err
}

// Write data to the UART.
func (uart *UARTE) Write(data []byte) (n int, err error) {
	uart.UARTE_Type.EVENTS_ENDTX.Set(0)
	uart.UARTE_Type.EVENTS_TXSTOPPED.Set(0)

	var buf [64]uint8
	count := len(data)
	copy(buf[:], data[:count])

	uart.UARTE_Type.TXD.SetPTR(uint32(uintptr(unsafe.Pointer(&buf))))
	uart.UARTE_Type.TXD.SetMAXCNT(uint32(count))

	uart.UARTE_Type.TASKS_STARTTX.Set(1)

	for uart.UARTE_Type.EVENTS_ENDTX.Get() == 0 && uart.UARTE_Type.EVENTS_TXSTOPPED.Get() == 0 && uart.UARTE_Type.EVENTS_ERROR.Get() == 0 {
		gosched()
	}

	if uart.UARTE_Type.EVENTS_TXSTOPPED.Get() == 0 {
		return 0, nil
	} else {
		// Transmitter has to be stopped by triggering the STOPTX task to achieve
		// the lowest possible level of the UARTE power consumption.
		uart.UARTE_Type.TASKS_STOPTX.Set(1)
		for uart.UARTE_Type.EVENTS_TXSTOPPED.Get() == 0 {
			gosched()
		}
	}

	return count, nil
}

// ReadByte reads a single byte from the RX buffer.
// If there is no data in the buffer, returns an error.
func (uart *UARTE) ReadByte() (byte, error) {
	// check if RX buffer is empty
	buf, ok := uart.Buffer.Get()
	if !ok {
		return 0, errUARTBufferEmpty
	}
	return buf, nil
}

// Buffered returns the number of bytes currently stored in the RX buffer.
func (uart *UARTE) Buffered() int {
	return int(uart.Buffer.Used())
}

// Receive handles adding data to the UART's data buffer.
// Usually called by the IRQ handler for a machine.
func (uart *UARTE) Receive(data byte) {
	uart.Buffer.Put(data)
}
