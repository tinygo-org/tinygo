//go:build nrf

package machine

import (
	"bytes"
	"device/nrf"
	"encoding/binary"
	"runtime/interrupt"
	"unsafe"
)

const deviceName = nrf.Device

const (
	PinInput         PinMode = (nrf.GPIO_PIN_CNF_DIR_Input << nrf.GPIO_PIN_CNF_DIR_Pos) | (nrf.GPIO_PIN_CNF_INPUT_Connect << nrf.GPIO_PIN_CNF_INPUT_Pos)
	PinInputPullup   PinMode = PinInput | (nrf.GPIO_PIN_CNF_PULL_Pullup << nrf.GPIO_PIN_CNF_PULL_Pos)
	PinInputPulldown PinMode = PinInput | (nrf.GPIO_PIN_CNF_PULL_Pulldown << nrf.GPIO_PIN_CNF_PULL_Pos)
	PinOutput        PinMode = (nrf.GPIO_PIN_CNF_DIR_Output << nrf.GPIO_PIN_CNF_DIR_Pos) | (nrf.GPIO_PIN_CNF_INPUT_Connect << nrf.GPIO_PIN_CNF_INPUT_Pos)
)

type PinChange uint8

// Pin change interrupt constants for SetInterrupt.
const (
	PinRising  PinChange = nrf.GPIOTE_CONFIG_POLARITY_LoToHi
	PinFalling PinChange = nrf.GPIOTE_CONFIG_POLARITY_HiToLo
	PinToggle  PinChange = nrf.GPIOTE_CONFIG_POLARITY_Toggle
)

// Callbacks to be called for pins configured with SetInterrupt.
var pinCallbacks [len(nrf.GPIOTE.CONFIG)]func(Pin)

// Configure this pin with the given configuration.
func (p Pin) Configure(config PinConfig) {
	cfg := config.Mode | nrf.GPIO_PIN_CNF_DRIVE_S0S1 | nrf.GPIO_PIN_CNF_SENSE_Disabled
	port, pin := p.getPortPin()
	port.PIN_CNF[pin].Set(uint32(cfg))
}

// Set the pin to high or low.
// Warning: only use this on an output pin!
func (p Pin) Set(high bool) {
	port, pin := p.getPortPin()
	if high {
		port.OUTSET.Set(1 << pin)
	} else {
		port.OUTCLR.Set(1 << pin)
	}
}

// Return the register and mask to enable a given GPIO pin. This can be used to
// implement bit-banged drivers.
func (p Pin) PortMaskSet() (*uint32, uint32) {
	port, pin := p.getPortPin()
	return &port.OUTSET.Reg, 1 << pin
}

// Return the register and mask to disable a given port. This can be used to
// implement bit-banged drivers.
func (p Pin) PortMaskClear() (*uint32, uint32) {
	port, pin := p.getPortPin()
	return &port.OUTCLR.Reg, 1 << pin
}

// Get returns the current value of a GPIO pin when the pin is configured as an
// input or as an output.
func (p Pin) Get() bool {
	port, pin := p.getPortPin()
	return (port.IN.Get()>>pin)&1 != 0
}

