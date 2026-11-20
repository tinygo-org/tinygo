//go:build esp32p4

// ESP32-P4 external PSRAM / Octal-PSRAM (MSPI) bring-up driver.
//
// The CrowPanel 1024x600 RGB565 panel needs a full 1.2MB framebuffer, which
// does not fit in the 640KB on-die SRAM, so the frame must live in the board's
// 32MB octal PSRAM (AP-Hex / "hexapole" die).  ESP-IDF allocates the DSI DPI
// frame buffer from PSRAM (MALLOC_CAP_SPIRAM) and write-backs the L1/L2 caches
// (esp_cache_msync) before the DW-GDMA reads it on AXI port 1, which bypasses
// the caches.  This driver replicates that MSPI+MMU+cache bring-up.
//
// Register layout notes:
//   - The PSRAM "system" controller is MSPI id 2 (SPIMEM2) @ 0x5008E000, and the
//     "peripheral" controller is MSPI id 3 (SPIMEM3) @ 0x5008F000.  These are the
//     same SPI_MEM IP as the flash SPI0 block, so SPI0_Type (which carries the
//     shared SPI_MEM_* / SPI_SMEM_* registers and their field accessors) is
//     reused here cast to the PSRAM base.
//   - The per-region MMU (SPI_MEM_S_MMU_ITEM_{CONTENT,INDEX} @ +0x37C/0x380)
//     maps PSRAM virtual pages (64KB) into CPU address space at 0x48000000.
//   - Cache write-back for DMA coherency uses the CACHE_SYNC_* registers at
//     CACHE base 0x3ff10000 (+0x98 CTRL, +0x9c MAP, +0xa0 ADDR, +0xa4 SIZE).
//
// Bring-up is intentionally incremental and committed at each step (Phase 6):
//
//	Step 1: enable PSRAM module clock + release dual-MSPI reset + clock source
//	Step 3: map PSRAM into CPU address space via the MMU (0x48000000..)
//	Step 4: L2 (and L1-D) cache write-back so the DMA sees fresh pixels
//	Step 5: allocate the full 1024x600 frame in PSRAM and feed the display
package machine

import (
	"device/esp"
	"runtime/volatile"
	"unsafe"
)

// ROM helper wrappers for the ESP32-P4 PSRAM user-command path.  These call
// the same esp_rom_spi_* functions ESP-IDF's AP-Hex PSRAM driver uses for its
// mode-register and probe transactions.  The RISC-V trampolines live in
// src/device/esp/esp32p4_psram.S, which is compiled as a target extra-file
// (TinyGo does not assemble .S files discovered inside Go packages).
//
// Addresses from components/esp_rom/esp32p4/ld/esp32p4.rom.ld (IDF 5.5.5).

//go:export psram_rom_set_op_mode
func psramROMSetOpMode(spiNum uint32, mode uint32)

//go:export psram_rom_cmd_config
func psramROMCmdConfig(spiNum uint32, cmd *psramROMCommand)

//go:export psram_rom_cmd_start
func psramROMCmdStart(spiNum uint32, rxBuf *uint32, rxLen uint32, csMask uint32, writeErase uint32)

// psramROMCommand mirrors esp_rom_spi_cmd_t from
// components/esp_rom/include/esp_rom_spiflash.h.
type psramROMCommand struct {
	cmd          uint16
	cmdBitLen    uint16
	addr         *uint32
	addrBitLen   uint32
	txData       *uint32
	txDataBitLen uint32
	rxData       *uint32
	rxDataBitLen uint32
	dummyBitLen  uint32
}

// ESP_ROM_SPIFLASH_OPI_DTR_MODE from esp_rom_spiflash.h.
const psramROMOPIDTRMode = 7

// psramROMSpiID maps an MSPI register base to the controller index expected by
// the ROM helpers: MSPI2 (system) = 2, MSPI3 (peripheral) = 3.
func psramROMSpiID(base uintptr) uint32 {
	if base == psramMspiPeriBase {
		return 3
	}
	return 2
}

// Base addresses for the dual PSRAM MSPI controllers (same SPI_MEM IP as SPI0).
const (
	psramMspiSystemBase = 0x5008E000 // MSPI id 2: SPIMEM2 controller (system/cache path)
	psramMspiPeriBase   = 0x5008F000 // MSPI id 3: SPIMEM3 (peripheral path)
)

// PSRAM clock-source selects (reg_psram_clk_src_sel).
const (
	psramClkSrcXTAL = 0
	psramClkSrcMPLL = 1
	psramClkSrcSPLL = 2 // ESP-IDF default; already up (drives MIPI-DSI PLL)
	psramClkSrcCPLL = 3
)

// PSRAM MMU configuration.  The P4 per-target MMU (SPI_MEM_S, MSPI id 2) maps
// 64KB pages from PSRAM0 into CPU virtual address space starting at 0x48000000
// (SOC_MMU_PSRAM_VADDR_BASE).  The physical page number is 0-based within the
// PSRAM0 region; the PSRAM0 physical AXI base is 0x88000000 (direct/uncached).
const (
	psramVAddrBase    = 0x48000000
	psramPageSize     = 0x10000 // 64KB
	psramMMUEntryMask = psramPageSize*1024 - 1

	// MMU ITEM_CONTENT flag bits within the PSRAM (SPI_MEM_S) MMU:
	socMMUPsramValid   = 1 << 11 // SOC_MMU_PSRAM_VALID
	socMMUPAccessPsram = 1 << 10 // SOC_MMU_ACCESS_PSRAM
)

// CACHE_SYNC_* registers describe the two-level cache write-back/invalidate
// operation used to make CPU-written PSRAM visible to the DW-GDMA (which reads
// via AXI and bypasses the caches).
const (
	cacheBase                       = 0x3ff10000
	cacheSyncCtrlOff                = 0x98
	cacheSyncMapOff                 = 0x9c
	cacheSyncAddrOff                = 0xa0
	cacheSyncSizeOff                = 0xa4
	cacheSyncInvalidateEna          = 1 << 0 // CACHE_INVALIDATE_ENA
	cacheSyncWriteBackEna           = 1 << 2 // CACHE_WRITEBACK_ENA
	cacheSyncWriteBackInvalidateEna = 1 << 3 // CACHE_WRITEBACK_INVALIDATE_ENA
	cacheSyncDone                   = 1 << 4 // CACHE_SYNC_DONE (set when operation finished)
	// CACHE_SYNC_MAP bits: [4]=L1-DCache, [5]=L2-Cache.
	cacheSyncMapL1DCache = 1 << 4
	cacheSyncMapL2Cache  = 1 << 5
	// L1 DCache line size on ESP32-P4 (used for cache-line alignment).
	cacheLineSize = 64
)

