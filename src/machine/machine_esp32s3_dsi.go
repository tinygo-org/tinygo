//go:build esp32s3

// ESP32-S3 DSI support using the LCD_CAM peripheral and GDMA for framebuffer streaming.

package machine

import (
	"device/esp"
	"runtime/volatile"
	"unsafe"
)

// GPIO MUX signal numbers for LCD_CAM.
const (
	lcdSignalDE    = 150
	lcdSignalHSYNC = 151
	lcdSignalVSYNC = 152
	lcdSignalPCLK  = 154
	lcdSignalDBase = 133
)

// GDMA descriptor chain entry, can only stream up to 4095 bytes per descriptor.
type llDMADescriptor struct {
	flags uint32
	buf   unsafe.Pointer
	next  *llDMADescriptor
}

// Enough to drive the framebuffer of up to 614_400 bytes (up to 640x480 with RGB565 format)
const maxLCDDMADesc = 160

// DSI implements generic DSIStreamer for ESP32-S3 video hardware. It is not instantiated immediately because even the
// unconfigured object consumes ~2K RAM for DMA descripors which is rather wasteful if the application doesn't need it.
// If needed, instantiate using `dsi := &DSI{}`
type DSI struct {
	framebuf     unsafe.Pointer
	framebufLen  uint32
	timings      VideoTimings
	pixelFormat  PixelFormat
	dmaDesc      [maxLCDDMADesc]llDMADescriptor
	dmaDescCount uint32
}

func (d *DSI) WriteShort(dt uint8, data0, data1 uint8) error {
	return errDSINotImplemented
}

func (d *DSI) WriteLong(dt uint8, payload []byte) error {
	return errDSINotImplemented
}

func (d *DSI) ReadShort(dt uint8) ([2]byte, error) {
	return [2]byte{0, 0}, errDSINotImplemented
}

// InitVideo configures D-PHY clocking, active lanes, and video timing parameters.
func (d *DSI) InitVideo(lanes LaneConfig, timings VideoTimings) error {
	d.timings = timings
	d.pixelFormat = lanes.PixelFormat
	d.initLCD()
	d.routePins(lanes)
	d.configureTimings(timings, lanes.PixelFormat)
	return nil
}

// StartVideo starts continuous DMA transmission over D-PHY lanes.
func (d *DSI) StartVideo() error {
	if d.framebuf == nil {
		return errDSIFramebufferError
	}

	d.startDMA(d.framebuf, d.framebufLen)
	return nil
}

// StopVideo halts video packet generation (e.g., prior to panel sleep mode).
func (d *DSI) StopVideo() error {
	return errDSINotImplemented
}

// SetFramebuffer assigns the primary buffer for continuous DMA reading. Must be called prior to StartVideo().
func (d *DSI) SetFramebuffer(primary []byte) error {
	if len(primary) == 0 {
		return errDSIFramebufferError
	}

	d.framebuf = unsafe.Pointer(&primary[0])
	d.framebufLen = uint32(len(primary))
	return nil
}

// SwapFramebuffer queues a new buffer address for the DMA engine.
func (d *DSI) SwapFramebuffer(next []byte) error {
	if len(next) == 0 {
		return errDSIFramebufferError
	}

	d.framebuf = unsafe.Pointer(&next[0])
	d.framebufLen = uint32(len(next))
	d.waitVSync()
	d.updateDMABuf(d.framebuf)
	return nil
}

func (d *DSI) waitVSync() {
	esp.LCD_CAM.SetLC_DMA_INT_RAW_LCD_VSYNC_INT_RAW(1)
	for i := 0; i < 1000000; i++ {
		if esp.LCD_CAM.GetLC_DMA_INT_RAW_LCD_VSYNC_INT_RAW() != 0 {
			return
		}
	}
}

func (d *DSI) buildDMADescriptorFlags(n, offset uint32, eof, owner bool) uint32 {
	var f uint32 = (n & 0xFFF) | ((n & 0xFFF) << 12) | ((offset & 0x3F) << 24)
	if eof {
		f |= 1 << 30
	}
	if owner {
		f |= 1 << 31
	}
	return f
}

