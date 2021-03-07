// +build mimxrt1062

package machine

import (
	"device/nxp"
	"runtime/interrupt"
	"runtime/volatile"
)

// UART peripheral abstraction layer for the MIMXRT1062

type UART struct {
	Bus       *nxp.LPUART_Type
	Buffer    *RingBuffer
	Interrupt interrupt.Interrupt

	// txBuffer should be allocated globally (such as when UART is created) to
	// prevent it being reclaimed or cleaned up prematurely.
	txBuffer *RingBuffer

	// these hold the input selector ("daisy chain") values that select which pins
	// are connected to the LPUART device, and should be defined where the UART
	// instance is declared. see the godoc comments on type muxSelect for more
	// details.
	muxRX, muxTX muxSelect

	// these are copied from UARTConfig, during (*UART).Configure(UARTConfig), and
	// should be considered read-only for internal reference (i.e., modifying them
	// will have no desirable effect).
	rx, tx Pin
	baud   uint32

	// auxiliary state data used internally
	configured   bool
	msbFirst     bool
	transmitting volatile.Register32
}

func (uart *UART) isTransmitting() bool { return uart.transmitting.Get() != 0 }
func (uart *UART) startTransmitting()   { uart.transmitting.Set(1) }
func (uart *UART) stopTransmitting()    { uart.transmitting.Set(0) }
func (uart *UART) resetTransmitting() {
	uart.stopTransmitting()
	uart.Bus.GLOBAL.SetBits(nxp.LPUART_GLOBAL_RST)
	uart.Bus.GLOBAL.ClearBits(nxp.LPUART_GLOBAL_RST)
}

