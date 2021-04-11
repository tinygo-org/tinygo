// +build gameboyadvance

package machine

import (
	"image/color"
	"runtime/volatile"
	"unsafe"
)

// Interrupt numbers as used on the GameBoy Advance. Register them with
// runtime/interrupt.New.
const (
	IRQ_VBLANK  = 0
	IRQ_HBLANK  = 1
	IRQ_VCOUNT  = 2
	IRQ_TIMER0  = 3
	IRQ_TIMER1  = 4
	IRQ_TIMER2  = 5
	IRQ_TIMER3  = 6
	IRQ_COM     = 7
	IRQ_DMA0    = 8
	IRQ_DMA1    = 9
	IRQ_DMA2    = 10
	IRQ_DMA3    = 11
	IRQ_KEYPAD  = 12
	IRQ_GAMEPAK = 13
)

// Make it easier to directly write to I/O RAM.
var ioram = (*[0x400]volatile.Register8)(unsafe.Pointer(uintptr(0x04000000)))

type PinMode uint8

// Set has not been implemented.
func (p Pin) Set(value bool) {
	// do nothing
}

var Display = FramebufDisplay{(*[160][240]volatile.Register16)(unsafe.Pointer(uintptr(0x06000000)))}

type FramebufDisplay struct {
	port *[160][240]volatile.Register16
}

func (d FramebufDisplay) Configure() {
	// Write into the I/O registers, setting video display parameters.
	ioram[0].Set(0x03) // Use video mode 3 (in BG2, a 16bpp bitmap in VRAM)
	ioram[1].Set(0x04) // Enable BG2 (BG0 = 1, BG1 = 2, BG2 = 4, ...)
}

func (d FramebufDisplay) Size() (x, y int16) {
	return 240, 160
}

func (d FramebufDisplay) SetPixel(x, y int16, c color.RGBA) {
	d.port[y][x].Set((uint16(c.R) >> 3) | ((uint16(c.G) >> 3) << 5) | ((uint16(c.B) >> 3) << 10))
}

func (d FramebufDisplay) Display() error {
	// Nothing to do here.
	return nil
}
