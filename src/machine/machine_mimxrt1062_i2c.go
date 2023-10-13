//go:build mimxrt1062

package machine

// I2C peripheral abstraction layer for the MIMXRT1062

import (
	"device/nxp"
	"errors"
)

var (
	errI2CWriteTimeout       = errors.New("I2C timeout during write")
	errI2CReadTimeout        = errors.New("I2C timeout during read")
	errI2CBusReadyTimeout    = errors.New("I2C timeout on bus ready")
	errI2CSignalStartTimeout = errors.New("I2C timeout on signal start")
	errI2CSignalReadTimeout  = errors.New("I2C timeout on signal read")
	errI2CSignalStopTimeout  = errors.New("I2C timeout on signal stop")
	errI2CAckExpected        = errors.New("I2C error: expected ACK not NACK")
	errI2CBusError           = errors.New("I2C bus error")
	errI2CNotConfigured      = errors.New("I2C interface is not yet configured")
)

// I2CConfig is used to store config info for I2C.
type I2CConfig struct {
	Frequency uint32
	SDA       Pin
	SCL       Pin
}

type I2C struct {
	Bus *nxp.LPI2C_Type

	// these pins are initialized by each global I2C variable declared in the
	// board_teensy4x.go file according to the board manufacturer's default pin
	// mapping. they can be overridden with the I2CConfig argument given to
	// (*I2C) Configure(I2CConfig).
	sda, scl Pin

	// these hold the input selector ("daisy chain") values that select which pins
	// are connected to the LPI2C device, and should be defined where the I2C
	// instance is declared (e.g., in the board definition). see the godoc
	// comments on type muxSelect for more details.
	muxSDA, muxSCL muxSelect
}

type i2cDirection bool

const (
	directionWrite i2cDirection = false
	directionRead  i2cDirection = true
)

func (dir i2cDirection) shift(addr uint16) uint32 {
	if addr <<= 1; dir == directionRead {
		addr |= 1
	}
	return uint32(addr) & 0xFF
}

// I2C enumerated types
type (
	resultFlag   uint32
	statusFlag   uint32
	transferFlag uint32
	commandFlag  uint32
	stateFlag    uint32
)

const (
	// general purpose results
	resultSuccess         resultFlag = 0x0 // success
	resultFail            resultFlag = 0x1 // fail
	resultReadOnly        resultFlag = 0x2 // read only failure
	resultOutOfRange      resultFlag = 0x3 // out of range access
	resultInvalidArgument resultFlag = 0x4 // invalid argument check
	// I2C-specific results
	resultBusy                 resultFlag = 0x0384 + 0x0 // the controller is already performing a transfer
	resultIdle                 resultFlag = 0x0384 + 0x1 // the peripheral driver is idle
	resultNak                  resultFlag = 0x0384 + 0x2 // the peripheral device sent a NAK in response to a byte
	resultFifoError            resultFlag = 0x0384 + 0x3 // FIFO under run or overrun
	resultBitError             resultFlag = 0x0384 + 0x4 // transferred bit was not seen on the bus
	resultArbitrationLost      resultFlag = 0x0384 + 0x5 // arbitration lost error
	resultPinLowTimeout        resultFlag = 0x0384 + 0x6 // SCL or SDA were held low longer than the timeout
	resultNoTransferInProgress resultFlag = 0x0384 + 0x7 // attempt to abort a transfer when one is not in progress
	resultDmaRequestFail       resultFlag = 0x0384 + 0x8 // DMA request failed
	resultTimeout              resultFlag = 0x0384 + 0x9 // timeout polling status flags
)

