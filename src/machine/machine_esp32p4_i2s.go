//go:build esp32p4

// ESP32-P4 I2S transmit (TX) driver.
//
// Implements DMA-backed I2S TX playback using the selected I2S controller
// (I2S0 or I2S1) and AHB-DMA channel 0.  Supports Philips I2S format
// (standard, MSB-first), 16-bit, stereo or mono.  An optional MCLK output is
// available for codecs such as the ES8311.
//
// Clock chain (integer-only division; XTAL is the only source on P4):
//
//	XTAL (40 MHz) → HP_SYS_CLKRST fractional divider → MCLK
//	MCLK → TX_BCK_DIV_NUM+1 → BCLK
//	BCLK = 40 MHz / (divN + divFrac) / bckDiv
//	Sample rate = BCLK / 32  (2 channels × 16 bits per stereo frame)
//
// DMA data format: []int16, interleaved L/R samples [L0, R0, L1, R1, ...].
// For mono output, pass single-channel samples; TX_MONO duplicates each sample
// to both I2S output slots so any downstream DAC receives a valid stereo stream.
//
// AHB-DMA channel 0 is used exclusively.  Do not share with other drivers.
//
// Reference: ESP32-P4 Technical Reference Manual, Chapter I2S; ESP-IDF i2s_ll.h.

package machine

import (
	"device/esp"
	"errors"
	"unsafe"
)

// i2sXtalHz is the XTAL frequency; the only I2S clock source on ESP32-P4.
const i2sXtalHz uint32 = 40_000_000

// maxI2SDescBytes is the maximum DMA payload per descriptor (12-bit size field).
const maxI2SDescBytes = 4092

// ErrI2STxBadBits is returned when BitsPerSample != 16.
var ErrI2STxBadBits = errors.New("machine: I2STx only supports 16 bits per sample")

// ErrI2STxBadSampleRate is returned when the sample rate is zero.
var ErrI2STxBadSampleRate = errors.New("machine: I2STx sample rate must be non-zero")

// ErrI2STxMCLKNotDivisible is returned when the requested MCLK cannot produce
// an integer BCLK divider in the range [2, 63].
var ErrI2STxMCLKNotDivisible = errors.New("machine: I2STx MCLK is not an integer multiple of the frame clock")

// ErrI2STxBadClock is returned when no I2S divider can be derived from XTAL.
var ErrI2STxBadClock = errors.New("machine: I2STx cannot derive the requested clock")

// ErrI2STxTimeout is returned when DMA playback does not complete in time.
var ErrI2STxTimeout = errors.New("machine: I2STx playback timed out")

// ErrI2STxBufferTooLarge is returned when a Write needs more descriptors than
// the built-in reusable descriptor pool can hold.
var ErrI2STxBufferTooLarge = errors.New("machine: I2STx buffer exceeds descriptor pool")

// i2sDMADesc is a 12-byte AHB-DMA linked-list descriptor.
// Matches the ESP32 dma_descriptor_t hardware layout (dma_types.h).
type i2sDMADesc struct {
	Config uint32 // [11:0]=byte_size [23:12]=byte_length [30]=suc_eof [31]=owner(1=DMA)
	Buf    uint32 // buffer start address (physical; equals virtual for SRAM on P4)
	Next   uint32 // next descriptor address, 0 if last
}

// i2sDMADescCfg builds the Config word for a DMA descriptor.
// byteLen is both the buffer size and valid byte count (equal for TX).
func i2sDMADescCfg(byteLen uint16, eof bool) uint32 {
	c := uint32(byteLen) | uint32(byteLen)<<12
	if eof {
		c |= 1 << 30 // suc_eof: this is the last descriptor in the chain
	}
	c |= 1 << 31 // owner = DMA (hardware may read the buffer)
	return c
}

// I2STxConfig holds the parameters for an I2S transmit session.
type I2STxConfig struct {
	Controller    I2SController // I2SController0 or I2SController1; default I2S1
	SCK           Pin           // Bit clock output (BCLK) — e.g. I2S_BCLK
	WS            Pin           // Word-select / LRCLK output — e.g. I2S_LRCLK
	SDO           Pin           // Serial data output — e.g. I2S_SDATA
	MCLK          Pin           // Optional master-clock output — e.g. I2S_MCLK
	MCLKFrequency uint32        // MCLK frequency in Hz; 0 disables MCLK output
	SampleRate    uint32        // Sample rate in Hz (e.g. 16000, 44100)
	BitsPerSample uint8         // Must be 16
	Mono          bool          // If true, Write accepts mono samples, duplicated to both outputs
}

