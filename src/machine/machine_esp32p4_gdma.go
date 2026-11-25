//go:build esp32p4

// ESP32-P4 DW-GDMA (DesignWare AXI DMA) driver.
//
// The DSI bridge is fed pixel data by the DW-GDMA controller at 0x50081000.
// This driver configures a controller, binds one of its four channels to a
// memory→peripheral transfer (DMA as the flow controller) and manages a
// single-item link list that describes the whole frame.
//
// Channel register layout (DW 32-bit mode):
//
//	channel N base = 0x50081000 + 0x100 + N*0x100
//	  SAR0    0x00  source address
//	  DAR0    0x08  destination address
//	  BLOCK_TS 0x10 (transfer items) - 1
//	  CTL0     0x18  control word low
//	  CTL1     0x1C  control word high (burst lengths + LLI markers)
//	  CFG0     0x20  multi-block transfer types
//	  CFG1     0x24  flow control / handshake / peripheral IDs / priority
//	  LLP0     0x28  link-list head (addr >> 6) | master select
//
// Global registers (0x50081000):
//
//	CFG0     0x10  DMAC_EN
//	CHEN0    0x18  per-channel enable (EN + EN_WE)
//
// The control-word encodings are computed by machine/internal/esp32p4dsi so
// the exact bit patterns are covered by host unit tests.
//
// Reference: ESP-IDF components/esp_hal_dma/esp32p4/include/hal/dw_gdma_ll.h
// and components/esp_lcd/dsi/esp_lcd_panel_dpi.c.
package machine

import (
	"device/esp"
	"device/riscv"
	"errors"
	"machine/internal/esp32p4dsi"
	"runtime/volatile"
	"unsafe"
)

// Channel register offsets relative to the channel base (64-bit transfer path).
const (
	gdmaRegSAR0          = 0x00
	gdmaRegSAR1          = 0x04
	gdmaRegDAR0          = 0x08
	gdmaRegDAR1          = 0x0C
	gdmaRegBlockTS       = 0x10
	gdmaRegCTL0          = 0x18
	gdmaRegCTL1          = 0x1C
	gdmaRegCFG0          = 0x20
	gdmaRegCFG1          = 0x24
	gdmaRegLLP0          = 0x28
	gdmaRegLLP1          = 0x2C
	gdmaRegStatus0       = 0x30
	gdmaRegSWHSSRC0      = 0x38 // CH_SWHSSRCREG0 — software handshake source
	gdmaRegSWHSDST0      = 0x40 // CH_SWHSDSTREG0 — software handshake destination
	gdmaRegINTSTATUS_EN0 = 0x80 // CH_INTSTATUS_ENABLE0 — gates status bit generation
	gdmaRegINTSTATUS0    = 0x88 // CH_INTSTATUS0 — read-only interrupt status
	gdmaRegINTSIGNAL_EN0 = 0x90 // CH_INTSIGNAL_ENABLE0 — gates interrupt signal to CPU
	gdmaRegINTCLEAR0     = 0x98 // CH_INTCLEARREG0 (0x198 from DMAC base for CH1)
)

// DW-GDMA channel enable / suspend bits in CHEN0.
const (
	gdmaChEnLow  = 0x101       // EN_WE + EN for channels 0..7 (2 bits each)
	gdmaChEnHigh = 0x101 << 16 // EN_WE + EN for channels 8..11
)

// ErrGDMAChannel is returned when an invalid channel index is requested.
var ErrGDMAChannel = errors.New("machine: DW-GDMA channel must be 0..3")

// dmaLLIItem is the hardware layout of a 64-byte DW-GDMA link-list item.
// Fields are written with the full word values produced by machine/internal/esp32p4dsi.
type dmaLLIItem struct {
	SarLo, SarHi       uint32 // source address (low/high)
	DarLo, DarHi       uint32 // destination address (low/high)
	BlockTS, _reserved uint32
	LLPLo, LLPHi       uint32 // next link-list item (addr>>6) | LMS  / high
	CtlLo, CtlHi       uint32 // control word low/high
	SStat, DStat       uint32 // periph status (writeback, unused)
	StatLo, StatHi     uint32 // channel status (written back here)
	_r2, _r3           uint32
}

// gdmaMaxLLI is the number of linked-list items in the static pool.  The
// display path normally uses one item per frame; a long chain (up to 256
// frames, ~4.5 s at 55 fps) is used to test continuous DMA without any
// full-transfer-done re-arm.
const gdmaMaxLLI = 256

// gdmaLLIPool holds gdmaMaxLLI 64-byte-aligned link-list items in a
// package global (GC-stable) buffer.
var gdmaLLIPool struct {
	buf [esp32p4dsi.GdmaLLISize*gdmaMaxLLI + 63]byte
}

// gdmaLLIAlloc returns a pointer to the first of gdmaMaxLLI 64-byte-aligned
// link-list items.
func gdmaLLIAlloc() *dmaLLIItem {
	base := uintptr(unsafe.Pointer(&gdmaLLIPool.buf[0]))
	aligned := (base + 63) &^ 63
	// IDF's dw_gdma_new_link_list does an esp_cache_msync(C2M|INVALIDATE) at
	// allocation, before any non-cacheable writes, so stale dirty lines from
	// the BSS zero-fill cannot later be written back over the descriptors.
	CacheWriteBack(aligned, uintptr(len(gdmaLLIPool.buf)))
	CacheInvalidateL1(aligned, uintptr(len(gdmaLLIPool.buf)))
	return (*dmaLLIItem)(unsafe.Pointer(aligned))
}

// gdmaLLIAt returns the Nth link-list item (0-indexed) from the base item.
func gdmaLLIAt(base *dmaLLIItem, n int) *dmaLLIItem {
	return (*dmaLLIItem)(unsafe.Add(unsafe.Pointer(base), uintptr(n)*esp32p4dsi.GdmaLLISize))
}

// DSI bridge memory FIFO that the DW-GDMA writes pixels into.
const dsiBridgeMemBase = 0x50105000

