package rpi3

import "unsafe"

// Psf is the raw header that is in the font file, minus the last field.
type Psf struct {
	magic         uint32
	version       uint32
	headersize    uint32
	flags         uint32
	numglyph      uint32
	bytesperglyph uint32
	height        uint32
	width         uint32
	//there is nothing by a sequence of bytes after this header...
}

// PSFFont is our wrapper around the Psf structure that knows how to display a string
// on the screen and knows how to do our GPU-assisted scrolling.
type PSFFont struct {
	raw          *Psf
	data         *uint8
	info         *FBInfo
	lines        uint16
	yLine        uint16
	widthInChars uint16
}

// DisplayString prints str at (x,y) using its own font data.  You probably only
// want to use this to wrte at a particular location on the screen.  For simply
// outputing text, see ConsolePrint.
func (p *PSFFont) DisplayString(x uint16, y uint16, str string) {
	for n := range str {
		s := str[n]
		glyphNum := uint32(s)
		if uint32(s) > p.raw.numglyph {
			glyphNum = 0
		}
		// NOTE NOTE NOTE: the fact that this type is *uint8 causes the compiler to
		// be careful about what it emits--and it uses an instruction that does not
		// require aligned access. If you use *int or similar, you will get an
		// abort due to alignment exception
		glyph := (*uint8)((unsafe.Pointer)(uintptr(unsafe.Pointer(p.data)) +
			uintptr(glyphNum*p.raw.bytesperglyph)))

		//glyph is now a pointer to the bytes we need for this character to be drawn
		offs := p.getOffset(int(x), int(y))
		bytesperline := (p.raw.width + 7) / 8
		if s == '\r' {
			x = 0
		} else if s == '\n' {
			x = 0
			y++
		} else {
			//print the character
			for j := 0; j < int(p.raw.height); j++ {
				// display one row
				line := offs
				mask := 1 << (p.raw.width - 1)
				glyphvalue := *glyph
				for i := 0; i < int(p.raw.width); i++ {
					// if bit set, we use yellow color, otherwise black
					color := uint32(0)
					if glyphvalue&uint8(mask) != 0 {
						color = 0xFFFF00
					}
					*((*uint32)(unsafe.Pointer(uintptr(p.info.Ptr) + uintptr(line)))) = color
					mask >>= 1
					line += 4
				}
				// adjust to next line
				glyph = (*uint8)(unsafe.Pointer(uintptr(unsafe.Pointer(glyph)) + uintptr(bytesperline)))
				offs += p.info.Pitch
			}
			x++
		}
	}
}

// getOffset is a helper to compute the offset into the framebuffer given a location
// and the pitch.
func (p *PSFFont) getOffset(x int, y int) int {
	return (y * int(p.raw.height) * p.info.Pitch) + (x * (int(p.raw.width) + 1) * 4)
}

// NewPSFFontViaLinker creates a proper PSFFont structure from the symbol
// provided. The symbol needs to be defined in the binary that is loaded.
func NewPSFFontViaLinker(ptr unsafe.Pointer, info *FBInfo) *PSFFont {
	var tmp Psf
	data := ((*uint8)((unsafe.Pointer)(uintptr(ptr) + unsafe.Sizeof(tmp))))
	result := &PSFFont{
		raw:  ((*Psf)(ptr)),
		data: data,
		info: info,
	}
	result.lines = uint16(info.Height) / uint16(result.raw.height)
	// we set to result.lines because we know screen 0 comes first
	result.yLine = result.lines // varies from result.lines -> 2 * result.lines (48 to 96 on screen 0, 144->192 on screen 1)
	result.widthInChars = uint16(info.Width) / uint16(result.raw.width)
	return result
}

// ConsolePrint scrolls the screen up one line and prints the given string on the display.
// This assumes that you have set your max_framebuffer_height=xxx in your config.txt
// on your pi to 4*lines
func (p *PSFFont) ConsolePrint(str string) {

	//truncate if too long
	if uint16(len(str)) > p.widthInChars {
		str = str[:p.widthInChars]
	}
	if p.yLine < p.lines || p.yLine >= 4*p.lines {
		print("scrolling contstraint violated! yLine is ", p.yLine, "\n")
		Abort()
	}
	if p.yLine >= (2*p.lines) && p.yLine < (3*p.lines) {
		print("scrolling contstraint violated (prohibited middle)! yLine is ", p.yLine, "\n")
		Abort()
	}

	// step 1: move the window down to expose the space where we are about to write
	var distance uint16
	twos := 2 * p.lines

	if p.yLine < twos { // isTop?
		distance = (p.yLine % p.lines) + 1
	} else {
		distance = ((p.yLine % p.lines) + 1) + twos //add on the second screen distance
	}
	yoffset := distance * uint16(p.raw.height)
	SetYScroll(yoffset)

	// display string works in character positions not pixels...
	// step2: we need to show this new content AND draw a copy in the other area
	p.DisplayString(0, p.yLine, str) //current
	if p.yLine < twos {              //isTop?
		p.DisplayString(0, p.yLine+p.lines, str) //other, bottom
	} else {
		p.DisplayString(0, p.yLine-(3*p.lines), str) //other, top
	}

	//step3: check for time to switch to other buffer
	if p.yLine%twos == twos-1 {
		//we can switch the virtual window to be on the other screen because the
		//content of the area switched to is the same as ours
		if p.yLine < twos { //isTop?
			p.yLine = 3 * p.lines
			// stick to TOP of lower screen so the next line will make sense at the bottom
			distance = twos
		} else {
			p.yLine = p.lines
			// go to very top of virtual buffer
			distance = 0
		}
		yoffset = distance * uint16(p.raw.height)
		SetYScroll(yoffset)
	} else {
		// the simple case is that we just increment and keep moving
		p.yLine++
	}
}