// I2STx is a DMA-backed I2S transmit driver using AHB-DMA channel 0.
// Create with NewI2STx; call Close when done.
type I2STx struct {
	hw   *esp.I2S_Type
	ctrl I2SController
	// GPIO.IN bit masks of the configured output pins (diagnostic probe).
	mBclk, mLrclk, mSdata, mMclk uint32
	hasMCLK                      bool
	sampleRate                   uint32
	mono                         bool
	// running is true once TX_START has been set.  The TX engine is kept
	// running across Write calls, matching ESP-IDF's i2s_channel_write path;
	// stopping and restarting it per buffer can leave the FIFO unread.
	running bool
	// descs is a reusable descriptor pool so streaming Write calls do not
	// allocate on every chunk.
	descs [128]i2sDMADesc
}

// NewI2STx initialises the selected I2S TX controller with the given
// configuration and returns a ready-to-use driver.  cfg.BitsPerSample must be 16.
//
// The function enables the APB clock, configures the TX clock chain via
// HP_SYS_CLKRST, sets up the TDM-mode registers, and enables AHB-DMA.
// When MCLKFrequency is non-zero, MCLK is routed to cfg.MCLK and the BCLK
// divider is derived from that MCLK instead of from the sample rate alone.
func NewI2STx(cfg I2STxConfig) (*I2STx, error) {
	if cfg.BitsPerSample != 16 {
		return nil, ErrI2STxBadBits
	}
	if cfg.SampleRate == 0 {
		return nil, ErrI2STxBadSampleRate
	}
	hw := i2sHardware(cfg.Controller)
	tx := &I2STx{hw: hw.dev, ctrl: hw.ctrl, sampleRate: cfg.SampleRate, mono: cfg.Mono}

	// ── I2S APB clock + reset ──────────────────────────────────────────────
	// CrowPanel ESP32-P4 10.1" routes audio (GPIO21/22/23) to I2S1.
	// The WaveShare ESP32-P4-NANO routes the ES8311 to I2S0 (GPIO9-13).
	hw.enableAPB(true)
	hw.resetAPB()

	// ── Reset I2S TX FIFO and datapath ────────────────────────────────────
	// ESP-IDF resets the TX slot before configuring the clock.  Do the same:
	// resetting after the clock/MCLK bind can leave the TX data path idle even
	// though BCLK/WS/MCLK continue to toggle.
	tx.hw.SetTX_CONF_TX_FIFO_RESET(1)
	tx.hw.SetTX_CONF_TX_FIFO_RESET(0)
	tx.hw.SetTX_CONF_TX_RESET(1)
	tx.hw.SetTX_CONF_TX_RESET(0)
	// i2s_ll_tx_enable_std() also clears PDM and the PCM2PDM converter.  A
	// stale PCM2PDM_CONV_EN routes the FIFO away from the I2S serializer.
	tx.hw.SetTX_CONF_TX_PDM_EN(0)
	tx.hw.SetTX_PCM2PDM_CONF_PCM2PDM_CONV_EN(0)

	// ── TX clock ───────────────────────────────────────────────────────────
	// XTAL (40 MHz) → fractional MCLK divider → integer BCK divider.
	// Without an explicit MCLK request, BCLK = 40 MHz / divN / bckDiv with an
	// integer divN, preserving the existing behavior.
	var bckDiv uint8
	if cfg.MCLKFrequency != 0 && cfg.MCLK != NoPin {
		frameHz := cfg.SampleRate * 32 // 2 channels × 16 bits
		if cfg.MCLKFrequency%frameHz != 0 {
			return nil, ErrI2STxMCLKNotDivisible
		}
		bd := cfg.MCLKFrequency / frameHz
		if bd < 2 || bd > 63 {
			return nil, ErrI2STxMCLKNotDivisible
		}
		bckDiv = uint8(bd)

		actual, ok := i2sSetTXMCLK(hw, cfg.MCLKFrequency)
		if !ok {
			return nil, ErrI2STxBadClock
		}
		println("i2s init: controller=", hw.name, "MCLK target=", cfg.MCLKFrequency,
			"actual=", actual, "bckDiv=", bckDiv)
		tx.hasMCLK = true
	} else {
		divN, bd := i2sBestDividers(i2sXtalHz, cfg.SampleRate*32)
		bckDiv = bd
		println("i2s init: controller=", hw.name, "divN=", divN, "bckDiv=", bckDiv)
		hw.setTXClockSrcXTAL()
		hw.setTXDivN(uint32(divN))
		// Fractional divider fields X/Y/Z remain 0 → integer-only division.
		hw.setTXDivX(0)
		hw.setTXDivY(0)
		hw.setTXDivZ(0)
		hw.setTXDivYN1(0)
		hw.setTXClockEnable(true)
	}
	a, b, c := hw.clockRegs()
	println("i2s init: clk ctrl a=", a, "b=", b, "c=", c)
	println("i2s init: SOC_CLK_CTRL2=", esp.HP_SYS_CLKRST.SOC_CLK_CTRL2.Get(),
		"HP_RST_EN2=", esp.HP_SYS_CLKRST.HP_RST_EN2.Get(),
		"LP_AON_CLK=", esp.LP_AON_CLKRST.LP_AONCLKRST_HP_CLK_CTRL.Get())

	// ── TX_CONF1: sample-format fields (all 0-indexed: store value − 1) ───
	// TX_TDM_WS_WIDTH [8:0]        = 15  (WS high for 16 BCLK cycles)
	// TX_BITS_MOD     [18:14]      = 15  (16-bit data slot width)
	// TX_HALF_SAMPLE_BITS [26:19]  = 15  (16 bits per half-frame)
	// TX_TDM_CHAN_BITS    [31:27]  = 15  (16-bit channel-slot size)
	const bitsM1 = uint32(15) // 16 − 1
	tx.hw.TX_CONF1.Set(
		bitsM1 | // TX_TDM_WS_WIDTH [8:0]
			(bitsM1 << 14) | // TX_BITS_MOD [18:14]
			(bitsM1 << 19) | // TX_HALF_SAMPLE_BITS [26:19]
			(bitsM1 << 27), // TX_TDM_CHAN_BITS [31:27]
	)

	// ── TX_TDM_CTRL: active channel mask and total channel count ──────────
	// Stereo: chan0_en[0]=1, chan1_en[1]=1, tot_chan_num[19:16]=1 (= 2 channels)
	// Mono:   chan0_en[0]=1,               tot_chan_num[19:16]=0 (= 1 channel)
	// (TX_MONO in TX_CONF duplicates slot-0 data to both WS phases.)
	if cfg.Mono {
		tx.hw.TX_TDM_CTRL.Set(0x1 | (0 << 16))
	} else {
		tx.hw.TX_TDM_CTRL.Set(0x3 | (1 << 16))
	}

	// ── TX_CONF: format control, BCK divisor, master mode ─────────────────
	// TX_PCM_BYPASS [12]      = 1  bypass PCM A-law/μ-law codec
	// TX_MSB_SHIFT  [13]      = 1  Philips I2S: data lags WS by 1 BCLK
	// TX_LEFT_ALIGN [15]      = 1  MSB in leftmost bit of the slot
	// TX_TDM_EN     [19]      = 1  TDM mode (required on ESP32-P4)
	// TX_BCK_DIV_NUM [26:21]  = bckDiv − 1  (0-indexed BCLK divisor from MCLK)
	// TX_MONO       [6]       = 1  (mono only: duplicate slot-0 to both outputs)
	// TX_CHAN_MOD   [29:27]   = 2  (mono only: all output channels = left-ch data)
	//
	// TX_STOP_EN [4] is deliberately left 0.  With auto-stop enabled, a
	// repeated Write can start TX before the DMA has filled the FIFO; TX then
	// stops and leaves data stuck in the FIFO.  Write explicitly stops TX when
	// the transfer is complete.
	txConf := uint32(0) |
		(1 << 12) | // TX_PCM_BYPASS
		(1 << 13) | // TX_MSB_SHIFT (Philips I2S)
		(1 << 15) | // TX_LEFT_ALIGN
		(1 << 19) | // TX_TDM_EN
		(uint32(bckDiv-1) << 21) // TX_BCK_DIV_NUM (0-indexed)
	if cfg.Mono {
		txConf |= (1 << 6)  // TX_MONO
		txConf |= (2 << 27) // TX_CHAN_MOD = 2
	}
	tx.hw.TX_CONF.Set(txConf)
	println("i2s init: TX_CONF=", tx.hw.TX_CONF.Get(), "TX_CONF1=", tx.hw.TX_CONF1.Get())

	// Commit TX_CONF and TX_CONF1 shadow registers to active hardware registers.
	i2sTxUpdate(tx.hw)

	// ── GPIO signal routing ────────────────────────────────────────────────
	i2sConnectOutPinP4(cfg.SCK, hw.bckOutIdx)
	i2sConnectOutPinP4(cfg.WS, hw.wsOutIdx)
	i2sConnectOutPinP4(cfg.SDO, hw.sdOutIdx)
	if tx.hasMCLK {
		i2sConnectOutPinP4(cfg.MCLK, hw.mclkOutIdx)
	}

	// Diagnostic: read back the GPIO-matrix out-sel registers.  The low 8
	// bits must show the selected controller's signal IDs.
	println("i2s: outsel BCLKpin=", cfg.SCK.outFunc().Get(),
		"LRCLKpin=", cfg.WS.outFunc().Get(),
		"SDATApin=", cfg.SDO.outFunc().Get())
	if tx.hasMCLK {
		println("i2s: outsel MCLKpin=", cfg.MCLK.outFunc().Get())
	}

	// Remember the pin masks for the pad-activity probe in Write.
	tx.mBclk = uint32(1) << uint32(cfg.SCK)
	tx.mLrclk = uint32(1) << uint32(cfg.WS)
	tx.mSdata = uint32(1) << uint32(cfg.SDO)
	if tx.hasMCLK {
		tx.mMclk = uint32(1) << uint32(cfg.MCLK)
	}

	// ── AHB-DMA clock, peripheral select, channel 0 reset ─────────────────
	esp.AHB_DMA.SetMISC_CONF_CLK_EN(1)
	esp.AHB_DMA.SetOUT_PERI_SEL_CH0_PERI_OUT_SEL_CH(uint32(hw.dmaPeriph))
	// Reset channel 0 to clear any stale state.
	esp.AHB_DMA.SetOUT_CONF0_CH0_OUT_RST_CH0(1)
	esp.AHB_DMA.SetOUT_CONF0_CH0_OUT_RST_CH0(0)

	return tx, nil
}

