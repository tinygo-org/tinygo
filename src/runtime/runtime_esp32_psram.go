//go:build esp32 && numa_psram_qspi

package runtime

// PSRAM on ESP32-D0WD modules such as the WROVER. Order and values follow ESP-IDF
// components/esp_psram/esp32/esp_psram_impl_quad.c, mode PSRAM_CACHE_F80M_S40M.

import (
	"device"
	"device/esp"
	"runtime/volatile"
	"unsafe"
)

const psramAddr = 0x3F800000 // ESP32 TRM v5.8 section 4.3.2.2

// Pins of the D0WD package with the default flash pins.
const (
	psramCSPin  = 16
	psramClkPin = 17
	flashCSPin  = 11
	spiQPin     = 7
	spiDPin     = 8
	spiHDPin    = 9
	spiWPPin    = 10
	loopPin28   = 28 // pads without a pin, used to delay the PSRAM clock
	loopPin29   = 29
)

// GPIO matrix signals, from ESP-IDF soc/esp32/include/soc/gpio_sig_map.h.
const (
	sigSPICLK    = 0
	sigSPIQ      = 1
	sigSPID      = 2
	sigSPIHD     = 3
	sigSPIWP     = 4
	sigSPICS0    = 5
	sigSPICS1    = 6
	sigInFunc224 = 224
	sigInFunc225 = 225
	sigGPIOOut   = 256
)

// PSRAM chip commands and ID fields.
const (
	psramExitQuad   = 0xF5
	psramReadID     = 0x9F
	psramEnterQuad  = 0x35
	psramGoodDie    = 0x5D
	psramReadQuad   = 0xEB
	psramWriteQuad  = 0x38
	psramReadWait   = 5 // wait cycles of psramReadQuad
	matrixWait40    = 1 // extra wait cycles for the GPIO matrix delay at 40MHz
	matrixWait80    = 2 // and at 80MHz
	psram32MbitVer0 = 0x20
)

// Reasons why PSRAM is off.
const (
	psramOK = iota
	psramWrongPackage
	psramOldChip
	psramFlashPins
	psramNoChip
	psramOldPSRAM
	psramBadReadBack
	psramTooSmall
	psramUnknownSize
	psramMapFailed
)

var (
	psramSize   uint32
	psramReason uint8 = psramNoChip
)

//go:extern _spsram
var _spsram [0]byte

//go:extern _epsram
var _epsram [0]byte

//export rom_gpio_matrix_out
func romMatrixOut(gpio, signal, invertOut, invertEnable uint32)

//export rom_gpio_matrix_in
func romMatrixIn(gpio, signal, invert uint32)

//export esp_rom_spiflash_config_clk
func romSPIClock(divider, port uint32) int32

//export ets_efuse_get_spiconfig
func romFlashPins() uint32

//export cache_sram_mmu_set_rom
func romCacheSRAMMap(cpu, pid, vaddr, paddr, pageKB, pages uint32) int32

// Extra wait cycles the ROM flash functions add: [0] SPI0, [1] SPI1.
//
//go:extern g_rom_spiflash_dummy_len_plus
var romExtraWait [2]uint8

// initPSRAM switches on the PSRAM and maps it at psramAddr. The start-up runs from
// IRAM with interrupts off, because it changes SPI0, which also reads the flash.
func initPSRAM() {
	state := device.AsmFull("rsil {}, 15", nil)
	psramReason = psramStart()
	device.AsmFull("wsr {state}, PS", map[string]interface{}{"state": state})
	device.Asm("rsync")
	if psramReason != psramOK {
		return
	}
	start := uintptr(unsafe.Pointer(&_spsram))
	end := uintptr(unsafe.Pointer(&_epsram))
	if end-psramAddr > uintptr(psramSize) {
		psramReason = psramTooSmall
		return
	}
	memzero(unsafe.Pointer(start), end-start)
}

// reportPSRAM stops the program when it has variables in PSRAM but PSRAM is off.
// Without such variables it only prints why PSRAM is off.
func reportPSRAM() {
	if psramReason == psramOK {
		return
	}
	if uintptr(unsafe.Pointer(&_epsram)) == uintptr(unsafe.Pointer(&_spsram)) {
		printstring("PSRAM is off: ")
		printstring(psramReasonText())
		printnl()
		return
	}
	printstring("panic: PSRAM is off: ")
	printstring(psramReasonText())
	printnl()
	abort()
}