// psramMSPI returns the system PSRAM MSPI controller register block, viewed
// through the shared SPI_MEM register layout of SPI0_Type.
func psramMSPI() *esp.SPI0_Type {
	return (*esp.SPI0_Type)(unsafe.Pointer(uintptr(psramMspiSystemBase)))
}

// EnablePSRAMModuleClock gates on the PSRAM MSPI functional clocks and releases
// the dual-MSPI reset, then selects the PSRAM clock source and sets the core
// clock divider.  This mirrors ESP-IDF psram_ctrlr_ll_enable_module_clock /
// _reset_module_clock / _select_clk_source / _set_core_clock_div.
//
// Step 1 of the PSRAM bring-up.  After calling this the PSRAM MSPI controller
// registers are readable and the APB/AXI paths out of reset.
func EnablePSRAMModuleClock() {
	// Gate the PSRAM system APB clock and the PLL/core functional clocks.
	esp.HP_SYS_CLKRST.SetSOC_CLK_CTRL0_REG_PSRAM_SYS_CLK_EN(1)
	esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL00_REG_PSRAM_PLL_CLK_EN(1)
	esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL00_REG_PSRAM_CORE_CLK_EN(1)

	// Pulse the dual-MSPI AXI/APB reset (assert, then release).  Order matches
	// ESP-IDF _psram_ctrlr_ll_reset_module_clock.
	esp.HP_SYS_CLKRST.SetHP_RST_EN0_REG_RST_EN_DUAL_MSPI_AXI(1)
	esp.HP_SYS_CLKRST.SetHP_RST_EN0_REG_RST_EN_DUAL_MSPI_APB(1)
	esp.HP_SYS_CLKRST.SetHP_RST_EN0_REG_RST_EN_DUAL_MSPI_APB(0)
	esp.HP_SYS_CLKRST.SetHP_RST_EN0_REG_RST_EN_DUAL_MSPI_AXI(0)

	// Select the PSRAM clock source.  The AP-Hex driver uses MPLL (320MHz at
	// 80MHz bus speed); ESP-IDF's esp_psram_impl_enable calls
	// psram_ctrlr_ll_select_clk_source(..., PSRAM_CLK_SRC_MPLL) after the
	// reset pulse.
	esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL00_REG_PSRAM_CLK_SRC_SEL(psramClkSrcMPLL)
}

// SetPSRAMCoreClockDiv sets the PSRAM core-clock divider (freqdiv-1 in the HW
// field, matching ESP-IDF _psram_ctrlr_ll_set_core_clock_div).
func SetPSRAMCoreClockDiv(freqdiv uint32) {
	if freqdiv == 0 {
		freqdiv = 1
	}
	esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL00_REG_PSRAM_CORE_CLK_DIV_NUM(freqdiv - 1)
}

// PSRAMClockState returns the raw clock source / enable register so bring-up
// code can confirm Step 1 took effect on-target.
func PSRAMClockState() (src, coreEn, pllEn, sysEn uint32) {
	return esp.HP_SYS_CLKRST.GetPERI_CLK_CTRL00_REG_PSRAM_CLK_SRC_SEL(),
		esp.HP_SYS_CLKRST.GetPERI_CLK_CTRL00_REG_PSRAM_CORE_CLK_EN(),
		esp.HP_SYS_CLKRST.GetPERI_CLK_CTRL00_REG_PSRAM_PLL_CLK_EN(),
		esp.HP_SYS_CLKRST.GetSOC_CLK_CTRL0_REG_PSRAM_SYS_CLK_EN()
}

// MapPSRAM maps `pages` (64KB each) of physical PSRAM0 into CPU virtual address
// space starting at the given vaddr (which must be 64KB aligned).  For the
// default region this is vaddr 0x48000000 with a 0-based physical page start.
// It mirrors ESP-IDF mmu_hal_map_region(mmu_id=1, MMU_TARGET_PSRAM0, ...).
//
// Step 3 of the PSRAM bring-up: after the module-clock/reset step this makes
// the PSRAM readable/writable from Go at the returned base address.
func MapPSRAM(vaddr, physPageStart uintptr, pages uint32) uintptr {
	mmu := psramMSPI()
	paddrPage := uint32(physPageStart >> 16)
	for i := uint32(0); i < pages; i++ {
		entryID := uint32((vaddr+uintptr(i)*psramPageSize)&psramMMUEntryMask) >> 16
		mmuVal := (paddrPage + i) | socMMUPAccessPsram | socMMUPsramValid
		mmu.SetSPI_MEM_MMU_ITEM_INDEX(entryID)
		mmu.SetSPI_MEM_MMU_ITEM_CONTENT(mmuVal)
	}
	return vaddr
}

// cacheSync ctrl pointer (CACHE_SYNC_CTRL_REG base + 0x98).
func cacheSyncCtrl() *volatile.Register32 {
	return (*volatile.Register32)(unsafe.Pointer(uintptr(cacheBase + cacheSyncCtrlOff)))
}
func cacheSyncMap() *volatile.Register32 {
	return (*volatile.Register32)(unsafe.Pointer(uintptr(cacheBase + cacheSyncMapOff)))
}
func cacheSyncAddr() *volatile.Register32 {
	return (*volatile.Register32)(unsafe.Pointer(uintptr(cacheBase + cacheSyncAddrOff)))
}
func cacheSyncSize() *volatile.Register32 {
	return (*volatile.Register32)(unsafe.Pointer(uintptr(cacheBase + cacheSyncSizeOff)))
}

