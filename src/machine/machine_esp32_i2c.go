//go:build esp32

package machine

import (
	"device/esp"
	"runtime/volatile"
	"unsafe"
)

var (
	I2C0 = &I2C{Bus: esp.I2C0, funcSCL: 29, funcSDA: 30}
	I2C1 = &I2C{Bus: esp.I2C1, funcSCL: 95, funcSDA: 96}
)

type I2C struct {
	Bus              *esp.I2C_Type
	funcSCL, funcSDA uint32
	config           I2CConfig
}

// I2CConfig is used to store config info for I2C.
type I2CConfig struct {
	Frequency uint32 // in Hz
	SCL       Pin
	SDA       Pin
}

const (
	i2cClkSourceFrequency = uint32(80 * MHz)
)

func (i2c *I2C) Configure(config I2CConfig) error {
	if config.Frequency == 0 {
		config.Frequency = 400 * KHz
	}
	if config.SCL == 0 {
		config.SCL = SCL_PIN
	}
	if config.SDA == 0 {
		config.SDA = SDA_PIN
	}
	i2c.config = config

	i2c.initAll()
	return nil
}

func (i2c *I2C) initAll() {
	i2c.initClock()
	i2c.initNoiseFilter()
	i2c.initPins()
	i2c.initFrequency()
	i2c.startMaster()
}

//go:inline
func (i2c *I2C) initClock() {
	// reset I2C clock
	if i2c.Bus == esp.I2C0 {
		esp.DPORT.SetPERIP_RST_EN_I2C0_EXT0_RST(1)
		esp.DPORT.SetPERIP_CLK_EN_I2C0_EXT0_CLK_EN(1)
		esp.DPORT.SetPERIP_RST_EN_I2C0_EXT0_RST(0)
	} else {
		esp.DPORT.SetPERIP_RST_EN_I2C_EXT1_RST(1)
		esp.DPORT.SetPERIP_CLK_EN_I2C_EXT1_CLK_EN(1)
		esp.DPORT.SetPERIP_RST_EN_I2C_EXT1_RST(0)
	}
	// disable interrupts
	i2c.Bus.INT_ENA.Set(0)
	i2c.Bus.INT_CLR.Set(0x3fff)

	i2c.Bus.SetCTR_CLK_EN(1)
}

//go:inline
func (i2c *I2C) initNoiseFilter() {
	i2c.Bus.SCL_FILTER_CFG.Set(0xF)
	i2c.Bus.SDA_FILTER_CFG.Set(0xF)
}

//go:inline
func (i2c *I2C) initPins() {
	var muxConfig uint32
	const function = 2 // function 2 is just GPIO

	// SDA
	muxConfig = function << esp.IO_MUX_GPIO0_MCU_SEL_Pos
	// Make this pin an input pin (always).
	muxConfig |= esp.IO_MUX_GPIO0_FUN_IE
	// Set drive strength: 0 is lowest, 3 is highest.
	muxConfig |= 1 << esp.IO_MUX_GPIO0_FUN_DRV_Pos
	i2c.config.SDA.mux().Set(muxConfig)
	i2c.config.SDA.outFunc().Set(i2c.funcSDA)
	inFunc(i2c.funcSDA).Set(uint32(esp.GPIO_FUNC_IN_SEL_CFG_SEL | i2c.config.SDA))
	i2c.config.SDA.Set(true)
	// Configure the pad with the given IO mux configuration.
	i2c.config.SDA.pinReg().SetBits(esp.GPIO_PIN_PAD_DRIVER)

	esp.GPIO.ENABLE_W1TS.Set(1 << int(i2c.config.SDA))
	i2c.Bus.SetCTR_SDA_FORCE_OUT(1)

	// SCL
	muxConfig = function << esp.IO_MUX_GPIO0_MCU_SEL_Pos
	// Make this pin an input pin (always).
	muxConfig |= esp.IO_MUX_GPIO0_FUN_IE
	// Set drive strength: 0 is lowest, 3 is highest.
	muxConfig |= 1 << esp.IO_MUX_GPIO0_FUN_DRV_Pos
	i2c.config.SCL.mux().Set(muxConfig)
	i2c.config.SCL.outFunc().Set(i2c.funcSCL)
	inFunc(i2c.funcSCL).Set(uint32(esp.GPIO_FUNC_IN_SEL_CFG_SEL | i2c.config.SCL))
	i2c.config.SCL.Set(true)
	// Configure the pad with the given IO mux configuration.
	i2c.config.SCL.pinReg().SetBits(esp.GPIO_PIN_PAD_DRIVER)

	esp.GPIO.ENABLE_W1TS.Set(1 << int(i2c.config.SCL))
	i2c.Bus.SetCTR_SCL_FORCE_OUT(1)
}

