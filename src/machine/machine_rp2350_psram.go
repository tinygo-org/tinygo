//go:build pico_plus2

package machine

import (
	"device/rp"
	"runtime/interrupt"
	"unsafe"
)

const (
	// Uncached alias of QMI window 1, XIP_NOCACHE_NOALLOC_BASE + 16MB.
	// See RP2350 datasheet section 4.4.1.
	psramBase = 0x15000000
)

var psramSize uintptr

// PSRAM returns the memory of the external PSRAM, or nil if none was found.
func PSRAM() []byte {
	if psramSize == 0 {
		return nil
	}
	return unsafe.Slice((*byte)(unsafe.Pointer(uintptr(psramBase))), psramSize)
}

// initPSRAM must run after clocks.init, as the timing assumes cpuFreq.
func initPSRAM() {
	psramCSPin.setFunc(fnQMI)
	state := interrupt.Disable()
	psramSize = psramSetup()
	interrupt.Restore(state)
}

// APS6404L timing at the 150MHz clk_sys, based on psram_init in MicroPython
// ports/rp2/rp2_psram.c.
const (
	psramClkdiv  = (cpuFreq + 133*MHz - 1) / (133 * MHz)
	psramRxdelay = psramClkdiv
	// Chip select may be held for at most 8us, in units of 64 clk_sys cycles.
	psramMaxSelect = 8 * cpuFreq / MHz / 64
	// Chip select must be released for at least 18ns.
	psramMinDeselect = (18*cpuFreq/MHz+999)/1000 - (psramClkdiv+1)/2
)

// psramSetup detects the PSRAM on CS1 and maps it into QMI window 1. It runs from
// RAM because QMI direct mode stalls XIP accesses to flash.
//
//go:section .ramfuncs
func psramSetup() uintptr {
	qmi := rp.QMI
	qmi.DIRECT_CSR.Set(30<<rp.QMI_DIRECT_CSR_CLKDIV_Pos | rp.QMI_DIRECT_CSR_EN)
	for qmi.DIRECT_CSR.HasBits(rp.QMI_DIRECT_CSR_BUSY) {
	}

	// Leave QPI mode in case it is still enabled from before a reset.
	qmi.DIRECT_CSR.SetBits(rp.QMI_DIRECT_CSR_ASSERT_CS1N)
	qmi.DIRECT_TX.Set(rp.QMI_DIRECT_TX_OE | rp.QMI_DIRECT_TX_IWIDTH_Q<<rp.QMI_DIRECT_TX_IWIDTH_Pos | 0xf5)
	for qmi.DIRECT_CSR.HasBits(rp.QMI_DIRECT_CSR_BUSY) {
	}
	qmi.DIRECT_RX.Get()
	qmi.DIRECT_CSR.ClearBits(rp.QMI_DIRECT_CSR_ASSERT_CS1N)

	// Read ID: command, 3 address bytes, MFID, KGD, EID.
	qmi.DIRECT_CSR.SetBits(rp.QMI_DIRECT_CSR_ASSERT_CS1N)
	var kgd, eid uint32
	for i := 0; i < 7; i++ {
		tx := uint32(0xff)
		if i == 0 {
			tx = 0x9f
		}
		qmi.DIRECT_TX.Set(tx)
		for !qmi.DIRECT_CSR.HasBits(rp.QMI_DIRECT_CSR_TXEMPTY) {
		}
		for qmi.DIRECT_CSR.HasBits(rp.QMI_DIRECT_CSR_BUSY) {
		}
		rx := qmi.DIRECT_RX.Get() & 0xff
		switch i {
		case 5:
			kgd = rx
		case 6:
			eid = rx
		}
	}
	qmi.DIRECT_CSR.ClearBits(rp.QMI_DIRECT_CSR_ASSERT_CS1N | rp.QMI_DIRECT_CSR_EN)

	if kgd != 0x5d {
		return 0
	}
	var size uintptr
	switch {
	case eid == 0x26 || eid>>5 == 2:
		size = 8 << 20
	case eid>>5 == 1:
		size = 4 << 20
	case eid>>5 == 0:
		size = 2 << 20
	default:
		size = 1 << 20
	}

	// Enable QPI mode.
	qmi.DIRECT_CSR.Set(10<<rp.QMI_DIRECT_CSR_CLKDIV_Pos | rp.QMI_DIRECT_CSR_EN | rp.QMI_DIRECT_CSR_AUTO_CS1N)
	for qmi.DIRECT_CSR.HasBits(rp.QMI_DIRECT_CSR_BUSY) {
	}
	qmi.DIRECT_TX.Set(rp.QMI_DIRECT_TX_NOPUSH | 0x35)
	for qmi.DIRECT_CSR.HasBits(rp.QMI_DIRECT_CSR_BUSY) {
	}

	qmi.M1_TIMING.Set(1<<rp.QMI_M1_TIMING_COOLDOWN_Pos |
		rp.QMI_M1_TIMING_PAGEBREAK_1024<<rp.QMI_M1_TIMING_PAGEBREAK_Pos |
		psramMaxSelect<<rp.QMI_M1_TIMING_MAX_SELECT_Pos |
		psramMinDeselect<<rp.QMI_M1_TIMING_MIN_DESELECT_Pos |
		psramRxdelay<<rp.QMI_M1_TIMING_RXDELAY_Pos |
		psramClkdiv<<rp.QMI_M1_TIMING_CLKDIV_Pos)

	// Quad 0xEB read with 6 dummy clocks, quad 0x38 write.
	const quad = rp.QMI_M1_RFMT_PREFIX_WIDTH_Q<<rp.QMI_M1_RFMT_PREFIX_WIDTH_Pos |
		rp.QMI_M1_RFMT_ADDR_WIDTH_Q<<rp.QMI_M1_RFMT_ADDR_WIDTH_Pos |
		rp.QMI_M1_RFMT_SUFFIX_WIDTH_Q<<rp.QMI_M1_RFMT_SUFFIX_WIDTH_Pos |
		rp.QMI_M1_RFMT_DUMMY_WIDTH_Q<<rp.QMI_M1_RFMT_DUMMY_WIDTH_Pos |
		rp.QMI_M1_RFMT_DATA_WIDTH_Q<<rp.QMI_M1_RFMT_DATA_WIDTH_Pos |
		rp.QMI_M1_RFMT_PREFIX_LEN_8<<rp.QMI_M1_RFMT_PREFIX_LEN_Pos
	qmi.M1_RFMT.Set(quad | 6<<rp.QMI_M1_RFMT_DUMMY_LEN_Pos)
	qmi.M1_RCMD.Set(0xeb)
	qmi.M1_WFMT.Set(quad)
	qmi.M1_WCMD.Set(0x38)

	qmi.DIRECT_CSR.Set(0)
	rp.XIP_CTRL.CTRL.SetBits(rp.XIP_CTRL_CTRL_WRITABLE_M1)
	return size
}
