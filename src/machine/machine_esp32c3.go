//go:build esp32c3
// +build esp32c3

package machine

import (
	"device/esp"
	"runtime/volatile"
	"unsafe"
)

const deviceName = esp.Device

// CPUFrequency returns the current CPU frequency of the chip.
// Currently it is a fixed frequency but it may allow changing in the future.
func CPUFrequency() uint32 {
	return 160e6 // 160MHz
}

const (
	PinOutput PinMode = iota
	PinInput
	PinInputPullup
	PinInputPulldown
)

// Configure this pin with the given configuration.
func (p Pin) Configure(config PinConfig) {
	if p == NoPin {
		// This simplifies pin configuration in peripherals such as SPI.
		return
	}

	var muxConfig uint32

	// Configure this pin as a GPIO pin.
	const function = 1 // function 1 is GPIO for every pin
	muxConfig |= function << esp.IO_MUX_GPIO_MCU_SEL_Pos

	// Make this pin an input pin (always).
	muxConfig |= esp.IO_MUX_GPIO_FUN_IE

	// Set drive strength: 0 is lowest, 3 is highest.
	muxConfig |= 2 << esp.IO_MUX_GPIO_FUN_DRV_Pos

	// Select pull mode.
	if config.Mode == PinInputPullup {
		muxConfig |= esp.IO_MUX_GPIO_FUN_WPU
	} else if config.Mode == PinInputPulldown {
		muxConfig |= esp.IO_MUX_GPIO_FUN_WPD
	}

	// Configure the pad with the given IO mux configuration.
	p.mux().Set(muxConfig)

	// Set the output signal to the simple GPIO output.
	p.outFunc().Set(0x80)

	switch config.Mode {
	case PinOutput:
		// Set the 'output enable' bit.
		esp.GPIO.ENABLE_W1TS.Set(1 << p)
	case PinInput, PinInputPullup, PinInputPulldown:
		// Clear the 'output enable' bit.
		esp.GPIO.ENABLE_W1TC.Set(1 << p)
	}
}

// outFunc returns the FUNCx_OUT_SEL_CFG register used for configuring the
// output function selection.
func (p Pin) outFunc() *volatile.Register32 {
	return (*volatile.Register32)(unsafe.Pointer((uintptr(unsafe.Pointer(&esp.GPIO.FUNC0_OUT_SEL_CFG)) + uintptr(p)*4)))
}

func (p Pin) pinReg() *volatile.Register32 {
	return (*volatile.Register32)(unsafe.Pointer((uintptr(unsafe.Pointer(&esp.GPIO.PIN0)) + uintptr(p)*4)))
}

// inFunc returns the FUNCy_IN_SEL_CFG register used for configuring the input
// function selection.
func inFunc(signal uint32) *volatile.Register32 {
	return (*volatile.Register32)(unsafe.Pointer((uintptr(unsafe.Pointer(&esp.GPIO.FUNC0_IN_SEL_CFG)) + uintptr(signal)*4)))
}

// mux returns the I/O mux configuration register corresponding to the given
// GPIO pin.
func (p Pin) mux() *volatile.Register32 {
	return (*volatile.Register32)(unsafe.Pointer((uintptr(unsafe.Pointer(&esp.IO_MUX.GPIO0)) + uintptr(p)*4)))
}

// Set the pin to high or low.
// Warning: only use this on an output pin!
func (p Pin) Set(value bool) {
	if value {
		reg, mask := p.portMaskSet()
		reg.Set(mask)
	} else {
		reg, mask := p.portMaskClear()
		reg.Set(mask)
	}
}

// Get returns the current value of a GPIO pin when configured as an input or as
// an output.
func (p Pin) Get() bool {
	reg := &esp.GPIO.IN
	return (reg.Get()>>p)&1 > 0
}

// Return the register and mask to enable a given GPIO pin. This can be used to
// implement bit-banged drivers.
//
// Warning: only use this on an output pin!
func (p Pin) PortMaskSet() (*uint32, uint32) {
	reg, mask := p.portMaskSet()
	return &reg.Reg, mask
}

// Return the register and mask to disable a given GPIO pin. This can be used to
// implement bit-banged drivers.
//
// Warning: only use this on an output pin!
func (p Pin) PortMaskClear() (*uint32, uint32) {
	reg, mask := p.portMaskClear()
	return &reg.Reg, mask
}

func (p Pin) portMaskSet() (*volatile.Register32, uint32) {
	return &esp.GPIO.OUT_W1TS, 1 << p
}

func (p Pin) portMaskClear() (*volatile.Register32, uint32) {
	return &esp.GPIO.OUT_W1TC, 1 << p
}

var DefaultUART = UART0

var (
	UART0  = &_UART0
	_UART0 = UART{Bus: esp.UART0, Buffer: NewRingBuffer()}
	UART1  = &_UART1
	_UART1 = UART{Bus: esp.UART1, Buffer: NewRingBuffer()}
)

