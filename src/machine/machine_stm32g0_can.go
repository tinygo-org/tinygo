//go:build stm32g0b1

package machine

import (
	"device/stm32"
	"errors"
	"runtime/interrupt"
	"unsafe"
)

// Exported API in src/machine/can.go

// FDCAN Message RAM configuration
// STM32G0B1 SRAMCAN base address: 0x4000B400
// Each FDCAN instance has its own message RAM area
const (
	sramcanBase = 0x4000B400

	// Message RAM layout sizes (matching STM32 HAL)
	sramcanFLSNbr = 28 // Max. Filter List Standard Number
	sramcanFLENbr = 8  // Max. Filter List Extended Number
	sramcanRF0Nbr = 3  // RX FIFO 0 Elements Number
	sramcanRF1Nbr = 3  // RX FIFO 1 Elements Number
	sramcanTEFNbr = 3  // TX Event FIFO Elements Number
	sramcanTFQNbr = 3  // TX FIFO/Queue Elements Number

	// Element sizes in bytes
	sramcanFLSSize = 1 * 4  // Filter Standard Element Size
	sramcanFLESize = 2 * 4  // Filter Extended Element Size
	sramcanRF0Size = 18 * 4 // RX FIFO 0 Element Size (for 64-byte data)
	sramcanRF1Size = 18 * 4 // RX FIFO 1 Element Size
	sramcanTEFSize = 2 * 4  // TX Event FIFO Element Size
	sramcanTFQSize = 18 * 4 // TX FIFO/Queue Element Size

	// Start addresses (offsets from base)
	sramcanFLSSA = 0
	sramcanFLESA = sramcanFLSSA + (sramcanFLSNbr * sramcanFLSSize)
	sramcanRF0SA = sramcanFLESA + (sramcanFLENbr * sramcanFLESize)
	sramcanRF1SA = sramcanRF0SA + (sramcanRF0Nbr * sramcanRF0Size)
	sramcanTEFSA = sramcanRF1SA + (sramcanRF1Nbr * sramcanRF1Size)
	sramcanTFQSA = sramcanTEFSA + (sramcanTEFNbr * sramcanTEFSize)
	sramcanSize  = sramcanTFQSA + (sramcanTFQNbr * sramcanTFQSize)
)

// FDCAN element masks (for parsing message RAM)
const (
	fdcanElementMaskSTDID = 0x1FFC0000 // Standard Identifier
	fdcanElementMaskEXTID = 0x1FFFFFFF // Extended Identifier
	fdcanElementMaskRTR   = 0x20000000 // Remote Transmission Request
	fdcanElementMaskXTD   = 0x40000000 // Extended Identifier flag
	fdcanElementMaskESI   = 0x80000000 // Error State Indicator
	fdcanElementMaskTS    = 0x0000FFFF // Timestamp
	fdcanElementMaskDLC   = 0x000F0000 // Data Length Code
	fdcanElementMaskBRS   = 0x00100000 // Bit Rate Switch
	fdcanElementMaskFDF   = 0x00200000 // FD Format
	fdcanElementMaskEFC   = 0x00800000 // Event FIFO Control
	fdcanElementMaskMM    = 0xFF000000 // Message Marker
	fdcanElementMaskFIDX  = 0x7F000000 // Filter Index
	fdcanElementMaskANMF  = 0x80000000 // Accepted Non-matching Frame
)

// Interrupt flags
const (
	FDCAN_IT_RX_FIFO0_NEW_MESSAGE = 0x00000001
	FDCAN_IT_RX_FIFO0_FULL        = 0x00000002
	FDCAN_IT_RX_FIFO0_MSG_LOST    = 0x00000004
	FDCAN_IT_RX_FIFO1_NEW_MESSAGE = 0x00000010
	FDCAN_IT_RX_FIFO1_FULL        = 0x00000020
	FDCAN_IT_RX_FIFO1_MSG_LOST    = 0x00000040
	FDCAN_IT_TX_COMPLETE          = 0x00000200
	FDCAN_IT_TX_ABORT_COMPLETE    = 0x00000400
	FDCAN_IT_TX_FIFO_EMPTY        = 0x00000800
	FDCAN_IT_BUS_OFF              = 0x02000000
	FDCAN_IT_ERROR_WARNING        = 0x01000000
	FDCAN_IT_ERROR_PASSIVE        = 0x00800000
)

