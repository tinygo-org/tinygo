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
		p.pin().ClearBits(esp.GPIO_PIN_INT_TYPE_Msk | esp.GPIO_PIN_INT_ENA_Msk)

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
		(p.pin().Get() & ^uint32(esp.GPIO_PIN_INT_TYPE_Msk|esp.GPIO_PIN_INT_ENA_Msk)) |
			uint32(change)<<esp.GPIO_PIN_INT_TYPE_Pos | uint32(1)<<esp.GPIO_PIN_INT_ENA_Pos)

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

	// - disable TX/RX clock to make sure the UART transmitter or receiver is not at work during configuration
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
	esp.RTC_CNTL.SetCLK_CONF_DIG_CLK8M_EN(1)
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
	inFunc(regs.gpioMatrixSignal).Set(esp.GPIO_FUNC_IN_SEL_CFG_SEL | uint32(config.RX))
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

// GetRNG returns 32-bit random numbers using the ESP32-C3 true random number generator,
// Random numbers are generated based on the thermal noise in the system and the
// asynchronous clock mismatch.
// For maximum entropy also make sure that the SAR_ADC is enabled.
// See esp32-c3_technical_reference_manual_en.pdf p.524
func GetRNG() (ret uint32, err error) {
	// ensure ADC clock is initialized
	initADCClock()

	// ensure fast RTC clock is enabled
	if esp.RTC_CNTL.GetCLK_CONF_DIG_CLK8M_EN() == 0 {
		esp.RTC_CNTL.SetCLK_CONF_DIG_CLK8M_EN(1)
	}

	return esp.APB_CTRL.GetRND_DATA(), nil
}

func initADCClock() {
	if esp.APB_SARADC.GetCLKM_CONF_CLK_EN() == 1 {
		return
	}

	// only support ADC_CTRL_CLK set to 1
	esp.APB_SARADC.SetCLKM_CONF_CLK_SEL(1)

	esp.APB_SARADC.SetCTRL_SARADC_SAR_CLK_GATED(1)

	esp.APB_SARADC.SetCLKM_CONF_CLKM_DIV_NUM(15)
	esp.APB_SARADC.SetCLKM_CONF_CLKM_DIV_B(1)
	esp.APB_SARADC.SetCLKM_CONF_CLKM_DIV_A(0)

	esp.APB_SARADC.SetCTRL_SARADC_SAR_CLK_DIV(1)
	esp.APB_SARADC.SetCLKM_CONF_CLK_EN(1)
}