// cacheSyncOp performs a cache sync operation (invalidate, writeback, etc.)
// matching ESP-IDF's Cache_WriteBack_Addr / Cache_Invalidate_Addr pattern:
//   - Aligns addr DOWN and size UP to cache-line boundaries
//   - Triggers the operation TWICE (hardware requirement on ESP32-P4)
func cacheSyncOp(addr uintptr, size uintptr, ena uint32) {
	// Align address down and size up to cache-line boundaries, matching ESP-IDF:
	//   plus = addr & (cache_line_size - 1); addr -= plus; size += plus;
	//   size = (size + cache_line_size - 1) & ~(cache_line_size - 1);
	plus := addr & (cacheLineSize - 1)
	addr -= plus
	size += plus
	size = (size + cacheLineSize - 1) &^ (cacheLineSize - 1)

	cacheSyncMap().Set(cacheSyncMapL1DCache | cacheSyncMapL2Cache)
	cacheSyncAddr().Set(uint32(addr))
	cacheSyncSize().Set(uint32(size))
	cacheSyncCtrl().Set(ena)
	for cacheSyncCtrl().Get()&cacheSyncDone == 0 {
	}
	// ESP-IDF triggers the operation a second time (hw requirement).
	cacheSyncCtrl().Set(ena)
	for cacheSyncCtrl().Get()&cacheSyncDone == 0 {
	}
}

// CacheInvalidate discards L1-DCACHE and L2-CACHE lines covering the virtual
// address range [addr, addr+size) WITHOUT writing dirty data back to memory.
func CacheInvalidate(addr uintptr, size uintptr) {
	if size == 0 {
		return
	}
	cacheSyncOp(addr, size, cacheSyncInvalidateEna)
}

// CacheInvalidateL1 discards L1-DCACHE-only lines for [addr, addr+size).
// Use for SRAM regions — SRAM is not cached by L2.  This prevents stale
// dirty .bss zero lines from evicting back over NC-alias-written data.
func CacheInvalidateL1(addr uintptr, size uintptr) {
	if size == 0 {
		return
	}
	plus := addr & (cacheLineSize - 1)
	addr -= plus
	size += plus
	size = (size + cacheLineSize - 1) &^ (cacheLineSize - 1)

	cacheSyncMap().Set(cacheSyncMapL1DCache) // L1 only, not L2
	cacheSyncAddr().Set(uint32(addr))
	cacheSyncSize().Set(uint32(size))
	cacheSyncCtrl().Set(cacheSyncInvalidateEna)
	for cacheSyncCtrl().Get()&cacheSyncDone == 0 {
	}
	cacheSyncCtrl().Set(cacheSyncInvalidateEna)
	for cacheSyncCtrl().Get()&cacheSyncDone == 0 {
	}
}

// CacheWriteBack flushes the dirty L1-DCACHE and L2-CACHE lines covering the
// virtual address range [addr, addr+size) out to memory, so the DW-GDMA (which
// reads via AXI, bypassing the caches) observes the freshly written data.
// Mirrors ESP-IDF esp_cache_msync(..., DIR_C2M) for the P4 cache-sync block.
func CacheWriteBack(addr uintptr, size uintptr) {
	if size == 0 {
		return
	}
	cacheSyncOp(addr, size, cacheSyncWriteBackEna)
}

// CacheWriteBackAll flushes ALL dirty L1-DCACHE and L2-CACHE lines to memory.
// This is the simplest way to ensure DMA coherency when the exact range is
// uncertain or when multiple scattered regions need flushing.
func CacheWriteBackAll() {
	cacheSyncMap().Set(cacheSyncMapL1DCache | cacheSyncMapL2Cache)
	cacheSyncAddr().Set(0)
	cacheSyncSize().Set(0)
	cacheSyncCtrl().Set(cacheSyncWriteBackEna)
	for cacheSyncCtrl().Get()&cacheSyncDone == 0 {
	}
	cacheSyncCtrl().Set(cacheSyncWriteBackEna)
	for cacheSyncCtrl().Get()&cacheSyncDone == 0 {
	}
}

// --- MPLL bring-up (Step 2 clock) -------------------------------------------
//
// The AP-Hex PSRAM driver selects PSRAM_CLK_SRC_MPLL (not SPLL) and configures
// MPLL to AP_HEX_PSRAM_MPLL_DEFAULT_FREQ_MHZ (400MHz) from the 40MHz XTAL, so
// the 200MHz PSRAM bus clock is MPLL/2 (freqdiv 2).  MPLL's analog divider is
// programmed through the LP analog I2C master (regi2c) on the MSPI block.

// regi2c (analog I2C master) register field layout of I2C0_CTRL (LP_I2C_ANA_MST).
const (
	r2cSlaveIDSh = 0  // bits [7:0] slave id (block)
	r2cAddrSh    = 8  // bits [15:8] register address
	r2cDataSh    = 16 // bits [23:16] data
	r2cWriteBit  = 1 << 24
	r2cBusy      = 1 << 25

	// ANA_CONF2 block select bits (from regi2c_impl.c).
	// Each analog I2C slave has a dedicated routing bit in ANA_CONF2.
	r2cBlockMSPI = 0x63 // MSPI/MPLL
	r2cBlockSPLL = 0x66 // System PLL
	r2cBlockCPLL = 0x67 // CPU PLL

	r2cSelMSPI = 1 << 9  // REGI2C_MSPI_XTAL_MST_SEL
	r2cSelSPLL = 1 << 5  // REGI2C_PLL_SYS_MST_SEL
	r2cSelCPLL = 1 << 11 // REGI2C_PLL_CPU_MST_SEL

	// MPLL regi2c slave id and register addresses (soc/regi2c_mpll.h).
	r2cMpllSlave     = r2cBlockMSPI
	r2cMpllIRCalRstb = 1 // bit 5
	r2cMpllDivReg    = 2 // div bits[7:3], ref_div bits[2:0]
	r2cMpllDHref     = 3 // bits[5:4]
)

// lpperiBase / clk-en bit for the low-power peripheral clock gate.
const (
	lpperiBase      = 0x50120000 // DR_REG_LPPERIPH_BASE
	lpperiClkEnReg  = lpperiBase + 0x0
	ckEnLpI2CMstBit = 1 << 27 // LPPERI_CK_EN_LP_I2CMST (analog I2C master clock)
)

