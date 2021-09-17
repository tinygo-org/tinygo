package machine

import (
	"device/arm"
	"errors"
)

// TWI_FREQ is the I2C bus speed. Normally either 100 kHz, or 400 kHz for high-speed bus.
const (
	TWI_FREQ_100KHZ = 100000
	TWI_FREQ_400KHZ = 400000
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
	errI2CConfigInvalid      = errors.New("I2C config error")
)

// SI2C is an I2C implementation by Software. Since it is implemented by
// software, it can be used with microcontrollers that do not have I2C
// function. This is not efficient but works around broken or missing drivers.
type SI2C struct {
	scl     Pin
	sda     Pin
	nack    bool
	waitNum uint32
}

// SI2CConfig is used to store config info for SI2C.
type SI2CConfig struct {
	Frequency uint32
	SCL       Pin
	SDA       Pin
}

// Configure is intended to setup the SI2C interface.
func (i2c *SI2C) Configure(config SI2CConfig) error {
	// Default SI2C bus speed is 100 kHz.
	if config.Frequency == 0 {
		config.Frequency = TWI_FREQ_100KHZ
	}

	if config.SDA == 0 || config.SCL == 0 {
		return errI2CConfigInvalid
	}
	i2c.scl = config.SCL
	i2c.sda = config.SDA

	i2c.SetBaudRate(config.Frequency)

	// enable pins
	i2c.sda.High()
	i2c.sda.Configure(PinConfig{Mode: PinOutput})
	i2c.scl.High()
	i2c.scl.Configure(PinConfig{Mode: PinOutput})

	return nil
}

// SetBaudRate sets the communication speed for the SI2C.
func (i2c *SI2C) SetBaudRate(br uint32) {
	// TODO: Should be more accurate
	i2c.waitNum = CPUFrequency() / br / 32
}

// Tx does a single SI2C transaction at the specified address.
// It clocks out the given address, writes the bytes in w, reads back len(r)
// bytes and stores them in r, and generates a stop condition on the bus.
func (i2c *SI2C) Tx(addr uint16, w, r []byte) error {
	var err error
	i2c.nack = false
	if len(w) != 0 {
		// send start/address for write
		i2c.sendAddress(addr, true)

		// wait until transmission complete

		// ACK received (0: ACK, 1: NACK)
		if i2c.nack {
			i2c.signalStop()
			return errI2CAckExpected
		}

		// write data
		for _, b := range w {
			err = i2c.WriteByte(b)
			if err != nil {
				return err
			}
		}

		err = i2c.signalStop()
		if err != nil {
			return err
		}
	}
	if len(r) != 0 {
		// send start/address for read
		i2c.sendAddress(addr, false)

		// wait transmission complete

		// ACK received (0: ACK, 1: NACK)
		if i2c.nack {
			i2c.signalStop()
			return errI2CAckExpected
		}

		// read first byte
		r[0] = i2c.readByte()
		for i := 1; i < len(r); i++ {
			// Send an ACK

			i2c.signalRead()

			// Read data and send the ACK
			r[i] = i2c.readByte()
		}

		// Send NACK to end transmission
		i2c.sendNack()

		err = i2c.signalStop()
		if err != nil {
			return err
		}
	}

	return nil
}

// WriteByte writes a single byte to the SI2C bus.
func (i2c *SI2C) WriteByte(data byte) error {
	// Send data byte
	i2c.scl.Low()
	i2c.sda.High()
	i2c.sda.Configure(PinConfig{Mode: PinOutput})
	i2c.waitHalfCycle()

	for i := 0; i < 8; i++ {
		i2c.scl.Low()
		if ((data >> (7 - i)) & 1) == 1 {
			i2c.sda.High()
		} else {
			i2c.sda.Low()
		}
		i2c.waitHalfCycle()
		i2c.waitHalfCycle()
		i2c.scl.High()
		i2c.waitHalfCycle()
		i2c.waitHalfCycle()
	}

	i2c.scl.Low()
	i2c.waitHalfCycle()
	i2c.waitHalfCycle()
	i2c.sda.Configure(PinConfig{Mode: PinInputPullup})
	i2c.scl.High()
	i2c.waitHalfCycle()

	i2c.nack = i2c.sda.Get()

	i2c.waitHalfCycle()

	// wait until transmission successful

	return nil
}

