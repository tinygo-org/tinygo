//go:build esp32

package machine

import (
	"device/esp"
	"errors"
	"runtime/interrupt"
	"runtime/volatile"
	"sync"
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

	// Internal pull resistors for pins with RTC function ignore
	// the IO_MUX_GPIO0_FUN_WPU and IO_MUX_GPIO0_FUN_WPD bits set
	// above and are instead controlled by the RTC_IO registers.
	p.configureRTCPull(config.Mode)

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
		return &esp.IO_MUX.GPIO14
	case 12:
		return &esp.IO_MUX.GPIO12
	case 13:
		return &esp.IO_MUX.GPIO13
	case 15:
		return &esp.IO_MUX.GPIO15
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
		return &esp.IO_MUX.GPIO9
	case 10:
		return &esp.IO_MUX.GPIO10
	case 11:
		return &esp.IO_MUX.GPIO11
	case 6:
		return &esp.IO_MUX.GPIO6
	case 7:
		return &esp.IO_MUX.GPIO7
	case 8:
		return &esp.IO_MUX.GPIO8
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
		return &esp.IO_MUX.GPIO3
	case 1:
		return &esp.IO_MUX.GPIO1
	case 23:
		return &esp.IO_MUX.GPIO23
	case 24:
		return &esp.IO_MUX.GPIO24
	default:
		return nil
	}
}

// configureRTCPull applies the pullup/pulldown setting to a pin.
// This mirrors ESP-IDF:
// https://github.com/espressif/esp-idf/blob/08e0d30/components/esp_driver_gpio/src/gpio.c#L283
// https://github.com/espressif/esp-idf/blob/08e0d30/components/esp_hal_gpio/esp32/rtc_io_periph.c#L57
// https://github.com/espressif/esp-idf/blob/08e0d30/components/esp_hal_gpio/esp32/include/hal/rtc_io_ll.h#L175
func (p Pin) configureRTCPull(mode PinMode) {
	var rue, rde uint32
	switch mode {
	case PinInputPullup:
		rue = 1
	case PinInputPulldown:
		rde = 1
	}

	switch p {
	case 0:
		esp.RTC_IO.SetTOUCH_PAD1_RUE(rue)
		esp.RTC_IO.SetTOUCH_PAD1_RDE(rde)
	case 2:
		esp.RTC_IO.SetTOUCH_PAD2_RUE(rue)
		esp.RTC_IO.SetTOUCH_PAD2_RDE(rde)
	case 4:
		esp.RTC_IO.SetTOUCH_PAD0_RUE(rue)
		esp.RTC_IO.SetTOUCH_PAD0_RDE(rde)
	case 12:
		esp.RTC_IO.SetTOUCH_PAD5_RUE(rue)
		esp.RTC_IO.SetTOUCH_PAD5_RDE(rde)
	case 13:
		esp.RTC_IO.SetTOUCH_PAD4_RUE(rue)
		esp.RTC_IO.SetTOUCH_PAD4_RDE(rde)
	case 14:
		esp.RTC_IO.SetTOUCH_PAD6_RUE(rue)
		esp.RTC_IO.SetTOUCH_PAD6_RDE(rde)
	case 15:
		esp.RTC_IO.SetTOUCH_PAD3_RUE(rue)
		esp.RTC_IO.SetTOUCH_PAD3_RDE(rde)
	case 25:
		esp.RTC_IO.SetPAD_DAC1_PDAC1_RUE(rue)
		esp.RTC_IO.SetPAD_DAC1_PDAC1_RDE(rde)
	case 26:
		esp.RTC_IO.SetPAD_DAC2_PDAC2_RUE(rue)
		esp.RTC_IO.SetPAD_DAC2_PDAC2_RDE(rde)
	case 27:
		esp.RTC_IO.SetTOUCH_PAD7_RUE(rue)
		esp.RTC_IO.SetTOUCH_PAD7_RDE(rde)
	case 32:
		esp.RTC_IO.SetXTAL_32K_PAD_X32P_RUE(rue)
		esp.RTC_IO.SetXTAL_32K_PAD_X32P_RDE(rde)
	case 33:
		esp.RTC_IO.SetXTAL_32K_PAD_X32N_RUE(rue)
		esp.RTC_IO.SetXTAL_32K_PAD_X32N_RDE(rde)
	}
}

const maxPin = 40

// cpuInterruptFromPin selects an edge-triggered CPU interrupt line for GPIO.
// CPU interrupt 10 is edge-triggered level-1 on the Xtensa LX6, which prevents
// the ISR from re-entering continuously when other peripherals (e.g. SPI via
// the GPIO Matrix) keep GPIO.STATUS bits asserted.
const cpuInterruptFromPin = 10

