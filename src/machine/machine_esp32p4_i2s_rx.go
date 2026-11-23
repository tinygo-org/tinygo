//go:build esp32p4

// ESP32-P4 I2S receive (RX) driver.
//
// Implements DMA-backed I2S RX capture using the selected I2S controller and
// AHB-DMA IN channel 0.  Supports Philips I2S format (standard, MSB-first),
// 16-bit, stereo or mono.  An optional MCLK output is available for codecs
// such as the ES8311.
//
// This is an RX-only driver: it configures the controller as the master and
// generates MCLK/BCLK/LRCLK itself.  It is intended for microphone capture when
// TX is not running.  Simultaneous TX+RX (duplex) needs a shared clock/DMA
// owner and is not implemented yet.
//
// Clock chain:
//
//	XTAL (40 MHz) → HP_SYS_CLKRST fractional divider → MCLK
//	MCLK → RX_BCK_DIV_NUM+1 → BCLK
//	Sample rate = BCLK / 32  (2 channels × 16 bits per stereo frame)
//
// DMA data format: []int16, interleaved L/R samples [L0, R0, L1, R1, ...].
//
// Reference: ESP32-P4 Technical Reference Manual, Chapter I2S; ESP-IDF i2s_ll.h.

package machine

import (
	"device/esp"
	"errors"
	"unsafe"
)

// Errors returned by the I2S RX driver.
var (
	ErrI2SRxBadBits          = errors.New("machine: I2SRx only supports 16 bits per sample")
	ErrI2SRxBadSampleRate    = errors.New("machine: I2SRx sample rate must be non-zero")
	ErrI2SRxMCLKNotDivisible = errors.New("machine: I2SRx MCLK is not an integer multiple of the frame clock")
	ErrI2SRxBadClock         = errors.New("machine: I2SRx cannot derive the requested clock")
	ErrI2SRxBufferTooLarge   = errors.New("machine: I2SRx buffer exceeds one DMA descriptor")
	ErrI2SRxTimeout          = errors.New("machine: I2SRx capture timed out")
)

// I2SRxConfig holds the parameters for an I2S receive session.
type I2SRxConfig struct {
	Controller    I2SController // I2SController0 or I2SController1; default I2S1
	SCK           Pin           // Bit clock output (BCLK) — e.g. I2S_BCLK
	WS            Pin           // Word-select / LRCLK output — e.g. I2S_LRCLK
	SDI           Pin           // Serial data input — e.g. I2S_DIN
	MCLK          Pin           // Optional master-clock output — e.g. I2S_MCLK
	MCLKFrequency uint32        // MCLK frequency in Hz; 0 disables MCLK output
	SampleRate    uint32        // Sample rate in Hz (e.g. 16000, 44100)
	BitsPerSample uint8         // Must be 16
	Mono          bool          // If true, capture only the first slot
}

// I2SRx is a DMA-backed I2S receive driver using AHB-DMA IN channel 0.
// Create with NewI2SRx; call Close when done.  Do not run concurrently with
// I2STx: both own the controller clock tree.
type I2SRx struct {
	hw      *esp.I2S_Type
	ctrl    I2SController
	hasMCLK bool
}

