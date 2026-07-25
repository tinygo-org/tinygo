//go:build esp32s3 && numa_psram_octal

// Octal SPI (OPI) PSRAM support for ESP32-S3 (e.g. ESP32-S3-WROOM-1U-N16R8,
// combo modules with in-package Octal PSRAM). Octal PSRAM uses 8 data
// lines and a DTR (double transfer rate / DDR) protocol, unlike the 4-line
// single-rate Quad PSRAM handled by runtime_esp32_psram_qspi.go; the two
// are not interchangeable and use different GPIOs (Octal adds D4-D7+DQS
// on GPIO33-37, on top of the CS1/D0-D3 pins shared with flash). Shared
// linker symbols, MMU/ROM-call helpers, and the address map are
// documented in runtime_esp32_psram.go.
//
// Unlike its own Quad PSRAM driver, ESP-IDF does not hand-roll the Octal
// DTR command sequence in C -- it delegates entirely to a mask-ROM
// function (esp_rom_opiflash_exec_cmd, esp32s3/rom/opi_flash.h) that every
// ESP32-S3 chip has at a fixed address. There is no accessible reference
// for what that ROM function does internally, so this file calls it
// directly (matching esp_psram_impl_octal.c's call sites) rather than
// guessing at the DDR-mode register sequence.
//
// This driver performs ESP-IDF's DQS/MSPI-delay timing calibration
// (mspi_timing_psram_tuning / s_select_best_tuning_config_dtr). It performs an
// iterative sweep of 14 MSPI input delay and extra dummy cycle candidate
// parameters at 80 MHz DTR (160 MHz MSPI core clock) using reference pattern
// reads written during 40 MHz bring-up, selecting the optimal configuration
// to achieve 80 MHz DTR (160 MB/s bandwidth).
//
// GPIO33-37 (the Octal-only D4-D7+DQS lines) are configured by calling the
// ROM's esp_rom_opiflash_pin_config(), which reads the efuse OPI-pin
// strapping. The ROM does not do this on its own at boot: its boot-time
// pin setup follows the flash boot mode (DIO on modules with a plain Quad
// flash chip, which only needs D0/D1), and a tinygo image has no
// 2nd-stage bootloader to do it either.
package runtime

import (
	"device/esp"
	"unsafe"
)

// ROM functions baked into every ESP32-S3 chip (fixed addresses provided
// via PROVIDE() in targets/esp32s3.ld). Declared per esp32s3/rom/opi_flash.h.

//export esp_rom_opiflash_exec_cmd
func romOpiflashExecCmd(spiNum int32, mode int32, cmd uint32, cmdBitLen int32, addr uint32, addrBitLen int32, dummyBits int32, mosiData *byte, mosiBitLen int32, misoData *byte, misoBitLen int32, csMask uint32, isWriteErase bool)

//export esp_rom_spi_set_dtr_swap_mode
func romSpiSetDtrSwapMode(spiNum int32, wrSwap bool, rdSwap bool)

//export esp_rom_opiflash_pin_config
func romOpiflashPinConfig()

// g_rom_spiflash_dummy_len_plus[1]: per-spi_num extra dummy cycle count that
// esp_rom_opiflash_exec_cmd() adds on top of the caller-supplied dummyBits.
//
//go:extern g_rom_spiflash_dummy_len_plus
var romDummyLenPlusSymbol [0]byte

