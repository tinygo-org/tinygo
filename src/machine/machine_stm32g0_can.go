//go:build stm32g0b1

package machine

import (
	"device/stm32"
	"errors"
	"runtime/interrupt"
	"unsafe"
)

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

// FDCAN represents an FDCAN peripheral
type FDCAN struct {
	Bus             *stm32.FDCAN_Type
	TxAltFuncSelect uint8
	RxAltFuncSelect uint8
	Interrupt       interrupt.Interrupt
	instance        uint8
}

// FDCANTransferRate represents CAN bus transfer rates
type FDCANTransferRate uint32

const (
	FDCANTransferRate125kbps  FDCANTransferRate = 125000
	FDCANTransferRate250kbps  FDCANTransferRate = 250000
	FDCANTransferRate500kbps  FDCANTransferRate = 500000
	FDCANTransferRate1000kbps FDCANTransferRate = 1000000
	FDCANTransferRate2000kbps FDCANTransferRate = 2000000 // FD only
	FDCANTransferRate4000kbps FDCANTransferRate = 4000000 // FD only
)

// FDCANMode represents the FDCAN operating mode
type FDCANMode uint8

const (
	FDCANModeNormal           FDCANMode = 0
	FDCANModeBusMonitoring    FDCANMode = 1
	FDCANModeInternalLoopback FDCANMode = 2
	FDCANModeExternalLoopback FDCANMode = 3
)

// FDCANConfig holds FDCAN configuration parameters
type FDCANConfig struct {
	TransferRate   FDCANTransferRate // Nominal bit rate (arbitration phase)
	TransferRateFD FDCANTransferRate // Data bit rate (data phase), must be >= TransferRate
	Mode           FDCANMode
	Tx             Pin
	Rx             Pin
	Standby        Pin // Optional standby pin for CAN transceiver (set to NoPin if not used)
}

// FDCANTxBufferElement represents a transmit buffer element
type FDCANTxBufferElement struct {
	ESI bool     // Error State Indicator
	XTD bool     // Extended ID flag
	RTR bool     // Remote Transmission Request
	ID  uint32   // CAN identifier (11-bit or 29-bit)
	MM  uint8    // Message Marker
	EFC bool     // Event FIFO Control
	FDF bool     // FD Frame indicator
	BRS bool     // Bit Rate Switch
	DLC uint8    // Data Length Code (0-15)
	DB  [64]byte // Data buffer
}

// FDCANRxBufferElement represents a receive buffer element
type FDCANRxBufferElement struct {
	ESI  bool     // Error State Indicator
	XTD  bool     // Extended ID flag
	RTR  bool     // Remote Transmission Request
	ID   uint32   // CAN identifier
	ANMF bool     // Accepted Non-matching Frame
	FIDX uint8    // Filter Index
	FDF  bool     // FD Frame
	BRS  bool     // Bit Rate Switch
	DLC  uint8    // Data Length Code
	RXTS uint16   // RX Timestamp
	DB   [64]byte // Data buffer
}

// FDCANFilterConfig represents a filter configuration
type FDCANFilterConfig struct {
	Index        uint8  // Filter index (0-27 for standard, 0-7 for extended)
	Type         uint8  // 0=Range, 1=Dual, 2=Classic (ID/Mask)
	Config       uint8  // 0=Disable, 1=FIFO0, 2=FIFO1, 3=Reject
	ID1          uint32 // First ID or filter
	ID2          uint32 // Second ID or mask
	IsExtendedID bool   // true for 29-bit ID, false for 11-bit
}

var (
	errFDCANInvalidTransferRate   = errors.New("FDCAN: invalid TransferRate")
	errFDCANInvalidTransferRateFD = errors.New("FDCAN: invalid TransferRateFD")
	errFDCANTimeout               = errors.New("FDCAN: timeout")
	errFDCANTxFifoFull            = errors.New("FDCAN: Tx FIFO full")
	errFDCANRxFifoEmpty           = errors.New("FDCAN: Rx FIFO empty")
	errFDCANNotStarted            = errors.New("FDCAN: not started")
)

// DLC to bytes lookup table
var dlcToBytes = [16]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 12, 16, 20, 24, 32, 48, 64}

