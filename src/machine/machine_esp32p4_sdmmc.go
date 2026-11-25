//go:build esp32p4

// ESP32-P4 SDMMC host driver, slot 0.
//
// This file implements the SD host controller bring-up, the native SD
// command protocol, and block-oriented PIO and IDMAC transfers.  It does
// not provide a filesystem or a byte-addressed ReadAt/WriteAt device.
//
// The controller uses shared static DMA and descriptor buffers, so only one
// controller may be initialized at a time.  Public block operations are
// serialized with a mutex.  Call Init before any block operation.
//
// Reference: ESP-IDF components/esp_hal_sd/esp32p4/include/hal/sdmmc_ll.h,
// components/esp_driver_sdmmc/src/sd_host_sdmmc.c, and
// components/sdmmc/sdmmc_sd.c.

package machine

import (
	"device/esp"
	"errors"
	"machine/internal/esp32p4sdmmc"
	"runtime/volatile"
	"sync"
	"unsafe"
)

const (
	sdmmcClockHz     = 160_000_000 // PLL_F160M
	sdmmcInitFreq    = 400_000
	sdmmcDataFreq    = 400_000
	sdmmcBlockSize   = 512
	sdmmcCmdTimeout  = 500_000 // microseconds
	sdmmcDataTimeout = 2_000_000
	sdmmcInitTimeout = 2_000_000

	// Multi-block reads use one descriptor per 4 KiB. The example compares
	// 16 blocks (8 KiB), which needs two descriptors.
	sdmmcMaxMultiBlocks = 16
	sdmmcMultiBytes     = sdmmcMaxMultiBlocks * sdmmcBlockSize
	sdmmcMaxDescBytes   = 4096
	sdmmcMaxMultiDescs  = (sdmmcMultiBytes + sdmmcMaxDescBytes - 1) / sdmmcMaxDescBytes
)

// SDHOST CTRL register bits that are not exposed as generated setters.
const (
	sdmmcCtrlDMAReset       = 1 << 2
	sdmmcCtrlDMAEnable      = 1 << 5
	sdmmcCtrlUseInternalDMA = 1 << 25
)

// SDHOST interrupt bits.
const (
	sdmmcIntRespErr  = 1 << 1
	sdmmcIntCmdDone  = 1 << 2
	sdmmcIntDataOver = 1 << 3
	sdmmcIntTXDR     = 1 << 4
	sdmmcIntRXDR     = 1 << 5
	sdmmcIntDCRC     = 1 << 7
	sdmmcIntRCRC     = 1 << 6
	sdmmcIntRTO      = 1 << 8
	sdmmcIntDRTO     = 1 << 9
	sdmmcIntHTO      = 1 << 10
	sdmmcIntFRUN     = 1 << 11
	sdmmcIntHLE      = 1 << 12
	sdmmcIntSBE      = 1 << 13
	sdmmcIntEBE      = 1 << 15

	sdmmcCmdErrMask  = sdmmcIntRTO | sdmmcIntRCRC | sdmmcIntRespErr
	sdmmcDataErrMask = sdmmcIntDCRC | sdmmcIntDRTO | sdmmcIntHTO | sdmmcIntFRUN | sdmmcIntHLE | sdmmcIntSBE | sdmmcIntEBE
)

// SDHOST IDMAC status bits.
const (
	sdmmcIDMACTI  = 1 << 0
	sdmmcIDMACRI  = 1 << 1
	sdmmcIDMACFBE = 1 << 2
	sdmmcIDMACDU  = 1 << 4
	sdmmcIDMACCES = 1 << 5
	sdmmcIDMACNI  = 1 << 8

	sdmmcIDMACError = sdmmcIDMACFBE | sdmmcIDMACDU | sdmmcIDMACCES
)

// SDHOST STATUS bits.
const (
	sdmmcStatusFIFOFull = 1 << 3
	sdmmcStatusDataBusy = 1 << 9
)

// SDHOST BMOD bits. On the P4, IDMAC enable is bit 7 (not bit 2 as on
// older DesignWare variants).
const (
	sdmmcBMODSWR = 1 << 0
	sdmmcBMODFB  = 1 << 1
	sdmmcBMODDE  = 1 << 7
)

var (
	// ErrSDMMCTimeout is returned when the card does not respond in time.
	ErrSDMMCTimeout = errors.New("machine: SDMMC timeout")
	// ErrSDMMCCommand is returned when a command reports an error.
	ErrSDMMCCommand = errors.New("machine: SDMMC command error")
	// ErrSDMMCData is returned when a data transfer reports an error.
	ErrSDMMCData = errors.New("machine: SDMMC data error")
	// ErrSDMMCBlockSize is returned for unaligned block transfers.
	ErrSDMMCBlockSize = errors.New("machine: SDMMC block size must be a multiple of 512")
	// ErrSDMMCEraseUnsupported is returned by EraseBlocks. FAT does not need
	// erase; littlefs would.
	ErrSDMMCEraseUnsupported = errors.New("machine: SDMMC erase not implemented")
	// ErrSDMMCPIONotImplemented is returned for PIO writes.
	ErrSDMMCPIONotImplemented = errors.New("machine: SDMMC PIO write not implemented")
	// Stage-specific errors make bring-up logs actionable.
	ErrSDMMCPower           = errors.New("machine: SDMMC power setup failed")
	ErrSDMMCControllerReset = errors.New("machine: SDMMC controller reset failed")
	ErrSDMMCFIFOReset       = errors.New("machine: SDMMC FIFO reset failed")
	ErrSDMMCDMAReset        = errors.New("machine: SDMMC DMA reset failed")
	ErrSDMMCInUse           = errors.New("machine: SDMMC controller already active")
	ErrSDMMCNotInitialized  = errors.New("machine: SDMMC controller not initialized")
	ErrSDMMCClock           = errors.New("machine: SDMMC clock setup failed")
	ErrSDMMCNoCard          = errors.New("machine: SDMMC no card response")
	ErrSDMMCCID             = errors.New("machine: SDMMC CID read failed (CRC/invalid response)")
	ErrSDMMCCSD             = errors.New("machine: SDMMC CSD read failed (CRC/invalid response)")
	ErrSDMMCInit            = errors.New("machine: SDMMC card initialization failed")
)

// SDMMCConfig describes the slot-0 pins and optional power control.
type SDMMCConfig struct {
	CLK Pin // clock pin
	CMD Pin // command pin
	D0  Pin // data line 0
	D1  Pin // data line 1; unused when OneBit is true
	D2  Pin // data line 2; unused when OneBit is true
	D3  Pin // data line 3; unused when OneBit is true

	Power Pin        // active-low card power enable; NoPin to skip
	LDO   LDOChannel // 0 to skip; LDO4 on the WaveShare ESP32-P4-NANO

	OneBit bool // force 1-bit data for bring-up
	PIO    bool // use FIFO/PIO transfers instead of IDMAC

	DataFreq uint32 // data clock in Hz; 0 uses the 400 kHz default
}

// SDCard is the block-level interface for an initialized SD card.
// It is intentionally limited to operations that can be shared safely by
// callers through a mutex.
type SDCard interface {
	Init() error
	ReadBlocks(dst []byte, startBlockIdx int64) (int, error)
	WriteBlocks(data []byte, startBlockIdx int64) (int, error)
	EraseBlocks(startBlock, numBlocks int64) error
	NumberOfBlocks() int64
	BlockSize() int
	Close() error
}