// SetInterrupt sets an interrupt to be executed when a particular pin changes
// state. The pin should already be configured as an input, including a pull up
// or down if no external pull is provided.
//
// This call will replace a previously set callback on this pin. You can pass a
// nil func to unset the pin change interrupt. If you do so, the change
// parameter is ignored and can be set to any value (such as 0).
func (p Pin) SetInterrupt(change PinChange, callback func(Pin)) error {
	// Some variables to easily check whether a channel was already configured
	// as an event channel for the given pin.
	// This is not just an optimization, this is requred: the datasheet says
	// that configuring more than one channel for a given pin results in
	// unpredictable behavior.
	expectedConfigMask := uint32(nrf.GPIOTE_CONFIG_MODE_Msk | nrf.GPIOTE_CONFIG_PSEL_Msk)
	expectedConfig := nrf.GPIOTE_CONFIG_MODE_Event<<nrf.GPIOTE_CONFIG_MODE_Pos | uint32(p)<<nrf.GPIOTE_CONFIG_PSEL_Pos

	foundChannel := false
	for i := range nrf.GPIOTE.CONFIG {
		config := nrf.GPIOTE.CONFIG[i].Get()
		if config == 0 || config&expectedConfigMask == expectedConfig {
			// Found an empty GPIOTE channel or one that was already configured
			// for this pin.
			if callback == nil {
				// Disable this channel.
				nrf.GPIOTE.INTENCLR.Set(uint32(1 << uint(i)))
				pinCallbacks[i] = nil
				return nil
			}
			// Enable this channel with the given callback.
			nrf.GPIOTE.INTENCLR.Set(uint32(1 << uint(i)))
			nrf.GPIOTE.CONFIG[i].Set(nrf.GPIOTE_CONFIG_MODE_Event<<nrf.GPIOTE_CONFIG_MODE_Pos |
				uint32(p)<<nrf.GPIOTE_CONFIG_PSEL_Pos |
				uint32(change)<<nrf.GPIOTE_CONFIG_POLARITY_Pos)
			pinCallbacks[i] = callback
			nrf.GPIOTE.INTENSET.Set(uint32(1 << uint(i)))
			foundChannel = true
			break
		}
	}

	if !foundChannel {
		return ErrNoPinChangeChannel
	}

	// Set and enable the GPIOTE interrupt. It's not a problem if this happens
	// more than once.
	interrupt.New(nrf.IRQ_GPIOTE, func(interrupt.Interrupt) {
		for i := range nrf.GPIOTE.EVENTS_IN {
			if nrf.GPIOTE.EVENTS_IN[i].Get() != 0 {
				nrf.GPIOTE.EVENTS_IN[i].Set(0)
				pin := Pin((nrf.GPIOTE.CONFIG[i].Get() & nrf.GPIOTE_CONFIG_PSEL_Msk) >> nrf.GPIOTE_CONFIG_PSEL_Pos)
				pinCallbacks[i](pin)
			}
		}
	}).Enable()

	// Everything was configured correctly.
	return nil
}

// UART on the NRF.
type UART struct {
	Buffer *RingBuffer
}

// UART
var (
	// UART0 is the hardware UART on the NRF SoC.
	_UART0 = UART{Buffer: NewRingBuffer()}
	UART0  = &_UART0
)

// Configure the UART.
func (uart *UART) Configure(config UARTConfig) {
	// Default baud rate to 115200.
	if config.BaudRate == 0 {
		config.BaudRate = 115200
	}

	uart.SetBaudRate(config.BaudRate)

	// Set TX and RX pins
	if config.TX == 0 && config.RX == 0 {
		// Use default pins
		uart.setPins(UART_TX_PIN, UART_RX_PIN)
	} else {
		uart.setPins(config.TX, config.RX)
	}

	nrf.UART0.ENABLE.Set(nrf.UART_ENABLE_ENABLE_Enabled)
	nrf.UART0.TASKS_STARTTX.Set(1)
	nrf.UART0.TASKS_STARTRX.Set(1)
	nrf.UART0.INTENSET.Set(nrf.UART_INTENSET_RXDRDY_Msk)

	// Enable RX IRQ.
	intr := interrupt.New(nrf.IRQ_UART0, _UART0.handleInterrupt)
	intr.SetPriority(0xc0) // low priority
	intr.Enable()
}

// SetBaudRate sets the communication speed for the UART.
func (uart *UART) SetBaudRate(br uint32) {
	// Magic: calculate 'baudrate' register from the input number.
	// Every value listed in the datasheet will be converted to the
	// correct register value, except for 192600. I suspect the value
	// listed in the nrf52 datasheet (0x0EBED000) is incorrectly rounded
	// and should be 0x0EBEE000, as the nrf51 datasheet lists the
	// nonrounded value 0x0EBEDFA4.
	// Some background:
	// https://devzone.nordicsemi.com/f/nordic-q-a/391/uart-baudrate-register-values/2046#2046
	rate := uint32((uint64(br/400)*uint64(400*0xffffffff/16000000) + 0x800) & 0xffffff000)

	nrf.UART0.BAUDRATE.Set(rate)
}

// WriteByte writes a byte of data to the UART.
func (uart *UART) WriteByte(c byte) error {
	nrf.UART0.EVENTS_TXDRDY.Set(0)
	nrf.UART0.TXD.Set(uint32(c))
	for nrf.UART0.EVENTS_TXDRDY.Get() == 0 {
	}
	return nil
}

func (uart *UART) flush() {}

func (uart *UART) handleInterrupt(interrupt.Interrupt) {
	if nrf.UART0.EVENTS_RXDRDY.Get() != 0 {
		uart.Receive(byte(nrf.UART0.RXD.Get()))
		nrf.UART0.EVENTS_RXDRDY.Set(0x0)
	}
}

