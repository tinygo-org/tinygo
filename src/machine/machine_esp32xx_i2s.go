//go:build esp32c3 || esp32c6 || esp32s3

package machine

import (
	"device/esp"
	"errors"
	"runtime/interrupt"
	"runtime/volatile"
	"unsafe"
)

// I2S on the ESP32-C3, ESP32-C6 and ESP32-S3 (I2S0 only).
//
// Samples are copied into a ring of RAM buffers that GDMA channel 0 plays or
// fills in the background, so audio keeps running between calls.
type I2S struct {
	Bus *esp.I2S_Type

	config     I2SConfig
	configured bool
	running    bool

	tx      i2sRing
	rx      i2sRing
	silence []uint32
	scratch []uint32
	writing volatile.Register8

	txDesc [i2sDescCount]i2sDescriptor
	rxDesc [i2sDescCount]i2sDescriptor
	txData [i2sDescCount]bool
	rxData [i2sDescCount]bool
	txNext int
	rxNext int
}

// i2sRing tracks buffers by counters that only increase. [done, queued) is
// owned by the hardware and [queued, filled) is ready for the hardware (TX)
// or for Read (RX).
type i2sRing struct {
	buf    []uint32
	done   uint32
	queued uint32
	filled uint32
	pos    int
}

// i2sDescriptor is a GDMA link descriptor (IDF hal/include/hal/dma_types.h).
type i2sDescriptor struct {
	dw0  uint32
	buf  uint32
	next uint32
}

const (
	i2sBufferWords = 128
	i2sBufferCount = 4
	i2sBufferBytes = i2sBufferWords * 4

	// GDMA can read the next descriptor early, so three descriptors leave one
	// buffer of time to update a finished descriptor.
	i2sDescCount = 3

	i2sDescOwnerDMA = 1 << 31
	i2sDescSucEOF   = 1 << 30
	i2sDescLenPos   = 12
)

// GDMA channel bits, the same on C3, C6 and S3 (IDF soc/*/include/soc/gdma_struct.h).
const (
	gdmaConf0Rst      = 1 << 0
	gdmaInConf0Burst  = 1<<2 | 1<<3
	gdmaOutConf0Burst = 1<<4 | 1<<5
	gdmaLinkAddrMask  = 0xfffff
	gdmaOutLinkStop   = 1 << 20
	gdmaOutLinkStart  = 1 << 21
	gdmaInLinkStop    = 1 << 21
	gdmaInLinkStart   = 1 << 22
	gdmaPeriSelI2S0   = 3
)

// I2S clock values from IDF hal/esp32c3/include/hal/i2s_ll.h.
const (
	i2sPLL160MClockSel   = 2
	i2sPLL160MClockFreq  = 160000000
	i2sMCLKMultiple      = 256
	i2sMCLKDividerMax    = 511
	i2sMCLKDivNumMax     = 255
	i2sUpdateTimeoutLoop = 100000
)

// i2sGDMA holds the GDMA channel 0 registers. The layout differs per chip.
type i2sGDMA struct {
	outConf0, outLink, outPeriSel, outIntEna, outIntClr, outIntSt, outEOFAddr *volatile.Register32
	inConf0, inLink, inPeriSel, inIntEna, inIntClr, inIntSt, inEOFAddr        *volatile.Register32
}

var I2S0 = I2S{Bus: esp.I2S0}

var (
	errI2SPins        = errors.New("i2s: missing pin")
	errI2SUnsupported = errors.New("i2s: unsupported configuration")
	errI2SNotRunning  = errors.New("i2s: not running")
)