const (
	opiSpiNum  = 1 // SPI1: command engine used for PSRAM setup, matches esp_psram_impl_octal.c's spi_num=1.
	opiDtrMode = 7 // ESP_ROM_SPIFLASH_OPI_DTR_MODE (esp_rom_spiflash.h enum: QIO=0,QOUT,DIO,DOUT,FASTRD,SLOWRD,OPI_STR,OPI_DTR=7).
	opiCS1Mask = 2 // BIT(1): esp_psram_impl_octal.c passes this literal cs_mask value for every PSRAM command.

	// Octal PSRAM command opcodes and phase lengths, matching the
	// OPI_PSRAM_*/OCT_PSRAM_* constants in esp_psram_impl_octal.c.
	opiRegRead    = 0x4040
	opiRegWrite   = 0xC0C0
	opiSyncRead   = 0x0000
	opiSyncWrite  = 0x8080
	opiCmdBitLen  = 16
	opiAddrBitLen = 32
	opiRegDummy   = 2 * (5 - 1)  // 8 dummy cycles for MRx register read.
	opiRdDummy    = 2 * (10 - 1) // 18 dummy cycles for sync read.
	opiWrDummy    = 2 * (5 - 1)  // 8 dummy cycles for sync write.

	opiVendorIDAP    = 0xD  // s_print_psram_info / OCT_PSRAM_VENDOR_ID_AP.
	opiVendorIDUnilc = 0x1A // OCT_PSRAM_VENDOR_ID_UNILC.

	csHoldTime  = 3 // OCT_PSRAM_CS_HOLD_TIME.
	csSetupTime = 3 // OCT_PSRAM_CS_SETUP_TIME.
	csHoldDelay = 2 // OCT_PSRAM_CS_HOLD_DELAY.

	// SPI0/SPI1 SMEM/FMEM clock divider for a fixed 40 MHz bring-up/fallback
	// speed. SCLK = source/(N+1), H=((N+1)/2)-1, L=N. Also the speed
	// calibratePSRAMTiming reverts to if the 80 MHz DTR timing sweep can't
	// find a working config -- same din_mode/din_num/extra-dummy of 0 as
	// this bring-up path, which already passed checkPSRAMConnected.
	spiClkCntN = 1
	spiClkCntH = 0
	spiClkCntL = 1
)

// setPSRAMClockDivider sets SPI0's SRAM (PSRAM) clock and SPI1's FMEM (flash)
// clock divider. SCLK = source/(N+1), H=((N+1)/2)-1, L=N.
//
// WARNING: this writes to SPI1.CLOCK as well, not just to the SRAM clock. During
// PSRAM bring-up and timing calibration this is intentional (SPI1 drives PSRAM
// command sequences); any caller that only wants to touch the SRAM clock must use
// the individual SPI0.SetSRAM_CLK_* calls directly.
//
// Lives in IRAM: calibratePSRAMTiming calls this with flash cache disabled,
// so the code itself can't be fetched from flash-mapped memory at that point.
//
//go:section .iram
func setPSRAMClockDivider(n, h, l uint32) {
	esp.SPI0.SetSRAM_CLK_SCLK_EQU_SYSCLK(0)
	esp.SPI0.SetSRAM_CLK_SCLKCNT_N(n)
	esp.SPI0.SetSRAM_CLK_SCLKCNT_H(h)
	esp.SPI0.SetSRAM_CLK_SCLKCNT_L(l)
	esp.SPI1.SetCLOCK_CLK_EQU_SYSCLK(0)
	esp.SPI1.SetCLOCK_CLKCNT_N(n)
	esp.SPI1.SetCLOCK_CLKCNT_H(h)
	esp.SPI1.SetCLOCK_CLKCNT_L(l)
}

