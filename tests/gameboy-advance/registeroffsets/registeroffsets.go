package main

import (
	"image/color"
	"machine"
	"strconv"
	"unsafe"

	"github.com/conejoninja/tinydraw"
	"github.com/conejoninja/tinyfont"
)

var (
	black = color.RGBA{}
	red   = color.RGBA{R: 255}
	green = color.RGBA{G: 255}
)

func test() ([]byte, color.RGBA) {
	for _, assert := range []struct {
		what string
		got  unsafe.Pointer
		want uintptr
	}{
		{"IO.LCD", unsafe.Pointer(&machine.IO.LCD), 0x04000000},
		{"IO.Sound", unsafe.Pointer(&machine.IO.Sound), 0x04000060},
		{"IO.Sound.WAVE_RAM", unsafe.Pointer(&machine.IO.Sound.WAVE_RAM), 0x04000090},
		{"IO.Sound.FIFO_A", unsafe.Pointer(&machine.IO.Sound.FIFO_A), 0x040000A0},
		{"IO.Timer", unsafe.Pointer(&machine.IO.Timer), 0x4000100},
		{"IO.Timer[3].Control", unsafe.Pointer(&machine.IO.Timer[3].Control), 0x400010E},
		{"IO.Keypad", unsafe.Pointer(&machine.IO.Keypad), 0x4000130},
		{"IO.Int", unsafe.Pointer(&machine.IO.Int), 0x4000200},
		{"IO.Int.Request", unsafe.Pointer(&machine.IO.Int.Request), 0x4000200},
		{"IO.Int.Ack", unsafe.Pointer(&machine.IO.Int.Ack), 0x4000202},
		{"IO.Int.Enable", unsafe.Pointer(&machine.IO.Int.Enable), 0x4000208},
		{"Tiles.Block4[S0][0]", unsafe.Pointer(&machine.Tiles.Block4[machine.TILE_BLOCK_S0][0]), 0x06010000},
		{"Tiles.Block4[S1][0]", unsafe.Pointer(&machine.Tiles.Block4[machine.TILE_BLOCK_S1][0]), 0x06014000},
		{"Tiles.Block8[S0][0]", unsafe.Pointer(&machine.Tiles.Block8[machine.TILE_BLOCK_S0][0]), 0x06010000},
		{"Tiles.Block8[S1][0]", unsafe.Pointer(&machine.Tiles.Block8[machine.TILE_BLOCK_S1][0]), 0x06014000},
		{"Sprites.Sprite[0].Y", unsafe.Pointer(&machine.Sprites.Sprite[0].Y), 0x07000000},
		{"Sprites.Sprite[127].Y", unsafe.Pointer(&machine.Sprites.Sprite[127].Y), 0x070003F8},
		{"Backgrounds.ScreenBlocks[0]", unsafe.Pointer(&machine.Backgrounds.ScreenBlocks[0]), 0x06000000},
		{"Backgrounds.Control[0]", unsafe.Pointer(&machine.Backgrounds.Control[0]), 0x04000008},
		{"Backgrounds.Offset[0].X", unsafe.Pointer(&machine.Backgrounds.Offset[0].X), 0x04000010},
		{"Backgrounds.Offset[0].Y", unsafe.Pointer(&machine.Backgrounds.Offset[0].Y), 0x04000012},
	} {
		if uintptr(assert.got) == assert.want {
			continue
		}

		var message []byte
		message = append(message, assert.what...)
		message = append(message, " = 0x"...)
		message = strconv.AppendUint(message, uint64(uintptr(assert.got)), 16)
		message = append(message, ", want 0x"...)
		message = strconv.AppendUint(message, uint64(uintptr(assert.want)), 16)

		return message, red
	}

	return []byte("Memory Tests: PASS"), green
}

func main() {
	display := machine.Display.Mode3 // machine.Display.Mode3
	display.Configure()

	str, color := test()

	const (
		BarHeight = 11
		BarWidth  = 240

		FontHeight  = 5
		TextXOffset = (BarHeight - FontHeight) / 2
		TextYOffset = BarHeight - TextXOffset
	)

	// Write a message in black-on-color
	tinydraw.FilledRectangle(display, 1, 1, BarWidth-2, BarHeight-2, color)
	tinyfont.WriteLine(display, &tinyfont.TomThumb, TextXOffset+1, TextYOffset, str, black)

	for {
	}
}