// sendAddress sends the address and start signal
func (i2c *SI2C) sendAddress(address uint16, write bool) error {
	data := (address << 1)
	if !write {
		data |= 1 // set read flag
	}

	i2c.scl.High()
	i2c.sda.Low()
	i2c.waitHalfCycle()
	i2c.waitHalfCycle()
	for i := 0; i < 8; i++ {
		i2c.scl.Low()
		if ((data >> (7 - i)) & 1) == 1 {
			i2c.sda.High()
		} else {
			i2c.sda.Low()
		}
		i2c.waitHalfCycle()
		i2c.waitHalfCycle()
		i2c.scl.High()
		i2c.waitHalfCycle()
		i2c.waitHalfCycle()
	}

	i2c.scl.Low()
	i2c.waitHalfCycle()
	i2c.waitHalfCycle()
	i2c.sda.Configure(PinConfig{Mode: PinInputPullup})
	i2c.scl.High()
	i2c.waitHalfCycle()

	i2c.nack = i2c.sda.Get()

	i2c.waitHalfCycle()

	// wait until bus ready

	return nil
}

func (i2c *SI2C) signalStop() error {
	i2c.scl.Low()
	i2c.sda.Low()
	i2c.sda.Configure(PinConfig{Mode: PinOutput})
	i2c.waitHalfCycle()
	i2c.waitHalfCycle()
	i2c.scl.High()
	i2c.waitHalfCycle()
	i2c.waitHalfCycle()
	i2c.sda.High()
	i2c.waitHalfCycle()
	i2c.waitHalfCycle()
	return nil
}

func (i2c *SI2C) signalRead() error {
	i2c.waitHalfCycle()
	i2c.waitHalfCycle()
	i2c.scl.Low()
	i2c.sda.Low()
	i2c.sda.Configure(PinConfig{Mode: PinOutput})
	i2c.waitHalfCycle()
	i2c.waitHalfCycle()
	i2c.scl.High()
	i2c.waitHalfCycle()
	i2c.waitHalfCycle()
	return nil
}

func (i2c *SI2C) readByte() byte {
	var data byte
	for i := 0; i < 8; i++ {
		i2c.scl.Low()
		i2c.sda.Configure(PinConfig{Mode: PinInputPullup})
		i2c.waitHalfCycle()
		i2c.waitHalfCycle()
		i2c.scl.High()
		if i2c.sda.Get() {
			data |= 1 << (7 - i)
		}
		i2c.waitHalfCycle()
		i2c.waitHalfCycle()
	}
	return data
}

func (i2c *SI2C) sendNack() error {
	i2c.waitHalfCycle()
	i2c.waitHalfCycle()
	i2c.scl.Low()
	i2c.sda.High()
	i2c.sda.Configure(PinConfig{Mode: PinOutput})
	i2c.waitHalfCycle()
	i2c.waitHalfCycle()
	i2c.scl.High()
	i2c.waitHalfCycle()
	i2c.waitHalfCycle()
	return nil
}

// WriteRegister transmits first the register and then the data to the
// peripheral device.
//
// Many I2C-compatible devices are organized in terms of registers. This method
// is a shortcut to easily write to such registers. Also, it only works for
// devices with 7-bit addresses, which is the vast majority.
func (i2c *SI2C) WriteRegister(address uint8, register uint8, data []byte) error {
	buf := make([]uint8, len(data)+1)
	buf[0] = register
	copy(buf[1:], data)
	return i2c.Tx(uint16(address), buf, nil)
}

// ReadRegister transmits the register, restarts the connection as a read
// operation, and reads the response.
//
// Many I2C-compatible devices are organized in terms of registers. This method
// is a shortcut to easily read such registers. Also, it only works for devices
// with 7-bit addresses, which is the vast majority.
func (i2c *SI2C) ReadRegister(address uint8, register uint8, data []byte) error {
	return i2c.Tx(uint16(address), []byte{register}, data)
}

func (i2c *SI2C) waitHalfCycle() {
	for i := uint32(0); i < i2c.waitNum; i++ {
		arm.Asm("nop")
	}
}