func (d *DSI) setupPin(gpio Pin, signalIdx uint32) {
	switch gpio {
	case GPIO0:
		esp.IO_MUX.SetGPIO0_MCU_SEL(1)
		esp.IO_MUX.SetGPIO0_FUN_DRV(2)
		esp.GPIO.SetFUNC0_OUT_SEL_CFG_OUT_SEL(signalIdx)
	case GPIO1:
		esp.IO_MUX.SetGPIO1_MCU_SEL(1)
		esp.IO_MUX.SetGPIO1_FUN_DRV(2)
		esp.GPIO.SetFUNC1_OUT_SEL_CFG_OUT_SEL(signalIdx)
	case GPIO2:
		esp.IO_MUX.SetGPIO2_MCU_SEL(1)
		esp.IO_MUX.SetGPIO2_FUN_DRV(2)
		esp.GPIO.SetFUNC2_OUT_SEL_CFG_OUT_SEL(signalIdx)
	case GPIO3:
		esp.IO_MUX.SetGPIO3_MCU_SEL(1)
		esp.IO_MUX.SetGPIO3_FUN_DRV(2)
		esp.GPIO.SetFUNC3_OUT_SEL_CFG_OUT_SEL(signalIdx)
	case GPIO4:
		esp.IO_MUX.SetGPIO4_MCU_SEL(1)
		esp.IO_MUX.SetGPIO4_FUN_DRV(2)
		esp.GPIO.SetFUNC4_OUT_SEL_CFG_OUT_SEL(signalIdx)
	case GPIO5:
		esp.IO_MUX.SetGPIO5_MCU_SEL(1)
		esp.IO_MUX.SetGPIO5_FUN_DRV(2)
		esp.GPIO.SetFUNC5_OUT_SEL_CFG_OUT_SEL(signalIdx)
	case GPIO6:
		esp.IO_MUX.SetGPIO6_MCU_SEL(1)
		esp.IO_MUX.SetGPIO6_FUN_DRV(2)
		esp.GPIO.SetFUNC6_OUT_SEL_CFG_OUT_SEL(signalIdx)
	case GPIO7:
		esp.IO_MUX.SetGPIO7_MCU_SEL(1)
		esp.IO_MUX.SetGPIO7_FUN_DRV(2)
		esp.GPIO.SetFUNC7_OUT_SEL_CFG_OUT_SEL(signalIdx)
	case GPIO8:
		esp.IO_MUX.SetGPIO8_MCU_SEL(1)
		esp.IO_MUX.SetGPIO8_FUN_DRV(2)
		esp.GPIO.SetFUNC8_OUT_SEL_CFG_OUT_SEL(signalIdx)
	case GPIO9:
		esp.IO_MUX.SetGPIO9_MCU_SEL(1)
		esp.IO_MUX.SetGPIO9_FUN_DRV(2)
		esp.GPIO.SetFUNC9_OUT_SEL_CFG_OUT_SEL(signalIdx)
	case GPIO10:
		esp.IO_MUX.SetGPIO10_MCU_SEL(1)
		esp.IO_MUX.SetGPIO10_FUN_DRV(2)
		esp.GPIO.SetFUNC10_OUT_SEL_CFG_OUT_SEL(signalIdx)
	case GPIO11:
		esp.IO_MUX.SetGPIO11_MCU_SEL(1)
		esp.IO_MUX.SetGPIO11_FUN_DRV(2)
		esp.GPIO.SetFUNC11_OUT_SEL_CFG_OUT_SEL(signalIdx)
	case GPIO12:
		esp.IO_MUX.SetGPIO12_MCU_SEL(1)
		esp.IO_MUX.SetGPIO12_FUN_DRV(2)
		esp.GPIO.SetFUNC12_OUT_SEL_CFG_OUT_SEL(signalIdx)
	case GPIO13:
		esp.IO_MUX.SetGPIO13_MCU_SEL(1)
		esp.IO_MUX.SetGPIO13_FUN_DRV(2)
		esp.GPIO.SetFUNC13_OUT_SEL_CFG_OUT_SEL(signalIdx)
	case GPIO14:
		esp.IO_MUX.SetGPIO14_MCU_SEL(1)
		esp.IO_MUX.SetGPIO14_FUN_DRV(2)
		esp.GPIO.SetFUNC14_OUT_SEL_CFG_OUT_SEL(signalIdx)
	case GPIO15:
		esp.IO_MUX.SetGPIO15_MCU_SEL(1)
		esp.IO_MUX.SetGPIO15_FUN_DRV(2)
		esp.GPIO.SetFUNC15_OUT_SEL_CFG_OUT_SEL(signalIdx)
	case GPIO16:
		esp.IO_MUX.SetGPIO16_MCU_SEL(1)
		esp.IO_MUX.SetGPIO16_FUN_DRV(2)
		esp.GPIO.SetFUNC16_OUT_SEL_CFG_OUT_SEL(signalIdx)
	case GPIO17:
		esp.IO_MUX.SetGPIO17_MCU_SEL(1)
		esp.IO_MUX.SetGPIO17_FUN_DRV(2)
		esp.GPIO.SetFUNC17_OUT_SEL_CFG_OUT_SEL(signalIdx)
	case GPIO18:
		esp.IO_MUX.SetGPIO18_MCU_SEL(1)
		esp.IO_MUX.SetGPIO18_FUN_DRV(2)
		esp.GPIO.SetFUNC18_OUT_SEL_CFG_OUT_SEL(signalIdx)
	case GPIO19:
		esp.IO_MUX.SetGPIO19_MCU_SEL(1)
		esp.IO_MUX.SetGPIO19_FUN_DRV(2)
		esp.GPIO.SetFUNC19_OUT_SEL_CFG_OUT_SEL(signalIdx)
	case GPIO20:
		esp.IO_MUX.SetGPIO20_MCU_SEL(1)
		esp.IO_MUX.SetGPIO20_FUN_DRV(2)
		esp.GPIO.SetFUNC20_OUT_SEL_CFG_OUT_SEL(signalIdx)
	case GPIO21:
		esp.IO_MUX.SetGPIO21_MCU_SEL(1)
		esp.IO_MUX.SetGPIO21_FUN_DRV(2)
		esp.GPIO.SetFUNC21_OUT_SEL_CFG_OUT_SEL(signalIdx)
	case Pin(22):
		esp.IO_MUX.SetGPIO22_MCU_SEL(1)
		esp.IO_MUX.SetGPIO22_FUN_DRV(2)
		esp.GPIO.SetFUNC22_OUT_SEL_CFG_OUT_SEL(signalIdx)
	case Pin(23):
		esp.IO_MUX.SetGPIO23_MCU_SEL(1)
		esp.IO_MUX.SetGPIO23_FUN_DRV(2)
		esp.GPIO.SetFUNC23_OUT_SEL_CFG_OUT_SEL(signalIdx)
	case Pin(24):
		esp.IO_MUX.SetGPIO24_MCU_SEL(1)
		esp.IO_MUX.SetGPIO24_FUN_DRV(2)
		esp.GPIO.SetFUNC24_OUT_SEL_CFG_OUT_SEL(signalIdx)
	case Pin(25):
		esp.IO_MUX.SetGPIO25_MCU_SEL(1)
		esp.IO_MUX.SetGPIO25_FUN_DRV(2)
		esp.GPIO.SetFUNC25_OUT_SEL_CFG_OUT_SEL(signalIdx)
	case GPIO26:
		esp.IO_MUX.SetGPIO26_MCU_SEL(1)
		esp.IO_MUX.SetGPIO26_FUN_DRV(2)
		esp.GPIO.SetFUNC26_OUT_SEL_CFG_OUT_SEL(signalIdx)
	case GPIO27:
		esp.IO_MUX.SetGPIO27_MCU_SEL(1)
		esp.IO_MUX.SetGPIO27_FUN_DRV(2)
		esp.GPIO.SetFUNC27_OUT_SEL_CFG_OUT_SEL(signalIdx)
	case GPIO28:
		esp.IO_MUX.SetGPIO28_MCU_SEL(1)
		esp.IO_MUX.SetGPIO28_FUN_DRV(2)
		esp.GPIO.SetFUNC28_OUT_SEL_CFG_OUT_SEL(signalIdx)
	case GPIO29:
		esp.IO_MUX.SetGPIO29_MCU_SEL(1)
		esp.IO_MUX.SetGPIO29_FUN_DRV(2)
		esp.GPIO.SetFUNC29_OUT_SEL_CFG_OUT_SEL(signalIdx)
	case GPIO30:
		esp.IO_MUX.SetGPIO30_MCU_SEL(1)
		esp.IO_MUX.SetGPIO30_FUN_DRV(2)
		esp.GPIO.SetFUNC30_OUT_SEL_CFG_OUT_SEL(signalIdx)
	case GPIO31:
		esp.IO_MUX.SetGPIO31_MCU_SEL(1)
		esp.IO_MUX.SetGPIO31_FUN_DRV(2)
		esp.GPIO.SetFUNC31_OUT_SEL_CFG_OUT_SEL(signalIdx)
	case GPIO32:
		esp.IO_MUX.SetGPIO32_MCU_SEL(1)
		esp.IO_MUX.SetGPIO32_FUN_DRV(2)
		esp.GPIO.SetFUNC32_OUT_SEL_CFG_OUT_SEL(signalIdx)
	case GPIO33:
		esp.IO_MUX.SetGPIO33_MCU_SEL(1)
		esp.IO_MUX.SetGPIO33_FUN_DRV(2)
		esp.GPIO.SetFUNC33_OUT_SEL_CFG_OUT_SEL(signalIdx)
	case GPIO34:
		esp.IO_MUX.SetGPIO34_MCU_SEL(1)
		esp.IO_MUX.SetGPIO34_FUN_DRV(2)
		esp.GPIO.SetFUNC34_OUT_SEL_CFG_OUT_SEL(signalIdx)
	case GPIO35:
		esp.IO_MUX.SetGPIO35_MCU_SEL(1)
		esp.IO_MUX.SetGPIO35_FUN_DRV(2)
		esp.GPIO.SetFUNC35_OUT_SEL_CFG_OUT_SEL(signalIdx)
	case GPIO36:
		esp.IO_MUX.SetGPIO36_MCU_SEL(1)
		esp.IO_MUX.SetGPIO36_FUN_DRV(2)
		esp.GPIO.SetFUNC36_OUT_SEL_CFG_OUT_SEL(signalIdx)
	case GPIO37:
		esp.IO_MUX.SetGPIO37_MCU_SEL(1)
		esp.IO_MUX.SetGPIO37_FUN_DRV(2)
		esp.GPIO.SetFUNC37_OUT_SEL_CFG_OUT_SEL(signalIdx)
	case GPIO38:
		esp.IO_MUX.SetGPIO38_MCU_SEL(1)
		esp.IO_MUX.SetGPIO38_FUN_DRV(2)
		esp.GPIO.SetFUNC38_OUT_SEL_CFG_OUT_SEL(signalIdx)
	case GPIO39:
		esp.IO_MUX.SetGPIO39_MCU_SEL(1)
		esp.IO_MUX.SetGPIO39_FUN_DRV(2)
		esp.GPIO.SetFUNC39_OUT_SEL_CFG_OUT_SEL(signalIdx)
	case GPIO40:
		esp.IO_MUX.SetGPIO40_MCU_SEL(1)
		esp.IO_MUX.SetGPIO40_FUN_DRV(2)
		esp.GPIO.SetFUNC40_OUT_SEL_CFG_OUT_SEL(signalIdx)
	case GPIO41:
		esp.IO_MUX.SetGPIO41_MCU_SEL(1)
		esp.IO_MUX.SetGPIO41_FUN_DRV(2)
		esp.GPIO.SetFUNC41_OUT_SEL_CFG_OUT_SEL(signalIdx)
	case GPIO42:
		esp.IO_MUX.SetGPIO42_MCU_SEL(1)
		esp.IO_MUX.SetGPIO42_FUN_DRV(2)
		esp.GPIO.SetFUNC42_OUT_SEL_CFG_OUT_SEL(signalIdx)
	case GPIO43:
		esp.IO_MUX.SetGPIO43_MCU_SEL(1)
		esp.IO_MUX.SetGPIO43_FUN_DRV(2)
		esp.GPIO.SetFUNC43_OUT_SEL_CFG_OUT_SEL(signalIdx)
	case GPIO44:
		esp.IO_MUX.SetGPIO44_MCU_SEL(1)
		esp.IO_MUX.SetGPIO44_FUN_DRV(2)
		esp.GPIO.SetFUNC44_OUT_SEL_CFG_OUT_SEL(signalIdx)
	case GPIO45:
		esp.IO_MUX.SetGPIO45_MCU_SEL(1)
		esp.IO_MUX.SetGPIO45_FUN_DRV(2)
		esp.GPIO.SetFUNC45_OUT_SEL_CFG_OUT_SEL(signalIdx)
	case GPIO46:
		esp.IO_MUX.SetGPIO46_MCU_SEL(1)
		esp.IO_MUX.SetGPIO46_FUN_DRV(2)
		esp.GPIO.SetFUNC46_OUT_SEL_CFG_OUT_SEL(signalIdx)
	case GPIO47:
		esp.IO_MUX.SetGPIO47_MCU_SEL(1)
		esp.IO_MUX.SetGPIO47_FUN_DRV(2)
		esp.GPIO.SetFUNC47_OUT_SEL_CFG_OUT_SEL(signalIdx)
	case GPIO48:
		esp.IO_MUX.SetGPIO48_MCU_SEL(1)
		esp.IO_MUX.SetGPIO48_FUN_DRV(2)
		esp.GPIO.SetFUNC48_OUT_SEL_CFG_OUT_SEL(signalIdx)
	}

	if gpio >= 32 {
		volatile.StoreUint32(&esp.GPIO.ENABLE1_W1TS.Reg, 1<<(gpio-32))
	} else {
		volatile.StoreUint32(&esp.GPIO.ENABLE_W1TS.Reg, 1<<gpio)
	}
}

