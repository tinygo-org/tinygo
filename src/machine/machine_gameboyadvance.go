//go:build gameboyadvance

package machine

import (
	"device/gba"

	"image/color"
	"runtime/volatile"
	"unsafe"
)

// Not sure what name to pick here. Not using ARM7TDMI because that's the CPU
// name, not the device name.
const deviceName = "GBA"

// Interrupt numbers as used on the GameBoy Advance. Register them with
// runtime/interrupt.New.
const (
	IRQ_VBLANK  = gba.IRQ_VBLANK
	IRQ_HBLANK  = gba.IRQ_HBLANK
	IRQ_VCOUNT  = gba.IRQ_VCOUNT
	IRQ_TIMER0  = gba.IRQ_TIMER0
	IRQ_TIMER1  = gba.IRQ_TIMER1
	IRQ_TIMER2  = gba.IRQ_TIMER2
	IRQ_TIMER3  = gba.IRQ_TIMER3
	IRQ_COM     = gba.IRQ_COM
	IRQ_DMA0    = gba.IRQ_DMA0
	IRQ_DMA1    = gba.IRQ_DMA1
	IRQ_DMA2    = gba.IRQ_DMA2
	IRQ_DMA3    = gba.IRQ_DMA3
	IRQ_KEYPAD  = gba.IRQ_KEYPAD
	IRQ_GAMEPAK = gba.IRQ_GAMEPAK
)

// Set has not been implemented.
func (p Pin) Set(value bool) {
	// do nothing
}

var Display = DisplayMode3{(*[160][240]volatile.Register16)(unsafe.Pointer(uintptr(gba.MEM_VRAM)))}

type DisplayMode3 struct {
	port *[160][240]volatile.Register16
}

func (d *DisplayMode3) Configure() {
	// Use video mode 3 (in BG2, a 16bpp bitmap in VRAM) and Enable BG2
	gba.DISP.DISPCNT.Set(gba.DISPCNT_BGMODE_3<<gba.DISPCNT_BGMODE_Pos |
		gba.DISPCNT_SCREENDISPLAY_BG2_ENABLE<<gba.DISPCNT_SCREENDISPLAY_BG2_Pos)
}

func (d *DisplayMode3) Size() (x, y int16) {
	return 240, 160
}

func (d *DisplayMode3) SetPixel(x, y int16, c color.RGBA) {
	d.port[y][x].Set((uint16(c.R) >> 3) | ((uint16(c.G) >> 3) << 5) | ((uint16(c.B) >> 3) << 10))
}

func (d *DisplayMode3) Display() error {
	// Nothing to do here.
	return nil
}
