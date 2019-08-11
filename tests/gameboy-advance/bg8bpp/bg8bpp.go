package main

import (
	"machine"
)

type BackgroundManager8 struct {
	nextPal int
	nextSB  int // screenblock index, also used to compute CB
}

func (m *BackgroundManager8) AddPalette(data []uint16) {
	m.nextPal += machine.BackgroundPalette.Full.Init(data, m.nextPal)
}

func (m *BackgroundManager8) AddTiles(data []uint32) (cbb int) {
	const (
		SBperCB    = 8
		TilesPerCB = len(machine.TileBlock8{})
		TilesPerSB = TilesPerCB / SBperCB
	)

	// Tiles need to start at a CharBlock, so skip ahead
	for m.nextSB%SBperCB != 0 {
		m.nextSB++
	}

	cbb = m.nextSB / SBperCB

	tiles := len(data) / 16
	m.nextSB += (tiles + TilesPerSB - 1) / TilesPerSB // round up

	for tile := 0; tile < tiles; tile++ {
		cb := cbb + tile/TilesPerCB
		i := tile % TilesPerCB
		machine.Tiles.Block8[cb][i].Init(data[16*tile:][:16])
	}
	return cbb
}

func (m *BackgroundManager8) AddTilemap(tilemap []uint16, tx, ty int) (sbb int) {
	const (
		SBperCB    = 8
		TilesPerCB = len(machine.TileBlock8{})
		TilesPerSB = TilesPerCB / SBperCB
	)

	// Store the screenblock we're starting with
	sbb = m.nextSB

	if tx > 32 || ty > 32 {
		return -1 // this code doesn't handle this yet...
	}

	// Screenblocks are 32x32, and this code only handles bitmaps fitting in one
	maxX, maxY := 32, 32
	m.nextSB += (maxX * maxY) / TilesPerSB // TODO(kevlar): probably not right

	reg := &machine.Backgrounds.ScreenBlocks[sbb].Regular
	for x := 0; x < tx; x++ {
		for y := 0; y < ty; y++ {
			reg[y][x].Set8(int(tilemap[y*tx+x]))
		}
	}
	return sbb
}

func main() {
	var bgm BackgroundManager8

	bgm.AddPalette(Palette_treasure_map_0[:])
	cbb := bgm.AddTiles(Tileset_0[:])
	sbb := bgm.AddTilemap(Tilemap_0[:], Tilemap_0_Dim.X, Tilemap_0_Dim.Y)
	machine.Backgrounds.Control[0].Configure8(0, cbb, sbb, 0)
	machine.Display.Mode0.Configure()

	for {
	}
}
