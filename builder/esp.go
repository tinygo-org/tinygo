package builder

// This file implements support for writing ESP image files. These image files
// are read by the ROM bootloader so have to be in a particular format.
//
// In the future, it may be necessary to implement support for other image
// formats, such as the ESP8266 image formats (again, used by the ROM bootloader
// to load the firmware).

import (
	"bytes"
	"crypto/sha256"
	"debug/elf"
	"encoding/binary"
	"fmt"
	"os"
	"sort"
	"strings"
)

type espImageSegment struct {
	addr uint32
	data []byte
}

// ESP32-S3 flash-mapped virtual address windows (esp-idf
// soc/esp32s3/ext_mem_defs.h, narrowed to what targets/esp32s3.ld uses).
// Segments in these ranges are XIP'd from flash through the cache MMU
// instead of being loaded into RAM by the ROM bootloader.
const (
	esp32s3DromLow  = 0x3C000000 // DBUS window; its upper half (0x3D000000+) is PSRAM, not flash.
	esp32s3DromHigh = 0x3D000000
	esp32s3IromLow  = 0x42000000 // IBUS window.
	esp32s3IromHigh = 0x44000000
)

const (
	// esp32FlashBase is the flash offset esptool writes the ESP32 image to.
	// ESP32-S3 images are written at offset 0 instead (see
	// flashBinUsingEsp32), so there image offsets are flash offsets.
	esp32FlashBase = 0x1000

	// espFlashPageSize is the flash cache MMU page size. The MMU supports
	// page sizes down to 256 B, but 64 KiB is the reset/default value and is
	// what the startup code relies on. If the startup code ever changes the
	// page size, this constant must change with it.
	espFlashPageSize = 0x10000
)

