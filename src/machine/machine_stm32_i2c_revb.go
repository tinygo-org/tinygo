//go:build stm32l5 || stm32f7 || stm32l4 || stm32l0 || stm32wlx

package machine

import (
	"device/stm32"
	"unsafe"
)

//go:linkname ticks runtime.ticks
func ticks() int64

// I2C implementation for 'newer' STM32 MCUs, including the F7, L5 and L4
// series of MCUs.
//
// Currently, only 100KHz mode is supported

const (
	flagBUSY  = stm32.I2C_ISR_BUSY
	flagTCR   = stm32.I2C_ISR_TCR
	flagRXNE  = stm32.I2C_ISR_RXNE
	flagSTOPF = stm32.I2C_ISR_STOPF
	flagAF    = stm32.I2C_ISR_NACKF
	flagTXIS  = stm32.I2C_ISR_TXIS
	flagTXE   = stm32.I2C_ISR_TXE
)

const (
	MAX_NBYTE_SIZE = 255

	// 100ms delay = 100e6ns / 16ns
	// In runtime_stm32_timers.go, tick is fixed at 16ns per tick
	TIMEOUT_TICKS = 100e6 / 16

	I2C_NO_STARTSTOP         = 0x0
	I2C_GENERATE_START_WRITE = 0x80000000 | stm32.I2C_CR2_START
	I2C_GENERATE_START_READ  = 0x80000000 | stm32.I2C_CR2_START | stm32.I2C_CR2_RD_WRN
	I2C_GENERATE_STOP        = 0x80000000 | stm32.I2C_CR2_STOP
)

type I2C struct {
	Bus             *stm32.I2C_Type
	AltFuncSelector uint8
}

// I2CConfig is used to store config info for I2C.
type I2CConfig struct {
	SCL Pin
	SDA Pin
}

func (i2c *I2C) Configure(config I2CConfig) error {
	// disable I2C interface before any configuration changes
	i2c.Bus.CR1.ClearBits(stm32.I2C_CR1_PE)

	// enable clock for I2C
	enableAltFuncClock(unsafe.Pointer(i2c.Bus))

	// init pins
	if config.SCL == 0 && config.SDA == 0 {
		config.SCL = I2C0_SCL_PIN
		config.SDA = I2C0_SDA_PIN
	}
	i2c.configurePins(config)

	// Frequency range
	i2c.Bus.TIMINGR.Set(i2c.getFreqRange())

	// Disable Own Address1 before set the Own Address1 configuration
	i2c.Bus.OAR1.ClearBits(stm32.I2C_OAR1_OA1EN)

	// 7 bit addressing, no self address
	i2c.Bus.OAR1.Set(stm32.I2C_OAR1_OA1EN)

	// Enable the AUTOEND by default, and enable NACK (should be disable only during Slave process
	i2c.Bus.CR2.Set(stm32.I2C_CR2_AUTOEND | stm32.I2C_CR2_NACK)

	// Disable Own Address2 / Dual Addressing
	i2c.Bus.OAR2.Set(0)

	// Disable Generalcall and NoStretch, Enable peripheral
	i2c.Bus.CR1.Set(stm32.I2C_CR1_PE)

	return nil
}

// SetBaudRate sets the communication speed for I2C.
func (i2c *I2C) SetBaudRate(br uint32) error {
	// TODO: implement
	return errI2CNotImplemented
}

func (i2c *I2C) Tx(addr uint16, w, r []byte) error {
	if len(w) > 0 {
		if err := i2c.controllerTransmit(addr, w); nil != err {
			return err
		}
	}

	if len(r) > 0 {
		if err := i2c.controllerReceive(addr, r); nil != err {
			return err
		}
	}

	return nil
}

func (i2c *I2C) configurePins(config I2CConfig) {
	config.SCL.ConfigureAltFunc(PinConfig{Mode: PinModeI2CSCL}, i2c.AltFuncSelector)
	config.SDA.ConfigureAltFunc(PinConfig{Mode: PinModeI2CSDA}, i2c.AltFuncSelector)
}