func (d *DSI) initLCD() {
	esp.SYSTEM.SetPERIP_CLK_EN1_LCD_CAM_CLK_EN(1)
	esp.SYSTEM.SetPERIP_RST_EN1_LCD_CAM_RST(1)
	esp.SYSTEM.SetPERIP_RST_EN1_LCD_CAM_RST(0)
}

func (d *DSI) routePins(cfg LaneConfig) {
	pins := cfg.Pins

	d.setupPin(pins.PCLK, lcdSignalPCLK)
	d.setupPin(pins.VSYNC, lcdSignalVSYNC)
	d.setupPin(pins.HSYNC, lcdSignalHSYNC)
	d.setupPin(pins.DE, lcdSignalDE)

	ndata := uint32(16)
	if cfg.PixelFormat == PixelFormatRGB888 {
		ndata = 24
	}
	for i := uint32(0); i < ndata; i++ {
		gpio := pins.Data[i]
		if gpio == NoPin {
			continue
		}
		d.setupPin(gpio, lcdSignalDBase+i)
	}
}

func (d *DSI) configureTimings(timings VideoTimings, pixFmt PixelFormat) {
	// f_LCD_CLK = 160_000_000 / (CLKM_DIV_NUM + (CLKM_DIV_B / CLKM_DIV_A))
	// f_LCD_PCLK = f_LCD_CLK / (reg_clkcnt_N + 1)
	// f_LCD_PCLK = 160_000_000 / 2 / (9 + 1) = 8 MHz pixel clock

	// Naive calculation of target pixel clock. We can make it more accurate by solving for
	// CLKM_DIV_NUM, CLKM_DIV_B and CLKM_DIV_A to arrive at the exact requested value, but
	// this in practice is not needed because in most cases we don't need accurate frame rate.
	div := (80_000_000 / timings.PCLK)
	if div < 1 {
		div = 1
	}
	if div > 64 {
		div = 64
	}

	esp.LCD_CAM.SetLCD_CLOCK_LCD_CLK_EQU_SYSCLK(0) // Enable clock divisor.
	esp.LCD_CAM.SetLCD_CLOCK_LCD_CLK_SEL(2)        // Clock source 2: PLL_160M
	esp.LCD_CAM.SetLCD_CLOCK_LCD_CLKM_DIV_NUM(1)   // Hardcode values to get to the 80 MHz LCD_CLK
	esp.LCD_CAM.SetLCD_CLOCK_LCD_CLKM_DIV_B(1)
	esp.LCD_CAM.SetLCD_CLOCK_LCD_CLKM_DIV_A(1)
	esp.LCD_CAM.SetLCD_CLOCK_LCD_CLKCNT_N(div - 1)
	esp.LCD_CAM.SetLCD_CLOCK_CLK_EN(1)

	hbFront := uint32(timings.HBackPorch + timings.HSyncWidth - 1)
	vaHeight := uint32(timings.VActive - 1)
	vtHeight := uint32(timings.VSyncWidth + timings.VBackPorch + timings.VActive + timings.VFrontPorch - 1)

	esp.LCD_CAM.SetLCD_CTRL_LCD_HB_FRONT(hbFront)
	esp.LCD_CAM.SetLCD_CTRL_LCD_VA_HEIGHT(vaHeight)
	esp.LCD_CAM.SetLCD_CTRL_LCD_VT_HEIGHT(vtHeight)
	esp.LCD_CAM.SetLCD_CTRL_LCD_RGB_MODE_EN(1)

	vbFront := uint32(timings.VBackPorch + timings.VSyncWidth - 1)
	haWidth := uint32(timings.HActive - 1)
	htWidth := uint32(timings.HSyncWidth + timings.HBackPorch + timings.HActive + timings.HFrontPorch - 1)

	esp.LCD_CAM.SetLCD_CTRL1_LCD_VB_FRONT(vbFront)
	esp.LCD_CAM.SetLCD_CTRL1_LCD_HA_WIDTH(haWidth)
	esp.LCD_CAM.SetLCD_CTRL1_LCD_HT_WIDTH(htWidth)

	vsyncWidth := uint32(timings.VSyncWidth - 1)
	hsyncWidth := uint32(timings.HSyncWidth - 1)

	esp.LCD_CAM.SetLCD_CTRL2_LCD_VSYNC_WIDTH(vsyncWidth)
	esp.LCD_CAM.SetLCD_CTRL2_LCD_VSYNC_IDLE_POL(1)
	esp.LCD_CAM.SetLCD_CTRL2_LCD_HS_BLANK_EN(1)
	esp.LCD_CAM.SetLCD_CTRL2_LCD_HSYNC_WIDTH(hsyncWidth)
	esp.LCD_CAM.SetLCD_CTRL2_LCD_HSYNC_IDLE_POL(1)

	if pixFmt == PixelFormatRGB565 {
		esp.LCD_CAM.SetLCD_USER_LCD_2BYTE_EN(1)
	} else {
		esp.LCD_CAM.SetLCD_USER_LCD_2BYTE_EN(0)
	}

	esp.LCD_CAM.SetLCD_USER_LCD_ALWAYS_OUT_EN(1)
	esp.LCD_CAM.SetLCD_USER_LCD_DOUT(1)
	esp.LCD_CAM.SetLCD_MISC_LCD_NEXT_FRAME_EN(1)
	esp.LCD_CAM.SetLCD_MISC_LCD_BK_EN(1)
}

