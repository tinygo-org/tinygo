//go:build rp2040

package machine

import (
	"device/rp"
	"errors"
	"internal/itoa"
)

// I2C on the RP2040.
var (
	I2C0  = &_I2C0
	_I2C0 = I2C{
		Bus: rp.I2C0,
	}
	I2C1  = &_I2C1
	_I2C1 = I2C{
		Bus: rp.I2C1,
	}
)

// The I2C target implementation is based on the C implementation from
// here: https://github.com/vmilea/pico_i2c_slave

// Features: Taken from datasheet.
// Default controller mode, with target mode available (not simultaneously).
// Default target address of RP2040: 0x055
// Supports 10-bit addressing in controller mode
// 16-element transmit buffer
// 16-element receive buffer
// Can be driven from DMA
// Can generate interrupts
// Fast mode plus max transfer speed (1000kb/s)

// GPIO config
// Each controller must connect its clock SCL and data SDA to one pair of GPIOs.
// The I2C standard requires that drivers drivea signal low, or when not driven the signal will be pulled high.
// This applies to SCL and SDA. The GPIO pads should be configured for:
//  Pull-up enabled
//  Slew rate limited
//  Schmitt trigger enabled
// Note: There should also be external pull-ups on the board as the internal pad pull-ups may not be strong enough to pull upexternal circuits.

// I2CConfig is used to store config info for I2C.
type I2CConfig struct {
	Frequency uint32
	// SDA/SCL Serial Data and clock pins. Refer to datasheet to see
	// which pins match the desired bus.
	SDA, SCL Pin
	Mode     I2CMode
}

type I2C struct {
	Bus          *rp.I2C0_Type
	mode         I2CMode
	txInProgress bool
}

var (
	ErrInvalidI2CBaudrate  = errors.New("invalid i2c baudrate")
	ErrInvalidTgtAddr      = errors.New("invalid target i2c address not in 0..0x80 or is reserved")
	ErrI2CGeneric          = errors.New("i2c error")
	ErrRP2040I2CDisable    = errors.New("i2c rp2040 peripheral timeout in disable")
	errInvalidI2CSDA       = errors.New("invalid I2C SDA pin")
	errInvalidI2CSCL       = errors.New("invalid I2C SCL pin")
	ErrI2CAlreadyListening = errors.New("i2c already listening")
	ErrI2CWrongMode        = errors.New("i2c wrong mode")
	ErrI2CUnderflow        = errors.New("i2c underflow")
)

// Tx performs a write and then a read transfer placing the result in
// in r.
//
// Passing a nil value for w or r skips the transfer corresponding to write
// or read, respectively.
//
//	i2c.Tx(addr, nil, r)
//
// Performs only a read transfer.
//
//	i2c.Tx(addr, w, nil)
//
// Performs only a write transfer.
func (i2c *I2C) Tx(addr uint16, w, r []byte) error {
	if i2c.mode != I2CModeController {
		return ErrI2CWrongMode
	}

	// timeout in microseconds.
	const timeout = 40 * 1000 // 40ms is a reasonable time for a real-time system.
	return i2c.tx(uint8(addr), w, r, timeout)
}

// Listen starts listening for I2C requests sent to specified address
//
// addr is the address to listen to
func (i2c *I2C) Listen(addr uint16) error {
	if i2c.mode != I2CModeTarget {
		return ErrI2CWrongMode
	}

	return i2c.listen(uint8(addr))
}

