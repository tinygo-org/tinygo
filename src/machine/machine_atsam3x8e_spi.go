//go:build sam && atsam3x8e

package machine

import (
	"device/sam"
	"errors"
	"unsafe"
)

var errSPILSBFirst = errors.New("SPI: LSB first is not supported")

// SPI0 TX is DMAC hardware interface 1, as in SdFat src/SpiDriver/SdSpiDue.cpp.
const (
	spiDMATxInterface = 1
	spiDMAMinLen      = 16
	spiDMAMaxLen      = 4095
)

// SPIConfig is used to store config info for SPI.
type SPIConfig struct {
	Frequency uint32
	SCK       Pin
	SDO       Pin
	SDI       Pin
	LSBFirst  bool
	Mode      uint8
}

// SPI on the SAM3X8E. The pins are fixed, so SCK, SDO and SDI are ignored.
type SPI struct {
	Bus *sam.SPI_Type
}

var SPI0 = &SPI{Bus: sam.SPI0}

// Configure sets up SPI0 as a controller on PA25, PA26 and PA27.
func (spi *SPI) Configure(config SPIConfig) error {
	if config.LSBFirst {
		return errSPILSBFirst
	}
	if config.Frequency == 0 {
		config.Frequency = 4_000_000
	}

	PA25.Configure(PinConfig{Mode: PinPeripheralA})
	PA26.Configure(PinConfig{Mode: PinPeripheralA})
	PA27.Configure(PinConfig{Mode: PinPeripheralA})
	sam.PMC.PMC_PCER0.Set(1 << sam.IRQ_SPI0)
	sam.PMC.PMC_PCER1.Set(1 << (sam.IRQ_DMAC - 32))
	sam.DMAC.EN.Set(sam.DMAC_EN_ENABLE)

	spi.Bus.CR.Set(sam.SPI_CR_SPIDIS)
	spi.Bus.CR.Set(sam.SPI_CR_SWRST)
	spi.Bus.CR.Set(sam.SPI_CR_SWRST)
	spi.Bus.MR.Set(sam.SPI_MR_MSTR | sam.SPI_MR_MODFDIS | 0xe<<sam.SPI_MR_PCS_Pos)

	scbr := (CPUFrequency() + config.Frequency - 1) / config.Frequency
	scbr = max(1, min(scbr, 255))
	csr := scbr<<sam.SPI_CSR_SCBR_Pos | sam.SPI_CSR_BITS_8_BIT<<sam.SPI_CSR_BITS_Pos
	switch config.Mode {
	case Mode0:
		csr |= sam.SPI_CSR_NCPHA
	case Mode2:
		csr |= sam.SPI_CSR_CPOL | sam.SPI_CSR_NCPHA
	case Mode3:
		csr |= sam.SPI_CSR_CPOL
	}
	spi.Bus.CSR[0].Set(csr)

	spi.Bus.CR.Set(sam.SPI_CR_SPIEN)
	return nil
}

// Transfer writes the byte and returns the byte read at the same time.
func (spi *SPI) Transfer(w byte) (byte, error) {
	for !spi.Bus.SR.HasBits(sam.SPI_SR_TDRE) {
	}
	spi.Bus.TDR.Set(uint32(w))
	for !spi.Bus.SR.HasBits(sam.SPI_SR_RDRF) {
	}
	return byte(spi.Bus.RDR.Get()), nil
}

// Tx writes w and reads into r. Either can be nil, otherwise they must be the
// same length.
func (spi *SPI) Tx(w, r []byte) error {
	switch {
	case r == nil:
		if len(w) >= spiDMAMinLen && inSRAM(&w[0]) {
			spi.txDMA(w)
		} else {
			for _, b := range w {
				for !spi.Bus.SR.HasBits(sam.SPI_SR_TDRE) {
				}
				spi.Bus.TDR.Set(uint32(b))
			}
		}
		for !spi.Bus.SR.HasBits(sam.SPI_SR_TXEMPTY) {
		}
		spi.Bus.RDR.Get()
	case w == nil:
		for i := range r {
			r[i], _ = spi.Transfer(0)
		}
	default:
		if len(w) != len(r) {
			return ErrTxInvalidSliceSize
		}
		for i, b := range w {
			r[i], _ = spi.Transfer(b)
		}
	}
	return nil
}

// txDMA writes w to the transmit register with DMAC channel 0.
func (spi *SPI) txDMA(w []byte) {
	const (
		ctrlb = sam.DMAC_CTRLB0_SRC_DSCR | sam.DMAC_CTRLB0_DST_DSCR |
			sam.DMAC_CTRLB0_FC_MEM2PER_DMA_FC<<sam.DMAC_CTRLB0_FC_Pos |
			sam.DMAC_CTRLB0_SRC_INCR_INCREMENTING<<sam.DMAC_CTRLB0_SRC_INCR_Pos |
			sam.DMAC_CTRLB0_DST_INCR_FIXED<<sam.DMAC_CTRLB0_DST_INCR_Pos
		cfg = spiDMATxInterface<<sam.DMAC_CFG0_DST_PER_Pos | sam.DMAC_CFG0_DST_H2SEL | sam.DMAC_CFG0_SOD
	)
	dma := sam.DMAC
	for len(w) > 0 {
		n := min(len(w), spiDMAMaxLen)
		dma.CHDR.Set(sam.DMAC_CHDR_DIS0)
		dma.SADDR0.Set(uint32(uintptr(unsafe.Pointer(&w[0]))))
		dma.DADDR0.Set(uint32(uintptr(unsafe.Pointer(&spi.Bus.TDR.Reg))))
		dma.DSCR0.Set(0)
		dma.CTRLA0.Set(uint32(n))
		dma.CTRLB0.Set(ctrlb)
		dma.CFG0.Set(cfg)
		dma.CHER.Set(sam.DMAC_CHER_ENA0)
		for dma.CHSR.HasBits(sam.DMAC_CHSR_ENA0) {
		}
		w = w[n:]
	}
}

func inSRAM(p *byte) bool {
	addr := uintptr(unsafe.Pointer(p))
	return addr >= 0x20000000 && addr < 0x20100000
}
