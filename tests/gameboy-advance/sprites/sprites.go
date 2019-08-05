package main

import (
	"image/color"
	"machine"

	"github.com/conejoninja/tinydraw"
	"github.com/conejoninja/tinyfont"
)

func rgb(r, g, b uint8) uint16 {
	return uint16(r>>3) | uint16(g>>3)<<5 | uint16(b>>3)<<10
}

// fillPalette sets up a palette with:
//   0   transparent
//   1   white
//   2   red
//   3   green
//   4   blue
//   5   black
func fillPalette(p *machine.PaletteBank) {
	p[1].Set(rgb(255, 255, 255))
	p[2].Set(rgb(255, 0, 0))
	p[3].Set(rgb(0, 255, 0))
	p[4].Set(rgb(0, 0, 255))
	p[5].Set(rgb(0, 0, 0))
}

func tileIndices(i1, i2, i3, i4 uint8) uint32 {
	return uint32(i1) | uint32(i2)<<8 | uint32(i3)<<16 | uint32(i4)<<24
}

func setInTile(t *machine.Tile4, x, y int, colorIndex int) {
	// This is super inefficient.  Generally we shouldn't have to mutate tiles.
	t[y&0x7].ClearBits(0xF << uint(4*(x&0x7)))
	t[y&0x7].SetBits(uint32(colorIndex&0xF) << uint(4*(x&0x7)))
}

func fillFont(block *machine.TileBlock4, offset *int, font *tinyfont.Font, color int) (start int) {
	// This is fun, but inefficient. It would be way better to have the tiles
	// already prepared in read-only memory.
	start = *offset
	for i, glyph := range font.Glyphs {
		// Only include basic glyphs.
		if i+int(font.First) > int(font.Last) {
			continue
		}

		// Allocate a tile for the character.
		tile := &block[*offset]
		*offset++

		// Manually blit the bitmaps onto the tile.  *shudder*
		bitmap := font.Bitmaps[glyph.BitmapIndex:]
		for y := 0; y < int(glyph.Height); y++ {
			for x := 0; x < int(glyph.Width); x++ {
				if bitmap[y]&(0x80>>uint(x)) == 0 {
					continue
				}
				setInTile(tile, x+int(glyph.XOffset), y+int(font.YAdvance)+int(glyph.YOffset), color)
			}
		}
	}
	return start
}

func main() {
	// At startup, FORCED_BLANK is set, which hides the entire screen.
	// During this period, we are allowed to spend as much time as we want
	// mucking about in OAM.  Once FORCED_BLANK is disabled, we can only acess
	// it during VBLANK (and HBLANK if DISPCNT_HBLANK_INTERVAL_FREE is set).

	// nextTile is updated as we create tiles
	nextTile := 1 // don't use tile 0, since that's the default

	palBank := 1
	fillPalette(&machine.SpritePalette.Bank[palBank])

	font := &tinyfont.TomThumb

	// Create font tiles that we can use for creating sprites.
	fontColor := 5 // black, see filPalette
	fontStart := fillFont(&machine.Tiles.Block4[machine.TILE_BLOCK_S1], &nextTile, font, fontColor)

	// Create a helper for creating and positioning font-based sprites.
	var ( // createSprite state variables
		spriteIndex, spritePriority = 0, 0
		xStart, yStart              = 10, 10
		xOffset, yOffset, xMax      = xStart, yStart, xStart
	)
	createSprite := func(ch byte) {
		if ch == '\n' || xOffset > 200 {
			xOffset = xStart
			yOffset += int(font.YAdvance)
		}
		if ch < font.First || ch > font.Last {
			return
		}
		glyph := font.Glyphs[int(ch-font.First)]

		sprite := &machine.Sprites.Sprite[spriteIndex]
		spriteIndex++

		tileOffset := fontStart + int(ch-font.First)
		sprite.Setup4(
			spritePriority, machine.SPRITE_4BPP_S1_OFFSET+tileOffset, palBank,
			machine.SPRITE_SQUARE,
			machine.SPRITE_SIZE_S,
		)
		sprite.SetPos(xOffset, yOffset)
		xOffset += int(glyph.XAdvance)
		if xOffset > xMax {
			xMax = xOffset
		}
	}

	// Create a bunch of sprites based on the letters.
	message := "If you an read this, sprites are rendering correctly!\nHere is a second line of text for your enjoyment."
	for _, r := range message {
		createSprite(byte(r))
	}

	// Enable the sprites (and mark them as 1-dim, though it's irrelevant)
	machine.Sprites.Enable1D()

	// Draw a pretty box around the sprites -- they'll be drawn over it.
	display := machine.Display.Mode3
	const padT, padR, padB, padL = 1, 1, 2, 2
	tinydraw.FilledRectangle(display,
		int16(xStart)-padL, int16(yStart)-padT,
		int16(xMax-xStart)+padL+padR, int16(yOffset-yStart)+int16(font.YAdvance)+padB+padT,
		color.RGBA{G: 255})
	display.Configure() // clears FORCED_BLANK, thus displaying everything

	// Wait forever.
	for {
	}
}
