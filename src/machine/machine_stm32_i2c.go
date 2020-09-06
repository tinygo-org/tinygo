// +build stm32,!stm32f103xx

package machine

// Peripheral abstraction layer for I2C on the stm32 family

import (
	"device/stm32"
	"unsafe"
)

// I2C fast mode (Fm) duty cycle
const (
	Duty2    = 0
	Duty16_9 = 1
)

// I2CConfig is used to store config info for I2C.
type I2CConfig struct {
	Frequency uint32
	SCL       Pin
	SDA       Pin
	DutyCycle uint8
}

// Configure is intended to setup the STM32 I2C interface.
func (i2c I2C) Configure(config I2CConfig) {

	// The following is the required sequence in master mode.
	// 1. Program the peripheral input clock in I2C_CR2 Register in order to
	//    generate correct timings
	// 2. Configure the clock control registers
	// 3. Configure the rise time register
	// 4. Program the I2C_CR1 register to enable the peripheral
	// 5. Set the START bit in the I2C_CR1 register to generate a Start condition

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

	// default to 100 kHz (Sm, standard mode) if no frequency is set
	if config.Frequency == 0 {
		config.Frequency = TWI_FREQ_100KHZ
	}

	// configure I2C input clock
	i2c.Bus.CR2.SetBits(i2c.getFreqRange(config))

	// configure clock control
	i2c.Bus.CCR.Set(i2c.getSpeed(config))

	// configure rise time
	i2c.Bus.TRISE.Set(i2c.getRiseTime(config))

	// enable I2C interface
	i2c.Bus.CR1.ClearBits(stm32.I2C_CR1_PE)
}

