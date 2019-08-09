package main

import (
	"image/color"
	"machine"
	"runtime/volatile"

	"github.com/conejoninja/tinydraw"
	"github.com/conejoninja/tinyfont"
)

var (
	black = color.RGBA{}
	red   = color.RGBA{R: 255}
	green = color.RGBA{G: 255}
)

const (
	BarHeight = 11
	BarWidth  = 240

	FontHeight  = 5
	TextXOffset = (BarHeight - FontHeight) / 2
	TextYOffset = BarHeight - TextXOffset
)

var mark [machine.InterruptCount]volatile.Register8

func handler(intr machine.Interrupt) {
	mark[intr].Set(1)
	machine.Interrupts.Disable(intr)
}

func main() {
	display := machine.Display.Mode3 // machine.Display.Mode3
	display.Configure()

	interrupts := []struct {
		reg       *volatile.Register8
		bad, good []byte
	}{
		{&mark[machine.INT_VBLANK], []byte("no vblank"), []byte("got vblank")},
		{&mark[machine.INT_HBLANK], []byte("no hblank"), []byte("got hblank")},
		{&mark[machine.INT_VCOUNTER], []byte("no vcounter"), []byte("got vcounter")},
		{&mark[machine.INT_TIMER0], []byte("waiting for timer0..."), []byte("got timer0")},
		{&mark[machine.INT_TIMER1], []byte("waiting for timer1..."), []byte("got timer1")},
		{&mark[machine.INT_TIMER2], []byte("waiting for timer2..."), []byte("got timer2")},
		{&mark[machine.INT_TIMER3], []byte("waiting for timer3..."), []byte("got timer3")},
		{&mark[machine.INT_KEYPAD], []byte("... press a key ..."), []byte("got keypad")},
	}

	for i, info := range interrupts {
		// Write a message on a red background
		tinydraw.FilledRectangle(
			display,
			1, int16(i)*BarHeight+1,
			BarWidth-2, BarHeight-1,
			red,
		)
		tinyfont.WriteLine(
			display, &tinyfont.TomThumb,
			TextXOffset+1, int16(i)*BarHeight+TextYOffset,
			info.bad, black,
		)
	}

	for i := range machine.IO.Timer {
		machine.IO.Timer[i].Stop()
		machine.IO.Timer[i].Counter.Set(0)
		machine.IO.Timer[i].Control.Set(uint16(i)) // 0..3 are the prescale constants
		machine.IO.Timer[i].Start()
	}

	machine.IO.Keypad.WakeOn(machine.KEY_ANY)

	machine.Interrupts.Enable(
		handler,
		machine.INT_VBLANK,
		machine.INT_HBLANK,
		machine.INT_VCOUNTER,
		machine.INT_TIMER0,
		machine.INT_TIMER1,
		machine.INT_TIMER2,
		machine.INT_TIMER3,
		machine.INT_KEYPAD,
	)

	for {
		for i, info := range interrupts {
			if info.reg.Get() == 0 {
				continue
			}
			info.reg.Set(0)

			// Write a message on a green background
			tinydraw.FilledRectangle(
				display,
				1, int16(i)*BarHeight+1,
				BarWidth-2, BarHeight-1,
				green,
			)
			tinyfont.WriteLine(
				display, &tinyfont.TomThumb,
				TextXOffset+1, int16(i)*BarHeight+TextYOffset,
				info.good, black,
			)
		}
	}
}
