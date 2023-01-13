//go:build (atmega || esp || nrf || sam || sifive || stm32 || k210 || nxp || rp2040) && !nrf9160

package machine

// To implement the UART interface for a board, you must declare a concrete type as follows:
//
// 		type UART struct {
// 			Buffer *RingBuffer
// 		}
//
// You can also add additional members to this struct depending on your implementation,
// but the *RingBuffer is required.
// When you are declaring your UARTs for your board, make sure that you also declare the
// RingBuffer using the NewRingBuffer() function when you declare your UART:
//
//		UART{Buffer: NewRingBuffer()}
//

// Read from the RX buffer.
func (uart *UART) Read(data []byte) (n int, err error) {
	// check if RX buffer is empty
	size := uart.Buffered()
	if size == 0 {
		return 0, nil
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

// Write data to the UART.
func (uart *UART) Write(data []byte) (n int, err error) {
	for _, v := range data {
		uart.WriteByte(v)
	}
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
}