func psramReasonText() string {
	switch psramReason {
	case psramWrongPackage:
		return "chip package is not ESP32-D0WD"
	case psramOldChip:
		return "chip revision is below 3"
	case psramFlashPins:
		return "flash does not use the default pins"
	case psramNoChip:
		return "no PSRAM chip answered"
	case psramOldPSRAM:
		return "32Mbit PSRAM version 0 is not supported"
	case psramBadReadBack:
		return "PSRAM read back test failed"
	case psramTooSmall:
		return ".psram section is bigger than the PSRAM"
	case psramUnknownSize:
		return "PSRAM ID has an unknown size"
	case psramMapFailed:
		return "PSRAM memory map failed"
	}
	return "unknown"
}

//go:section .iram.psram
func psramStart() uint8 {
	// Chip package and revision. See ESP-IDF hal/esp32/efuse_hal.c.
	pkg := esp.EFUSE.GetBLK0_RDATA3_RD_CHIP_PACKAGE_4BIT()<<3 | esp.EFUSE.GetBLK0_RDATA3_RD_CHIP_PACKAGE()
	if pkg > 1 {
		return psramWrongPackage
	}
	if esp.EFUSE.GetBLK0_RDATA3_RD_CHIP_VER_REV1() == 0 || esp.EFUSE.GetBLK0_RDATA5_RD_CHIP_VER_REV2() == 0 ||
		esp.APB_CTRL.DATE.Get()>>31 == 0 {
		return psramOldChip
	}
	if romFlashPins() != 0 {
		return psramFlashPins
	}

	esp.SPI0.EXT3.Set(1)
	esp.SPI1.SetUSER_USR_PREP_HOLD(0)
	psramSPI1Init()

	// Delay the PSRAM clock through two loop-back signals for the ID read.
	romMatrixOut(loopPin28, sigSPICLK, 0, 0)
	romMatrixIn(loopPin28, sigInFunc224, 0)
	romMatrixOut(loopPin29, sigInFunc224, 0, 0)
	romMatrixIn(loopPin29, sigInFunc225, 0)
	romMatrixOut(psramClkPin, sigInFunc225, 0, 0)

	psramPins()
	reason := psramTalk()
	if reason != psramOK {
		psramRelease()
	}
	return reason
}

// psramRelease gives the PSRAM clock and select pins back as plain GPIO pins.
//
//go:section .iram.psram
func psramRelease() {
	romMatrixOut(loopPin28, sigGPIOOut, 0, 0)
	romMatrixOut(loopPin29, sigGPIOOut, 0, 0)
	romMatrixOut(psramClkPin, sigGPIOOut, 0, 0)
	romMatrixOut(psramCSPin, sigGPIOOut, 0, 0)
}

// psramTalk reads the ID, switches the PSRAM to quad mode and maps it.
//
//go:section .iram.psram
func psramTalk() uint8 {
	// Exit quad mode, because the PSRAM keeps its power over a restart.
	psramCommand(&psramCmd{tx: psramExitQuad << 8, txBits: 16, quad: true})

	// Read the ID. Some 16Mbit chips answer wrong the first time.
	read := psramCmd{addr: psramReadID << 24, addrBits: 32, cmdBits: 2, waitBits: matrixWait40, rxBits: 64}
	id, _ := psramCommand(&read)
	if byte(id>>8) != psramGoodDie {
		id, _ = psramCommand(&read)
		if byte(id>>8) != psramGoodDie {
			return psramNoChip
		}
	}
	eid := byte(id >> 16)
	if eid == psram32MbitVer0 {
		return psramOldPSRAM
	}

	// Normal clock from now on.
	romMatrixOut(loopPin28, sigGPIOOut, 0, 0)
	romMatrixOut(loopPin29, sigGPIOOut, 0, 0)
	romMatrixOut(psramClkPin, sigSPICLK, 0, 0)
	psramSelectTiming()

	psramCommand(&psramCmd{addr: psramEnterQuad << 24, addrBits: 8})

	// Size codes 0, 1 and 2 are 2MB, 4MB and 8MB, see ESP-IDF PSRAM_EID_SIZE. 8MB
	// chips are used as 4MB. No switch here: its lookup table would be in flash.
	sizeCode := (eid >> 5) & 7
	if sizeCode > 2 {
		return psramUnknownSize
	}
	psramSize = 4 << 20
	if sizeCode == 0 {
		psramSize = 2 << 20
	}
	psramCacheInit()
	if romCacheSRAMMap(0, 0, psramAddr, 0, 32, psramSize>>15) != 0 {
		return psramMapFailed
	}

	// Writes can be lost without an error when the timing is wrong. The offset
	// in each word also finds a wrong size, because a mirror overwrites lower words.
	for off := uint32(0); off < psramSize; off += 1024 {
		*(*uint32)(unsafe.Pointer(uintptr(psramAddr + off))) = off
	}
	for off := uint32(0); off < psramSize; off += 1024 {
		if *(*uint32)(unsafe.Pointer(uintptr(psramAddr + off))) != off {
			return psramBadReadBack
		}
	}
	return psramOK
}