// CAN is a STM32G0's CAN/FDCAN peripheral.
type CAN struct {
	Bus             *stm32.FDCAN_Type
	TxAltFuncSelect uint8
	RxAltFuncSelect uint8
	Interrupt       interrupt.Interrupt
	instance        uint8
	alwaysFD        bool
	rxInterrupt     bool
}

// CANTransferRate represents CAN bus transfer rates
type CANTransferRate uint32

const (
	FDCANTransferRate125kbps  CANTransferRate = 125000
	FDCANTransferRate250kbps  CANTransferRate = 250000
	FDCANTransferRate500kbps  CANTransferRate = 500000
	FDCANTransferRate1000kbps CANTransferRate = 1000000
	FDCANTransferRate2000kbps CANTransferRate = 2000000 // FD only
	FDCANTransferRate4000kbps CANTransferRate = 4000000 // FD only
)

// CANMode represents the FDCAN operating mode
type CANMode uint8

const (
	CANModeNormal           CANMode = 0
	CANModeBusMonitoring    CANMode = 1
	CANModeInternalLoopback CANMode = 2
	CANModeExternalLoopback CANMode = 3
)

// CANConfig holds FDCAN configuration parameters
type CANConfig struct {
	TransferRate      CANTransferRate // Nominal bit rate (arbitration phase)
	TransferRateFD    CANTransferRate // Data bit rate (data phase), must be >= TransferRate
	Mode              CANMode
	Tx                Pin
	Rx                Pin
	Standby           Pin  // Optional standby pin for CAN transceiver (set to NoPin if not used)
	AlwaysFD          bool // Always transmit as FD frames, even when data fits in classic CAN
	EnableRxInterrupt bool // Enable interrupt-driven receive (messages delivered via SetRxCallback)
}

// CANFilterConfig represents a message filter configuration
type CANFilterConfig struct {
	Index        uint8  // Filter index (0-27 for standard, 0-7 for extended)
	Type         uint8  // 0=Range, 1=Dual, 2=Classic (ID/Mask)
	Config       uint8  // 0=Disable, 1=FIFO0, 2=FIFO1, 3=Reject
	ID1          uint32 // First ID or filter
	ID2          uint32 // Second ID or mask
	IsExtendedID bool   // true for 29-bit ID, false for 11-bit
}

var (
	errCANInvalidTransferRate   = errors.New("CAN: invalid TransferRate")
	errCANInvalidTransferRateFD = errors.New("CAN: invalid TransferRateFD")
	errCANTimeout               = errors.New("CAN: timeout")
	errCANTxFifoFull            = errors.New("CAN: Tx FIFO full")
)

// enableFDCANClock enables the FDCAN peripheral clock
func enableFDCANClock() {
	// FDCAN clock is on APB1
	stm32.RCC.SetAPBENR1_FDCANEN(1)
}

// flags implemented as described in [CAN.SetRxCallback]
var canRxCB [2]canRxCallback

// canInstances tracks CAN peripherals with interrupt-driven RX enabled.
// A non-nil entry means setRxCallback was called with a non-nil callback.
var canInstances [2]*CAN