// BridgeFIFOWrite writes one 64-bit word directly into the DSI bridge raw
// buffer FIFO.  This bypasses the DW-GDMA and is a bring-up diagnostic: it
// separates the DMA write path from the bridge's own read/output path.  The
// bridge normally receives these words from the DMA at the same address.
func BridgeFIFOWrite(word uint64) {
	volatile.StoreUint64((*uint64)(unsafe.Pointer(uintptr(dsiBridgeMemBase))), word)
}

// gdmaHW returns a pointer to the DW-GDMA register block, constructed from
// a constant address.  This avoids reading the esp.DMA package-level var,
// which is a .data global whose L1 D-cache line can be evicted and lost
// during large memory fills on ESP32-P4 HP SRAM (display76 root cause).
func gdmaHW() *esp.DMAC_Type {
	return (*esp.DMAC_Type)(unsafe.Pointer(uintptr(0x50081000)))
}

// gdmaNonCacheableOff is SOC_NON_CACHEABLE_OFFSET_SRAM on the ESP32-P4: adding
// it to an L2MEM address maps to the non-cacheable alias of the same physical
// SRAM (0x4FF0... -> 0x8FF0...).  Writes through this alias bypass the L1
// D-cache so the DW-GDMA reads fresh data.
const gdmaNonCacheableOff = 0x40000000

// GDMAChannel is one DW-GDMA channel bound to the DSI display path.
type GDMAChannel struct {
	dmac     *esp.DMAC_Type
	num      uint32 // 0..3
	lli      *dmaLLIItem
	lliCount int  // number of items in the linked list chain
	m2m      bool // channel temporarily reconfigured for a memory-to-memory test
}

// gdmaChannel returns the volatile register at the given channel-relative
// offset.
func (c *GDMAChannel) reg(off uint32) *volatile.Register32 {
	base := unsafe.Pointer(&c.dmac.CH1_SAR0) // channel 0 starts at 0x100
	return (*volatile.Register32)(unsafe.Add(base, uintptr(c.num)*0x100+uintptr(off)))
}

// gdmaClocksDone tracks whether gdmaGateClocks has already been called (e.g.
// by an early InitGDMAClocks) so the reset pulse is not repeated when
// NewGDMAChannel runs later.
var gdmaClocksDone bool

// InitGDMAClocks performs the one-time DW-GDMA clock enable, peripheral reset
// and software reset.  Calling this early — BEFORE configuring the DSI bridge
// — avoids the HP-AXI bus arbitration glitch that the RST_EN_GDMA pulse in
// HP_RST_EN0 causes.  ESP-IDF creates the DMA channel (and thus gates clocks
// + resets GDMA) BEFORE writing bridge registers; our earlier init order put
// the GDMA reset AFTER bridge configuration, which could corrupt the bridge's
// DMA handshake state machine.
//
// After this call, NewGDMAChannel will skip the redundant gdmaGateClocks().
func InitGDMAClocks() {
	if gdmaClocksDone {
		return
	}
	gdmaGateClocks()
	EnableGDMAEngine()
}

// NewGDMAChannel binds and configures a DW-GDMA channel for the DSI display
// path: memory→DSI-bridge, DMA as the flow controller, hardware handshake,
// 64-bit transfers, 512-item source bursts and 256-item destination bursts.
//
// The bus clock is enabled and the controller reset on the first channel.
func NewGDMAChannel(num uint8) (*GDMAChannel, error) {
	if num > 3 {
		return nil, ErrGDMAChannel
	}
	gdmaGateClocks()

	c := &GDMAChannel{dmac: gdmaHW(), num: uint32(num)}
	// d221: Disable the channel before reconfiguring, in case it's still
	// enabled from a previous phase.  This replaces the full GDMA software
	// reset that was previously done in gdmaGateClocks() — which broke the
	// bridge handshake (d205-d220).
	c.Disarm()
	ctl0 := esp32p4dsi.GDMAChannelCtl0(
		esp32p4dsi.GDMMasterMem, esp32p4dsi.GDMMasterDSI, // master ports
		esp32p4dsi.GdmaBurstModeIncr, esp32p4dsi.GdmaBurstModeFixed,
		esp32p4dsi.GdmaTransWidth64, esp32p4dsi.GdmaTransWidth64,
		esp32p4dsi.GdmaBurstItems512, esp32p4dsi.GdmaBurstItems256,
	)
	ctl1 := esp32p4dsi.GDMAChannelCtl1(15, 15, false, true, false) // IDF 0x400F87C0
	// Link-list (LLI) single-block transfer.  The LLI is written through its
	// non-cacheable alias (see SetupLinkList) so the DW-GDMA reads fresh
	// parameters, and the block markers (valid/last) are set so the transfer
	// completes against the DSI bridge handshake.
	//
	// IOC_BLKTFR (ctl1 bit 26, "interrupt on block transfer complete") is
	// deliberately CLEAR.  The live ESP-IDF reference reads ctl1=0x400F87C0
	// (bit 26 = 0); every TinyGo build through d309 read 0x440F87C0
	// (bit 26 = 1).  With IOC_BLKTFR set the channel raises a block-done
	// event at every full-frame DMA block boundary, and the DSI bridge
	// reports exactly one FIFO underrun at that same boundary — the
	// per-frame underrun that has kept the panel in reserve.  VALID/LAST
	// still generate the full-transfer-done event used by the re-arm ISR.
	cfg0 := esp32p4dsi.GDMAChannelCfg0(esp32p4dsi.GdmaBlockTransferList, esp32p4dsi.GdmaBlockTransferList)
	cfg1 := esp32p4dsi.GDMAChannelCfg1(
		esp32p4dsi.GdmaTTFCM2PDmac,
		// src=HW, dst=HW, matching the live ESP-IDF reference dump:
		//   DMA ch0 cfg1=0x0A020001  (bit3 hs_sel_src=0)
		// The reference reads PSRAM through the same master port, so the
		// d224 "HW source handshake causes DECERR" conclusion does not hold
		// for the current PSRAM/full-frame path.
		esp32p4dsi.GdmaHandshakeHW, esp32p4dsi.GdmaHandshakeHW, // src=HW, dst=HW (IDF reference)
		esp32p4dsi.GdmaHandshakePeriphDSI, esp32p4dsi.GdmaHandshakePeriphDSI, // src/dst peripheral IDs (0 = DSI)
		1,        // priority
		5-1, 2-1, // outstanding-request limits (stored minus one)
	)
	c.reg(gdmaRegCTL0).Set(ctl0)
	c.reg(gdmaRegCTL1).Set(ctl1)
	c.reg(gdmaRegCFG0).Set(cfg0)
	c.reg(gdmaRegCFG1).Set(cfg1)
	// Enable all interrupt status generation and signal delivery.  ESP-IDF
	// sets both to UINT32_MAX; without these the DW-GDMA's handshake state
	// machine may not properly cycle between bursts within a block transfer.
	c.reg(gdmaRegINTSTATUS_EN0).Set(0xFFFFFFFF)
	c.reg(gdmaRegINTSIGNAL_EN0).Set(0xFFFFFFFF)
	return c, nil
}