// NewI2SRx initialises the selected I2S RX controller with the given
// configuration and returns a ready-to-use driver.  cfg.BitsPerSample must be 16.
func NewI2SRx(cfg I2SRxConfig) (*I2SRx, error) {
	if cfg.BitsPerSample != 16 {
		return nil, ErrI2SRxBadBits
	}
	if cfg.SampleRate == 0 {
		return nil, ErrI2SRxBadSampleRate
	}
	hw := i2sHardware(cfg.Controller)
	rx := &I2SRx{hw: hw.dev, ctrl: hw.ctrl}

	// ── I2S APB clock + reset ──────────────────────────────────────────────
	hw.enableAPB(true)
	hw.resetAPB()

	// ── Clock tree ─────────────────────────────────────────────────────────
	// The RX module needs its own clock even in master mode.  When MCLK is
	// requested, MCLK is bound to the RX module and routed to cfg.MCLK.
	var bckDiv uint8
	if cfg.MCLKFrequency != 0 && cfg.MCLK != NoPin {
		frameHz := cfg.SampleRate * 32 // 2 channels × 16 bits
		if cfg.MCLKFrequency%frameHz != 0 {
			return nil, ErrI2SRxMCLKNotDivisible
		}
		bd := cfg.MCLKFrequency / frameHz
		if bd < 2 || bd > 63 {
			return nil, ErrI2SRxMCLKNotDivisible
		}
		bckDiv = uint8(bd)

		actual, ok := i2sSetRXClockAndMCLK(hw, cfg.MCLKFrequency)
		if !ok {
			return nil, ErrI2SRxBadClock
		}
		rx.hasMCLK = true
		println("i2s rx init: controller=", hw.name, "MCLK target=", cfg.MCLKFrequency,
			"actual=", actual, "bckDiv=", bckDiv)
	} else {
		divN, bd := i2sBestDividers(i2sXtalHz, cfg.SampleRate*32)
		bckDiv = bd
		if _, ok := i2sSetRXClock(hw, i2sXtalHz/uint32(divN)); !ok {
			return nil, ErrI2SRxBadClock
		}
		println("i2s rx init: controller=", hw.name, "divN=", divN, "bckDiv=", bckDiv)
	}

	// ── Reset RX FIFO and datapath ─────────────────────────────────────────
	rx.hw.SetRX_CONF_RX_FIFO_RESET(1)
	rx.hw.SetRX_CONF_RX_FIFO_RESET(0)
	rx.hw.SetRX_CONF_RX_RESET(1)
	rx.hw.SetRX_CONF_RX_RESET(0)

	// ── RX_CONF: format control, BCK divisor, master mode ──────────────────
	// RX_PCM_BYPASS [12]      = 1  bypass PCM A-law/μ-law codec
	// RX_MSB_SHIFT  [13]      = 1  Philips I2S: data lags WS by 1 BCLK
	// RX_LEFT_ALIGN [15]      = 1  MSB in leftmost bit of the slot
	// RX_TDM_EN     [19]      = 1  TDM mode (required on ESP32-P4)
	// RX_BCK_DIV_NUM [26:21]  = bckDiv − 1  (0-indexed BCLK divisor from MCLK)
	// RX_MONO       [6]       = 1  (mono only)
	rxConf := uint32(0) |
		(1 << 12) | // RX_PCM_BYPASS
		(1 << 13) | // RX_MSB_SHIFT (Philips I2S)
		(1 << 15) | // RX_LEFT_ALIGN
		(1 << 19) | // RX_TDM_EN
		(uint32(bckDiv-1) << 21) // RX_BCK_DIV_NUM (0-indexed)
	if cfg.Mono {
		rxConf |= 1 << 6 // RX_MONO
	}
	rx.hw.RX_CONF.Set(rxConf)

	// ── RX_CONF1: sample-format fields (all 0-indexed: store value − 1) ───
	const bitsM1 = uint32(15) // 16 − 1
	rx.hw.RX_CONF1.Set(
		bitsM1 | // RX_TDM_WS_WIDTH [8:0]
			(bitsM1 << 14) | // RX_BITS_MOD [18:14]
			(bitsM1 << 19) | // RX_HALF_SAMPLE_BITS [26:19]
			(bitsM1 << 27), // RX_TDM_CHAN_BITS [31:27]
	)

	// ── RX_TDM_CTRL: active channel mask and total channel count ───────────
	if cfg.Mono {
		rx.hw.RX_TDM_CTRL.Set(0x1 | (0 << 16))
	} else {
		rx.hw.RX_TDM_CTRL.Set(0x3 | (1 << 16))
	}

	// Commit RX_CONF / RX_CONF1 / RX_TDM_CTRL shadow registers.
	i2sRxUpdate(rx.hw)

	// ── GPIO signal routing ────────────────────────────────────────────────
	i2sConnectOutPinP4(cfg.SCK, hw.rxBckOutIdx)
	i2sConnectOutPinP4(cfg.WS, hw.rxWsOutIdx)
	if rx.hasMCLK {
		i2sConnectOutPinP4(cfg.MCLK, hw.mclkOutIdx)
	}
	i2sConnectInPinP4(cfg.SDI, hw.sdInIdx)

	println("i2s rx: outsel BCLKpin=", cfg.SCK.outFunc().Get(),
		"LRCLKpin=", cfg.WS.outFunc().Get(),
		"SDINpin=", inFunc(hw.sdInIdx).Get())
	if rx.hasMCLK {
		println("i2s rx: outsel MCLKpin=", cfg.MCLK.outFunc().Get())
	}

	// ── AHB-DMA IN channel setup ───────────────────────────────────────────
	esp.AHB_DMA.SetMISC_CONF_CLK_EN(1)
	esp.AHB_DMA.SetIN_PERI_SEL_CH0_PERI_IN_SEL_CH(uint32(hw.dmaPeriph))
	esp.AHB_DMA.SetIN_CONF0_CH0_IN_RST_CH(1)
	esp.AHB_DMA.SetIN_CONF0_CH0_IN_RST_CH(0)

	return rx, nil
}

