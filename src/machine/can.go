//go:build stm32g0

package machine

// unexported functions here are implemented in the device file
// and added to the build tags of this file.

// TxFIFOLevel returns amount of CAN frames stored for transmission and total Tx fifo length.
func (can *CAN) TxFIFOLevel() (level int, maxlevel int) {
	return can.txFIFOLevel()
}

// Tx puts a CAN frame in TxFIFO for transmission. Returns error if TxFIFO is full.
func (can *CAN) Tx(id uint32, extendedID bool, data []byte) error {
	return can.tx(id, extendedID, data)
}

// RxFIFOLevel returns amount of CAN frames received and stored and total Rx fifo length.
// If the hardware is interrupt driven RxFIFOLevel should return 0,0.
func (can *CAN) RxFIFOLevel() (level int, maxlevel int) {
	return can.rxFIFOLevel()
}

// SetRxCallback sets the receive callback. flags is a bitfield where bits set are:
//   - bit 0: Is a FD frame.
//   - bit 1: Is a RTR frame.
//   - bit 2: Bitrate switch was active in frame.
//   - bit 3: ESI error state indicator active.
func (can *CAN) SetRxCallback(cb func(data []byte, id uint32, extendedID bool, timestamp uint32, flags uint32)) {
	can.setRxCallback(cb)
}

// RxPoll is called periodically for poll driven drivers. If the driver is interrupt driven
// then RxPoll is a no-op and may return nil. Users may determine if a CAN is interrupt driven by
// checking if RxFIFOLevel returns 0,0.
func (can *CAN) RxPoll() error {
	return can.rxPoll()
}