// gdmaGateClocks enables the DW-GDMA CPU/sys clocks, releases the block from
// peripheral reset and resets the controller software-wise.  This mirrors the
// ESP-IDF dw_gdma_ll_enable_bus_clock + dw_gdma_ll_reset_register +
// dw_gdma_ll_reset sequence: without reg_rst_en_gdma being released, reading
// the DW-GDMA registers (e.g. RESET0) faults because the block is held in
// reset.
func gdmaGateClocks() {
	firstTime := !gdmaClocksDone
	gdmaClocksDone = true
	if !firstTime {
		// d221: Do NOT re-reset the entire GDMA controller on subsequent
		// calls.  ESP-IDF only resets GDMA once (during group creation);
		// subsequent channel allocations just disable and reconfigure.
		// The full software reset (RESET0 at 0x50081058) destroys the
		// handshake interface state machine, disconnecting the bridge's
		// dma_req signal — this was the root cause of the persistent
		// HW-handshake failure in d205-d220.
		// Channel disable (Disarm) is now done in NewGDMAChannel instead.
		return
	}
	// Force the GDMA internal memory clock on so register accesses don't
	// fault.  Without this the GDMA's internal clock domain may be gated,
	// causing bus errors on any register read (display73/74 crash at
	// 0x50081038).
	esp.HP_SYS.SetPERI_MEM_CLK_FORCE_ON_GDMA_MEM_CLK_FORCE_ON(1)
	esp.HP_SYS_CLKRST.SetSOC_CLK_CTRL0_REG_GDMA_CPU_CLK_EN(1)
	esp.HP_SYS_CLKRST.SetSOC_CLK_CTRL1_REG_GDMA_SYS_CLK_EN(1)
	// Release the DW-GDMA from peripheral reset (assert then deassert).
	esp.HP_SYS_CLKRST.SetHP_RST_EN0_REG_RST_EN_GDMA(1)
	esp.HP_SYS_CLKRST.SetHP_RST_EN0_REG_RST_EN_GDMA(0)
	// display73: fence + nops give the clock tree time to settle after
	// enable before the first DMA register access.
	riscv.Asm("fence iorw, iorw")
	for i := 0; i < 32; i++ {
		riscv.Asm("nop")
	}
	println("GDMA: clk cpu=", esp.HP_SYS_CLKRST.GetSOC_CLK_CTRL0_REG_GDMA_CPU_CLK_EN(),
		"sys=", esp.HP_SYS_CLKRST.GetSOC_CLK_CTRL1_REG_GDMA_SYS_CLK_EN(),
		"rst=", esp.HP_SYS_CLKRST.GetHP_RST_EN0_REG_RST_EN_GDMA(),
		"mem=", esp.HP_SYS.GetPERI_MEM_CLK_FORCE_ON_GDMA_MEM_CLK_FORCE_ON())
	// display80: COMMONREG_INTCLEAR0 at 0x38 is write-only — reading it
	// causes a bus decode error (load access fault).  SetRESET0_DMAC_RST
	// does a read-modify-write that somehow hits 0x38 instead of 0x58.
	// Use a raw write-1 to the RESET0 register at the known-good address,
	// then poll via raw read until the hardware clears the bit.
	resetReg := (*uint32)(unsafe.Pointer(uintptr(0x50081058)))
	println("GDMA: resetting via raw 0x50081058")
	volatile.StoreUint32(resetReg, 1)
	for volatile.LoadUint32(resetReg) != 0 {
		riscv.Asm("nop")
	}
	println("GDMA: reset done")
}

// EnableController turns on the DW-GDMA engine (CFG0.DMAC_EN).
func EnableGDMAEngine() {
	gdmaHW().SetCFG0_DMAC_EN(1)
	gdmaHW().SetCFG0_INT_EN(1)
}

// DisableController turns off the DW-GDMA engine.
func DisableGDMAEngine() {
	gdmaHW().SetCFG0_DMAC_EN(0)
}