// SDCardMulti exposes the optional multi-block read path.  It is separate
// from SDCard because the generic SD card contract only needs block I/O.
type SDCardMulti interface {
	SDCard
	ReadBlocksMulti(dst []byte, startBlockIdx int64) (int, error)
}

// SDCardInfo exposes card identity and negotiated bus information.
type SDCardInfo interface {
	Size() int64
	WriteBlockSize() int64
	EraseBlockSize() int64
	BusWidth() int
	OCRRaw() uint32
	HighCapacity() bool
	SDv2() bool
	CSDVersion() uint8
	CIDRaw() [16]byte
	CSDRaw() [16]byte
}

// SDCardDebug exposes the bring-up diagnostics used by board examples.
// It is separate from SDCard so the normal storage API stays small.
type SDCardDebug interface {
	DataLineLevels() (d0, d1, d2, d3 bool)
	DebugState() (ctrl, clkdiv, clksrc, clkena uint32)
	TransferDebug() (rintsts, idsts, ctrl, bmod, status uint32)
	LastR1() uint32
	DMAState() (dbaddr, bytcnt, blksiz, pldmnd, idinten uint32)
	PIOState() (rintsts, status, bytcnt, blksiz uint32)
	DescriptorState() (w0, w1, w2, w3 uint32)
	ClockState() (peri1, peri2, ref0, ref2, aonRst uint32)
}

type sdmmc struct {
	mu          sync.Mutex
	active      bool
	initialized bool

	hw  *esp.SDHOST_Type
	cfg SDMMCConfig

	rca          uint16
	highCapacity bool
	sdV2         bool
	ocr          uint32
	busWidth     int
	csd          [16]byte
	cid          [16]byte
	numBlocks    int64

	desc *esp32p4sdmmc.Descriptor

	lastRint, lastIdsts, lastCtrl, lastBmod, lastStatus uint32
	lastR1                                              uint32
}

// sdmmcBouncePool backs the DMA staging buffer. The IDMAC and cache
// maintenance operate on 64-byte cache lines, so align the usable buffer.
var sdmmcBouncePool [sdmmcBlockSize + 63]byte

func sdmmcBounce() []byte {
	base := uintptr(unsafe.Pointer(&sdmmcBouncePool[0]))
	aligned := (base + 63) &^ 63
	// CPU accesses use the non-cacheable alias; the IDMAC uses the physical
	// address returned by sdmmcPhysAddr.
	return unsafe.Slice((*byte)(unsafe.Pointer(aligned+sdmmcNonCacheableOff)), sdmmcBlockSize)
}

// sdmmcNonCacheableOff is SOC_NON_CACHEABLE_OFFSET_SRAM on the ESP32-P4.
const sdmmcNonCacheableOff = 0x40000000

// sdmmcDMAAddr returns the non-cacheable SRAM alias for addresses in the
// 0x4ff00000..0x4fffffff range.
func sdmmcDMAAddr(addr uintptr) uint32 {
	if addr >= 0x4ff00000 && addr < 0x50000000 {
		return uint32(addr + sdmmcNonCacheableOff)
	}
	return uint32(addr)
}

// sdmmcPhysAddr converts a CPU non-cacheable alias back to the physical SRAM
// address consumed by the IDMAC.
func sdmmcPhysAddr(addr uintptr) uint32 {
	if addr >= 0x8ff00000 && addr < 0x90000000 {
		return uint32(addr - sdmmcNonCacheableOff)
	}
	return uint32(addr)
}

// sdmmcMultiPool backs the multi-block DMA staging buffer.
var sdmmcMultiPool [sdmmcMultiBytes + 63]byte

func sdmmcMultiBuf() []byte {
	base := uintptr(unsafe.Pointer(&sdmmcMultiPool[0]))
	aligned := (base + 63) &^ 63
	return unsafe.Slice((*byte)(unsafe.Pointer(aligned+sdmmcNonCacheableOff)), sdmmcMultiBytes)
}

// sdmmcDescPool backs the IDMAC descriptor ring. The hardware requires
// 64-byte alignment on the P4 (L1 cache line), so align it manually.
var sdmmcDescPool struct {
	buf [esp32p4sdmmc.DescriptorSize*sdmmcMaxMultiDescs + 63]byte
}

func sdmmcDescriptorAt(i int) *esp32p4sdmmc.Descriptor {
	base := uintptr(unsafe.Pointer(&sdmmcDescPool.buf[0]))
	aligned := (base + 63) &^ 63
	return (*esp32p4sdmmc.Descriptor)(unsafe.Pointer(aligned + uintptr(i)*esp32p4sdmmc.DescriptorSize + sdmmcNonCacheableOff))
}

func sdmmcDescriptor() *esp32p4sdmmc.Descriptor {
	return sdmmcDescriptorAt(0)
}

// NewSDMMC returns a slot-0 SD card controller for the given pins.
// It does not access the card; call Init to power and initialize it.
func NewSDMMC(config SDMMCConfig) (SDCard, error) {
	return &sdmmc{
		hw:   esp.SDHOST,
		cfg:  config,
		desc: sdmmcDescriptor(),
	}, nil
}

// sdmmcActiveMu serializes controller creation and shutdown.  The DMA
// bounce pools and descriptor ring are package-global, so only one slot-0
// controller may be active at a time.
var sdmmcActiveMu sync.Mutex
var sdmmcActive bool

var _ SDCard = (*sdmmc)(nil)
var _ SDCardMulti = (*sdmmc)(nil)
var _ SDCardInfo = (*sdmmc)(nil)
var _ SDCardDebug = (*sdmmc)(nil)

// Init powers the card, brings up the controller, and runs the SD
// initialization sequence.  Only one controller may be active at a time.
func (d *sdmmc) Init() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.initialized {
		return nil
	}

	sdmmcActiveMu.Lock()
	if sdmmcActive {
		sdmmcActiveMu.Unlock()
		return ErrSDMMCInUse
	}
	sdmmcActive = true
	d.active = true
	sdmmcActiveMu.Unlock()

	if err := d.initLocked(); err != nil {
		d.releaseActiveLocked()
		return err
	}
	d.initialized = true
	return nil
}

func (d *sdmmc) initLocked() error {
	if err := d.powerOn(); err != nil {
		return ErrSDMMCPower
	}
	d.enableBusClock()
	d.configurePins()
	if err := d.resetController(); err != nil {
		return err
	}
	if err := d.setClock(sdmmcInitFreq); err != nil {
		return ErrSDMMCClock
	}
	if err := d.initCard(); err != nil {
		return err
	}
	dataFreq := d.cfg.DataFreq
	if dataFreq == 0 {
		dataFreq = sdmmcDataFreq
	}
	if err := d.setClock(dataFreq); err != nil {
		return ErrSDMMCClock
	}
	return nil
}