// Configure initializes the FDCAN peripheral and starts it.
func (can *CAN) Configure(config CANConfig) error {
	can.alwaysFD = config.AlwaysFD

	if config.Standby != NoPin {
		config.Standby.Configure(PinConfig{Mode: PinOutput})
		config.Standby.Low()
	}

	enableFDCANClock()

	config.Tx.ConfigureAltFunc(PinConfig{Mode: PinOutput}, can.TxAltFuncSelect)
	config.Rx.ConfigureAltFunc(PinConfig{Mode: PinInputFloating}, can.RxAltFuncSelect)

	// Exit sleep mode.
	can.Bus.SetCCCR_CSR(0)
	timeout := 10000
	for can.Bus.GetCCCR_CSA() != 0 {
		timeout--
		if timeout == 0 {
			return errCANTimeout
		}
	}

	// Request initialization.
	can.Bus.SetCCCR_INIT(1)
	timeout = 10000
	for can.Bus.GetCCCR_INIT() == 0 {
		timeout--
		if timeout == 0 {
			return errCANTimeout
		}
	}

	// Enable configuration change.
	can.Bus.SetCCCR_CCE(1)

	if can.Bus == stm32.FDCAN1 {
		can.Bus.SetCKDIV_PDIV(0) // No clock division.
	}

	can.Bus.SetCCCR_DAR(0)  // Enable auto retransmission.
	can.Bus.SetCCCR_TXP(0)  // Disable transmit pause.
	can.Bus.SetCCCR_PXHD(0) // Enable protocol exception handling.
	can.Bus.SetCCCR_FDOE(1) // FD operation.
	can.Bus.SetCCCR_BRSE(1) // Bit rate switching.

	// Reset mode bits, then apply requested mode.
	can.Bus.SetCCCR_TEST(0)
	can.Bus.SetCCCR_MON(0)
	can.Bus.SetCCCR_ASM(0)
	can.Bus.SetTEST_LBCK(0)
	switch config.Mode {
	case CANModeBusMonitoring:
		can.Bus.SetCCCR_MON(1)
	case CANModeInternalLoopback:
		can.Bus.SetCCCR_TEST(1)
		can.Bus.SetCCCR_MON(1)
		can.Bus.SetTEST_LBCK(1)
	case CANModeExternalLoopback:
		can.Bus.SetCCCR_TEST(1)
		can.Bus.SetTEST_LBCK(1)
	}

	// Nominal bit timing (64 MHz FDCAN clock, 16 tq/bit, ~80% sample point).
	if config.TransferRate == 0 {
		config.TransferRate = FDCANTransferRate500kbps
	}
	nbrp, ntseg1, ntseg2, nsjw, err := fdcanNominalBitTiming(config.TransferRate)
	if err != nil {
		return err
	}
	can.Bus.NBTP.Set(((nsjw - 1) << 25) | ((nbrp - 1) << 16) | ((ntseg1 - 1) << 8) | (ntseg2 - 1))

	// Data bit timing (FD phase).
	if config.TransferRateFD == 0 {
		config.TransferRateFD = FDCANTransferRate1000kbps
	}
	if config.TransferRateFD < config.TransferRate {
		return errCANInvalidTransferRateFD
	}
	dbrp, dtseg1, dtseg2, dsjw, err := fdcanDataBitTiming(config.TransferRateFD)
	if err != nil {
		return err
	}
	can.Bus.DBTP.Set(((dbrp - 1) << 16) | ((dtseg1 - 1) << 8) | ((dtseg2 - 1) << 4) | (dsjw - 1))

	// Enable timestamp counter (internal, prescaler=1).
	can.Bus.TSCC.Set(1)

	// Clear message RAM.
	base := can.sramBase()
	for addr := base; addr < base+sramcanSize; addr += 4 {
		*(*uint32)(unsafe.Pointer(addr)) = 0
	}

	// Set filter list sizes: LSS[20:16], LSE[27:24].
	rxgfc := can.Bus.RXGFC.Get()
	rxgfc &= ^uint32(0x0F1F0000)
	rxgfc |= uint32(sramcanFLSNbr) << 16
	rxgfc |= uint32(sramcanFLENbr) << 24
	can.Bus.RXGFC.Set(rxgfc)

	// Start peripheral.
	can.Bus.SetCCCR_CCE(0)
	can.Bus.SetCCCR_INIT(0)
	timeout = 10000
	for can.Bus.GetCCCR_INIT() != 0 {
		timeout--
		if timeout == 0 {
			return errCANTimeout
		}
	}

	return nil
}

// Stop puts the FDCAN peripheral back into initialization mode.
func (can *CAN) Stop() error {
	can.Bus.SetCCCR_INIT(1)
	timeout := 10000
	for can.Bus.GetCCCR_INIT() == 0 {
		timeout--
		if timeout == 0 {
			return errCANTimeout
		}
	}
	can.Bus.SetCCCR_CCE(1)
	return nil
}

// txFIFOLevel implements [CAN.TxFIFOLevel].
func (can *CAN) txFIFOLevel() (int, int) {
	free := int(can.Bus.TXFQS.Get() & 0x07) // TFFL[2:0]
	return sramcanTFQNbr - free, sramcanTFQNbr
}

