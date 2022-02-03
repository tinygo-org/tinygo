//go:build (sam && atsame51) || (sam && atsame54)
// +build sam,atsame51 sam,atsame54

package machine

import (
	"device/sam"
	"errors"
	"runtime/interrupt"
	"unsafe"
)

const (
	CANRxFifoSize = 16
	CANTxFifoSize = 16
	CANEvFifoSize = 16
)

// Message RAM can only be located in the first 64 KB area of the system RAM.
// TODO: when the go:section pragma is merged, add the section configuration

//go:align 4
var CANRxFifo [2][(8 + 64) * CANRxFifoSize]byte

//go:align 4
var CANTxFifo [2][(8 + 64) * CANTxFifoSize]byte

//go:align 4
var CANEvFifo [2][(8) * CANEvFifoSize]byte

type CAN struct {
	Bus *sam.CAN_Type
}

type CANTransferRate uint32

// CAN transfer rates for CANConfig
const (
	CANTransferRate125kbps  CANTransferRate = 125000
	CANTransferRate250kbps  CANTransferRate = 250000
	CANTransferRate500kbps  CANTransferRate = 500000
	CANTransferRate1000kbps CANTransferRate = 1000000
	CANTransferRate2000kbps CANTransferRate = 2000000
	CANTransferRate4000kbps CANTransferRate = 4000000
)

// CANConfig holds CAN configuration parameters. Tx and Rx need to be
// specified with some pins. When the Standby Pin is specified, configure it
// as an output pin and output Low in Configure(). If this operation is not
// necessary, specify NoPin.
type CANConfig struct {
	TransferRate   CANTransferRate
	TransferRateFD CANTransferRate
	Tx             Pin
	Rx             Pin
	Standby        Pin
}

var (
	errCANInvalidTransferRate   = errors.New("CAN: invalid TransferRate")
	errCANInvalidTransferRateFD = errors.New("CAN: invalid TransferRateFD")
)