func (d *sdmmc) powerOn() error {
	// Fully remove card power. A controller reset alone does not reset a card
	// that was left powered in a data/4-bit state; the physical reinsertion
	// test works because it removes both VDD and the pin bias.
	if d.cfg.LDO != 0 {
		_ = LDODisable(d.cfg.LDO)
	}
	if d.cfg.Power != NoPin {
		d.cfg.Power.Configure(PinConfig{Mode: PinOutput})
		d.cfg.Power.High() // active-low: off
	}
	d.highZPins()
	sdmmcDelayUS(500_000)

	if d.cfg.LDO != 0 {
		if err := LDOConfigure(d.cfg.LDO, 3300); err != nil {
			return err
		}
	}
	if d.cfg.Power != NoPin {
		d.cfg.Power.Low() // on
	}
	sdmmcDelayUS(100_000)
	return nil
}

// highZPins releases the SD lines while card power is off so the card cannot
// be parasitic-powered through CMD/DAT pullups.
func (d *sdmmc) highZPins() {
	for _, p := range []Pin{d.cfg.CLK, d.cfg.CMD, d.cfg.D0, d.cfg.D1, d.cfg.D2, d.cfg.D3} {
		if p != NoPin {
			p.Configure(PinConfig{Mode: PinInput})
		}
	}
}

func (d *sdmmc) enableBusClock() {
	esp.HP_SYS_CLKRST.SetSOC_CLK_CTRL1_REG_SDMMC_SYS_CLK_EN(1)

	// SDMMC's default source is PLL_F160M. This derived PLL has its own
	// reference gate; ESP-IDF's esp_clk_tree_enable_src() sets it.
	esp.HP_SYS_CLKRST.SetREF_CLK_CTRL2_REG_REF_160M_CLK_EN(1)

	// Reset the SDMMC register block.
	esp.LP_AON_CLKRST.SetLP_AONCLKRST_HP_SDMMC_EMAC_RST_CTRL_LP_AONCLKRST_RST_EN_SDMMC(1)
	esp.LP_AON_CLKRST.SetLP_AONCLKRST_HP_SDMMC_EMAC_RST_CTRL_LP_AONCLKRST_RST_EN_SDMMC(0)
	sdmmcDelayUS(10)

	// Match ESP-IDF's sd_host_set_clk_div order: program the host divider
	// first, then select/gate the clock source, then enable the internal
	// DRV/SAM/SLF clock phases. The controller/FIFO reset needs those clocks.
	d.setHostClockDiv(2)
	esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL01_REG_SDIO_LS_CLK_SRC_SEL(0)
	esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL01_REG_SDIO_LS_CLK_EN(1)
	d.initPhaseDelay()
}

// initPhaseDelay enables the SDIO driver/sample/self clock phases, matching
// sdmmc_ll_init_phase_delay() in ESP-IDF.
func (d *sdmmc) initPhaseDelay() {
	esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL02_REG_SDIO_LS_DRV_CLK_EN(1)
	esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL02_REG_SDIO_LS_SAM_CLK_EN(1)
	esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL02_REG_SDIO_LS_SLF_CLK_EN(1)
	esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL02_REG_SDIO_LS_DRV_CLK_EDGE_SEL(1)
	esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL02_REG_SDIO_LS_SAM_CLK_EDGE_SEL(0)
	esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL02_REG_SDIO_LS_SLF_CLK_EDGE_SEL(0)
	esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL02_REG_SDIO_LS_CLK_EDGE_CFG_UPDATE(1)
	esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL02_REG_SDIO_LS_CLK_EDGE_CFG_UPDATE(0)
}

// setHostClockDiv programs the first-stage SDMMC clock divider.
func (d *sdmmc) setHostClockDiv(div uint32) {
	h, n, l, update := esp32p4sdmmc.ClockEdge(div)
	esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL02_REG_SDIO_LS_CLK_EDGE_H(h)
	esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL02_REG_SDIO_LS_CLK_EDGE_N(n)
	esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL02_REG_SDIO_LS_CLK_EDGE_L(l)
	if update {
		esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL02_REG_SDIO_LS_CLK_EDGE_CFG_UPDATE(1)
		esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL02_REG_SDIO_LS_CLK_EDGE_CFG_UPDATE(0)
	}
}

func (d *sdmmc) configurePins() {
	// Slot 0 uses the dedicated IOMUX function 0 on GPIO39-44.
	d.configurePin(d.cfg.CLK, false)
	d.configurePin(d.cfg.CMD, true)
	d.configurePin(d.cfg.D0, true)
	d.configurePin(d.cfg.D1, true)
	d.configurePin(d.cfg.D2, true)
	d.configurePin(d.cfg.D3, true)
}

func (d *sdmmc) configurePin(p Pin, pullup bool) {
	if p == NoPin {
		return
	}
	var mux uint32
	mux |= 0 << esp.IO_MUX_GPIO_MCU_SEL_Pos // function 0 = SDMMC
	mux |= esp.IO_MUX_GPIO_FUN_IE
	mux |= 3 << esp.IO_MUX_GPIO_FUN_DRV_Pos
	if pullup {
		mux |= esp.IO_MUX_GPIO_FUN_WPU
	}
	p.mux().Set(mux)
}

func (d *sdmmc) resetController() error {
	// ESP-IDF asserts controller, FIFO, and DMA reset together and waits for
	// all three to self-clear. Write the full word to avoid a read-modify-write
	// on a self-clearing register.
	const resetMask = 1 | 1<<1 | sdmmcCtrlDMAReset
	d.hw.CTRL.Set(resetMask)
	if err := d.waitBitsClear(&d.hw.CTRL, resetMask, sdmmcCmdTimeout); err != nil {
		// On the P4 the FIFO reset can remain asserted until the card clock
		// domain is running. Enable the card clock and retry just the FIFO bit.
		if d.hw.CTRL.Get()&(1<<1) != 0 {
			d.hw.CLKDIV.Set(0)
			d.hw.CLKSRC.Set(0)
			d.hw.CLKENA.SetBits(1)
			_ = d.updateClock()
			d.hw.CTRL.SetBits(1 << 1)
			err = d.waitBitsClear(&d.hw.CTRL, 1<<1, sdmmcCmdTimeout)
		}
		if err != nil {
			ctrl := d.hw.CTRL.Get()
			switch {
			case ctrl&1 != 0:
				return ErrSDMMCControllerReset
			case ctrl&(1<<1) != 0:
				return ErrSDMMCFIFOReset
			case ctrl&sdmmcCtrlDMAReset != 0:
				return ErrSDMMCDMAReset
			}
			return err
		}
	}

	d.hw.TMOUT.Set(0xffffff<<8 | 0xff)
	d.hw.CTYPE.Set(0)
	d.hw.BLKSIZ.Set(sdmmcBlockSize)
	d.hw.BYTCNT.Set(0)
	d.hw.INTMASK.Set(0)
	d.hw.RINTSTS.Set(0xffffffff)
	d.hw.IDSTS.Set(0xffffffff)
	d.hw.IDINTEN.Set(sdmmcIDMACRI | sdmmcIDMACTI | sdmmcIDMACNI)
	d.hw.BMOD.Set(0)
	d.hw.BMOD.SetBits(sdmmcBMODSWR) // IDMAC software reset, as in sdmmc_ll_init_dma
	// ESP-IDF enables the DMA interface during host init, before any transfer.
	d.hw.CTRL.SetBits(sdmmcCtrlDMAEnable)
	return nil
}