// Configure initializes a UART with the given UARTConfig and other default
// settings.
func (uart *UART) Configure(config UARTConfig) {

	const defaultUartFreq = 115200

	// use default baud rate if not specified
	if config.BaudRate == 0 {
		config.BaudRate = defaultUartFreq
	}
	uart.baud = config.BaudRate

	// use default UART pins if not specified
	if config.RX == 0 && config.TX == 0 {
		config.RX = UART_RX_PIN
		config.TX = UART_TX_PIN
	}
	uart.rx = config.RX
	uart.tx = config.TX

	// configure the mux and pad control registers
	uart.rx.Configure(PinConfig{Mode: PinModeUARTRX})
	uart.tx.Configure(PinConfig{Mode: PinModeUARTTX})

	// configure the mux input selector
	uart.muxRX.connect()
	uart.muxTX.connect()

	// reset all internal logic and registers
	uart.resetTransmitting()

	// determine the baud rate and over-sample divisors
	sbr, osr := uart.getBaudRateDivisor(uart.baud)

	// for now we assume some configuration. in particular:
	//  Data bits         -> 8-bit
	//  Parity bit        -> None (parity bit generation disabled)
	//  Stop bits         -> 1 stop bit
	//  MSB first         -> false
	//  RX idle type      -> idle count starts after start bit
	//  RX idle config    -> 1 idle character
	//  RX RTS enabled    -> false
	//  TX CTS enabled    -> false

	// set the baud rate, over-sample configuration, stop bits
	baudBits := (((osr - 1) << nxp.LPUART_BAUD_OSR_Pos) & nxp.LPUART_BAUD_OSR_Msk) |
		((sbr << nxp.LPUART_BAUD_SBR_Pos) & nxp.LPUART_BAUD_SBR_Msk) |
		((nxp.LPUART_BAUD_SBNS_SBNS_0 << nxp.LPUART_BAUD_SBNS_Pos) & nxp.LPUART_BAUD_SBNS_Msk)
	if osr <= 8 {
		// if OSR less than or equal to 8, we must enable sampling on both edges
		baudBits |= nxp.LPUART_BAUD_BOTHEDGE
	}
	uart.Bus.BAUD.Set(baudBits)

	uart.Bus.PINCFG.Set(0) // disable triggers

	// use 8 data bits, disable parity, use 1 idle char, and idle count starts
	// after start bit
	ctrlBits := uint32(((nxp.LPUART_CTRL_M_M_0 << nxp.LPUART_CTRL_M_Pos) & nxp.LPUART_CTRL_M_Msk) |
		((nxp.LPUART_CTRL_PE_PE_0 << nxp.LPUART_CTRL_PE_Pos) & nxp.LPUART_CTRL_PE_Msk) |
		((nxp.LPUART_CTRL_ILT_ILT_0 << nxp.LPUART_CTRL_ILT_Pos) & nxp.LPUART_CTRL_ILT_Msk) |
		((nxp.LPUART_CTRL_IDLECFG_IDLECFG_0 << nxp.LPUART_CTRL_IDLECFG_Pos) & nxp.LPUART_CTRL_IDLECFG_Msk))
	uart.Bus.CTRL.Set(ctrlBits)

	rxSize, txSize := uart.getFIFOSize()

	rxWater := rxSize >> 1
	if rxWater > uint32(nxp.LPUART_FIFO_RXFIFOSIZE_Msk>>nxp.LPUART_FIFO_RXFIFOSIZE_Pos) {
		rxWater = uint32(nxp.LPUART_FIFO_RXFIFOSIZE_Msk >> nxp.LPUART_FIFO_RXFIFOSIZE_Pos)
	}

	txWater := txSize >> 1
	if txWater > uint32(nxp.LPUART_FIFO_TXFIFOSIZE_Msk>>nxp.LPUART_FIFO_TXFIFOSIZE_Pos) {
		txWater = uint32(nxp.LPUART_FIFO_TXFIFOSIZE_Msk >> nxp.LPUART_FIFO_TXFIFOSIZE_Pos)
	}

	uart.Bus.WATER.Set(
		((rxWater << nxp.LPUART_WATER_RXWATER_Pos) & nxp.LPUART_WATER_RXWATER_Msk) |
			((txWater << nxp.LPUART_WATER_TXWATER_Pos) & nxp.LPUART_WATER_TXWATER_Msk))

	// enable TX/RX FIFOs
	uart.Bus.FIFO.SetBits(nxp.LPUART_FIFO_RXFE | nxp.LPUART_FIFO_TXFE)

	// flush TX/RX FIFOs
	uart.Bus.FIFO.SetBits(nxp.LPUART_FIFO_RXFLUSH | nxp.LPUART_FIFO_TXFLUSH)

	uart.Bus.MODIR.SetBits( // set the CTS configuration/TX CTS source
		((nxp.LPUART_MODIR_TXCTSC_TXCTSC_0 << nxp.LPUART_MODIR_TXCTSC_Pos) & nxp.LPUART_MODIR_TXCTSC_Msk) |
			((nxp.LPUART_MODIR_TXCTSSRC_TXCTSSRC_0 << nxp.LPUART_MODIR_TXCTSSRC_Pos) & nxp.LPUART_MODIR_TXCTSSRC_Msk))

	// clear all status flags
	stat := uint32(nxp.LPUART_STAT_RXEDGIF_Msk | nxp.LPUART_STAT_IDLE_Msk | nxp.LPUART_STAT_OR_Msk |
		nxp.LPUART_STAT_NF_Msk | nxp.LPUART_STAT_FE_Msk | nxp.LPUART_STAT_PF_Msk |
		nxp.LPUART_STAT_LBKDIF_Msk | nxp.LPUART_STAT_MA1F_Msk | nxp.LPUART_STAT_MA2F_Msk)

	// set data bits order
	if uart.msbFirst {
		stat |= nxp.LPUART_STAT_MSBF
	} else {
		stat &^= nxp.LPUART_STAT_MSBF
	}

	uart.Bus.STAT.SetBits(stat)

	// enable RX/TX functions
	uart.Bus.CTRL.SetBits(nxp.LPUART_CTRL_TE | nxp.LPUART_CTRL_RE)

	// enable RX IRQ
	uart.Interrupt.SetPriority(0xc0)
	uart.Interrupt.Enable()

	uart.configured = true
}

