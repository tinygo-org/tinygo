//go:build esp32s3 && numa_psram_octal

// Shared PSRAM support for ESP32-S3: linker symbols, MMU/ROM-call helpers,
// and address-map documentation common to both the Quad SPI driver
// (runtime_esp32_psram_qspi.go, build tag numa_psram_qspi) and the Octal
// SPI driver (runtime_esp32_psram_octal.go, build tag numa_psram_octal).
// Everything protocol-specific (SPI1 command sequencing, chip ID/mode
// register parsing, SPI0 cache-phase configuration) lives in those files.
//
// Address map (esp-idf soc/ext_mem_defs.h; verified on hardware while
// bringing up the Octal driver):
//
//	0x600C5000  MMU table, 512 x 32-bit entries, one per 64KB page.
//	            A SINGLE table shared by IBUS (0x42000000 window) and DBUS
//	            (0x3C000000 window), indexed by linear address:
//	            entry = (vaddr & SOC_MMU_LINEAR_ADDR_MASK 0x1FFFFFF) >> 16.
//	            There is no ICache/DCache half-split: 0x3C800000 and
//	            0x42800000 both decode to entry 128.
//	            Entry format: bits[13:0] physical page number, bit 14 =
//	            invalid (SOC_MMU_INVALID), bit 15 = target type, 0 = flash,
//	            1 = PSRAM (SOC_MMU_ACCESS_SPIRAM).
//
//	Flash XIP identity-maps linear page N -> flash page N for both IROM
//	and DROM (esp32s3.S), so the flash image occupies entries
//	0..(image pages - 1), up to 16M. PSRAM lives in the upper half of the
//	DBUS window (0x3D000000, entries 256-511; targets/esp32s3.ld),
//	disjoint from all flash entries.
//
//	DCache (data cache) is the L1/L2 cache in front of the DBUS window.
//	It is disabled (ROM Cache_Disable_DCache, which also invalidates all
//	tag memory) while MMU entries are modified and re-enabled after, so
//	that no stale cache lines survive the remap.
package runtime

import "unsafe"

//go:extern _spsram
var _spsram [0]byte

//go:extern _epsram
var _epsram [0]byte

//go:extern _psram_start
var _psram_start [0]byte

//go:extern _psram_end
var _psram_end [0]byte

// ROM functions baked into every ESP32-S3 chip (fixed addresses provided
// via PROVIDE() in targets/esp32s3.ld); see esp32s3/rom/cache.h.

//export Cache_Dbus_MMU_Set
func romCacheDbusMMUSet(extRam uint32, vaddr uint32, paddr uint32, psize uint32, num uint32, fixed uint32) int32

// Cache_Disable_DCache / Cache_Enable_DCache: the documented-safe pairing
// around an MMU change (Cache_Disable_DCache also invalidates all DCache
// tag memory). Cache_Suspend_DCache/Resume, the alternative pairing, is
// explicitly documented as unsafe to use while changing the MMU.

//export Cache_Disable_DCache
func romCacheDisableDCache() uint32

//export Cache_Enable_DCache
func romCacheEnableDCache(autoload uint32)

const mmuAccessSpiram = 0x8000 // SOC_MMU_TYPE / SOC_MMU_ACCESS_SPIRAM: routes entry to PSRAM instead of flash.

// psramWindow returns the linker-provided PSRAM vaddr window
// (targets/esp32s3.ld) and its capacity in 64KB MMU pages.
func psramWindow() (start, end uintptr, pageCap int) {
	start = uintptr(unsafe.Pointer(&_psram_start))
	end = uintptr(unsafe.Pointer(&_psram_end))
	pageCap = int((end - start) / 65536)
	return
}

// zeroInitPSRAM zero-initializes .psram BSS after PSRAM has been mapped
// into the MMU, verifying with a known pattern first that writes actually
// reach the chip (a hardware or timing issue could otherwise silently
// drop them).
func zeroInitPSRAM() {
	spsram := uintptr(unsafe.Pointer(&_spsram))
	epsram := uintptr(unsafe.Pointer(&_epsram))
	if epsram <= spsram {
		return
	}

	*(*uint32)(unsafe.Pointer(spsram)) = 0x5A5A5A5A
	if *(*uint32)(unsafe.Pointer(spsram)) != 0x5A5A5A5A {
		panic("PSRAM readback test failed")
	}
	*(*uint32)(unsafe.Pointer(spsram)) = 0
	if *(*uint32)(unsafe.Pointer(spsram)) != 0 {
		panic("PSRAM readback test failed")
	}

	for addr := spsram + 4; addr < epsram; addr += 4 {
		*(*uint32)(unsafe.Pointer(addr)) = 0
	}
}