func (d *sdmmc) setClock(freq uint32) error {
	// Follow the ESP-IDF sequence: stop the card clock, update the CIU,
	// program both dividers, update again, then re-enable the clock.
	d.hw.CLKENA.ClearBits(1)
	if err := d.updateClock(); err != nil {
		return err
	}

	hostDiv, cardDiv := esp32p4sdmmc.ClockDiv(sdmmcClockHz, freq)
	d.setHostClockDiv(hostDiv)
	d.initPhaseDelay()

	d.hw.CLKDIV.Set(cardDiv) // clk_divider0
	d.hw.CLKSRC.Set(0)       // card0 -> divider 0
	if err := d.updateClock(); err != nil {
		return err
	}

	d.hw.CLKENA.SetBits(1) // enable card0 clock
	if err := d.updateClock(); err != nil {
		return err
	}
	sdmmcDelayUS(1_000)
	return nil
}

func (d *sdmmc) updateClock() error {
	word := (esp32p4sdmmc.CommandOptions{Index: 0, UpdateClock: true}).Word()
	d.hw.CMD.Set(word)
	deadline := sdmmcDeadlineUS(sdmmcCmdTimeout)
	for d.hw.CMD.Get()&(1<<31) != 0 {
		if sdmmcTicks() > deadline {
			return ErrSDMMCTimeout
		}
		sdmmcDelayUS(10)
	}
	return nil
}

func (d *sdmmc) initCard() error {
	// Reset twice: the first CMD0 also sends the required 80 initialization
	// clocks; the second gives a card that was left in a data/4-bit state a
	// second chance to return to idle.
	for i := 0; i < 2; i++ {
		if _, err := d.command(esp32p4sdmmc.CommandOptions{Index: esp32p4sdmmc.CmdGoIdleState}, 0); err != nil {
			return ErrSDMMCClock
		}
		sdmmcDelayUS(1_000)
	}

	// CMD8 identifies SD v2 cards. Retry it: a marginal response must not make
	// a 32 GB SDHC card fall back to the 121 MiB/CSD v1 legacy mode.
	v2 := false
	for i := 0; i < 3; i++ {
		if resp, err := d.command(esp32p4sdmmc.CommandOptions{
			Index:    esp32p4sdmmc.CmdSendIfCond,
			Response: esp32p4sdmmc.RespR7,
			CheckCRC: true,
		}, 0x1aa); err == nil {
			if resp[0]&0xfff != 0x1aa {
				return ErrSDMMCInit
			}
			v2 = true
			break
		}
		sdmmcDelayUS(1_000)
	}

	arg := uint32(esp32p4sdmmc.OCRVoltage)
	if v2 {
		arg |= esp32p4sdmmc.OCRCCS
	}
	deadline := sdmmcDeadlineUS(sdmmcInitTimeout)
	var ocr uint32
	for {
		if _, err := d.command(esp32p4sdmmc.CommandOptions{
			Index:    esp32p4sdmmc.CmdAppCmd,
			Response: esp32p4sdmmc.RespR1,
			CheckCRC: true,
		}, uint32(d.rca)<<16); err != nil {
			return ErrSDMMCNoCard
		}
		resp, err := d.command(esp32p4sdmmc.CommandOptions{
			Index:    esp32p4sdmmc.CmdSendOpCond,
			Response: esp32p4sdmmc.RespR3,
		}, arg)
		if err != nil {
			return ErrSDMMCNoCard
		}
		ocr = resp[0]
		if ocr&esp32p4sdmmc.OCRReady != 0 {
			// A v2 card that has not yet reported CCS is still in the legacy
			// 121 MiB fallback state. Keep polling until it reports high
			// capacity or the initialization deadline expires.
			if !v2 || ocr&esp32p4sdmmc.OCRCCS != 0 {
				break
			}
		}
		if sdmmcTicks() > deadline {
			return ErrSDMMCNoCard
		}
		sdmmcDelayUS(1_000)
	}
	d.highCapacity = ocr&esp32p4sdmmc.OCRCCS != 0
	d.ocr = ocr
	d.sdV2 = v2

	// CID (CMD2), RCA (CMD3), CSD (CMD9), then select the card (CMD7).
	resp, err := d.command(esp32p4sdmmc.CommandOptions{
		Index:    esp32p4sdmmc.CmdAllSendCID,
		Response: esp32p4sdmmc.RespR2,
		CheckCRC: true,
	}, 0)
	if err != nil {
		return ErrSDMMCCID
	}
	d.cid = esp32p4sdmmc.ResponseBytes(resp)

	resp, err = d.command(esp32p4sdmmc.CommandOptions{
		Index:    esp32p4sdmmc.CmdSendRelativeAddr,
		Response: esp32p4sdmmc.RespR6,
		CheckCRC: true,
	}, 0)
	if err != nil {
		return ErrSDMMCInit
	}
	d.rca = uint16(resp[0] >> 16)
	if d.rca == 0 {
		return ErrSDMMCInit
	}

	resp, err = d.command(esp32p4sdmmc.CommandOptions{
		Index:    esp32p4sdmmc.CmdSendCSD,
		Response: esp32p4sdmmc.RespR2,
		CheckCRC: true,
	}, uint32(d.rca)<<16)
	if err != nil {
		return ErrSDMMCCSD
	}
	d.csd = esp32p4sdmmc.ResponseBytes(resp)

	if _, err := d.command(esp32p4sdmmc.CommandOptions{
		Index:    esp32p4sdmmc.CmdSelectCard,
		Response: esp32p4sdmmc.RespR1b,
		CheckCRC: true,
	}, uint32(d.rca)<<16); err != nil {
		return ErrSDMMCInit
	}
	if err := d.waitReady(); err != nil {
		return ErrSDMMCInit
	}

	if !d.highCapacity {
		if _, err := d.command(esp32p4sdmmc.CommandOptions{
			Index:    esp32p4sdmmc.CmdSetBlockLen,
			Response: esp32p4sdmmc.RespR1,
			CheckCRC: true,
		}, sdmmcBlockSize); err != nil {
			return ErrSDMMCInit
		}
	}

	// Try 4-bit bus mode; fall back to 1-bit if the card rejects it.
	d.busWidth = 1
	if !d.cfg.OneBit {
		if _, err := d.command(esp32p4sdmmc.CommandOptions{
			Index:    esp32p4sdmmc.CmdAppCmd,
			Response: esp32p4sdmmc.RespR1,
			CheckCRC: true,
		}, uint32(d.rca)<<16); err == nil {
			if _, err := d.command(esp32p4sdmmc.CommandOptions{
				Index:    esp32p4sdmmc.CmdSetBusWidth,
				Response: esp32p4sdmmc.RespR1,
				CheckCRC: true,
			}, 2); err == nil {
				d.hw.CTYPE.SetBits(1) // slot0: 4-bit
				d.busWidth = 4
			}
		}
	}

	d.numBlocks = esp32p4sdmmc.CSDBlockCount(d.csd)
	if d.numBlocks <= 0 {
		return ErrSDMMCCommand
	}
	return nil
}