const (
	statusTxReady         statusFlag = nxp.LPI2C_MSR_TDF  // transmit data flag
	statusRxReady         statusFlag = nxp.LPI2C_MSR_RDF  // receive data flag
	statusEndOfPacket     statusFlag = nxp.LPI2C_MSR_EPF  // end Packet flag
	statusStopDetect      statusFlag = nxp.LPI2C_MSR_SDF  // stop detect flag
	statusNackDetect      statusFlag = nxp.LPI2C_MSR_NDF  // NACK detect flag
	statusArbitrationLost statusFlag = nxp.LPI2C_MSR_ALF  // arbitration lost flag
	statusFifoErr         statusFlag = nxp.LPI2C_MSR_FEF  // FIFO error flag
	statusPinLowTimeout   statusFlag = nxp.LPI2C_MSR_PLTF // pin low timeout flag
	statusI2CDataMatch    statusFlag = nxp.LPI2C_MSR_DMF  // data match flag
	statusBusy            statusFlag = nxp.LPI2C_MSR_MBF  // busy flag
	statusBusBusy         statusFlag = nxp.LPI2C_MSR_BBF  // bus busy flag

	// all flags which are cleared by the driver upon starting a transfer
	statusClear statusFlag = statusEndOfPacket | statusStopDetect | statusNackDetect |
		statusArbitrationLost | statusFifoErr | statusPinLowTimeout | statusI2CDataMatch

	// IRQ sources enabled by the non-blocking transactional API
	statusIrq statusFlag = statusArbitrationLost | statusTxReady | statusRxReady |
		statusStopDetect | statusNackDetect | statusPinLowTimeout | statusFifoErr

	// errors to check for
	statusError statusFlag = statusNackDetect | statusArbitrationLost | statusFifoErr |
		statusPinLowTimeout
)

// LPI2C transfer modes
const (
	transferDefault       transferFlag = 0x0 // transfer starts with a start signal, stops with a stop signal
	transferNoStart       transferFlag = 0x1 // don't send a start condition, address, and sub address
	transferRepeatedStart transferFlag = 0x2 // send a repeated start condition
	transferNoStop        transferFlag = 0x4 // don't send a stop condition
)

// LPI2C FIFO commands
const (
	commandTxData commandFlag = (0x0 << nxp.LPI2C_MTDR_CMD_Pos) & nxp.LPI2C_MTDR_CMD_Msk // transmit
	commandRxData commandFlag = (0x1 << nxp.LPI2C_MTDR_CMD_Pos) & nxp.LPI2C_MTDR_CMD_Msk // receive
	commandStop   commandFlag = (0x2 << nxp.LPI2C_MTDR_CMD_Pos) & nxp.LPI2C_MTDR_CMD_Msk // generate STOP condition
	commandStart  commandFlag = (0x4 << nxp.LPI2C_MTDR_CMD_Pos) & nxp.LPI2C_MTDR_CMD_Msk // generate (REPEATED)START and transmit
)

// LPI2C transactional states
const (
	stateIdle              stateFlag = 0x0
	stateSendCommand       stateFlag = 0x1
	stateIssueReadCommand  stateFlag = 0x2
	stateTransferData      stateFlag = 0x3
	stateStop              stateFlag = 0x4
	stateWaitForCompletion stateFlag = 0x5
)

func (i2c *I2C) setPins(c I2CConfig) (sda, scl Pin) {
	// if both given pins are defined, or either receiver pin is undefined.
	if 0 != c.SDA && 0 != c.SCL || 0 == i2c.sda || 0 == i2c.scl {
		// override the receiver's pins.
		i2c.sda, i2c.scl = c.SDA, c.SCL
	}
	// return the selected pins.
	return i2c.sda, i2c.scl
}

// Configure is intended to setup an I2C interface for transmit/receive.
func (i2c *I2C) Configure(config I2CConfig) {
	// init pins
	sda, scl := i2c.setPins(config)

	// configure the mux and pad control registers
	sda.Configure(PinConfig{Mode: PinModeI2CSDA})
	scl.Configure(PinConfig{Mode: PinModeI2CSCL})

	// configure the mux input selector
	i2c.muxSDA.connect()
	i2c.muxSCL.connect()

	freq := config.Frequency
	if 0 == freq {
		freq = 100 * KHz
	}

	// reset clock and registers, and enable LPI2C module interface
	i2c.reset(freq)
}