type UART struct {
	Bus    *esp.UART_Type
	Buffer *RingBuffer
}

func (uart *UART) WriteByte(b byte) error {
	for (uart.Bus.STATUS.Get()&esp.UART_STATUS_TXFIFO_CNT_Msk)>>esp.UART_STATUS_TXFIFO_CNT_Pos >= 128 {
		// Read UART_TXFIFO_CNT from the status register, which indicates how
		// many bytes there are in the transmit buffer. Wait until there are
		// less than 128 bytes in this buffer (the default buffer size).
	}
	uart.Bus.FIFO.Set(uint32(b))
	return nil
}

// I2C code

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
	i2c.initClock(config)
	i2c.initNoiseFilter()
	i2c.initPins(config)
	i2c.initFrequency(config)
	i2c.startMaster()
	return nil
}

//go: inline
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

//go: inline
func (i2c *I2C) initNoiseFilter() {
	esp.I2C.FILTER_CFG.Set(0x377)
}

//go: inline
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

//go: inline
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

//go: inline
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

//go: inline
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
	i2cCMD_RSTART i2cCommandType = 6 << 11
	i2cCMD_WRITE  i2cCommandType = 1<<11 | 1<<8 // WRITE + ack_check_en
	i2cCMD_READ   i2cCommandType = 3 << 11
	i2cCMD_STOP   i2cCommandType = 2 << 11
	i2cCMD_END    i2cCommandType = 4 << 11
)

const (
	i2cAckNone i2cAck = 0
	i2cAckHigh i2cAck = 0x00000700
	i2cAckLow  i2cAck = 0x00000100
)

type i2cCommand struct {
	cmd  i2cCommandType
	data []byte
	ack  i2cAck
	head int
}

//go:linkname nanotime runtime.nanotime
func nanotime() int64

func (i2c *I2C) transmit(addr uint16, cmd []i2cCommand, timeoutMS int) error {
	const intMask = esp.I2C_INT_CLR_END_DETECT_INT_CLR_Msk | esp.I2C_INT_CLR_TRANS_COMPLETE_INT_CLR_Msk | esp.I2C_INT_STATUS_TIME_OUT_INT_ST_Msk
	esp.I2C.INT_CLR.SetBits(intMask)
	esp.I2C.INT_ENA.SetBits(intMask)
	esp.I2C.SetCTR_CONF_UPGATE(1)

	defer func() {
		esp.I2C.INT_CLR.SetBits(intMask)
		esp.I2C.INT_ENA.ClearBits(intMask)
	}()

	timeoutNS := int64(timeoutMS) * 1000000
	needAddress := true
	var readTo []byte
	for cmdIdx, reg := 0, &esp.I2C.COMD0; cmdIdx < len(cmd); {
		c := &cmd[cmdIdx]

		switch c.cmd {
		case i2cCMD_RSTART:
			reg.Set(i2cCMD_RSTART | c.ack)
			reg = (*volatile.Register32)(unsafe.Pointer((uintptr(unsafe.Pointer(reg)) + 4)))
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
			reg.Set(i2cCMD_WRITE | uint32(32-count) | c.ack)
			reg = (*volatile.Register32)(unsafe.Pointer((uintptr(unsafe.Pointer(reg)) + 4)))

			if c.head < len(c.data) {
				reg.Set(i2cCMD_END)
				reg = nil
			} else {
				cmdIdx++
			}

		case i2cCMD_READ:
			if needAddress {
				needAddress = false
				esp.I2C.SetFIFO_DATA_FIFO_RDATA((uint32(addr)&0x7f)<<1 | 1)
				esp.I2C.SLAVE_ADDR.Set(uint32(addr))
				reg.Set(i2cCMD_WRITE | 1)
				reg = (*volatile.Register32)(unsafe.Pointer((uintptr(unsafe.Pointer(reg)) + 4)))
			}
			count := 32
			bytes := len(c.data) - c.head
			split := bytes <= count
			if split {
				bytes--
			}
			if bytes > 32 {
				bytes = 32
			}
			readTo = c.data[c.head : c.head+bytes]
			reg.Set(i2cCMD_READ | uint32(bytes) | c.ack)
			reg = (*volatile.Register32)(unsafe.Pointer((uintptr(unsafe.Pointer(reg)) + 4)))

			if split {
				reg.Set(i2cCMD_READ | 1 | c.ack)
				reg = (*volatile.Register32)(unsafe.Pointer((uintptr(unsafe.Pointer(reg)) + 4)))
				cmdIdx++
			} else {
				reg.Set(i2cCMD_END)
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
			if mask == esp.I2C_INT_STATUS_TIME_OUT_INT_ST_Msk {
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
	const timeout = 100 * 1000 // 40ms is a reasonable time for a real-time system.

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
