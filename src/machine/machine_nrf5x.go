//go:build nrf51 || nrf52

package machine

import "device/nrf"

// I2C on the NRF51 and NRF52.
type I2C struct {
	Bus *nrf.TWI_Type
}

// There are 2 I2C interfaces on the NRF.
var (
	I2C0 = &I2C{Bus: nrf.TWI0}
	I2C1 = &I2C{Bus: nrf.TWI1}
)

func (i2c *I2C) enableAsController() {
	i2c.Bus.ENABLE.Set(nrf.TWI_ENABLE_ENABLE_Enabled)
}

func (i2c *I2C) disable() {
	i2c.Bus.ENABLE.Set(0)
}

// Tx does a single I2C transaction at the specified address.
// It clocks out the given address, writes the bytes in w, reads back len(r)
// bytes and stores them in r, and generates a stop condition on the bus.
func (i2c *I2C) Tx(addr uint16, w, r []byte) (err error) {

	// Tricky stop condition.
	// After reads, the stop condition is generated implicitly with a shortcut.
	// After writes not followed by reads and in the case of errors, stop must be generated explicitly.

	i2c.Bus.ADDRESS.Set(uint32(addr))

	if len(w) != 0 {
		i2c.Bus.TASKS_STARTTX.Set(1) // start transmission for writing
		for _, b := range w {
			if err = i2c.writeByte(b); err != nil {
				i2c.signalStop()
				return
			}
		}
	}

	if len(r) != 0 {
		// To trigger suspend task when a byte is received
		i2c.Bus.SHORTS.Set(nrf.TWI_SHORTS_BB_SUSPEND)
		i2c.Bus.TASKS_STARTRX.Set(1) // re-start transmission for reading
		for i := range r {           // read each char
			if i+1 == len(r) {
				// To trigger stop task when last byte is received, set before resume task.
				i2c.Bus.SHORTS.Set(nrf.TWI_SHORTS_BB_STOP)
			}
			if i > 0 {
				i2c.Bus.TASKS_RESUME.Set(1) // re-start transmission for reading
			}
			if r[i], err = i2c.readByte(); err != nil {
				i2c.Bus.SHORTS.Set(nrf.TWI_SHORTS_BB_SUSPEND_Disabled)
				i2c.signalStop()
				return
			}
		}
		i2c.Bus.SHORTS.Set(nrf.TWI_SHORTS_BB_SUSPEND_Disabled)
	}

	// Stop explicitly when no reads were executed, stoping unconditionally would be a mistake.
	// It may execute after I2C peripheral has already been stopped by the shortcut in the read block,
	// so stop task will trigger first thing in a subsequent transaction, hanging it.
	if len(r) == 0 {
		i2c.signalStop()
	}

	return
}

// writeByte writes a single byte to the I2C bus and waits for confirmation.
func (i2c *I2C) writeByte(data byte) error {
	i2c.Bus.TXD.Set(uint32(data))
	for i2c.Bus.EVENTS_TXDSENT.Get() == 0 {
		if e := i2c.Bus.EVENTS_ERROR.Get(); e != 0 {
			i2c.Bus.EVENTS_ERROR.Set(0)
			return errI2CBusError
		}
	}
	i2c.Bus.EVENTS_TXDSENT.Set(0)
	return nil
}

// readByte reads a single byte from the I2C bus when it is ready.
func (i2c *I2C) readByte() (byte, error) {
	for i2c.Bus.EVENTS_RXDREADY.Get() == 0 {
		if e := i2c.Bus.EVENTS_ERROR.Get(); e != 0 {
			i2c.Bus.EVENTS_ERROR.Set(0)
			return 0, errI2CBusError
		}
	}
	i2c.Bus.EVENTS_RXDREADY.Set(0)
	return byte(i2c.Bus.RXD.Get()), nil
}
