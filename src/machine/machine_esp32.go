//go:build esp32

package machine

import (
	"device/esp"
	"errors"
	"runtime/volatile"
	"unsafe"
)

const deviceName = esp.Device

const peripheralClock = 80000000 // 80MHz

// CPUFrequency returns the current CPU frequency of the chip.
// Currently it is a fixed frequency but it may allow changing in the future.
func CPUFrequency() uint32 {
	return 160e6 // 160MHz
}

var (
	ErrInvalidSPIBus = errors.New("machine: invalid SPI bus")
)

const (
	PinOutput PinMode = iota
	PinInput
	PinInputPullup
	PinInputPulldown
)

// Hardware pin numbers
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
	GPIO21 Pin = 21
	GPIO22 Pin = 22
	GPIO23 Pin = 23
	GPIO25 Pin = 25
	GPIO26 Pin = 26
	GPIO27 Pin = 27
	GPIO32 Pin = 32
	GPIO33 Pin = 33
	GPIO34 Pin = 34
	GPIO35 Pin = 35
	GPIO36 Pin = 36
	GPIO37 Pin = 37
	GPIO38 Pin = 38
	GPIO39 Pin = 39
)

// Configure this pin with the given configuration.
func (p Pin) Configure(config PinConfig) {
	// Output function 256 is a special value reserved for use as a regular GPIO
	// pin. Peripherals (SPI etc) can set a custom output function by calling
	// lowercase configure() instead with a signal name.
	p.configure(config, 256)
}

// configure is the same as Configure, but allows for setting a specific input
// or output signal.
// Signals are always routed through the GPIO matrix for simplicity. Output
// signals are configured in FUNCx_OUT_SEL_CFG which selects a particular signal
// to output on a given pin. Input signals are configured in FUNCy_IN_SEL_CFG,
// which sets the pin to use for a particular input signal.
func (p Pin) configure(config PinConfig, signal uint32) {
	if p == NoPin {
		// This simplifies pin configuration in peripherals such as SPI.
		return
	}

	var muxConfig uint32 // The mux configuration.

	// Configure this pin as a GPIO pin.
	const function = 3 // function 3 is GPIO for every pin
	muxConfig |= (function - 1) << esp.IO_MUX_GPIO0_MCU_SEL_Pos

	// Make this pin an input pin (always).
	muxConfig |= esp.IO_MUX_GPIO0_FUN_IE

	// Set drive strength: 0 is lowest, 3 is highest.
	muxConfig |= 2 << esp.IO_MUX_GPIO0_FUN_DRV_Pos

	// Select pull mode.
	if config.Mode == PinInputPullup {
		muxConfig |= esp.IO_MUX_GPIO0_FUN_WPU
	} else if config.Mode == PinInputPulldown {
		muxConfig |= esp.IO_MUX_GPIO0_FUN_WPD
	}

	// Configure the pad with the given IO mux configuration.
	p.mux().Set(muxConfig)

	switch config.Mode {
	case PinOutput:
		// Set the 'output enable' bit.
		if p < 32 {
			esp.GPIO.ENABLE_W1TS.Set(1 << p)
		} else {
			esp.GPIO.ENABLE1_W1TS.Set(1 << (p - 32))
		}
		// Set the signal to read the output value from. It can be a peripheral
		// output signal, or the special value 256 which indicates regular GPIO
		// usage.
		p.outFunc().Set(signal)
	case PinInput, PinInputPullup, PinInputPulldown:
		// Clear the 'output enable' bit.
		if p < 32 {
			esp.GPIO.ENABLE_W1TC.Set(1 << p)
		} else {
			esp.GPIO.ENABLE1_W1TC.Set(1 << (p - 32))
		}
		if signal != 256 {
			// Signal is a peripheral function (not a simple GPIO). Connect this
			// signal to the pin.
			// Note that outFunc and inFunc work in the opposite direction.
			// outFunc configures a pin to use a given output signal, while
			// inFunc specifies a pin to use to read the signal from.
			inFunc(signal).Set(esp.GPIO_FUNC_IN_SEL_CFG_SEL | uint32(p)<<esp.GPIO_FUNC_IN_SEL_CFG_IN_SEL_Pos)
		}
	}
}

// outFunc returns the FUNCx_OUT_SEL_CFG register used for configuring the
// output function selection.
func (p Pin) outFunc() *volatile.Register32 {
	return (*volatile.Register32)(unsafe.Add(unsafe.Pointer(&esp.GPIO.FUNC0_OUT_SEL_CFG), uintptr(p)*4))
}