func (d *sdmmc) waitReady() error {
	deadline := sdmmcDeadlineUS(sdmmcCmdTimeout)
	for {
		resp, err := d.command(esp32p4sdmmc.CommandOptions{
			Index:    esp32p4sdmmc.CmdSendStatus,
			Response: esp32p4sdmmc.RespR1,
			CheckCRC: true,
		}, uint32(d.rca)<<16)
		if err != nil {
			return err
		}
		if resp[0]&esp32p4sdmmc.R1ReadyForData != 0 {
			return nil
		}
		if sdmmcTicks() > deadline {
			return ErrSDMMCTimeout
		}
		sdmmcDelayUS(100)
	}
}

func (d *sdmmc) command(opts esp32p4sdmmc.CommandOptions, arg uint32) ([4]uint32, error) {
	var resp [4]uint32
	if err := d.waitBitsClear(&d.hw.CMD, 1<<31, sdmmcCmdTimeout); err != nil {
		return resp, err
	}
	d.hw.RINTSTS.Set(0xffffffff)
	d.hw.CMDARG.Set(arg)
	d.hw.CMD.Set(opts.Word())

	deadline := sdmmcDeadlineUS(sdmmcCmdTimeout)
	for {
		st := d.hw.RINTSTS.Get()
		if st&sdmmcCmdErrMask != 0 {
			d.hw.RINTSTS.Set(st & (sdmmcCmdErrMask | sdmmcIntCmdDone))
			return resp, ErrSDMMCCommand
		}
		if st&sdmmcIntCmdDone != 0 {
			resp[0] = d.hw.RESP0.Get()
			resp[1] = d.hw.RESP1.Get()
			resp[2] = d.hw.RESP2.Get()
			resp[3] = d.hw.RESP3.Get()
			d.hw.RINTSTS.Set(st & (sdmmcCmdErrMask | sdmmcIntCmdDone))
			if opts.Response == esp32p4sdmmc.RespR1 || opts.Response == esp32p4sdmmc.RespR1b {
				d.lastR1 = resp[0]
				if esp32p4sdmmc.R1HasError(resp[0]) {
					return resp, ErrSDMMCCommand
				}
			}
			break
		}
		if sdmmcTicks() > deadline {
			return resp, ErrSDMMCTimeout
		}
		sdmmcDelayUS(10)
	}

	if opts.Response == esp32p4sdmmc.RespR1b {
		deadline = sdmmcDeadlineUS(sdmmcCmdTimeout)
		for d.hw.STATUS.Get()&sdmmcStatusDataBusy != 0 { // data_busy
			if sdmmcTicks() > deadline {
				return resp, ErrSDMMCTimeout
			}
			sdmmcDelayUS(10)
		}
	}
	return resp, nil
}

func (d *sdmmc) waitBitsClear(reg *volatile.Register32, mask uint32, timeoutUS uint32) error {
	deadline := sdmmcDeadlineUS(timeoutUS)
	for reg.Get()&mask != 0 {
		if sdmmcTicks() > deadline {
			return ErrSDMMCTimeout
		}
		sdmmcDelayUS(10)
	}
	return nil
}

// sdmmcTicks reads the free-running TIMG0 64-bit counter (40 MHz).
func sdmmcTicks() uint64 {
	esp.TIMG0.T0UPDATE.Set(1)
	for esp.TIMG0.T0UPDATE.Get() != 0 {
	}
	return uint64(esp.TIMG0.T0LO.Get()) | uint64(esp.TIMG0.T0HI.Get())<<32
}

// sdmmcDelayUS busy-waits for the given number of microseconds.
func sdmmcDelayUS(us uint32) {
	target := sdmmcTicks() + uint64(us)*40
	for sdmmcTicks() < target {
	}
}

func sdmmcDeadlineUS(us uint32) uint64 {
	return sdmmcTicks() + uint64(us)*40
}

// ReadBlocks reads len(dst) bytes starting at the 512-byte block index
// startBlockIdx.  len(dst) must be a multiple of 512.  It uses the selected
// PIO or IDMAC path and returns the number of bytes transferred.  DMA uses
// an internal bounce buffer, so the caller does not need cache alignment.
func (d *sdmmc) ReadBlocks(dst []byte, startBlockIdx int64) (int, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if !d.initialized {
		return 0, ErrSDMMCNotInitialized
	}
	return d.readBlocksLocked(dst, startBlockIdx)
}

func (d *sdmmc) readBlocksLocked(dst []byte, startBlockIdx int64) (int, error) {
	if len(dst)%sdmmcBlockSize != 0 {
		return 0, ErrSDMMCBlockSize
	}
	for i := 0; i < len(dst); i += sdmmcBlockSize {
		var err error
		if d.cfg.PIO {
			err = d.readBlockPIO(dst[i:i+sdmmcBlockSize], startBlockIdx+int64(i/sdmmcBlockSize))
		} else {
			err = d.transfer(dst[i:i+sdmmcBlockSize], startBlockIdx+int64(i/sdmmcBlockSize), false)
		}
		if err != nil {
			return i, err
		}
	}
	return len(dst), nil
}

// ReadBlocksMulti reads len(dst) bytes starting at the 512-byte block index
// startBlockIdx using CMD18 with automatic CMD12 stop.  len(dst) must be a
// multiple of 512.  Single-block remainders and PIO mode use the CMD17 path.
// Callers that need a per-block fallback can retry with ReadBlocks.
func (d *sdmmc) ReadBlocksMulti(dst []byte, startBlockIdx int64) (int, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if !d.initialized {
		return 0, ErrSDMMCNotInitialized
	}
	return d.readBlocksMultiLocked(dst, startBlockIdx)
}

func (d *sdmmc) readBlocksMultiLocked(dst []byte, startBlockIdx int64) (int, error) {
	if len(dst)%sdmmcBlockSize != 0 {
		return 0, ErrSDMMCBlockSize
	}
	if d.cfg.PIO {
		return d.readBlocksLocked(dst, startBlockIdx)
	}
	total := 0
	for total < len(dst) {
		n := len(dst) - total
		if n > sdmmcMultiBytes {
			n = sdmmcMultiBytes
		}
		block := startBlockIdx + int64(total/sdmmcBlockSize)
		var err error
		if n == sdmmcBlockSize {
			err = d.transfer(dst[total:total+n], block, false)
		} else {
			err = d.transferMulti(dst[total:total+n], block)
		}
		if err != nil {
			return total, err
		}
		total += n
	}
	return total, nil
}

// WriteBlocks writes len(data) bytes starting at the 512-byte block index
// startBlockIdx.  len(data) must be a multiple of 512.  It uses the selected
// PIO or IDMAC path and returns the number of bytes transferred.
func (d *sdmmc) WriteBlocks(data []byte, startBlockIdx int64) (int, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if !d.initialized {
		return 0, ErrSDMMCNotInitialized
	}
	if len(data)%sdmmcBlockSize != 0 {
		return 0, ErrSDMMCBlockSize
	}
	for i := 0; i < len(data); i += sdmmcBlockSize {
		var err error
		if d.cfg.PIO {
			err = d.writeBlockPIO(data[i:i+sdmmcBlockSize], startBlockIdx+int64(i/sdmmcBlockSize))
		} else {
			err = d.transfer(data[i:i+sdmmcBlockSize], startBlockIdx+int64(i/sdmmcBlockSize), true)
		}
		if err != nil {
			return i, err
		}
	}
	return len(data), nil
}