// Configure this CAN peripheral with the given configuration.
func (can *CAN) Configure(config CANConfig) error {
	if config.Standby != NoPin {
		config.Standby.Configure(PinConfig{Mode: PinOutput})
		config.Standby.Low()
	}

	mode := PinCAN0
	if can.instance() == 1 {
		mode = PinCAN1
	}

	config.Rx.Configure(PinConfig{Mode: mode})
	config.Tx.Configure(PinConfig{Mode: mode})

	can.Bus.CCCR.SetBits(sam.CAN_CCCR_INIT)
	for !can.Bus.CCCR.HasBits(sam.CAN_CCCR_INIT) {
	}

	can.Bus.CCCR.SetBits(sam.CAN_CCCR_CCE)

	can.Bus.CCCR.SetBits(sam.CAN_CCCR_BRSE | sam.CAN_CCCR_FDOE)
	can.Bus.MRCFG.Set(sam.CAN_MRCFG_QOS_MEDIUM)
	// base clock == 48 MHz
	if config.TransferRate == 0 {
		config.TransferRate = CANTransferRate500kbps
	}
	brp := uint32(6)
	switch config.TransferRate {
	case CANTransferRate125kbps:
		brp = 32
	case CANTransferRate250kbps:
		brp = 16
	case CANTransferRate500kbps:
		brp = 8
	case CANTransferRate1000kbps:
		brp = 4
	default:
		return errCANInvalidTransferRate
	}
	can.Bus.NBTP.Set(8<<sam.CAN_NBTP_NTSEG1_Pos | (brp-1)<<sam.CAN_NBTP_NBRP_Pos |
		1<<sam.CAN_NBTP_NTSEG2_Pos | 3<<sam.CAN_NBTP_NSJW_Pos)

	if config.TransferRateFD == 0 {
		config.TransferRateFD = CANTransferRate1000kbps
	}
	if config.TransferRateFD < config.TransferRate {
		return errCANInvalidTransferRateFD
	}
	brp = uint32(2)
	switch config.TransferRateFD {
	case CANTransferRate125kbps:
		brp = 32
	case CANTransferRate250kbps:
		brp = 16
	case CANTransferRate500kbps:
		brp = 8
	case CANTransferRate1000kbps:
		brp = 4
	case CANTransferRate2000kbps:
		brp = 2
	case CANTransferRate4000kbps:
		brp = 1
	default:
		return errCANInvalidTransferRateFD
	}
	can.Bus.DBTP.Set((brp-1)<<sam.CAN_DBTP_DBRP_Pos | 8<<sam.CAN_DBTP_DTSEG1_Pos |
		1<<sam.CAN_DBTP_DTSEG2_Pos | 3<<sam.CAN_DBTP_DSJW_Pos)

	can.Bus.RXF0C.Set(sam.CAN_RXF0C_F0OM | CANRxFifoSize<<sam.CAN_RXF0C_F0S_Pos | uint32(uintptr(unsafe.Pointer(&CANRxFifo[can.instance()][0])))&0xFFFF)
	can.Bus.RXESC.Set(sam.CAN_RXESC_F0DS_DATA64)
	can.Bus.TXESC.Set(sam.CAN_TXESC_TBDS_DATA64)
	can.Bus.TXBC.Set(CANTxFifoSize<<sam.CAN_TXBC_TFQS_Pos | 0<<sam.CAN_TXBC_NDTB_Pos | uint32(uintptr(unsafe.Pointer(&CANTxFifo[can.instance()][0])))&0xFFFF)
	can.Bus.TXEFC.Set(CANEvFifoSize<<sam.CAN_TXEFC_EFS_Pos | uint32(uintptr(unsafe.Pointer(&CANEvFifo[can.instance()][0])))&0xFFFF)

	can.Bus.TSCC.Set(sam.CAN_TSCC_TSS_INC)

	can.Bus.GFC.Set(0<<sam.CAN_GFC_ANFS_Pos | 0<<sam.CAN_GFC_ANFE_Pos)

	can.Bus.SIDFC.Set(0 << sam.CAN_SIDFC_LSS_Pos)
	can.Bus.XIDFC.Set(0 << sam.CAN_SIDFC_LSS_Pos)

	can.Bus.XIDAM.Set(0x1FFFFFFF << sam.CAN_XIDAM_EIDM_Pos)

	can.Bus.ILE.SetBits(sam.CAN_ILE_EINT0)

	can.Bus.CCCR.ClearBits(sam.CAN_CCCR_CCE)
	can.Bus.CCCR.ClearBits(sam.CAN_CCCR_INIT)
	for can.Bus.CCCR.HasBits(sam.CAN_CCCR_INIT) {
	}

	return nil
}

// Callbacks to be called for CAN.SetInterrupt(). Wre're using the magic
// constant 2 and 32 here beacuse th SAM E51/E54 has 2 CAN and 32 interrupt
// sources.
var (
	canInstances [2]*CAN
	canCallbacks [2][32]func(*CAN)
)

// SetInterrupt sets an interrupt to be executed when a particular CAN state.
//
// This call will replace a previously set callback. You can pass a nil func
// to unset the CAN interrupt. If you do so, the change parameter is ignored
// and can be set to any value (such as 0).
func (can *CAN) SetInterrupt(ie uint32, callback func(*CAN)) error {
	if callback == nil {
		// Disable this CAN interrupt
		can.Bus.IE.ClearBits(ie)
		return nil
	}
	can.Bus.IE.SetBits(ie)

	idx := 0
	switch can.Bus {
	case sam.CAN0:
		canInstances[0] = can
	case sam.CAN1:
		canInstances[1] = can
		idx = 1
	}

	for i := uint(0); i < 32; i++ {
		if ie&(1<<i) != 0 {
			canCallbacks[idx][i] = callback
		}
	}

	switch can.Bus {
	case sam.CAN0:
		interrupt.New(sam.IRQ_CAN0, func(interrupt.Interrupt) {
			ir := sam.CAN0.IR.Get()
			sam.CAN0.IR.Set(ir) // clear interrupt
			for i := uint(0); i < 32; i++ {
				if ir&(1<<i) != 0 && canCallbacks[0][i] != nil {
					canCallbacks[0][i](canInstances[0])
				}
			}
		}).Enable()
	case sam.CAN1:
		interrupt.New(sam.IRQ_CAN1, func(interrupt.Interrupt) {
			ir := sam.CAN1.IR.Get()
			sam.CAN1.IR.Set(ir) // clear interrupt
			for i := uint(0); i < 32; i++ {
				if ir&(1<<i) != 0 && canCallbacks[1][i] != nil {
					canCallbacks[1][i](canInstances[1])
				}
			}
		}).Enable()
	}

	return nil
}