//go:section .iram.psram
func psramSPI1Init() {
	spi := esp.SPI1
	spi.SLAVE.ClearBits(1 << 9) // transfer done interrupt
	spi.SetPIN_CK_IDLE_EDGE(0)
	spi.SetUSER_CK_OUT_EDGE(0)
	spi.SetCTRL_WR_BIT_ORDER(0)
	spi.SetCTRL_RD_BIT_ORDER(0)
	spi.SetUSER_DOUTDIN(0)
	spi.USER1.Set(0)
	spi.SetSLAVE_MODE(0)
	psramClearBuffer()
	spi.SetUSER_CS_HOLD(0) // no select hold or setup time with the delayed clock
	spi.SetUSER_CS_SETUP(0)
}

//go:section .iram.psram
func psramClearBuffer() {
	w := uintptr(unsafe.Pointer(&esp.SPI1.W0))
	for i := uintptr(0); i < 16; i++ {
		volatile.StoreUint32((*uint32)(unsafe.Pointer(w+i*4)), 0)
	}
}

// psramSelectTiming sets the select hold and setup time of SPI0 and SPI1 for the normal clock.
// Hold time 0 is for 40MHz PSRAM, see ESP-IDF PSRAM_CS_HOLD_TIME (1 only at 80MHz).
//
//go:section .iram.psram
func psramSelectTiming() {
	esp.SPI1.SetUSER_CS_HOLD(1)
	esp.SPI1.SetUSER_CS_SETUP(1)
	esp.SPI1.SetCTRL2_HOLD_TIME(0)
	esp.SPI1.SetCTRL2_SETUP_TIME(0)
	esp.SPI0.SetUSER_CS_HOLD(1)
	esp.SPI0.SetUSER_CS_SETUP(1)
	esp.SPI0.SetCTRL2_HOLD_TIME(0)
	esp.SPI0.SetCTRL2_SETUP_TIME(0)
}

// psramPins connects flash and PSRAM signals. SPI0 stays at 80MHz for the
// flash, SPI1 runs at 40MHz for the PSRAM commands.
//
//go:section .iram.psram
func psramPins() {
	// Wait cycles of the current flash read mode, see ESP-IDF psram_gpio_config.
	flashWait := uint32(7)
	if esp.SPI0.GetCTRL_FREAD_QIO() != 0 {
		flashWait = 3
	} else if esp.SPI0.GetCTRL_FREAD_DIO() != 0 {
		flashWait = 1
		esp.SPI0.SetUSER1_USR_ADDR_BITLEN(27)
	}
	romExtraWait[0] = matrixWait80
	romExtraWait[1] = matrixWait40
	esp.SPI0.SetUSER1_USR_DUMMY_CYCLELEN(flashWait + matrixWait80)
	romSPIClock(1, 0)
	romSPIClock(2, 1)
	ioMuxDrive(&esp.IO_MUX.GPIO6, 3)
	ioMuxDrive(&esp.IO_MUX.GPIO17, 2)
	esp.SPI0.SetUSER_USR_DUMMY(1)

	romMatrixOut(flashCSPin, sigSPICS0, 0, 0)
	romMatrixOut(psramCSPin, sigSPICS1, 0, 0)
	romMatrixOut(spiQPin, sigSPIQ, 0, 0)
	romMatrixIn(spiQPin, sigSPIQ, 0)
	romMatrixOut(spiDPin, sigSPID, 0, 0)
	romMatrixIn(spiDPin, sigSPID, 0)
	romMatrixOut(spiWPPin, sigSPIWP, 0, 0)
	romMatrixIn(spiWPPin, sigSPIWP, 0)
	romMatrixOut(spiHDPin, sigSPIHD, 0, 0)
	romMatrixIn(spiHDPin, sigSPIHD, 0)

	// Flash clock from the IO MUX (function 1), the rest through the GPIO matrix (function 2).
	ioMuxFunction(&esp.IO_MUX.GPIO6, 1)
	ioMuxFunction(&esp.IO_MUX.GPIO11, 2)
	ioMuxFunction(&esp.IO_MUX.GPIO16, 2)
	ioMuxFunction(&esp.IO_MUX.GPIO17, 2)
	ioMuxFunction(&esp.IO_MUX.GPIO7, 2)
	ioMuxFunction(&esp.IO_MUX.GPIO8, 2)
	ioMuxFunction(&esp.IO_MUX.GPIO9, 2)
	ioMuxFunction(&esp.IO_MUX.GPIO10, 2)
}