// SetupLinkList programmes the channel source/destination addresses, the block
// size (in 64-bit words) and points the LLP at a fully-configured link-list
// item.  The item is programmed with:
//
//	src = srcAddr (increment), dst = dsiBridgeMemBase (fixed)
//	block size = words, valid + last markers
//
// SetupLinkList programmes a single link-list item that carries the whole
// block, and points the channel at it.  The item is written through its
// non-cacheable alias (item address + SOC_NON_CACHEABLE_OFFSET_SRAM) so the
// DW-GDMA (which reads L2MEM SRAM directly, bypassing the CPU L1 D-cache)
// always fetches fresh transfer parameters instead of stale cached copies.
// The channel's LLP points at the (cached) item address, as ESP-IDF does.
func (c *GDMAChannel) SetupLinkList(lli *dmaLLIItem, srcAddr, dstAddr uintptr, words uint32) {
	// The link-list CtlLo carries the master ports.  In the display path the
	// destination is the DSI bridge (master 0); when ReconfigureM2M has
	// temporarily put the channel in memory-to-memory mode, both masters are
	// memory so the LLI agrees with the channel CFG registers.
	var dstMaster uint32 = esp32p4dsi.GDMMasterDSI
	var dstBurstMode uint32 = esp32p4dsi.GdmaBurstModeFixed
	var dstBurstItems uint32 = esp32p4dsi.GdmaBurstItems256
	if c.m2m {
		dstMaster = esp32p4dsi.GDMMasterMem
		dstBurstMode = esp32p4dsi.GdmaBurstModeIncr
		dstBurstItems = esp32p4dsi.GdmaBurstItems512
	}
	ctl0 := esp32p4dsi.GDMAChannelCtl0(
		esp32p4dsi.GDMMasterMem, dstMaster,
		esp32p4dsi.GdmaBurstModeIncr, dstBurstMode,
		esp32p4dsi.GdmaTransWidth64, esp32p4dsi.GdmaTransWidth64,
		esp32p4dsi.GdmaBurstItems512, dstBurstItems,
	)
	// Link-list items must carry SHADOWREG_OR_LLI_VALID | LLI_LAST for the
	// DW-GDMA to execute them.  ESP-IDF sets both markers on every LLI item
	// (dpi_panel_init: dw_gdma_lli_set_block_markers(is_valid=true,
	// is_last=true)).  Without VALID the channel enables but moves zero
	// words, which is exactly what the d425/d426 list-mode M2M test showed.
	ctl1 := esp32p4dsi.GDMAChannelCtl1(15, 15, false, true, true)

	// Write the LLI through the non-cacheable alias so the DMA sees fresh data.
	lliBase := uintptr(unsafe.Pointer(lli))
	lliNC := unsafe.Pointer(lliBase + gdmaNonCacheableOff)
	nc := (*dmaLLIItem)(lliNC)
	nc.SarLo = uint32(srcAddr)
	nc.SarHi = 0
	nc.DarLo = uint32(dstAddr)
	nc.DarHi = 0
	nc.BlockTS = words - 1
	// End of list: null next-item pointer (loc0 = 0) but with the link-list
	// master port set to MEMORY.  ESP-IDF sets lms = MEMORY on every LLI item
	// (dw_gdma.c dw_gdma_new_link_list); a zero lms leaves the DSI peripheral
	// port selected and the DW-GDMA stalls fetching/validating the list.
	nc.LLPLo = esp32p4dsi.GDMMasterMem
	nc.LLPHi = 0
	nc.CtlLo = ctl0
	nc.CtlHi = ctl1
	nc.SStat, nc.DStat, nc.StatLo, nc.StatHi = 0, 0, 0, 0
	riscv.Asm("fence")

	// Point the channel at the LLI.  LLP0.loc0 occupies register bits [31:6]
	// and holds the item address >> 6; writing the register therefore needs the
	// low 6 bits cleared (not a pre-shifted address).  lms (bit 0) selects the
	// master port the DMA uses to fetch the list item from memory.
	c.lli = lli
	c.reg(gdmaRegLLP0).Set(uint32(lliBase)&^0x3F | esp32p4dsi.GDMMasterMem)
	c.reg(gdmaRegLLP1).Set(0)
}

// ReconfigureM2M reprograms the channel registers for a memory-to-memory LLI
// block (both master ports = memory, DMA flow control, software handshake)
// so a diagnostic memcpy can be driven without the DSI-destination binding.
// The link-list items are rebuilt with matching masters by the next
// SetupLinkList call.  Use ReconfigureM2P to restore the display path.
func (c *GDMAChannel) ReconfigureM2M() {
	c.m2m = true
	ctl0 := esp32p4dsi.GDMAChannelCtl0(
		esp32p4dsi.GDMMasterMem, esp32p4dsi.GDMMasterMem,
		esp32p4dsi.GdmaBurstModeIncr, esp32p4dsi.GdmaBurstModeIncr,
		esp32p4dsi.GdmaTransWidth64, esp32p4dsi.GdmaTransWidth64,
		esp32p4dsi.GdmaBurstItems512, esp32p4dsi.GdmaBurstItems512,
	)
	ctl1 := esp32p4dsi.GDMAChannelCtl1(15, 15, false, true, false) // IDF 0x400F87C0
	cfg0 := esp32p4dsi.GDMAChannelCfg0(esp32p4dsi.GdmaBlockTransferList, esp32p4dsi.GdmaBlockTransferList)
	// M2M flow (tt_fc = DW_GDMA_LL_FLOW_M2M_DMAC = 0) with hardware
	// handshake, matching ESP-IDF's own DW_GDMA M2M test
	// (components/esp_hw_support/test_apps/dma/main/test_dw_gdma.c), which
	// leaves handshake_type at its zero default.  Selecting software
	// handshake here makes the channel wait forever for a request: the
	// channel enables (CHEN0=1) but moves zero words (STATUS0=0).
	cfg1 := esp32p4dsi.GDMAChannelCfg1(
		0, esp32p4dsi.GdmaHandshakeHW, esp32p4dsi.GdmaHandshakeHW,
		0, 0, 1, 5-1, 5-1,
	)
	c.reg(gdmaRegCTL0).Set(ctl0)
	c.reg(gdmaRegCTL1).Set(ctl1)
	c.reg(gdmaRegCFG0).Set(cfg0)
	c.reg(gdmaRegCFG1).Set(cfg1)
}

// ReconfigureM2P restores the channel registers to the DSI display path:
// memory→peripheral (DSI bridge) with hardware handshake.  Called after a
// memory-to-memory diagnostic transfer.
func (c *GDMAChannel) ReconfigureM2P() {
	c.m2m = false
	ctl0 := esp32p4dsi.GDMAChannelCtl0(
		esp32p4dsi.GDMMasterMem, esp32p4dsi.GDMMasterDSI,
		esp32p4dsi.GdmaBurstModeIncr, esp32p4dsi.GdmaBurstModeFixed,
		esp32p4dsi.GdmaTransWidth64, esp32p4dsi.GdmaTransWidth64,
		esp32p4dsi.GdmaBurstItems512, esp32p4dsi.GdmaBurstItems256,
	)
	ctl1 := esp32p4dsi.GDMAChannelCtl1(15, 15, false, true, false) // IDF 0x400F87C0
	cfg0 := esp32p4dsi.GDMAChannelCfg0(esp32p4dsi.GdmaBlockTransferList, esp32p4dsi.GdmaBlockTransferList)
	cfg1 := esp32p4dsi.GDMAChannelCfg1(
		esp32p4dsi.GdmaTTFCM2PDmac,
		esp32p4dsi.GdmaHandshakeHW, esp32p4dsi.GdmaHandshakeHW, // src=HW, dst=HW (IDF reference)
		esp32p4dsi.GdmaHandshakePeriphDSI, esp32p4dsi.GdmaHandshakePeriphDSI,
		1, 5-1, 2-1,
	)
	c.reg(gdmaRegCTL0).Set(ctl0)
	c.reg(gdmaRegCTL1).Set(ctl1)
	c.reg(gdmaRegCFG0).Set(cfg0)
	c.reg(gdmaRegCFG1).Set(cfg1)
}

