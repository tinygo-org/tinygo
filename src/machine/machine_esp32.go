// +build esp32

package machine

import (
	"device/esp"
	"errors"
	"runtime/volatile"
	"unsafe"
)

const peripheralClock = 80000000 // 80MHz

// CPUFrequency returns the current CPU frequency of the chip.
// Currently it is a fixed frequency but it may allow changing in the future.
func CPUFrequency() uint32 {
	return 160e6 // 160MHz
}

var (
	ErrInvalidSPIBus = errors.New("machine: invalid SPI bus")
)

type PinMode uint8

const (
	PinInput     PinMode = 0
	PinOutput            = (1 << 1)
	PinFunction          = (1 << 2)
	PinPullup            = (1 << 3)
	PinPulldown          = (1 << 4)
	PinOpenDrain         = (1 << 5)
	PinFunction0         = 0
	PinFunction1         = (1 << 6)
	PinFunction2         = (1 << 7)
	// PinFunction3         = (1 << 8)
	// PinFunction4         = (1 << 9)
	// PinSpecial = (1 << 13)
	PinInputPullup     = PinInput | PinPullup
	PinInputPulldown   = PinInput | PinPulldown
	PinOutputOpenDrain = PinOutput | PinOpenDrain
	PinInputFunction   = PinInput | PinFunction
	PinInputFunction0  = PinInput | PinFunction0
	PinInputFunction1  = PinInput | PinFunction1
	PinInputFunction2  = PinInput | PinFunction2
	// PinInputFunction3  = PinInput | PinFunction3
	// PinInputFunction4  = PinInput | PinFunction4
	PinOutputFunction  = PinOutput | PinFunction
	PinOutputFunction0 = PinOutput | PinFunction0
	PinOutputFunction1 = PinOutput | PinFunction1
	PinOutputFunction2 = PinOutput | PinFunction2
	// PinOutputFunction3 = PinOutput | PinFunction3
	// PinOutputFunction4 = PinOutput | PinFunction4
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
	return (*volatile.Register32)(unsafe.Pointer((uintptr(unsafe.Pointer(&esp.GPIO.FUNC0_OUT_SEL_CFG)) + uintptr(p)*4)))
}

