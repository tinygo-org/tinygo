// +build avr nrf stm32

package machine

import "errors"

type UARTConfig struct {
	BaudRate uint32
	TXPin    uint8
	RXPin    uint8
}

type UART struct {
}

// Read from the RX buffer.
func (uart UART) Read(data []byte) (n int, err error) {
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
func (uart UART) Write(data []byte) (n int, err error) {
	for _, v := range data {
		uart.WriteByte(v)
	}
	return len(data), nil
}

// ReadByte reads a single byte from the RX buffer.
// If there is no data in the buffer, returns an error.
func (uart UART) ReadByte() (byte, error) {
	// check if RX buffer is empty
	if uart.Buffered() == 0 {
		return 0, errors.New("Buffer empty")
	}

	return bufferGet(), nil
}

// Buffered returns the number of bytes currently stored in the RX buffer.
func (uart UART) Buffered() int {
	return int(bufferUsed())
}

const bufferSize = 64

// Minimal ring buffer implementation inspired by post at
// https://www.embeddedrelated.com/showthread/comp.arch.embedded/77084-1.php

//go:volatile
type volatileByte byte

var rxbuffer [bufferSize]volatileByte
var head volatileByte
var tail volatileByte

func bufferUsed() uint8 { return uint8(head - tail) }
func bufferPut(val byte) {
	if bufferUsed() != bufferSize {
		head++
		rxbuffer[head%bufferSize] = volatileByte(val)
	}
}
func bufferGet() byte {
	if bufferUsed() != 0 {
		tail++
		return byte(rxbuffer[tail%bufferSize])
	}
	return 0
}