// Configure initializes i2c peripheral and configures I2C config's pins passed.
// Here's a list of valid SDA and SCL GPIO pins on bus I2C0 of the rp2040:
//
//	SDA: 0, 4, 8, 12, 16, 20
//	SCL: 1, 5, 9, 13, 17, 21
//
// Same as above for I2C1 bus:
//
//	SDA: 2, 6, 10, 14, 18, 26
//	SCL: 3, 7, 11, 15, 19, 27
func (i2c *I2C) Configure(config I2CConfig) error {
	const defaultBaud uint32 = 100_000 // 100kHz standard mode
	if config.SCL == 0 && config.SDA == 0 {
		// If config pins are zero valued or clock pin is invalid then we set default values.
		switch i2c.Bus {
		case rp.I2C0:
			config.SCL = I2C0_SCL_PIN
			config.SDA = I2C0_SDA_PIN
		case rp.I2C1:
			config.SCL = I2C1_SCL_PIN
			config.SDA = I2C1_SDA_PIN
		}
	}
	var okSCL, okSDA bool
	switch i2c.Bus {
	case rp.I2C0:
		okSCL = (config.SCL+3)%4 == 0
		okSDA = (config.SDA+4)%4 == 0
	case rp.I2C1:
		okSCL = (config.SCL+1)%4 == 0
		okSDA = (config.SDA+2)%4 == 0
	}

	switch {
	case !okSCL:
		return errInvalidI2CSCL
	case !okSDA:
		return errInvalidI2CSDA
	}

	if config.Frequency == 0 {
		config.Frequency = defaultBaud
	}
	config.SDA.Configure(PinConfig{PinI2C})
	config.SCL.Configure(PinConfig{PinI2C})
	return i2c.init(config)
}

// SetBaudRate sets the I2C frequency. It has the side effect of also
// enabling the I2C hardware if disabled beforehand.
//
//go:inline
func (i2c *I2C) SetBaudRate(br uint32) error {

	if br == 0 {
		return ErrInvalidI2CBaudrate
	}

	// I2C is synchronous design that runs from clk_sys
	freqin := CPUFrequency()

	// TODO there are some subtleties to I2C timing which we are completely ignoring here
	period := (freqin + br/2) / br
	lcnt := period * 3 / 5 // oof this one hurts
	hcnt := period - lcnt
	// Check for out-of-range divisors:
	if hcnt > rp.I2C0_IC_FS_SCL_HCNT_IC_FS_SCL_HCNT_Msk || hcnt < 8 || lcnt > rp.I2C0_IC_FS_SCL_LCNT_IC_FS_SCL_LCNT_Msk || lcnt < 8 {
		return ErrInvalidI2CBaudrate
	}

	// Per I2C-bus specification a device in standard or fast mode must
	// internally provide a hold time of at least 300ns for the SDA signal to
	// bridge the undefined region of the falling edge of SCL. A smaller hold
	// time of 120ns is used for fast mode plus.

	// sda_tx_hold_count = freq_in [cycles/s] * 300ns * (1s / 1e9ns)
	// Reduce 300/1e9 to 3/1e7 to avoid numbers that don't fit in uint.
	// Add 1 to avoid division truncation.
	sdaTxHoldCnt := ((freqin * 3) / 10000000) + 1
	if br >= 1_000_000 {
		// sda_tx_hold_count = freq_in [cycles/s] * 120ns * (1s / 1e9ns)
		// Reduce 120/1e9 to 3/25e6 to avoid numbers that don't fit in uint.
		// Add 1 to avoid division truncation.
		sdaTxHoldCnt = ((freqin * 3) / 25000000) + 1
	}

	if sdaTxHoldCnt > lcnt-2 {
		return ErrInvalidI2CBaudrate
	}
	err := i2c.disable()
	if err != nil {
		return err
	}
	// Always use "fast" mode (<= 400 kHz, works fine for standard mode too)

	i2c.Bus.IC_CON.ReplaceBits(rp.I2C0_IC_CON_SPEED_FAST<<rp.I2C0_IC_CON_SPEED_Pos, rp.I2C0_IC_CON_SPEED_Msk, 0)
	i2c.Bus.IC_FS_SCL_HCNT.Set(hcnt)
	i2c.Bus.IC_FS_SCL_LCNT.Set(lcnt)

	i2c.Bus.IC_FS_SPKLEN.Set(u32max(1, lcnt/16))

	i2c.Bus.IC_SDA_HOLD.ReplaceBits(sdaTxHoldCnt<<rp.I2C0_IC_SDA_HOLD_IC_SDA_TX_HOLD_Pos, rp.I2C0_IC_SDA_HOLD_IC_SDA_TX_HOLD_Msk, 0)
	i2c.enable()
	return nil
}

//go:inline
func (i2c *I2C) enable() {
	i2c.Bus.IC_ENABLE.ReplaceBits(rp.I2C0_IC_ENABLE_ENABLE<<rp.I2C0_IC_ENABLE_ENABLE_Pos, rp.I2C0_IC_ENABLE_ENABLE_Msk, 0)
}

