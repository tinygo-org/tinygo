//go:build stm32f4 || stm32f1

package machine

// I2C implementation for 'older' STM32 MCUs, including the F1 and F4 series
// of MCUs.

import (
	"device/stm32"
	"unsafe"
)

const (
	flagOVR     = 0x00010800
	flagAF      = 0x00010400
	flagARLO    = 0x00010200
	flagBERR    = 0x00010100
	flagTXE     = 0x00010080
	flagRXNE    = 0x00010040
	flagSTOPF   = 0x00010010
	flagADD10   = 0x00010008
	flagBTF     = 0x00010004
	flagADDR    = 0x00010002
	flagSB      = 0x00010001
	flagDUALF   = 0x00100080
	flagGENCALL = 0x00100010
	flagTRA     = 0x00100004
	flagBUSY    = 0x00100002
	flagMSL     = 0x00100001
)

func (i2c *I2C) hasFlag(flag uint32) bool {
	const mask = 0x0000FFFF
	if uint8(flag>>16) == 1 {
		return i2c.Bus.SR1.HasBits(flag & mask)
	} else {
		return i2c.Bus.SR2.HasBits(flag & mask)
	}
}

func (i2c *I2C) clearFlag(flag uint32) {
	const mask = 0x0000FFFF
	i2c.Bus.SR1.Set(^(flag & mask))
}

// clearFlagADDR reads both status registers to clear any pending ADDR flags.
func (i2c *I2C) clearFlagADDR() {
	i2c.Bus.SR1.Get()
	i2c.Bus.SR2.Get()
}

func (i2c *I2C) waitForFlag(flag uint32, set bool) bool {
	const tryMax = 10000
	hasFlag := false
	for i := 0; !hasFlag && i < tryMax; i++ {
		hasFlag = i2c.hasFlag(flag) == set
	}
	return hasFlag
}

func (i2c *I2C) waitForFlagOrError(flag uint32, set bool) bool {
	const tryMax = 10000
	hasFlag := false
	for i := 0; !hasFlag && i < tryMax; i++ {
		if hasFlag = i2c.hasFlag(flag) == set; !hasFlag {
			// check for ACK failure
			if i2c.hasFlag(flagAF) {
				// generate stop condition
				i2c.Bus.CR1.SetBits(stm32.I2C_CR1_STOP)
				// clear pending flags
				i2c.clearFlag(flagAF)
				return false
			} else if i2c.hasFlag(flagSTOPF) {
				// clear stop flag
				i2c.clearFlag(flagSTOPF)
				return false
			}
		}
	}
	return hasFlag
}

type transferOption uint32

const (
	frameFirst        = 0x00000001
	frameFirstAndNext = 0x00000002
	frameNext         = 0x00000004
	frameFirstAndLast = 0x00000008
	frameLastNoStop   = 0x00000010
	frameLast         = 0x00000020
	frameNoOption     = 0xFFFF0000
)

// I2C fast mode (Fm) duty cycle
const (
	DutyCycle2    = 0
	DutyCycle16x9 = 1
)

// I2CConfig is used to store config info for I2C.
type I2CConfig struct {
	Frequency uint32
	SCL       Pin
	SDA       Pin
	DutyCycle uint8
}

