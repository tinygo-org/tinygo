//go:build stm32g0

package machine

// unexported functions here are implemented in the device file
// and added to the build tags of this file.

// These types are an alias for documentation purposes exclusively. We wish
// the interface to be used by other ecosystems besides TinyGo which is why
// we need these types to be a primitive types at the interface level.
// If these types are defined at machine or machine/can level they are not
// usable by non-TinyGo projects. This is not good news for fostering wider adoption
// of our API in "big-Go" embedded system projects like TamaGo and periph.io
type (
	// CAN IDs in tinygo are represented as 30 bit integers where
	// bits 1..29 store the actual ID and the 30th bit stores the IDE bit (if extended ID).
	// We include the extended ID bit in the ID itself to make comparison of IDs easier for users
	// since two identical IDs where one is extended and one is not are NOT equivalent IDs.
	canID = uint32
	// CAN flags bitmask are defined below.
	canFlags = uint32
)

// CAN ID definitions.
const (
	canIDStdMask      canID = (1 << 11) - 1
	canIDExtendedMask canID = (1 << 29) - 1
	canIDExtendedBit  canID = 1 << 30
)

// CAN Flag bit definitions.
const (
	canFlagBRS canFlags = 1 << 0 // Bit Rate Switch active on tx/rx of frame.
	canFlagFDF canFlags = 1 << 1 // Is a FD Frame.
	canFlagRTR canFlags = 1 << 2 // is a retransmission request frame.
	canFlagESI canFlags = 1 << 3 // Error status indicator active on tx/rx of frame.
	canFlagIDE canFlags = 1 << 4 // Extended ID.
)

// TxFIFOLevel returns amount of CAN frames stored for transmission and total Tx fifo length.
func (can *CAN) TxFIFOLevel() (level int, maxlevel int) {
	return can.txFIFOLevel()
}

// Tx puts a CAN frame in TxFIFO for transmission. Returns error if TxFIFO is full.
func (can *CAN) Tx(id canID, flags canFlags, data []byte) error {
	return can.tx(id, flags, data)
}

// RxFIFOLevel returns amount of CAN frames received and stored and total Rx fifo length.
// If the hardware is interrupt driven RxFIFOLevel should return 0,0.
func (can *CAN) RxFIFOLevel() (level int, maxlevel int) {
	return can.rxFIFOLevel()
}

type canRxCallback = func(data []byte, id canID, timestamp uint32, flags canFlags)

// SetRxCallback sets the receive callback. See [canFlags] for information on how bits are layed out.
func (can *CAN) SetRxCallback(cb canRxCallback) {
	can.setRxCallback(cb)
}

// RxPoll is called periodically for poll driven drivers. If the driver is interrupt driven
// then RxPoll is a no-op and may return nil. Users may determine if a CAN is interrupt driven by
// checking if RxFIFOLevel returns 0,0.
func (can *CAN) RxPoll() error {
	return can.rxPoll()
}

// DLC to bytes lookup table
var dlcToBytes = [16]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 12, 16, 20, 24, 32, 48, 64}

// dlcToLength converts a DLC value to actual byte length
func dlcToLength(dlc byte) uint8 {
	if dlc > 15 {
		dlc = 15
	}
	return dlcToBytes[dlc]
}

// lengthToDLC converts a byte length to DLC value
func lengthToDLC(length uint8) (dlc byte) {
	switch {
	case length <= 8:
		dlc = length
	case length <= 12:
		dlc = 9
	case length <= 16:
		dlc = 10
	case length <= 20:
		dlc = 11
	case length <= 24:
		dlc = 12
	case length <= 32:
		dlc = 13
	case length <= 48:
		dlc = 14
	default:
		dlc = 15
	}
	return dlc
}
