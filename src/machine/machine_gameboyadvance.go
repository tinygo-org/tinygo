//go:build gameboyadvance

package machine

import (
	"image/color"
	"runtime/interrupt"
	"runtime/volatile"
	"unsafe"
)

// Not sure what name to pick here. Not using ARM7TDMI because that's the CPU
// name, not the device name.
const deviceName = "GBA"

// Interrupt numbers as used on the GameBoy Advance. Register them with
// runtime/interrupt.New.
const (
	IRQ_VBLANK  = interrupt.IRQ_VBLANK
	IRQ_HBLANK  = interrupt.IRQ_HBLANK
	IRQ_VCOUNT  = interrupt.IRQ_VCOUNT
	IRQ_TIMER0  = interrupt.IRQ_TIMER0
	IRQ_TIMER1  = interrupt.IRQ_TIMER1
	IRQ_TIMER2  = interrupt.IRQ_TIMER2
	IRQ_TIMER3  = interrupt.IRQ_TIMER3
	IRQ_COM     = interrupt.IRQ_COM
	IRQ_DMA0    = interrupt.IRQ_DMA0
	IRQ_DMA1    = interrupt.IRQ_DMA1
	IRQ_DMA2    = interrupt.IRQ_DMA2
	IRQ_DMA3    = interrupt.IRQ_DMA3
	IRQ_KEYPAD  = interrupt.IRQ_KEYPAD
	IRQ_GAMEPAK = interrupt.IRQ_GAMEPAK
)

// Make it easier to directly write to I/O RAM.
var ioram = (*[0x400]volatile.Register8)(unsafe.Pointer(uintptr(0x04000000)))

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