type PinChange uint8

// Pin change interrupt constants for SetInterrupt.
const (
	PinRising PinChange = iota + 1
	PinFalling
	PinToggle
)

// SetInterrupt sets an interrupt to be executed when a particular pin changes
// state. The pin should already be configured as an input, including a pull up
// or down if no external pull is provided.
//
// You can pass a nil func to unset the pin change interrupt. If you do so,
// the change parameter is ignored and can be set to any value (such as 0).
// If the pin is already configured with a callback, you must first unset
// this pins interrupt before you can set a new callback.
func (p Pin) SetInterrupt(change PinChange, callback func(Pin)) error {
	if p >= maxPin {
		return ErrInvalidInputPin
	}

	if callback == nil {
		// Disable this pin interrupt
		p.pinReg().ClearBits(esp.GPIO_PIN_INT_TYPE_Msk | esp.GPIO_PIN_INT_ENA_Msk)

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
		setupPinInterruptErr = setupPinInterrupt()
	})
	if setupPinInterruptErr != nil {
		return setupPinInterruptErr
	}

	p.pinReg().Set(
		(p.pinReg().Get() & ^uint32(esp.GPIO_PIN_INT_TYPE_Msk|esp.GPIO_PIN_INT_ENA_Msk)) |
			uint32(change)<<esp.GPIO_PIN_INT_TYPE_Pos | uint32(1)<<esp.GPIO_PIN_INT_ENA_Pos)

	return nil
}

var (
	pinCallbacks          [maxPin]func(Pin)
	onceSetupPinInterrupt sync.Once
	setupPinInterruptErr  error
)

func setupPinInterrupt() error {
	esp.DPORT.SetPRO_GPIO_INTERRUPT_MAP_PRO_GPIO_INTERRUPT_PRO_MAP(cpuInterruptFromPin)
	return interrupt.New(cpuInterruptFromPin, handleGPIOInterrupt).Enable()
}

// handleGPIOInterrupt is the GPIO pin change interrupt handler. It must be a
// plain function (not a closure) because interrupt.New is a compiler intrinsic
// that does not support closures.
func handleGPIOInterrupt(interrupt.Interrupt) {
	// Read and immediately clear interrupt status bits.
	// Clearing before processing is critical for edge-triggered CPU
	// interrupts: any new GPIO events that arrive during callback
	// execution will set fresh STATUS bits, generating a new edge
	// on the CPU interrupt line so they are not lost.
	status := esp.GPIO.STATUS.Get()
	status1 := esp.GPIO.STATUS1.Get()
	esp.GPIO.STATUS_W1TC.Set(status)
	esp.GPIO.STATUS1_W1TC.Set(status1)

	// Check status for GPIO0-31
	for i, mask := 0, uint32(1); i < 32; i, mask = i+1, mask<<1 {
		if (status&mask) != 0 && pinCallbacks[i] != nil {
			pinCallbacks[i](Pin(i))
		}
	}
	// Check status for GPIO32-39
	for i, mask := 32, uint32(1); i < maxPin; i, mask = i+1, mask<<1 {
		if (status1&mask) != 0 && pinCallbacks[i] != nil {
			pinCallbacks[i](Pin(i))
		}
	}
}

var DefaultUART = UART0

var (
	UART0  = &_UART0
	_UART0 = UART{
		Bus:          esp.UART0,
		Buffer:       NewRingBuffer(),
		txrxSignal:   14,
		rtsctsSignal: 15,
	}
	UART1  = &_UART1
	_UART1 = UART{
		Bus:          esp.UART1,
		Buffer:       NewRingBuffer(),
		txrxSignal:   17,
		rtsctsSignal: 18,
	}
	UART2  = &_UART2
	_UART2 = UART{
		Bus:          esp.UART2,
		Buffer:       NewRingBuffer(),
		txrxSignal:   198,
		rtsctsSignal: 199,
	}

	onceUart = sync.Once{}
)

// CPU interrupt line used for all UART peripherals.
//
// On the ESP32 (Xtensa LX6) the 32 CPU interrupt lines have fixed hardware
// roles. Lines 6, 7, 11, 14, 15, 16 and 29 are internal (Xtensa timers,
// software and profiling/NMI) and are NOT wired to the peripheral interrupt
// matrix, so a peripheral routed to one of them via DPORT never fires.
// The usable level-1 peripheral lines are 2, 3, 5, 8, 9, 10 (edge), 12, 13,
// 17 and 18. We use line 8 for UART (9 is the timer alarm, 10 is GPIO).
const cpuInterruptFromUART = 8