// Write transmits audio samples over I2S using AHB-DMA and blocks until all
// samples have been clocked out by the I2S shift register.
//
// Stereo (Mono=false): samples must be interleaved [L0, R0, L1, R1, ...].
// Mono   (Mono=true):  pass single-channel samples; hardware copies each to both outputs.
//
// The samples slice must reside in internal SRAM and remain valid for the
// duration of the call.  Large buffers are split into 4092-byte DMA descriptors.
func (tx *I2STx) Write(samples []int16) error {
	if len(samples) == 0 {
		return nil
	}

	// Build AHB-DMA descriptor chain over the sample buffer.
	const samplesPerDesc = maxI2SDescBytes / 2 // 2046 int16 values per descriptor
	numDescs := (len(samples) + samplesPerDesc - 1) / samplesPerDesc
	if numDescs > len(tx.descs) {
		return ErrI2STxBufferTooLarge
	}
	descs := tx.descs[:numDescs]

	for i := range descs {
		start := i * samplesPerDesc
		end := start + samplesPerDesc
		if end > len(samples) {
			end = len(samples)
		}
		chunk := samples[start:end]
		nb := uint16(len(chunk) * 2) // byte count
		eof := i == numDescs-1
		descs[i] = i2sDMADesc{
			Config: i2sDMADescCfg(nb, eof),
			Buf:    uint32(uintptr(unsafe.Pointer(&chunk[0]))),
			Next:   0,
		}
		if !eof {
			descs[i].Next = uint32(uintptr(unsafe.Pointer(&descs[i+1])))
		}
	}

	// The AHB-DMA reads L2MEM directly and bypasses the CPU L1 cache.  Flush
	// the sample buffer and descriptor chain before handing them to the DMA;
	// otherwise the DMA can read stale zeros even though the CPU wrote the
	// tone/MP3/WAV data.  ESP-IDF does the equivalent with
	// esp_cache_msync(C2M) in i2s_channel_preload_data().
	CacheWriteBack(uintptr(unsafe.Pointer(&samples[0])), uintptr(len(samples))*2)
	CacheWriteBack(uintptr(unsafe.Pointer(&descs[0])), uintptr(len(descs))*unsafe.Sizeof(descs[0]))

	// Reset the AHB-DMA OUT channel and clear all pending interrupt flags so a
	// second Write starts from a clean state.  Without this, a short transfer
	// leaves OUT_DONE/OUT_TOTAL_EOF set and the next transfer never sees
	// OUT_EOF.
	esp.AHB_DMA.SetOUT_CONF0_CH0_OUT_RST_CH0(1)
	esp.AHB_DMA.SetOUT_CONF0_CH0_OUT_RST_CH0(0)
	esp.AHB_DMA.SetOUT_INT_CLR_CH0_OUT_DONE_CH_INT_CLR(1)
	esp.AHB_DMA.SetOUT_INT_CLR_CH0_OUT_EOF_CH_INT_CLR(1)
	esp.AHB_DMA.SetOUT_INT_CLR_CH0_OUT_DSCR_ERR_CH_INT_CLR(1)
	esp.AHB_DMA.SetOUT_INT_CLR_CH0_OUT_TOTAL_EOF_CH_INT_CLR(1)

	// Load descriptor chain base address and start DMA.
	descAddr := uint32(uintptr(unsafe.Pointer(&descs[0])))
	esp.AHB_DMA.SetOUT_LINK_ADDR_CH0(descAddr)
	esp.AHB_DMA.SetOUT_LINK_CH0_OUTLINK_START_CH(1)

	// Start I2S TX output once, then leave it running across Write calls.
	// ESP-IDF's i2s_channel_write keeps the TX channel enabled; stopping and
	// restarting it per buffer can leave the FIFO unread on the second write.
	// If TX was stopped by a timeout, start it again here.
	if !tx.running {
		i2sTxUpdate(tx.hw)
		tx.hw.SetTX_CONF_TX_START(1)
		tx.running = true
	}

	// Diagnostic (first Write only): count pad-level transitions on the I2S
	// output pins while TX is running.  This proves whether BCLK/LRCLK/SDATA
	// actually toggle on the pads (GPIO matrix routing + I2S1 clocking), which
	// register dumps cannot show.  Requires the pins to be < 32 (GPIO.IN).
	if !i2sPadProbeDone {
		i2sPadProbeDone = true
		nb, nl, ns, nm := i2sCountPadEdges(tx.mBclk, tx.mLrclk, tx.mSdata, tx.mMclk)
		println("i2s: pad edges BCLK=", nb, "LRCLK=", nl, "SDATA=", ns, "MCLK=", nm)
	}

	// Poll until DMA has consumed all descriptors (suc_eof descriptor).
	// Uses a //go:noinline helper so LLVM LTO cannot inline and apply LICM to
	// hoist the volatile register read out of the polling loop.
	frames := uint64(len(samples))
	if !tx.mono {
		frames /= 2
	}
	timeoutUS := frames*1_000_000/uint64(tx.sampleRate) + 2_000_000
	deadline := i2sTicks() + timeoutUS*40
	for i2sDMAEOFRaw() == 0 {
		if i2sTicks() > deadline {
			tx.hw.SetTX_CONF_TX_START(0)
			i2sTxUpdate(tx.hw)
			tx.running = false
			esp.AHB_DMA.SetOUT_LINK_CH0_OUTLINK_STOP_CH(1)
			return ErrI2STxTimeout
		}
	}
	esp.AHB_DMA.SetOUT_INT_CLR_CH0_OUT_EOF_CH_INT_CLR(1)

	// Keep TX running.  The next Write resets and reloads the DMA channel; the
	// I2S serializer continues consuming the FIFO, so there is no stop/start
	// race between buffers.
	return nil
}