// inFunc returns the FUNCy_IN_SEL_CFG register used for configuring the input
// function selection.
func inFunc(signal uint32) *volatile.Register32 {
	return (*volatile.Register32)(unsafe.Add(unsafe.Pointer(&esp.GPIO.FUNC0_IN_SEL_CFG), uintptr(signal)*4))
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
	if p < 32 {
		return &esp.GPIO.OUT_W1TS, 1 << p
	} else {
		return &esp.GPIO.OUT1_W1TS, 1 << (p - 32)
	}
}

func (p Pin) portMaskClear() (*volatile.Register32, uint32) {
	if p < 32 {
		return &esp.GPIO.OUT_W1TC, 1 << p
	} else {
		return &esp.GPIO.OUT1_W1TC, 1 << (p - 32)
	}
}

// Get returns the current value of a GPIO pin when the pin is configured as an
// input or as an output.
func (p Pin) Get() bool {
	if p < 32 {
		return esp.GPIO.IN.Get()&(1<<p) != 0
	} else {
		return esp.GPIO.IN1.Get()&(1<<(p-32)) != 0
	}
}

// mux returns the I/O mux configuration register corresponding to the given
// GPIO pin.
func (p Pin) mux() *volatile.Register32 {
	// I have no idea whether there is any pattern in the GPIO <-> pad mapping.
	// I couldn't find it.
	switch p {
	case 36:
		return &esp.IO_MUX.GPIO36
	case 37:
		return &esp.IO_MUX.GPIO37
	case 38:
		return &esp.IO_MUX.GPIO38
	case 39:
		return &esp.IO_MUX.GPIO39
	case 34:
		return &esp.IO_MUX.GPIO34
	case 35:
		return &esp.IO_MUX.GPIO35
	case 32:
		return &esp.IO_MUX.GPIO32
	case 33:
		return &esp.IO_MUX.GPIO33
	case 25:
		return &esp.IO_MUX.GPIO25
	case 26:
		return &esp.IO_MUX.GPIO26
	case 27:
		return &esp.IO_MUX.GPIO27
	case 14:
		return &esp.IO_MUX.MTMS
	case 12:
		return &esp.IO_MUX.MTDI
	case 13:
		return &esp.IO_MUX.MTCK
	case 15:
		return &esp.IO_MUX.MTDO
	case 2:
		return &esp.IO_MUX.GPIO2
	case 0:
		return &esp.IO_MUX.GPIO0
	case 4:
		return &esp.IO_MUX.GPIO4
	case 16:
		return &esp.IO_MUX.GPIO16
	case 17:
		return &esp.IO_MUX.GPIO17
	case 9:
		return &esp.IO_MUX.SD_DATA2
	case 10:
		return &esp.IO_MUX.SD_DATA3
	case 11:
		return &esp.IO_MUX.SD_CMD
	case 6:
		return &esp.IO_MUX.SD_CLK
	case 7:
		return &esp.IO_MUX.SD_DATA0
	case 8:
		return &esp.IO_MUX.SD_DATA1
	case 5:
		return &esp.IO_MUX.GPIO5
	case 18:
		return &esp.IO_MUX.GPIO18
	case 19:
		return &esp.IO_MUX.GPIO19
	case 20:
		return &esp.IO_MUX.GPIO20
	case 21:
		return &esp.IO_MUX.GPIO21
	case 22:
		return &esp.IO_MUX.GPIO22
	case 3:
		return &esp.IO_MUX.U0RXD
	case 1:
		return &esp.IO_MUX.U0TXD
	case 23:
		return &esp.IO_MUX.GPIO23
	case 24:
		return &esp.IO_MUX.GPIO24
	default:
		return nil
	}
}

var DefaultUART = UART0

var (
	UART0  = &_UART0
	_UART0 = UART{Bus: esp.UART0, Buffer: NewRingBuffer()}
	UART1  = &_UART1
	_UART1 = UART{Bus: esp.UART1, Buffer: NewRingBuffer()}
	UART2  = &_UART2
	_UART2 = UART{Bus: esp.UART2, Buffer: NewRingBuffer()}
)

type UART struct {
	Bus    *esp.UART_Type
	Buffer *RingBuffer
}

func (uart *UART) Configure(config UARTConfig) {
	if config.BaudRate == 0 {
		config.BaudRate = 115200
	}
	uart.Bus.CLKDIV.Set(peripheralClock / config.BaudRate)
}

func (uart *UART) writeByte(b byte) error {
	for (uart.Bus.STATUS.Get()>>16)&0xff >= 128 {
		// Read UART_TXFIFO_CNT from the status register, which indicates how
		// many bytes there are in the transmit buffer. Wait until there are
		// less than 128 bytes in this buffer (the default buffer size).
	}
	uart.Bus.TX_FIFO.Set(b)
	return nil
}

