// The tileprep command handles preparing tiles for the GBA in tinygo.
package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"image"
	"image/color"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	// Supported image file types:
	_ "image/png"
)

var (
	bits   = flag.Int("bits", 4, "Bits per pixel (4 or 8)")
	pkg    = flag.String("package", "main", "Package for output file")
	goOut  = flag.String("go_out", "-", "Output file")
	asmOut = flag.String("asm_out", "-", "Output file")
)

var usage = strings.TrimSpace(`
Usage: tileprep [--options] file.ext [file.ext ...]

Supported file extensions:
	.png

File names are expected to be of the following structure:
	<prefix>[.<index>].<ext>

In --bits=4 mode, all files with the same prefix will share tiles.
In --bits=8 mode, all files with the same index share a palette.

Supported flags:
`)

func main() {
	flag.Usage = func() {
		fmt.Println(usage)
		flag.PrintDefaults()
		os.Exit(17)
	}
	flag.Parse()

	files := flag.Args()
	if len(files) == 0 {
		log.Printf("No files to process")
		flag.Usage()
	}

	switch *bits {
	case 4, 8:
	default:
		log.Printf("Unrecognized bits-per-pixel value %d", *bits)
		flag.Usage()
	}

	var goFile io.Writer
	if *goOut == "-" {
		goFile = os.Stdout
	} else {
		f, err := os.Create(*goOut)
		if err != nil {
			log.Fatalf("preparing --go_out output: %s", err)
		}
		defer func() {
			if err := f.Close(); err != nil {
				log.Fatalf("finalizing --go-out output: %s", err)
			}
		}()
		goFile = f
	}

	var asmFile io.Writer
	if *asmOut == "-" {
		asmFile = os.Stdout
	} else {
		f, err := os.Create(*asmOut)
		if err != nil {
			log.Fatalf("preparing --asm_out output: %s", err)
		}
		defer func() {
			if err := f.Close(); err != nil {
				log.Fatalf("finalizing --asm-out output: %s", err)
			}
		}()
		asmFile = f
	}

	p := newProcessor(goFile)

	for _, file := range files {
		if err := p.process(file); err != nil {
			log.Fatalf("processing files: %s", err)
		}
	}

	fmt.Fprintf(goFile, "// Generated file -- DO NOT EDIT\n\n")
	fmt.Fprintf(goFile, "package %s\n\n", *pkg)

	fmt.Fprintf(asmFile, "# Generated file -- DO NOT EDIT\n\n")
	fmt.Fprintf(asmFile, ".arm\n")
	fmt.Fprintf(asmFile, ".section .rodata.bgtiles\n\n")

	p.writePalettes(goFile, asmFile)
	p.writeTilesets(goFile, asmFile)
}

type processor struct {
	w io.Writer

	palettes map[string]*palette
	tilesets map[string]*tileset
	files    map[string]file
}

func newProcessor(w io.Writer) *processor {
	return &processor{
		w:        w,
		palettes: make(map[string]*palette),
		tilesets: make(map[string]*tileset),
		files:    make(map[string]file),
	}
}

var nameFormat = regexp.MustCompile(`^([a-zA-Z][a-zA-Z0-9_]+)(?:\.([0-9]+))?\.[^.]+$`)

