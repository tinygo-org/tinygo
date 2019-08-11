package main

import (
	"machine"
)

type BackgroundManager4 struct {
	nextPal int
	nextSB  int // screenblock index, also used to compute CB
}

func (m *BackgroundManager4) AddPalette(data []uint16) (pal int) {
	pal = m.nextPal
	machine.BackgroundPalette.Bank[pal].Init(data)
	m.nextPal++
	return pal
}

func (m *BackgroundManager4) AddTiles(data []uint32) (cbb int) {
	const TilesPerCB = len(machine.TileBlock4{})
	const SBperCB = 8

	// Tiles need to start at a CharBlock, so skip ahead
	for m.nextSB%SBperCB != 0 {
		m.nextSB++
	}

	cbb = m.nextSB / SBperCB
	for len(data) > 8 { // 8 words are requird for an 8x8 4bpp tile
		cb := m.nextSB / SBperCB
		m.nextSB += SBperCB

		for i := 0; i < TilesPerCB; i++ {
			if len(data) < 8 {
				return cbb // 8 words are requird for an 8x8 4bpp tile
			}

			curr := data[:8]
			data = data[8:]
			machine.Tiles.Block4[cb][i].Init(curr)
		}
	}
	return cbb
}

func (m *BackgroundManager4) AddTilemap(pal int, tilemap []uint16) (sbb int) {
	sbb = m.nextSB
	for len(tilemap) > 0 {
		reg := &machine.Backgrounds.ScreenBlocks[m.nextSB].Regular
		m.nextSB++
		for r := range reg {
			for c := range reg[r] {
				if len(tilemap) <= 0 {
					return sbb
				}
				reg[r][c].Set4(pal, int(tilemap[0]))
				tilemap = tilemap[1:]
			}
		}
	}
	return sbb
}

func main() {
	var bgm BackgroundManager4

	pal := bgm.AddPalette(Palette_tinygo[:])
	cbb := bgm.AddTiles(Tileset_tinygo_0[:])
	sbb := bgm.AddTilemap(pal, Tilemap_tinygo_0[:])

	machine.Backgrounds.Control[0].Configure4(0, cbb, sbb, 0)
	machine.Backgrounds.Offset[0].X.Set(8)
	machine.Backgrounds.Offset[0].Y.Set(8 * 6)
	machine.Display.Mode0.Configure()

	for {
	}
}