// Configure is intended to setup the STM32 I2C interface.
func (i2c *I2C) Configure(config I2CConfig) error {

	// The following is the required sequence in controller mode.
	// 1. Program the peripheral input clock in I2C_CR2 Register in order to
	//    generate correct timings
	// 2. Configure the clock control registers
	// 3. Configure the rise time register
	// 4. Program the I2C_CR1 register to enable the peripheral
	// 5. Set the START bit in the I2C_CR1 register to generate a Start condition

	// disable I2C interface before any configuration changes
	i2c.Bus.CR1.ClearBits(stm32.I2C_CR1_PE)

	// reset I2C bus
	i2c.Bus.CR1.SetBits(stm32.I2C_CR1_SWRST)
	i2c.Bus.CR1.ClearBits(stm32.I2C_CR1_SWRST)

	// enable clock for I2C
	enableAltFuncClock(unsafe.Pointer(i2c.Bus))

	// init pins
	if config.SCL == 0 && config.SDA == 0 {
		config.SCL = I2C0_SCL_PIN
		config.SDA = I2C0_SDA_PIN
	}
	i2c.configurePins(config)

	// default to 100 kHz (Sm, standard mode) if no frequency is set
	if config.Frequency == 0 {
		config.Frequency = 100 * KHz
	}

	// configure I2C input clock
	i2c.Bus.CR2.SetBits(i2c.getFreqRange(config))

	// configure rise time
	i2c.Bus.TRISE.Set(i2c.getRiseTime(config))

	// configure clock control
	i2c.Bus.CCR.Set(i2c.getSpeed(config))

	// disable GeneralCall and NoStretch modes
	i2c.Bus.CR1.ClearBits(stm32.I2C_CR1_ENGC | stm32.I2C_CR1_NOSTRETCH)

	// enable I2C interface
	i2c.Bus.CR1.SetBits(stm32.I2C_CR1_PE)

	return nil
}

func (i2c *I2C) Tx(addr uint16, w, r []byte) error {

	if err := i2c.controllerTransmit(addr, w); nil != err {
		return err
	}

	if len(r) > 0 {
		if err := i2c.controllerReceive(addr, r); nil != err {
			return err
		}
	}

	return nil
}

func (i2c *I2C) controllerTransmit(addr uint16, w []byte) error {

	if !i2c.waitForFlag(flagBUSY, false) {
		return errI2CBusReadyTimeout
	}

	// disable POS
	i2c.Bus.CR1.ClearBits(stm32.I2C_CR1_POS)

	pos := 0
	rem := len(w)

	// send peripheral address
	if err := i2c.controllerRequestWrite(addr, frameNoOption); nil != err {
		return err
	}

	// clear ADDR flag
	i2c.clearFlagADDR()

	for rem > 0 {
		// wait for TXE flag set
		if !i2c.waitForFlagOrError(flagTXE, true) {
			return errI2CAckExpected
		}

		// write data to DR
		i2c.Bus.DR.Set(uint32(w[pos]))
		// update counters
		pos++
		rem--

		if i2c.hasFlag(flagBTF) && rem != 0 {
			// write data to DR
			i2c.Bus.DR.Set(uint32(w[pos]))
			// update counters
			pos++
			rem--
		}

		// wait for transfer finished flag BTF set
		if !i2c.waitForFlagOrError(flagBTF, true) {
			return errI2CWriteTimeout
		}
	}

	// generate stop condition
	i2c.Bus.CR1.SetBits(stm32.I2C_CR1_STOP)

	return nil
}

func (i2c *I2C) controllerRequestWrite(addr uint16, option transferOption) error {

	if frameFirstAndLast == option || frameFirst == option || frameNoOption == option {
		// generate start condition
		i2c.Bus.CR1.SetBits(stm32.I2C_CR1_START)
	} else if false /* (hi2c->PreviousState == I2C_STATE_MASTER_BUSY_RX) */ {
		// generate restart condition
		i2c.Bus.CR1.SetBits(stm32.I2C_CR1_START)
	}

	// ensure start bit is set
	if !i2c.waitForFlag(flagSB, true) {
		return errI2CSignalStartTimeout
	}

	// send peripheral address
	i2c.Bus.DR.Set(uint32(addr) << 1)

	// wait for address ACK from peripheral
	if !i2c.waitForFlagOrError(flagADDR, true) {
		return errI2CSignalStartTimeout
	}

	return nil
}