// Implemented as per 4.3.10.3. Disabling DW_apb_i2c section.
//
//go:inline
func (i2c *I2C) disable() error {
	const MAX_T_POLL_COUNT = 64 // 64 us timeout corresponds to around 1000kb/s i2c transfer rate.
	deadline := ticks() + MAX_T_POLL_COUNT
	i2c.Bus.IC_ENABLE.Set(0)
	for i2c.Bus.IC_ENABLE_STATUS.Get()&1 != 0 {
		if ticks() > deadline {
			return ErrRP2040I2CDisable
		}
	}
	return nil
}

//go:inline
func (i2c *I2C) init(config I2CConfig) error {
	i2c.reset()
	if err := i2c.disable(); err != nil {
		return err
	}

	i2c.mode = config.Mode

	// Configure as fast-mode with RepStart support, 7-bit addresses
	mode := uint32(rp.I2C0_IC_CON_SPEED_FAST<<rp.I2C0_IC_CON_SPEED_Pos) |
		rp.I2C0_IC_CON_IC_RESTART_EN | rp.I2C0_IC_CON_TX_EMPTY_CTRL // sets TX_EMPTY_CTRL to enable TX_EMPTY interrupt status
	if config.Mode == I2CModeController {
		mode |= rp.I2C0_IC_CON_MASTER_MODE | rp.I2C0_IC_CON_IC_SLAVE_DISABLE
	}
	i2c.Bus.IC_CON.Set(mode)

	// Set FIFO watermarks to 1 to make things simpler. This is encoded by a register value of 0.
	if config.Mode == I2CModeController {
		i2c.Bus.IC_TX_TL.Set(0)
		i2c.Bus.IC_RX_TL.Set(0)
	}

	// Always enable the DREQ signalling -- harmless if DMA isn't listening
	i2c.Bus.IC_DMA_CR.Set(rp.I2C0_IC_DMA_CR_TDMAE | rp.I2C0_IC_DMA_CR_RDMAE)
	return i2c.SetBaudRate(config.Frequency)
}

// reset sets I2C register RESET bits in the reset peripheral and then clears them.
//
//go:inline
func (i2c *I2C) reset() {
	resetVal := i2c.deinit()
	rp.RESETS.RESET.ClearBits(resetVal)
	// Wait until reset is done.
	for !rp.RESETS.RESET_DONE.HasBits(resetVal) {
	}
}

// deinit sets reset bit for I2C. Must call reset to reenable I2C after deinit.
//
//go:inline
func (i2c *I2C) deinit() (resetVal uint32) {
	switch {
	case i2c.Bus == rp.I2C0:
		resetVal = rp.RESETS_RESET_I2C0
	case i2c.Bus == rp.I2C1:
		resetVal = rp.RESETS_RESET_I2C1
	}
	// Perform I2C reset.
	rp.RESETS.RESET.SetBits(resetVal)

	return resetVal
}

