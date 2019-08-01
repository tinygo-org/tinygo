// +build 386,baremetal

package machine

import (
	"unsafe"
)

// PinSet is only provided for compatibility.
type PinMode uint8

// Set does nothing. The x86 platform usually doesn't expose any GPIO pins.
func (p Pin) Set(value bool) {
	// do nothing
}

//export outb
func outb(port uint16, val uint8)

var COM1 = UART{port: 0x3f8}

type UART struct {
	port uint16
}

// Configure allows the given COM port to be used for serial output.
func (uart UART) Configure() {
	// https://wiki.osdev.org/Serial_Ports#Initialization
	outb(uart.port+1, 0x00) // Disable all interrupts
	outb(uart.port+3, 0x80) // Enable DLAB (set baud rate divisor)
	outb(uart.port+0, 0x03) // Set divisor to 3 (lo byte) 38400 baud
	outb(uart.port+1, 0x00) //                  (hi byte)
	outb(uart.port+3, 0x03) // 8 bits, no parity, one stop bit
	outb(uart.port+2, 0xC7) // Enable FIFO, clear them, with 14-byte threshold
	outb(uart.port+4, 0x0B) // IRQs enabled, RTS/DSR set
}

// WriteByte sends a single byte to this COM port.
func (uart UART) WriteByte(c byte) error {
	// TODO: wait until the previous char has been sent.
	// This does not seem to be necessary in QEMU.
	outb(uart.port, c)
	return nil
}

const (
	ConsoleWidth  = 80
	ConsoleHeight = 25
)

// VGA colors, as used by the console interface.
const (
	VGAColorBlack = iota
	VGAColorBlue
	VGAColorGreen
	VGAColorCyan
	VGAColorRed
	VGAColorMagenta
	VGAColorBrown
	VGAColorLightGrey
	VGAColorDarkGrey
	VGAColorLightBlue
	VGAColorLightGreen
	VGAColorLightCyan
	VGAColorLightRed
	VGAColorLightMagenta
	VGAColorLightBrown
	VGAColorWhite
)

var consoleRawBuffer = (*[25][80]uint16)(unsafe.Pointer(uintptr(0xB8000)))

// Console0 is the standard console for this system.
var Console0 = Console{}

// Console represents the VGA console on this system.
type Console struct {
	x     int
	y     int
	color uint8
}

// Configure resets the console to the default state.
func (c *Console) Configure() {
	c.x = 0
	c.y = 0
	c.SetColor(VGAColorWhite, VGAColorBlack)
	c.Clear()
}

// Clear erases everything from the console and puts the cursor in the top left.
func (c *Console) Clear() {
	for x := 0; x < ConsoleWidth; x++ {
		for y := 0; y < ConsoleHeight; y++ {
			c.SetEntry(x, y, ' ', makeVGAColor(VGAColorWhite, VGAColorBlack))
		}
	}
}

// Scroll moves the screen up by the given number of lines.
func (c *Console) Scroll(lines int) {
	// Move the screen up by this amount.
	for y := lines; y < ConsoleHeight; y++ {
		for x := 0; x < ConsoleWidth; x++ {
			consoleRawBuffer[y-lines][x] = consoleRawBuffer[y][x]
		}
	}

	// Clear the last lines.
	for y := ConsoleHeight - lines; y < ConsoleHeight; y++ {
		for x := 0; x < ConsoleWidth; x++ {
			c.SetEntry(x, y, ' ', c.color)
		}
	}
}

// SetByte sets a byte at the given location, without affecting the cursor
// position.
func (c *Console) SetEntry(x, y int, ch byte, color uint8) {
	consoleRawBuffer[y][x] = uint16(ch) | uint16(c.color)<<8
}

// SetColor updates the current cursor color.
func (c *Console) SetColor(foreground, background uint8) {
	c.color = makeVGAColor(foreground, background)
}

// WriteByte writes one char to the console, wrapping around if necessary.
func (c *Console) WriteByte(ch byte) error {
	if ch == '\r' {
		return nil
	}
	if ch == '\n' {
		c.x = 0
		c.y++
		return nil
	}
	if c.x == ConsoleWidth {
		c.x = 0
		c.y++
	}
	if c.y >= ConsoleHeight {
		c.Scroll(c.y - ConsoleHeight + 1)
		c.y = ConsoleHeight - 1
	}
	c.SetEntry(c.x, c.y, ch, c.color)
	c.x++
	return nil
}

// makeVGAColor converts a background/foreground color into a single VGA color
// to be used directly in the VGA framebuffer.
func makeVGAColor(foreground, background uint8) uint8 {
	return foreground | background<<4
}

// Halt shuts down the system immediately.
//export halt
func Halt()
