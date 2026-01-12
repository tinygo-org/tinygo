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
	FDCANTransferRate100kbps  FDCANTransferRate = 100000
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
	Standby        Pin  // Optional standby pin for CAN transceiver (set to NoPin if not used)
	EnableFD       bool // Enable FD mode for larger payloads (up to 64 bytes) and higher data rates
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

// FDCANBusState represents the current CAN bus state
type FDCANBusState uint8

const (
	FDCANBusStateErrorActive  FDCANBusState = 0 // Normal operation, TEC/REC < 128
	FDCANBusStateErrorPassive FDCANBusState = 1 // TEC or REC >= 128
	FDCANBusStateBusOff       FDCANBusState = 2 // TEC >= 256, node disconnected from bus
)

// FDCANLastError represents the last error that occurred on the bus
type FDCANLastError uint8

const (
	FDCANErrorNone     FDCANLastError = 0 // No error
	FDCANErrorStuff    FDCANLastError = 1 // Stuff error - more than 5 equal bits
	FDCANErrorForm     FDCANLastError = 2 // Form error - fixed format part violated
	FDCANErrorAck      FDCANLastError = 3 // Ack error - no acknowledgement received
	FDCANErrorBit1     FDCANLastError = 4 // Bit1 error - sent recessive, monitored dominant
	FDCANErrorBit0     FDCANLastError = 5 // Bit0 error - sent dominant, monitored recessive
	FDCANErrorCRC      FDCANLastError = 6 // CRC error - CRC mismatch
	FDCANErrorNoChange FDCANLastError = 7 // No change since last read
)

