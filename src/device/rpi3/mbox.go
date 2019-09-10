package rpi3

import "unsafe"
import "runtime/volatile"

//this is really a C-style array with 36 elements of type uint32
// xxx not concurrent safe! xxx
var mboxData [256]byte

//
// Make a mailbox call. Returns true if the message has been returned successfully.
//
func MboxCall(ch byte) bool {
	q := uint32(0)
	mbox := uintptr(unsafe.Pointer(&mboxData)) //64 is bigger than 36 needed, but alloc produces only 4 byte aligned

	r := (uint32)(mbox | uintptr(uint64(ch)&0xF))
	for volatile.LoadUint32((*uint32)(MBOX_STATUS))&MBOX_FULL != 0 {
		volatile.StoreUint32(&q, volatile.LoadUint32(&q)+1)
	}
	volatile.StoreUint32((*uint32)(MBOX_WRITE), r)
	for {
		for volatile.LoadUint32((*uint32)(MBOX_STATUS))&MBOX_EMPTY != 0 {
			volatile.StoreUint32(&q, volatile.LoadUint32(&q)+1)
		}
		if r == volatile.LoadUint32((*uint32)(MBOX_READ)) {
			resp := volatile.LoadUint32((*uint32)(unsafe.Pointer(uintptr(mbox) + uintptr(1*4))))
			return resp == MBOX_RESPONSE
		}
	}
	return false
}

// this returns two 32bit numbers that make the 64bit id
// returns the hi 32bits first
// returns both values 0 on qemu
// and returns 0xffff in both if there was an error
func GetRPIID() (uint32, uint32) {
	mboxSetupData(
		8*4,
		MBOX_REQUEST,
		MBOX_TAG_GETSERIAL,
		8, //buffer size
		8,
		0, //clear output buffer
		0,
		MBOX_TAG_LAST,
	)

	channel := MBOX_CH_PROP
	if !MboxCall(byte(channel)) {
		return 0xffff, 0xffff
	}

	hi := MboxResultSlot(6)
	lo := MboxResultSlot(5) //little endian

	return hi, lo
}

func MboxResultSlot(i int) uint32 {
	//32 bit units
	mbox := align16bytes(uintptr(unsafe.Pointer(&mboxData)))
	return volatile.LoadUint32((*uint32)(unsafe.Pointer(uintptr(mbox) + uintptr(i*4))))
}

func mboxSetupData(data ...uint32) {
	mbox := align16bytes(uintptr(unsafe.Pointer(&mboxData)))
	for i := range data {
		volatile.StoreUint32((*uint32)(unsafe.Pointer(uintptr(mbox)+uintptr(i*4))), data[i])
	}
}

func LEDSet(on bool) bool {
	value := uint32(0)
	if on {
		value = 0xffff
	}
	mboxSetupData(
		8*4,
		0, //no response expected
		MBOX_TAG_SET_GPIO_STATE,
		8, //buffer size
		8,
		130, //port output 130
		value,
		MBOX_TAG_LAST,
	)
	channel := MBOX_CH_PROP
	MboxCall(byte(channel))
	s1 := MboxResultSlot(1)
	s2 := MboxResultSlot(4)

	if (s1 == 0x80000000) && (s2 == 0x80000008) {
		return true
	}
	return false
}