// tx implements [CAN.Tx].
func (can *CAN) tx(id canID, flags canFlags, data []byte) error {
	if can.Bus.TXFQS.Get()&0x00200000 != 0 { // TFQF bit
		return errCANTxFifoFull
	}

	length := byte(len(data))
	if length > 64 {
		length = 64
	}

	// Use FD framing if configured to always use FD, or if data exceeds classic CAN max.
	isFD := flags&canFlagFDF != 0 || length > 8

	putIndex := (can.Bus.TXFQS.Get() >> 16) & 0x03 // TFQPI[1:0]
	txAddr := can.sramBase() + sramcanTFQSA + uintptr(putIndex)*sramcanTFQSize

	// Header word 1: identifier and flags.
	var w1 uint32
	if flags&canFlagESI != 0 {
		w1 = (id & 0x1FFFFFFF) | fdcanElementMaskXTD
	} else {
		w1 = (id & 0x7FF) << 18
	}

	// Header word 2: DLC, FD/BRS flags.
	dlc := lengthToDLC(length)
	w2 := uint32(dlc) << 16
	if isFD {
		w2 |= fdcanElementMaskFDF | fdcanElementMaskBRS
	}

	*(*uint32)(unsafe.Pointer(txAddr)) = w1
	*(*uint32)(unsafe.Pointer(txAddr + 4)) = w2

	// Copy data with 32-bit word access (Cortex-M0+).
	for w := byte(0); w < (length+3)/4; w++ {
		var word uint32
		base := w * 4
		for b := byte(0); b < 4 && base+b < length; b++ {
			word |= uint32(data[base+b]) << (b * 8)
		}
		*(*uint32)(unsafe.Pointer(txAddr + 8 + uintptr(w)*4)) = word
	}

	can.Bus.TXBAR.Set(1 << putIndex)
	return nil
}

// rxFIFOLevel implements [CAN.RxFIFOLevel].
// Returns 0,0 when interrupt-driven (messages delivered via callback).
func (can *CAN) rxFIFOLevel() (int, int) {
	if canInstances[can.instance] != nil {
		return 0, 0
	}
	level := int(can.Bus.RXF0S.Get() & 0x0F) // F0FL[3:0]
	return level, sramcanRF0Nbr
}

// setRxCallback implements [CAN.SetRxCallback].
// When cb is non-nil, interrupt-driven receive is enabled on RX FIFO 0.
// The CAN.Interrupt field must be initialized with interrupt.New in the board file.
func (can *CAN) setRxCallback(cb canRxCallback) {
	canRxCB[can.instance] = cb
	if cb != nil {
		canInstances[can.instance] = can
		// Enable RX FIFO 0 new message interrupt, routed to interrupt line 0.
		can.Bus.SetIE_RF0NE(1)
		can.Bus.SetILS_RxFIFO0(0)
		can.Bus.SetILE_EINT0(1)
		can.Interrupt.Enable()
	} else {
		can.Bus.SetIE_RF0NE(0)
		canInstances[can.instance] = nil
	}
}

// rxPoll implements [CAN.RxPoll].
// No-op when interrupt-driven receive is active.
func (can *CAN) rxPoll() error {
	if canInstances[can.instance] != nil {
		return nil
	}
	cb := canRxCB[can.instance]
	if cb == nil {
		return nil
	}
	processRxFIFO0(can, cb)
	return nil
}

// processRxFIFO0 drains RX FIFO 0 and delivers each message to cb.
// Used by both rxPoll (poll mode) and canHandleInterrupt (interrupt mode).
func processRxFIFO0(can *CAN, cb canRxCallback) {
	for can.Bus.RXF0S.Get()&0x0F != 0 {
		getIndex := (can.Bus.RXF0S.Get() >> 8) & 0x03 // F0GI[1:0]
		rxAddr := can.sramBase() + sramcanRF0SA + uintptr(getIndex)*sramcanRF0Size

		w1 := *(*uint32)(unsafe.Pointer(rxAddr))
		w2 := *(*uint32)(unsafe.Pointer(rxAddr + 4))

		extendedID := w1&fdcanElementMaskXTD != 0
		var id uint32
		var flags uint32
		if extendedID {
			flags |= canFlagIDE
			id = w1 & fdcanElementMaskEXTID
		} else {
			id = (w1 & fdcanElementMaskSTDID) >> 18
		}

		timestamp := w2 & fdcanElementMaskTS
		dlc := byte((w2 & fdcanElementMaskDLC) >> 16)
		isFD := w2&fdcanElementMaskFDF != 0

		if isFD {
			flags |= canFlagFDF
		}
		if w1&fdcanElementMaskRTR != 0 {
			flags |= canFlagRTR
		}
		if w2&fdcanElementMaskBRS != 0 {
			flags |= canFlagBRS
		}
		if w1&fdcanElementMaskESI != 0 {
			flags |= canFlagESI
		}

		dataLen := dlcToLength(dlc)
		if !isFD && dataLen > 8 {
			dataLen = 8
		}
		var buf [64]byte
		for w := byte(0); w < (dataLen+3)/4; w++ {
			word := *(*uint32)(unsafe.Pointer(rxAddr + 8 + uintptr(w)*4))
			base := w * 4
			for b := byte(0); b < 4 && base+b < dataLen; b++ {
				buf[base+b] = byte(word >> (b * 8))
			}
		}

		// Acknowledge before callback so the FIFO slot is freed.
		can.Bus.RXF0A.Set(uint32(getIndex))
		cb(buf[:dataLen], id, timestamp, flags)
	}
}

