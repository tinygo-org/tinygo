//go:build nrf52840 || nrf52833

package machine

import (
	"device/nrf"
	"unsafe"
)

// I2C on the NRF528xx.
type I2C struct {
	Bus *nrf.TWIM_Type
}

// There are 2 I2C interfaces on the NRF.
var (
	I2C0 = &I2C{Bus: nrf.TWIM0}
	I2C1 = &I2C{Bus: nrf.TWIM1}
)

func (i2c *I2C) enableAsController() {
	i2c.Bus.ENABLE.Set(nrf.TWIM_ENABLE_ENABLE_Enabled)
}

func (i2c *I2C) disable() {
	i2c.Bus.ENABLE.Set(0)
}

// Tx does a single I2C transaction at the specified address (when in controller mode).
//
// It clocks out the given address, writes the bytes in w, reads back len(r)
// bytes and stores them in r, and generates a stop condition on the bus.
func (i2c *I2C) Tx(addr uint16, w, r []byte) (err error) {
	i2c.Bus.ADDRESS.Set(uint32(addr))

	i2c.Bus.EVENTS_STOPPED.Set(0)
	i2c.Bus.EVENTS_ERROR.Set(0)
	i2c.Bus.EVENTS_RXSTARTED.Set(0)
	i2c.Bus.EVENTS_TXSTARTED.Set(0)
	i2c.Bus.EVENTS_LASTRX.Set(0)
	i2c.Bus.EVENTS_LASTTX.Set(0)
	i2c.Bus.EVENTS_SUSPENDED.Set(0)

	// Configure for a single shot to perform both write and read (as applicable)
	if len(w) != 0 {
		i2c.Bus.TXD.PTR.Set(uint32(uintptr(unsafe.Pointer(&w[0]))))
		i2c.Bus.TXD.MAXCNT.Set(uint32(len(w)))

		// If no read, immediately signal stop after TX
		if len(r) == 0 {
			i2c.Bus.SHORTS.Set(nrf.TWIM_SHORTS_LASTTX_STOP)
		}
	}
	if len(r) != 0 {
		i2c.Bus.RXD.PTR.Set(uint32(uintptr(unsafe.Pointer(&r[0]))))
		i2c.Bus.RXD.MAXCNT.Set(uint32(len(r)))

		// Auto-start Rx after Tx and Stop after Rx
		i2c.Bus.SHORTS.Set(nrf.TWIM_SHORTS_LASTTX_STARTRX | nrf.TWIM_SHORTS_LASTRX_STOP)
	}

	// Fire the transaction
	i2c.Bus.TASKS_RESUME.Set(1)
	if len(w) != 0 {
		i2c.Bus.TASKS_STARTTX.Set(1)
	} else if len(r) != 0 {
		i2c.Bus.TASKS_STARTRX.Set(1)
	}

	// Wait until transaction stopped to ensure buffers fully processed
	for i2c.Bus.EVENTS_STOPPED.Get() == 0 {
		// Allow scheduler to run
		gosched()

		// Handle errors by ensuring STOP sent on bus
		if i2c.Bus.EVENTS_ERROR.Get() != 0 {
			if i2c.Bus.EVENTS_STOPPED.Get() == 0 {
				// STOP cannot be sent during SUSPEND
				i2c.Bus.TASKS_RESUME.Set(1)
				i2c.Bus.TASKS_STOP.Set(1)
			}
			err = twiCError(i2c.Bus.ERRORSRC.Get())
		}
	}

	return
}

// twiCError converts an I2C controller error to Go
func twiCError(val uint32) error {
	if val == 0 {
		return nil
	} else if val&nrf.TWIM_ERRORSRC_OVERRUN_Msk == nrf.TWIM_ERRORSRC_OVERRUN {
		return errI2CBusError
	} else if val&nrf.TWIM_ERRORSRC_ANACK_Msk == nrf.TWIM_ERRORSRC_ANACK {
		return errI2CAckExpected
	} else if val&nrf.TWIM_ERRORSRC_DNACK_Msk == nrf.TWIM_ERRORSRC_DNACK {
		return errI2CAckExpected
	}

	return errI2CBusError
}
