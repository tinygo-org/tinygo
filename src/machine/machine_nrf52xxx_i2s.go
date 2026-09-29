//go:build nrf52 || nrf52833 || nrf52840

package machine

import (
	"device/arm"
	"device/nrf"
	"errors"
	"runtime/interrupt"
	"runtime/volatile"
	"unsafe"
)

// I2S on the nRF52832, nRF52833 and nRF52840.
//
// Samples are copied into a ring of RAM buffers that EasyDMA plays or fills
// in the background, so audio keeps running between calls.
type I2S struct {
	Bus *nrf.I2S_Type

	config     I2SConfig
	configured bool
	running    bool
	hfclkOwned bool

	tx      i2sRing
	rx      i2sRing
	silence []uint32
	scratch []uint32
	writing volatile.Register8
}

// i2sRing tracks buffers by counters that only increase. [done, queued) is
// owned by the hardware and [queued, filled) is ready for the hardware (TX)
// or for Read (RX).
type i2sRing struct {
	buf        []uint32
	done       uint32
	queued     uint32
	filled     uint32
	pos        int
	curIsData  bool
	nextIsData bool
}

const (
	i2sBufferWords = 128
	i2sBufferCount = 4
)

var I2S0 = I2S{Bus: nrf.I2S}

var (
	errI2SPins        = errors.New("i2s: missing pin")
	errI2SUnsupported = errors.New("i2s: unsupported configuration")
	errI2SNotRunning  = errors.New("i2s: not running")
)

// Configure sets up the I2S peripheral and starts it.
// If you use Bluetooth, enable it first since the SoftDevice then owns the clock.
func (i2s *I2S) Configure(config I2SConfig) error {
	if i2s.configured {
		i2s.stop()
		i2s.configured = false
	}

	if config.DataFormat == I2SDataFormatDefault {
		config.DataFormat = I2SDataFormat16bit
	}
	if config.DataFormat != I2SDataFormat16bit && config.DataFormat != I2SDataFormat24bit {
		return errI2SUnsupported
	}
	if config.AudioFrequency == 0 {
		config.AudioFrequency = 44100
	}
	if config.Mode == I2SModePDM || config.MainClockOutput {
		return errI2SUnsupported
	}

	tx := config.Mode == I2SModeSource || config.Mode == I2SModeSourceReceiver
	rx := config.Mode == I2SModeReceiver || config.Mode == I2SModeSourceReceiver
	if config.SCK == NoPin || config.WS == NoPin || (tx && config.SDO == NoPin) || (rx && config.SDI == NoPin) {
		return errI2SPins
	}

	i2s.Bus.CONFIG.MODE.Set(nrf.I2S_CONFIG_MODE_MODE_Master)
	if config.ClockSource == I2SClockSourceExternal {
		i2s.Bus.CONFIG.MODE.Set(nrf.I2S_CONFIG_MODE_MODE_Slave)
	}

	switch config.Standard {
	case I2StandardPhilips:
		i2s.Bus.CONFIG.FORMAT.Set(nrf.I2S_CONFIG_FORMAT_FORMAT_I2S)
		i2s.Bus.CONFIG.ALIGN.Set(nrf.I2S_CONFIG_ALIGN_ALIGN_Left)
	case I2SStandardMSB:
		i2s.Bus.CONFIG.FORMAT.Set(nrf.I2S_CONFIG_FORMAT_FORMAT_Aligned)
		i2s.Bus.CONFIG.ALIGN.Set(nrf.I2S_CONFIG_ALIGN_ALIGN_Left)
	case I2SStandardLSB:
		i2s.Bus.CONFIG.FORMAT.Set(nrf.I2S_CONFIG_FORMAT_FORMAT_Aligned)
		i2s.Bus.CONFIG.ALIGN.Set(nrf.I2S_CONFIG_ALIGN_ALIGN_Right)
	}

	i2s.Bus.CONFIG.SWIDTH.Set(nrf.I2S_CONFIG_SWIDTH_SWIDTH_16Bit)
	if config.DataFormat == I2SDataFormat24bit {
		i2s.Bus.CONFIG.SWIDTH.Set(nrf.I2S_CONFIG_SWIDTH_SWIDTH_24Bit)
	}
	i2s.Bus.CONFIG.CHANNELS.Set(nrf.I2S_CONFIG_CHANNELS_CHANNELS_Stereo)
	i2s.Bus.CONFIG.TXEN.Set(0)
	i2s.Bus.CONFIG.RXEN.Set(0)
	if tx {
		i2s.Bus.CONFIG.TXEN.Set(nrf.I2S_CONFIG_TXEN_TXEN_Enabled)
	}
	if rx {
		i2s.Bus.CONFIG.RXEN.Set(nrf.I2S_CONFIG_RXEN_RXEN_Enabled)
	}
	i2s.Bus.CONFIG.MCKEN.Set(nrf.I2S_CONFIG_MCKEN_MCKEN_Enabled)

	i2s.config = config
	if err := i2s.setClock(config.AudioFrequency); err != nil {
		return err
	}

	if i2s.silence == nil {
		i2s.silence = make([]uint32, i2sBufferWords)
	}
	if rx && i2s.scratch == nil {
		i2s.scratch = make([]uint32, i2sBufferWords)
	}
	if tx && i2s.tx.buf == nil {
		i2s.tx.buf = make([]uint32, i2sBufferWords*i2sBufferCount)
	}
	if rx && i2s.rx.buf == nil {
		i2s.rx.buf = make([]uint32, i2sBufferWords*i2sBufferCount)
	}
	i2s.Bus.RXTXD.MAXCNT.Set(i2sBufferWords)

	intr := interrupt.New(nrf.IRQ_I2S, func(interrupt.Interrupt) {
		I2S0.handleInterrupt()
	})
	intr.SetPriority(0xc0) // low priority
	intr.Enable()

	i2s.configured = true
	i2s.start()
	return nil
}