// inFunc returns the FUNCy_IN_SEL_CFG register used for configuring the input
// function selection.
func inFunc(signal uint32) *volatile.Register32 {
	return (*volatile.Register32)(unsafe.Pointer((uintptr(unsafe.Pointer(&esp.GPIO.FUNC0_IN_SEL_CFG)) + uintptr(signal)*4)))
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
// input.
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

var (
	UART0 = UART{Bus: esp.UART0, Buffer: NewRingBuffer()}
	UART1 = UART{Bus: esp.UART1, Buffer: NewRingBuffer()}
	UART2 = UART{Bus: esp.UART2, Buffer: NewRingBuffer()}
)

type UART struct {
	Bus    *esp.UART_Type
	Buffer *RingBuffer
}

func (uart UART) Configure(config UARTConfig) {
	if config.BaudRate == 0 {
		config.BaudRate = 115200
	}
	uart.Bus.CLKDIV.Set(peripheralClock / config.BaudRate)
}

func (uart UART) WriteByte(b byte) error {
	for (uart.Bus.STATUS.Get()>>16)&0xff >= 128 {
		// Read UART_TXFIFO_CNT from the status register, which indicates how
		// many bytes there are in the transmit buffer. Wait until there are
		// less than 128 bytes in this buffer (the default buffer size).
	}
	uart.Bus.TX_FIFO.Set(b)
	return nil
}

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
		config.Frequency = 1e6 // default to 1MHz
	}

	// Configure the SPI clock. This assumes a peripheral clock of 80MHz.
	var clockReg uint32
	if config.Frequency >= 40e6 {
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
	var delayMode uint32
	switch config.Mode {
	case 0:
		delayMode = 2
	case 1:
		delayMode = 1
		userReg |= esp.SPI_USER_CK_OUT_EDGE
	case 2:
		delayMode = 1
		userReg |= esp.SPI_USER_CK_OUT_EDGE
		pinReg |= esp.SPI_PIN_CK_IDLE_EDGE
	case 3:
		delayMode = 2
		pinReg |= esp.SPI_PIN_CK_IDLE_EDGE
	}
	// Extra configuration necessary for correct data input at high frequencies.
	// This is only necessary when MISO goes through the GPIO matrix (which it
	// currently does).
	if config.Frequency >= 40e6 {
		// Delay mode must be set to 0 and SPI_USR_DUMMY_CYCLELEN should be set
		// to 0 (the default).
		userReg |= esp.SPI_USER_USR_DUMMY
	} else if config.Frequency >= 20e6 {
		// Nothing to do here, delay mode should be set to 0 according to the
		// datasheet.
	} else {
		// Follow the delay mode as given in table 29 on page 128 of the
		// reference manual.
		// Note that this is only specified for SPI frequency of 10MHz and
		// below (≤Fapb/8), so 13.3MHz appears to be left unspecified.
		ctrl2Reg |= delayMode << esp.SPI_CTRL2_MOSI_DELAY_MODE_Pos
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
//
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
		var outBuf [16]uint32
		txSize := 64
		if txSize > len(w) {
			txSize = len(w)
		}
		for i := 0; i < txSize; i++ {
			outBuf[i/4] = outBuf[i/4] | uint32(w[i])<<((i%4)*8)
		}
		for i, word := range outBuf {
			transferWords[i].Set(word)
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

// I2C on the ESP32.
type I2C struct {
	Bus *esp.I2C_Type
}

// I2CConfig is used to store config info for I2C.
type I2CConfig struct {
	Frequency uint32
	SCL       Pin
	SDA       Pin
}

const (
	i2cFIFOSize     = 255
	i2cFilterCYCNum = 7
	uint32Max       = 0xFFFFFFFF
)

const (
	i2cCmdRestart = iota //0,        /* I2C restart command */
	i2cCmdWrite          /* I2C write command */
	i2cCmdRead           /* I2C read command */
	i2cCmdStop           /* I2C stop command */
	i2cCmdEnd            /* I2C end command */
)

// Configure is intended to setup the I2C interface.
func (i2c I2C) Configure(config I2CConfig) error {
	// Default I2C bus speed is 100 kHz.
	if config.Frequency == 0 {
		config.Frequency = TWI_FREQ_100KHZ
	}
	// Default I2C pins if not set.
	if config.SDA == 0 && config.SCL == 0 {
		config.SDA = SDA_PIN
		config.SCL = SCL_PIN
	}

	// esp32_gpiowrite(config->scl_pin, 1);
	config.SCL.High()
	// esp32_gpiowrite(config->sda_pin, 1);
	config.SDA.High()

	// esp32_configgpio(config->scl_pin, OUTPUT | OPEN_DRAIN | FUNCTION_2);
	config.SCL.Configure(PinConfig{Mode: PinOutputOpenDrain | PinFunction2})

	// TODO: whatever these do?
	// gpio_matrix_out(config->scl_pin, config->scl_outsig, 0, 0);
	// gpio_matrix_in(config->scl_pin, config->scl_insig, 0);

	// esp32_configgpio(config->sda_pin, INPUT |
	// 								  OUTPUT |
	// 								  OPEN_DRAIN |
	// 								  FUNCTION_2);
	config.SDA.Configure(PinConfig{Mode: PinInput | PinOutputOpenDrain | PinFunction2})

	// TODO: whatever these do?
	// gpio_matrix_out(config->sda_pin, config->sda_outsig, 0, 0);
	// gpio_matrix_in(config->sda_pin, config->sda_insig, 0);

	// modifyreg32(DPORT_PERIP_CLK_EN_REG, 0, config->clk_bit);
	esp.DPORT.PERIP_CLK_EN.SetBits(esp.DPORT_PERIP_CLK_EN_I2C0)

	// modifyreg32(DPORT_PERIP_RST_EN_REG, config->rst_bit, 0);
	esp.DPORT.PERIP_RST_EN.SetBits(esp.DPORT_PERIP_RST_EN_I2C0)

	// esp32_i2c_set_reg(priv, I2C_INT_ENA_OFFSET, 0);
	i2c.Bus.INT_ENA.Set(0)

	// esp32_i2c_set_reg(priv, I2C_INT_CLR_OFFSET, UINT32_MAX);
	i2c.Bus.INT_CLR.Set(uint32Max)

	// esp32_i2c_set_reg(priv, I2C_CTR_OFFSET, I2C_MS_MODE |
	// 										I2C_SCL_FORCE_OUT |
	// 										I2C_SDA_FORCE_OUT);
	i2c.Bus.CTR.Set(esp.I2C_CTR_MS_MODE | esp.I2C_CTR_SCL_FORCE_OUT | esp.I2C_CTR_SDA_FORCE_OUT)

	// esp32_i2c_reset_reg_bits(priv, I2C_FIFO_CONF_OFFSET,
	// 						 I2C_NONFIFO_EN_M);
	i2c.Bus.FIFO_CONF.Set(esp.I2C_FIFO_CONF_NONFIFO_EN)

	i2c.resetFIFO()

	// esp32_i2c_set_reg(priv, I2C_SCL_FILTER_CFG_OFFSET, I2C_SCL_FILTER_EN_M |
	// 	I2C_FILTER_CYC_NUM_DEF);
	i2c.Bus.SCL_FILTER_CFG.Set(esp.I2C_SCL_FILTER_CFG_SCL_FILTER_EN | i2cFilterCYCNum)

	// esp32_i2c_set_reg(priv, I2C_SDA_FILTER_CFG_OFFSET, I2C_SDA_FILTER_EN_M |
	// 	I2C_FILTER_CYC_NUM_DEF);
	i2c.Bus.SDA_FILTER_CFG.Set(esp.I2C_SDA_FILTER_CFG_SDA_FILTER_EN | i2cFilterCYCNum)

	i2c.SetBaudRate(config.Frequency)
	return nil
}

// SetBaudRate sets the communication speed for the I2C.
func (i2c I2C) SetBaudRate(br uint32) {
	// uint32_t half_cycles = APB_CLK_FREQ / clk_freq / 2;
	hc := peripheralClock / br / 2

	// uint32_t timeout_cycles = half_cycles * 20;
	tc := hc * 20

	// esp32_i2c_set_reg(priv, I2C_SCL_LOW_PERIOD_OFFSET, half_cycles);
	i2c.Bus.SCL_LOW_PERIOD.Set(hc)

	// esp32_i2c_set_reg(priv, I2C_SCL_HIGH_PERIOD_OFFSET, half_cycles);
	i2c.Bus.SCL_HIGH_PERIOD.Set(hc)

	// esp32_i2c_set_reg(priv, I2C_SDA_HOLD_OFFSET, half_cycles / 2);
	i2c.Bus.SDA_HOLD.Set(hc >> 1)

	// esp32_i2c_set_reg(priv, I2C_SDA_SAMPLE_OFFSET, half_cycles / 2);
	i2c.Bus.SDA_SAMPLE.Set(hc >> 1)

	// esp32_i2c_set_reg(priv, I2C_SCL_RSTART_SETUP_OFFSET, half_cycles);
	i2c.Bus.SCL_RSTART_SETUP.Set(hc)

	// esp32_i2c_set_reg(priv, I2C_SCL_STOP_SETUP_OFFSET, half_cycles);
	i2c.Bus.SCL_STOP_SETUP.Set(hc)

	// esp32_i2c_set_reg(priv, I2C_SCL_START_HOLD_OFFSET, half_cycles);
	i2c.Bus.SCL_START_HOLD.Set(hc)

	// esp32_i2c_set_reg(priv, I2C_SCL_STOP_HOLD_OFFSET, half_cycles);
	i2c.Bus.SCL_STOP_HOLD.Set(hc)

	// esp32_i2c_set_reg(priv, I2C_TO_OFFSET, timeout_cycles);
	i2c.Bus.TO.Set(tc)
}

// Tx does a single I2C transaction at the specified address.
// It clocks out the given address, writes the bytes in w, reads back len(r)
// bytes and stores them in r, and generates a stop condition on the bus.
func (i2c I2C) Tx(addr uint16, w, r []byte) error {
	return nil
}

// WriteByte writes a single byte to the I2C bus.
func (i2c I2C) WriteByte(data byte) error {
	return nil
}

func i2CBaseCmd(cmd, checkAck int) uint32 {
	return uint32((cmd << 11) | (checkAck << 8))
}

func i2CSendCmd(cmd, checkAck, size int) uint32 {
	return uint32((cmd << 11) | (checkAck << 8) + size)
}

// sendAddress sends the address and start signal
func (i2c I2C) sendAddress(address uint16, write bool) error {
	return nil
}

func (i2c I2C) signalStart(size int) error {
	// 	esp32_i2c_set_reg(priv, I2C_COMD0_OFFSET,
	// 		I2C_BASE_CMD(I2C_CMD_RESTART, 0));
	i2c.Bus.COMD0.Set(i2CBaseCmd(i2cCmdRestart, 0))

	// esp32_i2c_set_reg(priv, I2C_COMD1_OFFSET,
	// 		I2C_SEND_CMD(I2C_CMD_WRITE, 1, 1));
	i2c.Bus.COMD1.Set(i2CSendCmd(i2cCmdWrite, 1, size))

	// esp32_i2c_set_reg(priv, I2C_COMD2_OFFSET,
	// 		I2C_BASE_CMD(I2C_CMD_END, 0));
	i2c.Bus.COMD2.Set(i2CBaseCmd(i2cCmdEnd, 0))

	// esp32_i2c_set_reg(priv, I2C_DATA_OFFSET, (msg->addr << 1) |
	// 							   (msg->flags & I2C_M_READ));
	//i2c.Bus.DATA.Set()

	// esp32_i2c_set_reg(priv, I2C_INT_ENA_OFFSET, I2C_END_DETECT_INT_ENA |
	// 								  I2C_INT_ERR_EN_BITS);
	i2c.Bus.INT_ENA.Set(esp.I2C_INT_ENA_END_DETECT_INT_ENA |
		esp.I2C_INT_ENA_ACK_ERR_INT_ENA)

	// esp32_i2c_set_reg_bits(priv, I2C_CTR_OFFSET, I2C_TRANS_START_M);
	i2c.Bus.CTR.SetBits(esp.I2C_INT_ENA_TRANS_START_INT_ENA)

	return nil
}

func (i2c I2C) signalStop() error {
	// esp32_i2c_set_reg(priv, I2C_COMD0_OFFSET, I2C_BASE_CMD(I2C_CMD_STOP, 0));
	i2c.Bus.COMD0.Set(i2CBaseCmd(i2cCmdStop, 0))

	// esp32_i2c_set_reg(priv, I2C_INT_ENA_OFFSET, I2C_TRANS_COMPLETE_INT_ENA |
	//                                               I2C_INT_ERR_EN_BITS);
	i2c.Bus.INT_ENA.Set(esp.I2C_INT_ENA_TRANS_COMPLETE_INT_ENA |
		esp.I2C_INT_ENA_ACK_ERR_INT_ENA)

	// esp32_i2c_set_reg_bits(priv, I2C_CTR_OFFSET, I2C_TRANS_START_M);
	i2c.Bus.CTR.SetBits(esp.I2C_INT_ENA_TRANS_START_INT_ENA)

	return nil
}

func (i2c I2C) signalRead() error {
	//uint32_t cmd = esp32_i2c_get_reg(priv, I2C_COMD0_OFFSET);
	//uint8_t n = cmd & 0xff;

	return nil
}

func (i2c I2C) readByte() byte {
	// esp32_i2c_get_reg(priv,
	// 	I2C_DATA_OFFSET);
	return byte(i2c.Bus.DATA.Get())
}

func (i2c I2C) resetFIFO() {
	// esp32_i2c_set_reg_bits(priv, I2C_FIFO_CONF_OFFSET, I2C_TX_FIFO_RST);
	i2c.Bus.FIFO_CONF.SetBits(esp.I2C_FIFO_CONF_TX_FIFO_RST)

	// esp32_i2c_reset_reg_bits(priv, I2C_FIFO_CONF_OFFSET, I2C_TX_FIFO_RST);
	i2c.Bus.FIFO_CONF.ClearBits(esp.I2C_FIFO_CONF_TX_FIFO_RST)

	// esp32_i2c_set_reg_bits(priv, I2C_FIFO_CONF_OFFSET, I2C_RX_FIFO_RST);
	i2c.Bus.FIFO_CONF.SetBits(esp.I2C_FIFO_CONF_RX_FIFO_RST)

	// esp32_i2c_reset_reg_bits(priv, I2C_FIFO_CONF_OFFSET, I2C_RX_FIFO_RST);
	i2c.Bus.FIFO_CONF.ClearBits(esp.I2C_FIFO_CONF_RX_FIFO_RST)
}