// I2CConfig is used to store config info for I2C.
type I2CConfig struct {
	Frequency uint32
	SCL       Pin
	SDA       Pin
	Mode      I2CMode
}

// Configure is intended to setup the I2C interface.
func (i2c *I2C) Configure(config I2CConfig) error {

	i2c.disable()

	// Default I2C bus speed is 100 kHz.
	if config.Frequency == 0 {
		config.Frequency = 100 * KHz
	}
	// Default I2C pins if not set.
	if config.SDA == 0 && config.SCL == 0 {
		config.SDA = SDA_PIN
		config.SCL = SCL_PIN
	}

	// do config
	sclPort, sclPin := config.SCL.getPortPin()
	sclPort.PIN_CNF[sclPin].Set((nrf.GPIO_PIN_CNF_DIR_Input << nrf.GPIO_PIN_CNF_DIR_Pos) |
		(nrf.GPIO_PIN_CNF_INPUT_Connect << nrf.GPIO_PIN_CNF_INPUT_Pos) |
		(nrf.GPIO_PIN_CNF_PULL_Pullup << nrf.GPIO_PIN_CNF_PULL_Pos) |
		(nrf.GPIO_PIN_CNF_DRIVE_S0D1 << nrf.GPIO_PIN_CNF_DRIVE_Pos) |
		(nrf.GPIO_PIN_CNF_SENSE_Disabled << nrf.GPIO_PIN_CNF_SENSE_Pos))

	sdaPort, sdaPin := config.SDA.getPortPin()
	sdaPort.PIN_CNF[sdaPin].Set((nrf.GPIO_PIN_CNF_DIR_Input << nrf.GPIO_PIN_CNF_DIR_Pos) |
		(nrf.GPIO_PIN_CNF_INPUT_Connect << nrf.GPIO_PIN_CNF_INPUT_Pos) |
		(nrf.GPIO_PIN_CNF_PULL_Pullup << nrf.GPIO_PIN_CNF_PULL_Pos) |
		(nrf.GPIO_PIN_CNF_DRIVE_S0D1 << nrf.GPIO_PIN_CNF_DRIVE_Pos) |
		(nrf.GPIO_PIN_CNF_SENSE_Disabled << nrf.GPIO_PIN_CNF_SENSE_Pos))

	i2c.setPins(config.SCL, config.SDA)

	i2c.mode = config.Mode

	if i2c.mode == I2CModeController {
		if config.Frequency >= 400*KHz {
			i2c.Bus.FREQUENCY.Set(nrf.TWI_FREQUENCY_FREQUENCY_K400)
		} else {
			i2c.Bus.FREQUENCY.Set(nrf.TWI_FREQUENCY_FREQUENCY_K100)
		}

		i2c.enableAsController()
	} else {
		i2c.enableAsTarget()
	}

	return nil
}

// signalStop sends a stop signal to the I2C peripheral and waits for confirmation.
func (i2c *I2C) signalStop() {
	i2c.Bus.TASKS_STOP.Set(1)
	for i2c.Bus.EVENTS_STOPPED.Get() == 0 {
	}
	i2c.Bus.EVENTS_STOPPED.Set(0)
}

var rngStarted = false

// GetRNG returns 32 bits of non-deterministic random data based on internal thermal noise.
// According to Nordic's documentation, the random output is suitable for cryptographic purposes.
func GetRNG() (ret uint32, err error) {
	// There's no apparent way to check the status of the RNG peripheral's task, so simply start it
	// to avoid deadlocking while waiting for output.
	if !rngStarted {
		nrf.RNG.TASKS_START.Set(1)
		nrf.RNG.SetCONFIG_DERCEN(nrf.RNG_CONFIG_DERCEN_Enabled)
		rngStarted = true
	}

	// The RNG returns one byte at a time, so stack up four bytes into a single uint32 for return.
	for i := 0; i < 4; i++ {
		// Wait for data to be ready.
		for nrf.RNG.EVENTS_VALRDY.Get() == 0 {
		}
		// Append random byte to output.
		ret = (ret << 8) ^ nrf.RNG.GetVALUE()
		// Unset the EVENTS_VALRDY register to avoid reading the same random output twice.
		nrf.RNG.EVENTS_VALRDY.Set(0)
	}

	return ret, nil
}