// SetBaudRate sets the communication speed for I2C.
func (i2c I2C) SetBaudRate(br uint32) error {
	// TODO: implement
	return errI2CNotImplemented
}

func (i2c I2C) Tx(addr uint16, w, r []byte) error {
	// perform transmit transfer
	if nil != w {
		// generate start condition on bus
		if result := i2c.start(addr, directionWrite); resultSuccess != result {
			return errI2CSignalStartTimeout
		}
		// ensure TX FIFO is empty
		if result := i2c.waitForTxEmpty(); resultSuccess != result {
			return errI2CBusReadyTimeout
		}
		// check if communication was successful
		if status := statusFlag(i2c.Bus.MSR.Get()); 0 != (status & statusNackDetect) {
			return errI2CAckExpected
		}
		// send transmit data
		if result := i2c.controllerTransmit(w); resultSuccess != result {
			return errI2CWriteTimeout
		}
	}

	// perform receive transfer
	if nil != r {
		// generate (repeated-)start condition on bus
		if result := i2c.start(addr, directionRead); resultSuccess != result {
			return errI2CSignalStartTimeout
		}
		// read received data
		if result := i2c.controllerReceive(r); resultSuccess != result {
			return errI2CReadTimeout
		}
	}

	// generate stop condition on bus
	if result := i2c.stop(); resultSuccess != result {
		return errI2CSignalStopTimeout
	}

	return nil
}

// WriteRegister transmits first the register and then the data to the
// peripheral device.
//
// Many I2C-compatible devices are organized in terms of registers. This method
// is a shortcut to easily write to such registers. Also, it only works for
// devices with 7-bit addresses, which is the vast majority.
func (i2c I2C) WriteRegister(address uint8, register uint8, data []byte) error {
	option := transferOption{
		flags:          transferDefault,  // transfer options bit mask (0 = normal transfer)
		peripheral:     uint16(address),  // 7-bit peripheral address
		direction:      directionWrite,   // directionRead or directionWrite
		subaddress:     uint16(register), // peripheral sub-address (transferred MSB first)
		subaddressSize: 1,                // byte length of sub-address (maximum = 4 bytes)
	}
	if result := i2c.controllerTransferPoll(option, data); resultSuccess != result {
		return errI2CWriteTimeout
	}
	return nil
}

// ReadRegister transmits the register, restarts the connection as a read
// operation, and reads the response.
//
// Many I2C-compatible devices are organized in terms of registers. This method
// is a shortcut to easily read such registers. Also, it only works for devices
// with 7-bit addresses, which is the vast majority.
func (i2c I2C) ReadRegister(address uint8, register uint8, data []byte) error {
	option := transferOption{
		flags:          transferDefault,  // transfer options bit mask (0 = normal transfer)
		peripheral:     uint16(address),  // 7-bit peripheral address
		direction:      directionRead,    // directionRead or directionWrite
		subaddress:     uint16(register), // peripheral sub-address (transferred MSB first)
		subaddressSize: 1,                // byte length of sub-address (maximum = 4 bytes)
	}
	if result := i2c.controllerTransferPoll(option, data); resultSuccess != result {
		return errI2CWriteTimeout
	}
	return nil
}

