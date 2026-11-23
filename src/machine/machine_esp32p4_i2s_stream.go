//go:build esp32p4

// Continuous two-buffer I2S TX streaming for ESP32-P4.
//
// I2STx.Write builds and tears down a DMA chain for every buffer.  That is
// fine for a single large buffer, but streaming callers hear a small gap at
// every boundary while the chain is restarted.  I2SStream keeps a circular
// descriptor ring running: two caller-provided buffers alternate, and the
// caller refills the buffer that just finished while the other one plays.

package machine

import (
	"device/esp"
	"unsafe"
)

// I2SStream is a circular two-buffer DMA stream.  Create it with
// I2STx.NewStream, fill both buffers, call Start, then repeatedly WaitFree
// and Update the returned buffer.
type I2SStream struct {
	tx      *I2STx
	bufs    [2][]int16
	descs   [2][]i2sDMADesc
	last    int
	started bool
}

// NewStream prepares a circular stream over two equal-sized sample buffers.
// The buffers must remain valid until Close.  The descriptor pool is reused;
// create a new stream after Close for another file.
func (tx *I2STx) NewStream(a, b []int16) (*I2SStream, error) {
	if len(a) == 0 || len(a) != len(b) {
		return nil, ErrI2STxBadSampleRate
	}
	per := (len(a) + maxI2SDescBytes/2 - 1) / (maxI2SDescBytes / 2)
	if per*2 > len(tx.descs) {
		return nil, ErrI2STxBufferTooLarge
	}

	s := &I2SStream{
		tx:    tx,
		bufs:  [2][]int16{a, b},
		descs: [2][]i2sDMADesc{tx.descs[:per], tx.descs[per : 2*per]},
	}
	for buf := 0; buf < 2; buf++ {
		for i := range s.descs[buf] {
			start := i * (maxI2SDescBytes / 2)
			end := start + maxI2SDescBytes/2
			if end > len(s.bufs[buf]) {
				end = len(s.bufs[buf])
			}
			s.descs[buf][i] = i2sDMADesc{
				Config: i2sDMADescCfg(uint16((end-start)*2), i == len(s.descs[buf])-1),
				Buf:    uint32(uintptr(unsafe.Pointer(&s.bufs[buf][start]))),
			}
			if i+1 < len(s.descs[buf]) {
				s.descs[buf][i].Next = uint32(uintptr(unsafe.Pointer(&s.descs[buf][i+1])))
			} else if buf == 0 {
				s.descs[buf][i].Next = uint32(uintptr(unsafe.Pointer(&s.descs[1][0])))
			} else {
				s.descs[buf][i].Next = uint32(uintptr(unsafe.Pointer(&s.descs[0][0])))
			}
		}
	}
	return s, nil
}

// Start flushes both buffers, resets the DMA channel, and starts the circular
// ring.  The I2S TX engine is started once and left running.
func (s *I2SStream) Start() error {
	if s.started {
		return nil
	}
	for i := 0; i < 2; i++ {
		CacheWriteBack(uintptr(unsafe.Pointer(&s.bufs[i][0])), uintptr(len(s.bufs[i]))*2)
		CacheWriteBack(uintptr(unsafe.Pointer(&s.descs[i][0])), uintptr(len(s.descs[i]))*unsafe.Sizeof(s.descs[i][0]))
	}

	esp.AHB_DMA.SetOUT_CONF0_CH0_OUT_RST_CH0(1)
	esp.AHB_DMA.SetOUT_CONF0_CH0_OUT_RST_CH0(0)
	esp.AHB_DMA.SetOUT_INT_CLR_CH0_OUT_DONE_CH_INT_CLR(1)
	esp.AHB_DMA.SetOUT_INT_CLR_CH0_OUT_EOF_CH_INT_CLR(1)
	esp.AHB_DMA.SetOUT_INT_CLR_CH0_OUT_DSCR_ERR_CH_INT_CLR(1)
	esp.AHB_DMA.SetOUT_INT_CLR_CH0_OUT_TOTAL_EOF_CH_INT_CLR(1)
	esp.AHB_DMA.SetOUT_EOF_DES_ADDR_CH0(0)

	descAddr := uint32(uintptr(unsafe.Pointer(&s.descs[0][0])))
	esp.AHB_DMA.SetOUT_LINK_ADDR_CH0(descAddr)
	esp.AHB_DMA.SetOUT_LINK_CH0_OUTLINK_START_CH(1)

	if !s.tx.running {
		i2sTxUpdate(s.tx.hw)
		s.tx.hw.SetTX_CONF_TX_START(1)
		s.tx.running = true
	}
	s.last = -1
	s.started = true
	return nil
}

// WaitFree blocks until one of the two buffers has been fully consumed and
// returns its index.  The other buffer is playing while the caller refills it.
func (s *I2SStream) WaitFree() (int, error) {
	if !s.started {
		return -1, ErrI2STxTimeout
	}
	descSize := uint32(unsafe.Sizeof(i2sDMADesc{}))
	aStart := uint32(uintptr(unsafe.Pointer(&s.descs[0][0])))
	aEnd := aStart + uint32(len(s.descs[0]))*descSize
	bStart := uint32(uintptr(unsafe.Pointer(&s.descs[1][0])))
	bEnd := bStart + uint32(len(s.descs[1]))*descSize

	deadline := i2sTicks() + 40_000*1000*5 // 5 seconds
	for {
		// The circular list sets suc_eof on every descriptor, so this register
		// advances through the ring.  Seeing an address in the other buffer
		// means the buffer returned here has been fully consumed.
		addr := esp.AHB_DMA.GetOUT_EOF_DES_ADDR_CH0()
		idx := -1
		if addr >= aStart && addr < aEnd {
			idx = 0
		} else if addr >= bStart && addr < bEnd {
			idx = 1
		}
		if idx >= 0 && idx != s.last {
			s.last = idx
			return idx, nil
		}
		if i2sTicks() > deadline {
			return -1, ErrI2STxTimeout
		}
	}
}

// Update rewrites the descriptors for one buffer and flushes it for DMA.  The
// buffer is the one returned by WaitFree; the other buffer is playing.
func (s *I2SStream) Update(index int) {
	for i := range s.descs[index] {
		start := i * (maxI2SDescBytes / 2)
		end := start + maxI2SDescBytes/2
		if end > len(s.bufs[index]) {
			end = len(s.bufs[index])
		}
		s.descs[index][i].Config = i2sDMADescCfg(uint16((end-start)*2), i == len(s.descs[index])-1)
	}
	CacheWriteBack(uintptr(unsafe.Pointer(&s.bufs[index][0])), uintptr(len(s.bufs[index]))*2)
	CacheWriteBack(uintptr(unsafe.Pointer(&s.descs[index][0])), uintptr(len(s.descs[index]))*unsafe.Sizeof(s.descs[index][0]))
}

// Buffer returns the sample buffer for one stream slot so the caller can fill
// it before Update.
func (s *I2SStream) Buffer(index int) []int16 {
	return s.bufs[index]
}

// Close stops the circular DMA and TX.  The APB/MCLK clocks are left enabled
// so another stream or I2STx.Close can follow.
func (s *I2SStream) Close() {
	s.tx.hw.SetTX_CONF_TX_START(0)
	i2sTxUpdate(s.tx.hw)
	s.tx.running = false
	esp.AHB_DMA.SetOUT_LINK_CH0_OUTLINK_STOP_CH(1)
	s.started = false
}