// Disable disables the UART interface.
//
// If any buffered data has not yet been transmitted, Disable waits until
// transmission completes before disabling the interface. The receiver UART's
// interrupt is also disabled, and the RX/TX pins are reconfigured for GPIO
// input (pull-up).
func (uart *UART) Disable() {

	// first ensure the device is enabled
	if uart.configured {

		// wait for any buffered data to send
		uart.Sync()

		// stop trapping RX interrupts
		uart.Interrupt.Disable()

		// reset all internal registers
		uart.resetTransmitting()

		// disable RX/TX functions
		uart.Bus.CTRL.ClearBits(nxp.LPUART_CTRL_TE | nxp.LPUART_CTRL_RE)

		// put pins back into GPIO mode
		uart.rx.Configure(PinConfig{Mode: PinInputPullUp})
		uart.tx.Configure(PinConfig{Mode: PinInputPullUp})
	}
	uart.configured = false
}

// Sync blocks the calling goroutine until all data in the output buffer has
// been transmitted.
func (uart *UART) Sync() error {
	for uart.isTransmitting() {
	}
	return nil
}

// WriteByte writes a single byte of data to the UART interface.
func (uart *UART) WriteByte(c byte) error {
	uart.startTransmitting()
	for !uart.txBuffer.Put(c) {
	}
	uart.Bus.CTRL.SetBits(nxp.LPUART_CTRL_TIE)
	return nil
}

// getBaudRateDivisor finds the greatest over-sampling factor (4..32) and
// corresponding baud rate divisor (1..8191) that best partition a given baud
// rate into equal intervals.
//
// This is an integral (i.e. non-floating point) port of the logic at the
// beginning of:
//       void HardwareSerial::begin(uint32_t baud, uint16_t format)
//         (from Teensyduino: `cores/teensy4/HardwareSerial.cpp`)
//
// We don't want to risk using floating point here in the machine package in
// case it gets called before the FPU or interrupts are ready (e.g., init()).
func (uart *UART) getBaudRateDivisor(baudRate uint32) (sbr uint32, osr uint32) {
	const clock = 24000000 // UART is muxed to 24 MHz OSC
	err := uint32(0xFFFFFFFF)
	sbr, osr = 0, 0
	for o := uint32(4); o <= 32; o++ {
		s := ((clock*10)/(baudRate*o) + 5) / 10
		if s == 0 {
			s = 1
		}
		b := clock / (s * o)
		var e uint32
		if b > baudRate {
			e = b - baudRate
		} else {
			e = baudRate - b
		}
		if e <= err {
			err = e
			osr = o
			sbr = s
		}
	}
	return sbr, osr
}

func (uart *UART) getFIFOSize() (rx, tx uint32) {
	fifo := uart.Bus.FIFO.Get()
	rx = uint32(1) << ((fifo & nxp.LPUART_FIFO_RXFIFOSIZE_Msk) >> nxp.LPUART_FIFO_RXFIFOSIZE_Pos)
	if rx > 1 {
		rx <<= 1
	}
	tx = uint32(1) << ((fifo & nxp.LPUART_FIFO_TXFIFOSIZE_Msk) >> nxp.LPUART_FIFO_TXFIFOSIZE_Pos)
	if tx > 1 {
		tx <<= 1
	}
	return rx, tx
}

func (uart *UART) getStatus() uint32 {
	return uart.Bus.STAT.Get() |
		((uart.Bus.FIFO.Get() & uint32(nxp.LPUART_FIFO_TXEMPT_Msk|nxp.LPUART_FIFO_RXEMPT_Msk|
			nxp.LPUART_FIFO_TXOF_Msk|nxp.LPUART_FIFO_RXUF_Msk)) >> 16)
}