// DebugState prints the TX controller and AHB-DMA state.  Call it after a
// timeout to see whether the I2S FIFO is draining and the DMA is advancing.
func (tx *I2STx) DebugState() {
	println("i2s debug: TX_CONF=", tx.hw.TX_CONF.Get(),
		"STATE=", tx.hw.STATE.Get(),
		"FIFO_CNT=", tx.hw.FIFO_CNT.Get(),
		"BCK_CNT=", tx.hw.BCK_CNT.Get())
	println("i2s debug: OUT_STATE=", esp.AHB_DMA.OUT_STATE_CH0.Get(),
		"OUT_INT_RAW=", esp.AHB_DMA.OUT_INT_RAW_CH0.Get(),
		"OUT_LINK=", esp.AHB_DMA.OUT_LINK_CH0.Get())
}

// Close stops TX playback and disables the selected I2S APB clock.
// Do not use the I2STx after calling Close.
func (tx *I2STx) Close() {
	tx.hw.SetTX_CONF_TX_START(0)
	i2sTxUpdate(tx.hw)
	tx.running = false
	hw := i2sHardware(tx.ctrl)
	hw.enableAPB(false)
	if tx.hasMCLK {
		hw.enableMCLKPad(false)
	}
}

// ─── Internal helpers ─────────────────────────────────────────────────────────

