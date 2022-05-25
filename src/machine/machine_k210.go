//go:build k210
// +build k210

package machine

import (
	"device/kendryte"
	"device/riscv"
	"errors"
	"runtime/interrupt"
	"unsafe"
)

const deviceName = kendryte.Device

func CPUFrequency() uint32 {
	return 390000000
}

type fpioaPullMode uint8
type PinChange uint8

// Pin modes.
const (
	PinInput PinMode = iota
	PinInputPullUp
	PinInputPullDown
	PinOutput
)

// FPIOA internal pull resistors.
const (
	fpioaPullNone fpioaPullMode = iota
	fpioaPullDown
	fpioaPullUp
)

// GPIOHS pin interrupt events.
const (
	PinRising PinChange = 1 << iota
	PinFalling
	PinToggle = PinRising | PinFalling
)

var (
	errUnsupportedSPIController = errors.New("SPI controller not supported. Use SPI0 or SPI1.")
	errI2CTxAbort               = errors.New("I2C transmition has been aborted.")
)

func (p Pin) setFPIOAIOPull(pull fpioaPullMode) {
	switch pull {
	case fpioaPullNone:
		kendryte.FPIOA.IO[uint8(p)].ClearBits(kendryte.FPIOA_IO_PU & kendryte.FPIOA_IO_PD)
	case fpioaPullUp:
		kendryte.FPIOA.IO[uint8(p)].SetBits(kendryte.FPIOA_IO_PU)
		kendryte.FPIOA.IO[uint8(p)].ClearBits(kendryte.FPIOA_IO_PD)
	case fpioaPullDown:
		kendryte.FPIOA.IO[uint8(p)].ClearBits(kendryte.FPIOA_IO_PU)
		kendryte.FPIOA.IO[uint8(p)].SetBits(kendryte.FPIOA_IO_PD)
	}
}

// SetFPIOAFunction is used to configure the pin for one of the FPIOA functions.
// Each pin on the Kendryte K210 can be configured with any of the available FPIOA functions.
func (p Pin) SetFPIOAFunction(f FPIOAFunction) {
	kendryte.FPIOA.IO[uint8(p)].Set(fpioaFuncDefaults[uint8(f)])
}

// FPIOAFunction returns the current FPIOA function of the pin.
func (p Pin) FPIOAFunction() FPIOAFunction {
	return FPIOAFunction((kendryte.FPIOA.IO[uint8(p)].Get() & kendryte.FPIOA_IO_CH_SEL_Msk))
}

// Configure this pin with the given configuration.
// The pin must already be set as GPIO or GPIOHS pin.
func (p Pin) Configure(config PinConfig) {
	var input bool

	// Check if the current pin's FPIOA function is either GPIO or GPIOHS.
	f := p.FPIOAFunction()
	if f < FUNC_GPIOHS0 || f > FUNC_GPIO7 {
		return // The pin is not configured as GPIO or GPIOHS.
	}

	// Configure pin.
	kendryte.FPIOA.IO[uint8(p)].SetBits(kendryte.FPIOA_IO_OE_EN | kendryte.FPIOA_IO_IE_EN | kendryte.FPIOA_IO_ST | kendryte.FPIOA_IO_DS_Msk)

	switch config.Mode {
	case PinInput:
		p.setFPIOAIOPull(fpioaPullNone)
		input = true
	case PinInputPullUp:
		p.setFPIOAIOPull(fpioaPullUp)
		input = true
	case PinInputPullDown:
		p.setFPIOAIOPull(fpioaPullDown)
		input = true
	case PinOutput:
		p.setFPIOAIOPull(fpioaPullNone)
		input = false
	}

	if f >= FUNC_GPIO0 && f <= FUNC_GPIO7 {
		// Converts the IO pin number in the effective GPIO number (based on the FPIOA function).
		gpioPin := uint8(f - FUNC_GPIO0)

		if input {
			kendryte.GPIO.DIRECTION.ClearBits(1 << gpioPin)
		} else {
			kendryte.GPIO.DIRECTION.SetBits(1 << gpioPin)
		}
	} else if f >= FUNC_GPIOHS0 && f <= FUNC_GPIOHS31 {
		// Converts the IO pin number in the effective GPIOHS number (based on the FPIOA function).
		gpioPin := uint8(f - FUNC_GPIOHS0)

		if input {
			kendryte.GPIOHS.INPUT_EN.SetBits(1 << gpioPin)
			kendryte.GPIOHS.OUTPUT_EN.ClearBits(1 << gpioPin)
		} else {
			kendryte.GPIOHS.OUTPUT_EN.SetBits(1 << gpioPin)
			kendryte.GPIOHS.INPUT_EN.ClearBits(1 << gpioPin)
		}
	}
}

