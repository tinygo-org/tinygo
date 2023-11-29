//go:build esp32c3 && !m5stamp_c3

package machine

import (
	"device/esp"
	"runtime/volatile"
	"unsafe"
)

var (
	I2C0 = &I2C{}
)

type I2C struct{}

// I2CConfig is used to store config info for I2C.
type I2CConfig struct {
	Frequency uint32 // in Hz
	SCL       Pin
	SDA       Pin
}

const (
	clkXTAL               = 0
	clkFOSC               = 1
	clkXTALFrequency      = uint32(40e6)
	clkFOSCFrequency      = uint32(17.5e6)
	i2cClkSourceFrequency = clkXTALFrequency
	i2cClkSource          = clkXTAL
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

	i2c.initClock(config)
	i2c.initNoiseFilter()
	i2c.initPins(config)
	i2c.initFrequency(config)
	i2c.startMaster()
	return nil
}

//go:inline
func (i2c *I2C) initClock(config I2CConfig) {
	// reset I2C clock
	esp.SYSTEM.SetPERIP_RST_EN0_EXT0_RST(1)
	esp.SYSTEM.SetPERIP_CLK_EN0_EXT0_CLK_EN(1)
	esp.SYSTEM.SetPERIP_RST_EN0_EXT0_RST(0)
	// disable interrupts
	esp.I2C.INT_ENA.ClearBits(0x3fff)
	esp.I2C.INT_CLR.ClearBits(0x3fff)

	esp.I2C.SetCLK_CONF_SCLK_SEL(i2cClkSource)
	esp.I2C.SetCLK_CONF_SCLK_ACTIVE(1)
	esp.I2C.SetCLK_CONF_SCLK_DIV_NUM(i2cClkSourceFrequency / (config.Frequency * 1024))
	esp.I2C.SetCTR_CLK_EN(1)
}

//go:inline
func (i2c *I2C) initNoiseFilter() {
	esp.I2C.FILTER_CFG.Set(0x377)
}

//go:inline
func (i2c *I2C) initPins(config I2CConfig) {
	var muxConfig uint32
	const function = 1 // function 1 is just GPIO

	// SDA
	muxConfig = function << esp.IO_MUX_GPIO_MCU_SEL_Pos
	// Make this pin an input pin (always).
	muxConfig |= esp.IO_MUX_GPIO_FUN_IE
	// Set drive strength: 0 is lowest, 3 is highest.
	muxConfig |= 1 << esp.IO_MUX_GPIO_FUN_DRV_Pos
	config.SDA.mux().Set(muxConfig)
	config.SDA.outFunc().Set(54)
	inFunc(54).Set(uint32(esp.GPIO_FUNC_IN_SEL_CFG_SIG_IN_SEL | config.SDA))
	config.SDA.Set(true)
	// Configure the pad with the given IO mux configuration.
	config.SDA.pinReg().SetBits(esp.GPIO_PIN_PIN_PAD_DRIVER)

	esp.GPIO.ENABLE.SetBits(1 << int(config.SDA))
	esp.I2C.SetCTR_SDA_FORCE_OUT(1)

	// SCL
	muxConfig = function << esp.IO_MUX_GPIO_MCU_SEL_Pos
	// Make this pin an input pin (always).
	muxConfig |= esp.IO_MUX_GPIO_FUN_IE
	// Set drive strength: 0 is lowest, 3 is highest.
	muxConfig |= 1 << esp.IO_MUX_GPIO_FUN_DRV_Pos
	config.SCL.mux().Set(muxConfig)
	config.SCL.outFunc().Set(53)
	inFunc(53).Set(uint32(config.SCL))
	config.SCL.Set(true)
	// Configure the pad with the given IO mux configuration.
	config.SCL.pinReg().SetBits(esp.GPIO_PIN_PIN_PAD_DRIVER)

	esp.GPIO.ENABLE.SetBits(1 << int(config.SCL))
	esp.I2C.SetCTR_SCL_FORCE_OUT(1)
}

//go:inline
func (i2c *I2C) initFrequency(config I2CConfig) {

	clkmDiv := i2cClkSourceFrequency/(config.Frequency*1024) + 1
	sclkFreq := i2cClkSourceFrequency / clkmDiv
	halfCycle := sclkFreq / config.Frequency / 2
	//SCL
	sclLow := halfCycle
	sclWaitHigh := uint32(0)
	if config.Frequency > 50000 {
		sclWaitHigh = halfCycle / 8 // compensate the time when freq > 50K
	}
	sclHigh := halfCycle - sclWaitHigh
	// SDA
	sdaHold := halfCycle / 4
	sda_sample := halfCycle / 2
	setup := halfCycle
	hold := halfCycle

	esp.I2C.SetSCL_LOW_PERIOD(sclLow - 1)
	esp.I2C.SetSCL_HIGH_PERIOD(sclHigh)
	esp.I2C.SetSCL_HIGH_PERIOD_SCL_WAIT_HIGH_PERIOD(25)
	esp.I2C.SetSCL_RSTART_SETUP_TIME(setup)
	esp.I2C.SetSCL_STOP_SETUP_TIME(setup)
	esp.I2C.SetSCL_START_HOLD_TIME(hold - 1)
	esp.I2C.SetSCL_STOP_HOLD_TIME(hold - 1)
	esp.I2C.SetSDA_SAMPLE_TIME(sda_sample)
	esp.I2C.SetSDA_HOLD_TIME(sdaHold)
}