// ConfigureContiguous sets up the channel for a contiguous (no LLI) block
// transfer.  SAR, DAR and BLOCK_TS are written directly to channel registers.
// CFG0 is set to contiguous mode (0,0) instead of linked-list mode.
// CTL1 is set with valid=true, last=true (single block).
func (c *GDMAChannel) ConfigureContiguous(srcAddr, dstAddr uintptr, words uint32) {
	c.reg(gdmaRegSAR0).Set(uint32(srcAddr))
	c.reg(gdmaRegSAR1).Set(0)
	c.reg(gdmaRegDAR0).Set(uint32(dstAddr))
	c.reg(gdmaRegDAR1).Set(0)
	c.reg(gdmaRegBlockTS).Set(words - 1)
	// Contiguous mode: no LLI, single block.
	c.reg(gdmaRegCFG0).Set(esp32p4dsi.GDMAChannelCfg0(
		esp32p4dsi.GdmaBlockTransferContiguous,
		esp32p4dsi.GdmaBlockTransferContiguous,
	))
	c.reg(gdmaRegLLP0).Set(0)
	c.reg(gdmaRegLLP1).Set(0)
	// CTL1: valid + last, keep arlen/awlen.  IOC_BLKTFR stays clear to
	// match the working ESP-IDF reference (see NewGDMAChannel).
	c.reg(gdmaRegCTL1).Set(esp32p4dsi.GDMAChannelCtl1(15, 15, false, true, true))
}

// ConfigureReload programmes a contiguous block in DW-GDMA auto-reload mode:
// after each block the channel reloads the same source/destination/size and
// continues, without a link-list fetch or a CPU re-arm.  This keeps the DSI
// bridge's request line asserted across frame boundaries, which is what a
// continuously scanned-out panel needs.
//
// d483: LAST must be false so the DW-GDMA never generates DMA_TFR_DONE or
// enters the "completing" state at block boundaries.  With LAST=true the
// channel briefly pauses for TFR_DONE processing at every reload boundary,
// draining the bridge FIFO enough to trigger the permanent reserve latch
// (~39 underruns/15s in d482).  VALID=true tells the DMA the shadow registers
// are ready; LAST=false tells it the transfer is not finished — keep going.
func (c *GDMAChannel) ConfigureReload(srcAddr, dstAddr uintptr, words uint32) {
	c.reg(gdmaRegSAR0).Set(uint32(srcAddr))
	c.reg(gdmaRegSAR1).Set(0)
	c.reg(gdmaRegDAR0).Set(uint32(dstAddr))
	c.reg(gdmaRegDAR1).Set(0)
	c.reg(gdmaRegBlockTS).Set(words - 1)
	c.reg(gdmaRegCFG0).Set(esp32p4dsi.GDMAChannelCfg0(
		esp32p4dsi.GdmaBlockTransferReload,
		esp32p4dsi.GdmaBlockTransferReload,
	))
	c.reg(gdmaRegLLP0).Set(0)
	c.reg(gdmaRegLLP1).Set(0)
	c.reg(gdmaRegCTL1).Set(esp32p4dsi.GDMAChannelCtl1(15, 15, false, false, true))
}

// RearmContiguous re-arms the channel for another contiguous block transfer
// by disarming first (DW-GDMA faults if SAR/BlockTS/CTL written while
// channel is enabled), re-writing SAR, BLOCK_TS, and CTL1.
func (c *GDMAChannel) RearmContiguous(srcAddr uintptr, words uint32) {
	c.Disarm()
	c.reg(gdmaRegSAR0).Set(uint32(srcAddr))
	c.reg(gdmaRegBlockTS).Set(words - 1)
	c.reg(gdmaRegCTL1).Set(esp32p4dsi.GDMAChannelCtl1(15, 15, false, true, true))
}

// PokeSWHandshakeDst asserts the destination software handshake request.
// Only effective when hs_sel_dst = SW (1) in CFG1.
func (c *GDMAChannel) PokeSWHandshakeDst() {
	// SWHSDST0: bit 0 = swhs_req_dst, bit 1 = swhs_req_dst_we
	c.reg(gdmaRegSWHSDST0).Set(0x3) // req=1, we=1
}

// SetSoftwareHandshakeDst changes hs_sel_dst to software (1) in CFG1
// while preserving all other fields.  Used for bring-up diagnostics to
// bypass the hardware handshake and manually trigger DMA bursts.
func (c *GDMAChannel) SetSoftwareHandshakeDst() {
	cfg1 := c.reg(gdmaRegCFG1).Get()
	cfg1 |= 1 << 4 // hs_sel_dst = 1 (software)
	c.reg(gdmaRegCFG1).Set(cfg1)
}

// ReconfigureM2PBridgeFC reprograms the channel for memory→peripheral with the
// DSI bridge as the flow controller (tt_fc = M2P_DST = 6).  All other settings
// (master ports, burst sizes, handshake, etc.) remain the same as the normal
// display path.  Use ReconfigureM2P to restore DMA-as-FC mode.
func (c *GDMAChannel) ReconfigureM2PBridgeFC() {
	c.m2m = false
	cfg1 := esp32p4dsi.GDMAChannelCfg1(
		esp32p4dsi.GdmaTTFCM2PDst,                              // bridge as flow controller
		esp32p4dsi.GdmaHandshakeSW, esp32p4dsi.GdmaHandshakeHW, // src=memory(SW), dst=DSI bridge(HW)
		esp32p4dsi.GdmaHandshakePeriphDSI, esp32p4dsi.GdmaHandshakePeriphDSI,
		1, 5-1, 2-1,
	)
	c.reg(gdmaRegCFG1).Set(cfg1)
}

