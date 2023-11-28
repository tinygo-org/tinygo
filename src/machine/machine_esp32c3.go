//go:build esp32c3

package machine

import (
	"device/esp"
	"device/riscv"
	"errors"
	"runtime/interrupt"
	"runtime/volatile"
	"sync"
	"unsafe"
)

const deviceName = esp.Device
const maxPin = 22
const cpuInterruptFromPin = 6

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

const (
	GPIO0  Pin = 0
	GPIO1  Pin = 1
	GPIO2  Pin = 2
	GPIO3  Pin = 3
	GPIO4  Pin = 4
	GPIO5  Pin = 5
	GPIO6  Pin = 6
	GPIO7  Pin = 7
	GPIO8  Pin = 8
	GPIO9  Pin = 9
	GPIO10 Pin = 10
	GPIO11 Pin = 11
	GPIO12 Pin = 12
	GPIO13 Pin = 13
	GPIO14 Pin = 14
	GPIO15 Pin = 15
	GPIO16 Pin = 16
	GPIO17 Pin = 17
	GPIO18 Pin = 18
	GPIO19 Pin = 19
	GPIO20 Pin = 20
	GPIO21 Pin = 21
)

type PinChange uint8

// Pin change interrupt constants for SetInterrupt.
const (
	PinRising PinChange = iota + 1
	PinFalling
	PinToggle
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
	return (*volatile.Register32)(unsafe.Add(unsafe.Pointer(&esp.GPIO.FUNC0_OUT_SEL_CFG), uintptr(p)*4))
}

func (p Pin) pinReg() *volatile.Register32 {
	return (*volatile.Register32)(unsafe.Pointer((uintptr(unsafe.Pointer(&esp.GPIO.PIN0)) + uintptr(p)*4)))
}

// inFunc returns the FUNCy_IN_SEL_CFG register used for configuring the input
// function selection.
func inFunc(signal uint32) *volatile.Register32 {
	return (*volatile.Register32)(unsafe.Add(unsafe.Pointer(&esp.GPIO.FUNC0_IN_SEL_CFG), uintptr(signal)*4))
}

// mux returns the I/O mux configuration register corresponding to the given
// GPIO pin.
func (p Pin) mux() *volatile.Register32 {
	return (*volatile.Register32)(unsafe.Add(unsafe.Pointer(&esp.IO_MUX.GPIO0), uintptr(p)*4))
}