// ioMuxFunction and ioMuxDrive use volatile.LoadUint32 and StoreUint32. In a test
// build Register32.Get and Set were not inlined here, so they ran from flash.
//
//go:section .iram.psram
func ioMuxFunction(reg *volatile.Register32, function uint32) {
	v := volatile.LoadUint32(&reg.Reg)&^esp.IO_MUX_GPIO16_MCU_SEL_Msk | function<<esp.IO_MUX_GPIO16_MCU_SEL_Pos
	volatile.StoreUint32(&reg.Reg, v)
}

//go:section .iram.psram
func ioMuxDrive(reg *volatile.Register32, drive uint32) {
	v := volatile.LoadUint32(&reg.Reg)&^esp.IO_MUX_GPIO16_FUN_DRV_Msk | drive<<esp.IO_MUX_GPIO16_FUN_DRV_Pos
	volatile.StoreUint32(&reg.Reg, v)
}

// psramCacheInit sets the PSRAM read and write commands of the cache.
//
//go:section .iram.psram
func psramCacheInit() {
	spi := esp.SPI0
	// PSRAM at 80MHz / 2. DATE bit 31 keeps the flash at 80MHz (not in the TRM, as ESP-IDF).
	spi.SetCLOCK_CLK_EQU_SYSCLK(0)
	spi.SetCLOCK_CLKDIV_PRE(0)
	spi.SetCLOCK_CLKCNT_N(1)
	spi.SetCLOCK_CLKCNT_H(0)
	spi.SetCLOCK_CLKCNT_L(1)
	spi.DATE.SetBits(1 << 31)
	spi.DATE.ClearBits(1 << 30)
	spi.SetCACHE_SCTRL_USR_SRAM_DIO(0)
	spi.SetCACHE_SCTRL_USR_SRAM_QIO(1)
	spi.SetCACHE_SCTRL_CACHE_SRAM_USR_RCMD(1)
	spi.SetCACHE_SCTRL_CACHE_SRAM_USR_WCMD(1)
	spi.SetCACHE_SCTRL_SRAM_ADDR_BITLEN(23)
	spi.SetCACHE_SCTRL_USR_RD_SRAM_DUMMY(1)
	spi.SetSRAM_DRD_CMD_CACHE_SRAM_USR_RD_CMD_BITLEN(7)
	spi.SetSRAM_DRD_CMD_CACHE_SRAM_USR_RD_CMD_VALUE(psramReadQuad)
	spi.SetSRAM_DWR_CMD_CACHE_SRAM_USR_WR_CMD_BITLEN(7)
	spi.SetSRAM_DWR_CMD_CACHE_SRAM_USR_WR_CMD_VALUE(psramWriteQuad)
	spi.SetCACHE_SCTRL_SRAM_DUMMY_CYCLELEN(psramReadWait + matrixWait40)

	// Normal mode with 32KB pages. ESP32 TRM v5.8 table 4.3-14.
	esp.DPORT.SetPRO_CACHE_CTRL_PRO_DRAM_HL(0)
	esp.DPORT.SetPRO_CACHE_CTRL_PRO_DRAM_SPLIT(0)
	esp.DPORT.SetPRO_CACHE_CTRL1_PRO_CACHE_MASK_DRAM1(0)
	esp.DPORT.SetPRO_CACHE_CTRL1_PRO_CACHE_MASK_OPSDRAM(0)
	esp.DPORT.SetPRO_CACHE_CTRL1_PRO_CMMU_SRAM_PAGE_MODE(0)
	spi.SetPIN_CS1_DIS(0) // the cache may now select the PSRAM
}