//go:inline
func (i2c *I2C) startMaster() {
	// FIFO mode for data
	esp.I2C.SetFIFO_CONF_NONFIFO_EN(0)
	// Reset TX & RX buffers
	esp.I2C.SetFIFO_CONF_RX_FIFO_RST(1)
	esp.I2C.SetFIFO_CONF_RX_FIFO_RST(0)
	esp.I2C.SetFIFO_CONF_TX_FIFO_RST(1)
	esp.I2C.SetFIFO_CONF_TX_FIFO_RST(0)
	// set timeout value
	esp.I2C.TO.Set(0x10)
	// enable master mode
	esp.I2C.CTR.Set(0x113)
	esp.I2C.SetCTR_CONF_UPGATE(1)
	resetMaster()
}

//go:inline
func resetMaster() {
	// reset FSM
	esp.I2C.SetCTR_FSM_RST(1)
	// clear the bus
	esp.I2C.SetSCL_SP_CONF_SCL_RST_SLV_NUM(9)
	esp.I2C.SetSCL_SP_CONF_SCL_RST_SLV_EN(1)
	esp.I2C.SetSCL_STRETCH_CONF_SLAVE_SCL_STRETCH_EN(1)
	esp.I2C.SetCTR_CONF_UPGATE(1)
	esp.I2C.FILTER_CFG.Set(0x377)
	// wait for SCL_RST_SLV_EN
	for esp.I2C.GetSCL_SP_CONF_SCL_RST_SLV_EN() != 0 {
	}
	esp.I2C.SetSCL_SP_CONF_SCL_RST_SLV_NUM(0)
}

type i2cCommandType = uint32
type i2cAck = uint32

const (
	i2cCMD_RSTART   i2cCommandType = 6 << 11
	i2cCMD_WRITE    i2cCommandType = 1<<11 | 1<<8 // WRITE + ack_check_en
	i2cCMD_READ     i2cCommandType = 3<<11 | 1<<8 // READ + ack_check_en
	i2cCMD_READLAST i2cCommandType = 3<<11 | 5<<8 // READ + ack_check_en + NACK
	i2cCMD_STOP     i2cCommandType = 2 << 11
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
	const intMask = esp.I2C_INT_STATUS_END_DETECT_INT_ST_Msk | esp.I2C_INT_STATUS_TRANS_COMPLETE_INT_ST_Msk | esp.I2C_INT_STATUS_TIME_OUT_INT_ST_Msk | esp.I2C_INT_STATUS_NACK_INT_ST_Msk
	esp.I2C.INT_CLR.SetBits(intMask)
	esp.I2C.INT_ENA.SetBits(intMask)
	esp.I2C.SetCTR_CONF_UPGATE(1)

	defer func() {
		esp.I2C.INT_CLR.SetBits(intMask)
		esp.I2C.INT_ENA.ClearBits(intMask)
	}()

	timeoutNS := int64(timeoutMS) * 1000000
	needAddress := true
	needRestart := false
	readLast := false
	var readTo []byte
	for cmdIdx, reg := 0, &esp.I2C.COMD0; cmdIdx < len(cmd); {
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
				esp.I2C.SetFIFO_DATA_FIFO_RDATA((uint32(addr) & 0x7f) << 1)
				count--
				esp.I2C.SLAVE_ADDR.Set(uint32(addr))
				esp.I2C.SetCTR_CONF_UPGATE(1)
			}
			for ; count > 0 && c.head < len(c.data); count, c.head = count-1, c.head+1 {
				esp.I2C.SetFIFO_DATA_FIFO_RDATA(uint32(c.data[c.head]))
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
				esp.I2C.SetFIFO_DATA_FIFO_RDATA((uint32(addr)&0x7f)<<1 | 1)
				esp.I2C.SLAVE_ADDR.Set(uint32(addr))
				reg.Set(i2cCMD_WRITE | 1)
				reg = nextAddress(reg)
			}
			if needRestart {
				// We need to send RESTART again after i2cCMD_WRITE.
				reg.Set(i2cCMD_RSTART)

				reg = nextAddress(reg)
				reg.Set(i2cCMD_WRITE | 1)

				reg = nextAddress(reg)
				esp.I2C.SetFIFO_DATA_FIFO_RDATA((uint32(addr)&0x7f)<<1 | 1)
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
			reg.Set(i2cCMD_READ | uint32(bytes))
			reg = nextAddress(reg)

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
			esp.I2C.SetCTR_CONF_UPGATE(1)
			esp.I2C.SetCTR_TRANS_START(1)
			end := nanotime() + timeoutNS
			var mask uint32
			for mask = esp.I2C.INT_STATUS.Get(); mask&intMask == 0; mask = esp.I2C.INT_STATUS.Get() {
				if nanotime() > end {
					if readTo != nil {
						return errI2CReadTimeout
					}
					return errI2CWriteTimeout
				}
			}
			switch {
			case mask&esp.I2C_INT_STATUS_NACK_INT_ST_Msk != 0 && !readLast:
				return errI2CAckExpected
			case mask&esp.I2C_INT_STATUS_TIME_OUT_INT_ST_Msk != 0:
				if readTo != nil {
					return errI2CReadTimeout
				}
				return errI2CWriteTimeout
			}
			esp.I2C.INT_CLR.SetBits(intMask)
			for i := 0; i < len(readTo); i++ {
				readTo[i] = byte(esp.I2C.GetFIFO_DATA_FIFO_RDATA() & 0xff)
				c.head++
			}
			readTo = nil
			reg = &esp.I2C.COMD0
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
	return nil
}

func nextAddress(reg *volatile.Register32) *volatile.Register32 {
	return (*volatile.Register32)(unsafe.Add(unsafe.Pointer(reg), 4))
}