// TxFifoIsFull returns whether TxFifo is full or not.
func (can *CAN) TxFifoIsFull() bool {
	return (can.Bus.TXFQS.Get() & sam.CAN_TXFQS_TFQF_Msk) == sam.CAN_TXFQS_TFQF_Msk
}

// TxRaw sends a CAN Frame according to CANTxBufferElement.
func (can *CAN) TxRaw(e *CANTxBufferElement) {
	putIndex := (can.Bus.TXFQS.Get() & sam.CAN_TXFQS_TFQPI_Msk) >> sam.CAN_TXFQS_TFQPI_Pos

	f := CANTxFifo[can.instance()][putIndex*(8+64) : (putIndex+1)*(8+64)]
	id := e.ID
	if !e.XTD {
		// standard identifier is stored into ID[28:18]
		id <<= 18
	}

	f[3] = byte(id >> 24)
	if e.ESI {
		f[3] |= 0x80
	}
	if e.XTD {
		f[3] |= 0x40
	}
	if e.RTR {
		f[3] |= 0x20
	}
	f[2] = byte(id >> 16)
	f[1] = byte(id >> 8)
	f[0] = byte(0)
	f[7] = e.MM
	f[6] = e.DLC
	if e.EFC {
		f[6] |= 0x80
	}
	if e.FDF {
		f[6] |= 0x20
	}
	if e.BRS {
		f[6] |= 0x10
	}
	f[5] = 0x00 // reserved
	f[4] = 0x00 // reserved

	length := CANDlcToLength(e.DLC, e.FDF)
	for i := byte(0); i < length; i++ {
		f[8+i] = e.DB[i]
	}

	can.Bus.TXBAR.SetBits(1 << putIndex)
}

// The Tx transmits CAN frames. It is easier to use than TxRaw, but not as
// flexible.
func (can *CAN) Tx(id uint32, data []byte, isFD, isExtendedID bool) {
	length := byte(len(data))
	dlc := CANLengthToDlc(length, true)

	e := CANTxBufferElement{
		ESI: false,
		XTD: isExtendedID,
		RTR: false,
		ID:  id,
		MM:  0x00,
		EFC: true,
		FDF: isFD,
		BRS: isFD,
		DLC: dlc,
	}

	if !isFD {
		if length > 8 {
			length = 8
		}
	}
	for i := byte(0); i < length; i++ {
		e.DB[i] = data[i]
	}

	can.TxRaw(&e)
}

// RxFifoSize returns the number of CAN Frames currently stored in the RXFifo.
func (can *CAN) RxFifoSize() int {
	sz := (can.Bus.RXF0S.Get() & sam.CAN_RXF0S_F0FL_Msk) >> sam.CAN_RXF0S_F0FL_Pos
	return int(sz)
}

// RxFifoIsFull returns whether RxFifo is full or not.
func (can *CAN) RxFifoIsFull() bool {
	sz := (can.Bus.RXF0S.Get() & sam.CAN_RXF0S_F0FL_Msk) >> sam.CAN_RXF0S_F0FL_Pos
	return sz == CANRxFifoSize
}

// RxFifoIsEmpty returns whether RxFifo is empty or not.
func (can *CAN) RxFifoIsEmpty() bool {
	sz := (can.Bus.RXF0S.Get() & sam.CAN_RXF0S_F0FL_Msk) >> sam.CAN_RXF0S_F0FL_Pos
	return sz == 0
}