// canHandleInterrupt is the shared interrupt handler for FDCAN interrupt line 0 (IRQ_TIM16).
// Both FDCAN1 and FDCAN2 share this IRQ vector.
func canHandleInterrupt(interrupt.Interrupt) {
	for i := range canInstances {
		can := canInstances[i]
		if can == nil {
			continue
		}
		ir := can.Bus.IR.Get()
		if ir&FDCAN_IT_RX_FIFO0_NEW_MESSAGE != 0 {
			can.Bus.IR.Set(FDCAN_IT_RX_FIFO0_NEW_MESSAGE) // Write 1 to clear
			if cb := canRxCB[i]; cb != nil {
				processRxFIFO0(can, cb)
			}
		}
	}
}

// ConfigureFilter configures a message acceptance filter.
func (can *CAN) ConfigureFilter(config CANFilterConfig) error {
	base := can.sramBase()

	if config.IsExtendedID {
		if config.Index >= sramcanFLENbr {
			return errors.New("CAN: filter index out of range")
		}

		filterAddr := base + sramcanFLESA + (uintptr(config.Index) * sramcanFLESize)

		w1 := (uint32(config.Config) << 29) | (config.ID1 & 0x1FFFFFFF)
		w2 := (uint32(config.Type) << 30) | (config.ID2 & 0x1FFFFFFF)

		*(*uint32)(unsafe.Pointer(filterAddr)) = w1
		*(*uint32)(unsafe.Pointer(filterAddr + 4)) = w2
	} else {
		if config.Index >= sramcanFLSNbr {
			return errors.New("CAN: filter index out of range")
		}

		filterAddr := base + sramcanFLSSA + (uintptr(config.Index) * sramcanFLSSize)

		w := (uint32(config.Type) << 30) |
			(uint32(config.Config) << 27) |
			((config.ID1 & 0x7FF) << 16) |
			(config.ID2 & 0x7FF)

		*(*uint32)(unsafe.Pointer(filterAddr)) = w
	}

	return nil
}

func (can *CAN) sramBase() uintptr {
	if can.Bus == stm32.FDCAN2 {
		return uintptr(sramcanBase) + sramcanSize
	}
	return uintptr(sramcanBase)
}

// fdcanNominalBitTiming returns prescaler and segment values for the nominal (arbitration) phase.
// STM32G0 FDCAN clock = 64 MHz, 16 time quanta per bit, ~80% sample point.
func fdcanNominalBitTiming(rate CANTransferRate) (brp, tseg1, tseg2, sjw uint32, err error) {
	switch rate {
	case FDCANTransferRate125kbps:
		return 32, 13, 2, 4, nil
	case FDCANTransferRate250kbps:
		return 16, 13, 2, 4, nil
	case FDCANTransferRate500kbps:
		return 8, 13, 2, 4, nil
	case FDCANTransferRate1000kbps:
		return 4, 13, 2, 4, nil
	default:
		return 0, 0, 0, 0, errCANInvalidTransferRate
	}
}

// fdcanDataBitTiming returns prescaler and segment values for the data phase (FD).
func fdcanDataBitTiming(rate CANTransferRate) (brp, tseg1, tseg2, sjw uint32, err error) {
	switch rate {
	case FDCANTransferRate125kbps:
		return 32, 13, 2, 4, nil
	case FDCANTransferRate250kbps:
		return 16, 13, 2, 4, nil
	case FDCANTransferRate500kbps:
		return 8, 13, 2, 4, nil
	case FDCANTransferRate1000kbps:
		return 4, 13, 2, 4, nil
	case FDCANTransferRate2000kbps:
		return 2, 13, 2, 4, nil
	case FDCANTransferRate4000kbps:
		return 1, 13, 2, 4, nil
	default:
		return 0, 0, 0, 0, errCANInvalidTransferRateFD
	}
}