// i2sDMAEOFRaw returns the raw OUT_EOF interrupt flag for AHB-DMA channel 0.
// Marked noinline so that LLVM LTO cannot inline this call site into a polling
// loop and apply Loop-Invariant Code Motion (LICM) to hoist the volatile load.
//
//go:noinline
func i2sDMAEOFRaw() uint32 {
	return esp.AHB_DMA.GetOUT_INT_RAW_CH0_OUT_EOF_CH_INT_RAW()
}

// i2sPadProbeDone gates the one-shot pad-activity probe in Write.
var i2sPadProbeDone bool

// i2sCountPadEdges counts pad input-level transitions observed on the given
// GPIO.IN bit masks over a fixed number of samples.  Used as a diagnostic to
// verify I2S output signals actually reach the pins (input buffer of an
// output pad reads back the driven level).  Masks must be for pins < 32.
//
//go:noinline
func i2sCountPadEdges(mBclk, mLrclk, mSdata, mMclk uint32) (nBclk, nLrclk, nSdata, nMclk uint32) {
	const iters = 200_000
	last := esp.GPIO.IN.Get()
	for i := 0; i < iters; i++ {
		v := esp.GPIO.IN.Get()
		ch := v ^ last
		last = v
		if ch&mBclk != 0 {
			nBclk++
		}
		if ch&mLrclk != 0 {
			nLrclk++
		}
		if ch&mSdata != 0 {
			nSdata++
		}
		if ch&mMclk != 0 {
			nMclk++
		}
	}
	return
}