func (i2c *I2C) reset(freq uint32) {
	// disable interface
	i2c.Bus.MCR.ClearBits(nxp.LPI2C_MCR_MEN)

	// software reset all interface registers
	i2c.Bus.MCR.Set(nxp.LPI2C_MCR_RST)

	// RST remains set until manually cleared!
	i2c.Bus.MCR.ClearBits(nxp.LPI2C_MCR_RST)

	// disable host request
	i2c.Bus.MCFGR0.Set(0)

	// enable ACK, use I2C 2-pin open drain mode
	i2c.Bus.MCFGR1.Set(0)

	// set FIFO watermarks (RX=1, TX=1)
	mfcr := (uint32(0x1) << nxp.LPI2C_MFCR_RXWATER_Pos) & nxp.LPI2C_MFCR_RXWATER_Msk
	mfcr |= (uint32(0x1) << nxp.LPI2C_MFCR_TXWATER_Pos) & nxp.LPI2C_MFCR_TXWATER_Msk
	i2c.Bus.MFCR.Set(mfcr)

	// configure clock using receiver frequency
	i2c.setFrequency(freq)

	// clear reset, and enable the interface
	i2c.Bus.MCR.Set(nxp.LPI2C_MCR_MEN)

	// wait for the I2C bus to idle
	for i2c.Bus.MSR.Get()&nxp.LPI2C_MSR_BBF != 0 {
	}
}

func (i2c *I2C) setFrequency(freq uint32) {
	var (
		bestPre   uint32 = 0
		bestClkHi uint32 = 0
		bestError uint32 = 0xFFFFFFFF
	)

	// disable interface
	wasEnabled := i2c.Bus.MCR.HasBits(nxp.LPI2C_MCR_MEN)
	i2c.Bus.MCR.ClearBits(nxp.LPI2C_MCR_MEN)

	// baud rate = (24MHz/(2^pre))/(CLKLO+1 + CLKHI+1 + FLOOR((2+FILTSCL)/(2^pre)))
	// assume: CLKLO=2*CLKHI, SETHOLD=CLKHI, DATAVD=CLKHI/2
	for pre := uint32(1); pre <= 128; pre *= 2 {
		if bestError == 0 {
			break
		}
		for clkHi := uint32(1); clkHi < 32; clkHi++ {
			var absError, rate uint32
			if clkHi == 1 {
				rate = (24 * MHz / pre) / (1 + 3 + 2 + 2/pre)
			} else {
				rate = (24 * MHz / pre) / (3*clkHi + 2 + 2/pre)
			}
			if freq > rate {
				absError = freq - rate
			} else {
				absError = rate - freq
			}
			if absError < bestError {
				bestPre = pre
				bestClkHi = clkHi
				bestError = absError
				// if the error is 0, then we can stop searching because we won't find a
				// better match
				if absError == 0 {
					break
				}
			}
		}
	}

	var (
		clklo   = func(n uint32) uint32 { return (n << nxp.LPI2C_MCCR0_CLKLO_Pos) & nxp.LPI2C_MCCR0_CLKLO_Msk }
		clkhi   = func(n uint32) uint32 { return (n << nxp.LPI2C_MCCR0_CLKHI_Pos) & nxp.LPI2C_MCCR0_CLKHI_Msk }
		datavd  = func(n uint32) uint32 { return (n << nxp.LPI2C_MCCR0_DATAVD_Pos) & nxp.LPI2C_MCCR0_DATAVD_Msk }
		sethold = func(n uint32) uint32 { return (n << nxp.LPI2C_MCCR0_SETHOLD_Pos) & nxp.LPI2C_MCCR0_SETHOLD_Msk }
	)
	// StandardMode, FastMode, FastModePlus, and UltraFastMode
	mccr0 := clkhi(bestClkHi)
	if bestClkHi < 2 {
		mccr0 |= (clklo(3) | sethold(2) | datavd(1))
	} else {
		mccr0 |= clklo(2*bestClkHi) | sethold(bestClkHi) | datavd(bestClkHi/2)
	}
	i2c.Bus.MCCR0.Set(mccr0)
	i2c.Bus.MCCR1.Set(i2c.Bus.MCCR0.Get())

	for i := uint32(0); i < 8; i++ {
		if bestPre == (1 << i) {
			bestPre = i
			break
		}
	}
	preMask := (bestPre << nxp.LPI2C_MCFGR1_PRESCALE_Pos) & nxp.LPI2C_MCFGR1_PRESCALE_Msk
	i2c.Bus.MCFGR1.Set((i2c.Bus.MCFGR1.Get() & ^uint32(nxp.LPI2C_MCFGR1_PRESCALE_Msk)) | preMask)

	var (
		filtsda = func(n uint32) uint32 { return (n << nxp.LPI2C_MCFGR2_FILTSDA_Pos) & nxp.LPI2C_MCFGR2_FILTSDA_Msk }
		filtscl = func(n uint32) uint32 { return (n << nxp.LPI2C_MCFGR2_FILTSCL_Pos) & nxp.LPI2C_MCFGR2_FILTSCL_Msk }
		busidle = func(n uint32) uint32 { return (n << nxp.LPI2C_MCFGR2_BUSIDLE_Pos) & nxp.LPI2C_MCFGR2_BUSIDLE_Msk }
		pinlow  = func(n uint32) uint32 { return (n << nxp.LPI2C_MCFGR3_PINLOW_Pos) & nxp.LPI2C_MCFGR3_PINLOW_Msk }

		mcfgr2, mcfgr3 uint32
	)
	const i2cClockStretchTimeout = 15000 // microseconds
	if freq >= 5*MHz {
		// I2C UltraFastMode 5 MHz
		mcfgr2 = 0 // disable glitch filters and timeout for UltraFastMode
		mcfgr3 = 0 //
	} else if freq >= 1*MHz {
		// I2C FastModePlus 1 MHz
		mcfgr2 = filtsda(1) | filtscl(1) | busidle(2400) // 100us timeout
		mcfgr3 = pinlow(i2cClockStretchTimeout*24/256 + 1)
	} else if freq >= 400*KHz {
		// I2C FastMode 400 kHz
		mcfgr2 = filtsda(2) | filtscl(2) | busidle(3600) // 150us timeout
		mcfgr3 = pinlow(i2cClockStretchTimeout*24/256 + 1)
	} else {
		// I2C StandardMode 100 kHz
		mcfgr2 = filtsda(5) | filtscl(5) | busidle(3000) // 250us timeout
		mcfgr3 = pinlow(i2cClockStretchTimeout*12/256 + 1)
	}
	i2c.Bus.MCFGR2.Set(mcfgr2)
	i2c.Bus.MCFGR3.Set(mcfgr3)

	// restore controller mode if it was enabled when called
	if wasEnabled {
		i2c.Bus.MCR.SetBits(nxp.LPI2C_MCR_MEN)
	}
}