// regi2cTimeout bounds the busy-wait on the analog I2C master controller so a
// stuck bus returns instead of hanging the bring-up.
const regi2cMaxBusy = 100000

func r2cCtrl() *volatile.Register32 {
	return &esp.LP_I2C_ANA_MST.I2C0_CTRL
}

// regi2cEnableGate enables the LP analog I2C master clock (regi2c) used to
// reach the MPLL analog registers on the MSPI block.
func regi2cEnableGate() {
	clk := (*volatile.Register32)(unsafe.Pointer(uintptr(lpperiClkEnReg)))
	clk.Set(clk.Get() | ckEnLpI2CMstBit)
}

// regi2cSelectBlock routes the analog-I2C master to the specified block.
// The block parameter is the I2C slave ID (e.g. r2cBlockMSPI, r2cBlockCPLL).
func regi2cSelectBlock(block uint8) {
	esp.LP_I2C_ANA_MST.ANA_CONF1.Set(0)
	esp.LP_I2C_ANA_MST.ANA_CONF2.Set(0)
	var sel uint32
	switch block {
	case r2cBlockMSPI:
		sel = r2cSelMSPI
	case r2cBlockSPLL:
		sel = r2cSelSPLL
	case r2cBlockCPLL:
		sel = r2cSelCPLL
	default:
		sel = r2cSelMSPI // fallback
	}
	esp.LP_I2C_ANA_MST.ANA_CONF2.Set(sel)
}

// regi2cWrite performs a single analog-I2C-master register write to `block`
// (slave id), `reg` (address) with `data`, mirroring ESP-IDF _regi2c_impl_write
// on the P4 LP_I2C_ANA_MST I2C0 controller.
func regi2cWrite(block, reg, data uint8) {
	regi2cSelectBlock(block)
	ctrl := r2cCtrl()
	for i := 0; ctrl.Get()&r2cBusy != 0 && i < regi2cMaxBusy; i++ {
	}
	ctrl.Set(uint32(block)<<r2cSlaveIDSh |
		uint32(reg)<<r2cAddrSh |
		r2cWriteBit |
		uint32(data)<<r2cDataSh)
	for i := 0; ctrl.Get()&r2cBusy != 0 && i < regi2cMaxBusy; i++ {
	}
}

// regi2cRead performs a single analog-I2C-master register read.
func regi2cRead(block, reg uint8) uint8 {
	regi2cSelectBlock(block)
	ctrl := r2cCtrl()
	for i := 0; ctrl.Get()&r2cBusy != 0 && i < regi2cMaxBusy; i++ {
	}
	ctrl.Set(uint32(block)<<r2cSlaveIDSh | uint32(reg)<<r2cAddrSh)
	for i := 0; ctrl.Get()&r2cBusy != 0 && i < regi2cMaxBusy; i++ {
	}
	return uint8(ctrl.Get() >> r2cDataSh)
}

// EnablePSRAMMPLL powers up and programmes the MPLL used as the PSRAM clock
// source at 400MHz (the 200MHz-bus default).  Prefer EnablePSRAMMPLLFreq for
// the CrowPanel's 80MHz bus, which needs MPLL=320MHz.
func EnablePSRAMMPLL() bool {
	return EnablePSRAMMPLLFreq(400)
}

// EnablePSRAMMPLLFreq powers up and programmes the MPLL for the requested
// frequency from the 40MHz XTAL.  ESP-IDF's AP-Hex driver uses 320MHz for an
// 80MHz PSRAM bus and 400MHz for a 200MHz bus.
//
// Like ESP-IDF's esp_clk_tree_mpll_acquire, the PSRAM+MPLL analog domain must be
// powered first: on the P4 this is the internal LDO channel 2 (VO2) at 1.8V
// (CONFIG_ESP_LDO_CHAN_PSRAM_DOMAIN=2, VOLTAGE=1800mV), plus the 3.3V peripheral
// rail (LDO4) that feeds the PSRAM I/O on the CrowPanel.
func EnablePSRAMMPLLFreq(mhz uint32) bool {
	if mhz < 20 {
		return false
	}
	// Power the PSRAM/MPLL + peripheral domains.  If the LDOs fail, abort.
	if lerr := LDOConfigure(LDO2, 1800); lerr != nil {
		return false
	}
	if lerr := LDOConfigure(LDO4, 3300); lerr != nil {
		return false
	}

	regi2cEnableGate()

	// Power up the MPLL analog circuit and gate its 500MHz clock.
	esp.PMU.SetRF_PWC_MSPI_PHY_XPD(1)
	esp.LP_AON_CLKRST.SetLP_AONCLKRST_HP_CLK_CTRL_LP_AONCLKRST_HP_MPLL_500M_CLK_EN(1)

	// Program the MPLL divider over the analog I2C master (clk_ll_mpll_set_config).
	dhref := regi2cRead(r2cMpllSlave, r2cMpllDHref)
	regi2cWrite(r2cMpllSlave, r2cMpllDHref, dhref|(3<<4))
	rstb := regi2cRead(r2cMpllSlave, r2cMpllIRCalRstb)
	regi2cWrite(r2cMpllSlave, r2cMpllIRCalRstb, rstb&0xdf)
	regi2cWrite(r2cMpllSlave, r2cMpllIRCalRstb, rstb|(1<<5))
	// MPLL = XTAL*(div+1)/(ref_div+1); ref_div=1, so div = mhz/20 - 1.
	div := mhz/20 - 1
	regi2cWrite(r2cMpllSlave, r2cMpllDivReg, (uint8(div)<<3)|1)

	// Self-calibration: start, wait done (bounded), stop.
	esp.HP_SYS_CLKRST.SetANA_PLL_CTRL0_REG_MSPI_CAL_STOP(0)
	for i := 0; esp.HP_SYS_CLKRST.GetANA_PLL_CTRL0_REG_MSPI_CAL_END() == 0 && i < regi2cMaxBusy; i++ {
	}
	done := esp.HP_SYS_CLKRST.GetANA_PLL_CTRL0_REG_MSPI_CAL_END() != 0
	esp.HP_SYS_CLKRST.SetANA_PLL_CTRL0_REG_MSPI_CAL_STOP(1)
	return done
}