// Tx does a single I2C transaction at the specified address.
// It clocks out the given address, writes the bytes in w, reads back len(r)
// bytes and stores them in r, and generates a stop condition on the bus.
func (i2c I2C) Tx(addr uint16, w, r []byte) error {
	var err error
	if len(w) != 0 {
		// start transmission for writing
		err = i2c.signalStart()
		if err != nil {
			return err
		}

		// send address
		err = i2c.sendAddress(uint8(addr), true)
		if err != nil {
			return err
		}

		for _, b := range w {
			err = i2c.WriteByte(b)
			if err != nil {
				return err
			}
		}

		// sending stop here for write
		err = i2c.signalStop()
		if err != nil {
			return err
		}
	}
	if len(r) != 0 {
		// re-start transmission for reading
		err = i2c.signalStart()
		if err != nil {
			return err
		}

		// 1 byte
		switch len(r) {
		case 1:
			// send address
			err = i2c.sendAddress(uint8(addr), false)
			if err != nil {
				return err
			}

			// Disable ACK of received data
			i2c.Bus.CR1.ClearBits(stm32.I2C_CR1_ACK)

			// clear timeout here
			timeout := i2cTimeout
			for !i2c.Bus.SR2.HasBits(stm32.I2C_SR2_MSL | stm32.I2C_SR2_BUSY) {
				timeout--
				if timeout == 0 {
					return errI2CWriteTimeout
				}
			}

			// Generate stop condition
			i2c.Bus.CR1.SetBits(stm32.I2C_CR1_STOP)

			timeout = i2cTimeout
			for !i2c.Bus.SR1.HasBits(stm32.I2C_SR1_RxNE) {
				timeout--
				if timeout == 0 {
					return errI2CReadTimeout
				}
			}

			// Read and return data byte from I2C data register
			r[0] = byte(i2c.Bus.DR.Get())

			// wait for stop
			return i2c.waitForStop()

		case 2:
			// enable pos
			i2c.Bus.CR1.SetBits(stm32.I2C_CR1_POS)

			// Enable ACK of received data
			i2c.Bus.CR1.SetBits(stm32.I2C_CR1_ACK)

			// send address
			err = i2c.sendAddress(uint8(addr), false)
			if err != nil {
				return err
			}

			// clear address here
			timeout := i2cTimeout
			for !i2c.Bus.SR2.HasBits(stm32.I2C_SR2_MSL | stm32.I2C_SR2_BUSY) {
				timeout--
				if timeout == 0 {
					return errI2CWriteTimeout
				}
			}

			// Disable ACK of received data
			i2c.Bus.CR1.ClearBits(stm32.I2C_CR1_ACK)

			// wait for btf. we need a longer timeout here than normal.
			timeout = 1000
			for !i2c.Bus.SR1.HasBits(stm32.I2C_SR1_BTF) {
				timeout--
				if timeout == 0 {
					return errI2CReadTimeout
				}
			}

			// Generate stop condition
			i2c.Bus.CR1.SetBits(stm32.I2C_CR1_STOP)

			// read the 2 bytes by reading twice.
			r[0] = byte(i2c.Bus.DR.Get())
			r[1] = byte(i2c.Bus.DR.Get())

			// wait for stop
			err = i2c.waitForStop()

			//disable pos
			i2c.Bus.CR1.ClearBits(stm32.I2C_CR1_POS)

			return err

		case 3:
			// Enable ACK of received data
			i2c.Bus.CR1.SetBits(stm32.I2C_CR1_ACK)

			// send address
			err = i2c.sendAddress(uint8(addr), false)
			if err != nil {
				return err
			}

			// clear address here
			timeout := i2cTimeout
			for !i2c.Bus.SR2.HasBits(stm32.I2C_SR2_MSL | stm32.I2C_SR2_BUSY) {
				timeout--
				if timeout == 0 {
					return errI2CWriteTimeout
				}
			}

			// Enable ACK of received data
			i2c.Bus.CR1.SetBits(stm32.I2C_CR1_ACK)

			// wait for btf. we need a longer timeout here than normal.
			timeout = 1000
			for !i2c.Bus.SR1.HasBits(stm32.I2C_SR1_BTF) {
				timeout--
				if timeout == 0 {
					return errI2CReadTimeout
				}
			}

			// Disable ACK of received data
			i2c.Bus.CR1.ClearBits(stm32.I2C_CR1_ACK)

			// read the first byte
			r[0] = byte(i2c.Bus.DR.Get())

			timeout = 1000
			for !i2c.Bus.SR1.HasBits(stm32.I2C_SR1_BTF) {
				timeout--
				if timeout == 0 {
					return errI2CReadTimeout
				}
			}

			// Generate stop condition
			i2c.Bus.CR1.SetBits(stm32.I2C_CR1_STOP)

			// read the last 2 bytes by reading twice.
			r[1] = byte(i2c.Bus.DR.Get())
			r[2] = byte(i2c.Bus.DR.Get())

			// wait for stop
			return i2c.waitForStop()

		default:
			// more than 3 bytes of data to read

			// send address
			err = i2c.sendAddress(uint8(addr), false)
			if err != nil {
				return err
			}

			// clear address here
			timeout := i2cTimeout
			for !i2c.Bus.SR2.HasBits(stm32.I2C_SR2_MSL | stm32.I2C_SR2_BUSY) {
				timeout--
				if timeout == 0 {
					return errI2CWriteTimeout
				}
			}

			for i := 0; i < len(r)-3; i++ {
				// Enable ACK of received data
				i2c.Bus.CR1.SetBits(stm32.I2C_CR1_ACK)

				// wait for btf. we need a longer timeout here than normal.
				timeout = 1000
				for !i2c.Bus.SR1.HasBits(stm32.I2C_SR1_BTF) {
					timeout--
					if timeout == 0 {
						return errI2CReadTimeout
					}
				}

				// read the next byte
				r[i] = byte(i2c.Bus.DR.Get())
			}

			// wait for btf. we need a longer timeout here than normal.
			timeout = 1000
			for !i2c.Bus.SR1.HasBits(stm32.I2C_SR1_BTF) {
				timeout--
				if timeout == 0 {
					return errI2CReadTimeout
				}
			}

			// Disable ACK of received data
			i2c.Bus.CR1.ClearBits(stm32.I2C_CR1_ACK)

			// get third from last byte
			r[len(r)-3] = byte(i2c.Bus.DR.Get())

			// Generate stop condition
			i2c.Bus.CR1.SetBits(stm32.I2C_CR1_STOP)

			// get second from last byte
			r[len(r)-2] = byte(i2c.Bus.DR.Get())

			timeout = i2cTimeout
			for !i2c.Bus.SR1.HasBits(stm32.I2C_SR1_RxNE) {
				timeout--
				if timeout == 0 {
					return errI2CReadTimeout
				}
			}

			// get last byte
			r[len(r)-1] = byte(i2c.Bus.DR.Get())

			// wait for stop
			return i2c.waitForStop()
		}
	}

	return nil
}