// checkStatus converts the status register to a resultFlag for return, and
// clears any errors if present.
func (i2c *I2C) checkStatus(status statusFlag) resultFlag {
	result := resultSuccess
	// check for error. these errors cause a stop to be sent automatically.
	// we must clear the errors before a new transfer can start.
	if status &= statusError; 0 != status {
		// select the correct error code ordered by severity, bus issues first.
		if 0 != (status & statusPinLowTimeout) {
			result = resultPinLowTimeout
		} else if 0 != (status & statusArbitrationLost) {
			result = resultArbitrationLost
		} else if 0 != (status & statusNackDetect) {
			result = resultNak
		} else if 0 != (status & statusFifoErr) {
			result = resultFifoError
		}
		// clear the flags
		i2c.Bus.MSR.Set(uint32(status))
		// reset fifos. these flags clear automatically.
		i2c.Bus.MCR.SetBits(nxp.LPI2C_MCR_RRF | nxp.LPI2C_MCR_RTF)
	}
	return result
}

func (i2c *I2C) getFIFOSize() (rx, tx uint32) { return 4, 4 }
func (i2c *I2C) getFIFOCount() (rx, tx uint32) {
	mfsr := i2c.Bus.MFSR.Get()
	return (mfsr & nxp.LPI2C_MFSR_RXCOUNT_Msk) >> nxp.LPI2C_MFSR_RXCOUNT_Pos,
		(mfsr & nxp.LPI2C_MFSR_TXCOUNT_Msk) >> nxp.LPI2C_MFSR_TXCOUNT_Pos
}

