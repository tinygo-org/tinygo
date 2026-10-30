//go:build nxp && mk20d7

package machine

import (
	"device/nxp"
	"runtime/volatile"
)

// I2C on the MK20DX256, in polled controller mode.
type I2C struct {
	Bus      *nxp.I2C_Type
	SCGC     *volatile.Register32
	SCGCMask uint32
	SDA, SCL Pin
	mux      uint32
}

// I2CConfig is used to store config info for I2C.
type I2CConfig struct {
	Frequency uint32
	SDA       Pin
	SCL       Pin
}

var (
	I2C0 = &_I2C0
	I2C1 = &_I2C1

	_I2C0 = I2C{Bus: nxp.I2C0, SCGC: &nxp.SIM.SCGC4, SCGCMask: nxp.SIM_SCGC4_I2C0, SDA: I2C0_SDA_PIN, SCL: I2C0_SCL_PIN, mux: 2}
	_I2C1 = I2C{Bus: nxp.I2C1, SCGC: &nxp.SIM.SCGC4, SCGCMask: nxp.SIM_SCGC4_I2C1, SDA: I2C1_SDA_PIN, SCL: I2C1_SCL_PIN, mux: 2}
)

// i2cDividers is the SCL divider for each ICR value with MULT = 1.
// K20 Sub-Family Reference Manual, I2C chapter, table "I2C divider and hold values".
var i2cDividers = [64]uint16{
	20, 22, 24, 26, 28, 30, 34, 40, 28, 32, 36, 40, 44, 48, 56, 68,
	48, 56, 64, 72, 80, 88, 104, 128, 80, 96, 112, 128, 144, 160, 192, 240,
	160, 192, 224, 256, 288, 320, 384, 480, 320, 384, 448, 512, 576, 640, 768, 960,
	640, 768, 896, 1024, 1152, 1280, 1536, 1920, 1280, 1536, 1792, 2048, 2304, 2560, 3072, 3840,
}

const i2cTimeout = 100000

// Configure is intended to setup the I2C interface.
func (i2c *I2C) Configure(config I2CConfig) error {
	if config.Frequency == 0 {
		config.Frequency = 100 * KHz
	}
	if config.SDA == 0 && config.SCL == 0 {
		config.SDA, config.SCL = i2c.SDA, i2c.SCL
	}

	i2c.SCGC.SetBits(i2c.SCGCMask)
	i2c.Bus.C1.Set(0)

	pcr := (i2c.mux << nxp.PORT_PCR_MUX_Pos) | nxp.PORT_PCR_ODE | nxp.PORT_PCR_SRE | nxp.PORT_PCR_DSE
	config.SDA.Control().Set(pcr)
	config.SCL.Control().Set(pcr)

	i2c.SetBaudRate(config.Frequency)
	i2c.Bus.FLT.Set(4)
	i2c.Bus.C2.Set(nxp.I2C_C2_HDRS)
	i2c.Bus.C1.Set(nxp.I2C_C1_IICEN)
	return nil
}

// SetBaudRate sets the SCL frequency to the nearest value that is not higher.
func (i2c *I2C) SetBaudRate(br uint32) error {
	div := BusFrequency() / br
	best := 63
	for icr, d := range i2cDividers {
		if uint32(d) >= div && d < i2cDividers[best] {
			best = icr
		}
	}
	i2c.Bus.F.Set(uint8(best))
	return nil
}

// Tx does a single I2C transaction at the specified address.
func (i2c *I2C) Tx(addr uint16, w, r []byte) error {
	err := i2c.tx(uint8(addr), w, r)
	if err != nil {
		i2c.stop()
	}
	return err
}

func (i2c *I2C) tx(addr uint8, w, r []byte) error {
	for n := 0; i2c.Bus.S.HasBits(nxp.I2C_S_BUSY); n++ {
		if n > i2cTimeout {
			return errI2CBusReadyTimeout
		}
	}
	i2c.Bus.S.Set(nxp.I2C_S_IICIF | nxp.I2C_S_ARBL)

	if len(w) > 0 || len(r) == 0 {
		if err := i2c.start(); err != nil {
			return err
		}
		if err := i2c.writeByte(addr << 1); err != nil {
			return err
		}
		for _, b := range w {
			if err := i2c.writeByte(b); err != nil {
				return err
			}
		}
		if len(r) == 0 {
			i2c.stop()
			return nil
		}
		i2c.Bus.C1.Set(nxp.I2C_C1_IICEN | nxp.I2C_C1_MST | nxp.I2C_C1_TX | nxp.I2C_C1_RSTA)
	} else if err := i2c.start(); err != nil {
		return err
	}
	if err := i2c.writeByte(addr<<1 | 1); err != nil {
		return err
	}

	// The last byte gets a NACK, and the STOP goes out before the last
	// read of D so that the read does not start a new transfer.
	c1 := uint8(nxp.I2C_C1_IICEN | nxp.I2C_C1_MST)
	if len(r) == 1 {
		c1 |= nxp.I2C_C1_TXAK
	}
	i2c.Bus.C1.Set(c1)
	i2c.Bus.D.Get()
	for i := range r {
		if err := i2c.wait(); err != nil {
			return err
		}
		switch {
		case i == len(r)-1:
			i2c.stop()
		case i == len(r)-2:
			i2c.Bus.C1.Set(nxp.I2C_C1_IICEN | nxp.I2C_C1_MST | nxp.I2C_C1_TXAK)
		}
		r[i] = i2c.Bus.D.Get()
	}
	return nil
}

// start sends a START and waits until the bus is busy, else D is ignored.
func (i2c *I2C) start() error {
	i2c.Bus.C1.Set(nxp.I2C_C1_IICEN | nxp.I2C_C1_MST | nxp.I2C_C1_TX)
	for n := 0; !i2c.Bus.S.HasBits(nxp.I2C_S_BUSY); n++ {
		if n > i2cTimeout {
			return errI2CSignalStartTimeout
		}
	}
	return nil
}

func (i2c *I2C) writeByte(b uint8) error {
	i2c.Bus.D.Set(b)
	if err := i2c.wait(); err != nil {
		return err
	}
	if i2c.Bus.S.HasBits(nxp.I2C_S_RXAK) {
		return errI2CAckExpected
	}
	return nil
}

func (i2c *I2C) wait() error {
	for n := 0; !i2c.Bus.S.HasBits(nxp.I2C_S_IICIF); n++ {
		if n > i2cTimeout {
			return errI2CWriteTimeout
		}
	}
	s := i2c.Bus.S.Get()
	i2c.Bus.S.Set(nxp.I2C_S_IICIF)
	if s&nxp.I2C_S_ARBL != 0 {
		i2c.Bus.S.Set(nxp.I2C_S_ARBL)
		return errI2CBusError
	}
	return nil
}

func (i2c *I2C) stop() {
	i2c.Bus.C1.Set(nxp.I2C_C1_IICEN | nxp.I2C_C1_TX)
	for n := 0; i2c.Bus.S.HasBits(nxp.I2C_S_BUSY) && n < i2cTimeout; n++ {
	}
}
