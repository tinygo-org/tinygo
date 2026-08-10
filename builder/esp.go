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

	// For ESP32 (original): separate RAM segments (loadable by ROM bootloader)
	// from flash-mapped segments (DROM/IROM, require MMU setup by startup code).
	// The ROM bootloader on ESP32 does NOT handle flash-mapped segments —
	// it tries to memcpy to the virtual address, which crashes.
	var flashSegments []*espImageSegment
	if chip == "esp32" {
		var ramSegments []*espImageSegment
		for _, seg := range segments {
			if (seg.addr >= 0x3F400000 && seg.addr < 0x3F800000) || // DROM
				(seg.addr >= 0x400D0000 && seg.addr < 0x40400000) { // IROM
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
	// This must happen before the checksum/hash are computed so the patched
	// value is covered by both.
	const esp32FlashBase = 0x1000 // esptool flashes the image at 0x1000
	// The ESP32 flash cache MMU supports configurable page sizes down to 256 B. 64 KiB is the reset/default size.
	// If the startup code ever changes the MMU page size, this constant must change too.
	const esp32PageSize = 0x10000 // 64KB MMU pages
	var esp32DromFlashAddr uint32
	if chip == "esp32" && len(flashSegments) > 0 {
		// Compute the size of the RAM portion of the image (everything the ROM
		// bootloader loads, up to and including the appended SHA256 hash).
		ramImageSize := 0
		ramImageSize += 24 // image header (8) + trailer fields (16)
		for _, seg := range segments {
			ramImageSize += 8 + len(seg.data) // segment header + data (4-aligned)
		}
		ramImageSize += 16 - ramImageSize%16 // footer padding + checksum byte
		ramImageSize += 32                   // appended SHA256 hash

		// DROM flash address must be 64KB page-aligned.
		esp32DromFlashAddr = uint32(esp32FlashBase+ramImageSize+esp32PageSize-1) &^ (esp32PageSize - 1)

		// Patch _drom_flash_addr in whichever RAM segment contains it.
		syms, _ := inf.Symbols()
		var dromSymAddr uint64
		for _, s := range syms {
			if s.Name == "_drom_flash_addr" {
				dromSymAddr = s.Value
				break
			}
		}
		if dromSymAddr == 0 {
			return fmt.Errorf("ESP32: _drom_flash_addr symbol not found")
		}
		patched := false
		for _, seg := range segments {
			if dromSymAddr >= uint64(seg.addr) && dromSymAddr+4 <= uint64(seg.addr)+uint64(len(seg.data)) {
				off := int(dromSymAddr - uint64(seg.addr))
				binary.LittleEndian.PutUint32(seg.data[off:], esp32DromFlashAddr)
				patched = true
				break
			}
		}
		if !patched {
			return fmt.Errorf("ESP32: _drom_flash_addr (0x%x) not in any RAM segment", dromSymAddr)
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
	if len(flashSegments) > 0 {
		const flashBase = esp32FlashBase
		const pageSize = esp32PageSize
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
			if makeImage {
				targetImageOffset = int(dromFlashAddr)
			}
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
			if makeImage {
				targetImageOffset = int(iromFlashAddr)
			}
			if outf.Len() > targetImageOffset {
				return fmt.Errorf("ESP32: DROM too large, overlaps IROM at flash 0x%x", iromFlashAddr)
			}
			outf.Write(make([]byte, targetImageOffset-outf.Len()))
			for _, seg := range iromSegs {
				outf.Write(seg.data)
			}
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