func (d *DSI) startDMA(buf unsafe.Pointer, bufLen uint32) {
	var descIdx uint32 = 0
	ptr := uintptr(buf)
	rem := bufLen
	isLast := false

	for rem > 0 && descIdx < maxLCDDMADesc {
		chunk := rem
		if chunk > 4092 {
			chunk = 4092 // max 12-bit size, word-aligned; 2046 RGB565 or 1364 RGB888 pixels
		}
		rem -= chunk

		if rem == 0 || descIdx == maxLCDDMADesc-1 {
			isLast = true
		}

		d.dmaDesc[descIdx].flags = d.buildDMADescriptorFlags(chunk, 0, isLast, true)
		d.dmaDesc[descIdx].buf = unsafe.Pointer(ptr)

		if isLast {
			d.dmaDesc[descIdx].next = &d.dmaDesc[0]
		} else {
			d.dmaDesc[descIdx].next = &d.dmaDesc[descIdx+1]
		}

		descIdx++
		ptr += uintptr(chunk)
	}
	d.dmaDescCount = descIdx

	esp.SYSTEM.SetPERIP_CLK_EN1_DMA_CLK_EN(1)
	esp.SYSTEM.SetPERIP_RST_EN1_DMA_RST(1)
	esp.SYSTEM.SetPERIP_RST_EN1_DMA_RST(0)

	esp.DMA.SetMISC_CONF_CLK_EN(1)

	esp.DMA.SetOUT_CONF0_CH0_OUT_RST(1)
	esp.DMA.SetOUT_CONF0_CH0_OUT_RST(0)
	esp.DMA.SetOUT_CONF0_CH0_OUTDSCR_BURST_EN(1)
	esp.DMA.SetOUT_CONF0_CH0_OUT_DATA_BURST_EN(1)

	esp.DMA.SetOUT_CONF1_CH0_OUT_CHECK_OWNER(0)

	esp.DMA.SetOUT_PERI_SEL_CH0_PERI_OUT_SEL(5)

	headAddr := uint32(uintptr(unsafe.Pointer(&d.dmaDesc[0])))
	esp.DMA.SetOUT_LINK_CH0_OUTLINK_ADDR((headAddr & 0xFFFFF) | (1 << 21))

	esp.LCD_CAM.SetLCD_MISC_LCD_AFIFO_RESET(1)
	esp.LCD_CAM.SetLCD_USER_LCD_RESET(1)

	for i := 0; i < 1000; i++ {
		_ = i
	}

	esp.LCD_CAM.SetLCD_USER_LCD_UPDATE(1)
	esp.LCD_CAM.SetLCD_USER_LCD_START(1)
}

func (d *DSI) updateDMABuf(buf unsafe.Pointer) {
	if d.dmaDescCount == 0 {
		return
	}
	ptr := uintptr(buf)
	for i := uint32(0); i < d.dmaDescCount; i++ {
		chunk := uintptr(d.dmaDesc[i].flags & 0xFFF)
		d.dmaDesc[i].buf = unsafe.Pointer(ptr)
		ptr += chunk
	}
}