//go:inline
func (i2c *I2C) initFrequency() {
	clkmDiv := i2cClkSourceFrequency/(i2c.config.Frequency*1024) + 1
	sclkFreq := i2cClkSourceFrequency / clkmDiv
	halfCycle := sclkFreq / i2c.config.Frequency / 2
	//SCL
	sclLow := halfCycle
	sclWaitHigh := uint32(0)
	if i2c.config.Frequency > 50000 {
		sclWaitHigh = halfCycle / 8 // compensate the time when freq > 50K
	}
	sclHigh := halfCycle - sclWaitHigh
	// SDA
	sdaHold := halfCycle / 4
	sda_sample := halfCycle / 2
	setup := halfCycle
	hold := halfCycle

	i2c.Bus.SetSCL_LOW_PERIOD(sclLow - 1)
	i2c.Bus.SetSCL_HIGH_PERIOD(sclHigh)
	i2c.Bus.SetSCL_RSTART_SETUP_TIME(setup)
	i2c.Bus.SetSCL_STOP_SETUP_TIME(setup)
	i2c.Bus.SetSCL_START_HOLD_TIME(hold - 1)
	i2c.Bus.SetSCL_STOP_HOLD_TIME(hold - 1)
	i2c.Bus.SetSDA_SAMPLE_TIME(sda_sample)
	i2c.Bus.SetSDA_HOLD_TIME(sdaHold)
	// set timeout value
	i2c.Bus.SetTO_TIME_OUT(20 * halfCycle)
}

//go:inline
func (i2c *I2C) startMaster() {
	// FIFO mode for data
	i2c.Bus.SetFIFO_CONF_NONFIFO_EN(0)
	// Reset TX & RX buffers
	i2c.Bus.SetFIFO_CONF_RX_FIFO_RST(1)
	i2c.Bus.SetFIFO_CONF_RX_FIFO_RST(0)
	i2c.Bus.SetFIFO_CONF_TX_FIFO_RST(1)
	i2c.Bus.SetFIFO_CONF_TX_FIFO_RST(0)
	// enable master mode
	i2c.Bus.SetCTR_MS_MODE(1)
}

func (i2c *I2C) resetBus() {
	// unlike esp32c3, the esp32 i2c modules do not have a reset fsm register,
	// so we need to:
	//   1. disconnect the pins
	//   2. generate a stop condition manually
	//   3. do a full reset
	//   4. redo all configuration

	i2c.config.SDA.mux().Set(2<<esp.IO_MUX_GPIO0_MCU_SEL_Pos | esp.IO_MUX_GPIO0_FUN_IE | 1<<esp.IO_MUX_GPIO0_FUN_DRV_Pos)
	i2c.config.SDA.outFunc().Set(0x500)
	i2c.config.SDA.pinReg().SetBits(esp.GPIO_PIN_PAD_DRIVER)
	i2c.config.SCL.mux().Set(2<<esp.IO_MUX_GPIO0_MCU_SEL_Pos | esp.IO_MUX_GPIO0_FUN_IE | 1<<esp.IO_MUX_GPIO0_FUN_DRV_Pos)
	i2c.config.SCL.outFunc().Set(0x500)
	i2c.config.SCL.pinReg().SetBits(esp.GPIO_PIN_PAD_DRIVER)

	// bit-bang a read-NACK in case any device on the bus is in the middle of a write
	i2c.config.SCL.Low()
	i2c.config.SDA.High()
	wait()
	for i := 0; i < 9; i++ {
		if i2c.config.SDA.Get() {
			break
		}
		i2c.config.SCL.High()
		wait()
		i2c.config.SCL.Low()
		wait()
	}
	i2c.config.SDA.Low()
	i2c.config.SCL.High()
	wait()
	i2c.config.SDA.High()

	// initAll contains initClock which contains a reset
	i2c.initAll()
}

