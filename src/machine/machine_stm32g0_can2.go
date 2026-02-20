//go:build stm32g0b1

package machine

import (
	"device/stm32"
	"unsafe"
)

var canRxCB [2]func(data []byte, id uint32, extendedID bool, timestamp uint32, flags uint32)

// Configure initializes the FDCAN peripheral and starts it.
func (can *CAN) Configure(config FDCANConfig) error {
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
			return errFDCANTimeout
		}
	}

	// Request initialization.
	can.Bus.SetCCCR_INIT(1)
	timeout = 10000
	for can.Bus.GetCCCR_INIT() == 0 {
		timeout--
		if timeout == 0 {
			return errFDCANTimeout
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
		return errFDCANInvalidTransferRateFD
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
			return errFDCANTimeout
		}
	}

	return nil
}

func (can *CAN) txFIFOLevel() (int, int) {
	free := int(can.Bus.TXFQS.Get() & 0x07) // TFFL[2:0]
	return sramcanTFQNbr - free, sramcanTFQNbr
}

func (can *CAN) tx(id uint32, extendedID bool, data []byte) error {
	if can.Bus.TXFQS.Get()&0x00200000 != 0 { // TFQF bit
		return errFDCANTxFifoFull
	}

	putIndex := (can.Bus.TXFQS.Get() >> 16) & 0x03 // TFQPI[1:0]
	txAddr := can.sramBase() + sramcanTFQSA + uintptr(putIndex)*sramcanTFQSize

	// Header word 1: identifier and flags.
	var w1 uint32
	if extendedID {
		w1 = (id & 0x1FFFFFFF) | fdcanElementMaskXTD
	} else {
		w1 = (id & 0x7FF) << 18
	}

	// Header word 2: DLC only (classic CAN, no FD/BRS).
	length := byte(len(data))
	if length > 8 {
		length = 8
	}
	w2 := uint32(length) << 16

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

func (can *CAN) rxFIFOLevel() (int, int) {
	level := int(can.Bus.RXF0S.Get() & 0x0F) // F0FL[3:0]
	return level, sramcanRF0Nbr
}

func (can *CAN) setRxCallback(cb func(data []byte, id uint32, extendedID bool, timestamp uint32, flags uint32)) {
	canRxCB[can.instance] = cb
}

func (can *CAN) rxPoll() error {
	cb := canRxCB[can.instance]
	if cb == nil {
		return nil
	}

	for can.Bus.RXF0S.Get()&0x0F != 0 {
		getIndex := (can.Bus.RXF0S.Get() >> 8) & 0x03 // F0GI[1:0]
		rxAddr := can.sramBase() + sramcanRF0SA + uintptr(getIndex)*sramcanRF0Size

		w1 := *(*uint32)(unsafe.Pointer(rxAddr))
		w2 := *(*uint32)(unsafe.Pointer(rxAddr + 4))

		extendedID := w1&fdcanElementMaskXTD != 0
		var id uint32
		if extendedID {
			id = w1 & fdcanElementMaskEXTID
		} else {
			id = (w1 & fdcanElementMaskSTDID) >> 18
		}

		timestamp := w2 & fdcanElementMaskTS
		dlc := byte((w2 & fdcanElementMaskDLC) >> 16)

		var flags uint32
		if w2&fdcanElementMaskFDF != 0 {
			flags |= 1 // bit 0: FD frame
		}
		if w1&fdcanElementMaskRTR != 0 {
			flags |= 2 // bit 1: RTR
		}
		if w2&fdcanElementMaskBRS != 0 {
			flags |= 4 // bit 2: BRS
		}
		if w1&fdcanElementMaskESI != 0 {
			flags |= 8 // bit 3: ESI
		}

		dataLen := dlcToBytes[dlc&0x0F]
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
		cb(buf[:dataLen], id, extendedID, timestamp, flags)
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
func fdcanNominalBitTiming(rate FDCANTransferRate) (brp, tseg1, tseg2, sjw uint32, err error) {
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
		return 0, 0, 0, 0, errFDCANInvalidTransferRate
	}
}

// fdcanDataBitTiming returns prescaler and segment values for the data phase (FD).
func fdcanDataBitTiming(rate FDCANTransferRate) (brp, tseg1, tseg2, sjw uint32, err error) {
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
		return 0, 0, 0, 0, errFDCANInvalidTransferRateFD
	}
}