// initPSRAM performs the Octal PSRAM bring-up sequence, mirroring
// esp_psram_impl_octal.c's esp_psram_impl_enable (minus DQS/timing
// calibration, see file doc comment):
//  1. CS1 pin select, CS timing, fixed low-speed clock.
//  2. Mode register (MR0) init: fixed read latency.
//  3. Connectivity check via a sync write/read of a reference word.
//  4. Read vendor ID (MR1) and density (MR2), validate and size the MMU window.
//  5. Configure SPI0's cache-facing read/write command phases for Octal DTR.
//  6. Map PSRAM into the MMU, zero-initialize .psram BSS.
func initPSRAM() {
	psramStart, _, psramMmuPagesCap := psramWindow()

	// Enable SPI0/SPI1 peripheral clock and clear resets.
	esp.SYSTEM.SetPERIP_CLK_EN0_SPI01_CLK_EN(1)
	esp.SYSTEM.SetPERIP_RST_EN0_SPI01_RST(0)

	// Ensure MMU table power is enabled and force-on.
	esp.EXTMEM.SetCACHE_MMU_POWER_CTRL_CACHE_MMU_MEM_FORCE_ON(1)
	esp.EXTMEM.SetCACHE_MMU_POWER_CTRL_CACHE_MMU_MEM_FORCE_PU(1)

	// CS1 pin (GPIO26): function 0 is the dedicated SPICS1 signal and the
	// IO_MUX reset default; function 1 switches the pin to plain GPIO,
	// disconnecting it from SPI1 (soc/io_mux_reg.h: FUNC_SPICS1_SPICS1 = 0).
	// FUN_DRV(3) sets max drive strength (3 = 40 mA) for the CS1 and
	// SPICLK pads; Octal OPI at 80 MHz DTR needs stronger drive than the
	// Quad SPI driver's FUN_WPU pull-up approach (the Quad driver operates
	// at single-data-rate 40 MHz and the pull-up ensures CS1 isn't
	// floating when deasserted).
	esp.IO_MUX.SetGPIO26_MCU_SEL(0)
	esp.IO_MUX.SetGPIO26_FUN_DRV(3)
	esp.SPI0.SetDATE_SPI_SMEM_SPICLK_FUN_DRV(3)

	// GPIO33-37 (Octal D4-D7+DQS) are not configured by the ROM at boot
	// (see file doc comment); the efuse.h "GPIO33-37 powered by VDDSPI"
	// note is about voltage domain, not pin function. This ROM call does
	// the actual per-efuse pad configuration.
	romOpiflashPinConfig()

	// CS/hold timing, shared by SPI0 and SPI1 for PSRAM
	// (esp_psram_impl_octal.c s_set_psram_cs_timing).
	esp.SPI0.SetSPI_SMEM_AC_SPI_SMEM_CS_SETUP(1)
	esp.SPI0.SetSPI_SMEM_AC_SPI_SMEM_CS_HOLD(1)
	esp.SPI0.SetSPI_SMEM_AC_SPI_SMEM_CS_HOLD_TIME(csHoldTime)
	esp.SPI0.SetSPI_SMEM_AC_SPI_SMEM_CS_SETUP_TIME(csSetupTime)
	esp.SPI0.SetSPI_SMEM_AC_SPI_SMEM_CS_HOLD_DELAY(csHoldDelay)

	// Fixed low-speed clock for SPI0 and SPI1 (no DQS/timing calibration).
	setPSRAMClockDivider(spiClkCntN, spiClkCntH, spiClkCntL)

	// Variable-dummy DDR mode on SPI1, no DTR read/write swap
	// (esp_psram_impl_octal.c esp_psram_impl_enable).
	esp.SPI1.SetDDR_SPI_FMEM_VAR_DUMMY(1)
	romSpiSetDtrSwapMode(opiSpiNum, false, false)

	// Mode register MR0: fixed latency, read_latency=2, drive_str=0
	// (esp_psram_impl_octal.c: mode_reg.mr0 = {lt:1, read_latency:2, drive_str:0}).
	mr0 := readPSRAMModeReg(0x0)
	mr0 = (mr0 &^ 0x3f) | (2 << 2) | (1 << 5)
	writePSRAMModeReg(0x0, mr0)

	if !checkPSRAMConnected() {
		panic("PSRAM init failed: sync read/write mismatch, chip not detected")
	}

	mr1 := readPSRAMModeReg(0x1) // vendor id
	mr2 := readPSRAMModeReg(0x2) // density / dev id / gb
	vendorID := mr1 & 0x1f
	if vendorID != opiVendorIDAP && vendorID != opiVendorIDUnilc {
		panic("PSRAM init failed: unrecognized vendor id")
	}

	// Density encoding (esp_psram_impl_octal.c:372-376): 0x1=4MB, 0x3=8MB,
	// 0x5=16MB, in 64KB pages. Chips reporting 0x7 (32MB) or 0x6 (64MB)
	// fall through unmatched and are simply capped at psramMmuPagesCap,
	// same as any chip bigger than the window: targets/esp32s3.ld only
	// reserves a 16M PSRAM window, so there's nothing a bigger case value
	// could map to.
	density := mr2 & 0x7
	mmuPages := psramMmuPagesCap
	switch density {
	case 0x1:
		if mmuPages > 64 {
			mmuPages = 64 // 4MB
		}
	case 0x3:
		if mmuPages > 128 {
			mmuPages = 128 // 8MB
		}
	case 0x5:
		if mmuPages > 256 {
			mmuPages = 256 // 16MB
		}
	}

	// Perform DQS / MSPI delay timing calibration to upgrade Octal PSRAM to
	// 80 MHz DTR. On failure it reverts to the untuned 40 MHz bring-up
	// speed instead of running at an unverified timing; either way, the
	// cache-facing register setup below is unaffected by which speed won.
	calibratePSRAMTiming()

	// Disable DCache before touching the SPI0 cache-facing registers and
	// the MMU table below (see the Cache_Disable_DCache doc comment in
	// runtime_esp32_psram.go).
	romCacheDisableDCache()

	// Enable SPI0's CS1 output for its cache-triggered SRAM (PSRAM)
	// transactions. SPI_MEM_CS1_DIS on SPI0.MISC defaults to 1 (disabled;
	// soc/spi_mem_reg.h bitpos:[1]), so without this the chip is never
	// selected for SPI0 accesses. Matches ESP-IDF's quad PSRAM driver:
	// "ENABLE SPI0 CS1 TO PSRAM (CS0--FLASH; CS1--SRAM)"
	// (esp32s2/esp_psram_impl_quad.c:546); the generated SPI0_Type has no
	// named CS1_DIS accessor, hence ClearBits.
	esp.SPI0.MISC.ClearBits(1 << 1)

	// Configure SPI0's cache-facing read/write command phases for Octal
	// DTR PSRAM access (esp_psram_impl_octal.c s_config_psram_spi_phases).
	esp.SPI0.SetCACHE_SCTRL_CACHE_SRAM_USR_WCMD(1)
	esp.SPI0.SetSRAM_DWR_CMD_CACHE_SRAM_USR_WR_CMD_BITLEN(opiCmdBitLen - 1)
	esp.SPI0.SetSRAM_DWR_CMD_CACHE_SRAM_USR_WR_CMD_VALUE(opiSyncWrite)

	esp.SPI0.SetCACHE_SCTRL_CACHE_SRAM_USR_RCMD(1)
	esp.SPI0.SetSRAM_DRD_CMD_CACHE_SRAM_USR_RD_CMD_BITLEN(opiCmdBitLen - 1)
	esp.SPI0.SetSRAM_DRD_CMD_CACHE_SRAM_USR_RD_CMD_VALUE(opiSyncRead)

	esp.SPI0.SetCACHE_SCTRL_SRAM_ADDR_BITLEN(opiAddrBitLen - 1)
	esp.SPI0.SetCACHE_SCTRL_CACHE_USR_SCMD_4BYTE(1)

	esp.SPI0.SetCACHE_SCTRL_USR_RD_SRAM_DUMMY(1)
	esp.SPI0.SetCACHE_SCTRL_USR_WR_SRAM_DUMMY(1)
	esp.SPI0.SetCACHE_SCTRL_SRAM_RDUMMY_CYCLELEN(opiRdDummy - 1)
	esp.SPI0.SetSPI_SMEM_DDR_SPI_SMEM_VAR_DUMMY(1)
	esp.SPI0.SetCACHE_SCTRL_SRAM_WDUMMY_CYCLELEN(opiWrDummy - 1)

	esp.SPI0.SetSPI_SMEM_DDR_WDAT_SWP(0)
	esp.SPI0.SetSPI_SMEM_DDR_RDAT_SWP(0)
	esp.SPI0.SetSPI_SMEM_DDR_EN(1)

	esp.SPI0.SetSRAM_CMD_SDUMMY_OUT(1)
	esp.SPI0.SetSRAM_CMD_SCMD_OCT(1)
	esp.SPI0.SetSRAM_CMD_SADDR_OCT(1)
	esp.SPI0.SetSRAM_CMD_SDOUT_OCT(1)
	esp.SPI0.SetSRAM_CMD_SDIN_OCT(1)
	esp.SPI0.SetCACHE_SCTRL_SRAM_OCT(1)

	// Map PSRAM into the MMU: PSRAM physical page 0 at _psram_start, 64KB
	// pages, fixed=0 (physical pages grow linearly with virtual pages).
	if romCacheDbusMMUSet(mmuAccessSpiram, uint32(psramStart), 0, 64, uint32(mmuPages), 0) != 0 {
		panic("PSRAM init failed: Cache_Dbus_MMU_Set error")
	}

	// Re-enable DCache. Cache_Disable_DCache above already invalidated all
	// tag memory.
	romCacheEnableDCache(0)

	zeroInitPSRAM()
}