// RxRaw copies the received CAN frame to CANRxBufferElement.
func (can *CAN) RxRaw(e *CANRxBufferElement) {
	idx := (can.Bus.RXF0S.Get() & sam.CAN_RXF0S_F0GI_Msk) >> sam.CAN_RXF0S_F0GI_Pos
	f := CANRxFifo[can.instance()][idx*(8+64):]

	e.ESI = false
	if (f[3] & 0x80) != 0x00 {
		e.ESI = true
	}

	e.XTD = false
	if (f[3] & 0x40) != 0x00 {
		e.XTD = true
	}

	e.RTR = false
	if (f[3] & 0x20) != 0x00 {
		e.RTR = true
	}

	id := ((uint32(f[3]) << 24) + (uint32(f[2]) << 16) + (uint32(f[1]) << 8) + uint32(f[0])) & 0x1FFFFFFF
	if (f[3] & 0x20) == 0 {
		id >>= 18
		id &= 0x000007FF
	}
	e.ID = id

	e.ANMF = false
	if (f[7] & 0x80) != 0x00 {
		e.ANMF = true
	}

	e.FIDX = f[7] & 0x7F

	e.FDF = false
	if (f[6] & 0x20) != 0x00 {
		e.FDF = true
	}

	e.BRS = false
	if (f[6] & 0x10) != 0x00 {
		e.BRS = true
	}

	e.DLC = f[6] & 0x0F

	e.RXTS = (uint16(f[5]) << 8) + uint16(f[4])

	for i := byte(0); i < CANDlcToLength(e.DLC, e.FDF); i++ {
		e.DB[i] = f[i+8]
	}

	can.Bus.RXF0A.ReplaceBits(idx, sam.CAN_RXF0A_F0AI_Msk, sam.CAN_RXF0A_F0AI_Pos)
}

// Rx receives a CAN frame. It is easier to use than RxRaw, but not as
// flexible.
func (can *CAN) Rx() (id uint32, dlc byte, data []byte, isFd, isExtendedID bool) {
	e := CANRxBufferElement{}
	can.RxRaw(&e)
	length := CANDlcToLength(e.DLC, e.FDF)
	return e.ID, length, e.DB[:length], e.FDF, e.XTD
}

func (can *CAN) instance() byte {
	if can.Bus == sam.CAN0 {
		return 0
	} else {
		return 1
	}
}

// CANTxBufferElement is a struct that corresponds to the same5x' Tx Buffer
// Element.
type CANTxBufferElement struct {
	ESI bool
	XTD bool
	RTR bool
	ID  uint32
	MM  uint8
	EFC bool
	FDF bool
	BRS bool
	DLC uint8
	DB  [64]uint8
}

// CANRxBufferElement is a struct that corresponds to the same5x Rx Buffer and
// FIFO Element.
type CANRxBufferElement struct {
	ESI  bool
	XTD  bool
	RTR  bool
	ID   uint32
	ANMF bool
	FIDX uint8
	FDF  bool
	BRS  bool
	DLC  uint8
	RXTS uint16
	DB   [64]uint8
}

// Data returns the received data as a slice of the size according to dlc.
func (e CANRxBufferElement) Data() []byte {
	return e.DB[:CANDlcToLength(e.DLC, e.FDF)]
}

// CANDlcToLength() converts a DLC value to its actual length.
func CANDlcToLength(dlc byte, isFD bool) byte {
	length := dlc
	if dlc == 0x09 {
		length = 12
	} else if dlc == 0x0A {
		length = 16
	} else if dlc == 0x0B {
		length = 20
	} else if dlc == 0x0C {
		length = 24
	} else if dlc == 0x0D {
		length = 32
	} else if dlc == 0x0E {
		length = 48
	} else if dlc == 0x0F {
		length = 64
	}
	return length

}

// CANLengthToDlc() converts its actual length to a DLC value.
func CANLengthToDlc(length byte, isFD bool) byte {
	dlc := length
	if length <= 0x08 {
	} else if length <= 12 {
		dlc = 0x09
	} else if length <= 16 {
		dlc = 0x0A
	} else if length <= 20 {
		dlc = 0x0B
	} else if length <= 24 {
		dlc = 0x0C
	} else if length <= 32 {
		dlc = 0x0D
	} else if length <= 48 {
		dlc = 0x0E
	} else if length <= 64 {
		dlc = 0x0F
	}
	return dlc
}