// i2sTXIdleRaw returns the TX_IDLE bit from the I2S STATE register.
// Marked noinline for the same LICM reason as i2sDMAEOFRaw.
//
//go:noinline
func i2sTXIdleRaw(hw *esp.I2S_Type) uint32 {
	return hw.GetSTATE_TX_IDLE()
}

// i2sTxUpdate commits TX_CONF / TX_CONF1 shadow registers to the active I2S
// hardware registers.  TX_UPDATE is a self-clearing strobe; we poll until it
// deasserts to confirm the update has taken effect.
func i2sTxUpdate(hw *esp.I2S_Type) {
	hw.SetTX_CONF_TX_UPDATE(1)
	for hw.GetTX_CONF_TX_UPDATE() != 0 {
	}
}

// i2sConnectOutPinP4 routes an I2S0 peripheral output signal to a GPIO pin
// via the GPIO matrix on ESP32-P4.
func i2sConnectOutPinP4(p Pin, sig uint32) {
	if p == NoPin {
		return
	}
	// Configure the pin as a GPIO output (sets IO MUX to GPIO function and
	// enables the output driver).
	p.Configure(PinConfig{Mode: PinOutput})
	// Override the default SIG_GPIO_OUT_IDX (256) with the I2S signal index,
	// routing the I2S peripheral output through the GPIO matrix to this pin.
	p.outFunc().Set(sig)
}