// EnablePSRAMClockSource selects the MPLL source and sets the 80MHz bus clock
// (MPLL 320MHz / freqdiv 4), mirroring ESP-IDF's psram_ctrlr_ll_select_clk_source
// + set_bus_clock during esp_psram_impl_enable.
func EnablePSRAMClockSource() {
	esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL00_REG_PSRAM_CLK_SRC_SEL(psramClkSrcMPLL)
	psramSetBusClock(psramMspiSystemBase, psram80Div)
	psramSetBusClock(psramMspiPeriBase, psram80Div)
}

// psramSetBusClock writes the PSRAM bus-clock divider.  MSPI2 uses
// SPI_MEM_S_SRAM_CLK (+0x50); MSPI3 uses SPI1_MEM_S_CLOCK (+0x14).  Both encode
// the same SCLKCNT_N/H/L fields.
func psramSetBusClock(base uintptr, freqdiv uint32) {
	if freqdiv == 0 {
		freqdiv = 1
	}
	var freqbits uint32
	if freqdiv == 1 {
		freqbits = 1 << 31 // SCLK_EQU_SYSCLK
	} else {
		freqbits = ((freqdiv - 1) << 16) | ((freqdiv/2 - 1) << 8) | (freqdiv - 1)
	}
	if base == psramMspiPeriBase {
		(*esp.SPI1_Type)(unsafe.Pointer(base)).SPI_MEM_CLOCK.Set(freqbits)
		return
	}
	psramMSPI().SPI_MEM_SRAM_CLK.Set(freqbits)
}

// AP-Hex PSRAM command codes and bus constants for the CrowPanel's 80 MHz
// hex-mode configuration (V1.2 sdkconfig: CONFIG_SPIRAM_MODE_HEX=y,
// CONFIG_SPIRAM_SPEED_80M=y).  ESP-IDF's AP-Hex driver selects these values
// from the "else" branch of esp_psram_impl_ap_hex.c (any speed below 200M).
const (
	apHexSyncWrite        = 0x8080 // PSDRAM sync write command
	apHexSyncRead         = 0x0000 // PSRAM sync read command
	apHexWrCmdBitlen      = 16
	apHexRdCmdBitlen      = 16
	apHexAddrBitlen       = 32
	apHexRdDummy          = 2 * (10 - 1) // 80MHz: 18
	apHexWrDummy          = 2 * (5 - 1)  // 80MHz: 8
	apHexRegRdDummy       = 2 * (5 - 1)  // 80MHz register read: 8
	apHexRdLatency        = 2            // MR0 read latency at 80MHz
	apHexWrLatency        = 2            // MR4 write latency at 80MHz
	apHexCSSetup          = 4
	apHexCSHold           = 4
	apHexCSHoldDelay      = 3
	psramPageSizeCode2048 = 3 // SPI_SMEM_ECC_CTRL page-size code for 2048B

	// MPLL is programmed to 320MHz and the PSRAM bus runs at MPLL/4 = 80MHz.
	psram80MPLLMHz = 320
	psram80BusMHz  = 80
	psram80Div     = psram80MPLLMHz / psram80BusMHz
)

// ConfigurePSRAMController programmes the PSRAM system MSPI (SPIMEM2) for the
// AP-Hex octal-DDR device: sync read/write command codes, 32-bit address depth,
// read/write dummy cycles (200MHz: rd 26, wr 12), DDR mode, octal line mode, and
// AXI access with read/write splicing.  Mirrors ESP-IDF s_config_mspi_for_psram.
func ConfigurePSRAMController() {
	m := psramMSPI()

	// Write cmd: 0x8080, 16-bit (value set is an RMW on bits [15:0]).
	m.SetSPI_MEM_CACHE_SCTRL_SPI_MEM_CACHE_SRAM_USR_WCMD(1)
	m.SetSPI_MEM_SRAM_DWR_CMD_SPI_MEM_CACHE_SRAM_USR_WR_CMD_BITLEN(apHexWrCmdBitlen - 1)
	m.SetSPI_MEM_SRAM_DWR_CMD_SPI_MEM_CACHE_SRAM_USR_WR_CMD_VALUE(apHexSyncWrite)
	// Read cmd: 0x0000, 16-bit.
	m.SetSPI_MEM_CACHE_SCTRL_SPI_MEM_CACHE_SRAM_USR_RCMD(1)
	m.SetSPI_MEM_SRAM_DRD_CMD_SPI_MEM_CACHE_SRAM_USR_RD_CMD_BITLEN(apHexRdCmdBitlen - 1)
	m.SetSPI_MEM_SRAM_DRD_CMD_SPI_MEM_CACHE_SRAM_USR_RD_CMD_VALUE(apHexSyncRead)

	// Address: 32-bit, 4-byte mode.
	m.SetSPI_MEM_CACHE_SCTRL_SPI_MEM_SRAM_ADDR_BITLEN(apHexAddrBitlen - 1)
	m.SetSPI_MEM_CACHE_SCTRL_SPI_MEM_CACHE_USR_SADDR_4BYTE(1)

	// Dummy: write 8, read 18 (80MHz), variable dummy + write dummy level ctrl.
	m.SetSPI_MEM_CACHE_SCTRL_SPI_MEM_USR_WR_SRAM_DUMMY(1)
	m.SetSPI_MEM_CACHE_SCTRL_SPI_MEM_SRAM_WDUMMY_CYCLELEN(apHexWrDummy - 1)
	m.SetSPI_MEM_CACHE_SCTRL_SPI_MEM_USR_RD_SRAM_DUMMY(1)
	m.SetSPI_MEM_CACHE_SCTRL_SPI_MEM_SRAM_RDUMMY_CYCLELEN(apHexRdDummy - 1)
	m.SetSPI_SMEM_DDR_SPI_SMEM_VAR_DUMMY(1)
	m.SetSPI_MEM_SRAM_CMD_SPI_MEM_SDUMMY_WOUT(1)

	// DDR with no data swap.
	m.SetSPI_SMEM_DDR_WDAT_SWP(0)
	m.SetSPI_SMEM_DDR_RDAT_SWP(0)
	m.SetSPI_SMEM_DDR_EN(1)

	// Octal command/address lines and x16 ("hex") data lines.  The AP-Hex
	// device transfers 16 data bits per clock, so SDIN_HEX/SDOUT_HEX must be
	// set (IDF psram_ctrlr_ll_enable_hex_data_line_mode(true)); leaving them
	// zero configures x8 data lines and makes every read/write corrupt.
	m.SetSPI_MEM_CACHE_SCTRL_SPI_MEM_SRAM_OCT(1)
	m.SetSPI_MEM_SRAM_CMD_SPI_MEM_SCMD_OCT(1)
	m.SetSPI_MEM_SRAM_CMD_SPI_MEM_SADDR_OCT(1)
	m.SetSPI_MEM_SRAM_CMD_SPI_MEM_SDOUT_OCT(1)
	m.SetSPI_MEM_SRAM_CMD_SPI_MEM_SDIN_OCT(1)
	m.SetSPI_MEM_SRAM_CMD_SPI_MEM_SDIN_HEX(1)
	m.SetSPI_MEM_SRAM_CMD_SPI_MEM_SDOUT_HEX(1)

	// AXI access + read/write splice.
	m.SetSPI_MEM_CACHE_FCTRL_SPI_MEM_AXI_REQ_EN(1)
	m.SetSPI_MEM_CACHE_FCTRL_SPI_CLOSE_AXI_INF_EN(0)
	m.SetSPI_MEM_CTRL1_SPI_MEM_AW_SPLICE_EN(1)
	m.SetSPI_MEM_CTRL1_SPI_MEM_AR_SPLICE_EN(1)
}