// Configure initializes the FDCAN peripheral
func (can *FDCAN) Configure(config FDCANConfig) error {
	// Configure standby pin if specified (for CAN transceivers with standby control)
	// Setting it low enables the transceiver
	if config.Standby != NoPin {
		config.Standby.Configure(PinConfig{Mode: PinOutput})
		config.Standby.Low()
	}

	// Enable FDCAN clock
	enableFDCANClock()

	// Configure TX and RX pins
	config.Tx.ConfigureAltFunc(PinConfig{Mode: PinOutput}, can.TxAltFuncSelect)
	config.Rx.ConfigureAltFunc(PinConfig{Mode: PinInputFloating}, can.RxAltFuncSelect)

	// Exit from sleep mode
	can.Bus.SetCCCR_CSR(0)

	// Wait for sleep mode exit
	timeout := 10000
	for can.Bus.GetCCCR_CSA() != 0 {
		timeout--
		if timeout == 0 {
			return errFDCANTimeout
		}
	}

	// Request initialization
	can.Bus.SetCCCR_INIT(1)

	// Wait for init mode
	timeout = 10000
	for can.Bus.GetCCCR_INIT() == 0 {
		timeout--
		if timeout == 0 {
			return errFDCANTimeout
		}
	}

	// Enable configuration change
	can.Bus.SetCCCR_CCE(1)

	// Configure clock divider (only for FDCAN1)
	if can.Bus == stm32.FDCAN1 {
		can.Bus.SetCKDIV_PDIV(0)
		//can.Bus.CKDIV.Set(0) // No division
	}

	// Enable automatic retransmission
	can.Bus.SetCCCR_DAR(0)

	// Disable transmit pause
	can.Bus.SetCCCR_TXP(0)

	// Enable protocol exception handling
	can.Bus.SetCCCR_PXHD(0)

	// Enable FD mode with bit rate switching
	can.Bus.SetCCCR_FDOE(1)
	can.Bus.SetCCCR_BRSE(1)

	// Configure operating mode
	can.Bus.SetCCCR_TEST(0)
	can.Bus.SetCCCR_MON(0)
	can.Bus.SetCCCR_ASM(0)
	can.Bus.SetTEST_LBCK(0)

	switch config.Mode {
	case FDCANModeBusMonitoring:
		can.Bus.SetCCCR_MON(1)
	case FDCANModeInternalLoopback:
		can.Bus.SetCCCR_TEST(1)
		can.Bus.SetCCCR_MON(1)
		can.Bus.SetTEST_LBCK(1)
	case FDCANModeExternalLoopback:
		can.Bus.SetCCCR_TEST(1)
		can.Bus.SetTEST_LBCK(1)
	}

	// Set nominal bit timing
	// STM32G0 runs at 64MHz, FDCAN clock = PCLK = 64MHz
	// Bit time = (1 + NTSEG1 + NTSEG2) * tq
	// tq = (NBRP + 1) / fCAN_CLK
	if config.TransferRate == 0 {
		config.TransferRate = FDCANTransferRate500kbps
	}

	nbrp, ntseg1, ntseg2, nsjw, err := can.calculateNominalBitTiming(config.TransferRate)
	if err != nil {
		return err
	}
	can.Bus.NBTP.Set(((nsjw - 1) << 25) | ((nbrp - 1) << 16) | ((ntseg1 - 1) << 8) | (ntseg2 - 1))

	// Set data bit timing (for FD mode)
	if config.TransferRateFD == 0 {
		config.TransferRateFD = FDCANTransferRate1000kbps
	}
	if config.TransferRateFD < config.TransferRate {
		return errFDCANInvalidTransferRateFD
	}

	dbrp, dtseg1, dtseg2, dsjw, err := can.calculateDataBitTiming(config.TransferRateFD)
	if err != nil {
		return err
	}
	can.Bus.DBTP.Set(((dbrp - 1) << 16) | ((dtseg1 - 1) << 8) | ((dtseg2 - 1) << 4) | (dsjw - 1))

	// Configure message RAM
	can.configureMessageRAM()

	return nil
}

// Start enables the FDCAN peripheral for communication
func (can *FDCAN) Start() error {
	// Disable configuration change
	can.Bus.SetCCCR_CCE(0)

	// Exit initialization mode
	can.Bus.SetCCCR_INIT(0)

	// Wait for normal operation
	timeout := 10000

	for can.Bus.GetCCCR_INIT() != 0 {
		timeout--
		if timeout == 0 {
			return errFDCANTimeout
		}
	}

	return nil
}

// Stop disables the FDCAN peripheral
func (can *FDCAN) Stop() error {
	// Request initialization
	can.Bus.SetCCCR_INIT(1)

	// Wait for init mode
	timeout := 10000
	for can.Bus.GetCCCR_INIT() == 0 {
		timeout--
		if timeout == 0 {
			return errFDCANTimeout
		}
	}

	// Enable configuration change
	can.Bus.SetCCCR_CCE(1)

	return nil
}