func (p *processor) process(filename string) error {
	log.Printf("Processing %q...", filename)

	match := nameFormat.FindStringSubmatch(filepath.Base(filename))
	if match == nil {
		return fmt.Errorf("file %q does not match required pattern /%s/", filename, nameFormat)
	}
	prefix := match[1]
	index := 0
	if i, err := strconv.Atoi(match[2]); err != nil {
		index = i
	}

	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return err
	}
	log.Printf("  Size:    %d bytes", stat.Size())

	img, format, err := image.Decode(f)
	if err != nil {
		return fmt.Errorf("file %q: %s", filename, err)
	}

	bounds := img.Bounds()
	dx, dy := bounds.Dx(), bounds.Dy()
	log.Printf("  Format:  %dx%d %s (colorspace: %T)", dx, dy, format, img.ColorModel())

	if dx%8 != 0 || dy%8 != 0 {
		return fmt.Errorf("file %q: dimensions %dx%d not a multiple of 8x8", filename, dx, dy)
	}

	var paletteKey, tilesetKey string
	switch *bits {
	case 4:
		paletteKey = prefix
		tilesetKey = fmt.Sprintf("%s_%d", prefix, index)
	case 8:
		paletteKey = fmt.Sprintf("%s_%d", prefix, index)
		tilesetKey = strconv.Itoa(index)
	}
	log.Printf("  Groups:  %s / %s", paletteKey, tilesetKey)
	p.files[filepath.Base(filename)] = file{dx, dy, paletteKey, tilesetKey}

	if _, ok := p.palettes[paletteKey]; !ok {
		p.palettes[paletteKey] = newPalette()
	}
	pal := p.palettes[paletteKey]

	if _, ok := p.tilesets[tilesetKey]; !ok {
		p.tilesets[tilesetKey] = newTileset(dx, dy)
	}
	tset := p.tilesets[tilesetKey]
	if dx != tset.px || dy != tset.py {
		return fmt.Errorf("tileset %q had size %dx%d, file %q has %dx%d", tilesetKey, tset.px, tset.py, filename, dx, dy)
	}

	for tx := 0; tx < tset.tx; tx++ {
		for ty := 0; ty < tset.ty; ty++ {
			var tile [8 * 8]uint8
			for dx := 0; dx < 8; dx++ {
				for dy := 0; dy < 8; dy++ {
					c := img.At(bounds.Min.X+8*tx+dx, bounds.Min.Y+8*ty+dy)
					pindex, err := pal.index(c)
					if err != nil {
						return fmt.Errorf("preparing palette %q: %s", paletteKey, err)
					}
					tile[8*dy+dx] = pindex
				}
			}
			tset.set(tx, ty, tile)
		}
	}

	log.Printf("  Palette: %d colors (total)", len(pal.colors))
	log.Printf("  Tiles:   %dx%d (%d total, %d bytes)", tset.tx, tset.ty, tset.tx*tset.ty, tset.tx*tset.ty*64**bits/8)

	return nil
}

func (p *processor) writeData32(asmFile io.Writer, symbol string, data []uint32) {
	fmt.Fprintf(asmFile, ".global %s\n", symbol)
	fmt.Fprintf(asmFile, ".align  4\n")
	fmt.Fprintf(asmFile, "%s:\n", symbol)
	for _, v := range data {
		fmt.Fprintf(asmFile, "    .word %#08x\n", v)
	}
	fmt.Fprintf(asmFile, "\n")
}

func (p *processor) writeData16(asmFile io.Writer, symbol string, data []uint16) {
	fmt.Fprintf(asmFile, ".global %s\n", symbol)
	fmt.Fprintf(asmFile, ".align  4\n")
	fmt.Fprintf(asmFile, "%s:\n", symbol)
	for _, v := range data {
		fmt.Fprintf(asmFile, "    .hword %#04x\n", v)
	}
	fmt.Fprintf(asmFile, "\n")
}

func (p *processor) writePalettes(goFile, asmFile io.Writer) {
	var keys []string
	for k := range p.palettes {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		v := p.palettes[k]
		symbol := fmt.Sprintf("Palette_%s", k)
		packed := v.pack()

		fmt.Fprintf(goFile, "//go:extern %s\n", symbol)
		fmt.Fprintf(goFile, "var %s [%d]uint16\n\n", symbol, len(packed))

		p.writeData16(asmFile, symbol, packed)
	}
}

func (p *processor) writeTilesets(goFile, asmFile io.Writer) {
	var keys []string
	for k := range p.tilesets {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		ts := p.tilesets[k]
		tileset, tilemap := ts.pack()

		tilesetSymbol := fmt.Sprintf("Tileset_%s", k)
		fmt.Fprintf(goFile, "//go:extern %s\n", tilesetSymbol)
		fmt.Fprintf(goFile, "var %s [%d]uint32\n\n", tilesetSymbol, len(tileset))
		p.writeData32(asmFile, tilesetSymbol, tileset)

		tilemapSymbol := fmt.Sprintf("Tilemap_%s", k)
		fmt.Fprintf(goFile, "//go:extern %s\n", tilemapSymbol)
		fmt.Fprintf(goFile, "var %s [%d]uint16\n\n", tilemapSymbol, len(tilemap))
		p.writeData16(asmFile, tilemapSymbol, tilemap)

		fmt.Fprintf(goFile, "var %s_Dim = struct{X,Y int}{%d,%d}", tilemapSymbol, ts.tx, ts.ty)
	}
}