// SetSampleFrequency sets the sample rate. The MCK divider only has fixed
// steps, so the rate is the closest one within 2%.
func (i2s *I2S) SetSampleFrequency(freq uint32) error {
	running := i2s.running
	if running {
		i2s.stop()
	}
	err := i2s.setClock(freq)
	if err == nil {
		i2s.config.AudioFrequency = freq
	}
	if running {
		i2s.start()
	}
	return err
}

func (i2s *I2S) setClock(freq uint32) error {
	if freq == 0 {
		return ErrInvalidSampleFrequency
	}
	// MCKFREQ values from lib/nrfx/mdk/nrf52840_bitfields.h line 2310.
	dividers := [...]struct{ reg, div uint32 }{
		{nrf.I2S_CONFIG_MCKFREQ_MCKFREQ_32MDIV8, 8},
		{nrf.I2S_CONFIG_MCKFREQ_MCKFREQ_32MDIV10, 10},
		{nrf.I2S_CONFIG_MCKFREQ_MCKFREQ_32MDIV11, 11},
		{nrf.I2S_CONFIG_MCKFREQ_MCKFREQ_32MDIV15, 15},
		{nrf.I2S_CONFIG_MCKFREQ_MCKFREQ_32MDIV16, 16},
		{nrf.I2S_CONFIG_MCKFREQ_MCKFREQ_32MDIV21, 21},
		{nrf.I2S_CONFIG_MCKFREQ_MCKFREQ_32MDIV23, 23},
		{nrf.I2S_CONFIG_MCKFREQ_MCKFREQ_32MDIV30, 30},
		{nrf.I2S_CONFIG_MCKFREQ_MCKFREQ_32MDIV31, 31},
		{nrf.I2S_CONFIG_MCKFREQ_MCKFREQ_32MDIV32, 32},
		{nrf.I2S_CONFIG_MCKFREQ_MCKFREQ_32MDIV42, 42},
		{nrf.I2S_CONFIG_MCKFREQ_MCKFREQ_32MDIV63, 63},
		{nrf.I2S_CONFIG_MCKFREQ_MCKFREQ_32MDIV125, 125},
	}
	ratios := [...]uint32{32, 48, 64, 96, 128, 192, 256, 384, 512}

	// The MCK to LRCK ratio must be a multiple of 2 times the sample width.
	// See lib/nrfx/hal/nrf_i2s.h line 648.
	frame := 2 * uint32(i2s.config.DataFormat)
	bestReg, bestRatio, bestErr := uint32(0), 0, uint32(0xffffffff)
	for _, d := range dividers {
		for r, ratio := range ratios {
			if ratio%frame != 0 {
				continue
			}
			lrck := 32000000 / d.div / ratio
			diff := lrck - freq
			if lrck < freq {
				diff = freq - lrck
			}
			if diff < bestErr {
				bestReg, bestRatio, bestErr = d.reg, r, diff
			}
		}
	}
	if uint64(bestErr)*50 > uint64(freq) {
		return ErrInvalidSampleFrequency
	}
	i2s.Bus.CONFIG.MCKFREQ.Set(bestReg)
	i2s.Bus.CONFIG.RATIO.Set(uint32(bestRatio))
	return nil
}