// i2sSetTXMCLK configures the selected I2S TX module clock from XTAL to
// targetHz, binds MCLK to the TX module, and enables the MCLK pad clock.  It
// returns the achieved MCLK frequency, or false if no divider fits.
func i2sSetTXMCLK(hw i2sHw, targetHz uint32) (uint32, bool) {
	divN, divX, divY, divZ, divYN1, actual := i2sFracDivider(i2sXtalHz, targetHz)
	if divN == 0 {
		return 0, false
	}
	println("i2s: MCLK controller=", hw.name, "divN=", divN, "x=", divX, "y=", divY, "z=", divZ, "yn1=", divYN1, "actual=", actual)

	// Match ESP-IDF's order: enable the module clock, bind MCLK to TX, set the
	// source, then program the fractional divider.
	hw.setTXClockEnable(true)
	hw.setMSTClkTX()
	hw.setTXClockSrcXTAL()

	// Apply the ESP-IDF double-division workaround: briefly select a small
	// integer divisor, clear the fractional fields, then write the target
	// coefficients and finally the integer divisor.
	hw.setTXDivN(2)
	hw.setTXDivYN1(0)
	hw.setTXDivY(1)
	hw.setTXDivZ(0)
	hw.setTXDivX(0)
	hw.setTXDivYN1(uint32(divYN1))
	hw.setTXDivZ(uint32(divZ))
	hw.setTXDivY(uint32(divY))
	hw.setTXDivX(uint32(divX))
	hw.setTXDivN(uint32(divN))

	hw.enableMCLKPad(true)
	return actual, true
}

// i2sFracDivider finds a fractional MCLK divider N + num/den such that
// srcHz / (N + num/den) is as close as possible to targetHz.  It returns the
// register fields used by HP_SYS_CLKRST: integer divisor N, fractional X/Y/Z
// and the YN1 sign/borrow bit.  A zero divN means no valid divider was found.
func i2sFracDivider(srcHz, targetHz uint32) (divN, divX, divY, divZ, divYN1 uint8, actual uint32) {
	if targetHz == 0 || srcHz < targetHz {
		return 0, 0, 0, 0, 0, 0
	}
	const maxFract = 512

	integer := srcHz / targetHz
	freqErr := srcHz % targetHz
	denom := uint32(2)
	numer := uint32(0)

	if freqErr != 0 {
		// Search for the closest fraction num/den to freqErr/targetHz.
		if freqErr < targetHz-targetHz/(maxFract-1)*2 {
			min := ^uint32(0)
			for a := uint32(2); a < maxFract && min != 0; a++ {
				b := (a*freqErr + targetHz/2) / targetHz
				var sub uint32
				if targetHz*b > freqErr*a {
					sub = targetHz*b - freqErr*a
				} else {
					sub = freqErr*a - targetHz*b
				}
				if sub < min {
					denom = a
					numer = b
					min = sub
				}
			}
		} else {
			integer++
		}
	}

	// Normalize an improper fraction.
	if numer >= denom {
		integer += numer / denom
		numer %= denom
	}
	if integer < 1 || integer > 255 {
		return 0, 0, 0, 0, 0, 0
	}

	divN = uint8(integer)
	if numer != 0 {
		divYN1 = 0
		if numer*2 > denom {
			divYN1 = 1
			divZ = uint8(denom - numer)
		} else {
			divZ = uint8(numer)
		}
		divX = uint8(denom/uint32(divZ) - 1)
		divY = uint8(denom % uint32(divZ))
	}

	if numer == 0 {
		actual = srcHz / integer
	} else {
		actual = uint32((uint64(srcHz) * uint64(denom)) / (uint64(integer)*uint64(denom) + uint64(numer)))
	}
	return
}

// i2sBestDividers finds integer clock dividers (divN ∈ [1..255], bckDiv ∈ [2..63])
// that minimise the error between the achieved and target BCLK frequency:
//
//	achieved_BCLK = srcHz / divN / bckDiv
//
// divN is written directly to PERI_CLK_CTRL13_REG_I2S0_TX_DIV_N.
// bckDiv is stored as (bckDiv − 1) in TX_CONF TX_BCK_DIV_NUM (0-indexed hardware field).
func i2sBestDividers(srcHz, targetHz uint32) (divN, bckDiv uint8) {
	var bestErr uint32 = 0xFFFFFFFF
	bestN, bestM := 1, 2
	for n := 1; n <= 255; n++ {
		nDiv := srcHz / uint32(n)
		for m := 2; m <= 63; m++ {
			actual := nDiv / uint32(m)
			var diff uint32
			if actual > targetHz {
				diff = actual - targetHz
			} else {
				diff = targetHz - actual
			}
			if diff < bestErr {
				bestErr = diff
				bestN, bestM = n, m
			}
			if bestErr == 0 {
				return uint8(bestN), uint8(bestM)
			}
			if actual < targetHz {
				break // increasing m only reduces frequency further
			}
		}
	}
	return uint8(bestN), uint8(bestM)
}