//export rom_spi_flash_disable_cache
func romSpiFlashDisableCache(cpuid uint32, savedState *uint32)

//export rom_spi_flash_restore_cache
func romSpiFlashRestoreCache(cpuid uint32, savedState uint32)

//go:section .iram
func calibratePSRAMTiming() (ok bool) {
	// Disable both caches so that no flash fetches can be initiated by the
	// cache controller during clock switches.
	var cacheState uint32
	romSpiFlashDisableCache(0, &cacheState)

	origCoreClkSel := esp.SPI0.GetCORE_CLK_SEL()

	// Disable variable dummy mode on SPI1 during timing calibration (matching ESP-IDF).
	esp.SPI1.SetDDR_SPI_FMEM_VAR_DUMMY(0)

	// 1. Write reference data pattern (64 bytes) to PSRAM address 0 at low speed (40 MHz).
	var refData [16]uint32
	seed := uint32(0xa5ff005a)
	for i := range refData {
		seed = seed*1664525 + 1013904223
		refData[i] = seed
	}
	romOpiflashExecCmd(opiSpiNum, opiDtrMode, opiSyncWrite, opiCmdBitLen, 0, opiAddrBitLen, opiWrDummy, (*byte)(unsafe.Pointer(&refData[0])), 512, nil, 0, opiCS1Mask, false)

	// 2. Read the original flash clock configuration to scale the divider when the core clock increases.
	origFlashClock := esp.SPI0.CLOCK.Get()
	var origFlashDiv uint32 = 1
	if (origFlashClock & (1 << 31)) == 0 {
		origFlashDiv = ((origFlashClock >> 16) & 0xff) + 1
	}
	newFlashDiv := origFlashDiv * 2
	newFlashClock := ((newFlashDiv - 1) << 16) | (((newFlashDiv/2 - 1) & 0xff) << 8) | (newFlashDiv - 1)

	// Apply scaled divider to SPI0 and SPI1 flash clock registers.
	esp.SPI0.CLOCK.Set(newFlashClock)
	esp.SPI1.CLOCK.Set(newFlashClock)

	// Switch MSPI core clock to 160 MHz.
	esp.SPI0.SetCORE_CLK_SEL(2) // 160 MHz core clock

	// Set SPI0/SPI1 SRAM clocks to 80 MHz (divider = 2).
	setPSRAMClockDivider(1, 0, 1)

	// Declare candidate delay parameters inside IRAM to avoid RODATA flash access.
	psramTuningParams := [14]struct {
		dinMode       uint8
		dinNum        uint8
		extraDummyLen uint8
	}{
		{0, 0, 0},
		{4, 2, 2},
		{2, 1, 2},
		{4, 1, 2},
		{1, 0, 1},
		{4, 0, 2}, // default config index 5
		{0, 0, 1},
		{4, 2, 3},
		{2, 1, 3},
		{4, 1, 3},
		{1, 0, 2},
		{4, 0, 3},
		{0, 0, 2},
		{4, 2, 4},
	}

	romDummyPlus1 := (*uint8)(unsafe.Pointer(&romDummyLenPlusSymbol))
	origRomDummyPlus1 := *romDummyPlus1

	// 3. Sweep all 14 timing configurations and record successful reads.
	var success [14]bool
	for i := 0; i < 14; i++ {
		p := psramTuningParams[i]
		dinModeReg := (uint32(p.dinMode) & 7) * 0x01249249
		dinNumReg := (uint32(p.dinNum) & 3) * 0x00015555
		esp.SPI0.SPI_SMEM_DIN_MODE.Set(dinModeReg)
		esp.SPI0.SPI_SMEM_DIN_NUM.Set(dinNumReg)

		if p.extraDummyLen > 0 {
			esp.SPI1.TIMING_CALI.Set(2 | ((uint32(p.extraDummyLen) & 7) << 2))
			esp.SPI0.SPI_SMEM_TIMING_CALI.Set(2 | ((uint32(p.extraDummyLen) & 7) << 2))
		} else {
			esp.SPI1.TIMING_CALI.Set(0)
			esp.SPI0.SPI_SMEM_TIMING_CALI.Set(0)
		}
		*romDummyPlus1 = origRomDummyPlus1 + p.extraDummyLen

		// Inline clear of SPI1 FIFO (W0..W15 registers).
		esp.SPI1.W0.Set(0)
		esp.SPI1.W1.Set(0)
		esp.SPI1.W2.Set(0)
		esp.SPI1.W3.Set(0)
		esp.SPI1.W4.Set(0)
		esp.SPI1.W5.Set(0)
		esp.SPI1.W6.Set(0)
		esp.SPI1.W7.Set(0)
		esp.SPI1.W8.Set(0)
		esp.SPI1.W9.Set(0)
		esp.SPI1.W10.Set(0)
		esp.SPI1.W11.Set(0)
		esp.SPI1.W12.Set(0)
		esp.SPI1.W13.Set(0)
		esp.SPI1.W14.Set(0)
		esp.SPI1.W15.Set(0)

		var readData [16]uint32
		romOpiflashExecCmd(opiSpiNum, opiDtrMode, opiSyncRead, opiCmdBitLen, 0, opiAddrBitLen, opiRdDummy, nil, 0, (*byte)(unsafe.Pointer(&readData[0])), 512, opiCS1Mask, false)

		if readData == refData {
			success[i] = true
		}
	}

	// 5. Select best calibration index from passing candidates, in order of
	// preference; this covers all 14 measured candidates, so a working but
	// non-preferred config is never silently dropped in favor of an
	// unverified default.
	priority := [14]uint32{4, 10, 6, 12, 0, 5, 1, 2, 3, 7, 8, 9, 11, 13}
	var bestIdx uint32
	ok = false
	for _, idx := range priority {
		if success[idx] {
			bestIdx = idx
			ok = true
			break
		}
	}

	if ok {
		// 6. Apply best calibration parameters to SPI0 (cache) and
		// clear SPI1's leftover sweep timing.
		best := psramTuningParams[bestIdx]
		dinModeReg := (uint32(best.dinMode) & 7) * 0x01249249
		dinNumReg := (uint32(best.dinNum) & 3) * 0x00015555
		esp.SPI0.SPI_SMEM_DIN_MODE.Set(dinModeReg)
		esp.SPI0.SPI_SMEM_DIN_NUM.Set(dinNumReg)

		if best.extraDummyLen > 0 {
			esp.SPI0.SPI_SMEM_TIMING_CALI.Set(2 | ((uint32(best.extraDummyLen) & 7) << 2))
		} else {
			esp.SPI0.SPI_SMEM_TIMING_CALI.Set(0)
		}
		esp.SPI1.TIMING_CALI.Set(0)
	} else {
		// No candidate passed the 80 MHz DTR read-back: revert clocks and
		// din_mode/din_num/timing-cali to the same untuned 40 MHz state
		// used for bring-up (spiClkCntN/H/L), which already passed
		// checkPSRAMConnected, instead of running PSRAM at an unverified
		// timing.
		esp.SPI0.CLOCK.Set(origFlashClock)
		esp.SPI1.CLOCK.Set(origFlashClock)
		esp.SPI0.SetCORE_CLK_SEL(origCoreClkSel)

		// Restore SPI0 SRAM clock for 40 MHz PSRAM.
		setPSRAMClockDivider(spiClkCntN, spiClkCntH, spiClkCntL)

		esp.SPI0.SPI_SMEM_DIN_MODE.Set(0)
		esp.SPI0.SPI_SMEM_DIN_NUM.Set(0)
		esp.SPI0.SPI_SMEM_TIMING_CALI.Set(0)
		esp.SPI1.TIMING_CALI.Set(0)
	}

	// Restore variable dummy mode on SPI1.
	esp.SPI1.SetDDR_SPI_FMEM_VAR_DUMMY(1)

	// Restore ROM global dummy plus value.
	*romDummyPlus1 = origRomDummyPlus1

	// Re-enable caches.
	romSpiFlashRestoreCache(0, cacheState)
	return ok
}