// uartInterrupts is the set of UART interrupt flags we care about for RX.
const uartInterrupts = esp.UART_INT_ENA_RXFIFO_FULL_INT_ENA |
	esp.UART_INT_ENA_RXFIFO_TOUT_INT_ENA |
	esp.UART_INT_ENA_PARITY_ERR_INT_ENA |
	esp.UART_INT_ENA_FRM_ERR_INT_ENA |
	esp.UART_INT_ENA_RXFIFO_OVF_INT_ENA |
	esp.UART_INT_ENA_GLITCH_DET_INT_ENA

type UART struct {
	Bus    *esp.UART_Type
	Buffer *RingBuffer

	txrxSignal           uint32
	rtsctsSignal         uint32
	parityErrorDetected  bool
	dataErrorDetected    bool
	dataOverflowDetected bool
}

func (uart *UART) Configure(config UARTConfig) error {
	if config.BaudRate == 0 {
		config.BaudRate = 115200
	}

	// If no pins are specified (the zero value is GPIO0, which is never a
	// sensible default for both TX and RX), pick sensible defaults per UART.
	//
	// For UART0 (the console) we deliberately leave the pins untouched: the
	// ROM bootloader has already wired GPIO1 (TX) and GPIO3 (RX) directly via
	// the IO MUX to the USB-serial bridge. Re-routing them through the GPIO
	// matrix is unnecessary and can break RX, so we keep the bootloader setup
	// which is exactly what makes the boot log and greeting appear.
	// We still fall through to configure baud rate, interrupts, and the RX
	// FIFO even when pins are already wired.
	if config.TX == 0 && config.RX == 0 {
		switch uart.Bus {
		case esp.UART0:
			config.TX = NoPin
			config.RX = NoPin
		case esp.UART1:
			config.TX = 10
			config.RX = 9
		case esp.UART2:
			config.TX = 17
			config.RX = 16
		}
	}

	uart.Bus.CLKDIV.Set(peripheralClock / config.BaudRate)

	if config.RX != NoPin {
		config.RX.configure(PinConfig{Mode: PinInputPullup}, uart.txrxSignal)
		if config.InvertRX {
			inFunc(uart.txrxSignal).Set(esp.GPIO_FUNC_IN_SEL_CFG_SEL | uint32(config.RX)<<esp.GPIO_FUNC_IN_SEL_CFG_IN_SEL_Pos | esp.GPIO_FUNC_IN_SEL_CFG_IN_INV_SEL)
		} else {
			inFunc(uart.txrxSignal).Set(esp.GPIO_FUNC_IN_SEL_CFG_SEL | uint32(config.RX)<<esp.GPIO_FUNC_IN_SEL_CFG_IN_SEL_Pos)
		}
	}

	if config.TX != NoPin {
		config.TX.configure(PinConfig{Mode: PinOutput}, uart.txrxSignal)
		if config.InvertTX {
			config.TX.outFunc().Set(uart.txrxSignal | esp.GPIO_FUNC_OUT_SEL_CFG_INV_SEL)
		} else {
			config.TX.outFunc().Set(uart.txrxSignal)
		}
	}

	if config.RTS != NoPin {
		config.RTS.configure(PinConfig{Mode: PinOutput}, uart.rtsctsSignal)
	}

	if config.CTS != NoPin {
		config.CTS.configure(PinConfig{Mode: PinInputPullup}, uart.rtsctsSignal)
	}

	uart.configureInterrupt()
	uart.enableReceiver()

	return nil
}

func (uart *UART) configureInterrupt() {
	// Disable all UART interrupts while configuring.
	uart.Bus.INT_ENA.ClearBits(0x0ffff)

	// Map this UART's peripheral interrupt to a CPU interrupt line via DPORT.
	switch uart.Bus {
	case esp.UART0:
		esp.DPORT.SetPRO_UART_INTR_MAP(cpuInterruptFromUART)
	case esp.UART1:
		esp.DPORT.SetPRO_UART1_INTR_MAP(cpuInterruptFromUART)
	case esp.UART2:
		esp.DPORT.SetPRO_UART2_INTR_MAP(cpuInterruptFromUART)
	}

	// Register the ISR only once (shared across all UARTs on the same CPU int).
	// interrupt.New is a compiler intrinsic and requires a plain (non-capturing)
	// handler function, so we use a named package-level function.
	onceUart.Do(func() {
		_ = interrupt.New(cpuInterruptFromUART, handleUARTInterrupt).Enable()
	})
}

// handleUARTInterrupt is the shared UART interrupt handler. It must be a plain
// function (not a closure) because interrupt.New is a compiler intrinsic that
// does not support closures.
func handleUARTInterrupt(interrupt.Interrupt) {
	UART0.serveInterrupt()
	UART1.serveInterrupt()
	UART2.serveInterrupt()
}