func (d *sdmmc) writeBlockPIO(src []byte, block int64) error {
	// Disable IDMAC and use the CPU-visible FIFO.
	d.hw.CTRL.ClearBits(sdmmcCtrlDMAEnable | sdmmcCtrlUseInternalDMA)
	d.hw.BMOD.ClearBits(sdmmcBMODFB | sdmmcBMODDE)
	d.hw.IDINTEN.Set(0)
	d.hw.FIFOTH.Set(0)
	d.hw.CTRL.SetBits(1 << 1) // FIFO reset
	_ = d.waitBitsClear(&d.hw.CTRL, 1<<1, sdmmcCmdTimeout)

	d.hw.BLKSIZ.Set(sdmmcBlockSize)
	d.hw.BYTCNT.Set(sdmmcBlockSize)
	d.hw.RINTSTS.Set(0xffffffff)

	arg := uint32(block)
	if !d.highCapacity {
		arg = uint32(block * sdmmcBlockSize)
	}
	opts := esp32p4sdmmc.CommandOptions{
		Index:    esp32p4sdmmc.CmdWriteSingleBlock,
		Response: esp32p4sdmmc.RespR1,
		CheckCRC: true,
		Data:     true,
		Write:    true,
	}
	if _, err := d.command(opts, arg); err != nil {
		d.captureTransferState()
		return err
	}

	const wordsTotal = sdmmcBlockSize / 4
	words := 0
	deadline := sdmmcDeadlineUS(sdmmcDataTimeout)
	for words < wordsTotal {
		st := d.hw.RINTSTS.Get()
		if st&sdmmcDataErrMask != 0 {
			d.captureTransferState()
			d.hw.RINTSTS.Set(0xffffffff)
			return ErrSDMMCData
		}
		if d.hw.STATUS.Get()&sdmmcStatusFIFOFull == 0 {
			off := words * 4
			w := uint32(src[off]) | uint32(src[off+1])<<8 | uint32(src[off+2])<<16 | uint32(src[off+3])<<24
			d.hw.BUFFIFO.Set(w)
			words++
		}
		if sdmmcTicks() > deadline {
			d.captureTransferState()
			return ErrSDMMCTimeout
		}
		sdmmcDelayUS(10)
	}

	deadline = sdmmcDeadlineUS(sdmmcDataTimeout)
	for d.hw.RINTSTS.Get()&sdmmcIntDataOver == 0 {
		st := d.hw.RINTSTS.Get()
		if st&sdmmcDataErrMask != 0 {
			d.captureTransferState()
			d.hw.RINTSTS.Set(0xffffffff)
			return ErrSDMMCData
		}
		if sdmmcTicks() > deadline {
			d.captureTransferState()
			return ErrSDMMCTimeout
		}
		sdmmcDelayUS(10)
	}
	// Wait for the card to release DAT0 after the write.
	deadline = sdmmcDeadlineUS(sdmmcDataTimeout)
	for d.hw.STATUS.Get()&sdmmcStatusDataBusy != 0 {
		if sdmmcTicks() > deadline {
			d.captureTransferState()
			return ErrSDMMCTimeout
		}
		sdmmcDelayUS(10)
	}
	d.hw.RINTSTS.Set(0xffffffff)

	// Ask the card for its status after the data phase. This is where a data
	// CRC/timeout/write error is reported; the CMD24 R1 is pre-data.
	if _, err := d.command(esp32p4sdmmc.CommandOptions{
		Index:    esp32p4sdmmc.CmdSendStatus,
		Response: esp32p4sdmmc.RespR1,
		CheckCRC: true,
	}, uint32(d.rca)<<16); err != nil {
		return err
	}
	return nil
}

func (d *sdmmc) readBlockPIO(dst []byte, block int64) error {
	// Disable IDMAC and take ownership of the FIFO. Reset the IDMAC first:
	// a stale IDMAC/FIFO state can make BUFFIFO return the same word without
	// advancing the read pointer.
	d.hw.BMOD.SetBits(sdmmcBMODSWR)
	resetDeadline := sdmmcDeadlineUS(sdmmcCmdTimeout)
	for d.hw.BMOD.Get()&sdmmcBMODSWR != 0 {
		if sdmmcTicks() > resetDeadline {
			return ErrSDMMCTimeout
		}
		sdmmcDelayUS(10)
	}
	d.hw.CTRL.ClearBits(sdmmcCtrlDMAEnable | sdmmcCtrlUseInternalDMA)
	d.hw.BMOD.ClearBits(sdmmcBMODFB | sdmmcBMODDE)
	d.hw.IDINTEN.Set(0)
	d.hw.FIFOTH.Set(0)
	d.hw.CTRL.SetBits(1 << 1) // FIFO reset
	_ = d.waitBitsClear(&d.hw.CTRL, 1<<1, sdmmcCmdTimeout)

	d.hw.BLKSIZ.Set(sdmmcBlockSize)
	d.hw.BYTCNT.Set(sdmmcBlockSize)
	d.hw.RINTSTS.Set(0xffffffff)

	arg := uint32(block)
	if !d.highCapacity {
		arg = uint32(block * sdmmcBlockSize)
	}
	opts := esp32p4sdmmc.CommandOptions{
		Index:    esp32p4sdmmc.CmdReadSingleBlock,
		Response: esp32p4sdmmc.RespR1,
		CheckCRC: true,
		Data:     true,
	}
	if _, err := d.command(opts, arg); err != nil {
		d.captureTransferState()
		return err
	}

	const wordsTotal = sdmmcBlockSize / 4
	words := 0
	deadline := sdmmcDeadlineUS(sdmmcDataTimeout)
	for words < wordsTotal {
		st := d.hw.RINTSTS.Get()
		if st&sdmmcDataErrMask != 0 {
			d.captureTransferState()
			d.hw.RINTSTS.Set(0xffffffff)
			return ErrSDMMCData
		}
		status := d.hw.STATUS.Get()
		count := int((status >> 17) & 0x1fff)
		for count > 0 && words < wordsTotal {
			w := d.hw.BUFFIFO.Get()
			off := words * 4
			dst[off+0] = byte(w)
			dst[off+1] = byte(w >> 8)
			dst[off+2] = byte(w >> 16)
			dst[off+3] = byte(w >> 24)
			words++
			count--
			// Give the FIFO read pointer time to advance. The P4 returns the
			// same word if BUFFIFO is read back-to-back too quickly.
			sdmmcDelayUS(1)
		}
		if words == wordsTotal {
			break
		}
		if sdmmcTicks() > deadline {
			d.captureTransferState()
			return ErrSDMMCTimeout
		}
		sdmmcDelayUS(10)
	}

	deadline = sdmmcDeadlineUS(sdmmcDataTimeout)
	for d.hw.RINTSTS.Get()&sdmmcIntDataOver == 0 {
		st := d.hw.RINTSTS.Get()
		if st&sdmmcDataErrMask != 0 {
			d.captureTransferState()
			d.hw.RINTSTS.Set(0xffffffff)
			return ErrSDMMCData
		}
		if sdmmcTicks() > deadline {
			d.captureTransferState()
			return ErrSDMMCTimeout
		}
		sdmmcDelayUS(10)
	}
	d.hw.RINTSTS.Set(0xffffffff)
	return nil
}

