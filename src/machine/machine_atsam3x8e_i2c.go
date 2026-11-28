//go:build sam && atsam3x8e

package machine

import (
	"device/sam"
	"errors"
)

var (
	errI2CTimeout     = errors.New("i2c: timeout")
	errI2CInvalidRate = errors.New("i2c: invalid baud rate")
)

const i2cTimeout = 1_000_000

// I2CConfig is used to store config info for I2C.
type I2CConfig struct {
	Frequency uint32
	SCL       Pin
	SDA       Pin
}

// I2C on the SAM3X8E. The pins of each TWI are fixed, so SCL and SDA are ignored.
type I2C struct {
	Bus      *sam.TWI_Type
	id       uint32
	sda, scl Pin
}

// I2C0 is TWI1 on SDA and SCL (D20 and D21), which Arduino calls Wire.
// I2C1 is TWI0 on SDA1 and SCL1, which Arduino calls Wire1.
var (
	I2C0 = &I2C{Bus: sam.TWI1, id: sam.IRQ_TWI1, sda: PB12, scl: PB13}
	I2C1 = &I2C{Bus: sam.TWI0, id: sam.IRQ_TWI0, sda: PA17, scl: PA18}
)

// Configure enables the TWI as a controller.
func (i2c *I2C) Configure(config I2CConfig) error {
	if config.Frequency == 0 {
		config.Frequency = TWI_FREQ_100KHZ
	}
	i2c.recover()
	sam.PMC.PMC_PCER0.Set(1 << i2c.id)

	i2c.Bus.IDR.Set(0xffffffff)
	i2c.Bus.SR.Get()
	i2c.Bus.CR.Set(sam.TWI_CR_SWRST)
	i2c.Bus.RHR.Get()
	i2c.Bus.CR.Set(sam.TWI_CR_SVDIS)
	i2c.Bus.CR.Set(sam.TWI_CR_MSDIS)
	i2c.Bus.CR.Set(sam.TWI_CR_MSEN)
	if err := i2c.SetBaudRate(config.Frequency); err != nil {
		return err
	}
	i2c.sda.Configure(PinConfig{Mode: PinPeripheralA})
	i2c.scl.Configure(PinConfig{Mode: PinPeripheralA})
	return nil
}

// SetBaudRate sets the SCL frequency. The result is never above br.
func (i2c *I2C) SetBaudRate(br uint32) error {
	if br == 0 {
		return errI2CInvalidRate
	}
	period := (CPUFrequency() + br - 1) / br
	low := period / 2
	if br > TWI_FREQ_100KHZ {
		// Fast mode needs SCL low for at least 1.3 us, see the I2C specification UM10204 table 10.
		low = period * 2 / 3
	}
	high := period - low

	// Each half lasts (DIV << CKDIV) + 4 clock cycles, see TWI_CWGR in the SAM3X/A datasheet 11057C.
	for ckdiv := uint32(0); ckdiv < 8; ckdiv++ {
		cl := divCeil(sub4(low), 1<<ckdiv)
		ch := divCeil(sub4(high), 1<<ckdiv)
		if cl <= 255 && ch <= 255 {
			i2c.Bus.CWGR.Set(cl<<sam.TWI_CWGR_CLDIV_Pos | ch<<sam.TWI_CWGR_CHDIV_Pos | ckdiv<<sam.TWI_CWGR_CKDIV_Pos)
			return nil
		}
	}
	return errI2CInvalidRate
}

func sub4(n uint32) uint32 {
	if n < 4 {
		return 0
	}
	return n - 4
}

func divCeil(n, d uint32) uint32 {
	return (n + d - 1) / d
}

// Tx writes w and then reads into r from the target at addr. With both empty
// it only sends the address, which can be used to scan the bus.
func (i2c *I2C) Tx(addr uint16, w, r []byte) error {
	i2c.Bus.SR.Get()
	if len(r) == 0 {
		return i2c.write(addr, w)
	}
	// The TWI sends up to 3 bytes as an internal address with a repeated start.
	if len(w) > 3 {
		if err := i2c.write(addr, w); err != nil {
			return err
		}
		w = nil
	}
	return i2c.read(addr, w, r)
}

