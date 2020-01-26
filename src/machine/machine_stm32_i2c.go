// +build stm32

package machine

// Peripheral abstraction layer for I2C on the stm32 family

import (
	"device/stm32"
	"errors"
	"unsafe"
)

// I2C on the STM32F103xx.
type I2C struct {
	Bus *stm32.I2C_Type
}

// Since the first interface is named I2C1, both I2C0 and I2C1 refer to I2C1.
// TODO: implement I2C2.
var (
	I2C1 = I2C{Bus: stm32.I2C1}
	I2C0 = I2C1
)

// I2CConfig is used to store config info for I2C.
type I2CConfig struct {
	Frequency uint32
	SCL       Pin
	SDA       Pin
}

// Configure is intended to setup the I2C interface.
func (i2c I2C) Configure(config I2CConfig) {
	// Default I2C bus speed is 100 kHz.
	if config.Frequency == 0 {
		config.Frequency = TWI_FREQ_100KHZ
	}

	// enable clock for I2C
	EnableAltFuncClock(unsafe.Pointer(i2c.Bus))

	// I2C1 pins
	i2c.configurePins(config)

	// Disable the selected I2C peripheral to configure
	i2c.Bus.CR1.ClearBits(stm32.I2C_CR1_PE)

	// pclk1 clock speed is main frequency divided by PCLK1 prescaler (div 2)
	pclk1 := uint32(CPUFrequency() / 2)

	// set freqency range to PCLK1 clock speed in MHz
	// aka setting the value 36 means to use 36 MHz clock
	pclk1Mhz := pclk1 / 1000000
	i2c.Bus.CR2.SetBits(pclk1Mhz)

	switch config.Frequency {
	case TWI_FREQ_100KHZ:
		// Normal mode speed calculation
		ccr := pclk1 / (config.Frequency * 2)
		i2c.Bus.CCR.Set(ccr)

		// duty cycle 2
		i2c.Bus.CCR.ClearBits(stm32.I2C_CCR_DUTY)

		// frequency standard mode
		i2c.Bus.CCR.ClearBits(stm32.I2C_CCR_F_S)

		// Set Maximum Rise Time for standard mode
		i2c.Bus.TRISE.Set(pclk1Mhz)

	case TWI_FREQ_400KHZ:
		// Fast mode speed calculation
		ccr := pclk1 / (config.Frequency * 3)
		i2c.Bus.CCR.Set(ccr)

		// duty cycle 2
		i2c.Bus.CCR.ClearBits(stm32.I2C_CCR_DUTY)

		// frequency fast mode
		i2c.Bus.CCR.SetBits(stm32.I2C_CCR_F_S)

		// Set Maximum Rise Time for fast mode
		i2c.Bus.TRISE.Set(((pclk1Mhz * 300) / 1000))
	}

	// re-enable the selected I2C peripheral
	i2c.Bus.CR1.SetBits(stm32.I2C_CR1_PE)
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
					return errors.New("I2C timeout on read clear address")
				}
			}

			// Generate stop condition
			i2c.Bus.CR1.SetBits(stm32.I2C_CR1_STOP)

			timeout = i2cTimeout
			for !i2c.Bus.SR1.HasBits(stm32.I2C_SR1_RxNE) {
				timeout--
				if timeout == 0 {
					return errors.New("I2C timeout on read 1 byte")
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
					return errors.New("I2C timeout on read clear address")
				}
			}

			// Disable ACK of received data
			i2c.Bus.CR1.ClearBits(stm32.I2C_CR1_ACK)

			// wait for btf. we need a longer timeout here than normal.
			timeout = 1000
			for !i2c.Bus.SR1.HasBits(stm32.I2C_SR1_BTF) {
				timeout--
				if timeout == 0 {
					return errors.New("I2C timeout on read 2 bytes")
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
					return errors.New("I2C timeout on read clear address")
				}
			}

			// Enable ACK of received data
			i2c.Bus.CR1.SetBits(stm32.I2C_CR1_ACK)

			// wait for btf. we need a longer timeout here than normal.
			timeout = 1000
			for !i2c.Bus.SR1.HasBits(stm32.I2C_SR1_BTF) {
				timeout--
				if timeout == 0 {
					println("I2C timeout on read 3 bytes")
					return errors.New("I2C timeout on read 3 bytes")
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
					return errors.New("I2C timeout on read 3 bytes")
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
					return errors.New("I2C timeout on read clear address")
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
						println("I2C timeout on read 3 bytes")
						return errors.New("I2C timeout on read 3 bytes")
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
					return errors.New("I2C timeout on read more than 3 bytes")
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
					return errors.New("I2C timeout on read last byte of more than 3")
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
			return errors.New("I2C busy on start")
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
			return errors.New("I2C timeout on start")
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
			println("I2C timeout on wait for stop signal")
			return errors.New("I2C timeout on wait for stop signal")
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
				return errors.New("I2C timeout on send write address")
			}
		}

		timeout = i2cTimeout
		for !i2c.Bus.SR2.HasBits(stm32.I2C_SR2_MSL | stm32.I2C_SR2_BUSY | stm32.I2C_SR2_TRA) {
			timeout--
			if timeout == 0 {
				return errors.New("I2C timeout on send write address")
			}
		}
	} else {
		// I2C_EVENT_MASTER_RECEIVER_MODE_SELECTED which is ADDR flag.
		for !i2c.Bus.SR1.HasBits(stm32.I2C_SR1_ADDR) {
			timeout--
			if timeout == 0 {
				return errors.New("I2C timeout on send read address")
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
			return errors.New("I2C timeout on write")
		}
	}

	return nil
}