func (d *sdmmc) transfer(buf []byte, block int64, write bool) error {
	if len(buf) != sdmmcBlockSize {
		return ErrSDMMCBlockSize
	}

	// Always stage through the DMA-safe global buffer. This keeps the IDMAC
	// out of heap/PSRAM and guarantees 4-byte alignment.
	if write {
		copy(sdmmcBounce(), buf)
	}
	bounce := sdmmcBounce()
	ptr := uintptr(unsafe.Pointer(&bounce[0]))
	physPtr := sdmmcPhysAddr(ptr)

	*d.desc = esp32p4sdmmc.Descriptor{}
	d.desc.SetFirst(true)
	d.desc.SetLast(true)
	d.desc.SetOwned(true)
	d.desc.SetBuffer(physPtr, sdmmcBlockSize)
	d.desc.SetNext(0)
	descAddr := uintptr(unsafe.Pointer(d.desc))
	descPhys := sdmmcPhysAddr(descAddr)

	d.hw.BLKSIZ.Set(sdmmcBlockSize)
	d.hw.BYTCNT.Set(sdmmcBlockSize)
	// The IDMAC consumes physical SRAM addresses. Do not feed it the
	// 0x8FF... non-cacheable CPU alias.
	d.hw.DBADDR.Set(descPhys)
	d.hw.RINTSTS.Set(0xffffffff)
	d.hw.IDSTS.Set(0xffffffff)
	d.dmaEnable()
	d.hw.PLDMND.Set(1)

	arg := uint32(block)
	if !d.highCapacity {
		arg = uint32(block * sdmmcBlockSize)
	}
	opts := esp32p4sdmmc.CommandOptions{
		Index:    esp32p4sdmmc.CmdReadSingleBlock,
		Response: esp32p4sdmmc.RespR1,
		CheckCRC: true,
		Data:     true,
	}
	if write {
		opts.Index = esp32p4sdmmc.CmdWriteSingleBlock
		opts.Write = true
	}
	if _, err := d.command(opts, arg); err != nil {
		d.captureTransferState()
		d.dmaStop()
		return err
	}
	// Some P4 revisions do not latch the pre-command demand. Poll again after
	// the card has accepted the data command.
	d.hw.PLDMND.Set(1)
	sdmmcDelayUS(10)

	deadline := sdmmcDeadlineUS(sdmmcDataTimeout)
	for {
		// Keep the IDMAC demand asserted. On the P4 the FSM can suspend
		// before the card data reaches the FIFO; a later demand resumes it.
		d.hw.PLDMND.Set(1)
		rint := d.hw.RINTSTS.Get()
		idsts := d.hw.IDSTS.Get()
		if rint&sdmmcDataErrMask != 0 || idsts&sdmmcIDMACError != 0 {
			d.captureTransferState()
			d.hw.RINTSTS.Set(0xffffffff)
			d.hw.IDSTS.Set(0xffffffff)
			d.dmaStop()
			return ErrSDMMCData
		}
		if rint&sdmmcIntDataOver != 0 && idsts&(sdmmcIDMACRI|sdmmcIDMACTI) != 0 {
			break
		}
		if sdmmcTicks() > deadline {
			d.captureTransferState()
			d.dmaStop()
			return ErrSDMMCTimeout
		}
		sdmmcDelayUS(10)
	}

	d.hw.RINTSTS.Set(0xffffffff)
	d.hw.IDSTS.Set(0xffffffff)
	d.dmaStop()
	d.desc.SetOwned(false)

	if !write {
		copy(buf, bounce)
	}
	return nil
}

// transferMulti reads len(buf) bytes with CMD18 and automatic CMD12 stop.
// The data is staged through a non-cacheable DMA buffer; the caller's buffer
// is copied after the transfer.
func (d *sdmmc) transferMulti(buf []byte, block int64) error {
	if len(buf) == 0 || len(buf)%sdmmcBlockSize != 0 || len(buf) > sdmmcMultiBytes {
		return ErrSDMMCBlockSize
	}

	bounce := sdmmcMultiBuf()
	ptr := uintptr(unsafe.Pointer(&bounce[0]))
	physPtr := sdmmcPhysAddr(ptr)

	nDesc := (len(buf) + sdmmcMaxDescBytes - 1) / sdmmcMaxDescBytes
	for i := 0; i < nDesc; i++ {
		desc := sdmmcDescriptorAt(i)
		*desc = esp32p4sdmmc.Descriptor{}
		off := i * sdmmcMaxDescBytes
		size := len(buf) - off
		if size > sdmmcMaxDescBytes {
			size = sdmmcMaxDescBytes
		}
		desc.SetFirst(i == 0)
		desc.SetLast(i == nDesc-1)
		desc.SetOwned(true)
		desc.SetBuffer(physPtr+uint32(off), size)
		if i == nDesc-1 {
			desc.SetNext(0)
		} else {
			next := uintptr(unsafe.Pointer(sdmmcDescriptorAt(i + 1)))
			desc.SetNext(sdmmcPhysAddr(next))
		}
	}
	descPhys := sdmmcPhysAddr(uintptr(unsafe.Pointer(sdmmcDescriptorAt(0))))

	d.hw.BLKSIZ.Set(sdmmcBlockSize)
	d.hw.BYTCNT.Set(uint32(len(buf)))
	d.hw.DBADDR.Set(descPhys)
	d.hw.RINTSTS.Set(0xffffffff)
	d.hw.IDSTS.Set(0xffffffff)
	d.dmaEnable()
	d.hw.PLDMND.Set(1)

	arg := uint32(block)
	if !d.highCapacity {
		arg = uint32(block * sdmmcBlockSize)
	}
	opts := esp32p4sdmmc.CommandOptions{
		Index:    esp32p4sdmmc.CmdReadMultipleBlock,
		Response: esp32p4sdmmc.RespR1,
		CheckCRC: true,
		Data:     true,
		AutoStop: true,
	}
	if _, err := d.command(opts, arg); err != nil {
		d.captureTransferState()
		d.dmaStop()
		return err
	}
	d.hw.PLDMND.Set(1)
	sdmmcDelayUS(10)

	deadline := sdmmcDeadlineUS(sdmmcDataTimeout)
	for {
		d.hw.PLDMND.Set(1)
		rint := d.hw.RINTSTS.Get()
		idsts := d.hw.IDSTS.Get()
		if rint&sdmmcDataErrMask != 0 || idsts&sdmmcIDMACError != 0 {
			d.captureTransferState()
			d.hw.RINTSTS.Set(0xffffffff)
			d.hw.IDSTS.Set(0xffffffff)
			d.dmaStop()
			return ErrSDMMCData
		}
		if rint&sdmmcIntDataOver != 0 && idsts&sdmmcIDMACRI != 0 {
			break
		}
		if sdmmcTicks() > deadline {
			d.captureTransferState()
			d.dmaStop()
			return ErrSDMMCTimeout
		}
		sdmmcDelayUS(10)
	}

	d.hw.RINTSTS.Set(0xffffffff)
	d.hw.IDSTS.Set(0xffffffff)
	d.dmaStop()
	for i := 0; i < nDesc; i++ {
		sdmmcDescriptorAt(i).SetOwned(false)
	}
	copy(buf, bounce[:len(buf)])
	return nil
}