// makeESPFirmwareImage converts an input ELF file to an image file for an ESP32 or
// ESP8266 chip. This is a special purpose image format just for the ESP chip
// family, and is parsed by the on-chip mask ROM bootloader.
//
// The following documentation has been used:
// https://github.com/espressif/esptool/wiki/Firmware-Image-Format
// https://github.com/espressif/esp-idf/blob/8fbb63c2a701c22ccf4ce249f43aded73e134a34/components/bootloader_support/include/esp_image_format.h#L58
// https://github.com/espressif/esptool/blob/master/esptool.py
func makeESPFirmwareImage(infile, outfile, format string) error {
	inf, err := elf.Open(infile)
	if err != nil {
		return err
	}
	defer inf.Close()

	// Load all segments to be written to the image. These are actually ELF
	// sections, not true ELF segments (similar to how esptool does it).
	var segments []*espImageSegment
	for _, section := range inf.Sections {
		if section.Type != elf.SHT_PROGBITS || section.Size == 0 || section.Flags&elf.SHF_ALLOC == 0 {
			continue
		}
		data, err := section.Data()
		if err != nil {
			return fmt.Errorf("failed to read section data: %w", err)
		}
		for len(data)%4 != 0 {
			// Align segment to 4 bytes.
			data = append(data, 0)
		}
		if uint64(uint32(section.Addr)) != section.Addr {
			return fmt.Errorf("section address too big: 0x%x", section.Addr)
		}
		segments = append(segments, &espImageSegment{
			addr: uint32(section.Addr),
			data: data,
		})
	}

	// Sort the segments by address. This is what esptool does too.
	sort.SliceStable(segments, func(i, j int) bool { return segments[i].addr < segments[j].addr })

	// Write first to an in-memory buffer, primarily so that we can easily
	// calculate a hash over the entire image.
	// An added benefit is that we don't need to check for errors all the time.
	outf := &bytes.Buffer{}

	// Separate esp32 and esp32-img. The -img suffix indicates we should make an
	// image, not just a binary to be flashed at 0x1000 for example.
	chip := format
	makeImage := false
	if strings.HasSuffix(format, "-img") {
		makeImage = true
		chip = format[:len(format)-len("-img")]
	}

	// Separate RAM segments loaded by the ROM bootloader from flash-mapped
	// segments initialized by the TinyGo startup code.
	var flashSegments []*espImageSegment
	switch chip {
	case "esp32":
		var ramSegments []*espImageSegment
		for _, seg := range segments {
			if (seg.addr >= 0x3F400000 && seg.addr < 0x3F800000) ||
				(seg.addr >= 0x400D0000 && seg.addr < 0x40400000) {
				flashSegments = append(flashSegments, seg)
			} else {
				ramSegments = append(ramSegments, seg)
			}
		}
		segments = ramSegments

	case "esp32s3":
		// The DBUS cache window runs to 0x3E000000, but its upper half is
		// reserved for PSRAM (targets/esp32s3.ld), which is never backed by
		// flash. Only the DROM half below 0x3D000000 is flash-mapped.
		var ramSegments []*espImageSegment
		for _, seg := range segments {
			if (seg.addr >= esp32s3DromLow && seg.addr < esp32s3DromHigh) ||
				(seg.addr >= esp32s3IromLow && seg.addr < esp32s3IromHigh) {
				flashSegments = append(flashSegments, seg)
			} else {
				ramSegments = append(ramSegments, seg)
			}
		}
		segments = ramSegments
	}

	// ESP32 flash XIP: compute where the DROM segment will be placed in flash
	// (page-aligned, right after the RAM segments) and patch the
	// _drom_flash_addr variable so the startup code can program the cache MMU.
	var esp32DromFlashAddr uint32
	if chip == "esp32" && len(flashSegments) > 0 {
		esp32DromFlashAddr = uint32(alignUpFlashPage(esp32FlashBase + ramImageSize(segments, makeImage)))

		syms, err := inf.Symbols()
		if err != nil {
			return fmt.Errorf("ESP32: %w", err)
		}
		if err := patchFlashAddr(syms, segments, "_drom_flash_addr", esp32DromFlashAddr); err != nil {
			return fmt.Errorf("ESP32: %w", err)
		}
	}

	// ESP32-S3 flash XIP: the ROM bootloader rejects images with a DROM
	// segment over 1MB, so IROM and DROM are kept out of the segment table
	// (see above) and appended at page-aligned flash offsets instead. Those
	// offsets are patched into the RAM image for the startup code to program
	// the cache MMU with.
	var esp32s3Irom, esp32s3Drom *espImageSegment
	var esp32s3IromFlashAddr, esp32s3DromFlashAddr uint32
	if chip == "esp32s3" && len(flashSegments) > 0 {
		var err error
		esp32s3Irom, err = singleFlashSegment(flashSegments, esp32s3IromLow, esp32s3IromHigh, "IROM")
		if err != nil {
			return fmt.Errorf("ESP32-S3: %w", err)
		}
		esp32s3Drom, err = singleFlashSegment(flashSegments, esp32s3DromLow, esp32s3DromHigh, "DROM")
		if err != nil {
			return fmt.Errorf("ESP32-S3: %w", err)
		}

		// The image is flashed at offset 0, so image offsets are flash
		// offsets. IROM goes right after the RAM image and DROM right after
		// IROM, each rounded up to an MMU page. The linker script rounds the
		// virtual addresses up the same way, so the offsets within a page
		// match on both sides of the mapping.
		esp32s3IromFlashAddr = uint32(alignUpFlashPage(ramImageSize(segments, makeImage)))
		esp32s3DromFlashAddr = esp32s3IromFlashAddr + uint32(alignUpFlashPage(len(esp32s3Irom.data)))

		syms, err := inf.Symbols()
		if err != nil {
			return fmt.Errorf("ESP32-S3: %w", err)
		}
		if err := patchFlashAddr(syms, segments, "_irom_flash_addr", esp32s3IromFlashAddr); err != nil {
			return fmt.Errorf("ESP32-S3: %w", err)
		}
		if err := patchFlashAddr(syms, segments, "_drom_flash_addr", esp32s3DromFlashAddr); err != nil {
			return fmt.Errorf("ESP32-S3: %w", err)
		}
	}

	// Calculate checksum over the segment data. This is used in the image
	// footer.
	checksum := uint8(0xef)
	for _, segment := range segments {
		for _, b := range segment.data {
			checksum ^= b
		}
	}

	if makeImage {
		// The bootloader starts at 0x1000, or 4096.
		// TinyGo doesn't use a separate bootloader and runs the entire
		// application in the bootloader location.
		outf.Write(make([]byte, 4096))
	}

	// Chip IDs. Source:
	// https://github.com/espressif/esp-idf/blob/v4.3/components/bootloader_support/include/esp_app_format.h#L22
	chip_id := map[string]uint16{
		"esp32":   0x0000,
		"esp32c3": 0x0005,
		"esp32c6": 0x000d,
		"esp32s3": 0x0009,
	}[chip]

	// SPI flash speed/size byte (byte 3 of header):
	//   Upper nibble = flash size, lower nibble = flash frequency.
	//   The espflasher auto-detects and patches the flash size (upper nibble),
	//   but the frequency (lower nibble) must be correct per chip.
	spiSpeedSize := map[string]uint8{
		"esp32":   0x1f, // 80MHz=0x0F, 2MB=0x10
		"esp32c3": 0x1f, // 80MHz=0x0F, 2MB=0x10
		"esp32c6": 0x10, // 80MHz=0x00, 2MB=0x10 (C6 uses different freq encoding)
		"esp32s3": 0x1f, // 80MHz=0x0F, 2MB=0x10
	}[chip]

	// Image header.
	switch chip {
	case "esp32", "esp32c3", "esp32s3", "esp32c6":
		// Header format:
		// https://github.com/espressif/esp-idf/blob/v4.3/components/bootloader_support/include/esp_app_format.h#L71
		// Note: not adding a SHA256 hash as the binary is modified by
		// esptool.py while flashing and therefore the hash won't be valid
		// anymore.
		binary.Write(outf, binary.LittleEndian, struct {
			magic          uint8
			segment_count  uint8
			spi_mode       uint8
			spi_speed_size uint8
			entry_addr     uint32
			wp_pin         uint8
			spi_pin_drv    [3]uint8
			chip_id        uint16
			min_chip_rev   uint8
			reserved       [8]uint8
			hash_appended  bool
		}{
			magic:          0xE9,
			segment_count:  byte(len(segments)),
			spi_mode:       2, // ESP_IMAGE_SPI_MODE_DIO
			spi_speed_size: spiSpeedSize,
			entry_addr:     uint32(inf.Entry),
			wp_pin:         0xEE, // disable WP pin
			chip_id:        chip_id,
			hash_appended:  true, // add a SHA256 hash
		})
	case "esp8266":
		// Header format:
		// https://github.com/espressif/esptool/wiki/Firmware-Image-Format
		// Basically a truncated version of the ESP32 header.
		binary.Write(outf, binary.LittleEndian, struct {
			magic          uint8
			segment_count  uint8
			spi_mode       uint8
			spi_speed_size uint8
			entry_addr     uint32
		}{
			magic:          0xE9,
			segment_count:  byte(len(segments)),
			spi_mode:       0,    // irrelevant, replaced by esptool when flashing
			spi_speed_size: 0x20, // spi_speed, spi_size: replaced by esptool when flashing
			entry_addr:     uint32(inf.Entry),
		})
	default:
		return fmt.Errorf("builder: unknown binary format %#v, expected esp32 or esp8266", format)
	}

	// Write all segments to the image.
	// https://github.com/espressif/esptool/wiki/Firmware-Image-Format#segment
	for _, segment := range segments {
		binary.Write(outf, binary.LittleEndian, struct {
			addr   uint32
			length uint32
		}{
			addr:   segment.addr,
			length: uint32(len(segment.data)),
		})
		outf.Write(segment.data)
	}

	// Footer, including checksum.
	// The entire image size must be a multiple of 16, so pad the image to one
	// byte less than that before writing the checksum.
	outf.Write(make([]byte, 15-outf.Len()%16))
	outf.WriteByte(checksum)

	if chip != "esp8266" {
		// SHA256 hash (to protect against image corruption, not for security).
		hash := sha256.Sum256(outf.Bytes())
		outf.Write(hash[:])
	}

	// For ESP32: append flash-mapped segments (DROM/IROM) at page-aligned flash
	// offsets after the RAM portion. The startup code maps them via the flash
	// cache MMU (DROM at esp32DromFlashAddr, patched into _drom_flash_addr).
	if chip == "esp32" && len(flashSegments) > 0 {
		const flashBase = esp32FlashBase
		const pageSize = espFlashPageSize
		dromFlashAddr := esp32DromFlashAddr

		// Separate DROM and IROM segments.
		var dromSegs, iromSegs []*espImageSegment
		for _, seg := range flashSegments {
			if seg.addr >= 0x3F400000 && seg.addr < 0x3F800000 {
				dromSegs = append(dromSegs, seg)
			} else {
				iromSegs = append(iromSegs, seg)
			}
		}

		// Write DROM segments at the computed page-aligned flash offset.
		dromSize := 0
		if len(dromSegs) > 0 {
			targetImageOffset := int(dromFlashAddr - flashBase)
			if outf.Len() > targetImageOffset {
				return fmt.Errorf("ESP32: RAM segments too large (%d bytes), overlap DROM at flash 0x%x", outf.Len(), dromFlashAddr)
			}
			outf.Write(make([]byte, targetImageOffset-outf.Len()))
			for _, seg := range dromSegs {
				outf.Write(seg.data)
				dromSize += len(seg.data)
			}
		}

		// Write IROM segments immediately after DROM, at the next page boundary.
		// IROM flash addr = dromFlashAddr + ceil(dromSize/pageSize)*pageSize
		// (must match the computation in the startup assembly).
		if len(iromSegs) > 0 {
			dromPages := (dromSize + pageSize - 1) / pageSize
			if dromPages == 0 {
				dromPages = 1
			}
			iromFlashAddr := dromFlashAddr + uint32(dromPages)*pageSize
			targetImageOffset := int(iromFlashAddr - flashBase)
			if outf.Len() > targetImageOffset {
				return fmt.Errorf("ESP32: DROM too large, overlaps IROM at flash 0x%x", iromFlashAddr)
			}
			outf.Write(make([]byte, targetImageOffset-outf.Len()))
			for _, seg := range iromSegs {
				outf.Write(seg.data)
			}
		}
	}

	// For ESP32-S3: append the XIP segments at the flash offsets patched into
	// the image above. Both are page-aligned and in ascending order, so each
	// one only needs padding up to its own offset.
	if chip == "esp32s3" && len(flashSegments) > 0 {
		for _, region := range []struct {
			name    string
			offset  uint32
			segment *espImageSegment
		}{
			{"IROM", esp32s3IromFlashAddr, esp32s3Irom},
			{"DROM", esp32s3DromFlashAddr, esp32s3Drom},
		} {
			if outf.Len() > int(region.offset) {
				return fmt.Errorf("ESP32-S3: image is %d bytes, overlapping %s at flash offset 0x%x",
					outf.Len(), region.name, region.offset)
			}
			outf.Write(make([]byte, int(region.offset)-outf.Len()))
			outf.Write(region.segment.data)
		}
	}

	// QEMU (or more precisely, qemu-system-xtensa from Espressif) expects the
	// image to be a certain size.
	if makeImage {
		// Use a default image size of 4MB.
		grow := 4096*1024 - outf.Len()
		if grow > 0 {
			outf.Write(make([]byte, grow))
		}
	}

	// Write the image to the output file.
	return os.WriteFile(outfile, outf.Bytes(), 0666)
}