func (i2c *I2C) controllerTransmit(addr uint16, w []byte) error {
	start := ticks()

	if !i2c.waitOnFlagUntilTimeout(flagBUSY, false, start) {
		return errI2CBusReadyTimeout
	}

	pos := 0
	xferCount := len(w)
	xferSize := uint8(xferCount)
	if xferCount > MAX_NBYTE_SIZE {
		// Large write, indicate reload
		xferSize = MAX_NBYTE_SIZE
		i2c.transferConfig(addr, xferSize, stm32.I2C_CR2_RELOAD, I2C_GENERATE_START_WRITE)
	} else {
		// Small write, auto-end
		i2c.transferConfig(addr, xferSize, stm32.I2C_CR2_AUTOEND, I2C_GENERATE_START_WRITE)
	}

	for xferCount > 0 {
		if !i2c.waitOnTXISFlagUntilTimeout(start) {
			return errI2CWriteTimeout
		}

		i2c.Bus.TXDR.Set(uint32(w[pos]))
		pos++
		xferCount--
		xferSize--

		// If we've written the last byte of this chunk
		if xferCount != 0 && xferSize == 0 {
			// Wait for Transfer Complete Reload to be flagged
			if !i2c.waitOnFlagUntilTimeout(flagTCR, true, start) {
				return errI2CWriteTimeout
			}

			if xferCount > MAX_NBYTE_SIZE {
				// Large write remaining, indicate reload
				xferSize = MAX_NBYTE_SIZE
				i2c.transferConfig(addr, xferSize, stm32.I2C_CR2_RELOAD, I2C_NO_STARTSTOP)
			} else {
				// Small write, auto-end
				xferSize = uint8(xferCount)
				i2c.transferConfig(addr, xferSize, stm32.I2C_CR2_AUTOEND, I2C_NO_STARTSTOP)
			}
		}
	}

	if !i2c.waitOnStopFlagUntilTimeout(start) {
		return errI2CWriteTimeout
	}

	i2c.clearFlag(stm32.I2C_ISR_STOPF)

	i2c.resetCR2()

	return nil
}

func (i2c *I2C) controllerReceive(addr uint16, r []byte) error {
	start := ticks()

	if !i2c.waitOnFlagUntilTimeout(flagBUSY, false, start) {
		return errI2CBusReadyTimeout
	}

	pos := 0
	xferCount := len(r)
	xferSize := uint8(xferCount)
	if xferCount > MAX_NBYTE_SIZE {
		// Large read, indicate reload
		xferSize = MAX_NBYTE_SIZE
		i2c.transferConfig(addr, xferSize, stm32.I2C_CR2_RELOAD, I2C_GENERATE_START_READ)
	} else {
		// Small read, auto-end
		i2c.transferConfig(addr, xferSize, stm32.I2C_CR2_AUTOEND, I2C_GENERATE_START_READ)
	}

	for xferCount > 0 {
		if !i2c.waitOnRXNEFlagUntilTimeout(start) {
			return errI2CWriteTimeout
		}

		r[pos] = uint8(i2c.Bus.RXDR.Get())
		pos++
		xferCount--
		xferSize--

		// If we've read the last byte of this chunk
		if xferCount != 0 && xferSize == 0 {
			// Wait for Transfer Complete Reload to be flagged
			if !i2c.waitOnFlagUntilTimeout(flagTCR, true, start) {
				return errI2CWriteTimeout
			}

			if xferCount > MAX_NBYTE_SIZE {
				// Large read remaining, indicate reload
				xferSize = MAX_NBYTE_SIZE
				i2c.transferConfig(addr, xferSize, stm32.I2C_CR2_RELOAD, I2C_NO_STARTSTOP)
			} else {
				// Small read, auto-end
				xferSize = uint8(xferCount)
				i2c.transferConfig(addr, xferSize, stm32.I2C_CR2_AUTOEND, I2C_NO_STARTSTOP)
			}
		}
	}

	if !i2c.waitOnStopFlagUntilTimeout(start) {
		return errI2CWriteTimeout
	}

	i2c.clearFlag(stm32.I2C_ISR_STOPF)

	i2c.resetCR2()

	return nil
}