// Set the pin to high or low.
func (p Pin) Set(high bool) {

	// Check if the current pin's FPIOA function is either GPIO or GPIOHS.
	f := p.FPIOAFunction()
	if f < FUNC_GPIOHS0 || f > FUNC_GPIO7 {
		return // The pin is not configured as GPIO or GPIOHS.
	}

	if f >= FUNC_GPIO0 && f <= FUNC_GPIO7 {
		gpioPin := uint8(f - FUNC_GPIO0)

		if high {
			kendryte.GPIO.DATA_OUTPUT.SetBits(1 << gpioPin)
		} else {
			kendryte.GPIO.DATA_OUTPUT.ClearBits(1 << gpioPin)
		}
	} else if f >= FUNC_GPIOHS0 && f <= FUNC_GPIOHS31 {
		gpioPin := uint8(f - FUNC_GPIOHS0)

		if high {
			kendryte.GPIOHS.OUTPUT_VAL.SetBits(1 << gpioPin)
		} else {
			kendryte.GPIOHS.OUTPUT_VAL.ClearBits(1 << gpioPin)
		}
	}
}

// Get returns the current value of a GPIO pin.
func (p Pin) Get() bool {

	// Check if the current pin's FPIOA function is either GPIO or GPIOHS.
	f := p.FPIOAFunction()
	if f < FUNC_GPIOHS0 || f > FUNC_GPIO7 {
		return false // The pin is not configured as GPIO or GPIOHS.
	}

	var val uint32
	if f >= FUNC_GPIO0 && f <= FUNC_GPIO7 {
		gpioPin := uint8(f - FUNC_GPIO0)
		val = kendryte.GPIO.DATA_INPUT.Get() & (1 << gpioPin)
	} else if f >= FUNC_GPIOHS0 && f <= FUNC_GPIOHS31 {
		gpioPin := uint8(f - FUNC_GPIOHS0)
		val = kendryte.GPIOHS.INPUT_VAL.Get() & (1 << gpioPin)
	}
	return (val > 0)
}

// Callbacks to be called for GPIOHS pins configured with SetInterrupt.
var pinCallbacks [32]func(Pin)