type file struct {
	dx, dy     int
	paletteKey string
	tilesetKey string
}

type palette struct {
	colors map[color.RGBA]uint8
}

func newPalette() *palette {
	return &palette{
		colors: make(map[color.RGBA]uint8),
	}
}

func (pal *palette) index(c color.Color) (uint8, error) {
	rgba := color.RGBAModel.Convert(c).(color.RGBA)
	if index, ok := pal.colors[rgba]; ok {
		return index, nil
	}
	switch *bits {
	case 4:
		if len(pal.colors) >= 16 {
			return 0, fmt.Errorf("16-color palette exhausted")
		}
	case 8:
		if len(pal.colors) >= 256 {
			return 0, fmt.Errorf("256-color palette exhausted")
		}
	}
	index := len(pal.colors)
	pal.colors[rgba] = uint8(index)
	return uint8(index), nil
}

func (pal *palette) pack() []uint16 {
	packed := make([]uint16, len(pal.colors))
	for c, i := range pal.colors {
		packed[i] = uint16(c.R>>3) | uint16(c.G>>3)<<5 | uint16(c.B>>3)<<10
	}
	return packed
}

type tileset struct {
	px, py  int                  // in pixels
	tx, ty  int                  // in tiles
	indexOf map[[8 * 8]uint8]int // location of tiles in the tileset
	tileset [][8 * 8]uint8       // tiles
	tilemap [][]uint16           // where tiles go
	tilearr []uint16             // straight-line tilemap
}

func newTileset(dx, dy int) *tileset {
	tx, ty := dx/8, dy/8

	tilearr := make([]uint16, tx*ty)
	tilemap := make([][]uint16, ty)
	for i := range tilemap {
		tilemap[i] = tilearr[i*tx:][:tx:tx]
	}

	return &tileset{
		px:      dx,
		py:      dy,
		tx:      tx,
		ty:      ty,
		indexOf: make(map[[8 * 8]uint8]int),
		tilemap: tilemap,
		tilearr: tilearr,
	}
}

func (tset *tileset) set(tx, ty int, tile [8 * 8]uint8) {
	index, ok := tset.indexOf[tile]
	if !ok {
		index = len(tset.tileset)
		tset.indexOf[tile] = index
		tset.tileset = append(tset.tileset, tile)
	}
	tset.tilemap[ty][tx] = uint16(index)
}

func (tset *tileset) pack() (tiles []uint32, tilemap []uint16) {
	switch *bits {
	case 4:
		tiles = make([]uint32, 0, len(tset.tileset)*8)
		for _, tile := range tset.tileset {
			for r := 0; r < 8; r++ {
				row := tile[8*r:][:8]
				packed := uint32(row[0]&0xF) |
					uint32(row[1]&0xF)<<4 |
					uint32(row[2]&0xF)<<8 |
					uint32(row[3]&0xF)<<12 |
					uint32(row[4]&0xF)<<16 |
					uint32(row[5]&0xF)<<20 |
					uint32(row[6]&0xF)<<24 |
					uint32(row[7]&0xF)<<28
				tiles = append(tiles, packed)
			}
		}
	case 8:
		tiles = make([]uint32, 0, len(tset.tileset)*16)
		for _, tile := range tset.tileset {
			for r := 0; r < 8; r++ {
				row := tile[8*r:][:8]
				lo := binary.LittleEndian.Uint32(row[:4])
				hi := binary.LittleEndian.Uint32(row[4:])
				tiles = append(tiles, lo, hi)
			}
		}
	}
	return tiles, tset.tilearr
}