// TxFifoIsFull returns true if the TX FIFO is full
func (can *FDCAN) TxFifoIsFull() bool {
	return (can.Bus.TXFQS.Get() & 0x00200000) != 0 // TFQF bit
}

// TxFifoFreeLevel returns the number of free TX FIFO elements
func (can *FDCAN) TxFifoFreeLevel() int {
	return int(can.Bus.TXFQS.Get() & 0x07) // TFFL[2:0]
}

// RxFifoSize returns the number of messages in RX FIFO 0
func (can *FDCAN) RxFifoSize() int {
	return int(can.Bus.RXF0S.Get() & 0x0F) // F0FL[3:0]
}

// RxFifoIsEmpty returns true if RX FIFO 0 is empty
func (can *FDCAN) RxFifoIsEmpty() bool {
	return (can.Bus.RXF0S.Get() & 0x0F) == 0
}

// TxRaw transmits a CAN frame using the raw buffer element structure
func (can *FDCAN) TxRaw(e *FDCANTxBufferElement) error {
	// Check if TX FIFO is full
	if can.TxFifoIsFull() {
		return errFDCANTxFifoFull
	}

	// Get put index
	putIndex := (can.Bus.TXFQS.Get() >> 16) & 0x03 // TFQPI[1:0]

	// Calculate TX buffer address
	sramBase := can.getSRAMBase()
	txAddress := sramBase + sramcanTFQSA + (uintptr(putIndex) * sramcanTFQSize)

	// Build first word
	var w1 uint32
	id := e.ID
	if !e.XTD {
		// Standard ID - shift to bits [28:18]
		id = (id & 0x7FF) << 18
	}
	w1 = id & 0x1FFFFFFF
	if e.ESI {
		w1 |= fdcanElementMaskESI
	}
	if e.XTD {
		w1 |= fdcanElementMaskXTD
	}
	if e.RTR {
		w1 |= fdcanElementMaskRTR
	}

	// Build second word
	var w2 uint32
	w2 = uint32(e.DLC) << 16
	if e.FDF {
		w2 |= fdcanElementMaskFDF
	}
	if e.BRS {
		w2 |= fdcanElementMaskBRS
	}
	if e.EFC {
		w2 |= fdcanElementMaskEFC
	}
	w2 |= uint32(e.MM) << 24

	// Write to message RAM
	*(*uint32)(unsafe.Pointer(txAddress)) = w1
	*(*uint32)(unsafe.Pointer(txAddress + 4)) = w2

	// Copy data bytes - must use 32-bit word access on Cortex-M0+
	dataLen := dlcToBytes[e.DLC&0x0F]
	numWords := (dataLen + 3) / 4
	for w := byte(0); w < numWords; w++ {
		var word uint32
		baseIdx := w * 4
		for b := byte(0); b < 4 && baseIdx+b < dataLen; b++ {
			word |= uint32(e.DB[baseIdx+b]) << (b * 8)
		}
		*(*uint32)(unsafe.Pointer(txAddress + 8 + uintptr(w)*4)) = word
	}

	// Request transmission
	can.Bus.TXBAR.Set(1 << putIndex)

	return nil
}

// Tx transmits a CAN frame with the specified ID and data
func (can *FDCAN) Tx(id uint32, data []byte, isFD, isExtendedID bool) error {
	length := byte(len(data))
	if length > 64 {
		length = 64
	}
	if !isFD && length > 8 {
		length = 8
	}

	e := FDCANTxBufferElement{
		ESI: false,
		XTD: isExtendedID,
		RTR: false,
		ID:  id,
		MM:  0,
		EFC: false,
		FDF: isFD,
		BRS: isFD,
		DLC: FDCANLengthToDlc(length, isFD),
	}

	for i := byte(0); i < length; i++ {
		e.DB[i] = data[i]
	}

	return can.TxRaw(&e)
}