// SetInterrupt sets an interrupt to be executed when a particular pin changes
// state. The pin should already be configured as an input, including a pull up
// or down if no external pull is provided.
//
// You can pass a nil func to unset the pin change interrupt. If you do so,
// the change parameter is ignored and can be set to any value (such as 0).
// If the pin is already configured with a callback, you must first unset
// this pins interrupt before you can set a new callback.
func (p Pin) SetInterrupt(change PinChange, callback func(Pin)) error {

	// Check if the pin is a GPIOHS pin.
	f := p.FPIOAFunction()
	if f < FUNC_GPIOHS0 || f > FUNC_GPIOHS31 {
		return ErrInvalidDataPin
	}

	gpioPin := uint8(f - FUNC_GPIOHS0)

	// Clear all interrupts.
	kendryte.GPIOHS.RISE_IE.ClearBits(1 << gpioPin)
	kendryte.GPIOHS.FALL_IE.ClearBits(1 << gpioPin)
	kendryte.GPIOHS.HIGH_IE.ClearBits(1 << gpioPin)
	kendryte.GPIOHS.LOW_IE.ClearBits(1 << gpioPin)

	// Clear all the pending bits for this pin.
	kendryte.GPIOHS.RISE_IP.SetBits(1 << gpioPin)
	kendryte.GPIOHS.FALL_IP.SetBits(1 << gpioPin)
	kendryte.GPIOHS.HIGH_IP.SetBits(1 << gpioPin)
	kendryte.GPIOHS.LOW_IP.SetBits(1 << gpioPin)

	if callback == nil {
		if pinCallbacks[gpioPin] != nil {
			pinCallbacks[gpioPin] = nil
		}
		return nil
	}

	if pinCallbacks[gpioPin] != nil {
		// The pin was already configured.
		// To properly re-configure a pin, unset it first and set a new
		// configuration.
		return ErrNoPinChangeChannel
	}

	pinCallbacks[gpioPin] = callback

	// Enable interrupts.
	if change&PinRising != 0 {
		kendryte.GPIOHS.RISE_IE.SetBits(1 << gpioPin)
	}
	if change&PinFalling != 0 {
		kendryte.GPIOHS.FALL_IE.SetBits(1 << gpioPin)
	}

	handleInterrupt := func(inter interrupt.Interrupt) {

		pin := uint8(inter.GetNumber() - kendryte.IRQ_GPIOHS0)

		if kendryte.GPIOHS.RISE_IE.HasBits(1 << pin) {
			kendryte.GPIOHS.RISE_IE.ClearBits(1 << pin)
			// Acknowledge interrupt atomically.
			riscv.AsmFull(
				"amoor.w {}, {mask}, ({reg})",
				map[string]interface{}{
					"mask": uint32(1 << pin),
					"reg":  uintptr(unsafe.Pointer(&kendryte.GPIOHS.RISE_IP.Reg)),
				})
			kendryte.GPIOHS.RISE_IE.SetBits(1 << pin)
		}

		if kendryte.GPIOHS.FALL_IE.HasBits(1 << pin) {
			kendryte.GPIOHS.FALL_IE.ClearBits(1 << pin)
			// Acknowledge interrupt atomically.
			riscv.AsmFull(
				"amoor.w {}, {mask}, ({reg})",
				map[string]interface{}{
					"mask": uint32(1 << pin),
					"reg":  uintptr(unsafe.Pointer(&kendryte.GPIOHS.FALL_IP.Reg)),
				})
			kendryte.GPIOHS.FALL_IE.SetBits(1 << pin)
		}

		pinCallbacks[pin](Pin(pin))
	}

	var ir interrupt.Interrupt

	switch f {
	case FUNC_GPIOHS0:
		ir = interrupt.New(kendryte.IRQ_GPIOHS0, handleInterrupt)
	case FUNC_GPIOHS1:
		ir = interrupt.New(kendryte.IRQ_GPIOHS1, handleInterrupt)
	case FUNC_GPIOHS2:
		ir = interrupt.New(kendryte.IRQ_GPIOHS2, handleInterrupt)
	case FUNC_GPIOHS3:
		ir = interrupt.New(kendryte.IRQ_GPIOHS3, handleInterrupt)
	case FUNC_GPIOHS4:
		ir = interrupt.New(kendryte.IRQ_GPIOHS4, handleInterrupt)
	case FUNC_GPIOHS5:
		ir = interrupt.New(kendryte.IRQ_GPIOHS5, handleInterrupt)
	case FUNC_GPIOHS6:
		ir = interrupt.New(kendryte.IRQ_GPIOHS6, handleInterrupt)
	case FUNC_GPIOHS7:
		ir = interrupt.New(kendryte.IRQ_GPIOHS7, handleInterrupt)
	case FUNC_GPIOHS8:
		ir = interrupt.New(kendryte.IRQ_GPIOHS8, handleInterrupt)
	case FUNC_GPIOHS9:
		ir = interrupt.New(kendryte.IRQ_GPIOHS9, handleInterrupt)
	case FUNC_GPIOHS10:
		ir = interrupt.New(kendryte.IRQ_GPIOHS10, handleInterrupt)
	case FUNC_GPIOHS11:
		ir = interrupt.New(kendryte.IRQ_GPIOHS11, handleInterrupt)
	case FUNC_GPIOHS12:
		ir = interrupt.New(kendryte.IRQ_GPIOHS12, handleInterrupt)
	case FUNC_GPIOHS13:
		ir = interrupt.New(kendryte.IRQ_GPIOHS13, handleInterrupt)
	case FUNC_GPIOHS14:
		ir = interrupt.New(kendryte.IRQ_GPIOHS14, handleInterrupt)
	case FUNC_GPIOHS15:
		ir = interrupt.New(kendryte.IRQ_GPIOHS15, handleInterrupt)
	case FUNC_GPIOHS16:
		ir = interrupt.New(kendryte.IRQ_GPIOHS16, handleInterrupt)
	case FUNC_GPIOHS17:
		ir = interrupt.New(kendryte.IRQ_GPIOHS17, handleInterrupt)
	case FUNC_GPIOHS18:
		ir = interrupt.New(kendryte.IRQ_GPIOHS18, handleInterrupt)
	case FUNC_GPIOHS19:
		ir = interrupt.New(kendryte.IRQ_GPIOHS19, handleInterrupt)
	case FUNC_GPIOHS20:
		ir = interrupt.New(kendryte.IRQ_GPIOHS20, handleInterrupt)
	case FUNC_GPIOHS21:
		ir = interrupt.New(kendryte.IRQ_GPIOHS21, handleInterrupt)
	case FUNC_GPIOHS22:
		ir = interrupt.New(kendryte.IRQ_GPIOHS22, handleInterrupt)
	case FUNC_GPIOHS23:
		ir = interrupt.New(kendryte.IRQ_GPIOHS23, handleInterrupt)
	case FUNC_GPIOHS24:
		ir = interrupt.New(kendryte.IRQ_GPIOHS24, handleInterrupt)
	case FUNC_GPIOHS25:
		ir = interrupt.New(kendryte.IRQ_GPIOHS25, handleInterrupt)
	case FUNC_GPIOHS26:
		ir = interrupt.New(kendryte.IRQ_GPIOHS26, handleInterrupt)
	case FUNC_GPIOHS27:
		ir = interrupt.New(kendryte.IRQ_GPIOHS27, handleInterrupt)
	case FUNC_GPIOHS28:
		ir = interrupt.New(kendryte.IRQ_GPIOHS28, handleInterrupt)
	case FUNC_GPIOHS29:
		ir = interrupt.New(kendryte.IRQ_GPIOHS29, handleInterrupt)
	case FUNC_GPIOHS30:
		ir = interrupt.New(kendryte.IRQ_GPIOHS30, handleInterrupt)
	case FUNC_GPIOHS31:
		ir = interrupt.New(kendryte.IRQ_GPIOHS31, handleInterrupt)
	}

	ir.SetPriority(5)
	ir.Enable()

	return nil

}