// psramCmd is one transfer on SPI1. A phase with 0 bits is left out.
type psramCmd struct {
	cmd, cmdBits   uint32
	addr, addrBits uint32
	tx, txBits     uint32
	waitBits       uint32
	rxBits         uint32
	quad           bool
}

// psramCommand sends one transfer to the PSRAM with SPI1 and returns the first
// 64 bits that came back. See ESP-IDF psram_cmd_config and psram_cmd_recv_start.
//
//go:section .iram.psram
func psramCommand(c *psramCmd) (uint32, uint32) {
	spi := esp.SPI1
	for spi.GetCMD_USR() != 0 {
	}
	user, user1, user2, ctrl := spi.USER.Get(), spi.USER1.Get(), spi.USER2.Get(), spi.CTRL.Get()

	spi.SetUSER_USR_COMMAND(onIf(c.cmdBits))
	spi.SetUSER2_USR_COMMAND_BITLEN(minusOne(c.cmdBits))
	spi.SetUSER2_USR_COMMAND_VALUE(c.cmd)
	spi.SetUSER_USR_ADDR(onIf(c.addrBits))
	spi.SetUSER1_USR_ADDR_BITLEN(minusOne(c.addrBits))
	spi.ADDR.Set(c.addr)
	spi.SetUSER_USR_MOSI(onIf(c.txBits))
	spi.SetMOSI_DLEN_USR_MOSI_DBITLEN(minusOne(c.txBits))
	spi.W0.Set(c.tx)
	spi.SetUSER_USR_MISO(onIf(c.rxBits))
	spi.SetMISO_DLEN_USR_MISO_DBITLEN(minusOne(c.rxBits))
	spi.SetUSER_USR_DUMMY(onIf(c.waitBits))
	spi.SetUSER1_USR_DUMMY_CYCLELEN(minusOne(c.waitBits))

	quad := uint32(0)
	if c.quad {
		quad = 1
	}
	spi.SetUSER_FWRITE_QIO(quad)
	spi.SetUSER_FWRITE_DIO(0)
	spi.SetUSER_FWRITE_QUAD(0)
	spi.SetUSER_FWRITE_DUAL(0)
	spi.SetCTRL_FREAD_QIO(quad)
	spi.SetCTRL_FREAD_QUAD(0)
	spi.SetCTRL_FREAD_DUAL(0)
	spi.SetCTRL_FREAD_DIO(0)

	// Select line 1 (PSRAM) instead of line 0 (flash).
	spi.SetPIN_CS1_DIS(0)
	spi.SetPIN_CS0_DIS(1)
	for esp.SPI0.EXT2.Get() != 0 { // the cache shares the wires, so wait until it is idle
	}
	esp.DPORT.HOST_INF_SEL.SetBits(1 << 14) // not in the TRM, as ESP-IDF
	spi.SetCMD_USR(1)
	for spi.GetCMD_USR() != 0 {
	}
	esp.DPORT.HOST_INF_SEL.ClearBits(1 << 14)
	spi.SetPIN_CS1_DIS(1)
	spi.SetPIN_CS0_DIS(0)

	w0, w1 := spi.W0.Get(), spi.W1.Get()
	spi.USER.Set(user)
	spi.USER1.Set(user1)
	spi.USER2.Set(user2)
	spi.CTRL.Set(ctrl)
	return w0, w1
}

//go:section .iram.psram
func onIf(bits uint32) uint32 {
	if bits != 0 {
		return 1
	}
	return 0
}

//go:section .iram.psram
func minusOne(bits uint32) uint32 {
	if bits == 0 {
		return 0
	}
	return bits - 1
}