func wait() {
	end := nanotime() + 5_000
	for nanotime() < end {
		//spin
	}
}

type i2cCommandType = uint32
type i2cAck = uint32

const (
	i2cCMD_RSTART   i2cCommandType = 0 << 11
	i2cCMD_WRITE    i2cCommandType = 1<<11 | 1<<8 // WRITE + ack_check_en
	i2cCMD_READ     i2cCommandType = 2 << 11
	i2cCMD_READLAST i2cCommandType = 2<<11 | 1<<10 // READ + NACK
	i2cCMD_STOP     i2cCommandType = 3 << 11
	i2cCMD_END      i2cCommandType = 4 << 11
)

type i2cCommand struct {
	cmd  i2cCommandType
	data []byte
	head int
}

//go:linkname nanotime runtime.nanotime
func nanotime() int64

func (i2c *I2C) transmit(addr uint16, cmd []i2cCommand, timeoutMS int) error {
	if i2c.Bus.GetSR_BUS_BUSY() == 1 {
		i2c.resetBus()
	}

	const intMask = esp.I2C_INT_STATUS_END_DETECT_INT_ST_Msk | esp.I2C_INT_STATUS_TRANS_COMPLETE_INT_ST_Msk | esp.I2C_INT_STATUS_TIME_OUT_INT_ST_Msk | esp.I2C_INT_STATUS_ACK_ERR_INT_ST_Msk | esp.I2C_INT_STATUS_ARBITRATION_LOST_INT_ST_Msk
	i2c.Bus.INT_CLR.Set(intMask)
	i2c.Bus.INT_ENA.Set(intMask)

	defer func() {
		i2c.Bus.INT_CLR.Set(intMask)
		i2c.Bus.INT_ENA.Set(0)
	}()

	timeoutNS := int64(timeoutMS) * 1000000
	needAddress := true
	needRestart := false
	readLast := false
	var readTo []byte
	for cmdIdx, reg := 0, &i2c.Bus.COMD0; cmdIdx < len(cmd); {
		c := &cmd[cmdIdx]

		switch c.cmd {
		case i2cCMD_RSTART:
			reg.Set(i2cCMD_RSTART)
			reg = nextAddress(reg)
			cmdIdx++

		case i2cCMD_WRITE:
			count := 32
			if needAddress {
				needAddress = false
				i2c.Bus.SetDATA_FIFO_RDATA((uint32(addr) & 0x7f) << 1)
				count--
				i2c.Bus.SLAVE_ADDR.Set(uint32(addr))
			}
			for ; count > 0 && c.head < len(c.data); count, c.head = count-1, c.head+1 {
				i2c.Bus.SetDATA_FIFO_RDATA(uint32(c.data[c.head]))
			}
			reg.Set(i2cCMD_WRITE | uint32(32-count))
			reg = nextAddress(reg)

			if c.head < len(c.data) {
				reg.Set(i2cCMD_END)
				reg = nil
			} else {
				cmdIdx++
			}
			needRestart = true

		case i2cCMD_READ:
			if needAddress {
				needAddress = false
				i2c.Bus.SetDATA_FIFO_RDATA((uint32(addr)&0x7f)<<1 | 1)
				i2c.Bus.SLAVE_ADDR.Set(uint32(addr))
				reg.Set(i2cCMD_WRITE | 1)
				reg = nextAddress(reg)
			}
			if needRestart {
				// We need to send RESTART again after i2cCMD_WRITE.
				reg.Set(i2cCMD_RSTART)

				reg = nextAddress(reg)
				reg.Set(i2cCMD_WRITE | 1)

				reg = nextAddress(reg)
				i2c.Bus.SetDATA_FIFO_RDATA((uint32(addr)&0x7f)<<1 | 1)
				needRestart = false
			}
			count := 32
			bytes := len(c.data) - c.head
			// Only last byte in sequence must be sent with ACK set to 1 to indicate end of data.
			split := bytes <= count
			if split {
				bytes--
			}
			if bytes > 32 {
				bytes = 32
			}
			if bytes > 0 {
				reg.Set(i2cCMD_READ | uint32(bytes))
				reg = nextAddress(reg)
			}

			if split {
				readLast = true
				reg.Set(i2cCMD_READLAST | 1)
				reg = nextAddress(reg)
				readTo = c.data[c.head : c.head+bytes+1] // read bytes + 1 last byte
				cmdIdx++
			} else {
				reg.Set(i2cCMD_END)
				readTo = c.data[c.head : c.head+bytes]
				reg = nil
			}

		case i2cCMD_STOP:
			reg.Set(i2cCMD_STOP)
			reg = nil
			cmdIdx++
		}
		if reg == nil {
			// transmit now
			i2c.Bus.SetCTR_TRANS_START(1)
			end := nanotime() + timeoutNS
			var mask uint32
			for mask = i2c.Bus.INT_STATUS.Get(); mask&intMask == 0; mask = i2c.Bus.INT_STATUS.Get() {
				if nanotime() > end {
					// timeout leaves the bus in an undefined state, reset
					i2c.resetBus()
					if readTo != nil {
						return errI2CReadTimeout
					}
					return errI2CWriteTimeout
				}
			}
			switch {
			case mask&esp.I2C_INT_STATUS_ACK_ERR_INT_ST_Msk != 0 && !readLast:
				return errI2CAckExpected
			case mask&esp.I2C_INT_STATUS_TIME_OUT_INT_ST_Msk != 0:
				// timeout leaves the bus in an undefined state, reset
				i2c.resetBus()
				if readTo != nil {
					return errI2CReadTimeout
				}
				return errI2CWriteTimeout
			}
			i2c.Bus.INT_CLR.SetBits(intMask)
			for i := 0; i < len(readTo); i++ {
				readTo[i] = byte(i2c.Bus.GetDATA_FIFO_RDATA() & 0xff)
				c.head++
			}
			readTo = nil
			reg = &i2c.Bus.COMD0
		}
	}
	return nil
}