const i2cTimeout = 500

// signalStart sends a start signal.
func (i2c I2C) signalStart() error {
	// Wait until I2C is not busy
	timeout := i2cTimeout
	for i2c.Bus.SR2.HasBits(stm32.I2C_SR2_BUSY) {
		timeout--
		if timeout == 0 {
			return errI2CSignalStartTimeout
		}
	}

	// clear stop
	i2c.Bus.CR1.ClearBits(stm32.I2C_CR1_STOP)

	// Generate start condition
	i2c.Bus.CR1.SetBits(stm32.I2C_CR1_START)

	// Wait for I2C EV5 aka SB flag.
	timeout = i2cTimeout
	for !i2c.Bus.SR1.HasBits(stm32.I2C_SR1_SB) {
		timeout--
		if timeout == 0 {
			return errI2CSignalStartTimeout
		}
	}

	return nil
}

// signalStop sends a stop signal and waits for it to succeed.
func (i2c I2C) signalStop() error {
	// Generate stop condition
	i2c.Bus.CR1.SetBits(stm32.I2C_CR1_STOP)

	// wait for stop
	return i2c.waitForStop()
}

// waitForStop waits after a stop signal.
func (i2c I2C) waitForStop() error {
	// Wait until I2C is stopped
	timeout := i2cTimeout
	for i2c.Bus.SR1.HasBits(stm32.I2C_SR1_STOPF) {
		timeout--
		if timeout == 0 {
			return errI2CSignalStopTimeout
		}
	}

	return nil
}

// Send address of device we want to talk to
func (i2c I2C) sendAddress(address uint8, write bool) error {
	data := (address << 1)
	if !write {
		data |= 1 // set read flag
	}

	i2c.Bus.DR.Set(uint32(data))

	// Wait for I2C EV6 event.
	// Destination device acknowledges address
	timeout := i2cTimeout
	if write {
		// EV6 which is ADDR flag.
		for !i2c.Bus.SR1.HasBits(stm32.I2C_SR1_ADDR) {
			timeout--
			if timeout == 0 {
				return errI2CWriteTimeout
			}
		}

		timeout = i2cTimeout
		for !i2c.Bus.SR2.HasBits(stm32.I2C_SR2_MSL | stm32.I2C_SR2_BUSY | stm32.I2C_SR2_TRA) {
			timeout--
			if timeout == 0 {
				return errI2CWriteTimeout
			}
		}
	} else {
		// I2C_EVENT_MASTER_RECEIVER_MODE_SELECTED which is ADDR flag.
		for !i2c.Bus.SR1.HasBits(stm32.I2C_SR1_ADDR) {
			timeout--
			if timeout == 0 {
				return errI2CWriteTimeout
			}
		}
	}

	return nil
}

// WriteByte writes a single byte to the I2C bus.
func (i2c I2C) WriteByte(data byte) error {
	// Send data byte
	i2c.Bus.DR.Set(uint32(data))

	// Wait for I2C EV8_2 when data has been physically shifted out and
	// output on the bus.
	// I2C_EVENT_MASTER_BYTE_TRANSMITTED is TXE flag.
	timeout := i2cTimeout
	for !i2c.Bus.SR1.HasBits(stm32.I2C_SR1_TxE) {
		timeout--
		if timeout == 0 {
			return errI2CWriteTimeout
		}
	}

	return nil
}