// SetupLinkListChain programmes a multi-item linked list where each item
// transfers the same source block (repeating the framebuffer) to the same
// destination (bridge FIFO).  This allows a partial framebuffer to fill a
// full DPI frame: e.g. 6 items × 100 lines = 600 lines.
//
// The items are written through their non-cacheable aliases.  The last item
// has is_last=true and LLP=0 (end of list); intermediate items chain to
// the next via LLP.
func (c *GDMAChannel) SetupLinkListChain(lli *dmaLLIItem, count int, srcAddr, dstAddr uintptr, wordsPerBlock uint32) {
	c.SetupLinkListChainStride(lli, count, srcAddr, dstAddr, wordsPerBlock, 0)
}

// SetupLinkListChainStride is SetupLinkListChain with a per-item source stride
// in bytes: item i reads from srcAddr + i*strideBytes.  A stride of
// wordsPerBlock*8 tiles consecutive framebuffer regions (one item per
// horizontal band), which is the historical d256/d262/d263 shape.
func (c *GDMAChannel) SetupLinkListChainStride(lli *dmaLLIItem, count int, srcAddr, dstAddr uintptr, wordsPerBlock uint32, strideBytes uintptr) {
	ctl0 := esp32p4dsi.GDMAChannelCtl0(
		esp32p4dsi.GDMMasterMem, esp32p4dsi.GDMMasterDSI,
		esp32p4dsi.GdmaBurstModeIncr, esp32p4dsi.GdmaBurstModeFixed,
		esp32p4dsi.GdmaTransWidth64, esp32p4dsi.GdmaTransWidth64,
		esp32p4dsi.GdmaBurstItems512, esp32p4dsi.GdmaBurstItems256,
	)
	for i := 0; i < count; i++ {
		item := gdmaLLIAt(lli, i)
		itemBase := uintptr(unsafe.Pointer(item))
		nc := (*dmaLLIItem)(unsafe.Pointer(itemBase + gdmaNonCacheableOff))
		nc.SarLo = uint32(srcAddr + uintptr(i)*strideBytes)
		nc.SarHi = 0
		nc.DarLo = uint32(dstAddr)
		nc.DarHi = 0
		nc.BlockTS = wordsPerBlock - 1
		nc.CtlLo = ctl0
		isLast := i == count-1
		nc.CtlHi = esp32p4dsi.GDMAChannelCtl1(15, 15, false, isLast, true)
		if isLast {
			nc.LLPLo = esp32p4dsi.GDMMasterMem // null pointer, LMS=memory
		} else {
			next := gdmaLLIAt(lli, i+1)
			nc.LLPLo = uint32(uintptr(unsafe.Pointer(next)))&^0x3F | esp32p4dsi.GDMMasterMem
		}
		nc.LLPHi = 0
		nc.SStat, nc.DStat, nc.StatLo, nc.StatHi = 0, 0, 0, 0
	}
	riscv.Asm("fence")
	CacheInvalidateL1(uintptr(unsafe.Pointer(lli)), uintptr(count)*esp32p4dsi.GdmaLLISize)
	c.lli = lli
	c.lliCount = count
	lliBase := uintptr(unsafe.Pointer(lli))
	c.reg(gdmaRegLLP0).Set(uint32(lliBase)&^0x3F | esp32p4dsi.GDMMasterMem)
	c.reg(gdmaRegLLP1).Set(0)
}

// SetupCircularChain programmes a circular linked list where the last item
// points back to the first, and no item has is_last=true.  The DMA loops
// forever through the chain without CPU intervention.  This avoids the
// re-arm problem entirely: the bridge continuously receives data, and the
// CPU can update the framebuffer via NC alias at any time.
//
// RAW_NUM_TOTAL must be set to the full frame size (count × wordsPerBlock)
// so the bridge's auto_reload counter stays synchronised with the DMA cycle.
func (c *GDMAChannel) SetupCircularChain(lli *dmaLLIItem, count int, srcAddr, dstAddr uintptr, wordsPerBlock uint32) {
	ctl0 := esp32p4dsi.GDMAChannelCtl0(
		esp32p4dsi.GDMMasterMem, esp32p4dsi.GDMMasterDSI,
		esp32p4dsi.GdmaBurstModeIncr, esp32p4dsi.GdmaBurstModeFixed,
		esp32p4dsi.GdmaTransWidth64, esp32p4dsi.GdmaTransWidth64,
		esp32p4dsi.GdmaBurstItems512, esp32p4dsi.GdmaBurstItems256,
	)
	for i := 0; i < count; i++ {
		item := gdmaLLIAt(lli, i)
		itemBase := uintptr(unsafe.Pointer(item))
		nc := (*dmaLLIItem)(unsafe.Pointer(itemBase + gdmaNonCacheableOff))
		nc.SarLo = uint32(srcAddr)
		nc.SarHi = 0
		nc.DarLo = uint32(dstAddr)
		nc.DarHi = 0
		nc.BlockTS = wordsPerBlock - 1
		nc.CtlLo = ctl0
		// Circular: all items are valid, none is last.
		nc.CtlHi = esp32p4dsi.GDMAChannelCtl1(15, 15, false, false, true)
		// Every item chains to the next; last wraps to first.
		nextIdx := (i + 1) % count
		next := gdmaLLIAt(lli, nextIdx)
		nc.LLPLo = uint32(uintptr(unsafe.Pointer(next)))&^0x3F | esp32p4dsi.GDMMasterMem
		nc.LLPHi = 0
		nc.SStat, nc.DStat, nc.StatLo, nc.StatHi = 0, 0, 0, 0
	}
	riscv.Asm("fence")
	CacheInvalidateL1(uintptr(unsafe.Pointer(lli)), uintptr(count)*esp32p4dsi.GdmaLLISize)
	c.lli = lli
	c.lliCount = count
	lliBase := uintptr(unsafe.Pointer(lli))
	c.reg(gdmaRegLLP0).Set(uint32(lliBase)&^0x3F | esp32p4dsi.GDMMasterMem)
	c.reg(gdmaRegLLP1).Set(0)
}