func (i2c *I2C) waitOnFlagUntilTimeout(flag uint32, set bool, startTicks int64) bool {
	for i2c.hasFlag(flag) != set {
		if (ticks() - startTicks) > TIMEOUT_TICKS {
			return false
		}
	}
	return true
}

func (i2c *I2C) waitOnRXNEFlagUntilTimeout(startTicks int64) bool {
	for !i2c.hasFlag(flagRXNE) {
		if i2c.isAcknowledgeFailed(startTicks) {
			return false
		}

		if i2c.hasFlag(flagSTOPF) {
			i2c.clearFlag(flagSTOPF)
			i2c.resetCR2()
			return false
		}

		if (ticks() - startTicks) > TIMEOUT_TICKS {
			return false
		}
	}

	return true
}

func (i2c *I2C) waitOnTXISFlagUntilTimeout(startTicks int64) bool {
	for !i2c.hasFlag(flagTXIS) {
		if i2c.isAcknowledgeFailed(startTicks) {
			return false
		}

		if (ticks() - startTicks) > TIMEOUT_TICKS {
			return false
		}
	}

	return true
}

func (i2c *I2C) waitOnStopFlagUntilTimeout(startTicks int64) bool {
	for !i2c.hasFlag(flagSTOPF) {
		if i2c.isAcknowledgeFailed(startTicks) {
			return false
		}

		if (ticks() - startTicks) > TIMEOUT_TICKS {
			return false
		}
	}

	return true
}

func (i2c *I2C) isAcknowledgeFailed(startTicks int64) bool {
	if i2c.hasFlag(flagAF) {
		// Wait until STOP Flag is reset
		// AutoEnd should be initiate after AF
		for !i2c.hasFlag(flagSTOPF) {
			if (ticks() - startTicks) > TIMEOUT_TICKS {
				return true
			}
		}

		i2c.clearFlag(flagAF)
		i2c.clearFlag(flagSTOPF)
		i2c.flushTXDR()
		i2c.resetCR2()

		return true
	}

	return false
}

func (i2c *I2C) flushTXDR() {
	// If a pending TXIS flag is set, write a dummy data in TXDR to clear it
	if i2c.hasFlag(flagTXIS) {
		i2c.Bus.TXDR.Set(0)
	}

	// Flush TX register if not empty
	if !i2c.hasFlag(flagTXE) {
		i2c.clearFlag(flagTXE)
	}
}

func (i2c *I2C) resetCR2() {
	i2c.Bus.CR2.ClearBits(stm32.I2C_CR2_SADD_Msk |
		stm32.I2C_CR2_HEAD10R_Msk |
		stm32.I2C_CR2_NBYTES_Msk |
		stm32.I2C_CR2_RELOAD_Msk |
		stm32.I2C_CR2_RD_WRN_Msk)
}

func (i2c *I2C) transferConfig(addr uint16, size uint8, mode uint32, request uint32) {
	mask := uint32(stm32.I2C_CR2_SADD_Msk |
		stm32.I2C_CR2_NBYTES_Msk |
		stm32.I2C_CR2_RELOAD_Msk |
		stm32.I2C_CR2_AUTOEND_Msk |
		(stm32.I2C_CR2_RD_WRN & uint32(request>>(31-stm32.I2C_CR2_RD_WRN_Pos))) |
		stm32.I2C_CR2_START_Msk |
		stm32.I2C_CR2_STOP_Msk)

	value := (uint32(addr<<1) & stm32.I2C_CR2_SADD_Msk) |
		((uint32(size) << stm32.I2C_CR2_NBYTES_Pos) & stm32.I2C_CR2_NBYTES_Msk) |
		mode | request

	i2c.Bus.CR2.ReplaceBits(value, mask, 0)
}

func (i2c *I2C) hasFlag(flag uint32) bool {
	return i2c.Bus.ISR.HasBits(flag)
}

func (i2c *I2C) clearFlag(flag uint32) {
	if flag == stm32.I2C_ISR_TXE {
		i2c.Bus.ISR.SetBits(flag)
	} else {
		i2c.Bus.ICR.SetBits(flag)
	}
}