// Tx does a single I2C transaction at the specified address.
// It clocks out the given address, writes the bytes in w, reads back len(r)
// bytes and stores them in r, and generates a stop condition on the bus.
func (i2c *I2C) Tx(addr uint16, w, r []byte) (err error) {
	// timeout in microseconds.
	const timeout = 40 // 40ms is a reasonable time for a real-time system.

	cmd := make([]i2cCommand, 0, 8)
	cmd = append(cmd, i2cCommand{cmd: i2cCMD_RSTART})
	if len(w) > 0 {
		cmd = append(cmd, i2cCommand{cmd: i2cCMD_WRITE, data: w})
	}
	if len(r) > 0 {
		cmd = append(cmd, i2cCommand{cmd: i2cCMD_READ, data: r})
	}
	cmd = append(cmd, i2cCommand{cmd: i2cCMD_STOP})

	return i2c.transmit(addr, cmd, timeout)
}

func (i2c *I2C) SetBaudRate(br uint32) error {
	return errI2CNotImplemented
}

func (p Pin) pinReg() *volatile.Register32 {
	return (*volatile.Register32)(unsafe.Pointer((uintptr(unsafe.Pointer(&esp.GPIO.PIN0)) + uintptr(p)*4)))
}

func nextAddress(reg *volatile.Register32) *volatile.Register32 {
	return (*volatile.Register32)(unsafe.Add(unsafe.Pointer(reg), 4))
}

// CheckDevice does an empty I2C transaction at the specified address.
// This can be used to find out if any device with that address is
// connected, e.g. for enumerating all devices on the bus.
func (i2c *I2C) CheckDevice(addr uint16) bool {
	// timeout in microseconds.
	const timeout = 40 // 40ms is a reasonable time for a real-time system.

	cmd := []i2cCommand{
		{cmd: i2cCMD_RSTART},
		{cmd: i2cCMD_WRITE},
		{cmd: i2cCMD_STOP},
	}
	return i2c.transmit(addr, cmd, timeout) == nil
}
