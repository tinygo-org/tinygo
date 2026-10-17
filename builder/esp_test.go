package builder

// Tests for the ESP32-S3 image layout in makeESPFirmwareImage. The ESP32-S3
// keeps DROM/IROM out of the ROM-loaded segment table (the ROM bootloader
// rejects images with a DROM segment over 1MB) and appends them at
// 64KB-aligned flash offsets, patching those offsets into the RAM image so
// the startup code can program the cache MMU.
//
// The input ELF is synthesized here rather than compiled, so these tests run
// without an Xtensa toolchain.

import (
	"bytes"
	"crypto/sha256"
	"debug/elf"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

// elf32Section describes one allocated section of the synthetic ELF.
type elf32Section struct {
	name  string
	addr  uint32
	data  []byte
	flags elf.SectionFlag
}

// writeTestELF writes a minimal 32-bit little-endian Xtensa ELF containing
// the given allocated PROGBITS sections plus a symbol table defining syms
// (name -> address). It returns the file path.
func writeTestELF(t *testing.T, entry uint32, sections []elf32Section, syms map[string]uint32) string {
	t.Helper()

	const (
		ehSize  = 52 // Elf32_Ehdr
		shSize  = 40 // Elf32_Shdr
		symSize = 16 // Elf32_Sym
	)

	// Section 0 is the mandatory null entry, followed by the allocated
	// sections, then .symtab, .strtab and .shstrtab.
	symtabIdx := len(sections) + 1
	strtabIdx := symtabIdx + 1
	shstrtabIdx := strtabIdx + 1
	numSections := shstrtabIdx + 1

	// Build the section name table.
	var shstrtab []byte
	shName := func(name string) uint32 {
		off := uint32(len(shstrtab))
		shstrtab = append(shstrtab, name...)
		shstrtab = append(shstrtab, 0)
		return off
	}
	shstrtab = append(shstrtab, 0)

	// Build the symbol table and its string table. Symbols are attributed to
	// the first section, which is enough for elf.File.Symbols.
	strtab := []byte{0}
	symtab := make([]byte, symSize) // index 0 is the null symbol
	for name, value := range syms {
		nameOff := uint32(len(strtab))
		strtab = append(strtab, name...)
		strtab = append(strtab, 0)

		var sym [symSize]byte
		binary.LittleEndian.PutUint32(sym[0:], nameOff)
		binary.LittleEndian.PutUint32(sym[4:], value)
		binary.LittleEndian.PutUint32(sym[8:], 4)
		sym[12] = byte(elf.ST_INFO(elf.STB_GLOBAL, elf.STT_OBJECT))
		binary.LittleEndian.PutUint16(sym[14:], 1)
		symtab = append(symtab, sym[:]...)
	}

	// Lay out the file: header, section headers, then section contents.
	offset := uint32(ehSize + numSections*shSize)
	type placed struct {
		nameOff uint32
		offset  uint32
	}
	allocated := make([]placed, len(sections))
	for i, section := range sections {
		allocated[i] = placed{nameOff: shName(section.name), offset: offset}
		offset += uint32(len(section.data))
	}
	symtabName, symtabOff := shName(".symtab"), offset
	offset += uint32(len(symtab))
	strtabName, strtabOff := shName(".strtab"), offset
	offset += uint32(len(strtab))
	shstrtabName, shstrtabOff := shName(".shstrtab"), offset

	buf := &bytes.Buffer{}
	header := make([]byte, ehSize)
	copy(header, []byte{0x7f, 'E', 'L', 'F'})
	header[4] = byte(elf.ELFCLASS32)
	header[5] = byte(elf.ELFDATA2LSB)
	header[6] = byte(elf.EV_CURRENT)
	binary.LittleEndian.PutUint16(header[16:], uint16(elf.ET_EXEC))
	binary.LittleEndian.PutUint16(header[18:], uint16(elf.EM_XTENSA))
	binary.LittleEndian.PutUint32(header[20:], uint32(elf.EV_CURRENT))
	binary.LittleEndian.PutUint32(header[24:], entry)
	binary.LittleEndian.PutUint32(header[32:], ehSize) // e_shoff
	binary.LittleEndian.PutUint16(header[40:], ehSize) // e_ehsize
	binary.LittleEndian.PutUint16(header[46:], shSize) // e_shentsize
	binary.LittleEndian.PutUint16(header[48:], uint16(numSections))
	binary.LittleEndian.PutUint16(header[50:], uint16(shstrtabIdx))
	buf.Write(header)

	writeSectionHeader := func(nameOff uint32, typ elf.SectionType, flags elf.SectionFlag, addr, off, size, link, info, entsize uint32) {
		var sh [shSize]byte
		binary.LittleEndian.PutUint32(sh[0:], nameOff)
		binary.LittleEndian.PutUint32(sh[4:], uint32(typ))
		binary.LittleEndian.PutUint32(sh[8:], uint32(flags))
		binary.LittleEndian.PutUint32(sh[12:], addr)
		binary.LittleEndian.PutUint32(sh[16:], off)
		binary.LittleEndian.PutUint32(sh[20:], size)
		binary.LittleEndian.PutUint32(sh[24:], link)
		binary.LittleEndian.PutUint32(sh[28:], info)
		binary.LittleEndian.PutUint32(sh[32:], 4)
		binary.LittleEndian.PutUint32(sh[36:], entsize)
		buf.Write(sh[:])
	}

	writeSectionHeader(0, elf.SHT_NULL, 0, 0, 0, 0, 0, 0, 0)
	for i, section := range sections {
		writeSectionHeader(allocated[i].nameOff, elf.SHT_PROGBITS, section.flags,
			section.addr, allocated[i].offset, uint32(len(section.data)), 0, 0, 0)
	}
	writeSectionHeader(symtabName, elf.SHT_SYMTAB, 0, 0, symtabOff, uint32(len(symtab)),
		uint32(strtabIdx), 1, symSize)
	writeSectionHeader(strtabName, elf.SHT_STRTAB, 0, 0, strtabOff, uint32(len(strtab)), 0, 0, 0)
	writeSectionHeader(shstrtabName, elf.SHT_STRTAB, 0, 0, shstrtabOff, uint32(len(shstrtab)), 0, 0, 0)

	for _, section := range sections {
		buf.Write(section.data)
	}
	buf.Write(symtab)
	buf.Write(strtab)
	buf.Write(shstrtab)

	path := filepath.Join(t.TempDir(), "test.elf")
	if err := os.WriteFile(path, buf.Bytes(), 0666); err != nil {
		t.Fatal(err)
	}
	return path
}

// esp32s3TestImage builds an ESP32-S3 image from a synthetic ELF with the
// given .text and .rodata sizes, and returns the image bytes.
func esp32s3TestImage(t *testing.T, textSize, rodataSize int) []byte {
	t.Helper()

	// Mirrors targets/esp32s3.ld: .text starts at ORIGIN(IROM) and .rodata
	// starts one whole 64KB page past the end of .text.
	const iromBase = 0x42000000
	const dromBase = 0x3C000000
	const pageSize = 0x10000
	dromAddr := uint32(dromBase + (textSize+pageSize-1)/pageSize*pageSize)

	text := bytes.Repeat([]byte{0x11}, textSize)
	rodata := bytes.Repeat([]byte{0x22}, rodataSize)
	iram := bytes.Repeat([]byte{0x33}, 256)
	// .data holds _irom_flash_addr and _drom_flash_addr, both zero until the
	// image builder patches them.
	data := make([]byte, 16)

	const dataAddr = 0x3FC88000
	sections := []elf32Section{
		{name: ".rodata", addr: dromAddr, data: rodata, flags: elf.SHF_ALLOC},
		{name: ".data", addr: dataAddr, data: data, flags: elf.SHF_ALLOC | elf.SHF_WRITE},
		{name: ".iram", addr: 0x40378000, data: iram, flags: elf.SHF_ALLOC | elf.SHF_EXECINSTR},
		{name: ".text", addr: iromBase, data: text, flags: elf.SHF_ALLOC | elf.SHF_EXECINSTR},
	}
	syms := map[string]uint32{
		"_irom_flash_addr": dataAddr + 8,
		"_drom_flash_addr": dataAddr + 12,
	}

	infile := writeTestELF(t, 0x40378000, sections, syms)
	outfile := filepath.Join(t.TempDir(), "test.bin")
	if err := makeESPFirmwareImage(infile, outfile, "esp32s3"); err != nil {
		t.Fatal("makeESPFirmwareImage failed:", err)
	}
	image, err := os.ReadFile(outfile)
	if err != nil {
		t.Fatal(err)
	}
	return image
}

// TestESP32S3ImageLayout checks that DROM and IROM stay out of the ROM-loaded
// segment table, that they land at 64KB-aligned flash offsets, and that those
// offsets are patched into the RAM image.
func TestESP32S3ImageLayout(t *testing.T) {
	const pageSize = 0x10000

	for _, tc := range []struct {
		name       string
		textSize   int
		rodataSize int
	}{
		{"small", 0x800, 0x400},
		{"multipage text", 0x30000, 0x400},
		// A DROM segment over 1MB is what the ROM bootloader used to reject.
		{"large rodata", 0x1000, 0x780000},
	} {
		t.Run(tc.name, func(t *testing.T) {
			image := esp32s3TestImage(t, tc.textSize, tc.rodataSize)

			// The ROM only sees the RAM segments: .data and .iram.
			if got := image[1]; got != 2 {
				t.Errorf("segment count = %d, want 2 (.data and .iram only)", got)
			}

			// Both flash regions must be page-aligned, in IROM/DROM order,
			// and must not overlap the RAM image.
			iromAddr := binary.LittleEndian.Uint32(image[0x18+8+8:])
			dromAddr := binary.LittleEndian.Uint32(image[0x18+8+12:])
			if iromAddr%pageSize != 0 || dromAddr%pageSize != 0 {
				t.Errorf("flash offsets not 64KB-aligned: irom=%#x drom=%#x", iromAddr, dromAddr)
			}
			if dromAddr < iromAddr+uint32(tc.textSize) {
				t.Errorf("DROM at %#x overlaps IROM at %#x (%#x bytes)", dromAddr, iromAddr, tc.textSize)
			}
			ramImageEnd := checkRAMImage(t, image)
			if int(iromAddr) < ramImageEnd {
				t.Errorf("IROM at %#x overlaps the RAM image ending at %#x", iromAddr, ramImageEnd)
			}

			// The image is flashed at offset 0 on the ESP32-S3, so the patched
			// flash offsets are also offsets into the image, and the startup
			// code reads the section contents through them.
			if int(iromAddr)+tc.textSize > len(image) || int(dromAddr)+tc.rodataSize > len(image) {
				t.Fatalf("image too short: %d bytes, irom=%#x drom=%#x", len(image), iromAddr, dromAddr)
			}
			if got := image[iromAddr]; got != 0x11 {
				t.Errorf("byte at IROM offset %#x = %#x, want 0x11 (.text)", iromAddr, got)
			}
			if got := image[dromAddr]; got != 0x22 {
				t.Errorf("byte at DROM offset %#x = %#x, want 0x22 (.rodata)", dromAddr, got)
			}
		})
	}
}

// checkRAMImage walks the ROM-loaded part of the image the way the ROM
// bootloader does, verifying the segment table, the trailing checksum byte
// and the appended SHA256 hash. This also covers the flash offsets patched
// into .data, which must be in place before either is computed. It returns
// the offset just past the hash, where the XIP segments start.
func checkRAMImage(t *testing.T, image []byte) int {
	t.Helper()

	if image[0] != 0xE9 {
		t.Fatalf("image magic = %#x, want 0xe9", image[0])
	}

	offset := 0x18 // image header
	checksum := byte(0xEF)
	for i := 0; i < int(image[1]); i++ {
		length := int(binary.LittleEndian.Uint32(image[offset+4:]))
		offset += 8
		if offset+length > len(image) {
			t.Fatalf("segment %d runs past the end of the image", i)
		}
		for _, b := range image[offset : offset+length] {
			checksum ^= b
		}
		offset += length
	}

	offset += 15 - offset%16 // footer padding
	if got := image[offset]; got != checksum {
		t.Errorf("checksum byte = %#x, want %#x", got, checksum)
	}
	offset++

	want := sha256.Sum256(image[:offset])
	if got := image[offset : offset+sha256.Size]; !bytes.Equal(got, want[:]) {
		t.Errorf("appended SHA256 = %x, want %x", got, want[:])
	}
	return offset + sha256.Size
}

// TestESP32S3ImageErrors checks that an ELF the startup code could not boot
// from is rejected, rather than silently producing an unbootable image.
func TestESP32S3ImageErrors(t *testing.T) {
	const dataAddr = 0x3FC88000
	data := elf32Section{name: ".data", addr: dataAddr, data: make([]byte, 16),
		flags: elf.SHF_ALLOC | elf.SHF_WRITE}
	text := elf32Section{name: ".text", addr: 0x42000000, data: make([]byte, 64),
		flags: elf.SHF_ALLOC | elf.SHF_EXECINSTR}
	rodata := elf32Section{name: ".rodata", addr: 0x3C010000, data: make([]byte, 64),
		flags: elf.SHF_ALLOC}
	bothSyms := map[string]uint32{
		"_irom_flash_addr": dataAddr + 8,
		"_drom_flash_addr": dataAddr + 12,
	}

	for _, tc := range []struct {
		name     string
		sections []elf32Section
		syms     map[string]uint32
	}{
		{"no IROM", []elf32Section{rodata, data}, bothSyms},
		{"no DROM", []elf32Section{text, data}, bothSyms},
		{
			// The startup code maps each region as one run of MMU pages, so
			// a second output section in the window cannot be represented.
			name: "two DROM sections",
			sections: []elf32Section{text, rodata, data,
				{name: ".rodata2", addr: 0x3C020000, data: make([]byte, 64), flags: elf.SHF_ALLOC}},
			syms: bothSyms,
		},
		{
			name:     "missing symbol",
			sections: []elf32Section{text, rodata, data},
			syms:     map[string]uint32{"_drom_flash_addr": dataAddr + 12},
		},
		{
			// A flash offset that is not in a RAM segment cannot be patched,
			// so the startup code would read an uninitialized value.
			name:     "symbol outside the RAM segments",
			sections: []elf32Section{text, rodata, data},
			syms: map[string]uint32{
				"_irom_flash_addr": 0x3FC99000,
				"_drom_flash_addr": dataAddr + 12,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			infile := writeTestELF(t, 0x40378000, tc.sections, tc.syms)
			outfile := filepath.Join(t.TempDir(), "test.bin")
			if err := makeESPFirmwareImage(infile, outfile, "esp32s3"); err == nil {
				t.Error("expected an error, got none")
			} else {
				t.Log(err)
			}
		})
	}
}