// ReadTemperature reads the silicon die temperature of the chip. The return
// value is in milli-celsius.
func ReadTemperature() int32 {
	nrf.TEMP.TASKS_START.Set(1)
	for nrf.TEMP.EVENTS_DATARDY.Get() == 0 {
	}
	temp := int32(nrf.TEMP.TEMP.Get()) * 250 // the returned value is in units of 0.25°C
	nrf.TEMP.EVENTS_DATARDY.Set(0)
	return temp
}

const memoryStart = 0x0

// compile-time check for ensuring we fulfill BlockDevice interface
var _ BlockDevice = flashBlockDevice{}

var Flash flashBlockDevice

type flashBlockDevice struct {
}

// ReadAt reads the given number of bytes from the block device.
func (f flashBlockDevice) ReadAt(p []byte, off int64) (n int, err error) {
	if FlashDataStart()+uintptr(off)+uintptr(len(p)) > FlashDataEnd() {
		return 0, errFlashCannotReadPastEOF
	}

	data := unsafe.Slice((*byte)(unsafe.Pointer(FlashDataStart()+uintptr(off))), len(p))
	copy(p, data)

	return len(p), nil
}

// WriteAt writes the given number of bytes to the block device.
// Only double-word (64 bits) length data can be programmed. See rm0461 page 78.
// If the length of p is not long enough it will be padded with 0xFF bytes.
// This method assumes that the destination is already erased.
func (f flashBlockDevice) WriteAt(p []byte, off int64) (n int, err error) {
	if FlashDataStart()+uintptr(off)+uintptr(len(p)) > FlashDataEnd() {
		return 0, errFlashCannotWritePastEOF
	}

	address := FlashDataStart() + uintptr(off)
	padded := f.pad(p)

	waitWhileFlashBusy()

	nrf.NVMC.SetCONFIG_WEN(nrf.NVMC_CONFIG_WEN_Wen)
	defer nrf.NVMC.SetCONFIG_WEN(nrf.NVMC_CONFIG_WEN_Ren)

	for j := 0; j < len(padded); j += int(f.WriteBlockSize()) {
		// write word
		*(*uint32)(unsafe.Pointer(address)) = binary.LittleEndian.Uint32(padded[j : j+int(f.WriteBlockSize())])
		address += uintptr(f.WriteBlockSize())
		waitWhileFlashBusy()
	}

	return len(padded), nil
}

// Size returns the number of bytes in this block device.
func (f flashBlockDevice) Size() int64 {
	return int64(FlashDataEnd() - FlashDataStart())
}

const writeBlockSize = 4

// WriteBlockSize returns the block size in which data can be written to
// memory. It can be used by a client to optimize writes, non-aligned writes
// should always work correctly.
func (f flashBlockDevice) WriteBlockSize() int64 {
	return writeBlockSize
}

// EraseBlockSize returns the smallest erasable area on this particular chip
// in bytes. This is used for the block size in EraseBlocks.
// It must be a power of two, and may be as small as 1. A typical size is 4096.
func (f flashBlockDevice) EraseBlockSize() int64 {
	return eraseBlockSize()
}

// EraseBlocks erases the given number of blocks. An implementation may
// transparently coalesce ranges of blocks into larger bundles if the chip
// supports this. The start and len parameters are in block numbers, use
// EraseBlockSize to map addresses to blocks.
func (f flashBlockDevice) EraseBlocks(start, len int64) error {
	address := FlashDataStart() + uintptr(start*f.EraseBlockSize())
	waitWhileFlashBusy()

	nrf.NVMC.SetCONFIG_WEN(nrf.NVMC_CONFIG_WEN_Een)
	defer nrf.NVMC.SetCONFIG_WEN(nrf.NVMC_CONFIG_WEN_Ren)

	for i := start; i < start+len; i++ {
		nrf.NVMC.ERASEPAGE.Set(uint32(address))
		waitWhileFlashBusy()
		address += uintptr(f.EraseBlockSize())
	}

	return nil
}

// pad data if needed so it is long enough for correct byte alignment on writes.
func (f flashBlockDevice) pad(p []byte) []byte {
	overflow := int64(len(p)) % f.WriteBlockSize()
	if overflow == 0 {
		return p
	}

	padding := bytes.Repeat([]byte{0xff}, int(f.WriteBlockSize()-overflow))
	return append(p, padding...)
}

func waitWhileFlashBusy() {
	for nrf.NVMC.GetREADY() != nrf.NVMC_READY_READY_Ready {
	}
}