// alignUpFlashPage rounds size up to the next flash cache MMU page boundary.
func alignUpFlashPage(size int) int {
	return (size + espFlashPageSize - 1) &^ (espFlashPageSize - 1)
}

// ramImageSize returns the size of the part of the image that the ROM
// bootloader loads: the header, the segment headers and their data, the
// footer holding the checksum, and the appended SHA256 hash.
//
// Flash-mapped (XIP) segments are appended after this portion, and their
// flash offsets have to be known before the image is written, because they
// are patched into the image itself. This function must therefore predict
// exactly what makeESPFirmwareImage writes; keep the two in sync.
func ramImageSize(segments []*espImageSegment, makeImage bool) int {
	size := 0
	if makeImage {
		size += 4096 // padding in front of the image header
	}
	size += 24 // image header (8) + trailer fields (16)
	for _, segment := range segments {
		size += 8 + len(segment.data) // segment header + data (4-aligned)
	}
	size += 16 - size%16 // footer padding + checksum byte
	size += 32           // appended SHA256 hash
	return size
}

// patchFlashAddr stores value in the 32-bit variable named by symbol, which
// must live in one of the RAM segments. The startup code reads it to program
// the flash cache MMU. Patching must happen before the checksum and hash are
// computed, so that the patched value is covered by both.
func patchFlashAddr(syms []elf.Symbol, segments []*espImageSegment, symbol string, value uint32) error {
	var symbolAddr uint64
	found := false
	for _, sym := range syms {
		if sym.Name == symbol {
			symbolAddr = sym.Value
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("symbol %s not found", symbol)
	}

	for _, segment := range segments {
		start := uint64(segment.addr)
		end := start + uint64(len(segment.data))
		if symbolAddr >= start && symbolAddr+4 <= end {
			binary.LittleEndian.PutUint32(segment.data[symbolAddr-start:], value)
			return nil
		}
	}
	return fmt.Errorf("symbol %s (0x%x) not in a RAM segment", symbol, symbolAddr)
}

// singleFlashSegment returns the one segment inside the [low, high) virtual
// address window, named name in error messages. The startup code maps each
// XIP region as a single run of MMU pages and the linker script sizes the
// regions accordingly, so anything other than exactly one segment per window
// means the two have drifted apart and the image would not boot.
func singleFlashSegment(segments []*espImageSegment, low, high uint32, name string) (*espImageSegment, error) {
	var found *espImageSegment
	for _, segment := range segments {
		if segment.addr < low || segment.addr >= high {
			continue
		}
		if found != nil {
			return nil, fmt.Errorf("expected a single %s segment, found more than one", name)
		}
		found = segment
	}
	if found == nil {
		return nil, fmt.Errorf("%s segment not found", name)
	}
	return found, nil
}