// ConfigurePSRAMTiming applies the pre-chip CS timing, split transaction, page
// size, bus clock (MPLL/4 = 80MHz) and DLL enable on both MSPI controllers,
// mirroring the lead-in portion of esp_psram_impl_enable.
func ConfigurePSRAMTiming() {
	m := psramMSPI()
	// CS timing (s_set_psram_cs_timing).
	m.SetSPI_SMEM_AC_SPI_SMEM_CS_SETUP(1)
	m.SetSPI_SMEM_AC_SPI_SMEM_CS_SETUP_TIME(apHexCSSetup - 1)
	m.SetSPI_SMEM_AC_SPI_SMEM_CS_HOLD(1)
	m.SetSPI_SMEM_AC_SPI_SMEM_CS_HOLD_TIME(apHexCSHold - 1)
	m.SetSPI_SMEM_AC_SPI_SMEM_CS_HOLD_DELAY(apHexCSHoldDelay - 1)
	// Split transaction + 2048B page.
	m.SetSPI_SMEM_AC_SPI_SMEM_SPLIT_TRANS_EN(1)
	m.SetSPI_SMEM_ECC_CTRL_SPI_SMEM_PAGE_SIZE(psramPageSizeCode2048)
	// Bus clock on both controllers (MPLL 320MHz / 80MHz = freqdiv 4).
	psramSetBusClock(psramMspiSystemBase, psram80Div)
	psramSetBusClock(psramMspiPeriBase, psram80Div)
	// DLL enable on both controllers (SPI_MEM_S / SPI1_MEM_S timing-cali).
	m.SetSPI_SMEM_TIMING_CALI_SPI_SMEM_DLL_TIMING_CALI(1)
	psramPeriTimingCali().SetBits(psramDLLTimingCaliBit)
}

// psramPeriTimingCali returns the MSPI3 timing-cali register.
func psramPeriTimingCali() *volatile.Register32 {
	return (*volatile.Register32)(unsafe.Pointer(uintptr(psramMspiPeriBase + 0x180)))
}

const psramDLLTimingCaliBit = 1 << 5

// AP-Hex register read/write command codes and constants.
const (
	apHexRegRead   = 0x4040
	apHexRegWrite  = 0xC0C0
	apHexCmdBitlen = 16
	apHexRefData   = 0x5a6b7c8d
)

// psramUserCommand drives a single SPI_MEM user transaction on the given
// controller (MSPI2 @0x5008E000 or MSPI3 @0x5008F000) through the ESP32-P4
// mask-ROM helpers used by ESP-IDF's AP-Hex PSRAM driver.  Using the ROM path
// gives the exact OPI-DTR line-mode setup and bit-length encodings (the ROM
// subtracts one from the address/dummy/data lengths internally).
//
// The write data word must be placed in SPI_MEM_W0 by the caller (psramDataWord);
// the read result is returned when misoBits > 0.
func psramUserCommand(base uintptr, cmd uint32, addr uint32, addrBitlen, dummy, mosiBits, misoBits uint32) uint32 {
	u := (*esp.SPI1_Type)(unsafe.Pointer(base))
	spi := psramROMSpiID(base)

	tx := u.SPI_MEM_W0.Get()
	var rx uint32
	c := psramROMCommand{
		cmd:          uint16(cmd),
		cmdBitLen:    apHexCmdBitlen,
		addr:         &addr,
		addrBitLen:   addrBitlen,
		txData:       &tx,
		txDataBitLen: mosiBits,
		rxData:       &rx,
		rxDataBitLen: misoBits,
		dummyBitLen:  dummy,
	}

	psramROMSetOpMode(spi, psramROMOPIDTRMode)
	psramROMCmdConfig(spi, &c)
	if misoBits > 0 {
		psramROMCmdStart(spi, &rx, misoBits/8, 1<<1, 0)
	} else {
		psramROMCmdStart(spi, nil, 0, 1<<1, 0)
	}
	return rx
}

// psramReg returns a raw 32-bit register pointer at `base + off`.
func psramReg(base uintptr, off uintptr) *volatile.Register32 {
	return (*volatile.Register32)(unsafe.Pointer(base + off))
}