// readPSRAMModeReg reads a 16-bit PSRAM mode register (MR0..MR8) via the
// ROM's Octal DTR command executor, matching esp_psram_impl_octal.c's
// s_get_psram_mode_reg / s_init_psram_mode_reg read calls.
func readPSRAMModeReg(addr uint32) uint32 {
	var val uint32
	romOpiflashExecCmd(opiSpiNum, opiDtrMode, opiRegRead, opiCmdBitLen, addr, opiAddrBitLen, opiRegDummy, nil, 0, (*byte)(unsafe.Pointer(&val)), 16, opiCS1Mask, false)
	return val
}

// writePSRAMModeReg writes a 16-bit PSRAM mode register, matching
// esp_psram_impl_octal.c's s_init_psram_mode_reg write calls.
func writePSRAMModeReg(addr uint32, val uint32) {
	romOpiflashExecCmd(opiSpiNum, opiDtrMode, opiRegWrite, opiCmdBitLen, addr, opiAddrBitLen, 0, (*byte)(unsafe.Pointer(&val)), 16, nil, 0, opiCS1Mask, false)
}

// checkPSRAMConnected writes a reference word to PSRAM address 0 and reads
// it back, matching esp_psram_impl_octal.c's s_check_psram_connected.
func checkPSRAMConnected() bool {
	refData := uint32(0x5a6b7c8d)
	romOpiflashExecCmd(opiSpiNum, opiDtrMode, opiSyncWrite, opiCmdBitLen, 0, opiAddrBitLen, opiWrDummy, (*byte)(unsafe.Pointer(&refData)), 32, nil, 0, opiCS1Mask, false)

	var got uint32
	romOpiflashExecCmd(opiSpiNum, opiDtrMode, opiSyncRead, opiCmdBitLen, 0, opiAddrBitLen, opiRdDummy, nil, 0, (*byte)(unsafe.Pointer(&got)), 32, opiCS1Mask, false)

	return got == refData
}