// Read captures interleaved int16 samples into samples and blocks until the
// DMA descriptor is full or the timeout expires.
//
// Stereo: samples are interleaved [L0, R0, L1, R1, ...].
// Mono:   the first slot is captured; hardware discards the second slot.
//
// len(samples)*2 must not exceed 4092 bytes (one DMA descriptor).  The buffer
// must reside in internal SRAM and remain valid for the duration of the call.
func (rx *I2SRx) Read(samples []int16) error {
	if len(samples) == 0 {
		return nil
	}
	nb := len(samples) * 2
	if nb > maxI2SDescBytes {
		return ErrI2SRxBufferTooLarge
	}

	// Reset the RX FIFO so stale samples from a previous read are discarded.
	rx.hw.SetRX_CONF_RX_FIFO_RESET(1)
	rx.hw.SetRX_CONF_RX_FIFO_RESET(0)

	// The I2S RX FIFO raises in_suc_eof after rx_eof_num bytes.
	rx.hw.SetRXEOF_NUM_RX_EOF_NUM(uint32(nb))

	desc := i2sDMADesc{
		Config: i2sDMADescCfg(uint16(nb), true),
		Buf:    uint32(uintptr(unsafe.Pointer(&samples[0]))),
		Next:   0,
	}

	// Clear any pending EOF and reset the DMA IN channel.
	esp.AHB_DMA.SetIN_INT_CLR_CH0_IN_SUC_EOF_CH_INT_CLR(1)
	esp.AHB_DMA.SetIN_CONF0_CH0_IN_RST_CH(1)
	esp.AHB_DMA.SetIN_CONF0_CH0_IN_RST_CH(0)

	// Flush the descriptor before the DMA reads it.  The RX buffer itself is
	// invalidated after the transfer so the CPU does not read stale cache
	// lines.
	CacheWriteBack(uintptr(unsafe.Pointer(&desc)), unsafe.Sizeof(desc))
	descAddr := uint32(uintptr(unsafe.Pointer(&desc)))
	println("i2s rx: dma start addr=", descAddr, "bytes=", nb)
	esp.AHB_DMA.SetIN_LINK_ADDR_CH0(descAddr)
	esp.AHB_DMA.SetIN_LINK_CH0_INLINK_START_CH(1)

	// Start RX: commit shadow registers first, then set rx_start.
	i2sRxUpdate(rx.hw)
	rx.hw.SetRX_CONF_RX_START(1)

	// Poll for the descriptor to fill, with a 1 second timeout.
	deadline := i2sTicks() + 40_000*1000
	for i2sInEOFRaw() == 0 {
		if i2sTicks() > deadline {
			rx.hw.SetRX_CONF_RX_START(0)
			i2sRxUpdate(rx.hw)
			esp.AHB_DMA.SetIN_LINK_CH0_INLINK_STOP_CH(1)
			return ErrI2SRxTimeout
		}
	}
	esp.AHB_DMA.SetIN_INT_CLR_CH0_IN_SUC_EOF_CH_INT_CLR(1)

	// The DMA wrote the samples directly to L2MEM, bypassing the CPU L1
	// cache.  Invalidate the range before the caller reads it.
	CacheInvalidate(uintptr(unsafe.Pointer(&samples[0])), uintptr(nb))

	// Stop RX and the DMA link.
	rx.hw.SetRX_CONF_RX_START(0)
	i2sRxUpdate(rx.hw)
	esp.AHB_DMA.SetIN_LINK_CH0_INLINK_STOP_CH(1)
	return nil
}