// InitAPHexPSRAM performs the AP-Hex PSRAM chip init: programs MR0/MR4/MR8 mode
// registers (fixed read/write latency, drive strength, burst length, x16 mode)
// over the peripheral controller (MSPI3) and confirms the part is connected by
// a write/read-back of a reference pattern.  Mirrors ESP-IDF
// s_init_psram_mode_reg + s_check_psram_connected.
func InitAPHexPSRAM(mspi3Base uintptr) bool {
	// MR0: lt=1 (fixed latency), read_latency=2, drive_str=0 @80MHz.
	// mr0 bits: drive_str[1:0], read_latency[4:2], lt[5], rsvd6[6], tso[7].
	// Read MR0+MR1 (addr 0x0): low byte = MR0, high byte = MR1.
	v01 := psramUserCommand(mspi3Base, apHexRegRead, 0x0, 32, apHexRegRdDummy, 0, 16)
	mr0 := byte(v01 & 0xff)
	mr1 := byte(v01 >> 8)
	mr0 = (mr0 & 0xc2) | (1 << 5) | uint8(apHexRdLatency<<2)
	psramUserCommand16(mspi3Base, apHexRegWrite, 0x0, uint16(mr0)|uint16(mr1)<<8)

	// MR4 (addr 0x4): low byte = MR4, high byte = MR8; set wr_latency[7:5]=2.
	v48 := psramUserCommand(mspi3Base, apHexRegRead, 0x4, 32, apHexRegRdDummy, 0, 16)
	mr4 := byte(v48 & 0xff)
	mr8Old := byte(v48 >> 8)
	mr4 = (mr4 & 0x1f) | uint8(apHexWrLatency<<5)
	psramUserCommand16(mspi3Base, apHexRegWrite, 0x4, uint16(mr4)|uint16(mr8Old)<<8)

	// MR8 (addr 0x8, 8-bit read): set bl=3, bt=0, rbx=1, x16=1.
	mr8 := byte(psramUserCommand(mspi3Base, apHexRegRead, 0x8, 32, apHexRegRdDummy, 0, 8) & 0xff)
	// mr8 bits: bl[1:0], bt[2], rbx[3], x16[6].
	mr8 = (mr8 & 0x68) | 3 | (1 << 3) | (1 << 6)
	psramUserCommand16(mspi3Base, apHexRegWrite, 0x8, uint16(mr8))

	// Connected check: write a reference pattern, read it back (round trip).
	psramDataWord(mspi3Base, apHexRefData)
	psramUserCommand(mspi3Base, apHexSyncWrite, 0x0, 32, apHexWrDummy, 32, 0)
	rd := psramUserCommand(mspi3Base, apHexSyncRead, 0x0, 32, apHexRdDummy, 0, 32)
	return rd == apHexRefData
}

// psramUserCommand16 is a convenience wrapper for a 16-bit data-phase write.
func psramUserCommand16(base uintptr, cmd, addr uint32, data uint16) {
	psramDataWord(base, uint32(data))
	psramUserCommand(base, cmd, addr, 32, 0, 16, 0)
}

// psramDataWord writes a 32-bit data word into the SPI_MEM W0 FIFO of the block.
func psramDataWord(base uintptr, w uint32) {
	u := (*esp.SPI1_Type)(unsafe.Pointer(base))
	u.SPI_MEM_W0.Set(w)
}

// PSRAMProbe exercises the PSRAM through the SYSTEM controller (MSPI2) with a
// bounded user-command write/read round-trip and returns the read-back value.
// Unlike a raw CPU access (which stalls forever on a non-responsive part), this
// is non-hanging: it returns 0 on failure so the bring-up can report instead of
// wedging.  Used for on-target bring-up checks before the CPU/MMU path is
// trusted, and it maps page 0 so the CPU path is also primed.
func PSRAMProbe() uint32 {
	MapPSRAM(psramVAddrBase, 0, 1)
	psramDataWord(psramMspiSystemBase, apHexRefData)
	psramUserCommand(psramMspiSystemBase, apHexSyncWrite, 0x0, 32, apHexWrDummy, 32, 0)
	rd := psramUserCommand(psramMspiSystemBase, apHexSyncRead, 0x0, 32, apHexRdDummy, 0, 32)
	if rd != apHexRefData {
		return 0
	}
	return rd
}

// PSRAMBase is the CPU virtual address where InitPSRAM80MHz maps PSRAM0.
const PSRAMBase = psramVAddrBase

// PSRAMFramebufferBytes is the size of one full 1024x600 RGB565 frame.
const PSRAMFramebufferBytes = 1024 * 600 * 2

// psramFramebufferPages is the number of 64KB MMU pages needed for one frame.
const psramFramebufferPages = (PSRAMFramebufferBytes + psramPageSize - 1) / psramPageSize

// InitPSRAM80MHz performs the full AP-Hex PSRAM bring-up used by the working
// CrowPanel V1.2 example: MPLL=320MHz, PSRAM bus=80MHz, x16 hex mode,
// 2048-byte pages, DLL enabled, then chip mode-register init over MSPI3.
// On success it maps enough pages at PSRAMBase for one full 1024x600 RGB565
// framebuffer and validates a 1MB CPU write/read-back through the cache.
//
// The function is deliberately non-fatal: callers should print an error and
// stay in the small on-chip framebuffer path if it returns false.
func InitPSRAM80MHz() bool {
	if !EnablePSRAMMPLLFreq(psram80MPLLMHz) {
		return false
	}
	EnablePSRAMModuleClock()
	ConfigPSRAMPinIO(2)
	ConfigurePSRAMTiming()

	// Configure the chip itself through the peripheral controller (MSPI3).
	if !InitAPHexPSRAM(psramMspiPeriBase) {
		return false
	}

	// Now configure the system controller (MSPI2) for cache/AXI access and map
	// the physical pages into the CPU data address space.
	ConfigurePSRAMController()
	MapPSRAM(psramVAddrBase, 0, psramFramebufferPages)

	return PSRAMValidate(psramVAddrBase, PSRAMFramebufferBytes)
}