// tx performs blocking write followed by read to I2C bus.
func (i2c *I2C) tx(addr uint8, tx, rx []byte, timeout_us uint64) (err error) {
	deadline := ticks() + timeout_us
	if addr >= 0x80 || isReservedI2CAddr(addr) {
		return ErrInvalidTgtAddr
	}
	txlen := len(tx)
	rxlen := len(rx)
	// Quick return if possible.
	if txlen == 0 && rxlen == 0 {
		return nil
	}

	err = i2c.disable()
	if err != nil {
		return err
	}
	i2c.Bus.IC_TAR.Set(uint32(addr))
	i2c.enable()
	abort := false
	var abortReason i2cAbortError
	txStop := rxlen == 0
	for txCtr := 0; txCtr < txlen; txCtr++ {
		if abort {
			break
		}
		first := txCtr == 0
		last := txCtr == txlen-1 && rxlen == 0
		i2c.Bus.IC_DATA_CMD.Set(
			(boolToBit(first) << rp.I2C0_IC_DATA_CMD_RESTART_Pos) |
				(boolToBit(last && txStop) << rp.I2C0_IC_DATA_CMD_STOP_Pos) |
				uint32(tx[txCtr]))

		// Wait until the transmission of the address/data from the internal
		// shift register has completed. For this to function correctly, the
		// TX_EMPTY_CTRL flag in IC_CON must be set. The TX_EMPTY_CTRL flag
		// was set in i2c_init.

		// IC_RAW_INTR_STAT_TX_EMPTY: This bit is set to 1 when the transmit buffer is at or below
		// the threshold value set in the IC_TX_TL register and the
		// transmission of the address/data from the internal shift
		// register for the most recently popped command is
		// completed. It is automatically cleared by hardware when
		// the buffer level goes above the threshold. When
		// IC_ENABLE[0] is set to 0, the TX FIFO is flushed and held
		// in reset. There the TX FIFO looks like it has no data within
		// it, so this bit is set to 1, provided there is activity in the
		// controller or target state machines. When there is no longer
		// any activity, then with ic_en=0, this bit is set to 0.
		for !i2c.interrupted(rp.I2C0_IC_RAW_INTR_STAT_TX_EMPTY) {
			if ticks() > deadline {
				return errI2CWriteTimeout // If there was a timeout, don't attempt to do anything else.
			}

			gosched()
		}

		abortReason = i2c.getAbortReason()
		if abortReason != 0 {
			i2c.clearAbortReason()
			abort = true
		}
		if abort || last {
			// If the transaction was aborted or if it completed
			// successfully wait until the STOP condition has occurred.

			// TODO Could there be an abort while waiting for the STOP
			// condition here? If so, additional code would be needed here
			// to take care of the abort.
			for !i2c.interrupted(rp.I2C0_IC_RAW_INTR_STAT_STOP_DET) {
				if ticks() > deadline {
					if abort {
						return abortReason
					}
					return errI2CWriteTimeout
				}

				gosched()
			}
			i2c.Bus.IC_CLR_STOP_DET.Get()
		}
	}

	// Midway check for abort. Related issue https://github.com/tinygo-org/tinygo/issues/3671.
	// The root cause for an abort after writing registers was "tx data no ack" (abort code=8).
	// If the abort code was not registered then the whole peripheral would remain in disabled state forever.
	abortReason = i2c.getAbortReason()
	if abortReason != 0 {
		i2c.clearAbortReason()
		abort = true
	}

	rxStart := txlen == 0
	if rxlen > 0 && !abort {
		for rxCtr := 0; rxCtr < rxlen; rxCtr++ {
			first := rxCtr == 0
			last := rxCtr == rxlen-1
			for i2c.writeAvailable() == 0 {
				gosched()
			}
			i2c.Bus.IC_DATA_CMD.Set(
				boolToBit(first && rxStart)<<rp.I2C0_IC_DATA_CMD_RESTART_Pos |
					boolToBit(last)<<rp.I2C0_IC_DATA_CMD_STOP_Pos |
					rp.I2C0_IC_DATA_CMD_CMD) // -> 1 for read

			for !abort && i2c.readAvailable() == 0 {
				abortReason = i2c.getAbortReason()
				if abortReason != 0 {
					i2c.clearAbortReason()
					abort = true
				}
				if ticks() > deadline {
					return errI2CReadTimeout // If there was a timeout, don't attempt to do anything else.
				}

				gosched()
			}
			if abort {
				break
			}
			rx[rxCtr] = uint8(i2c.Bus.IC_DATA_CMD.Get())
		}
	}
	// From Pico SDK: A lot of things could have just happened due to the ingenious and
	// creative design of I2C. Try to figure things out.
	if abort {
		switch {
		case abortReason == 0 || abortReason&rp.I2C0_IC_TX_ABRT_SOURCE_ABRT_7B_ADDR_NOACK != 0:
			// No reported errors - seems to happen if there is nothing connected to the bus.
			// Address byte not acknowledged
			err = ErrI2CGeneric
		case abortReason&rp.I2C0_IC_TX_ABRT_SOURCE_ABRT_TXDATA_NOACK != 0:
			// Address acknowledged, some data not acknowledged
			fallthrough
		default:
			err = abortReason
		}
	}
	return err
}

// listen sets up for async handling of requests on the I2C bus.
func (i2c *I2C) listen(addr uint8) error {
	if addr >= 0x80 || isReservedI2CAddr(addr) {
		return ErrInvalidTgtAddr
	}

	err := i2c.disable()
	if err != nil {
		return err
	}

	i2c.Bus.IC_SAR.Set(uint32(addr))

	i2c.enable()

	return nil
}