func (i2c *I2C) controllerReceive(addr uint16, r []byte) error {

	if !i2c.waitForFlag(flagBUSY, false) {
		return errI2CBusReadyTimeout
	}

	// disable POS
	i2c.Bus.CR1.ClearBits(stm32.I2C_CR1_POS)

	pos := 0
	rem := len(r)

	// send peripheral address
	if err := i2c.controllerRequestRead(addr, frameNoOption); nil != err {
		return err
	}

	switch rem {
	case 0:
		// clear ADDR flag
		i2c.clearFlagADDR()
		// generate stop condition
		i2c.Bus.CR1.SetBits(stm32.I2C_CR1_STOP)

	case 1:
		// disable ACK
		i2c.Bus.CR1.ClearBits(stm32.I2C_CR1_ACK)
		// clear ADDR flag
		i2c.clearFlagADDR()
		// generate stop condition
		i2c.Bus.CR1.SetBits(stm32.I2C_CR1_STOP)

	case 2:
		// disable ACK
		i2c.Bus.CR1.ClearBits(stm32.I2C_CR1_ACK)
		// enable POS
		i2c.Bus.CR1.SetBits(stm32.I2C_CR1_POS)
		// clear ADDR flag
		i2c.clearFlagADDR()

	default:
		// enable ACK
		i2c.Bus.CR1.SetBits(stm32.I2C_CR1_ACK)
		// clear ADDR flag
		i2c.clearFlagADDR()
	}

	for rem > 0 {
		switch rem {
		case 1:
			// wait until RXNE flag is set
			if !i2c.waitForFlagOrError(flagRXNE, true) {
				return errI2CReadTimeout
			}

			// read data from DR
			r[pos] = byte(i2c.Bus.DR.Get())

			// update counters
			pos++
			rem--

		case 2:
			// wait until transfer finished flag BTF is set
			if !i2c.waitForFlag(flagBTF, true) {
				return errI2CReadTimeout
			}

			// generate stop condition
			i2c.Bus.CR1.SetBits(stm32.I2C_CR1_STOP)

			// read data from DR
			r[pos] = byte(i2c.Bus.DR.Get())

			// update counters
			pos++
			rem--

			// read data from DR
			r[pos] = byte(i2c.Bus.DR.Get())

			// update counters
			pos++
			rem--

		case 3:
			// wait until transfer finished flag BTF is set
			if !i2c.waitForFlag(flagBTF, true) {
				return errI2CReadTimeout
			}

			// disable ACK
			i2c.Bus.CR1.ClearBits(stm32.I2C_CR1_ACK)

			// read data from DR
			r[pos] = byte(i2c.Bus.DR.Get())

			// update counters
			pos++
			rem--

			// wait until transfer finished flag BTF is set
			if !i2c.waitForFlag(flagBTF, true) {
				return errI2CReadTimeout
			}

			// generate stop condition
			i2c.Bus.CR1.SetBits(stm32.I2C_CR1_STOP)

			// read data from DR
			r[pos] = byte(i2c.Bus.DR.Get())

			// update counters
			pos++
			rem--

			// read data from DR
			r[pos] = byte(i2c.Bus.DR.Get())

			// update counters
			pos++
			rem--

		default:
			// wait until RXNE flag is set
			if !i2c.waitForFlagOrError(flagRXNE, true) {
				return errI2CReadTimeout
			}

			// read data from DR
			r[pos] = byte(i2c.Bus.DR.Get())

			// update counters
			pos++
			rem--

			if i2c.hasFlag(flagBTF) {
				// read data from DR
				r[pos] = byte(i2c.Bus.DR.Get())

				// update counters
				pos++
				rem--
			}
		}
	}

	return nil
}

func (i2c *I2C) controllerRequestRead(addr uint16, option transferOption) error {

	// enable ACK
	i2c.Bus.CR1.SetBits(stm32.I2C_CR1_ACK)

	if frameFirstAndLast == option || frameFirst == option || frameNoOption == option {
		// generate start condition
		i2c.Bus.CR1.SetBits(stm32.I2C_CR1_START)
	} else if false /* (hi2c->PreviousState == I2C_STATE_MASTER_BUSY_TX) */ {
		// generate restart condition
		i2c.Bus.CR1.SetBits(stm32.I2C_CR1_START)
	}

	// ensure start bit is set
	if !i2c.waitForFlag(flagSB, true) {
		return errI2CSignalStartTimeout
	}

	// send peripheral address
	i2c.Bus.DR.Set(uint32(addr)<<1 | 1)

	// wait for address ACK from peripheral
	if !i2c.waitForFlagOrError(flagADDR, true) {
		return errI2CSignalStartTimeout
	}

	return nil
}