func (i2c *I2C) write(addr uint16, w []byte) error {
	bus := i2c.Bus
	bus.MMR.Set(uint32(addr) << sam.TWI_MMR_DADR_Pos)
	if len(w) == 0 {
		bus.CR.Set(sam.TWI_CR_QUICK)
		// TXCOMP stays set from the last transfer until this one starts.
		for i := 0; i < 1000 && bus.SR.HasBits(sam.TWI_SR_TXCOMP); i++ {
		}
		return i2c.wait(sam.TWI_SR_TXCOMP)
	}
	for _, b := range w {
		bus.THR.Set(uint32(b))
		if err := i2c.wait(sam.TWI_SR_TXRDY); err != nil {
			return err
		}
	}
	bus.CR.Set(sam.TWI_CR_STOP)
	return i2c.wait(sam.TWI_SR_TXCOMP)
}

func (i2c *I2C) read(addr uint16, iadr, r []byte) error {
	bus := i2c.Bus
	var ia uint32
	for _, b := range iadr {
		ia = ia<<8 | uint32(b)
	}
	bus.MMR.Set(uint32(addr)<<sam.TWI_MMR_DADR_Pos | sam.TWI_MMR_MREAD | uint32(len(iadr))<<sam.TWI_MMR_IADRSZ_Pos)
	bus.IADR.Set(ia)

	// STOP must be set with START for one byte, else after the next to last byte.
	// See Master Receiver Mode in the TWI chapter of the SAM3X/A datasheet 11057C.
	if len(r) == 1 {
		bus.CR.Set(sam.TWI_CR_START | sam.TWI_CR_STOP)
	} else {
		bus.CR.Set(sam.TWI_CR_START)
	}
	for i := range r {
		if i == len(r)-1 && len(r) > 1 {
			bus.CR.Set(sam.TWI_CR_STOP)
		}
		if err := i2c.wait(sam.TWI_SR_RXRDY); err != nil {
			return err
		}
		r[i] = byte(bus.RHR.Get())
	}
	return i2c.wait(sam.TWI_SR_TXCOMP)
}

// wait returns when a flag in mask is set, or with an error on NACK, lost
// arbitration or timeout.
func (i2c *I2C) wait(mask uint32) error {
	for i := 0; i < i2cTimeout; i++ {
		sr := i2c.Bus.SR.Get()
		switch {
		case sr&sam.TWI_SR_NACK != 0:
			i2c.waitIdle()
			return errI2CAckExpected
		case sr&sam.TWI_SR_ARBLST != 0:
			return errI2CBusError
		case sr&mask != 0:
			return nil
		}
	}
	return errI2CTimeout
}

// waitIdle waits for the transfer to end after a NACK.
func (i2c *I2C) waitIdle() {
	for i := 0; i < i2cTimeout; i++ {
		if i2c.Bus.SR.HasBits(sam.TWI_SR_TXCOMP) {
			return
		}
	}
}

// recover frees the bus from a target that holds SDA low after an
// interrupted transfer, with up to 9 SCL pulses and a STOP. See the I2C
// specification UM10204 section 3.1.16.
func (i2c *I2C) recover() {
	release := func(p Pin) { p.Configure(PinConfig{Mode: PinInputPullup}) }
	drive := func(p Pin) { p.Configure(PinConfig{Mode: PinOutput}); p.Low() }
	wait := func() {
		for i := 0; i < 400; i++ {
			i2c.sda.Get()
		}
	}

	release(i2c.sda)
	release(i2c.scl)
	wait()
	for i := 0; i < 9 && !i2c.sda.Get(); i++ {
		drive(i2c.scl)
		wait()
		release(i2c.scl)
		wait()
	}
	drive(i2c.sda)
	wait()
	release(i2c.sda)
	wait()
}