func (i2c *I2C) WaitForEvent(buf []byte) (evt I2CTargetEvent, count int, err error) {
	rxPtr := 0
	for {
		stat := i2c.Bus.IC_RAW_INTR_STAT.Get()

		if stat&rp.I2C0_IC_INTR_MASK_M_RX_FULL != 0 {
			b := uint8(i2c.Bus.IC_DATA_CMD.Get())
			if rxPtr < len(buf) {
				buf[rxPtr] = b
				rxPtr++
			}
		}

		// Stop
		if stat&rp.I2C0_IC_INTR_MASK_M_STOP_DET != 0 {
			if rxPtr > 0 {
				return I2CReceive, rxPtr, nil
			}

			i2c.Bus.IC_CLR_STOP_DET.Get() // clear
			return I2CFinish, 0, nil
		}

		// Start or restart - ignore start, return on restart
		if stat&rp.I2C0_IC_INTR_MASK_M_START_DET != 0 {
			i2c.Bus.IC_CLR_START_DET.Get() // clear restart

			// Restart
			if rxPtr > 0 {
				return I2CReceive, rxPtr, nil
			}
		}

		// Read request - leave flag set until we start to reply.
		if stat&rp.I2C0_IC_INTR_MASK_M_RD_REQ != 0 {
			return I2CRequest, 0, nil
		}

		gosched()
	}
}

func (i2c *I2C) Reply(buf []byte) error {
	txPtr := 0

	stat := i2c.Bus.IC_RAW_INTR_STAT.Get()

	if stat&rp.I2C0_IC_INTR_MASK_M_RD_REQ == 0 {
		return ErrI2CWrongMode
	}
	i2c.Bus.IC_CLR_RD_REQ.Get() // clear restart

	// Clear any dangling TX abort
	if stat&rp.I2C0_IC_INTR_MASK_M_TX_ABRT != 0 {
		i2c.Bus.IC_CLR_TX_ABRT.Get()
	}

	for txPtr < len(buf) {
		if i2c.Bus.GetIC_RAW_INTR_STAT_TX_EMPTY() != 0 {
			i2c.Bus.SetIC_DATA_CMD_DAT(uint32(buf[txPtr]))
			txPtr++
			// The DW_apb_i2c flushes/resets/empties the
			// TX_FIFO and RX_FIFO whenever there is a transmit abort
			// caused by any of the events tracked by the
			// IC_TX_ABRT_SOURCE register.
			// In other words, it's safe to block until TX FIFO is
			// EMPTY--it will empty from being transmitted or on error.
			for i2c.Bus.GetIC_RAW_INTR_STAT_TX_EMPTY() == 0 {
			}
		}

		// This Tx abort is a normal case - we're sending more
		// data than controller wants to receive
		if i2c.Bus.GetIC_RAW_INTR_STAT_TX_ABRT() != 0 {
			i2c.Bus.GetIC_CLR_TX_ABRT_CLR_TX_ABRT()
			return nil
		}

		gosched()
	}

	return nil
}

// writeAvailable determines non-blocking write space available
//
//go:inline
func (i2c *I2C) writeAvailable() uint32 {
	return rp.I2C0_IC_COMP_PARAM_1_TX_BUFFER_DEPTH_Pos - i2c.Bus.IC_TXFLR.Get()
}

// readAvailable determines number of bytes received
//
//go:inline
func (i2c *I2C) readAvailable() uint32 {
	return i2c.Bus.IC_RXFLR.Get()
}

// Equivalent to IC_CLR_TX_ABRT.Get() (side effect clears ABORT_REASON)
//
//go:inline
func (i2c *I2C) clearAbortReason() {
	// Note clearing the abort flag also clears the reason, and
	// this instance of flag is clear-on-read! Note also the
	// IC_CLR_TX_ABRT register always reads as 0.
	i2c.Bus.IC_CLR_TX_ABRT.Get()
}

// getAbortReason reads IC_TX_ABRT_SOURCE register.
//
//go:inline
func (i2c *I2C) getAbortReason() i2cAbortError {
	return i2cAbortError(i2c.Bus.IC_TX_ABRT_SOURCE.Get())
}

