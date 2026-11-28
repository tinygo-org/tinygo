// Package esp32p4sdmmc contains the hardware-independent pieces of the
// ESP32-P4 SDMMC host driver: command encoding, clock divider selection,
// CSD decoding, and IDMAC descriptor layout.
//
// Keeping these pieces free of volatile register access makes them testable
// on the host, while machine_esp32p4_sdmmc.go handles the actual MMIO.
package esp32p4sdmmc

// SD command indices used by the driver.
const (
	CmdGoIdleState        = 0
	CmdAllSendCID         = 2
	CmdSendRelativeAddr   = 3
	CmdSwitchFunc         = 6
	CmdSelectCard         = 7
	CmdSendIfCond         = 8
	CmdSendCSD            = 9
	CmdSendCID            = 10
	CmdStopTransmission   = 12
	CmdSendStatus         = 13
	CmdSetBlockLen        = 16
	CmdReadSingleBlock    = 17
	CmdReadMultipleBlock  = 18
	CmdSetBlockCount      = 23 // ACMD23
	CmdWriteSingleBlock   = 24
	CmdWriteMultipleBlock = 25
	CmdAppCmd             = 55
	CmdSendOpCond         = 41 // ACMD41
	CmdSetBusWidth        = 6  // ACMD6
)

// ResponseType describes the response expected from a command.
type ResponseType uint8

const (
	RespNone ResponseType = iota
	RespR1
	RespR1b
	RespR2
	RespR3
	RespR6
	RespR7
)

// CommandOptions describes one command for the SDHOST CMD register.
type CommandOptions struct {
	Index         uint8
	Response      ResponseType
	Data          bool
	Write         bool
	AutoStop      bool
	CheckCRC      bool
	CardNum       uint32
	SendInit      bool
	StopAbort     bool
	VoltageSwitch bool
	UpdateClock   bool
}

// Word returns the value to write to the SDHOST CMD register.
func (o CommandOptions) Word() uint32 {
	var w uint32

	w |= uint32(o.Index) & 0x3f
	if o.Response != RespNone {
		w |= 1 << 6 // response_expect
		if o.Response == RespR2 {
			w |= 1 << 7 // response_long
		}
	}
	if o.CheckCRC {
		w |= 1 << 8
	}
	if o.Data {
		w |= 1 << 9 // data_expected
		if o.Write {
			w |= 1 << 10
		}
	}
	if o.AutoStop {
		w |= 1 << 12
	}
	if o.Index == CmdGoIdleState && !o.UpdateClock {
		o.SendInit = true
	}
	if o.Index == CmdStopTransmission {
		o.StopAbort = true
	}
	if o.SendInit {
		w |= 1 << 15
	} else if !o.StopAbort && !o.VoltageSwitch && !o.UpdateClock {
		w |= 1 << 13 // wait_complete
	}
	if o.StopAbort {
		w |= 1 << 14
	}
	w |= (o.CardNum & 0x1f) << 16
	if o.UpdateClock {
		w |= 1 << 21
	}
	if o.VoltageSwitch {
		w |= 1 << 28
	}
	w |= 1 << 29 // use_hold_reg (hardware default)
	w |= 1 << 31 // start_command
	return w
}

// ClockDiv returns the host and card clock dividers for a target card
// clock. hostDiv divides the 160 MHz SDMMC source; cardDiv is the SDHOST
// divider and divides by 2*cardDiv (0 means bypass).
func ClockDiv(srcHz, targetHz uint32) (hostDiv, cardDiv uint32) {
	if targetHz == 0 {
		targetHz = 400_000
	}
	// Match the ESP-IDF probing choice on the P4: 160 MHz / 10 / (2*20).
	if targetHz == 400_000 {
		return 10, 20
	}

	total := (srcHz + targetHz - 1) / targetHz
	if total <= 16 {
		if total == 0 {
			total = 1
		}
		return total, 0
	}

	// Two-stage divider: keep the intermediate clock as high as possible
	// while fitting the card divider in 8 bits.
	for h := uint32(2); h <= 16; h++ {
		intermediate := srcHz / h
		c := (intermediate + 2*targetHz - 1) / (2 * targetHz)
		if c < 1 {
			c = 1
		}
		if c <= 255 {
			return h, c
		}
	}
	return 2, 255
}

// ClockEdge returns the HP_SYS_CLKRST edge fields for a host divider.
// A divider of 1 bypasses the edge divider.
func ClockEdge(div uint32) (h, n, l uint32, update bool) {
	if div <= 1 {
		return 0, 0, 0, false
	}
	return div/2 - 1, div - 1, div - 1, true
}

// CSD version helpers.

// CSDVersion returns 1 or 2 for the CSD structure version.
func CSDVersion(csd [16]byte) uint8 {
	return 1 + csd[0]>>6
}

