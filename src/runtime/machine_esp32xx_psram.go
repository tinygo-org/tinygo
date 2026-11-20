//go:build esp32s3 && machine.psram

// Shared PSRAM support for ESP32-S3: linker symbols, MMU/ROM-call helpers,
// and address-map documentation for the Octal SPI driver
// (machine_esp32s3_psram_octal.go, build tag machine.psram).
// Everything protocol-specific (SPI1 command sequencing, chip ID/mode
// register parsing, SPI0 cache-phase configuration) lives in that file.
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
//	Flash XIP takes the low entries: IROM starts at entry 0, DROM follows
//	after it, and machine.Flash follows DROM (esp32s3.S, targets/esp32s3.ld).
//	PSRAM lives in the upper half of the DBUS window (0x3D000000, entries
//	256-511; targets/esp32s3.ld), disjoint from all flash entries.
//
//	The IROM/DROM boundary is variable, since it follows the size of
//	.text. The PSRAM window sits above it either way: the boundary can
//	only move down, never above 0x3D000000, so the vaddr the drivers pass
//	to romCacheDbusMMUSet always stays inside whatever DROM range
//	Cache_Set_IDROM_MMU_Size left behind.
//
//	DCache (data cache) is the L1/L2 cache in front of the DBUS window.
//	It is disabled (ROM Cache_Disable_DCache, which also invalidates all
//	tag memory) while MMU entries are modified and re-enabled after, so
//	that no stale cache lines survive the remap.
package runtime

import "unsafe"

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

// psramWindow returns the PSRAM vaddr window that the target declares with
// __psram_size (targets/esp32s3.ld) and its size in 64KB MMU pages.
func psramWindow() (start, end uintptr, pageCap int) {
	start = uintptr(unsafe.Pointer(&_psram_start))
	end = uintptr(unsafe.Pointer(&_psram_end))
	pageCap = int((end - start) / 65536)
	return
}