var (
	errFDCANInvalidTransferRate   = errors.New("FDCAN: invalid TransferRate")
	errFDCANInvalidTransferRateFD = errors.New("FDCAN: invalid TransferRateFD")
	errFDCANTimeout               = errors.New("FDCAN: timeout")
	errFDCANTxFifoFull            = errors.New("FDCAN: Tx FIFO full")
	errFDCANRxFifoEmpty           = errors.New("FDCAN: Rx FIFO empty")
	errFDCANNotStarted            = errors.New("FDCAN: not started")
	errFDCANTxCancelled           = errors.New("FDCAN: Tx cancelled")
	errFDCANBusOff                = errors.New("FDCAN: bus off")
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

	// Configure TX and RX pins in Alternate Function Push-Pull mode (matches HAL GPIO_MODE_AF_PP)
	// Use PinModePWMOutput which sets MODER to alternate function mode and calls SetAltFunc
	config.Tx.ConfigureAltFunc(PinConfig{Mode: PinModePWMOutput}, can.TxAltFuncSelect)
	config.Rx.ConfigureAltFunc(PinConfig{Mode: PinModePWMOutput}, can.RxAltFuncSelect)

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

	// Disable automatic retransmission (matches HAL behavior)
	// DAR=1 means retransmission is disabled
	can.Bus.SetCCCR_DAR(1)

	// Disable transmit pause
	can.Bus.SetCCCR_TXP(0)

	// Disable protocol exception handling (matches HAL PXHD=DISABLE)
	can.Bus.SetCCCR_PXHD(1)

	// Configure FD mode
	if config.EnableFD {
		// Enable FD mode for larger payloads (up to 64 bytes) and bit rate switching
		can.Bus.SetCCCR_FDOE(1) // Enable FD operation
		can.Bus.SetCCCR_BRSE(1) // Enable bit rate switching for data phase
	} else {
		// Classic CAN frame format (matches HAL FDCAN_FRAME_CLASSIC)
		can.Bus.SetCCCR_FDOE(0)
		can.Bus.SetCCCR_BRSE(0)
	}

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

// TxPendingCount returns the number of messages pending transmission
func (can *FDCAN) TxPendingCount() int {
	pending := can.Bus.TXBRP.Get() & 0x07 // TXBRP[2:0] for 3 TX buffers
	count := 0
	for pending != 0 {
		count += int(pending & 1)
		pending >>= 1
	}
	return count
}

// TxIsPending returns true if there are any pending transmissions
func (can *FDCAN) TxIsPending() bool {
	return (can.Bus.TXBRP.Get() & 0x07) != 0
}

// TxCancelAll cancels all pending transmissions.
// Returns the number of transmissions that were cancelled.
func (can *FDCAN) TxCancelAll() int {
	pending := can.Bus.TXBRP.Get() & 0x07
	if pending == 0 {
		return 0
	}

	// Request cancellation for all pending buffers
	can.Bus.TXBCR.Set(pending)

	// Wait for cancellation to complete (with timeout)
	timeout := 10000
	for can.Bus.TXBCF.Get()&pending != pending {
		timeout--
		if timeout == 0 {
			break
		}
	}

	// Count how many were cancelled
	cancelled := can.Bus.TXBCF.Get() & 0x07
	count := 0
	for cancelled != 0 {
		count += int(cancelled & 1)
		cancelled >>= 1
	}
	return count
}

// TxCancel cancels a specific TX buffer (0-2).
// Returns true if the buffer was pending and cancellation was requested.
func (can *FDCAN) TxCancel(bufferIndex uint8) bool {
	if bufferIndex > 2 {
		return false
	}

	bufferMask := uint32(1) << bufferIndex

	// Check if buffer has pending transmission
	if can.Bus.TXBRP.Get()&bufferMask == 0 {
		return false // Nothing to cancel
	}

	// Request cancellation
	can.Bus.TXBCR.Set(bufferMask)

	// Wait for cancellation to complete (with timeout)
	timeout := 10000
	for can.Bus.TXBCF.Get()&bufferMask == 0 {
		timeout--
		if timeout == 0 {
			return false
		}
	}

	return true
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

// TxBlocking transmits a CAN frame and waits for transmission to complete.
// Returns error if transmission fails or times out.
// timeoutMs specifies the maximum time to wait in milliseconds (0 = wait forever).
func (can *FDCAN) TxBlocking(id uint32, data []byte, isFD, isExtendedID bool, timeoutMs uint32) error {
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

	return can.TxRawBlocking(&e, timeoutMs)
}

// TxRawBlocking transmits a CAN frame and waits for transmission to complete.
// Returns error if transmission fails or times out.
// timeoutMs specifies the maximum time to wait in milliseconds (0 = wait forever).
func (can *FDCAN) TxRawBlocking(e *FDCANTxBufferElement, timeoutMs uint32) error {
	// Check if TX FIFO is full
	if can.TxFifoIsFull() {
		return errFDCANTxFifoFull
	}

	// Get put index before adding to FIFO
	putIndex := (can.Bus.TXFQS.Get() >> 16) & 0x03 // TFQPI[1:0]
	bufferMask := uint32(1) << putIndex

	// Calculate TX buffer address
	sramBase := can.getSRAMBase()
	txAddress := sramBase + sramcanTFQSA + (uintptr(putIndex) * sramcanTFQSize)

	// Build first word
	var w1 uint32
	id := e.ID
	if !e.XTD {
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

	// Copy data bytes
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
	can.Bus.TXBAR.Set(bufferMask)

	// Wait for transmission to complete
	// TXBTO (TX Buffer Transmission Occurred) bit is set when transmission completes
	timeout := timeoutMs * 1000 // Convert to microseconds for finer granularity
	if timeoutMs == 0 {
		timeout = 0xFFFFFFFF // Effectively infinite
	}

	for {
		// Check if transmission occurred
		if can.Bus.TXBTO.Get()&bufferMask != 0 {
			return nil // Success
		}

		// Check if transmission was cancelled
		if can.Bus.TXBCF.Get()&bufferMask != 0 {
			return errFDCANTxCancelled
		}

		// Check for bus-off (can't transmit)
		if can.IsBusOff() {
			return errFDCANBusOff
		}

		timeout--
		if timeout == 0 {
			return errFDCANTimeout
		}

		// Small delay to avoid busy-waiting too hard
		// On Cortex-M0+ at 64MHz, a simple loop iteration is ~10-20 cycles
		// This gives roughly microsecond-level timing
	}
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

// Rx8 receives a classic CAN frame (up to 8 bytes) with no heap allocation.
// Returns the data in a fixed-size array and the actual data length.
// For FD frames with more than 8 bytes, only the first 8 bytes are returned.
func (can *FDCAN) Rx8() (id uint32, data [8]byte, length uint8, isFD, isExtendedID bool, err error) {
	var e FDCANRxBufferElement
	err = can.RxRaw(&e)
	if err != nil {
		return 0, [8]byte{}, 0, false, false, err
	}

	length = FDCANDlcToLength(e.DLC, e.FDF)
	if length > 8 {
		length = 8
	}

	var data8 [8]byte
	copy(data8[:], e.DB[:length])
	return e.ID, data8, length, e.FDF, e.XTD, nil
}

// Rx64 receives a CAN FD frame (up to 64 bytes) with no heap allocation.
// Returns the full 64-byte data buffer and the actual data length.
// Works for both classic CAN (up to 8 bytes) and CAN FD (up to 64 bytes).
func (can *FDCAN) Rx64() (id uint32, data [64]byte, length uint8, isFD, isExtendedID bool, err error) {
	var e FDCANRxBufferElement
	err = can.RxRaw(&e)
	if err != nil {
		return 0, [64]byte{}, 0, false, false, err
	}

	length = FDCANDlcToLength(e.DLC, e.FDF)
	return e.ID, e.DB, length, e.FDF, e.XTD, nil
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

// Filter type constants
const (
	FDCANFilterTypeRange   = 0 // Accept messages with ID in range [ID1, ID2]
	FDCANFilterTypeDual    = 1 // Accept messages matching ID1 or ID2
	FDCANFilterTypeMask    = 2 // Accept messages matching (ID & ID2) == ID1
	FDCANFilterTypeDisable = 3 // Filter disabled
)

// Filter config constants (destination)
const (
	FDCANFilterConfigDisable = 0 // Filter disabled
	FDCANFilterConfigFIFO0   = 1 // Store in RX FIFO 0
	FDCANFilterConfigFIFO1   = 2 // Store in RX FIFO 1
	FDCANFilterConfigReject  = 3 // Reject matching messages
)

// AcceptAll configures the FDCAN to accept all messages (no filtering).
// All standard and extended ID messages will be accepted into RX FIFO 0.
func (can *FDCAN) AcceptAll() error {
	// Configure RXGFC to accept all non-matching frames to FIFO0
	// ANFS = 0 (accept to FIFO0), ANFE = 0 (accept to FIFO0)
	// LSS = 0 (no standard filters), LSE = 0 (no extended filters)
	can.Bus.RXGFC.Set(0)
	return nil
}

// RejectNonMatching configures the FDCAN to reject all frames that don't match
// any configured filter. Call this after setting up your filters.
func (can *FDCAN) RejectNonMatching() {
	// RXGFC register bits:
	// Bits 27:24 - LSE[3:0] - List Size Extended
	// Bits 20:16 - LSS[4:0] - List Size Standard
	// Bits 5:4 - ANFS[1:0] - Accept Non-matching Frames Standard (2 = reject)
	// Bits 3:2 - ANFE[1:0] - Accept Non-matching Frames Extended (2 = reject)
	rxgfc := can.Bus.RXGFC.Get()
	rxgfc &^= (0x3 << 4) | (0x3 << 2) // Clear ANFS and ANFE
	rxgfc |= (2 << 4) | (2 << 2)      // ANFS=2, ANFE=2 (reject non-matching)
	can.Bus.RXGFC.Set(rxgfc)
}

// AcceptID configures a filter to accept only messages with the specified ID.
// Use filterIndex 0-27 for standard IDs, 0-7 for extended IDs.
func (can *FDCAN) AcceptID(filterIndex uint8, id uint32, isExtended bool) error {
	return can.ConfigureFilter(FDCANFilterConfig{
		Index:        filterIndex,
		Type:         FDCANFilterTypeDual, // Dual ID mode - match either ID1 or ID2
		Config:       FDCANFilterConfigFIFO0,
		ID1:          id,
		ID2:          id, // Same ID for both = single ID match
		IsExtendedID: isExtended,
	})
}

// AcceptRange configures a filter to accept messages with IDs in the range [idLow, idHigh].
func (can *FDCAN) AcceptRange(filterIndex uint8, idLow, idHigh uint32, isExtended bool) error {
	return can.ConfigureFilter(FDCANFilterConfig{
		Index:        filterIndex,
		Type:         FDCANFilterTypeRange,
		Config:       FDCANFilterConfigFIFO0,
		ID1:          idLow,
		ID2:          idHigh,
		IsExtendedID: isExtended,
	})
}

// AcceptMask configures a classic ID/mask filter.
// A message is accepted if: (receivedID & mask) == (id & mask)
// Example: AcceptMask(0, 0x100, 0x700, false) accepts IDs 0x100-0x1FF
func (can *FDCAN) AcceptMask(filterIndex uint8, id, mask uint32, isExtended bool) error {
	return can.ConfigureFilter(FDCANFilterConfig{
		Index:        filterIndex,
		Type:         FDCANFilterTypeMask,
		Config:       FDCANFilterConfigFIFO0,
		ID1:          id,
		ID2:          mask,
		IsExtendedID: isExtended,
	})
}

// RejectID configures a filter to reject messages with the specified ID.
func (can *FDCAN) RejectID(filterIndex uint8, id uint32, isExtended bool) error {
	return can.ConfigureFilter(FDCANFilterConfig{
		Index:        filterIndex,
		Type:         FDCANFilterTypeDual,
		Config:       FDCANFilterConfigReject,
		ID1:          id,
		ID2:          id,
		IsExtendedID: isExtended,
	})
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

	// Configure RXGFC register
	// LSS[4:0] at bits 20:16 = number of standard filters (28 max)
	// LSE[3:0] at bits 27:24 = number of extended filters (8 max)
	// ANFS[1:0] at bits 5:4 = Accept Non-matching Frames Standard (0 = accept to RxFIFO0)
	// ANFE[1:0] at bits 3:2 = Accept Non-matching Frames Extended (0 = accept to RxFIFO0)
	rxgfc := uint32(sramcanFLSNbr<<16) | uint32(sramcanFLENbr<<24) // LSS=28, LSE=8
	can.Bus.RXGFC.Set(rxgfc)

	// Configure TX buffer for FIFO mode (matches HAL: FDCAN_TX_FIFO_OPERATION)
	// TFQM bit 24 = 0 for FIFO mode
	can.Bus.TXBC.Set(0)
}

func (can *FDCAN) calculateNominalBitTiming(rate FDCANTransferRate) (brp, tseg1, tseg2, sjw uint32, err error) {
	// STM32G0 FDCAN clock = 64MHz
	// Target: 80% sample point
	// Bit time = (1 + TSEG1 + TSEG2) time quanta
	// SJW = 1 to match HAL configuration
	switch rate {
	case FDCANTransferRate100kbps:
		// 64MHz / 40 = 1.6MHz, 16 tq per bit = 100kbps
		return 40, 13, 2, 1, nil
	case FDCANTransferRate125kbps:
		// 64MHz / 32 = 2MHz, 16 tq per bit = 125kbps
		return 32, 13, 2, 1, nil
	case FDCANTransferRate250kbps:
		// 64MHz / 16 = 4MHz, 16 tq per bit = 250kbps
		return 16, 13, 2, 1, nil
	case FDCANTransferRate500kbps:
		// 64MHz / 8 = 8MHz, 16 tq per bit = 500kbps
		return 8, 13, 2, 1, nil
	case FDCANTransferRate1000kbps:
		// 64MHz / 4 = 16MHz, 16 tq per bit = 1Mbps
		return 4, 13, 2, 1, nil
	default:
		return 0, 0, 0, 0, errFDCANInvalidTransferRate
	}
}

func (can *FDCAN) calculateDataBitTiming(rate FDCANTransferRate) (brp, tseg1, tseg2, sjw uint32, err error) {
	// STM32G0 FDCAN clock = 64MHz
	// For data phase, we need higher bit rates
	switch rate {
	case FDCANTransferRate100kbps:
		return 40, 13, 2, 4, nil
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
	fdcanInstances   [2]*FDCAN
	fdcanCallbacks   [2][32]func(*FDCAN)
	fdcanRxCallbacks [2]func(*FDCANRxBufferElement)
	fdcanRxBuffers   [2]FDCANRxBufferElement // Pre-allocated buffers to avoid heap alloc in interrupt
)

func fdcanHandleInterrupt(idx int) {
	if fdcanInstances[idx] == nil {
		return
	}

	can := fdcanInstances[idx]
	ir := can.Bus.IR.Get()
	can.Bus.IR.Set(ir) // Clear interrupt flags

	// Handle RX FIFO 0 new message interrupt
	if ir&FDCAN_IT_RX_FIFO0_NEW_MESSAGE != 0 && fdcanRxCallbacks[idx] != nil {
		// Read all available messages using pre-allocated buffer
		for !can.RxFifoIsEmpty() {
			if can.RxRaw(&fdcanRxBuffers[idx]) == nil {
				fdcanRxCallbacks[idx](&fdcanRxBuffers[idx])
			}
		}
	}

	// Handle other registered callbacks
	for i := uint(0); i < 32; i++ {
		if ir&(1<<i) != 0 && fdcanCallbacks[idx][i] != nil {
			fdcanCallbacks[idx][i](can)
		}
	}
}

// SetRxCallback registers a callback function that will be called when a CAN message is received.
// The callback receives the full RxBufferElement with ID, data, flags, etc.
// Pass nil to disable the callback.
//
// Example:
//
//	CAN1.SetRxCallback(func(msg *machine.FDCANRxBufferElement) {
//	    println("Received ID:", msg.ID, "Data:", msg.DB[0], msg.DB[1])
//	})
func (can *FDCAN) SetRxCallback(callback func(*FDCANRxBufferElement)) error {
	idx := can.instance
	fdcanInstances[idx] = can
	fdcanRxCallbacks[idx] = callback

	if callback == nil {
		// Disable RX FIFO 0 new message interrupt
		can.Bus.IE.ClearBits(FDCAN_IT_RX_FIFO0_NEW_MESSAGE)
		return nil
	}

	// Enable interrupt line 0 (ILE bit 0)
	can.Bus.ILE.Set(1)

	// Route RX FIFO 0 interrupts to line 0 (default, bit 0 of ILS = 0)
	// ILS register: 0 = interrupt line 0, 1 = interrupt line 1
	// RF0NE (bit 0) should go to line 0, so we clear bit 0
	ils := can.Bus.ILS.Get()
	ils &^= FDCAN_IT_RX_FIFO0_NEW_MESSAGE
	can.Bus.ILS.Set(ils)

	// Enable RX FIFO 0 new message interrupt
	can.Bus.IE.SetBits(FDCAN_IT_RX_FIFO0_NEW_MESSAGE)

	// Enable the NVIC interrupt
	can.Interrupt.Enable()

	return nil
}

// Data returns the received data as a slice
func (e *FDCANRxBufferElement) Data() []byte {
	return e.DB[:FDCANDlcToLength(e.DLC, e.FDF)]
}

// Length returns the actual data length
func (e *FDCANRxBufferElement) Length() byte {
	return FDCANDlcToLength(e.DLC, e.FDF)
}

// GetErrorCounters returns the transmit and receive error counters.
// TEC >= 256 means bus-off state. TEC or REC >= 128 means error passive state.
func (can *FDCAN) GetErrorCounters() (txErrors, rxErrors uint8) {
	ecr := can.Bus.ECR.Get()
	txErrors = uint8(ecr & 0xFF)        // TEC[7:0]
	rxErrors = uint8((ecr >> 8) & 0x7F) // REC[6:0]
	return
}

// GetBusState returns the current CAN bus state based on error counters.
func (can *FDCAN) GetBusState() FDCANBusState {
	psr := can.Bus.PSR.Get()

	// Check Bus_Off status (bit 7)
	if psr&(1<<7) != 0 {
		return FDCANBusStateBusOff
	}

	// Check Error Passive status (bit 5)
	if psr&(1<<5) != 0 {
		return FDCANBusStateErrorPassive
	}

	return FDCANBusStateErrorActive
}

// GetLastError returns the last error that occurred on the CAN bus.
// The error is cleared after reading.
func (can *FDCAN) GetLastError() FDCANLastError {
	psr := can.Bus.PSR.Get()
	lec := (psr >> 0) & 0x07 // LEC[2:0] - Last Error Code
	return FDCANLastError(lec)
}

// GetDataPhaseLastError returns the last error that occurred during data phase (FD only).
func (can *FDCAN) GetDataPhaseLastError() FDCANLastError {
	psr := can.Bus.PSR.Get()
	dlec := (psr >> 8) & 0x07 // DLEC[2:0] - Data Phase Last Error Code
	return FDCANLastError(dlec)
}

// IsBusOff returns true if the CAN controller is in bus-off state.
func (can *FDCAN) IsBusOff() bool {
	return can.Bus.PSR.Get()&(1<<7) != 0
}

// IsErrorPassive returns true if the CAN controller is in error passive state.
func (can *FDCAN) IsErrorPassive() bool {
	return can.Bus.PSR.Get()&(1<<5) != 0
}

// IsErrorWarning returns true if at least one error counter has reached the warning level (>= 96).
func (can *FDCAN) IsErrorWarning() bool {
	return can.Bus.PSR.Get()&(1<<6) != 0
}

// GetBusActivity returns true if the CAN bus is currently active (transmitting or receiving).
func (can *FDCAN) GetBusActivity() (transmitting, receiving bool) {
	psr := can.Bus.PSR.Get()
	// Activity bits: RESI (bit 11), RBRS (bit 12), etc. indicate recent activity
	// For actual TX/RX activity, we check the protocol status
	act := (psr >> 3) & 0x03 // ACT[1:0] - Activity
	// 00 = Synchronizing, 01 = Idle, 10 = Receiver, 11 = Transmitter
	receiving = act == 2
	transmitting = act == 3
	return
}

// String returns a human-readable string for the bus state.
func (s FDCANBusState) String() string {
	switch s {
	case FDCANBusStateErrorActive:
		return "ErrorActive"
	case FDCANBusStateErrorPassive:
		return "ErrorPassive"
	case FDCANBusStateBusOff:
		return "BusOff"
	default:
		return "Unknown"
	}
}

// String returns a human-readable string for the error type.
func (e FDCANLastError) String() string {
	switch e {
	case FDCANErrorNone:
		return "None"
	case FDCANErrorStuff:
		return "StuffError"
	case FDCANErrorForm:
		return "FormError"
	case FDCANErrorAck:
		return "AckError"
	case FDCANErrorBit1:
		return "Bit1Error"
	case FDCANErrorBit0:
		return "Bit0Error"
	case FDCANErrorCRC:
		return "CRCError"
	case FDCANErrorNoChange:
		return "NoChange"
	default:
		return "Unknown"
	}
}

// enableFDCANClock enables the FDCAN peripheral clock
func enableFDCANClock() {
	// Select PCLK1 as FDCAN clock source (matches HAL: RCC_FDCANCLKSOURCE_PCLK1)
	// FDCANSEL[1:0] = 00 in RCC_CCIPR2
	ccipr2 := stm32.RCC.CCIPR2.Get()
	ccipr2 &= ^uint32(0x3 << 8) // Clear FDCANSEL bits
	stm32.RCC.CCIPR2.Set(ccipr2)

	// Enable FDCAN peripheral clock on APB1
	stm32.RCC.SetAPBENR1_FDCANEN(1)
}