func (i2c *I2C) waitForTxReady() resultFlag {
	result := resultSuccess
	_, txSize := i2c.getFIFOSize()
	for {
		_, txCount := i2c.getFIFOCount()
		status := statusFlag(i2c.Bus.MSR.Get())
		if result = i2c.checkStatus(status); resultSuccess != result {
			break
		}
		if txSize-txCount > 0 {
			break
		}
	}
	return result
}

func (i2c *I2C) waitForTxEmpty() resultFlag {
	result := resultSuccess
	for {
		_, txCount := i2c.getFIFOCount()
		status := statusFlag(i2c.Bus.MSR.Get())
		if result = i2c.checkStatus(status); resultSuccess != result {
			break
		}
		if 0 == txCount {
			break
		}
	}
	return result
}

// isBusBusy checks if the I2C bus is busy, returning true if it is busy and we
// are not the ones driving it, otherwise false.
func (i2c *I2C) isBusBusy() bool {
	status := statusFlag(i2c.Bus.MSR.Get())
	return (0 != (status & statusBusBusy)) && (0 == (status & statusBusy))
}

// start sends a START signal and peripheral address on the I2C bus.
//
// This function is used to initiate a new controller mode transfer. First, the
// bus state is checked to ensure that another controller is not occupying the
// bus. Then a START signal is transmitted, followed by the 7-bit peripheral
// address. Note that this function does not actually wait until the START and
// address are successfully sent on the bus before returning.
func (i2c *I2C) start(address uint16, dir i2cDirection) resultFlag {
	// return an error if the bus is already in use by another controller
	if i2c.isBusBusy() {
		return resultBusy
	}
	// clear all flags
	i2c.Bus.MSR.Set(uint32(statusClear))
	// turn off auto-stop
	i2c.Bus.MCFGR1.ClearBits(nxp.LPI2C_MCFGR1_AUTOSTOP)
	// wait until there is room in the FIFO
	if result := i2c.waitForTxReady(); resultSuccess != result {
		return result
	}

	// issue start command
	i2c.Bus.MTDR.Set(uint32(commandStart) | dir.shift(address))
	return resultSuccess
}

// stop sends a STOP signal on the I2C bus.
//
// This function does not return until the STOP signal is seen on the bus, or
// an error occurs.
func (i2c *I2C) stop() resultFlag {
	const tryMax = 0 // keep waiting forever
	// wait until there is room in the FIFO
	result := i2c.waitForTxReady()
	if resultSuccess != result {
		return result
	}
	// send the STOP signal
	i2c.Bus.MTDR.Set(uint32(commandStop))
	// wait for the stop detected flag to set, indicating the transfer has
	// completed on the bus. also check for errors while waiting.
	try := 0
	for resultSuccess == result && (0 == tryMax || try < tryMax) {
		status := statusFlag(i2c.Bus.MSR.Get())
		result = i2c.checkStatus(status)
		if (0 != (status & statusStopDetect)) && (0 != (status & statusTxReady)) {
			i2c.Bus.MSR.Set(uint32(statusStopDetect))
			break
		}
		try++
	}
	if 0 != tryMax && try >= tryMax {
		return resultTimeout
	}
	return result
}