// WriteMono plays each sample on both channels. It returns once all samples
// are queued. Only 16-bit samples are supported.
func (i2s *I2S) WriteMono(b []uint16) (int, error) {
	if i2s.config.DataFormat != I2SDataFormat16bit {
		return 0, errI2SUnsupported
	}
	return i2sWrite(i2s, b, true)
}

// WriteStereo queues stereo frames. At 16-bit each value is one frame with
// left in the low half. At 24-bit each value is one sample, left then right.
func (i2s *I2S) WriteStereo(b []uint32) (int, error) {
	return i2sWrite(i2s, b, false)
}

// ReadMono reads the left channel. Only 16-bit samples are supported.
func (i2s *I2S) ReadMono(b []uint16) (int, error) {
	if i2s.config.DataFormat != I2SDataFormat16bit {
		return 0, errI2SUnsupported
	}
	return i2sRead(i2s, b)
}

// ReadStereo reads frames in the same layout WriteStereo uses.
func (i2s *I2S) ReadStereo(b []uint32) (int, error) {
	return i2sRead(i2s, b)
}

func i2sWrite[T uint16 | uint32](i2s *I2S, b []T, mono bool) (int, error) {
	if !i2s.txEnabled() {
		return 0, errI2SUnsupported
	}
	if !i2s.running {
		return 0, errI2SNotRunning
	}
	r := &i2s.tx
	i2s.writing.Set(1)
	defer i2s.writing.Set(0)
	for i := 0; i < len(b); {
		if r.pos == 0 {
			for volatile.LoadUint32(&r.filled)-volatile.LoadUint32(&r.done) >= i2sBufferCount {
				if !i2s.running {
					return i, errI2SNotRunning
				}
				gosched()
			}
		}
		buf := r.slot(r.filled)
		n := min(len(buf)-r.pos, len(b)-i)
		for j := 0; j < n; j++ {
			v := uint32(b[i+j])
			if mono {
				v |= v << 16
			}
			buf[r.pos+j] = v
		}
		i += n
		r.pos += n
		if r.pos == len(buf) {
			mask := interrupt.Disable()
			r.pos = 0
			r.filled++
			interrupt.Restore(mask)
		}
	}
	return len(b), nil
}

func i2sRead[T uint16 | uint32](i2s *I2S, b []T) (int, error) {
	if !i2s.rxEnabled() {
		return 0, errI2SUnsupported
	}
	if !i2s.running {
		return 0, errI2SNotRunning
	}
	r := &i2s.rx
	for i := 0; i < len(b); {
		for volatile.LoadUint32(&r.filled) == volatile.LoadUint32(&r.queued) {
			if !i2s.running {
				return i, errI2SNotRunning
			}
			gosched()
		}
		buf := r.slot(r.queued)
		n := min(len(buf)-r.pos, len(b)-i)
		for j := 0; j < n; j++ {
			b[i+j] = T(buf[r.pos+j])
		}
		i += n
		r.pos += n
		if r.pos == len(buf) {
			mask := interrupt.Disable()
			r.pos = 0
			r.queued++
			interrupt.Restore(mask)
		}
	}
	return len(b), nil
}

func (i2s *I2S) txEnabled() bool {
	return i2s.config.Mode == I2SModeSource || i2s.config.Mode == I2SModeSourceReceiver
}

func (i2s *I2S) rxEnabled() bool {
	return i2s.config.Mode == I2SModeReceiver || i2s.config.Mode == I2SModeSourceReceiver
}

func (r *i2sRing) slot(n uint32) []uint32 {
	start := (n % i2sBufferCount) * i2sBufferWords
	return r.buf[start : start+i2sBufferWords]
}

