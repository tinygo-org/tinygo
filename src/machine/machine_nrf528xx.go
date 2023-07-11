//go:build nrf52840 || nrf52833

package machine

import (
	"device/nrf"
	"unsafe"
)

// I2C on the NRF528xx.
type I2C struct {
	Bus  *nrf.TWIM_Type // Called Bus to align with Bus field in nrf51
	BusT *nrf.TWIS_Type
	mode I2CMode
}

// There are 2 I2C interfaces on the NRF.
var (
	I2C0 = &I2C{Bus: nrf.TWIM0, BusT: nrf.TWIS0}
	I2C1 = &I2C{Bus: nrf.TWIM1, BusT: nrf.TWIS1}
)

func (i2c *I2C) enableAsController() {
	i2c.Bus.ENABLE.Set(nrf.TWIM_ENABLE_ENABLE_Enabled)
}

func (i2c *I2C) enableAsTarget() {
	i2c.BusT.ENABLE.Set(nrf.TWIS_ENABLE_ENABLE_Enabled)
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

// Listen starts listening for I2C requests sent to specified address
//
// addr is the address to listen to
func (i2c *I2C) Listen(addr uint8) error {
	i2c.BusT.ADDRESS[0].Set(uint32(addr))
	i2c.BusT.CONFIG.Set(nrf.TWIS_CONFIG_ADDRESS0_Enabled)

	i2c.BusT.EVENTS_STOPPED.Set(0)
	i2c.BusT.EVENTS_ERROR.Set(0)
	i2c.BusT.EVENTS_RXSTARTED.Set(0)
	i2c.BusT.EVENTS_TXSTARTED.Set(0)
	i2c.BusT.EVENTS_WRITE.Set(0)
	i2c.BusT.EVENTS_READ.Set(0)

	return nil
}

// WaitForEvent blocks the current go-routine until an I2C event is received (when in Target mode).
//
// The passed buffer will be populated for receive events, with the number of bytes
// received returned in count.  For other event types, buf is not modified and a count
// of zero is returned.
//
// For request events, the caller MUST call `Reply` to avoid hanging the i2c bus indefinitely.
func (i2c *I2C) WaitForEvent(buf []byte) (evt I2CTargetEvent, count int, err error) {
	i2c.BusT.RXD.PTR.Set(uint32(uintptr(unsafe.Pointer(&buf[0]))))
	i2c.BusT.RXD.MAXCNT.Set(uint32(len(buf)))

	i2c.BusT.TASKS_PREPARERX.Set(nrf.TWIS_TASKS_PREPARERX_TASKS_PREPARERX_Trigger)

	i2c.Bus.TASKS_RESUME.Set(1)

	for i2c.BusT.EVENTS_STOPPED.Get() == 0 &&
		i2c.BusT.EVENTS_READ.Get() == 0 {
		gosched()

		if i2c.BusT.EVENTS_ERROR.Get() != 0 {
			i2c.BusT.EVENTS_ERROR.Set(0)
			return I2CReceive, 0, twisError(i2c.BusT.ERRORSRC.Get())
		}
	}

	count = 0
	evt = I2CFinish
	err = nil

	if i2c.BusT.EVENTS_WRITE.Get() != 0 {
		i2c.BusT.EVENTS_WRITE.Set(0)

		// Data was sent to this target.  We've waited for
		// READ or STOPPED event, so transmission should be
		// complete.
		count = int(i2c.BusT.RXD.AMOUNT.Get())
		evt = I2CReceive
	} else if i2c.BusT.EVENTS_READ.Get() != 0 {
		i2c.BusT.EVENTS_READ.Set(0)

		// Data is requested from this target, hw will stretch
		// the controller's clock until there is a reply to
		// send
		evt = I2CRequest
	} else if i2c.BusT.EVENTS_STOPPED.Get() != 0 {
		i2c.BusT.EVENTS_STOPPED.Set(0)
		evt = I2CFinish
	}

	return
}

// Reply supplies the response data the controller.
func (i2c *I2C) Reply(buf []byte) error {
	i2c.BusT.TXD.PTR.Set(uint32(uintptr(unsafe.Pointer(&buf[0]))))
	i2c.BusT.TXD.MAXCNT.Set(uint32(len(buf)))

	i2c.BusT.EVENTS_STOPPED.Set(0)

	// Trigger Tx
	i2c.BusT.TASKS_PREPARETX.Set(nrf.TWIS_TASKS_PREPARETX_TASKS_PREPARETX_Trigger)

	// Block, waiting for Tx to complete
	for i2c.BusT.EVENTS_STOPPED.Get() == 0 {
		gosched()

		if i2c.BusT.EVENTS_ERROR.Get() != 0 {
			return twisError(i2c.BusT.ERRORSRC.Get())
		}
	}

	i2c.BusT.EVENTS_STOPPED.Set(0)

	return nil
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

// twisError converts an I2C target error to Go
func twisError(val uint32) error {
	if val == 0 {
		return nil
	} else if val&nrf.TWIS_ERRORSRC_OVERFLOW_Msk == nrf.TWIS_ERRORSRC_OVERFLOW {
		return errI2COverflow
	} else if val&nrf.TWIS_ERRORSRC_DNACK_Msk == nrf.TWIS_ERRORSRC_DNACK {
		return errI2CAckExpected
	} else if val&nrf.TWIS_ERRORSRC_OVERREAD_Msk == nrf.TWIS_ERRORSRC_OVERREAD {
		return errI2COverread
	}

	return errI2CBusError
}