func (uart *UART) serveInterrupt() {
	// Check masked interrupt status.
	interruptFlag := uart.Bus.INT_ST.Get()
	if (interruptFlag & uartInterrupts) == 0 {
		return
	}

	// Block UART interrupts while processing.
	uart.Bus.INT_ENA.ClearBits(uartInterrupts)

	if interruptFlag&(esp.UART_INT_ENA_RXFIFO_FULL_INT_ENA|esp.UART_INT_ENA_RXFIFO_TOUT_INT_ENA) != 0 {
		for uart.Bus.GetSTATUS_RXFIFO_CNT() > 0 {
			// The ESP32 UART FIFO must be accessed through the AHB address
			// (base + 0x200C0000 == 0x60000000 for UART0), not the APB FIFO
			// register, due to a silicon erratum. This mirrors writeByte.
			b := (*volatile.Register8)(unsafe.Add(unsafe.Pointer(uart.Bus), 0x200C0000)).Get()
			if !uart.Buffer.Put(b) {
				uart.dataOverflowDetected = true
			}
		}
	}
	if interruptFlag&esp.UART_INT_ENA_PARITY_ERR_INT_ENA > 0 {
		uart.parityErrorDetected = true
	}
	if interruptFlag&esp.UART_INT_ENA_FRM_ERR_INT_ENA != 0 {
		uart.dataErrorDetected = true
	}
	if interruptFlag&esp.UART_INT_ENA_RXFIFO_OVF_INT_ENA != 0 {
		uart.dataOverflowDetected = true
	}
	if interruptFlag&esp.UART_INT_ENA_GLITCH_DET_INT_ENA != 0 {
		uart.dataErrorDetected = true
	}

	// Clear the interrupt status bits.
	uart.Bus.INT_CLR.SetBits(interruptFlag)
	uart.Bus.INT_CLR.ClearBits(interruptFlag)
	// Re-enable UART interrupts.
	uart.Bus.INT_ENA.Set(uartInterrupts)
}

func (uart *UART) enableReceiver() {
	// Reset the RX FIFO.
	uart.Bus.SetCONF0_RXFIFO_RST(1)
	uart.Bus.SetCONF0_RXFIFO_RST(0)
	// Trigger interrupt when 1 byte is available (low latency).
	uart.Bus.SetCONF1_RXFIFO_FULL_THRHD(1)
	// Enable the RX timeout so that a single byte still generates an interrupt
	// once the line has been idle for the given number of bit periods. Without
	// this, RXFIFO_FULL only fires once more than the threshold has arrived.
	uart.Bus.SetCONF1_RX_TOUT_THRHD(2)
	uart.Bus.SetCONF1_RX_TOUT_EN(1)
	// Enable RX-related interrupts.
	uart.Bus.SetINT_ENA_RXFIFO_FULL_INT_ENA(1)
	uart.Bus.SetINT_ENA_RXFIFO_TOUT_INT_ENA(1)
	uart.Bus.SetINT_ENA_FRM_ERR_INT_ENA(1)
	uart.Bus.SetINT_ENA_PARITY_ERR_INT_ENA(1)
	uart.Bus.SetINT_ENA_GLITCH_DET_INT_ENA(1)
	uart.Bus.SetINT_ENA_RXFIFO_OVF_INT_ENA(1)
}

func (uart *UART) writeByte(b byte) error {
	for (uart.Bus.STATUS.Get()>>16)&0xff >= 128 {
		// Read UART_TXFIFO_CNT from the status register, which indicates how
		// many bytes there are in the transmit buffer. Wait until there are
		// less than 128 bytes in this buffer (the default buffer size).
		gosched()
	}
	// Write to the TX_FIFO register.
	(*volatile.Register8)(unsafe.Add(unsafe.Pointer(uart.Bus), 0x200C0000)).Set(b)
	return nil
}

func (uart *UART) flush() {}

// Serial Peripheral Interface on the ESP32.
type SPI struct {
	Bus *esp.SPI_Type
}

var (
	// SPI0 and SPI1 are reserved for use by the caching system etc.
	SPI2 = &SPI{esp.SPI2}
	SPI3 = &SPI{esp.SPI3}
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
func (spi *SPI) Configure(config SPIConfig) error {
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
func (spi *SPI) Transfer(w byte) (byte, error) {
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

// Tx handles read/write operation for SPI interface. Since SPI is a synchronous write/read
// interface, there must always be the same number of bytes written as bytes read.
// This is accomplished by sending zero bits if r is bigger than w or discarding
// the incoming data if w is bigger than r.
func (spi *SPI) Tx(w, r []byte) error {
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
