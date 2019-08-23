package rpi3

import "unsafe"
import "runtime/volatile"

//this is really a C-style array with 128 elements of type uint32
// xxx not concurrent safe! xxx
var mboxData [512]byte

// Make a mailbox call. Returns true if the message has been returned successfully.
// XXX verify the return value logic is ok
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

// FBInfo is a struct that holds the parameters of the screen as well as a
//pointer to the framebuffer  itself.
type FBInfo struct {
	Height int
	Width  int
	Pitch  int
	Ptr    unsafe.Pointer
	Size   int
}

//FrameBufferInfo is an FBInfo struct that gets filled in by a successful
//call to InitFramebuffer.
var FrameBufferInfo FBInfo

// this is called hackdata because it SHOULD be the case that we can put this inside
// the function below.  However, we cannot because tinygo+llvm will generate an
// unaligned access (sp + constant that is not aligned) and the code generates
// an exception.  For some reason, it's ok when the code is here.
var hackdata = []uint32{
	35 * 4,
	MBOX_REQUEST,

	0x48003, //set phy wh
	8,
	8,
	1024, //FrameBufferInfo.width -- 1024,1920
	768,  //FrameBufferInfo.height -- 768,1200

	0x48004, //set virt wh
	8,
	8,
	1024, //FrameBufferInfo.virtual_width
	3072, //FrameBufferInfo.virtual_height

	0x48009, //set virt offset
	8,
	8,
	0, //FrameBufferInfo.x_offset
	0, //FrameBufferInfo.y.offset

	0x48005, //set depth
	4,
	4,
	32, //FrameBufferInfo.depth

	0x48006, //set pixel order
	4,
	4,
	1, //RGB, not BGR preferably

	0x40001, //get framebuffer, gets alignment on request
	8,
	8,
	4096, //FrameBufferInfo.pointer
	0,    //FrameBufferInfo.size

	0x40008, //get pitch
	4,
	4,
	0, //FrameBufferInfo.pitch
	MBOX_TAG_LAST,
}

//InitFramebuffer does a series of Mbox calls to the GPU to init the framebuffer
//hardware. On success, results are placed into FrameBufferInfo.
//go:export InitFramebuffer
func InitFramebuffer() bool {
	mbox := align16bytes(uintptr(unsafe.Pointer(&mboxData)))

	for i := range hackdata {
		volatile.StoreUint32((*uint32)(unsafe.Pointer(uintptr(mbox)+uintptr(i*4))), hackdata[i])
	}

	channel := byte(MBOX_CH_PROP)
	if MboxCall(byte(channel)) && MboxResultSlot(20) == 32 && MboxResultSlot(28) != 0 {
		FrameBufferInfo.Ptr = unsafe.Pointer(uintptr(MboxResultSlot(28) & 0x3FFFFFFF))
		FrameBufferInfo.Width = int(MboxResultSlot(5))
		FrameBufferInfo.Height = int(MboxResultSlot(6))
		FrameBufferInfo.Pitch = int(MboxResultSlot(33))
		return true
	}

	return false
}

// SetYScroll sets the Y position of the screen against the virtual framebuffer.
func SetYScroll(yoffset uint16) bool {
	mboxSetupData(
		7*4,
		MBOX_REQUEST,
		0x48009, //set virt offset
		8,
		8,
		0,               //FrameBufferInfo.x_offset
		uint32(yoffset), //FrameBufferInfo.y.offset
		MBOX_TAG_LAST,
	)

	channel := byte(MBOX_CH_PROP)
	_ = MboxCall(byte(channel))
	newValue := MboxResultSlot(6)
	if newValue != uint32(yoffset) {
		print("could not set to yoffset ", yoffset, " and returned ", newValue, "\n")
		print("do you need set max_framebuffer_height=1536 in your config.txt?\n")
		Abort()
	}
	return true
}