type UART struct {
	Bus    *kendryte.UARTHS_Type
	Buffer *RingBuffer
}

var (
	UART0  = &_UART0
	_UART0 = UART{Bus: kendryte.UARTHS, Buffer: NewRingBuffer()}
)

func (uart *UART) Configure(config UARTConfig) {

	// Use default baudrate  if not set.
	if config.BaudRate == 0 {
		config.BaudRate = 115200
	}

	// Use default pins if not set.
	if config.TX == 0 && config.RX == 0 {
		config.TX = UART_TX_PIN
		config.RX = UART_RX_PIN
	}

	config.TX.SetFPIOAFunction(FUNC_UARTHS_TX)
	config.RX.SetFPIOAFunction(FUNC_UARTHS_RX)

	div := CPUFrequency()/config.BaudRate - 1

	uart.Bus.DIV.Set(div)
	uart.Bus.TXCTRL.Set(kendryte.UARTHS_TXCTRL_TXEN)
	uart.Bus.RXCTRL.Set(kendryte.UARTHS_RXCTRL_RXEN)

	// Enable interrupts on receive.
	uart.Bus.IE.Set(kendryte.UARTHS_IE_RXWM)

	intr := interrupt.New(kendryte.IRQ_UARTHS, _UART0.handleInterrupt)
	intr.SetPriority(5)
	intr.Enable()
}