func (uart *UART) flush() {}

// Serial Peripheral Interface on the ESP32.
type SPI struct {
	Bus *esp.SPI_Type
}

var (
	// SPI0 and SPI1 are reserved for use by the caching system etc.
	SPI2 = SPI{esp.SPI2}
	SPI3 = SPI{esp.SPI3}
)

// SPIConfig configures a SPI peripheral on the ESP32. Make sure to set at least
// SCK, SDO and SDI (possibly to NoPin if not in use). The default for LSBFirst
// (false) and Mode (0) are good for most applications. The frequency defaults
// to 1MHz if not set but can be configured up to 40MHz. Possible values are
// 40MHz and integer divisions from 40MHz such as 20MHz, 13.3MHz, 10MHz, 8MHz,
// etc.
type SPIConfig struct {
	Frequency uint32
	SCK       Pin
	SDO       Pin
	SDI       Pin
	LSBFirst  bool
	Mode      uint8
}

// Configure and make the SPI peripheral ready to use.
func (spi SPI) Configure(config SPIConfig) error {
	if config.Frequency == 0 {
		config.Frequency = 4e6 // default to 4MHz
	}

	// Configure the SPI clock. This assumes a peripheral clock of 80MHz.
	var clockReg uint32
	if config.Frequency > 40e6 {
		// Don't use a prescaler, but directly connect to the APB clock. This
		// results in a SPI clock frequency of 40MHz.
		clockReg |= esp.SPI_CLOCK_CLK_EQU_SYSCLK
	} else {
		// Use a prescaler for frequencies below 40MHz. They will get rounded
		// down to the next possible frequency (20MHz, 13.3MHz, 10MHz, 8MHz,
		// 6.7MHz, 5.7MHz, 5MHz, etc).
		// This code is much simpler than how ESP-IDF configures the frequency,
		// but should be just as accurate. The only exception is for frequencies
		// below 4883Hz, which will need special support.
		if config.Frequency < 4883 {
			// The current lower limit is 4883Hz.
			// The hardware supports lower frequencies by setting the h and n
			// variables, but that's not yet implemented.
			config.Frequency = 4883
		}
		// The prescaler value is 40e6 / config.Frequency, but rounded up so
		// that the actual frequency is never higher than the frequency
		// requested in config.Frequency.
		var (
			pre uint32 = (40e6 + config.Frequency - 1) / config.Frequency
			n   uint32 = 2 // this value seems to equal the number of ticks per SPI clock tick
			h   uint32 = 1 // must be half of n according to the formula in the reference manual
			l   uint32 = n // must equal n according to the reference manual
		)
		clockReg |= (pre - 1) << esp.SPI_CLOCK_CLKDIV_PRE_Pos
		clockReg |= (n - 1) << esp.SPI_CLOCK_CLKCNT_N_Pos
		clockReg |= (h - 1) << esp.SPI_CLOCK_CLKCNT_H_Pos
		clockReg |= (l - 1) << esp.SPI_CLOCK_CLKCNT_L_Pos
	}
	spi.Bus.CLOCK.Set(clockReg)

	// SPI_CTRL_REG controls bit order.
	var ctrlReg uint32
	if config.LSBFirst {
		ctrlReg |= esp.SPI_CTRL_WR_BIT_ORDER
		ctrlReg |= esp.SPI_CTRL_RD_BIT_ORDER
	}
	spi.Bus.CTRL.Set(ctrlReg)

	// SPI_CTRL2_REG, SPI_USER_REG and SPI_PIN_REG control SPI clock polarity
	// (mode), among others.
	var ctrl2Reg, userReg, pinReg uint32
	// For mode configuration, see table 29 in the reference manual (page 128).
	switch config.Mode {
	case 0:
	case 1:
		userReg |= esp.SPI_USER_CK_OUT_EDGE
	case 2:
		userReg |= esp.SPI_USER_CK_OUT_EDGE
		pinReg |= esp.SPI_PIN_CK_IDLE_EDGE
	case 3:
		pinReg |= esp.SPI_PIN_CK_IDLE_EDGE
	}
	// Enable full-duplex communication.
	userReg |= esp.SPI_USER_DOUTDIN
	userReg |= esp.SPI_USER_USR_MOSI
	// Write values to registers.
	spi.Bus.CTRL2.Set(ctrl2Reg)
	spi.Bus.USER.Set(userReg)
	spi.Bus.PIN.Set(pinReg)

	// Configure pins.
	// TODO: use direct output if possible, if the configured pins match the
	// possible direct configurations (e.g. for SPI2, when SCK is pin 14 etc).
	if spi.Bus == esp.SPI2 {
		config.SCK.configure(PinConfig{Mode: PinOutput}, 8)  // HSPICLK
		config.SDI.configure(PinConfig{Mode: PinInput}, 9)   // HSPIQ
		config.SDO.configure(PinConfig{Mode: PinOutput}, 10) // HSPID
	} else if spi.Bus == esp.SPI3 {
		config.SCK.configure(PinConfig{Mode: PinOutput}, 63) // VSPICLK
		config.SDI.configure(PinConfig{Mode: PinInput}, 64)  // VSPIQ
		config.SDO.configure(PinConfig{Mode: PinOutput}, 65) // VSPID
	} else {
		// Don't know how to configure this bus.
		return ErrInvalidSPIBus
	}

	return nil
}