// Rearm re-arms the channel for a new transfer by re-writing the LLI block
// markers (valid + last) via the non-cacheable alias for all items in the
// chain, and re-pointing the channel's LLP register at the first LLI.
// This matches the ESP-IDF mipi_dsi_dma_trans_done_cb re-arm sequence.
func (c *GDMAChannel) Rearm() {
	if c.lli == nil {
		return
	}
	count := c.lliCount
	if count <= 0 {
		count = 1
	}
	for i := 0; i < count; i++ {
		item := gdmaLLIAt(c.lli, i)
		itemBase := uintptr(unsafe.Pointer(item))
		nc := (*dmaLLIItem)(unsafe.Pointer(itemBase + gdmaNonCacheableOff))
		isLast := i == count-1
		nc.CtlHi = esp32p4dsi.GDMAChannelCtl1(15, 15, false, isLast, true)
	}
	riscv.Asm("fence")
	lliBase := uintptr(unsafe.Pointer(c.lli))
	c.reg(gdmaRegLLP0).Set(uint32(lliBase)&^0x3F | esp32p4dsi.GDMMasterMem)
}

// RearmNoFence re-validates the link-list items and re-points LLP without the
// trailing fence.  The fence is only needed when the caller immediately polls
// CHEN0 for completion; the display re-arm path is latency-sensitive and the
// channel is already disabled, so the fence can be omitted.
func (c *GDMAChannel) RearmNoFence() {
	if c.lli == nil {
		return
	}
	count := c.lliCount
	if count <= 0 {
		count = 1
	}
	for i := 0; i < count; i++ {
		item := gdmaLLIAt(c.lli, i)
		itemBase := uintptr(unsafe.Pointer(item))
		nc := (*dmaLLIItem)(unsafe.Pointer(itemBase + gdmaNonCacheableOff))
		isLast := i == count-1
		nc.CtlHi = esp32p4dsi.GDMAChannelCtl1(15, 15, false, isLast, true)
	}
	lliBase := uintptr(unsafe.Pointer(c.lli))
	c.reg(gdmaRegLLP0).Set(uint32(lliBase)&^0x3F | esp32p4dsi.GDMMasterMem)
	// The items were written through the non-cacheable alias, but the DMA
	// fetches the cached alias address.  Drop any stale dirty cache lines for
	// the cached alias so a later write-back cannot clobber the descriptor
	// with zeros (which stalls the descriptor fetch and drains the bridge).
	CacheInvalidateL1(lliBase, uintptr(count)*esp32p4dsi.GdmaLLISize)
}

// Start enables the channel, launching the transfer.
func (c *GDMAChannel) Start() {
	// CHEN0: EN_WE | EN for the channel selection.
	gdmaHW().CHEN0.Set(gdmaChEnLow << (c.num * 2))
	// Fence ensures the CHEN0 write reaches the bus before the caller polls
	// CHEN0 for done detection — without this, a back-to-back read could
	// see the old value (0) and falsely report the channel as finished.
	riscv.Asm("fence iorw, iorw")
}

// StartSuspended enables the channel with the suspend request already set.
// The channel is enabled (so the DSI bridge sees a live DMA writer) but does
// not transfer until Resume is called.  Used to enable DPI with both a live
// DMA channel and an empty bridge FIFO.
func (c *GDMAChannel) StartSuspended() {
	gdmaHW().CHEN0.Set((gdmaChEnLow << (c.num * 2)) | (0x1010000 << c.num))
	riscv.Asm("fence iorw, iorw")
}

// Stop disables the channel (EN_W|!EN).
func (c *GDMAChannel) Stop() {
	gdmaHW().CHEN0.Set((0x100 << (c.num * 2)) & 0xFFFF)
}

// Suspend requests the channel to stop at the next transaction boundary.
// Unlike Stop, this preserves the channel's position so Resume can continue
// the same block.  This is the DW-GDMA CHEN0 suspend protocol.
func (c *GDMAChannel) Suspend() {
	gdmaHW().CHEN0.Set(0x1010000 << c.num)
	riscv.Asm("fence iorw, iorw")
}

// Resume clears the channel suspend request and continues the block.
func (c *GDMAChannel) Resume() {
	gdmaHW().CHEN0.Set(0x1000000 << c.num)
	riscv.Asm("fence iorw, iorw")
}

// Suspended reports whether the channel's suspend bit is set.
func (c *GDMAChannel) Suspended() bool {
	return gdmaHW().CHEN0.Get()&(1<<(16+c.num)) != 0
}

// Disarm disables the channel and waits for it to reach the disabled state.
// The DW-GDMA faults (store access fault, mcause=7) if a transfer descriptor
// register such as LLP0 is written while the channel is still enabled; the
// display loop must disarm between frames before re-arming the link list.
func (c *GDMAChannel) Disarm() {
	gdmaHW().CHEN0.Set(0x100 << (c.num * 2))
	for i := 0; i < 1000000; i++ {
		if gdmaHW().CHEN0.Get()&(1<<c.num) == 0 {
			break
		}
		riscv.Asm("nop")
	}
}

// Transferred returns the number of 64-bit words completed by the channel, and
// whether the transfer has finished.  Done detection uses the CHEN0 register
// (channel-enable bit) instead of INTSTATUS0: the DW-GDMA automatically
// disables a channel when its last LLI block completes (llLast=true in ctl1),
// so CHEN0 bit N going from 1→0 is the definitive transfer-done signal.
// INTSTATUS0 was unreliable because CH_INTSTATUS_ENABLE0 defaults enable
// multiple interrupt sources after reset, causing the aggregated bit to
// re-assert even after ClearDone.
func (c *GDMAChannel) Transferred() (words uint32, done bool) {
	words = c.reg(gdmaRegStatus0).Get()
	done = gdmaHW().CHEN0.Get()&(1<<c.num) == 0
	return
}

// ClearDone clears the channel's transfer-done interrupt.  The channel-level
// interrupt (CH_INTCLEARREG0) must be cleared, not the controller-level
// COMMONREG_INTCLEAR0 (which only clears bus-error interrupts).  The
// aggregated INTSTATUS0.CH_INTSTAT bit clears automatically once the
// underlying channel interrupt is cleared.
func (c *GDMAChannel) ClearDone() {
	// Clear exactly the asserted channel events, matching IDF's
	// dw_gdma_ll_channel_clear_intr().  Writing all ones can leave the
	// controller aggregate INTSTATUS0 bit asserted, which prevents a
	// level-triggered CLIC source from re-dispatching.
	status := c.reg(gdmaRegINTSTATUS0).Get()
	if status != 0 {
		c.reg(gdmaRegINTCLEAR0).Set(status)
	}
}