func (uart *UART) handleInterrupt(interrupt.Interrupt) {
	rxdata := uart.Bus.RXDATA.Get()
	c := byte(rxdata)
	if uint32(c) != rxdata {
		// The rxdata has other bits set than just the low 8 bits. This probably
		// means that the 'empty' flag is set, which indicates there is no data
		// to be read and the byte is garbage. Ignore this byte.
		return
	}
	uart.Receive(c)
}

func (uart *UART) WriteByte(c byte) {
	for uart.Bus.TXDATA.Get()&kendryte.UARTHS_TXDATA_FULL != 0 {
	}

	uart.Bus.TXDATA.Set(uint32(c))
}

type SPI struct {
	Bus *kendryte.SPI_Type
}

// SPIConfig is used to store config info for SPI.
type SPIConfig struct {
	Frequency uint32
	SCK       Pin
	SDO       Pin
	SDI       Pin
	LSBFirst  bool
	Mode      uint8
}

// Configure is intended to setup the SPI interface.
// Only SPI controller 0 and 1 can be used because SPI2 is a special
// peripheral-mode controller and SPI3 is used for flashing.
func (spi SPI) Configure(config SPIConfig) error {
	// Use default pins if not set.
	if config.SCK == 0 && config.SDO == 0 && config.SDI == 0 {
		config.SCK = SPI0_SCK_PIN
		config.SDO = SPI0_SDO_PIN
		config.SDI = SPI0_SDI_PIN
	}

	// Enable APB2 clock.
	kendryte.SYSCTL.CLK_EN_CENT.SetBits(kendryte.SYSCTL_CLK_EN_CENT_APB2_CLK_EN)

	switch spi.Bus {
	case kendryte.SPI0:
		// Initialize SPI clock.
		kendryte.SYSCTL.CLK_EN_PERI.SetBits(kendryte.SYSCTL_CLK_EN_PERI_SPI0_CLK_EN)
		kendryte.SYSCTL.CLK_TH1.ClearBits(kendryte.SYSCTL_CLK_TH1_SPI0_CLK_Msk)

		// Initialize pins.
		config.SCK.SetFPIOAFunction(FUNC_SPI0_SCLK)
		config.SDO.SetFPIOAFunction(FUNC_SPI0_D0)
		config.SDI.SetFPIOAFunction(FUNC_SPI0_D1)
	case kendryte.SPI1:
		// Initialize SPI clock.
		kendryte.SYSCTL.CLK_EN_PERI.SetBits(kendryte.SYSCTL_CLK_EN_PERI_SPI1_CLK_EN)
		kendryte.SYSCTL.CLK_TH1.ClearBits(kendryte.SYSCTL_CLK_TH1_SPI1_CLK_Msk)

		// Initialize pins.
		config.SCK.SetFPIOAFunction(FUNC_SPI1_SCLK)
		config.SDO.SetFPIOAFunction(FUNC_SPI1_D0)
		config.SDI.SetFPIOAFunction(FUNC_SPI1_D1)
	default:
		return errUnsupportedSPIController
	}

	// Set default frequency.
	if config.Frequency == 0 {
		config.Frequency = 4000000 // 4MHz
	}

	baudr := CPUFrequency() / config.Frequency
	spi.Bus.BAUDR.Set(baudr)

	// Configure SPI mode 0, standard frame format, 8-bit data, little-endian.
	spi.Bus.IMR.Set(0)
	spi.Bus.DMACR.Set(0)
	spi.Bus.DMATDLR.Set(0x10)
	spi.Bus.DMARDLR.Set(0)
	spi.Bus.SER.Set(0)
	spi.Bus.SSIENR.Set(0)
	spi.Bus.CTRLR0.Set((7 << 16))
	spi.Bus.SPI_CTRLR0.Set(0)
	spi.Bus.ENDIAN.Set(0)

	return nil
}