func (uart *UART) getEnabledInterrupts() uint32 {
	return ((uart.Bus.BAUD.Get() & uint32(nxp.LPUART_BAUD_LBKDIE_Msk|nxp.LPUART_BAUD_RXEDGIE_Msk)) >> 8) |
		((uart.Bus.FIFO.Get() & uint32(nxp.LPUART_FIFO_TXOFE_Msk|nxp.LPUART_FIFO_RXUFE_Msk)) >> 8) |
		(uart.Bus.CTRL.Get() & uint32(0xFF0C000))
}

func (uart *UART) disableInterrupts(mask uint32) {
	uart.Bus.BAUD.ClearBits((mask << 8) & uint32(nxp.LPUART_BAUD_LBKDIE_Msk|nxp.LPUART_BAUD_RXEDGIE_Msk))
	uart.Bus.FIFO.Set((uart.Bus.FIFO.Get() & ^uint32(nxp.LPUART_FIFO_TXOF_Msk|nxp.LPUART_FIFO_RXUF_Msk)) &
		^uint32((mask<<8)&(nxp.LPUART_FIFO_TXOFE_Msk|nxp.LPUART_FIFO_RXUFE_Msk)))
	mask &= uint32(0xFFFFFF00)
	uart.Bus.CTRL.ClearBits(mask)
}

func (uart *UART) handleInterrupt(interrupt.Interrupt) {

	stat := uart.getStatus()
	inte := uart.getEnabledInterrupts()

	_, txSize := uart.getFIFOSize()

	// check for and clear overrun, otherwise RX will not work
	if (stat & uint32(nxp.LPUART_STAT_OR)) != 0 {
		uart.Bus.STAT.Set((uart.Bus.STAT.Get() & uint32(0x3FE00000)) | nxp.LPUART_STAT_OR)
	}

	// idle or receive data register is full
	if (stat & uint32(nxp.LPUART_STAT_RDRF|nxp.LPUART_STAT_IDLE)) != 0 {
		count := (uart.Bus.WATER.Get() & uint32(nxp.LPUART_WATER_RXCOUNT_Msk)) >> nxp.LPUART_WATER_RXCOUNT_Pos
		for ; count > 0; count-- {
			// read up to 8 bits of data at a time
			// TODO: 7, 9, and 10-bit support?
			uart.Buffer.Put(uint8(uart.Bus.DATA.Get() & uint32(0xFF)))
		}
		// if it was an IDLE status, clear the flag
		if (stat & uint32(nxp.LPUART_STAT_IDLE)) != 0 {
			uart.Bus.STAT.SetBits(nxp.LPUART_STAT_IDLE)
		}
		// disable idle line interrupts
		uart.disableInterrupts(nxp.LPUART_CTRL_RIE | nxp.LPUART_CTRL_ORIE)
	}

	// check if we have data to write
	if ((inte & nxp.LPUART_CTRL_TIE) != 0) && ((stat & nxp.LPUART_STAT_TDRE) != 0) {
		for ((uart.Bus.WATER.Get() & uint32(nxp.LPUART_WATER_TXCOUNT_Msk)) >> nxp.LPUART_WATER_TXCOUNT_Pos) < txSize {
			if b, ok := uart.txBuffer.Get(); ok {
				uart.Bus.DATA.Set(uint32(b))
			} else {
				break
			}
		}
		if uart.Bus.STAT.HasBits(nxp.LPUART_STAT_TDRE) {
			uart.Bus.CTRL.Set((uart.Bus.CTRL.Get() & ^uint32(nxp.LPUART_CTRL_TIE)) | nxp.LPUART_CTRL_TCIE)
		}
	}

	if ((inte & nxp.LPUART_CTRL_TCIE) != 0) && ((stat & nxp.LPUART_STAT_TC) != 0) {
		uart.stopTransmitting()
		uart.Bus.CTRL.ClearBits(nxp.LPUART_CTRL_TCIE)
	}
}