func (r *i2sRing) reset() {
	r.done, r.queued, r.filled, r.pos = 0, 0, 0, 0
	r.curIsData, r.nextIsData = false, false
}

func (i2s *I2S) handleInterrupt() {
	if i2s.Bus.EVENTS_TXPTRUPD.Get() != 0 {
		i2s.Bus.EVENTS_TXPTRUPD.Set(0)
		i2s.nextTx()
	}
	if i2s.Bus.EVENTS_RXPTRUPD.Get() != 0 {
		i2s.Bus.EVENTS_RXPTRUPD.Set(0)
		i2s.nextRx()
	}
}

// nextTx runs when the hardware took TXD.PTR. The buffer that was playing
// is done, so TXD.PTR gets the next full buffer or silence.
func (i2s *I2S) nextTx() {
	r := &i2s.tx
	if r.curIsData {
		r.done++
	}
	r.curIsData = r.nextIsData
	if r.queued == r.filled && r.pos > 0 && i2s.writing.Get() == 0 {
		clear(r.slot(r.filled)[r.pos:])
		r.pos = 0
		r.filled++
	}
	if r.queued != r.filled {
		i2s.Bus.TXD.PTR.Set(uint32(uintptr(unsafe.Pointer(&r.slot(r.queued)[0]))))
		r.queued++
		r.nextIsData = true
	} else {
		i2s.Bus.TXD.PTR.Set(uint32(uintptr(unsafe.Pointer(&i2s.silence[0]))))
		r.nextIsData = false
	}
}

// nextRx runs when the hardware took RXD.PTR. The buffer that was filling is
// ready for Read. RX counters use done for buffers given to the hardware,
// filled for buffers ready to read and queued for buffers already read.
func (i2s *I2S) nextRx() {
	r := &i2s.rx
	if r.curIsData {
		r.filled++
	}
	r.curIsData = r.nextIsData
	if r.done-r.queued < i2sBufferCount {
		i2s.Bus.RXD.PTR.Set(uint32(uintptr(unsafe.Pointer(&r.slot(r.done)[0]))))
		r.done++
		r.nextIsData = true
	} else {
		// Read is too slow, so drop this buffer of samples.
		i2s.Bus.RXD.PTR.Set(uint32(uintptr(unsafe.Pointer(&i2s.scratch[0]))))
		r.nextIsData = false
	}
}

// Enable starts or stops the I2S peripheral. Queued samples are dropped.
func (i2s *I2S) Enable(enabled bool) {
	if !i2s.configured {
		return
	}
	if enabled {
		i2s.start()
	} else {
		i2s.stop()
	}
}

func (i2s *I2S) start() {
	if i2s.running {
		return
	}
	i2s.requestHFXO()

	c := i2s.config
	sckMode := PinOutput
	if c.ClockSource == I2SClockSourceExternal {
		sckMode = PinInput
	}
	c.SCK.Configure(PinConfig{Mode: sckMode})
	c.WS.Configure(PinConfig{Mode: sckMode})
	i2s.Bus.PSEL.SCK.Set(uint32(c.SCK))
	i2s.Bus.PSEL.LRCK.Set(uint32(c.WS))
	i2s.Bus.PSEL.MCK.Set(i2sDisconnected)
	i2s.Bus.PSEL.SDOUT.Set(i2sDisconnected)
	i2s.Bus.PSEL.SDIN.Set(i2sDisconnected)

	i2s.tx.reset()
	i2s.rx.reset()
	i2s.writing.Set(0)
	if i2s.txEnabled() {
		c.SDO.Configure(PinConfig{Mode: PinOutput})
		i2s.Bus.PSEL.SDOUT.Set(uint32(c.SDO))
		i2s.Bus.TXD.PTR.Set(uint32(uintptr(unsafe.Pointer(&i2s.silence[0]))))
	}
	if i2s.rxEnabled() {
		c.SDI.Configure(PinConfig{Mode: PinInput})
		i2s.Bus.PSEL.SDIN.Set(uint32(c.SDI))
		i2s.Bus.RXD.PTR.Set(uint32(uintptr(unsafe.Pointer(&i2s.rx.slot(0)[0]))))
		i2s.rx.done = 1
		i2s.rx.nextIsData = true
	}

	i2s.Bus.EVENTS_TXPTRUPD.Set(0)
	i2s.Bus.EVENTS_RXPTRUPD.Set(0)
	i2s.Bus.EVENTS_STOPPED.Set(0)
	i2s.Bus.ENABLE.Set(nrf.I2S_ENABLE_ENABLE_Enabled)
	i2s.Bus.INTENSET.Set(nrf.I2S_INTENSET_TXPTRUPD_Msk | nrf.I2S_INTENSET_RXPTRUPD_Msk)
	i2s.running = true
	i2s.Bus.TASKS_START.Set(1)
}