// Transfer writes/reads a single byte using the SPI interface.
func (spi SPI) Transfer(w byte) (byte, error) {
	spi.Bus.SSIENR.Set(0)

	// Set transfer-receive mode.
	spi.Bus.CTRLR0.ClearBits(0x3 << 8)

	// Enable/disable SPI.
	spi.Bus.SSIENR.Set(1)
	defer spi.Bus.SSIENR.Set(0)

	// Enable/disable device.
	spi.Bus.SER.Set(0x1)
	defer spi.Bus.SER.Set(0)

	spi.Bus.DR0.Set(uint32(w))

	// Wait for transfer.
	for spi.Bus.SR.Get()&0x05 != 0x04 {
	}

	// Wait for data.
	for spi.Bus.RXFLR.Get() == 0 {
	}

	return byte(spi.Bus.DR0.Get()), nil
}

// I2C on the K210.
type I2C struct {
	Bus kendryte.I2C_Type
}

var (
	I2C0 = (*I2C)(unsafe.Pointer(kendryte.I2C0))
	I2C1 = (*I2C)(unsafe.Pointer(kendryte.I2C1))
	I2C2 = (*I2C)(unsafe.Pointer(kendryte.I2C2))
)

// I2CConfig is used to store config info for I2C.
type I2CConfig struct {
	Frequency uint32
	SCL       Pin
	SDA       Pin
}

// Configure is intended to setup the I2C interface.
func (i2c *I2C) Configure(config I2CConfig) error {

	if config.Frequency == 0 {
		config.Frequency = TWI_FREQ_100KHZ
	}

	if config.SDA == 0 && config.SCL == 0 {
		config.SDA = I2C0_SDA_PIN
		config.SCL = I2C0_SCL_PIN
	}

	// Enable APB0 clock.
	kendryte.SYSCTL.CLK_EN_CENT.SetBits(kendryte.SYSCTL_CLK_EN_CENT_APB0_CLK_EN)

	switch &i2c.Bus {
	case kendryte.I2C0:
		// Initialize I2C0 clock.
		kendryte.SYSCTL.CLK_EN_PERI.SetBits(kendryte.SYSCTL_CLK_EN_PERI_I2C0_CLK_EN)
		kendryte.SYSCTL.CLK_TH5.ReplaceBits(0x03, kendryte.SYSCTL_CLK_TH5_I2C0_CLK_Msk, kendryte.SYSCTL_CLK_TH5_I2C0_CLK_Pos)

		// Initialize pins.
		config.SDA.SetFPIOAFunction(FUNC_I2C0_SDA)
		config.SCL.SetFPIOAFunction(FUNC_I2C0_SCLK)
	case kendryte.I2C1:
		// Initialize I2C1 clock.
		kendryte.SYSCTL.CLK_EN_PERI.SetBits(kendryte.SYSCTL_CLK_EN_PERI_I2C1_CLK_EN)
		kendryte.SYSCTL.CLK_TH5.ReplaceBits(0x03, kendryte.SYSCTL_CLK_TH5_I2C1_CLK_Msk, kendryte.SYSCTL_CLK_TH5_I2C1_CLK_Pos)

		// Initialize pins.
		config.SDA.SetFPIOAFunction(FUNC_I2C1_SDA)
		config.SCL.SetFPIOAFunction(FUNC_I2C1_SCLK)
	case kendryte.I2C2:
		// Initialize I2C2 clock.
		kendryte.SYSCTL.CLK_EN_PERI.SetBits(kendryte.SYSCTL_CLK_EN_PERI_I2C2_CLK_EN)
		kendryte.SYSCTL.CLK_TH5.ReplaceBits(0x03, kendryte.SYSCTL_CLK_TH5_I2C2_CLK_Msk, kendryte.SYSCTL_CLK_TH5_I2C2_CLK_Pos)

		// Initialize pins.
		config.SDA.SetFPIOAFunction(FUNC_I2C2_SDA)
		config.SCL.SetFPIOAFunction(FUNC_I2C2_SCLK)
	}

	div := CPUFrequency() / config.Frequency / 16

	// Disable controller before setting the prescale register.
	i2c.Bus.ENABLE.Set(0)

	i2c.Bus.CON.Set(0x63)

	// Set prescaler registers.
	i2c.Bus.SS_SCL_HCNT.Set(uint32(div))
	i2c.Bus.SS_SCL_LCNT.Set(uint32(div))

	i2c.Bus.INTR_MASK.Set(0)
	i2c.Bus.DMA_CR.Set(0x03)
	i2c.Bus.DMA_RDLR.Set(0)
	i2c.Bus.DMA_TDLR.Set(0x4)

	return nil
}

