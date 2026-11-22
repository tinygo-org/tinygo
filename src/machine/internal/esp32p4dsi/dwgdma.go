package esp32p4dsi

// DW-GDMA channel and link-list configuration mathematics for the ESP32-P4.
//
// The ESP32-P4 feeds the MIPI-DSI bridge with pixel data using the DesignWare
// AXI DW-GDMA controller (0x50081000).  The controller is programmed once for
// a memory→peripheral (DSI bridge) transfer, with the DMA engine as the flow
// controller and a single link-list item carrying the whole frame:
//
//	src: L2MEM (or PSRAM once Phase 6 lands) framebuffer, 64-bit transfers
//	dst: MIPI_DSI_BRG_MEM_BASE (0x50105000), burst 256
//
// The register-field encodings below mirror:
//   - components/esp_hal_dma/esp32p4/include/hal/dw_gdma_ll.h
//   - components/esp_lcd/dsi/esp_lcd_panel_dpi.c  (dpi_panel_create_dma_link)
//
// They are pure functions so the graph can be unit-tested on the host.

// DW-GDMA master port numbers (dw_gdma_ll.h).
const (
	GDMMasterDSI = 0 // master port 0 can reach the DSI bridge FIFO
	GDMMasterCSI = 0 // master port 0 can reach the CSI bridge FIFO
	GDMMasterMem = 1 // master port 1 reaches L2MEM / ROM / flash / PSRAM
)

// DW-GDMA flow-control transfer-type codes (dw_gdma_ll.h dw_gdma_flow_t).
const (
	GdmaTTFCM2PDmac uint32 = 1 // M2P, DMA as flow controller
	GdmaTTFCM2PDst  uint32 = 6 // M2P, peripheral (DSI bridge) as flow controller
)

// DW-GDMA link-list related constants.
const (
	GdmaBlockTransferContiguous uint32 = 0 // DW_GDMA_BLOCK_TRANSFER_CONTIGUOUS (channel regs, no LLI)
	GdmaBlockTransferReload     uint32 = 1 // DW_GDMA_BLOCK_TRANSFER_RELOAD (repeat the block automatically)
	GdmaBlockTransferShadow     uint32 = 2 // DW_GDMA_BLOCK_TRANSFER_SHADOW
	GdmaBlockTransferList       uint32 = 3 // DW_GDMA_BLOCK_TRANSFER_LIST (link-list driven)
	GdmaHandshakeHW             uint32 = 0 // DW_GDMA_HANDSHAKE_HW
	GdmaHandshakeSW             uint32 = 1 // DW_GDMA_HANDSHAKE_SW
	GdmaRoleMem                 uint32 = 0 // DW_GDMA_ROLE_MEM
	GdmaRolePeriphDSI           uint32 = 1 // DW_GDMA_ROLE_PERIPH_DSI
	GdmaTransWidth64            uint32 = 3 // DW_GDMA_TRANS_WIDTH_64 (0=8,1=16,2=32,3=64)
	GdmaBurstModeIncr           uint32 = 0 // DW_GDMA_BURST_MODE_INCREMENT
	GdmaBurstModeFixed          uint32 = 1 // DW_GDMA_BURST_MODE_FIXED
	GdmaBurstItems256           uint32 = 7 // DW_GDMA_BURST_ITEMS_256 (ESP-IDF enum: 1=0,4=1,...,256=7,512=8,1024=9)
	GdmaBurstItems512           uint32 = 8 // DW_GDMA_BURST_ITEMS_512
)

// Hardware handshake peripheral IDs for the CFG1 src_per/dst_per fields
// (dw_gdma_ll.h dw_gdma_ll_hw_handshake_periph_t).  These are distinct from
// the dw_gdma_role_t values above: the LL maps the DSI *role* onto the DSI
// *peripheral ID* 0, CSI onto 1 and ISP onto 2.
const (
	GdmaHandshakePeriphDSI uint32 = 0 // DW_GDMA_LL_HW_HANDSHAKE_PERIPH_DSI
	GdmaHandshakePeriphCSI uint32 = 1 // DW_GDMA_LL_HW_HANDSHAKE_PERIPH_CSI
	GdmaHandshakePeriphISP uint32 = 2 // DW_GDMA_LL_HW_HANDSHAKE_PERIPH_ISP
)