const i2sDisconnected = nrf.I2S_PSEL_SCK_CONNECT_Disconnected << nrf.I2S_PSEL_SCK_CONNECT_Pos

// stop follows nrfx_i2s_stop and nrfx_i2s_uninit in lib/nrfx/drivers/src/nrfx_i2s.c.
func (i2s *I2S) stop() {
	if !i2s.running {
		return
	}
	// Disable the interrupts first to skip stray PTRUPD events (nRF52 anomaly 55).
	i2s.Bus.INTENCLR.Set(nrf.I2S_INTENCLR_TXPTRUPD_Msk | nrf.I2S_INTENCLR_RXPTRUPD_Msk)
	i2s.Bus.TASKS_STOP.Set(1)

	// STOP does not switch off all resources (nRF52 anomaly 194).
	base := uintptr(unsafe.Pointer(i2s.Bus))
	(*volatile.Register32)(unsafe.Pointer(base + 0x38)).Set(1)
	(*volatile.Register32)(unsafe.Pointer(base + 0x3C)).Set(1)

	for i2s.Bus.EVENTS_STOPPED.Get() == 0 {
		gosched()
	}
	i2s.Bus.EVENTS_STOPPED.Set(0)
	i2s.Bus.EVENTS_TXPTRUPD.Set(0)
	i2s.Bus.EVENTS_RXPTRUPD.Set(0)
	i2s.Bus.ENABLE.Set(nrf.I2S_ENABLE_ENABLE_Disabled)

	// Disabling I2S does not release the pins (nRF52 anomaly 196).
	i2s.Bus.PSEL.MCK.Set(i2sDisconnected)
	i2s.Bus.PSEL.SCK.Set(i2sDisconnected)
	i2s.Bus.PSEL.LRCK.Set(i2sDisconnected)
	i2s.Bus.PSEL.SDOUT.Set(i2sDisconnected)
	i2s.Bus.PSEL.SDIN.Set(i2sDisconnected)

	i2s.running = false
	i2s.releaseHFXO()
}

// SoftDevice clock calls, from nrf_soc.h in the S140 6.1.1 headers.
const (
	sdClockHFCLKRequest   = 0x2C + 22
	sdClockHFCLKRelease   = 0x2C + 23
	sdClockHFCLKIsRunning = 0x2C + 24
)

var i2sHFCLKRunning uint32

// requestHFXO starts the 32 MHz crystal for better MCK accuracy and jitter, as
// Zephyr does by default (dts/bindings/i2s/nordic,nrf-i2s.yaml line 31).
func (i2s *I2S) requestHFXO() {
	if isSoftDeviceEnabled() {
		arm.SVCall0(sdClockHFCLKRequest)
		i2s.hfclkOwned = true
		for {
			arm.SVCall1(sdClockHFCLKIsRunning, &i2sHFCLKRunning)
			if i2sHFCLKRunning != 0 {
				return
			}
			gosched()
		}
	}
	if nrf.CLOCK.GetHFCLKSTAT_SRC() == nrf.CLOCK_HFCLKSTAT_SRC_Xtal {
		return
	}
	nrf.CLOCK.EVENTS_HFCLKSTARTED.Set(0)
	nrf.CLOCK.TASKS_HFCLKSTART.Set(1)
	for nrf.CLOCK.EVENTS_HFCLKSTARTED.Get() == 0 {
		gosched()
	}
	i2s.hfclkOwned = true
}

func (i2s *I2S) releaseHFXO() {
	if !i2s.hfclkOwned {
		return
	}
	i2s.hfclkOwned = false
	if isSoftDeviceEnabled() {
		arm.SVCall0(sdClockHFCLKRelease)
		return
	}
	nrf.CLOCK.TASKS_HFCLKSTOP.Set(1)
}