// PSRAMValidate writes a position-dependent pattern through the cacheable PSRAM
// mapping, writes the dirty lines back, invalidates them, and reads the pattern
// again from the device.  This distinguishes a real PSRAM path from a cache hit
// on stale data.  It is bounded to one framebuffer (1.2MB).
func PSRAMValidate(addr uintptr, size uintptr) bool {
	const patternSeed = 0xA5C30000
	words := int(size / 4)

	for i := 0; i < words; i++ {
		v := uint32(patternSeed) ^ uint32(i)*0x9E3779B1
		(*volatile.Register32)(unsafe.Pointer(addr + uintptr(i)*4)).Set(v)
	}
	CacheWriteBack(addr, size)
	CacheInvalidate(addr, size)

	for i := 0; i < words; i++ {
		want := uint32(patternSeed) ^ uint32(i)*0x9E3779B1
		got := (*volatile.Register32)(unsafe.Pointer(addr + uintptr(i)*4)).Get()
		if got != want {
			return false
		}
	}
	return true
}

// --- IOMUX PSRAM pin IO + timing (Step 2c) --------------------------------
//
// The PSRAM octal-DDR bus uses dedicated IOMUX MSPI pins.  For signal
// integrity at 200MHz the pad drive strength and the DQS input/output must be
// configured, and the DQS phase / per-pin delaylines aligned (the DQS-based
// timing tuning).  These registers live in the IOMUX_MSPI_PIN block.

// IOMUX MSPI PSRAM pin register block base and field layout.
const (
	iomuxMspiPinBase = 0x500E1200 // DR_REG_HPPERIPH1_BASE(0x500C0000)+0x21200
	psramDrvSh       = 12         // PSRAM_*_DRV[2] bit offset in the data/control pins
	psramDqsDrvSh    = 15         // PSRAM_DQS_*_DRV[2] bit offset
	psramDqsXPD      = 1 << 0     // PSRAM_DQS_*_XPD (power up DQS pad)
	psramDqsPhaseSh  = 1          // PSRAM_DQS_*_PHASE
)

// PSRAM data/control pin register offsets within the IOMUX MSPI pin block.
var psramPinOffsets = [20]uint32{
	0x1c, 0x20, 0x24, 0x28, 0x2c, 0x30, 0x34, 0x38, // D,Q,WP,HD,DQ4..DQ7
	0x3c,       // DQS_0
	0x40, 0x44, // CK, CS
	0x48, 0x4c, 0x50, 0x54, 0x58, 0x5c, 0x60, 0x64, // DQ8..DQ15
	0x68, // DQS_1
}

func psramPinReg(off uint32) *volatile.Register32 {
	return (*volatile.Register32)(unsafe.Pointer(uintptr(iomuxMspiPinBase + off)))
}

// ConfigPSRAMPinIO sets every PSRAM IOMUX pad drive strength (drv, matches
// ESP-IDF mspi_timing_ll_pin_drv_set(drv=2)) and powers up the DQS pads
// (mspi_timing_ll_enable_dqs(true)).
func ConfigPSRAMPinIO(drv uint8) {
	for _, off := range psramPinOffsets {
		reg := psramPinReg(off)
		if off == 0x3c || off == 0x68 { // DQS_0 / DQS_1: drive at bit 15 + XPD
			reg.ReplaceBits(uint32(drv), 0x3<<psramDqsDrvSh, psramDqsDrvSh)
			reg.SetBits(psramDqsXPD)
		} else {
			reg.ReplaceBits(uint32(drv), 0x3<<psramDrvSh, psramDrvSh)
		}
	}
}

// SetPSRAMDQSPhase sets the DQS phase field of both DQS pads (used by timing
// tuning).  phase maps 0..3 to {67.5,78.75,90,101.25} degrees.
func SetPSRAMDQSPhase(phase uint8) {
	psramPinReg(0x3c).ReplaceBits(uint32(phase), 0x3<<psramDqsPhaseSh, psramDqsPhaseSh)
	psramPinReg(0x68).ReplaceBits(uint32(phase), 0x3<<psramDqsPhaseSh, psramDqsPhaseSh)
}

// SetPSRAMDelayline sets the per-pin delayline for the given pin index
// (0..MSPI_LL_PIN_MAX, enum as in mspi_ll.h: D=0,Q=1,WP,HD,D4..D7=7,DQS0=8,
// CLK=9,CS=10,D8..D15=17,DQS1=18).  DQS pins use the dqs delay_90/270 fields;
// others use the per-pin delayline field.
func SetPSRAMDelayline(pin uint8, delayline uint8) {
	dlr := psramDelaylineReg(pin)
	if dlr == nil {
		return
	}
	dlr.ReplaceBits(uint32(delayline), 0xFF, 0)
}

// psramDelaylineReg returns the register+field-address for a pin's delayline,
// mapping the MSPI_LL pin enum to the IOMUX psram_pin_group layout.
func psramDelaylineReg(pin uint8) *volatile.Register32 {
	// DQS0 and DQS1 use a dedicated pair of delay-90/270 fields in separate
	// registers (dqs0/dqs1).  These are not on the same 4-byte stride as the
	// per-pin delayline fields, so we return a placeholder on the DQS pin regs
	// (the delay-90 field is bit-aligned there) -- see PSRAM pin group B/C.
	if pin == 8 { // DQS0
		return psramPinReg(0x3c)
	}
	if pin == 18 { // DQS1
		return psramPinReg(0x68)
	}
	// data/control pins: D(0)..D7(7) -> offsets 0x1c..0x38; CLK(9)/CS(10) and
	// D8..D15 -> the low nibble delayline field in pin_group0/pin_group1 regs.
	var off uint32
	switch {
	case pin <= 7:
		off = 0x1c + uint32(pin)*4
	case pin == 9:
		off = 0x40
	case pin == 10:
		off = 0x44
	default: // D8..D15 (11..18 in enum after CLK/CS): 0x48 + (pin-11)*4
		off = 0x48 + uint32(pin-11)*4
	}
	return psramPinReg(off)
}

// TunePSRAM applies a deterministic timing baseline (Step 2c-2 first cut):
// DQS phase and per-pin delaylines are zeroed, which is the low-speed/starting
// point the DQS sweep tunes from; a full runtime sweep can be layered on later.
func TunePSRAM() {
	SetPSRAMDQSPhase(0)
	for p := uint8(0); p < 19; p++ {
		SetPSRAMDelayline(p, 0)
	}
}