// controllerReceive performs a polling receive transfer on the I2C bus.
func (i2c *I2C) controllerReceive(rxBuffer []byte) resultFlag {
	const tryMax = 0 // keep trying forever
	rxSize := len(rxBuffer)
	if rxSize == 0 {
		return resultSuccess
	}
	// wait until there is room in the FIFO
	result := i2c.waitForTxReady()
	if resultSuccess != result {
		return result
	}
	sizeMask := (uint32(rxSize-1) << nxp.LPI2C_MTDR_DATA_Pos) & nxp.LPI2C_MTDR_DATA_Msk
	i2c.Bus.MTDR.Set(uint32(commandRxData) | sizeMask)

	// receive data
	for rxSize > 0 {
		// read LPI2C receive FIFO register. the register includes a flag to
		// indicate whether the FIFO is empty, so we can both get the data and check
		// if we need to keep reading using a single register read.
		var data uint32
		try := 0
		for 0 == tryMax || try < tryMax {
			// check for errors on the bus
			status := statusFlag(i2c.Bus.MSR.Get())
			result = i2c.checkStatus(status)
			if resultSuccess != result {
				return result
			}
			// read received data, break if FIFO was non-empty
			data = i2c.Bus.MRDR.Get()
			if 0 == (data & nxp.LPI2C_MRDR_RXEMPTY_Msk) {
				break
			}
			try++
		}
		// ensure we didn't timeout waiting for data
		if 0 != tryMax && try >= tryMax {
			return resultTimeout
		}
		// copy data to RX buffer
		rxBuffer[len(rxBuffer)-rxSize] = byte(data & nxp.LPI2C_MRDR_DATA_Msk)
		rxSize--
	}
	return result
}

// controllerReceive performs a polling transmit transfer on the I2C bus.
func (i2c *I2C) controllerTransmit(txBuffer []byte) resultFlag {
	txSize := len(txBuffer)
	for txSize > 0 {
		// wait until there is room in the FIFO
		result := i2c.waitForTxReady()
		if resultSuccess != result {
			return result
		}
		// write byte into LPI2C data register
		i2c.Bus.MTDR.Set(uint32(txBuffer[len(txBuffer)-txSize] & nxp.LPI2C_MTDR_DATA_Msk))
		txSize--
	}
	return resultSuccess
}

type transferOption struct {
	flags          transferFlag // transfer options bit mask (0 = normal transfer)
	peripheral     uint16       // 7-bit peripheral address
	direction      i2cDirection // directionRead or directionWrite
	subaddress     uint16       // peripheral sub-address (transferred MSB first)
	subaddressSize uint16       // byte length of sub-address (maximum = 4 bytes)
}

func (i2c *I2C) controllerTransferPoll(option transferOption, data []byte) resultFlag {
	// return an error if the bus is already in use by another controller
	if i2c.isBusBusy() {
		return resultBusy
	}
	// clear all flags
	i2c.Bus.MSR.Set(uint32(statusClear))
	// turn off auto-stop
	i2c.Bus.MCFGR1.ClearBits(nxp.LPI2C_MCFGR1_AUTOSTOP)

	cmd := make([]uint16, 0, 7)
	size := len(data)

	direction := option.direction
	if option.subaddressSize > 0 {
		direction = directionWrite
	}
	// peripheral address
	if 0 == (option.flags & transferNoStart) {
		addr := direction.shift(option.peripheral)
		cmd = append(cmd, uint16(uint32(commandStart)|addr))
	}
	// sub-address (MSB-first)
	rem := option.subaddressSize
	for rem > 0 {
		rem--
		cmd = append(cmd, (option.subaddress>>(8*rem))&0xFF)
	}
	// need to send repeated start if switching directions to read
	if (0 != size) && (directionRead == option.direction) {
		if directionWrite == direction {
			addr := directionRead.shift(option.peripheral)
			cmd = append(cmd, uint16(uint32(commandStart)|addr))
		}
	}
	// send command buffer
	result := resultSuccess
	for _, c := range cmd {
		// wait until there is room in the FIFO
		if result = i2c.waitForTxReady(); resultSuccess != result {
			return result
		}
		// write byte into LPI2C controller data register
		i2c.Bus.MTDR.Set(uint32(c))
	}
	// send data
	if option.direction == directionWrite && size > 0 {
		result = i2c.controllerTransmit(data)
	}
	// receive data
	if option.direction == directionRead && size > 0 {
		result = i2c.controllerReceive(data)
	}
	if resultSuccess != result {
		return result
	}
	if 0 == (option.flags & transferNoStop) {
		result = i2c.stop()
	}
	return result
}
