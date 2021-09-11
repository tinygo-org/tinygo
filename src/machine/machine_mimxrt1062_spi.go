// +build mimxrt1062

package machine

// SPI peripheral abstraction layer for the MIMXRT1062

import (
	"device/nxp"
	"errors"
	"unsafe"
)

// SPIConfig is used to store config info for SPI.
type SPIConfig struct {
	Frequency uint32
	SDI       Pin
	SDO       Pin
	SCK       Pin
	CS        Pin
	LSBFirst  bool
	Mode      uint8
}

func (c SPIConfig) getPins() (di, do, ck, cs Pin) {
	if 0 == c.SDI && 0 == c.SDO && 0 == c.SCK && 0 == c.CS {
		// default pins if none specified
		return SPI_SDI_PIN, SPI_SDO_PIN, SPI_SCK_PIN, SPI_CS_PIN
	}
	return c.SDI, c.SDO, c.SCK, c.CS
}

type SPI struct {
	Bus *nxp.LPSPI_Type

	// these hold the input selector ("daisy chain") values that select which pins
	// are connected to the LPSPI device, and should be defined where the SPI
	// instance is declared (e.g., in the board definition). see the godoc
	// comments on type muxSelect for more details.
	muxSDI, muxSDO, muxSCK, muxCS muxSelect

	// these are copied from SPIConfig, during (*SPI).Configure(SPIConfig), and
	// should be considered read-only for internal reference (i.e., modifying them
	// will have no desirable effect).
	sdi, sdo, sck, cs Pin
	frequency         uint32

	// auxiliary state data used internally
	configured bool
}

const (
	statusTxDataRequest    = nxp.LPSPI_SR_TDF // Transmit data flag
	statusRxDataReady      = nxp.LPSPI_SR_RDF // Receive data flag
	statusWordComplete     = nxp.LPSPI_SR_WCF // Word Complete flag
	statusFrameComplete    = nxp.LPSPI_SR_FCF // Frame Complete flag
	statusTransferComplete = nxp.LPSPI_SR_TCF // Transfer Complete flag
	statusTransmitError    = nxp.LPSPI_SR_TEF // Transmit Error flag (FIFO underrun)
	statusReceiveError     = nxp.LPSPI_SR_REF // Receive Error flag (FIFO overrun)
	statusDataMatch        = nxp.LPSPI_SR_DMF // Data Match flag
	statusModuleBusy       = nxp.LPSPI_SR_MBF // Module Busy flag
	statusAll              = nxp.LPSPI_SR_TDF | nxp.LPSPI_SR_RDF |
		nxp.LPSPI_SR_WCF | nxp.LPSPI_SR_FCF | nxp.LPSPI_SR_TCF | nxp.LPSPI_SR_TEF |
		nxp.LPSPI_SR_REF | nxp.LPSPI_SR_DMF | nxp.LPSPI_SR_MBF
)

var (
	errSPINotConfigured = errors.New("SPI interface is not yet configured")
)

// Configure is intended to setup an SPI interface for transmit/receive.
func (spi *SPI) Configure(config SPIConfig) {

	const defaultSpiFreq = 4000000 // 4 MHz

	// init pins
	spi.sdi, spi.sdo, spi.sck, spi.cs = config.getPins()

	// configure the mux and pad control registers
	spi.sdi.Configure(PinConfig{Mode: PinModeSPISDI})
	spi.sdo.Configure(PinConfig{Mode: PinModeSPISDO})
	spi.sck.Configure(PinConfig{Mode: PinModeSPICLK})
	spi.cs.Configure(PinConfig{Mode: PinModeSPICS})

	// configure the mux input selector
	spi.muxSDI.connect()
	spi.muxSDO.connect()
	spi.muxSCK.connect()
	spi.muxCS.connect()

	// software reset of LPSPI state registers
	spi.Bus.CR.SetBits(nxp.LPSPI_CR_RST)
	// also reset FIFOs (not performed by software reset above)
	spi.Bus.CR.SetBits(nxp.LPSPI_CR_RRF | nxp.LPSPI_CR_RTF)
	spi.Bus.CR.Set(0)

	// set controller mode, and input data is sampled on delayed SCK edge
	spi.Bus.CFGR1.Set(nxp.LPSPI_CFGR1_MASTER | nxp.LPSPI_CFGR1_SAMPLE)

	spi.frequency = config.Frequency
	if 0 == spi.frequency {
		spi.frequency = defaultSpiFreq
	}

	// configure LPSPI clock divisor and CS assertion delays
	div := spi.getClockDivisor(config.Frequency)
	ccr := (div << nxp.LPSPI_CCR_SCKDIV_Pos) & nxp.LPSPI_CCR_SCKDIV_Msk
	ccr |= ((div / 2) << nxp.LPSPI_CCR_DBT_Pos) & nxp.LPSPI_CCR_DBT_Msk
	ccr |= ((div / 2) << nxp.LPSPI_CCR_PCSSCK_Pos) & nxp.LPSPI_CCR_PCSSCK_Msk
	spi.Bus.CCR.Set(ccr)

	// 8-bit frame size (words)
	tcr := uint32(7)
	if config.LSBFirst {
		tcr |= nxp.LPSPI_TCR_LSBF
	}
	// set polarity and phase
	switch config.Mode {
	case Mode1:
		tcr |= nxp.LPSPI_TCR_CPHA
	case Mode2:
		tcr |= nxp.LPSPI_TCR_CPOL
	case Mode3:
		tcr |= nxp.LPSPI_TCR_CPOL
		tcr |= nxp.LPSPI_TCR_CPHA
	}
	spi.Bus.TCR.Set(tcr)

	// clear FIFO water marks
	spi.setWatermark(0, 0)

	// enable LPSPI module
	spi.Bus.CR.Set(nxp.LPSPI_CR_MEN)

	spi.configured = true
}