// Close stops capture and disables the RX clock.  The APB clock is left
// enabled because a TX session may share it.
func (rx *I2SRx) Close() {
	rx.hw.SetRX_CONF_RX_START(0)
	i2sRxUpdate(rx.hw)
	esp.AHB_DMA.SetIN_LINK_CH0_INLINK_STOP_CH(1)
	hw := i2sHardware(rx.ctrl)
	hw.setRXClockEnable(false)
	if rx.hasMCLK {
		hw.enableMCLKPad(false)
	}
}

// i2sSetRXClock configures the selected RX module clock from XTAL to targetHz
// using the same fractional divider representation as the TX path.  It returns
// the achieved RX clock frequency.
func i2sSetRXClock(hw i2sHw, targetHz uint32) (uint32, bool) {
	return i2sSetRXClockMode(hw, targetHz, false)
}

// i2sSetRXClockAndMCLK configures the RX clock and binds MCLK to the RX
// module, matching ESP-IDF's RX-only clock routing.
func i2sSetRXClockAndMCLK(hw i2sHw, targetHz uint32) (uint32, bool) {
	return i2sSetRXClockMode(hw, targetHz, true)
}

func i2sSetRXClockMode(hw i2sHw, targetHz uint32, bindMCLK bool) (uint32, bool) {
	divN, divX, divY, divZ, divYN1, actual := i2sFracDivider(i2sXtalHz, targetHz)
	if divN == 0 {
		return 0, false
	}
	println("i2s rx: clock controller=", hw.name, "divN=", divN, "x=", divX, "y=", divY, "z=", divZ, "yn1=", divYN1, "actual=", actual)

	// Match ESP-IDF's order: enable the module clock, bind MCLK to RX, set the
	// source, then program the fractional divider.
	hw.setRXClockEnable(true)
	if bindMCLK {
		hw.setMSTClkRX()
	}
	hw.setRXClockSrcXTAL()

	// Same double-division workaround as the TX path.
	hw.setRXDivN(2)
	hw.setRXDivYN1(0)
	hw.setRXDivY(1)
	hw.setRXDivZ(0)
	hw.setRXDivX(0)
	hw.setRXDivYN1(uint32(divYN1))
	hw.setRXDivZ(uint32(divZ))
	hw.setRXDivY(uint32(divY))
	hw.setRXDivX(uint32(divX))
	hw.setRXDivN(uint32(divN))

	if bindMCLK {
		hw.enableMCLKPad(true)
	}
	return actual, true
}

// i2sConnectInPinP4 routes a GPIO pin to an I2S input signal via the GPIO
// matrix on ESP32-P4.
func i2sConnectInPinP4(p Pin, sig uint32) {
	if p == NoPin {
		return
	}
	p.Configure(PinConfig{Mode: PinInput})
	inFunc(sig).Set(esp.GPIO_FUNC_IN_SEL_CFG_SEL | uint32(p))
}

// i2sRxUpdate commits RX_CONF / RX_CONF1 / RX_TDM_CTRL shadow registers to the
// active I2S hardware registers.  RX_UPDATE is a self-clearing strobe.
func i2sRxUpdate(hw *esp.I2S_Type) {
	hw.SetRX_CONF_RX_UPDATE(1)
	for hw.GetRX_CONF_RX_UPDATE() != 0 {
	}
}

// i2sInEOFRaw returns the raw IN_SUC_EOF flag for AHB-DMA IN channel 0.
// Marked noinline so LLVM LTO cannot hoist the volatile load out of the poll.
//
//go:noinline
func i2sInEOFRaw() uint32 {
	return esp.AHB_DMA.GetIN_INT_RAW_CH0_IN_SUC_EOF_CH_INT_RAW()
}

// i2sTicks returns the TIMG0 64-bit counter value (40 ticks per microsecond).
func i2sTicks() uint64 {
	esp.TIMG0.T0UPDATE.Set(1)
	for esp.TIMG0.T0UPDATE.Get() != 0 {
	}
	return uint64(esp.TIMG0.T0LO.Get()) | uint64(esp.TIMG0.T0HI.Get())<<32
}
