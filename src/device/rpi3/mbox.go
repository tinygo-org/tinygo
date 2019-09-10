package rpi3

import "unsafe"
import "runtime/volatile"

//this is really a C-style array with 36 elements of type uint32
var mboxData [256]byte

//
// Make a mailbox call. Returns true if the message has been returned successfully.
//
func MboxCall(ch byte) bool {
	q := uint32(0)
	mbox := uintptr(unsafe.Pointer(&mboxData)) //64 is bigger than 36 needed, but alloc produces only 4 byte aligned
	//addrMbox := uintptr(unsafe.Pointer(mbox)) & uintptr(f)
	//or on the channel number to he end
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
	mbox := align16bytes(uintptr(unsafe.Pointer(&mboxData)))

	mboxData := []uint32{
		8 * 4,
		MBOX_REQUEST,
		MBOX_TAG_GETSERIAL,
		8, //buffer size
		8,
		0, //clear output buffer
		0,
		MBOX_TAG_LAST,
	}

	channel := MBOX_CH_PROP
	for i := 0; i < len(mboxData); i++ {
		volatile.StoreUint32((*uint32)(unsafe.Pointer(uintptr(mbox)+uintptr(i*4))), mboxData[i])
	}
	if !MboxCall(byte(channel)) {
		return 0xffff, 0xffff
	}

	hi := MboxResultSlot(6)
	lo := MboxResultSlot(5) //little endian
	print("hex values:")
	UART0Hex(hi)
	UART0Hex(lo)

	return hi, lo
}

func MboxResultSlot(i int) uint32 {
	//32 bit units
	mbox := align16bytes(uintptr(unsafe.Pointer(&mboxData)))
	return volatile.LoadUint32((*uint32)(unsafe.Pointer(uintptr(mbox) + uintptr(i*4))))
}

func LEDSet(on bool) bool {
	mbox := align16bytes(uintptr(unsafe.Pointer(&mboxData)))

	val := uint32(0)
	if on {
		val = 0xffff
	}
	mboxData := []uint32{
		8 * 4,
		0, //no response expected
		MBOX_TAG_SET_GPIO_STATE,
		8, //buffer size
		8,
		130, //port output 130
		val,
		MBOX_TAG_LAST,
	}

	channel := MBOX_CH_PROP
	for i := 0; i < len(mboxData); i++ {
		volatile.StoreUint32((*uint32)(unsafe.Pointer(uintptr(mbox)+uintptr(i*4))), mboxData[i])
	}
	MboxCall(byte(channel))
	if (mboxData[1] == 0x80000000) && (mboxData[4] == 0x80000008) {
		return true
	}
	return false
}