func (d *sdmmc) dmaEnable() {
	d.hw.CTRL.SetBits(sdmmcCtrlDMAEnable | sdmmcCtrlUseInternalDMA)
	d.hw.BMOD.SetBits(sdmmcBMODFB | sdmmcBMODDE)
	d.hw.IDINTEN.Set(sdmmcIDMACRI | sdmmcIDMACTI | sdmmcIDMACNI)
}

func (d *sdmmc) dmaStop() {
	d.hw.CTRL.ClearBits(sdmmcCtrlDMAEnable | sdmmcCtrlUseInternalDMA)
	d.hw.BMOD.ClearBits(sdmmcBMODFB | sdmmcBMODDE)
	d.hw.CTRL.SetBits(sdmmcCtrlDMAReset)
	deadline := sdmmcDeadlineUS(sdmmcCmdTimeout)
	for d.hw.CTRL.Get()&sdmmcCtrlDMAReset != 0 {
		if sdmmcTicks() > deadline {
			break
		}
		sdmmcDelayUS(10)
	}
}

// EraseBlocks is not implemented.  It returns ErrSDMMCEraseUnsupported;
// normal FAT writes use WriteBlocks and do not require an erase.
func (d *sdmmc) EraseBlocks(startBlock, numBlocks int64) error {
	return ErrSDMMCEraseUnsupported
}

// Sync is a no-op.  Card writes are committed by the transfer itself.
func (d *sdmmc) Sync() error { return nil }

// Close disables the card clock and powers the card off.  It does not
// reset controller state; do not issue block operations afterward.
func (d *sdmmc) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if !d.active {
		return nil
	}

	d.hw.CLKENA.ClearBits(1)
	if d.cfg.Power != NoPin {
		d.cfg.Power.High() // active-low: off
	}
	var err error
	if d.cfg.LDO != 0 {
		err = LDODisable(d.cfg.LDO)
	}
	d.initialized = false
	d.releaseActiveLocked()
	return err
}

func (d *sdmmc) releaseActiveLocked() {
	sdmmcActiveMu.Lock()
	if d.active {
		sdmmcActive = false
		d.active = false
	}
	sdmmcActiveMu.Unlock()
}

// NumberOfBlocks returns the number of 512-byte blocks on the card.
func (d *sdmmc) NumberOfBlocks() int64 {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.numBlocks
}

// BlockSize returns the card block size in bytes.
func (d *sdmmc) BlockSize() int { return sdmmcBlockSize }

// Size returns the card capacity in bytes.
func (d *sdmmc) Size() int64 {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.numBlocks * int64(sdmmcBlockSize)
}

// WriteBlockSize returns the block size accepted by WriteBlocks.
func (d *sdmmc) WriteBlockSize() int64 { return sdmmcBlockSize }

// EraseBlockSize returns the nominal erase granularity.  EraseBlocks is
// unsupported; this value is provided for block-device adapters.
func (d *sdmmc) EraseBlockSize() int64 { return sdmmcBlockSize }

// BusWidth returns the negotiated data bus width (1 or 4).
func (d *sdmmc) BusWidth() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.busWidth
}

// OCRRaw returns the OCR response from ACMD41.
func (d *sdmmc) OCRRaw() uint32 {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.ocr
}

// HighCapacity reports whether the card reported SDHC/SDXC capacity.
func (d *sdmmc) HighCapacity() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.highCapacity
}

// SDv2 reports whether CMD8 identified an SD v2 card.
func (d *sdmmc) SDv2() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.sdV2
}

// CSDVersion returns the decoded CSD structure version (1 or 2).
func (d *sdmmc) CSDVersion() uint8 {
	d.mu.Lock()
	defer d.mu.Unlock()
	return esp32p4sdmmc.CSDVersion(d.csd)
}

// DataLineLevels returns the current GPIO input levels of D0-D3. This is a
// read-only diagnostic; it does not drive the lines.
func (d *sdmmc) DataLineLevels() (d0, d1, d2, d3 bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.cfg.D0.Get(), d.cfg.D1.Get(), d.cfg.D2.Get(), d.cfg.D3.Get()
}

// DebugState returns the raw SDHOST control/clock registers for bring-up.
func (d *sdmmc) DebugState() (ctrl, clkdiv, clksrc, clkena uint32) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.hw.CTRL.Get(), d.hw.CLKDIV.Get(), d.hw.CLKSRC.Get(), d.hw.CLKENA.Get()
}

func (d *sdmmc) captureTransferState() {
	d.lastRint = d.hw.RINTSTS.Get()
	d.lastIdsts = d.hw.IDSTS.Get()
	d.lastCtrl = d.hw.CTRL.Get()
	d.lastBmod = d.hw.BMOD.Get()
	d.lastStatus = d.hw.STATUS.Get()
}

// TransferDebug returns the data-transfer status captured before cleanup.
func (d *sdmmc) TransferDebug() (rintsts, idsts, ctrl, bmod, status uint32) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.lastRint, d.lastIdsts, d.lastCtrl, d.lastBmod, d.lastStatus
}

// LastR1 returns the most recent R1 response word.
func (d *sdmmc) LastR1() uint32 {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.lastR1
}

// DMAState returns the raw IDMAC setup registers.
func (d *sdmmc) DMAState() (dbaddr, bytcnt, blksiz, pldmnd, idinten uint32) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.hw.DBADDR.Get(), d.hw.BYTCNT.Get(), d.hw.BLKSIZ.Get(), d.hw.PLDMND.Get(), d.hw.IDINTEN.Get()
}

// PIOState returns the raw status registers used by the FIFO path.
func (d *sdmmc) PIOState() (rintsts, status, bytcnt, blksiz uint32) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.hw.RINTSTS.Get(), d.hw.STATUS.Get(), d.hw.BYTCNT.Get(), d.hw.BLKSIZ.Get()
}

// DescriptorState returns the first four IDMAC descriptor words.
func (d *sdmmc) DescriptorState() (w0, w1, w2, w3 uint32) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.desc.Word[0], d.desc.Word[1], d.desc.Word[2], d.desc.Word[3]
}

// ClockState returns the raw HP_SYS_CLKRST and LP_AON reset registers.
func (d *sdmmc) ClockState() (peri1, peri2, ref0, ref2, aonRst uint32) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return esp.HP_SYS_CLKRST.PERI_CLK_CTRL01.Get(),
		esp.HP_SYS_CLKRST.PERI_CLK_CTRL02.Get(),
		esp.HP_SYS_CLKRST.REF_CLK_CTRL0.Get(),
		esp.HP_SYS_CLKRST.REF_CLK_CTRL2.Get(),
		esp.LP_AON_CLKRST.LP_AONCLKRST_HP_SDMMC_EMAC_RST_CTRL.Get()
}

// CIDRaw returns the 16-byte card identification register.
func (d *sdmmc) CIDRaw() [16]byte {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.cid
}

// CSDRaw returns the 16-byte card-specific data register.
func (d *sdmmc) CSDRaw() [16]byte {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.csd
}