// CSDReadBlockLen returns the read block length in bytes.
func CSDReadBlockLen(csd [16]byte) int {
	return 1 << (csd[5] & 0x0f)
}

// CSDCapacityBytes decodes the total card capacity from a CSD.
func CSDCapacityBytes(csd [16]byte) int64 {
	switch CSDVersion(csd) {
	case 2:
		return int64(csdV2CSize(csd)+1) * 512 * 1024
	case 1:
		csize := int64(csdV1CSize(csd))
		mult := int64(1) << (2 + csdV1CSizeMult(csd))
		return (csize + 1) * mult * int64(CSDReadBlockLen(csd))
	default:
		return 0
	}
}

// CSDBlockCount returns the number of 512-byte blocks on the card.
func CSDBlockCount(csd [16]byte) int64 {
	if CSDReadBlockLen(csd) != 512 {
		// SDHC/SDXC always use 512-byte blocks; for SDSC the card is put
		// into 512-byte block mode by CMD16 before this is used.
		return CSDCapacityBytes(csd) / 512
	}
	return CSDCapacityBytes(csd) / 512
}

func csdV2CSize(csd [16]byte) uint32 {
	return uint32(csd[7]&0x3f)<<16 | uint32(csd[8])<<8 | uint32(csd[9])
}

func csdV1CSize(csd [16]byte) uint16 {
	return uint16(csd[8]>>6) | uint16(csd[7])<<2 | uint16(csd[6]&0b11)<<10
}

func csdV1CSizeMult(csd [16]byte) uint8 {
	return (csd[9]&0b11)<<1 | csd[10]>>7
}

// ResponseBytes converts the four SDHOST response words (RESP0..RESP3) into
// the 16-byte big-endian representation used by CID/CSD. RESP0 is the least
// significant word of the 128-bit response.
func ResponseBytes(resp [4]uint32) [16]byte {
	var b [16]byte
	for i := 0; i < 4; i++ {
		w := resp[3-i]
		b[i*4+0] = byte(w >> 24)
		b[i*4+1] = byte(w >> 16)
		b[i*4+2] = byte(w >> 8)
		b[i*4+3] = byte(w)
	}
	return b
}

// R1 response bits.
const (
	R1ReadyForData = 1 << 8
	R1AppCmd       = 1 << 5
	R1ErrorMask    = 0xffff0000
)

// R1HasError reports whether the R1 response carries a card error.
func R1HasError(r1 uint32) bool {
	return r1&R1ErrorMask != 0
}

// OCR bits.
const (
	OCRReady   = 1 << 31
	OCRCCS     = 1 << 30
	OCRVoltage = 0x00300000 // 3.2-3.3 V and 3.3-3.4 V
)

// Descriptor is the 64-byte SDMMC IDMAC descriptor. Only the first four
// words are used; the remaining twelve pad the descriptor to the P4 L1
// cache line size.
type Descriptor struct {
	Word [16]uint32
}

// Descriptor size and maximum buffer length.
const (
	DescriptorSize     = 64
	MaxDescriptorBytes = 4096
)

// Descriptor word 0 bits.
const (
	DescLast           = 1 << 2
	DescFirst          = 1 << 3
	DescSecondChained  = 1 << 4
	DescEndOfRing      = 1 << 5
	DescCardErrSummary = 1 << 30
	DescOwnedByIDMAC   = 1 << 31
)

// SetFirst marks the descriptor as the first in a chain.
func (d *Descriptor) SetFirst(first bool) {
	set := func(v bool, bit uint32) {
		if v {
			d.Word[0] |= bit
		} else {
			d.Word[0] &^= bit
		}
	}
	set(first, DescFirst)
}

// SetLast marks the descriptor as the last in a chain.
func (d *Descriptor) SetLast(last bool) {
	if last {
		d.Word[0] |= DescLast
	} else {
		d.Word[0] &^= DescLast
	}
}

// SetOwned transfers ownership to or from the IDMAC.
func (d *Descriptor) SetOwned(owned bool) {
	if owned {
		d.Word[0] |= DescOwnedByIDMAC
	} else {
		d.Word[0] &^= DescOwnedByIDMAC
	}
}

// Owned reports whether the IDMAC currently owns the descriptor.
func (d *Descriptor) Owned() bool {
	return d.Word[0]&DescOwnedByIDMAC != 0
}

// SetBuffer sets the first buffer address and size.
func (d *Descriptor) SetBuffer(addr uint32, size int) {
	d.Word[0] |= DescSecondChained
	d.Word[1] = uint32(size) & 0x1fff
	d.Word[2] = addr
}

// SetNext sets the next descriptor address.
func (d *Descriptor) SetNext(addr uint32) {
	d.Word[3] = addr
}