// Transfer writes/reads a single byte using the SPI interface.
func (spi *SPI) Transfer(w byte) (byte, error) {
	if !spi.configured {
		return 0, errSPINotConfigured
	}

	const readTryMax = 10000

	for spi.Bus.SR.HasBits(statusModuleBusy) {
	} // wait for SPI busy bit to clear

	_, txFIFOSize := spi.getFIFOSize()

	spi.flushFIFO(true, true)
	spi.Bus.SR.Set(statusAll) // clear all status flags (W1C)

	// enable LPSPI module
	spi.Bus.CR.Set(nxp.LPSPI_CR_MEN)

	// TODO: unnecessary since we just flushed the FIFO?
	for { // wait for TX FIFO to not be full
		if _, txFIFO := spi.getFIFOCount(); txFIFO < txFIFOSize {
			break
		}
	}

	// write out byte to TX FIFO
	spi.Bus.TDR.Set(uint32(w))

	// try to read from RX FIFO if anything exists
	didRead := false
	data := byte(0)
	for i := 0; !didRead && (i < readTryMax); i++ {
		rxFIFO, _ := spi.getFIFOCount()
		didRead = rxFIFO > 0
		if didRead {
			data = byte(spi.Bus.RDR.Get())
		}
	}

	// if nothing was read, then wait for transfer complete flag to decide when
	// we are finished
	if !didRead {
		for !spi.Bus.SR.HasBits(nxp.LPSPI_SR_TCF) {
		} // wait for all transfers complete flag to set
	}

	return data, nil
}

func (spi *SPI) isHardwareCSPin(pin Pin) bool {
	switch unsafe.Pointer(spi.Bus) {
	case unsafe.Pointer(nxp.LPSPI1):
		return SPI1_CS_PIN == pin
	case unsafe.Pointer(nxp.LPSPI2):
		return SPI2_CS_PIN == pin
	case unsafe.Pointer(nxp.LPSPI3):
		return SPI3_CS_PIN == pin
	}
	return false
}

func (spi *SPI) hasHardwareCSPin() bool {
	return spi.isHardwareCSPin(spi.cs)
}

// getClockDivisor finds the SPI prescalar that minimizes the error between
// requested frequency and possible frequencies available with the LPSPI clock.
// this routine is based on Teensyduino (libraries/SPI/SPI.cpp):
//     `void SPIClass::setClockDivider_noInline(uint32_t clk)`
func (spi *SPI) getClockDivisor(freq uint32) uint32 {
	const clock = 132000000 // LPSPI root clock frequency (PLL2)
	d := uint32(clock)
	if freq > 0 {
		d /= freq
	}
	if d > 0 && clock/d > freq {
		d++
	}
	if d > 257 {
		return 255
	}
	if d > 2 {
		return d - 2
	}
	return 0
}

func (spi *SPI) getFIFOSize() (rx, tx uint32) {
	param := spi.Bus.PARAM.Get()
	return uint32(1) << ((param & nxp.LPSPI_PARAM_RXFIFO_Msk) >> nxp.LPSPI_PARAM_RXFIFO_Pos),
		uint32(1) << ((param & nxp.LPSPI_PARAM_TXFIFO_Msk) >> nxp.LPSPI_PARAM_TXFIFO_Pos)
}

func (spi *SPI) getFIFOCount() (rx, tx uint32) {
	fsr := spi.Bus.FSR.Get()
	return (fsr & nxp.LPSPI_FSR_RXCOUNT_Msk) >> nxp.LPSPI_FSR_RXCOUNT_Pos,
		(fsr & nxp.LPSPI_FSR_TXCOUNT_Msk) >> nxp.LPSPI_FSR_TXCOUNT_Pos
}

func (spi *SPI) flushFIFO(rx, tx bool) {
	var flush uint32
	if rx {
		flush |= nxp.LPSPI_CR_RRF
	}
	if tx {
		flush |= nxp.LPSPI_CR_RTF
	}
	spi.Bus.CR.SetBits(flush)
}

func (spi *SPI) setWatermark(rx, tx uint32) {
	spi.Bus.FCR.Set(((rx << nxp.LPSPI_FCR_RXWATER_Pos) & nxp.LPSPI_FCR_RXWATER_Msk) |
		((tx << nxp.LPSPI_FCR_TXWATER_Pos) & nxp.LPSPI_FCR_TXWATER_Msk))
}