// pin returns the PIN register corresponding to the given GPIO pin.
func (p Pin) pin() *volatile.Register32 {
	return (*volatile.Register32)(unsafe.Add(unsafe.Pointer(&esp.GPIO.PIN0), uintptr(p)*4))
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

// SetInterrupt sets an interrupt to be executed when a particular pin changes
// state. The pin should already be configured as an input, including a pull up
// or down if no external pull is provided.
//
// You can pass a nil func to unset the pin change interrupt. If you do so,
// the change parameter is ignored and can be set to any value (such as 0).
// If the pin is already configured with a callback, you must first unset
// this pins interrupt before you can set a new callback.
func (p Pin) SetInterrupt(change PinChange, callback func(Pin)) (err error) {
	if p >= maxPin {
		return ErrInvalidInputPin
	}

	if callback == nil {
		// Disable this pin interrupt
		p.pin().ClearBits(esp.GPIO_PIN_PIN_INT_TYPE_Msk | esp.GPIO_PIN_PIN_INT_ENA_Msk)

		if pinCallbacks[p] != nil {
			pinCallbacks[p] = nil
		}
		return nil
	}

	if pinCallbacks[p] != nil {
		// The pin was already configured.
		// To properly re-configure a pin, unset it first and set a new
		// configuration.
		return ErrNoPinChangeChannel
	}
	pinCallbacks[p] = callback

	onceSetupPinInterrupt.Do(func() {
		err = setupPinInterrupt()
	})
	if err != nil {
		return err
	}

	p.pin().Set(
		(p.pin().Get() & ^uint32(esp.GPIO_PIN_PIN_INT_TYPE_Msk|esp.GPIO_PIN_PIN_INT_ENA_Msk)) |
			uint32(change)<<esp.GPIO_PIN_PIN_INT_TYPE_Pos | uint32(1)<<esp.GPIO_PIN_PIN_INT_ENA_Pos)

	return nil
}

var (
	pinCallbacks          [maxPin]func(Pin)
	onceSetupPinInterrupt sync.Once
)

func setupPinInterrupt() error {
	esp.INTERRUPT_CORE0.GPIO_INTERRUPT_PRO_MAP.Set(cpuInterruptFromPin)
	return interrupt.New(cpuInterruptFromPin, func(interrupt.Interrupt) {
		status := esp.GPIO.STATUS.Get()
		for i, mask := 0, uint32(1); i < maxPin; i, mask = i+1, mask<<1 {
			if (status&mask) != 0 && pinCallbacks[i] != nil {
				pinCallbacks[i](Pin(i))
			}
		}
		// clear interrupt bit
		esp.GPIO.STATUS_W1TC.SetBits(status)
	}).Enable()
}

var (
	DefaultUART = UART0

	UART0  = &_UART0
	_UART0 = UART{Bus: esp.UART0, Buffer: NewRingBuffer()}
	UART1  = &_UART1
	_UART1 = UART{Bus: esp.UART1, Buffer: NewRingBuffer()}

	onceUart            = sync.Once{}
	errSamePins         = errors.New("UART: invalid pin combination")
	errWrongUART        = errors.New("UART: unsupported UARTn")
	errWrongBitSize     = errors.New("UART: invalid data size")
	errWrongStopBitSize = errors.New("UART: invalid bit size")
)

type UART struct {
	Bus                  *esp.UART_Type
	Buffer               *RingBuffer
	ParityErrorDetected  bool // set when parity error detected
	DataErrorDetected    bool // set when data corruption detected
	DataOverflowDetected bool // set when data overflow detected in UART FIFO buffer or RingBuffer
}

const (
	defaultDataBits = 8
	defaultStopBit  = 1
	defaultParity   = ParityNone

	uartInterrupts = esp.UART_INT_ENA_RXFIFO_FULL_INT_ENA |
		esp.UART_INT_ENA_PARITY_ERR_INT_ENA |
		esp.UART_INT_ENA_FRM_ERR_INT_ENA |
		esp.UART_INT_ENA_RXFIFO_OVF_INT_ENA |
		esp.UART_INT_ENA_GLITCH_DET_INT_ENA

	pplClockFreq = 80e6
)

type registerSet struct {
	interruptMapReg  *volatile.Register32
	uartClockBitMask uint32
	gpioMatrixSignal uint32
}

func (uart *UART) Configure(config UARTConfig) error {
	if config.BaudRate == 0 {
		config.BaudRate = 115200
	}
	if config.TX == config.RX {
		return errSamePins
	}
	switch {
	case uart.Bus == esp.UART0:
		return uart.configure(config, registerSet{
			interruptMapReg:  &esp.INTERRUPT_CORE0.UART_INTR_MAP,
			uartClockBitMask: esp.SYSTEM_PERIP_CLK_EN0_UART_CLK_EN,
			gpioMatrixSignal: 6,
		})
	case uart.Bus == esp.UART1:
		return uart.configure(config, registerSet{
			interruptMapReg:  &esp.INTERRUPT_CORE0.UART1_INTR_MAP,
			uartClockBitMask: esp.SYSTEM_PERIP_CLK_EN0_UART1_CLK_EN,
			gpioMatrixSignal: 9,
		})
	}
	return errWrongUART
}

func (uart *UART) configure(config UARTConfig, regs registerSet) error {

	initUARTClock(uart.Bus, regs)

	// - disbale TX/RX clock to make sure the UART transmitter or receiver is not at work during configuration
	uart.Bus.SetCLK_CONF_TX_SCLK_EN(0)
	uart.Bus.SetCLK_CONF_RX_SCLK_EN(0)

	// Configure static registers (Ref: Configuring URATn Communication)

	// - default clock source: 1=APB_CLK, 2=FOSC_CLK, 3=XTAL_CLK
	uart.Bus.SetCLK_CONF_SCLK_SEL(1)
	// reset divisor of the divider via UART_SCLK_DIV_NUM, UART_SCLK_DIV_A, and UART_SCLK_DIV_B
	uart.Bus.SetCLK_CONF_SCLK_DIV_NUM(0)
	uart.Bus.SetCLK_CONF_SCLK_DIV_A(0)
	uart.Bus.SetCLK_CONF_SCLK_DIV_B(0)

	// - the baud rate
	uart.SetBaudRate(config.BaudRate)
	// - the data format
	uart.SetFormat(defaultDataBits, defaultStopBit, defaultParity)
	// - set UART mode
	uart.Bus.SetRS485_CONF_RS485_EN(0)
	uart.Bus.SetRS485_CONF_RS485TX_RX_EN(0)
	uart.Bus.SetRS485_CONF_RS485RXBY_TX_EN(0)
	uart.Bus.SetCONF0_IRDA_EN(0)
	// - disable hw-flow control
	uart.Bus.SetCONF0_TX_FLOW_EN(0)
	uart.Bus.SetCONF1_RX_FLOW_EN(0)

	// synchronize values into Core Clock
	uart.Bus.SetID_REG_UPDATE(1)

	uart.setupPins(config, regs)
	uart.configureInterrupt(regs.interruptMapReg)
	uart.enableTransmitter()
	uart.enableReceiver()

	// Start TX/RX
	uart.Bus.SetCLK_CONF_TX_SCLK_EN(1)
	uart.Bus.SetCLK_CONF_RX_SCLK_EN(1)
	return nil
}

func (uart *UART) SetFormat(dataBits, stopBits int, parity UARTParity) error {
	if dataBits < 5 {
		return errWrongBitSize
	}
	if stopBits > 1 {
		return errWrongStopBitSize
	}
	// - data length
	uart.Bus.SetCONF0_BIT_NUM(uint32(dataBits - 5))
	// - stop bit
	uart.Bus.SetCONF0_STOP_BIT_NUM(uint32(stopBits))
	// - parity check
	switch parity {
	case ParityNone:
		uart.Bus.SetCONF0_PARITY_EN(0)
	case ParityEven:
		uart.Bus.SetCONF0_PARITY_EN(1)
		uart.Bus.SetCONF0_PARITY(0)
	case ParityOdd:
		uart.Bus.SetCONF0_PARITY_EN(1)
		uart.Bus.SetCONF0_PARITY(1)
	}
	return nil
}

func initUARTClock(bus *esp.UART_Type, regs registerSet) {
	uartClock := &esp.SYSTEM.PERIP_CLK_EN0
	uartClockReset := &esp.SYSTEM.PERIP_RST_EN0

	// Initialize/reset URATn (Ref: Initializing URATn)
	// - enable the clock for UART RAM
	uartClock.SetBits(esp.SYSTEM_PERIP_CLK_EN0_UART_MEM_CLK_EN)
	// - enable APB_CLK for UARTn
	uartClock.SetBits(regs.uartClockBitMask)
	// - reset sequence
	uartClockReset.ClearBits(regs.uartClockBitMask)
	bus.SetCLK_CONF_RST_CORE(1)
	uartClockReset.SetBits(regs.uartClockBitMask)
	uartClockReset.ClearBits(regs.uartClockBitMask)
	bus.SetCLK_CONF_RST_CORE(0)
	// synchronize core register
	bus.SetID_REG_UPDATE(0)
	// enable RTC clock
	esp.RTC_CNTL.SetRTC_CLK_CONF_DIG_CLK8M_EN(1)
	// wait for Core Clock to ready for configuration
	for bus.GetID_REG_UPDATE() > 0 {
		riscv.Asm("nop")
	}
}

func (uart *UART) SetBaudRate(baudRate uint32) {
	// based on esp-idf
	max_div := uint32((1 << 12) - 1)
	sclk_div := (pplClockFreq + (max_div * baudRate) - 1) / (max_div * baudRate)
	clk_div := (pplClockFreq << 4) / (baudRate * sclk_div)
	uart.Bus.SetCLKDIV(clk_div >> 4)
	uart.Bus.SetCLKDIV_FRAG(clk_div & 0xf)
	uart.Bus.SetCLK_CONF_SCLK_DIV_NUM(sclk_div - 1)
}

func (uart *UART) setupPins(config UARTConfig, regs registerSet) {
	config.RX.Configure(PinConfig{Mode: PinInputPullup})
	config.TX.Configure(PinConfig{Mode: PinInputPullup})

	// link TX with GPIO signal X (technical reference manual 5.10) (this is not interrupt signal!)
	config.TX.outFunc().Set(regs.gpioMatrixSignal)
	// link RX with GPIO signal X and route signals via GPIO matrix (GPIO_SIGn_IN_SEL 0x40)
	inFunc(regs.gpioMatrixSignal).Set(esp.GPIO_FUNC_IN_SEL_CFG_SIG_IN_SEL | uint32(config.RX))
}

func (uart *UART) configureInterrupt(intrMapReg *volatile.Register32) { // Disable all UART interrupts
	// Disable all UART interrupts
	uart.Bus.INT_ENA.ClearBits(0x0ffff)

	intrMapReg.Set(7)
	onceUart.Do(func() {
		_ = interrupt.New(7, func(i interrupt.Interrupt) {
			UART0.serveInterrupt(0)
			UART1.serveInterrupt(1)
		}).Enable()
	})
}

func (uart *UART) serveInterrupt(num int) {
	// get interrupt status
	interrutFlag := uart.Bus.INT_ST.Get()
	if (interrutFlag & uartInterrupts) == 0 {
		return
	}

	// block UART interrupts while processing
	uart.Bus.INT_ENA.ClearBits(uartInterrupts)

	if interrutFlag&esp.UART_INT_ENA_RXFIFO_FULL_INT_ENA > 0 {
		for uart.Bus.GetSTATUS_RXFIFO_CNT() > 0 {
			b := uart.Bus.GetFIFO_RXFIFO_RD_BYTE()
			if !uart.Buffer.Put(byte(b & 0xff)) {
				uart.DataOverflowDetected = true
			}
		}
	}
	if interrutFlag&esp.UART_INT_ENA_PARITY_ERR_INT_ENA > 0 {
		uart.ParityErrorDetected = true
	}
	if 0 != interrutFlag&esp.UART_INT_ENA_FRM_ERR_INT_ENA {
		uart.DataErrorDetected = true
	}
	if 0 != interrutFlag&esp.UART_INT_ENA_RXFIFO_OVF_INT_ENA {
		uart.DataOverflowDetected = true
	}
	if 0 != interrutFlag&esp.UART_INT_ENA_GLITCH_DET_INT_ENA {
		uart.DataErrorDetected = true
	}

	// Clear the UART interrupt status
	uart.Bus.INT_CLR.SetBits(interrutFlag)
	uart.Bus.INT_CLR.ClearBits(interrutFlag)
	// Enable interrupts
	uart.Bus.INT_ENA.Set(uartInterrupts)
}

const uart_empty_thresh_default = 10

func (uart *UART) enableTransmitter() {
	uart.Bus.SetCONF0_TXFIFO_RST(1)
	uart.Bus.SetCONF0_TXFIFO_RST(0)
	// TXINFO empty threshold is when txfifo_empty_int interrupt produced after the amount of data in Tx-FIFO is less than this register value.
	uart.Bus.SetCONF1_TXFIFO_EMPTY_THRHD(uart_empty_thresh_default)
	// we are not using interrut on TX since write we are waiting for FIFO to have space.
	// uart.Bus.INT_ENA.SetBits(esp.UART_INT_ENA_TXFIFO_EMPTY_INT_ENA)
}

func (uart *UART) enableReceiver() {
	uart.Bus.SetCONF0_RXFIFO_RST(1)
	uart.Bus.SetCONF0_RXFIFO_RST(0)
	// using value 1 so that we can start populate ring buffer with data as we get it
	uart.Bus.SetCONF1_RXFIFO_FULL_THRHD(1)
	// enable interrupts for:
	uart.Bus.SetINT_ENA_RXFIFO_FULL_INT_ENA(1)
	uart.Bus.SetINT_ENA_FRM_ERR_INT_ENA(1)
	uart.Bus.SetINT_ENA_PARITY_ERR_INT_ENA(1)
	uart.Bus.SetINT_ENA_GLITCH_DET_INT_ENA(1)
	uart.Bus.SetINT_ENA_RXFIFO_OVF_INT_ENA(1)
}

func (uart *UART) writeByte(b byte) error {
	for (uart.Bus.STATUS.Get()&esp.UART_STATUS_TXFIFO_CNT_Msk)>>esp.UART_STATUS_TXFIFO_CNT_Pos >= 128 {
		// Read UART_TXFIFO_CNT from the status register, which indicates how
		// many bytes there are in the transmit buffer. Wait until there are
		// less than 128 bytes in this buffer (the default buffer size).
	}
	uart.Bus.FIFO.Set(uint32(b))
	return nil
}

func (uart *UART) flush() {}

type Serialer interface {
	WriteByte(c byte) error
	Write(data []byte) (n int, err error)
	Configure(config UARTConfig) error
	Buffered() int
	ReadByte() (byte, error)
	DTR() bool
	RTS() bool
}

// USB Serial/JTAG Controller
// See esp32-c3_technical_reference_manual_en.pdf
// pg. 736
type USB_DEVICE struct {
	Bus *esp.USB_DEVICE_Type
}

var (
	_USBCDC = &USB_DEVICE{
		Bus: esp.USB_DEVICE,
	}

	USBCDC Serialer = _USBCDC
)

var (
	errUSBWrongSize            = errors.New("USB: invalid write size")
	errUSBCouldNotWriteAllData = errors.New("USB: could not write all data")
	errUSBBufferEmpty          = errors.New("USB: read buffer empty")
)

func (usbdev *USB_DEVICE) Configure(config UARTConfig) error {
	return nil
}

func (usbdev *USB_DEVICE) WriteByte(c byte) error {
	if usbdev.Bus.GetEP1_CONF_SERIAL_IN_EP_DATA_FREE() == 0 {
		return errUSBCouldNotWriteAllData
	}

	usbdev.Bus.SetEP1_RDWR_BYTE(uint32(c))
	usbdev.flush()

	return nil
}

func (usbdev *USB_DEVICE) Write(data []byte) (n int, err error) {
	if len(data) == 0 || len(data) > 64 {
		return 0, errUSBWrongSize
	}

	for i, c := range data {
		if usbdev.Bus.GetEP1_CONF_SERIAL_IN_EP_DATA_FREE() == 0 {
			if i > 0 {
				usbdev.flush()
			}

			return i, errUSBCouldNotWriteAllData
		}
		usbdev.Bus.SetEP1_RDWR_BYTE(uint32(c))
	}

	usbdev.flush()
	return len(data), nil
}

func (usbdev *USB_DEVICE) Buffered() int {
	return int(usbdev.Bus.GetEP1_CONF_SERIAL_OUT_EP_DATA_AVAIL())
}

func (usbdev *USB_DEVICE) ReadByte() (byte, error) {
	if usbdev.Bus.GetEP1_CONF_SERIAL_OUT_EP_DATA_AVAIL() != 0 {
		return byte(usbdev.Bus.GetEP1_RDWR_BYTE()), nil
	}

	return 0, nil
}

func (usbdev *USB_DEVICE) DTR() bool {
	return false
}

func (usbdev *USB_DEVICE) RTS() bool {
	return false
}

func (usbdev *USB_DEVICE) flush() {
	usbdev.Bus.SetEP1_CONF_WR_DONE(1)
	for usbdev.Bus.GetEP1_CONF_SERIAL_IN_EP_DATA_FREE() == 0 {
	}
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
	const intMask = esp.I2C_INT_CLR_END_DETECT_INT_CLR_Msk | esp.I2C_INT_CLR_TRANS_COMPLETE_INT_CLR_Msk | esp.I2C_INT_STATUS_TIME_OUT_INT_ST_Msk | esp.I2C_INT_STATUS_NACK_INT_ST_Msk
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
	var readTo []byte
	for cmdIdx, reg := 0, &esp.I2C.COMD0; cmdIdx < len(cmd); {
		c := &cmd[cmdIdx]

		switch c.cmd {
		case i2cCMD_RSTART:
			reg.Set(i2cCMD_RSTART)
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
			reg.Set(i2cCMD_WRITE | uint32(32-count))
			reg = (*volatile.Register32)(unsafe.Pointer((uintptr(unsafe.Pointer(reg)) + 4)))

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
				reg = (*volatile.Register32)(unsafe.Pointer((uintptr(unsafe.Pointer(reg)) + 4)))
			}
			if needRestart {
				// We need to send RESTART again after i2cCMD_WRITE.
				reg.Set(i2cCMD_RSTART)
				reg = (*volatile.Register32)(unsafe.Pointer((uintptr(unsafe.Pointer(reg)) + 4)))
				reg.Set(i2cCMD_WRITE | 1)
				reg = (*volatile.Register32)(unsafe.Pointer((uintptr(unsafe.Pointer(reg)) + 4)))
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
			reg = (*volatile.Register32)(unsafe.Pointer((uintptr(unsafe.Pointer(reg)) + 4)))

			if split {
				reg.Set(i2cCMD_READLAST | 1)
				reg = (*volatile.Register32)(unsafe.Pointer((uintptr(unsafe.Pointer(reg)) + 4)))
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