// RxRaw receives a CAN frame into the raw buffer element structure
func (can *FDCAN) RxRaw(e *FDCANRxBufferElement) error {
	if can.RxFifoIsEmpty() {
		return errFDCANRxFifoEmpty
	}

	// Get get index
	getIndex := (can.Bus.RXF0S.Get() >> 8) & 0x03 // F0GI[1:0]

	// Calculate RX buffer address
	sramBase := can.getSRAMBase()
	rxAddress := sramBase + sramcanRF0SA + (uintptr(getIndex) * sramcanRF0Size)

	// Read first word
	w1 := *(*uint32)(unsafe.Pointer(rxAddress))
	e.ESI = (w1 & fdcanElementMaskESI) != 0
	e.XTD = (w1 & fdcanElementMaskXTD) != 0
	e.RTR = (w1 & fdcanElementMaskRTR) != 0

	if e.XTD {
		e.ID = w1 & fdcanElementMaskEXTID
	} else {
		e.ID = (w1 & fdcanElementMaskSTDID) >> 18
	}

	// Read second word
	w2 := *(*uint32)(unsafe.Pointer(rxAddress + 4))
	e.RXTS = uint16(w2 & fdcanElementMaskTS)
	e.DLC = uint8((w2 & fdcanElementMaskDLC) >> 16)
	e.BRS = (w2 & fdcanElementMaskBRS) != 0
	e.FDF = (w2 & fdcanElementMaskFDF) != 0
	e.FIDX = uint8((w2 & fdcanElementMaskFIDX) >> 24)
	e.ANMF = (w2 & fdcanElementMaskANMF) != 0

	// Copy data bytes - must use 32-bit word access on Cortex-M0+
	dataLen := dlcToBytes[e.DLC&0x0F]
	numWords := (dataLen + 3) / 4
	for w := byte(0); w < numWords; w++ {
		word := *(*uint32)(unsafe.Pointer(rxAddress + 8 + uintptr(w)*4))
		baseIdx := w * 4
		for b := byte(0); b < 4 && baseIdx+b < dataLen; b++ {
			e.DB[baseIdx+b] = byte(word >> (b * 8))
		}
	}

	// Acknowledge the read
	can.Bus.RXF0A.Set(uint32(getIndex))

	return nil
}

// Rx receives a CAN frame and returns its components
func (can *FDCAN) Rx() (id uint32, dlc byte, data []byte, isFD, isExtendedID bool, err error) {
	e := FDCANRxBufferElement{}
	err = can.RxRaw(&e)
	if err != nil {
		return 0, 0, nil, false, false, err
	}

	length := FDCANDlcToLength(e.DLC, e.FDF)
	return e.ID, length, e.DB[:length], e.FDF, e.XTD, nil
}

// SetInterrupt configures interrupt handling for the FDCAN peripheral
func (can *FDCAN) SetInterrupt(ie uint32, callback func(*FDCAN)) error {
	if callback == nil {
		can.Bus.IE.ClearBits(ie)
		return nil
	}

	can.Bus.IE.SetBits(ie)

	idx := can.instance
	fdcanInstances[idx] = can

	for i := uint(0); i < 32; i++ {
		if ie&(1<<i) != 0 {
			fdcanCallbacks[idx][i] = callback
		}
	}

	can.Interrupt.Enable()
	return nil
}

// ConfigureFilter configures a message filter
func (can *FDCAN) ConfigureFilter(config FDCANFilterConfig) error {
	sramBase := can.getSRAMBase()

	if config.IsExtendedID {
		// Extended filter
		if config.Index >= sramcanFLENbr {
			return errors.New("FDCAN: filter index out of range")
		}

		filterAddr := sramBase + sramcanFLESA + (uintptr(config.Index) * sramcanFLESize)

		// Build filter elements
		w1 := (uint32(config.Config) << 29) | (config.ID1 & 0x1FFFFFFF)
		w2 := (uint32(config.Type) << 30) | (config.ID2 & 0x1FFFFFFF)

		*(*uint32)(unsafe.Pointer(filterAddr)) = w1
		*(*uint32)(unsafe.Pointer(filterAddr + 4)) = w2
	} else {
		// Standard filter
		if config.Index >= sramcanFLSNbr {
			return errors.New("FDCAN: filter index out of range")
		}

		filterAddr := sramBase + sramcanFLSSA + (uintptr(config.Index) * sramcanFLSSize)

		// Build filter element
		w := (uint32(config.Type) << 30) |
			(uint32(config.Config) << 27) |
			((config.ID1 & 0x7FF) << 16) |
			(config.ID2 & 0x7FF)

		*(*uint32)(unsafe.Pointer(filterAddr)) = w
	}

	return nil
}

func (can *FDCAN) getSRAMBase() uintptr {
	base := uintptr(sramcanBase)
	if can.Bus == stm32.FDCAN2 {
		base += sramcanSize
	}
	return base
}