// Transfer writes/reads a single byte using the SPI interface. If you need to
// transfer larger amounts of data, Tx will be faster.
func (spi SPI) Transfer(w byte) (byte, error) {
	spi.Bus.MISO_DLEN.Set(7 << esp.SPI_MISO_DLEN_USR_MISO_DBITLEN_Pos)
	spi.Bus.MOSI_DLEN.Set(7 << esp.SPI_MOSI_DLEN_USR_MOSI_DBITLEN_Pos)

	spi.Bus.W0.Set(uint32(w))

	// Send/receive byte.
	spi.Bus.CMD.Set(esp.SPI_CMD_USR)
	for spi.Bus.CMD.Get() != 0 {
	}

	// The received byte is stored in W0.
	return byte(spi.Bus.W0.Get()), nil
}

// Tx handles read/write operation for SPI interface. Since SPI is a syncronous write/read
// interface, there must always be the same number of bytes written as bytes read.
// This is accomplished by sending zero bits if r is bigger than w or discarding
// the incoming data if w is bigger than r.
func (spi SPI) Tx(w, r []byte) error {
	toTransfer := len(w)
	if len(r) > toTransfer {
		toTransfer = len(r)
	}

	for toTransfer != 0 {
		// Do only 64 bytes at a time.
		chunkSize := toTransfer
		if chunkSize > 64 {
			chunkSize = 64
		}

		// Fill tx buffer.
		transferWords := (*[16]volatile.Register32)(unsafe.Pointer(uintptr(unsafe.Pointer(&spi.Bus.W0))))
		if len(w) >= 64 {
			// We can fill the entire 64-byte transfer buffer with data.
			// This loop is slightly faster than the loop below.
			for i := 0; i < 16; i++ {
				word := uint32(w[i*4])<<0 | uint32(w[i*4+1])<<8 | uint32(w[i*4+2])<<16 | uint32(w[i*4+3])<<24
				transferWords[i].Set(word)
			}
		} else {
			// We can't fill the entire transfer buffer, so we need to be a bit
			// more careful.
			// Note that parts of the transfer buffer that aren't used still
			// need to be set to zero, otherwise we might be transferring
			// garbage from a previous transmission if w is smaller than r.
			for i := 0; i < 16; i++ {
				var word uint32
				if i*4+3 < len(w) {
					word |= uint32(w[i*4+3]) << 24
				}
				if i*4+2 < len(w) {
					word |= uint32(w[i*4+2]) << 16
				}
				if i*4+1 < len(w) {
					word |= uint32(w[i*4+1]) << 8
				}
				if i*4+0 < len(w) {
					word |= uint32(w[i*4+0]) << 0
				}
				transferWords[i].Set(word)
			}
		}

		// Do the transfer.
		spi.Bus.MISO_DLEN.Set((uint32(chunkSize)*8 - 1) << esp.SPI_MISO_DLEN_USR_MISO_DBITLEN_Pos)
		spi.Bus.MOSI_DLEN.Set((uint32(chunkSize)*8 - 1) << esp.SPI_MOSI_DLEN_USR_MOSI_DBITLEN_Pos)
		spi.Bus.CMD.Set(esp.SPI_CMD_USR)
		for spi.Bus.CMD.Get() != 0 {
		}

		// Read rx buffer.
		rxSize := 64
		if rxSize > len(r) {
			rxSize = len(r)
		}
		for i := 0; i < rxSize; i++ {
			r[i] = byte(transferWords[i/4].Get() >> ((i % 4) * 8))
		}

		// Cut off some part of the output buffer so the next iteration we will
		// only send the remaining bytes.
		if len(w) < chunkSize {
			w = nil
		} else {
			w = w[chunkSize:]
		}
		if len(r) < chunkSize {
			r = nil
		} else {
			r = r[chunkSize:]
		}
		toTransfer -= chunkSize
	}

	return nil
}