// Tx does a single I2C transaction at the specified address.
// It clocks out the given address, writes the bytes in w, reads back len(r)
// bytes and stores them in r, and generates a stop condition on the bus.
func (i2c *I2C) Tx(addr uint16, w, r []byte) error {
	// Set peripheral address.
	i2c.Bus.TAR.Set(uint32(addr))
	// Enable controller.
	i2c.Bus.ENABLE.Set(1)

	if len(w) != 0 {
		i2c.Bus.CLR_TX_ABRT.Set(i2c.Bus.CLR_TX_ABRT.Get())
		dataLen := uint32(len(w))
		di := 0

		for dataLen != 0 {
			fifoLen := 8 - i2c.Bus.TXFLR.Get()
			if dataLen < fifoLen {
				fifoLen = dataLen
			}

			for i := uint32(0); i < fifoLen; i++ {
				i2c.Bus.DATA_CMD.Set(uint32(w[di]))
				di += 1
			}
			if i2c.Bus.TX_ABRT_SOURCE.Get() != 0 {
				return errI2CTxAbort
			}
			dataLen -= fifoLen
		}

		// Wait for transmition to complete.
		for i2c.Bus.STATUS.HasBits(kendryte.I2C_STATUS_ACTIVITY) || !i2c.Bus.STATUS.HasBits(kendryte.I2C_STATUS_TFE) {
		}

		if i2c.Bus.TX_ABRT_SOURCE.Get() != 0 {
			return errI2CTxAbort
		}
	}
	if len(r) != 0 {
		dataLen := uint32(len(r))
		cmdLen := uint32(len(r))
		di := 0

		for dataLen != 0 || cmdLen != 0 {
			fifoLen := i2c.Bus.RXFLR.Get()
			if dataLen < fifoLen {
				fifoLen = dataLen
			}
			for i := uint32(0); i < fifoLen; i++ {
				r[di] = byte(i2c.Bus.DATA_CMD.Get())
				di += 1
			}
			dataLen -= fifoLen

			fifoLen = 8 - i2c.Bus.TXFLR.Get()
			if cmdLen < fifoLen {
				fifoLen = cmdLen
			}
			for i := uint32(0); i < fifoLen; i++ {
				i2c.Bus.DATA_CMD.Set(0x100)
			}
			if i2c.Bus.TX_ABRT_SOURCE.Get() != 0 {
				return errI2CTxAbort
			}
			cmdLen -= fifoLen
		}
	}

	return nil
}