func (can *FDCAN) configureMessageRAM() {
	sramBase := can.getSRAMBase()

	// Clear message RAM
	for addr := sramBase; addr < sramBase+sramcanSize; addr += 4 {
		*(*uint32)(unsafe.Pointer(addr)) = 0
	}

	// Configure filter counts (using RXGFC register)
	// LSS = number of standard filters, LSE = number of extended filters
	rxgfc := can.Bus.RXGFC.Get()
	rxgfc &= ^uint32(0xFF000000)            // Clear LSS and LSE
	rxgfc |= (sramcanFLSNbr << 24)          // Standard filters
	rxgfc |= (sramcanFLENbr << 24) & 0xFF00 // Extended filters (shifted)
	can.Bus.RXGFC.Set(rxgfc)
}

func (can *FDCAN) calculateNominalBitTiming(rate FDCANTransferRate) (brp, tseg1, tseg2, sjw uint32, err error) {
	// STM32G0 FDCAN clock = 64MHz
	// Target: 80% sample point
	// Bit time = (1 + TSEG1 + TSEG2) time quanta
	switch rate {
	case FDCANTransferRate125kbps:
		// 64MHz / 32 = 2MHz, 16 tq per bit = 125kbps
		return 32, 13, 2, 4, nil
	case FDCANTransferRate250kbps:
		// 64MHz / 16 = 4MHz, 16 tq per bit = 250kbps
		return 16, 13, 2, 4, nil
	case FDCANTransferRate500kbps:
		// 64MHz / 8 = 8MHz, 16 tq per bit = 500kbps
		return 8, 13, 2, 4, nil
	case FDCANTransferRate1000kbps:
		// 64MHz / 4 = 16MHz, 16 tq per bit = 1Mbps
		return 4, 13, 2, 4, nil
	default:
		return 0, 0, 0, 0, errFDCANInvalidTransferRate
	}
}

func (can *FDCAN) calculateDataBitTiming(rate FDCANTransferRate) (brp, tseg1, tseg2, sjw uint32, err error) {
	// STM32G0 FDCAN clock = 64MHz
	// For data phase, we need higher bit rates
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
		// 64MHz / 2 = 32MHz, 16 tq per bit = 2Mbps
		return 2, 13, 2, 4, nil
	case FDCANTransferRate4000kbps:
		// 64MHz / 1 = 64MHz, 16 tq per bit = 4Mbps
		return 1, 13, 2, 4, nil
	default:
		return 0, 0, 0, 0, errFDCANInvalidTransferRateFD
	}
}

// FDCANDlcToLength converts a DLC value to actual byte length
func FDCANDlcToLength(dlc byte, isFD bool) byte {
	if dlc > 15 {
		dlc = 15
	}
	length := dlcToBytes[dlc]
	if !isFD && length > 8 {
		return 8
	}
	return length
}

// FDCANLengthToDlc converts a byte length to DLC value
func FDCANLengthToDlc(length byte, isFD bool) byte {
	if !isFD {
		if length > 8 {
			return 8
		}
		return length
	}

	switch {
	case length <= 8:
		return length
	case length <= 12:
		return 9
	case length <= 16:
		return 10
	case length <= 20:
		return 11
	case length <= 24:
		return 12
	case length <= 32:
		return 13
	case length <= 48:
		return 14
	default:
		return 15
	}
}

// Interrupt handling
var (
	fdcanInstances [2]*FDCAN
	fdcanCallbacks [2][32]func(*FDCAN)
)

func fdcanHandleInterrupt(idx int) {
	if fdcanInstances[idx] == nil {
		return
	}

	can := fdcanInstances[idx]
	ir := can.Bus.IR.Get()
	can.Bus.IR.Set(ir) // Clear interrupt flags

	for i := uint(0); i < 32; i++ {
		if ir&(1<<i) != 0 && fdcanCallbacks[idx][i] != nil {
			fdcanCallbacks[idx][i](can)
		}
	}
}

// Data returns the received data as a slice
func (e *FDCANRxBufferElement) Data() []byte {
	return e.DB[:FDCANDlcToLength(e.DLC, e.FDF)]
}

// Length returns the actual data length
func (e *FDCANRxBufferElement) Length() byte {
	return FDCANDlcToLength(e.DLC, e.FDF)
}

// enableFDCANClock enables the FDCAN peripheral clock
func enableFDCANClock() {
	// FDCAN clock is on APB1
	stm32.RCC.SetAPBENR1_FDCANEN(1)
}