// returns true if RAW_INTR_STAT bits in mask are all set. performs:
//
//	RAW_INTR_STAT & mask == mask
//
//go:inline
func (i2c *I2C) interrupted(mask uint32) bool {
	reg := i2c.Bus.IC_RAW_INTR_STAT.Get()
	return reg&mask == mask
}

type i2cAbortError uint32

func (b i2cAbortError) Error() string {
	return "i2c abort, reason " + itoa.Uitoa(uint(b))
}

func (b i2cAbortError) Reasons() (reasons []string) {
	if b == 0 {
		return nil
	}
	if b&rp.I2C0_IC_TX_ABRT_SOURCE_ABRT_7B_ADDR_NOACK != 0 {
		reasons = append(reasons, "7-bit address no ack")
	}
	if b&rp.I2C0_IC_TX_ABRT_SOURCE_ABRT_10ADDR1_NOACK != 0 {
		reasons = append(reasons, "10-bit address first byte no ack")
	}
	if b&rp.I2C0_IC_TX_ABRT_SOURCE_ABRT_10ADDR2_NOACK != 0 {
		reasons = append(reasons, "10-bit address second byte no ack")
	}
	if b&rp.I2C0_IC_TX_ABRT_SOURCE_ABRT_TXDATA_NOACK != 0 {
		reasons = append(reasons, "tx data no ack")
	}
	if b&rp.I2C0_IC_TX_ABRT_SOURCE_ABRT_GCALL_NOACK != 0 {
		reasons = append(reasons, "general call no ack")
	}
	if b&rp.I2C0_IC_TX_ABRT_SOURCE_ABRT_GCALL_READ != 0 {
		reasons = append(reasons, "general call read")
	}
	if b&rp.I2C0_IC_TX_ABRT_SOURCE_ABRT_HS_ACKDET != 0 {
		reasons = append(reasons, "high speed ack detect")
	}
	if b&rp.I2C0_IC_TX_ABRT_SOURCE_ABRT_SBYTE_ACKDET != 0 {
		reasons = append(reasons, "start byte ack detect")
	}
	if b&rp.I2C0_IC_TX_ABRT_SOURCE_ABRT_HS_NORSTRT != 0 {
		reasons = append(reasons, "high speed no restart")
	}
	if b&rp.I2C0_IC_TX_ABRT_SOURCE_ABRT_SBYTE_NORSTRT != 0 {
		reasons = append(reasons, "start byte no restart")
	}
	if b&rp.I2C0_IC_TX_ABRT_SOURCE_ABRT_10B_RD_NORSTRT != 0 {
		reasons = append(reasons, "10-bit read no restart")
	}
	if b&rp.I2C0_IC_TX_ABRT_SOURCE_ABRT_MASTER_DIS != 0 {
		reasons = append(reasons, "master disabled")
	}
	if b&rp.I2C0_IC_TX_ABRT_SOURCE_ARB_LOST != 0 {
		reasons = append(reasons, "arbitration lost")
	}
	if b&rp.I2C0_IC_TX_ABRT_SOURCE_ABRT_SLVFLUSH_TXFIFO != 0 {
		reasons = append(reasons, "slave flush tx fifo")
	}
	if b&rp.I2C0_IC_TX_ABRT_SOURCE_ABRT_SLV_ARBLOST != 0 {
		reasons = append(reasons, "slave arbitration lost")
	}
	if b&rp.I2C0_IC_TX_ABRT_SOURCE_ABRT_SLVRD_INTX != 0 {
		reasons = append(reasons, "slave read while inactive")
	}
	if b&rp.I2C0_IC_TX_ABRT_SOURCE_ABRT_USER_ABRT != 0 {
		reasons = append(reasons, "user abort")
	}
	return reasons
}

//go:inline
func boolToBit(a bool) uint32 {
	if a {
		return 1
	}
	return 0
}

//go:inline
func u32max(a, b uint32) uint32 {
	if a > b {
		return a
	}
	return b
}

//go:inline
func isReservedI2CAddr(addr uint8) bool {
	return (addr&0x78) == 0 || (addr&0x78) == 0x78
}