// ReadSAR returns the current DMA source address register value for this
// channel.  Used for bring-up diagnostics to track how far DMA has advanced
// through the source buffer.
func (c *GDMAChannel) ReadSAR() uint32 {
	return c.reg(gdmaRegSAR0).Get()
}

// IRQRegisters returns the interrupt-related DW-GDMA registers for bring-up
// diagnostics: controller aggregate status, common status/enables, and this
// channel's status/enables/clear.
func (c *GDMAChannel) IRQRegisters() (ctrlStatus, commonStatus, commonEnable, commonSignal, chStatus, chEnable, chSignal, chClear uint32) {
	return gdmaHW().INTSTATUS0.Get(),
		gdmaHW().COMMONREG_INTSTATUS0.Get(),
		gdmaHW().COMMONREG_INTSTATUS_ENABLE0.Get(),
		gdmaHW().COMMONREG_INTSIGNAL_ENABLE0.Get(),
		c.reg(gdmaRegINTSTATUS0).Get(),
		c.reg(gdmaRegINTSTATUS_EN0).Get(),
		c.reg(gdmaRegINTSIGNAL_EN0).Get(),
		0 // CH_INTCLEARREG0 is write-only; reading it faults (0x50081198)
}

// State returns a snapshot of the DW-GDMA controller and this channel's
// registers for bring-up diagnostics.
func (c *GDMAChannel) State() (chen0, intstatus0, status0, llp0, cfg1, cfg0, ctl0 uint32) {
	return gdmaHW().CHEN0.Get(),
		gdmaHW().INTSTATUS0.Get(),
		c.reg(gdmaRegStatus0).Get(),
		c.reg(gdmaRegLLP0).Get(),
		c.reg(gdmaRegCFG1).Get(),
		c.reg(gdmaRegCFG0).Get(),
		c.reg(gdmaRegCTL0).Get()
}

// DebugDump prints a full snapshot of the DW-GDMA controller and channel-0
// registers for bring-up diagnostics.
func (c *GDMAChannel) DebugDump() {
	println("  dma chen0=", gdmaHW().CHEN0.Get(), "intst0=", gdmaHW().INTSTATUS0.Get())
	println("  dma cfg0=", c.reg(gdmaRegCFG0).Get(), "cfg1=", c.reg(gdmaRegCFG1).Get())
	println("  dma ctl0=", c.reg(gdmaRegCTL0).Get(), "ctl1=", c.reg(gdmaRegCTL1).Get())
	println("  dma sar0=", c.reg(gdmaRegSAR0).Get(), "sar1=", c.reg(gdmaRegSAR1).Get())
	println("  dma dar0=", c.reg(gdmaRegDAR0).Get(), "dar1=", c.reg(gdmaRegDAR1).Get())
	println("  dma blkts=", c.reg(gdmaRegBlockTS).Get(), "status0=", c.reg(gdmaRegStatus0).Get())
	// Channel-0 interrupt/status word.  Bit positions follow the ESP32-P4
	// hw_ver1 dmac_chn_intstatus0_reg_t (dw_gdma_struct.h), NOT the older
	// DW-GDMA field order.  The d288 circular-DMA failure was hidden because
	// this decode used bit 12 for lli_invalid and bit 16 for disabled; the
	// real bits are 13 and 30.
	//   0=block_tfr_done       1=dma_tfr_done
	//   3=src_transcomp        4=dst_transcomp
	//   5=src_dec_err          6=dst_dec_err
	//   7=src_slv_err          8=dst_slv_err
	//   9=lli_rd_dec_err      10=lli_wr_dec_err
	//  11=lli_rd_slv_err      12=lli_wr_slv_err
	//  13=shadowreg_or_lli_invalid_err
	//  30=ch_disabled         31=ch_aborted
	ist0 := c.reg(gdmaRegINTSTATUS0).Get()
	println("  dma ch0 int_st0=", ist0, " bits: blk_done=", ist0&1, "dma_tfr_done=", (ist0>>1)&1,
		"src_dec=", (ist0>>5)&1, "dst_dec=", (ist0>>6)&1,
		"src_slv=", (ist0>>7)&1, "dst_slv=", (ist0>>8)&1,
		"lli_invalid=", (ist0>>13)&1, "disabled=", (ist0>>30)&1, "aborted=", (ist0>>31)&1)
	println("  dma ch0 int_st_en0=", c.reg(gdmaRegINTSTATUS_EN0).Get(),
		"int_sig_en0=", c.reg(gdmaRegINTSIGNAL_EN0).Get())
	println("  dma ch0 int_st1=", gdmaHW().CH1_INTSTATUS1.Get(),
		"ch0 status0=", gdmaHW().CH1_STATUS0.Get()&0xFFFFFFFF, "status1=", gdmaHW().CH1_STATUS1.Get())
}

// LLIWords returns the raw 64-byte link-list item contents (16 words) for
// bring-up diagnostics.
func (c *GDMAChannel) LLIWords() [16]uint32 {
	var out [16]uint32
	if c.lli == nil {
		return out
	}
	lli := (*[16]uint32)(unsafe.Pointer(c.lli))
	for i := range out {
		out[i] = lli[i]
	}
	return out
}

// LLIAddr returns the cached and non-cacheable addresses of the link-list item
// so diagnostics can compare what the CPU wrote vs what the DMA reads.
func (c *GDMAChannel) LLIAddr() (cached, nonCacheable uintptr) {
	if c.lli == nil {
		return 0, 0
	}
	cached = uintptr(unsafe.Pointer(c.lli))
	return cached, cached + gdmaNonCacheableOff
}

// LLIWordsAt returns the 16-word contents of a link-list item at the given
// address (cached or non-cacheable), for coherency diagnostics.
func LLIWordsAt(addr uintptr) [16]uint32 {
	var out [16]uint32
	lli := (*[16]uint32)(unsafe.Pointer(addr))
	for i := range out {
		out[i] = lli[i]
	}
	return out
}