// Configure sets up the I2S peripheral and starts it.
func (i2s *I2S) Configure(config I2SConfig) error {
	if i2s.configured {
		i2s.stop()
		i2s.configured = false
	}

	if config.DataFormat == I2SDataFormatDefault {
		config.DataFormat = I2SDataFormat16bit
	}
	if config.DataFormat != I2SDataFormat16bit && config.DataFormat != I2SDataFormat32bit {
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
	slave := config.ClockSource == I2SClockSourceExternal

	i2sEnableClocks()

	bits := uint32(config.DataFormat) - 1
	txConf := uint32(esp.I2S_TX_CONF_TX_TDM_EN | esp.I2S_TX_CONF_TX_PCM_BYPASS | esp.I2S_TX_CONF_TX_STOP_EN |
		esp.I2S_TX_CONF_TX_LEFT_ALIGN | esp.I2S_TX_CONF_TX_MONO_FST_VLD)
	rxConf := uint32(esp.I2S_RX_CONF_RX_TDM_EN | esp.I2S_RX_CONF_RX_PCM_BYPASS | esp.I2S_RX_CONF_RX_LEFT_ALIGN |
		esp.I2S_RX_CONF_RX_MONO_FST_VLD | 2<<esp.I2S_RX_CONF_RX_STOP_MODE_Pos)
	if slave {
		txConf |= esp.I2S_TX_CONF_TX_SLAVE_MOD
		rxConf |= esp.I2S_RX_CONF_RX_SLAVE_MOD
	}
	if config.Mode == I2SModeSourceReceiver {
		// RX uses the TX BCK and WS (esp-hal src/i2s/master.rs, Config::Tdm).
		txConf |= esp.I2S_TX_CONF_SIG_LOOPBACK
		rxConf |= esp.I2S_RX_CONF_RX_SLAVE_MOD
	}
	msbShift := uint32(0)
	if config.Standard == I2StandardPhilips {
		msbShift = 1
	}

	i2s.Bus.TX_CONF.Set(txConf)
	i2s.Bus.SetTX_CONF1_TX_TDM_WS_WIDTH(bits)
	i2s.Bus.SetTX_CONF1_TX_BITS_MOD(bits)
	i2s.Bus.SetTX_CONF1_TX_HALF_SAMPLE_BITS(bits)
	i2s.Bus.SetTX_CONF1_TX_TDM_CHAN_BITS(bits)
	i2s.Bus.SetTX_CONF1_TX_MSB_SHIFT(msbShift)
	i2s.Bus.TX_TDM_CTRL.Set(1<<esp.I2S_TX_TDM_CTRL_TX_TDM_TOT_CHAN_NUM_Pos | 3)

	i2s.Bus.RX_CONF.Set(rxConf)
	i2s.Bus.SetRX_CONF1_RX_TDM_WS_WIDTH(bits)
	i2s.Bus.SetRX_CONF1_RX_BITS_MOD(bits)
	i2s.Bus.SetRX_CONF1_RX_HALF_SAMPLE_BITS(bits)
	i2s.Bus.SetRX_CONF1_RX_TDM_CHAN_BITS(bits)
	i2s.Bus.SetRX_CONF1_RX_MSB_SHIFT(msbShift)
	i2s.Bus.RX_TDM_CTRL.Set(1<<esp.I2S_RX_TDM_CTRL_RX_TDM_TOT_CHAN_NUM_Pos | 3)

	i2s.config = config
	if err := i2s.setClock(config.AudioFrequency); err != nil {
		return err
	}

	i2s.configurePins(tx, rx, slave)

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

	for i := range i2s.txDesc {
		next := (i + 1) % i2sDescCount
		volatile.StoreUint32(&i2s.txDesc[i].next, uint32(uintptr(unsafe.Pointer(&i2s.txDesc[next]))))
		volatile.StoreUint32(&i2s.rxDesc[i].next, uint32(uintptr(unsafe.Pointer(&i2s.rxDesc[next]))))
	}

	i2sGDMARegs.outIntEna.ClearBits(gdmaOutEOF)
	i2sGDMARegs.inIntEna.ClearBits(gdmaInSucEOF)
	i2sEnableInterrupt()

	i2s.configured = true
	i2s.start()
	return nil
}

func (i2s *I2S) configurePins(tx, rx, slave bool) {
	c := i2s.config
	sckSignal, wsSignal := uint32(i2sSignalTxBCK), uint32(i2sSignalTxWS)
	if !tx {
		sckSignal, wsSignal = i2sSignalRxBCK, i2sSignalRxWS
	}
	clockMode := PinOutput
	if slave {
		clockMode = PinInput
	}
	c.SCK.configure(PinConfig{Mode: clockMode}, sckSignal)
	c.WS.configure(PinConfig{Mode: clockMode}, wsSignal)
	if tx {
		c.SDO.configure(PinConfig{Mode: PinOutput}, i2sSignalTxSD)
	}
	if rx {
		c.SDI.configure(PinConfig{Mode: PinInput}, i2sSignalRxSD)
	}
}

// SetSampleFrequency sets the sample rate.
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

// setClock derives MCLK = 256 * freq from the 160 MHz PLL with the
// fractional divider search from IDF components/hal/i2s_hal.c.
func (i2s *I2S) setClock(freq uint32) error {
	if freq == 0 {
		return ErrInvalidSampleFrequency
	}
	mclk := uint64(freq) * i2sMCLKMultiple
	integ := uint64(i2sPLL160MClockFreq) / mclk
	rem := uint64(i2sPLL160MClockFreq) % mclk
	denom, numer := uint64(1), uint64(0)
	if rem != 0 {
		if rem*2*i2sMCLKDividerMax <= mclk*(2*i2sMCLKDividerMax-1) {
			best := ^uint64(0)
			for a := uint64(2); a <= i2sMCLKDividerMax; a++ {
				b := (a*rem*2 + mclk) / (2 * mclk)
				ma, mb := rem*a, mclk*b
				diff := ma - mb
				if mb > ma {
					diff = mb - ma
				}
				if diff < best {
					denom, numer, best = a, b, diff
				}
				if diff == 0 {
					break
				}
			}
		} else {
			integ++
		}
	}
	if integ < 2 || integ > i2sMCLKDivNumMax {
		return ErrInvalidSampleFrequency
	}

	// Same as i2s_ll_tx_set_mclk in IDF hal/esp32c3/include/hal/i2s_ll.h.
	var x, y, z, yn1 uint64
	if denom != 0 && numer != 0 {
		if numer*2 > denom {
			yn1 = 1
			z = denom - numer
		} else {
			z = numer
		}
		x = denom/z - 1
		y = denom % z
	}

	bckDiv := uint32(i2sMCLKMultiple/(2*uint32(i2s.config.DataFormat))) - 1
	i2sSetTxClock(uint32(integ), uint32(x), uint32(y), uint32(z), uint32(yn1))
	i2sSetRxClock(uint32(integ), uint32(x), uint32(y), uint32(z), uint32(yn1))
	i2s.Bus.SetTX_CONF1_TX_BCK_DIV_NUM(bckDiv)
	i2s.Bus.SetRX_CONF1_RX_BCK_DIV_NUM(bckDiv)
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
// left in the low half. At 32-bit each value is one sample, left then right.
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
}

// TX and RX share one edge-triggered CPU interrupt on C3 and C6. Loop until both
// are clear, otherwise a late event keeps the line high and no new edge comes.
func (i2s *I2S) handleInterrupt() {
	g := &i2sGDMARegs
	for {
		out := g.outIntSt.Get()&gdmaOutEOF != 0
		in := g.inIntSt.Get()&gdmaInSucEOF != 0
		if !out && !in {
			return
		}
		if out {
			g.outIntClr.Set(gdmaOutEOF)
			last := i2sDescIndex(&i2s.txDesc, g.outEOFAddr.Get())
			for last >= 0 {
				d := i2s.txNext
				i2s.txNext = (d + 1) % i2sDescCount
				i2s.nextTx(d)
				if d == last {
					break
				}
			}
		}
		if in {
			g.inIntClr.Set(gdmaInSucEOF)
			last := i2sDescIndex(&i2s.rxDesc, g.inEOFAddr.Get())
			for last >= 0 {
				d := i2s.rxNext
				i2s.rxNext = (d + 1) % i2sDescCount
				i2s.nextRx(d)
				if d == last {
					break
				}
			}
		}
	}
}

func i2sDescIndex(desc *[i2sDescCount]i2sDescriptor, addr uint32) int {
	for i := range desc {
		if uint32(uintptr(unsafe.Pointer(&desc[i]))) == addr {
			return i
		}
	}
	return -1
}

// nextTx runs when GDMA has read all of descriptor d. The descriptor gets
// the next full buffer or silence.
func (i2s *I2S) nextTx(d int) {
	r := &i2s.tx
	if i2s.txData[d] {
		r.done++
	}
	if r.queued == r.filled && r.pos > 0 && i2s.writing.Get() == 0 {
		clear(r.slot(r.filled)[r.pos:])
		r.pos = 0
		r.filled++
	}
	buf := i2s.silence
	i2s.txData[d] = r.queued != r.filled
	if i2s.txData[d] {
		buf = r.slot(r.queued)
		r.queued++
	}
	i2s.setTxDesc(d, buf)
}

// nextRx runs when GDMA has filled descriptor d. RX counters use done for
// buffers given to the hardware, filled for buffers ready to read and queued
// for buffers already read.
func (i2s *I2S) nextRx(d int) {
	r := &i2s.rx
	if i2s.rxData[d] {
		r.filled++
	}
	// Read is too slow when all buffers are in use, so drop these samples.
	buf := i2s.scratch
	i2s.rxData[d] = r.done-r.queued < i2sBufferCount
	if i2s.rxData[d] {
		buf = r.slot(r.done)
		r.done++
	}
	i2s.setRxDesc(d, buf)
}

func (i2s *I2S) setTxDesc(d int, buf []uint32) {
	volatile.StoreUint32(&i2s.txDesc[d].buf, uint32(uintptr(unsafe.Pointer(&buf[0]))))
	volatile.StoreUint32(&i2s.txDesc[d].dw0, i2sDescOwnerDMA|i2sDescSucEOF|i2sBufferBytes<<i2sDescLenPos|i2sBufferBytes)
}

func (i2s *I2S) setRxDesc(d int, buf []uint32) {
	volatile.StoreUint32(&i2s.rxDesc[d].buf, uint32(uintptr(unsafe.Pointer(&buf[0]))))
	volatile.StoreUint32(&i2s.rxDesc[d].dw0, i2sDescOwnerDMA|i2sBufferBytes)
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

// start follows i2s_start in IDF components/esp_driver_i2s/i2s_common.c.
func (i2s *I2S) start() {
	if i2s.running {
		return
	}
	g := &i2sGDMARegs
	i2s.tx.reset()
	i2s.rx.reset()
	i2s.writing.Set(0)
	i2s.running = true

	// RX starts first because in loopback mode it runs from the TX clock.
	if i2s.rxEnabled() {
		for d := range i2s.rxDesc {
			i2s.setRxDesc(d, i2s.rx.slot(uint32(d)))
			i2s.rxData[d] = true
		}
		i2s.rx.done = i2sDescCount
		i2s.rxNext = 0

		i2s.Bus.RX_CONF.SetBits(esp.I2S_RX_CONF_RX_RESET | esp.I2S_RX_CONF_RX_FIFO_RESET)
		i2s.Bus.RX_CONF.ClearBits(esp.I2S_RX_CONF_RX_RESET | esp.I2S_RX_CONF_RX_FIFO_RESET)
		i2s.Bus.RXEOF_NUM.Set(i2sBufferBytes)
		g.inConf0.Set(gdmaConf0Rst)
		g.inConf0.Set(gdmaInConf0Burst)
		g.inPeriSel.Set(gdmaPeriSelI2S0)
		g.inLink.Set(uint32(uintptr(unsafe.Pointer(&i2s.rxDesc[0])))&gdmaLinkAddrMask | gdmaInLinkStart)
		g.inIntClr.Set(gdmaInSucEOF)
		g.inIntEna.SetBits(gdmaInSucEOF)
		i2s.update(&i2s.Bus.RX_CONF, esp.I2S_RX_CONF_RX_UPDATE)
		i2s.Bus.RX_CONF.SetBits(esp.I2S_RX_CONF_RX_START)
	}

	if i2s.txEnabled() {
		for d := range i2s.txDesc {
			i2s.setTxDesc(d, i2s.silence)
			i2s.txData[d] = false
		}
		i2s.txNext = 0

		i2s.Bus.TX_CONF.SetBits(esp.I2S_TX_CONF_TX_RESET | esp.I2S_TX_CONF_TX_FIFO_RESET)
		i2s.Bus.TX_CONF.ClearBits(esp.I2S_TX_CONF_TX_RESET | esp.I2S_TX_CONF_TX_FIFO_RESET)
		g.outConf0.Set(gdmaConf0Rst)
		g.outConf0.Set(gdmaOutConf0Burst)
		g.outPeriSel.Set(gdmaPeriSelI2S0)
		g.outLink.Set(uint32(uintptr(unsafe.Pointer(&i2s.txDesc[0])))&gdmaLinkAddrMask | gdmaOutLinkStart)
		g.outIntClr.Set(gdmaOutEOF)
		g.outIntEna.SetBits(gdmaOutEOF)
		i2s.update(&i2s.Bus.TX_CONF, esp.I2S_TX_CONF_TX_UPDATE)
		i2s.Bus.TX_CONF.SetBits(esp.I2S_TX_CONF_TX_START)
	}
}

// update syncs the register values into the I2S clock domain.
func (i2s *I2S) update(reg *volatile.Register32, bit uint32) {
	reg.SetBits(bit)
	for i := 0; i < i2sUpdateTimeoutLoop && reg.HasBits(bit); i++ {
	}
}

func (i2s *I2S) stop() {
	if !i2s.running {
		return
	}
	g := &i2sGDMARegs
	g.outIntEna.ClearBits(gdmaOutEOF)
	g.inIntEna.ClearBits(gdmaInSucEOF)
	i2s.Bus.TX_CONF.ClearBits(esp.I2S_TX_CONF_TX_START)
	i2s.Bus.RX_CONF.ClearBits(esp.I2S_RX_CONF_RX_START)
	g.outLink.Set(gdmaOutLinkStop)
	g.inLink.Set(gdmaInLinkStop)
	g.outIntClr.Set(gdmaOutEOF)
	g.inIntClr.Set(gdmaInSucEOF)
	i2s.running = false
}

// i2sSetClockDiv writes a fixed divider first and then the real one, as
// i2s_ll_tx_set_mclk does (IDF hal/esp32c3/include/hal/i2s_ll.h).
func i2sSetClockDiv(conf, divConf *volatile.Register32, divNumPos, div, x, y, z, yn1 uint32) {
	conf.ReplaceBits(7, 0xff, uint8(divNumPos))
	divConf.Set(i2sClockDivConf(317, 7, 3, 0))
	conf.ReplaceBits(div, 0xff, uint8(divNumPos))
	divConf.Set(i2sClockDivConf(x, y, z, yn1))
}

func i2sClockDivConf(x, y, z, yn1 uint32) uint32 {
	return x<<18 | y<<9 | z | yn1<<27
}
