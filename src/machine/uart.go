//go:build atmega || esp || nrf || sam || sifive || stm32 || k210 || nxp || rp2040

package machine

import "errors"

var errUARTBufferEmpty = errors.New("UART buffer empty")

// UARTParity is the parity setting to be used for UART communication.
type UARTParity uint8

type UARTCommon struct {
	Buffer *RingBuffer
	RXChan chan bool
}

const (
	// ParityNone means to not use any parity checking. This is
	// the most common setting.
	ParityNone UARTParity = iota

	// ParityEven means to expect that the total number of 1 bits sent
	// should be an even number.
	ParityEven

	// ParityOdd means to expect that the total number of 1 bits sent
	// should be an odd number.
	ParityOdd
)

// To implement the UART interface for a board, you must declare a concrete type as follows:
//
// 		type UART struct {
// 			UARTCommon
// 		}
//
// You can also add additional members to this struct depending on your
// implementation. UARTCommon is required. When you are declaring your UARTs for
// your board, make sure that you also initialize UARTCommon using the
// NewUARTCommon() function:
//
//		UART{UARTCommon: NewUARTCommon()}
//

// Read from the RX buffer.
func (uart *UART) Read(data []byte) (n int, err error) {
	// If RX buffer is empty, block until data is available.
	size := uart.Buffered()
	if size == 0 {
		<-uart.RXChan
		size = uart.Buffered()
	}

	// Make sure we do not read more from buffer than the data slice can hold.
	if len(data) < size {
		size = len(data)
	}

	// only read number of bytes used from buffer
	for i := 0; i < size; i++ {
		v, _ := uart.ReadByte()
		data[i] = v
	}

	return size, nil
}

// WriteByte writes a byte of data over the UART's Tx.
// This function blocks until the data is finished being sent.
func (uart *UART) WriteByte(c byte) error {
	err := uart.writeByte(c)
	if err != nil {
		return err
	}
	uart.flush() // flush() blocks until all data has been transmitted.
	return nil
}

// Write data over the UART's Tx.
// This function blocks until the data is finished being sent.
func (uart *UART) Write(data []byte) (n int, err error) {
	for i, v := range data {
		err = uart.writeByte(v)
		if err != nil {
			return i, err
		}
	}
	uart.flush() // flush() blocks until all data has been transmitted.
	return len(data), nil
}

// ReadByte reads a single byte from the RX buffer.
// If there is no data in the buffer, returns an error.
func (uart *UART) ReadByte() (byte, error) {
	// check if RX buffer is empty
	buf, ok := uart.Buffer.Get()
	if !ok {
		return 0, errUARTBufferEmpty
	}
	return buf, nil
}

// Buffered returns the number of bytes currently stored in the RX buffer.
func (uart *UART) Buffered() int {
	return int(uart.Buffer.Used())
}

// Receive handles adding data to the UART's data buffer.
// Usually called by the IRQ handler for a machine.
func (uart *UART) Receive(data byte) {
	uart.Buffer.Put(data)

	select {
	case uart.RXChan <- true:
	default:
	}
}

func NewUARTCommon() UARTCommon {
	return UARTCommon{
		Buffer: NewRingBuffer(),
		RXChan: make(chan bool),
	}
}