// Link-list item layout offsets (64-byte items, 64-byte aligned).
const (
	GdmaLLISARLo    = 0x00 // source address low
	GdmaLLISARHi    = 0x04 // source address high
	GdmaLLIDARLo    = 0x08 // destination address low
	GdmaLLIDARHi    = 0x0C // destination address high
	GdmaLLIBlockTS  = 0x10 // (transfer size items) - 1
	GdmaLLIReserved = 0x14
	GdmaLLILLPLo    = 0x18 // next link-list item (addr >> 6) | LMS
	GdmaLLILLPHi    = 0x1C // high 32 bits of the next item pointer
	GdmaLLICtlLo    = 0x20 // control word low
	GdmaLLICtlHi    = 0x24 // control word high
	GdmaLLISStat    = 0x28 // source peripheral status
	GdmaLLIDStat    = 0x2C // destination peripheral status
	GdmaLLIStatLo   = 0x30 // channel status low
	GdmaLLIStatHi   = 0x34 // channel status high
	GdmaLLISize     = 0x40 // 64 bytes per item
)

// GDMAChannelCtl0 encodes the channel CTL0 register (control word low) from
// the master ports, burst modes, transfer widths and burst item counts.
func GDMAChannelCtl0(srcMaster, dstMaster, srcBurstMode, dstBurstMode, srcTrWidth, dstTrWidth, srcBurstItems, dstBurstItems uint32) uint32 {
	return srcMaster<<0 |
		dstMaster<<2 |
		srcBurstMode<<4 |
		dstBurstMode<<6 |
		(srcTrWidth&0x7)<<8 |
		(dstTrWidth&0x7)<<11 |
		(srcBurstItems&0xF)<<14 |
		(dstBurstItems&0xF)<<18
}

// GDMAChannelCtl1 encodes the channel CTL1 register (control word high):
// burst lengths and the LLI block markers.
func GDMAChannelCtl1(arlen, awlen uint8, iocBlkTfr, llLast, llValid bool) uint32 {
	var word uint32
	if arlen > 0 {
		word |= 1 << 6 // ARLEN_EN
		word |= uint32(arlen&0x7F) << 7
	}
	if awlen > 0 {
		word |= 1 << 15 // AWLEN_EN
		word |= uint32(awlen&0x7F) << 16
	}
	if iocBlkTfr {
		word |= 1 << 26 // IOC_BLKTFR
	}
	if llLast {
		word |= 1 << 30 // SHADOWREG_OR_LLI_LAST
	}
	if llValid {
		word |= 1 << 31 // SHADOWREG_OR_LLI_VALID
	}
	return word
}

// GDMAChannelCfg0 encodes the channel CFG0 register: multi-block transfer
// types for source and destination (both LIST for a link-list driven transfer).
func GDMAChannelCfg0(srcMultblkType, dstMultblkType uint32) uint32 {
	return srcMultblkType<<0 | dstMultblkType<<2
}

// GDMAChannelCfg1 encodes the channel CFG1 register: transfer-flow type,
// handshake selection, peripheral IDs, priority and outstanding-request limits.
func GDMAChannelCfg1(ttFC, hsSelSrc, hsSelDst, srcPer, dstPer uint32, priority uint32, srcOsrLmt, dstOsrLmt uint32) uint32 {
	return ttFC<<0 |
		hsSelSrc<<3 |
		hsSelDst<<4 |
		(srcPer&0xF)<<7 |
		(dstPer&0xF)<<12 |
		(priority&0x7)<<17 |
		(srcOsrLmt&0x7)<<23 |
		(dstOsrLmt&0x7)<<27
}

// GDMAFrameWords returns the number of 64-bit words in a frame of the given
// width×height×bits-per-pixel, rounded up to cover partial final words.
func GDMAFrameWords(hSize, vSize, bitsPerPixel uint32) uint32 {
	bits := hSize * vSize * bitsPerPixel
	return (bits + 63) / 64
}
