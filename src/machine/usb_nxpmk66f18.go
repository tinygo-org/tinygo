// +build nxp,mk66f18

package machine

import (
	"device/nxp"
	"fmt"
	"reflect"
	"runtime/interrupt"
	"runtime/volatile"
	"unsafe"
)

const debugUSB = true

var USB0 = USB{USB0_Type: nxp.USB0, SCGC: &nxp.SIM.SCGC4, SCGCMask: nxp.SIM_SCGC4_USBOTG, ENDPT: (*[16]register8Aligned32)(unsafe.Pointer(&nxp.USB0.ENDPT0))}

//go:align 512
var usbBufferDescriptorRecords [usb_MaxBDTSize]usbBufferDescriptor
var usbBufferDescriptorTable = usbBufferDescriptorTable_t{records: &usbBufferDescriptorRecords}

func init() {
	USB0.Interrupt = interrupt.New(nxp.IRQ_USB0, USB0.serviceInterrupt)
}

//go:linkname millisSinceBoot runtime.millisSinceBoot
func millisSinceBoot() uint64

type USB struct {
	*nxp.USB0_Type
	SCGC      *volatile.Register32
	SCGCMask  uint32
	Interrupt interrupt.Interrupt
	ENDPT     *[16]register8Aligned32

	CDCRXBuffer RingBuffer
	Configured  bool

	Log []string
}

func (u *USB) logf(format string, args ...interface{}) {
	u.Log = append(u.Log, fmt.Sprintf(format, args...))
}

func (u *USB) debugf(format string, args ...interface{}) {
	if debugUSB {
		u.logf(format, args...)
	}
}

func (u *USB) DumpLog() []string {
	var log []string

	state := interrupt.Disable()
	log, u.Log = u.Log, nil
	interrupt.Restore(state)

	if log == nil {
		return []string{}
	}
	return log
}

func (u *USB) Configure() {
	if u.Configured {
		return
	}
	u.Configured = true

	for i := 0; i < usb_MaxBDTSize; i++ {
		usbBufferDescriptorTable.Entry(i).Reset()
	}

	// this basically follows the flowchart in the Kinetis
	// Quick Reference User Guide, Rev. 1, 03/2012, page 141

	// assume 48 MHz clock already running
	// SIM - enable clock
	u.SCGC.SetBits(u.SCGCMask)
	nxp.SYSMPU.RGDAAC0.SetBits(0x03000000)

	// if using IRC48M, turn on the USB clock recovery hardware
	u.CLK_RECOVER_IRC_EN.Set(nxp.USB0_CLK_RECOVER_IRC_EN_IRC_EN | nxp.USB0_CLK_RECOVER_IRC_EN_REG_EN)
	u.CLK_RECOVER_CTRL.Set(nxp.USB0_CLK_RECOVER_CTRL_CLOCK_RECOVER_EN | nxp.USB0_CLK_RECOVER_CTRL_RESTART_IFRTRIM_EN)

	// // reset USB module
	// u.USBTRC0.Set(nxp.USB0_USBTRC_USBRESET)
	// for u.USBTRC0.HasBits(nxp.USB0_USBTRC_USBRESET) {} // wait for reset to end

	// set desc table base addr
	table := usbBufferDescriptorTable.Address()
	u.BDTPAGE1.Set(uint8(table >> 8))
	u.BDTPAGE2.Set(uint8(table >> 16))
	u.BDTPAGE3.Set(uint8(table >> 24))

	// clear all ISR flags
	u.ISTAT.Set(0xFF)
	u.ERRSTAT.Set(0xFF)
	u.OTGISTAT.Set(0xFF)

	//u.USBTRC0.SetBits(0x40) // undocumented bit

	// enable USB
	u.CTL.Set(nxp.USB0_CTL_USBENSOFEN)
	u.USBCTRL.Set(0)

	// enable reset interrupt
	u.INTEN.Set(nxp.USB0_INTEN_USBRSTEN)

	u.Interrupt.SetPriority(112)
	u.Interrupt.Enable()

	// enable d+ pullup
	u.CONTROL.Set(nxp.USB0_CONTROL_DPPULLUPNONOTG)
	u.Log = append(u.Log, "done")
}

func (u *USB) serviceInterrupt(interrupt.Interrupt) {
	// from: usb_isr

	// // ARM errata 838869: Store immediate overlapping exception return operation
	// // might vector to incorrect interrupt
	// defer func() { arm.Asm("dsb 0xF") }()

again:
	status := u.ISTAT.Get() & u.INTEN.Get()

	if status&^nxp.USB0_ISTAT_SOFTOK != 0 {
		// log interrupt status, but not start of frame token
		u.debugf("istat 0x%02X", status)
	}

	if status&nxp.USB0_ISTAT_SOFTOK != 0 {
		// TODO: usb reboot timer, cdc transmit flush timer
		u.ISTAT.Set(nxp.USB0_ISTAT_SOFTOK)
	}

	if status&nxp.USB0_ISTAT_TOKDNE != 0 {
		stat := u.STAT.Get()
		u.debugf("stat 0x%02X", stat)

		if ep := stat >> 4; ep == 0 {
			u.controlPacket(stat)
		} else {
			panic("can't handle endpoints beyond 0")
		}

		u.ISTAT.Set(nxp.USB0_ISTAT_TOKDNE)
		goto again
	}

	if status&nxp.USB0_ISTAT_USBRST != 0 {
		// initialize BDT toggle bits
		u.CTL.Set(nxp.USB0_CTL_ODDRST)
		ep0_tx_bdt_bank = false

		// set up buffers to receive Setup and OUT packets
		usbBufferDescriptorTable.Get(0, false, false).Set(ep0_rx0_buf[:], false) // EP0 RX even
		usbBufferDescriptorTable.Get(0, false, true).Set(ep0_rx1_buf[:], false)  // EP0 RX odd
		usbBufferDescriptorTable.Get(0, true, false).Reset()                     // EP0 TX even
		usbBufferDescriptorTable.Get(0, true, true).Reset()                      // EP0 TX odd

		// activate endpoint 0
		u.configureEndpoint(0, USBEndpointConfiguration{
			Transmit:  true,
			Receive:   true,
			Handshake: true,
		})

		// clear all ending interrupts
		u.ERRSTAT.Set(0xFF)
		u.ISTAT.Set(0xFF)

		// set the address to zero during enumeration
		u.ADDR.Set(0)

		// enable other interrupts
		u.ERREN.Set(0xFF)
		u.INTEN.Set(nxp.USB0_INTEN_TOKDNEEN |
			nxp.USB0_INTEN_SOFTOKEN |
			nxp.USB0_INTEN_STALLEN |
			nxp.USB0_INTEN_ERROREN |
			nxp.USB0_INTEN_USBRSTEN |
			nxp.USB0_INTEN_SLEEPEN)

		// is this necessary?
		u.CTL.Set(nxp.USB0_CTL_USBENSOFEN)
		return
	}

	if status&nxp.USB0_ISTAT_STALL != 0 {
		// u.unstall(0)
		u.ENDPT0.Set(nxp.USB0_ENDPT_EPRXEN | nxp.USB0_ENDPT_EPTXEN | nxp.USB0_ENDPT_EPHSHK)
		u.ISTAT.Set(nxp.USB0_ISTAT_STALL)
	}

	if status&nxp.USB0_ISTAT_ERROR != 0 {
		err := u.ERRSTAT.Get()
		u.ERRSTAT.Set(err)
		u.logf("errstat 0x%02X", err)
		u.ISTAT.Set(nxp.USB0_ISTAT_ERROR)
	}

	if status&nxp.USB0_ISTAT_SLEEP != 0 {
		u.ISTAT.Set(nxp.USB0_ISTAT_SLEEP)
	}
}

func (u *USB) controlPacket(stat uint8) {
	// from: usb_control

	descriptor := usbBufferDescriptorTable.GetStat(stat)

	pid := descriptor.TokenPID()
	u.debugf("bd tok 0x%X", pid)

	switch pid {
	case usb_TOK_PID_OUT, 0x2: // OUT transaction received from host
		// give the buffer back
		descriptor.Describe(usb_BUFFER_SIZE, true)

	case usb_TOK_PID_IN: // IN transaction completed to host
		if len(ep0_tx_remainder) > 0 {
			ep0_tx_remainder = sendUSB0EP0Chunk(ep0_tx_remainder)
		}

		if addr := usb0_pending_address_change; addr != nil {
			usb0_pending_address_change = nil
			u.ADDR.Set(*addr)
		}

	case usb_TOK_PID_SETUP:
		// read setup info
		setup := newUSBSetup(descriptor.Data())
		u.debugf("setup request=%x type=%x value=%x index=%x", setup.bRequest, setup.bmRequestType, setup.wValue(), setup.wIndex)

		// give the buffer back
		descriptor.Describe(usb_BUFFER_SIZE, true)

		// clear any leftover pending IN transactions
		ep0_tx_remainder = nil
		usbBufferDescriptorTable.Get(0, true, false).Reset() // EP0 TX even
		usbBufferDescriptorTable.Get(0, true, true).Reset()  // EP0 TX odd

		// first IN after Setup is always DATA1
		ep0_tx_data_toggle = true

		if !u.applySetup(setup) {
			u.stall(0)
		}
	}

	// unfreeze the USB, now that we're ready (clear TXSUSPENDTOKENBUSY bit)
	u.CTL.Set(nxp.USB0_CTL_USBENSOFEN)
}

func (u *USB) applySetup(setup usbSetup) bool {
	// from: usb_setup

	requestDirection := setup.bmRequestType & usb_REQUEST_DIRECTION
	requestType := setup.bmRequestType & usb_REQUEST_TYPE
	requestRecipient := setup.bmRequestType & usb_REQUEST_RECIPIENT

	if requestType == usb_REQUEST_CLASS && requestRecipient == usb_REQUEST_INTERFACE {
		switch setup.bRequest {
		case usb_CDC_SET_LINE_CODING:
			sendZlp()
			return true

		case usb_CDC_SET_CONTROL_LINE_STATE:
			// usb_cdc_line_rtsdtr_millis.Set(uint32(millisSinceBoot()))
			// usb_cdc_line_rtsdtr.Set(uint32(setup.wValue()))
			sendZlp()
			return true

		case usb_CDC_SEND_BREAK:
			sendZlp()
			return true
		}

	} else if requestType == usb_REQUEST_STANDARD {
		switch setup.bRequest {
		case usb_GET_STATUS:
			if requestDirection == usb_REQUEST_DEVICETOHOST {
				if requestRecipient == usb_REQUEST_DEVICE {
					sendUSB0EP0([]byte{0, 0})
					return true
				}

				if requestRecipient == usb_REQUEST_ENDPOINT {
					i := int(setup.wIndex & 0x7F)
					if i > len(u.ENDPT) {
						return false
					}

					b := []byte{0, 0}
					if u.ENDPT[i].HasBits(nxp.USB0_ENDPT_EPSTALL) {
						b[0] = 1
					}
					sendUSB0EP0(b)
					return true
				}
			}

		case usb_CLEAR_FEATURE:
			if requestDirection == usb_REQUEST_HOSTTODEVICE && requestRecipient == usb_REQUEST_ENDPOINT {
				i := int(setup.wIndex & 0x7F)
				if i > len(u.ENDPT) || setup.wValue() != 0 {
					return false
				}

				u.unstall(i)
				sendZlp()
				return true
			}

		case usb_SET_FEATURE:
			if requestDirection == usb_REQUEST_HOSTTODEVICE && requestRecipient == usb_REQUEST_ENDPOINT {
				i := int(setup.wIndex & 0x7F)
				if i > len(u.ENDPT) || setup.wValue() != 0 {
					return false
				}

				u.stall(i)
				sendZlp()
				return true
			}

		case usb_SET_ADDRESS:
			if requestDirection == usb_REQUEST_HOSTTODEVICE && requestRecipient == usb_REQUEST_DEVICE {
				v := setup.wValueL
				usb0_pending_address_change = &v
				sendZlp()
				return true
			}

		case usb_GET_DESCRIPTOR:
			if requestDirection == usb_REQUEST_DEVICETOHOST && (requestRecipient == usb_REQUEST_DEVICE || requestRecipient == usb_REQUEST_INTERFACE) {
				sendDescriptor(setup)
				return true
			}

		case usb_SET_DESCRIPTOR:
			return false

		case usb_GET_CONFIGURATION:
			if requestDirection == usb_REQUEST_DEVICETOHOST && requestRecipient == usb_REQUEST_DEVICE {
				sendUSB0EP0([]byte{usb0_configuration.Get()})
				return true
			}

		case usb_SET_CONFIGURATION:
			if requestDirection == usb_REQUEST_HOSTTODEVICE && requestRecipient == usb_REQUEST_DEVICE {
				usb0_configuration.Set(setup.wValueL)

				// clear all BDT entries, free any allocated memory...
				for i := 4; i < usb_MaxBDTSize; i++ {
					e := usbBufferDescriptorTable.Entry(i)
					if e.Owner() == usbBufferOwnedByUSBFS {
						e.retain = nil
						e.record.address.Set(0)
					}
				}

				u.configureEndpoint(usb_CDC_ENDPOINT_ACM, USBEndpointConfiguration{
					Transmit:     true,
					Handshake:    true,
					DisableSetup: true,
				})
				u.configureEndpoint(usb_CDC_ENDPOINT_OUT, USBEndpointConfiguration{
					Receive:      true,
					Handshake:    true,
					DisableSetup: true,
				})
				u.configureEndpoint(usb_CDC_ENDPOINT_IN, USBEndpointConfiguration{
					Transmit:     true,
					Handshake:    true,
					DisableSetup: true,
				})

				usbBufferDescriptorTable.Get(usb_CDC_ENDPOINT_ACM, true, false).Reset()
				usbBufferDescriptorTable.Get(usb_CDC_ENDPOINT_ACM, true, true).Reset()
				usbBufferDescriptorTable.Get(usb_CDC_ENDPOINT_OUT, true, false).Reset()
				usbBufferDescriptorTable.Get(usb_CDC_ENDPOINT_OUT, true, true).Reset()
				usbBufferDescriptorTable.Get(usb_CDC_ENDPOINT_IN, true, false).Reset()
				usbBufferDescriptorTable.Get(usb_CDC_ENDPOINT_IN, true, true).Reset()

				usbBufferDescriptorTable.Get(usb_CDC_ENDPOINT_OUT, false, false).Set(make([]byte, usb_BUFFER_SIZE), false)
				usbBufferDescriptorTable.Get(usb_CDC_ENDPOINT_OUT, false, true).Set(make([]byte, usb_BUFFER_SIZE), true)

				sendZlp()
				return true
			}
		}
	}

	return false
}

func (u *USB) stall(n int) {
	u.ENDPT[n].SetBits(nxp.USB0_ENDPT_EPSTALL)
}

func (u *USB) unstall(n int) {
	u.ENDPT[n].ClearBits(nxp.USB0_ENDPT_EPSTALL)
}

func (u *USB) configureEndpoint(n int, c USBEndpointConfiguration) {
	var v uint8
	if c.Transmit {
		v |= nxp.USB0_ENDPT_EPTXEN
	}
	if c.Receive {
		v |= nxp.USB0_ENDPT_EPRXEN
	}
	if c.Handshake {
		v |= nxp.USB0_ENDPT_EPHSHK
	}
	if c.DisableSetup {
		v |= nxp.USB0_ENDPT_EPCTLDIS
	}
	u.ENDPT[n].Set(v)
}

var (
	usb0_configuration          volatile.Register8
	usb0_pending_address_change *uint8

	//go:align 4
	ep0_rx0_buf [usb_BUFFER_SIZE]byte

	//go:align 4
	ep0_rx1_buf [usb_BUFFER_SIZE]byte

	ep0_tx_remainder   []byte
	ep0_tx_bdt_bank    bool
	ep0_tx_data_toggle bool
)

func sendUSB0EP0Chunk(data []byte) (remainder []byte) {
	if len(data) > usb_EP0_SIZE {
		data, remainder = data[:usb_EP0_SIZE], data[usb_EP0_SIZE:]
	}

	USB0.debugf("ep0 transmit %d byte(s): %x", len(data), data)

	bd := usbBufferDescriptorTable.Get(0, true, ep0_tx_bdt_bank)
	bd.Set(data, ep0_tx_data_toggle)
	ep0_tx_data_toggle = !ep0_tx_data_toggle
	ep0_tx_bdt_bank = !ep0_tx_bdt_bank
	return remainder
}

func sendUSB0EP0(data []byte) {
	// write to first bank
	data = sendUSB0EP0Chunk(data)
	if len(data) == 0 {
		return
	}

	// write to second bank
	data = sendUSB0EP0Chunk(data)
	if len(data) == 0 {
		return
	}

	// save remainder
	ep0_tx_remainder = data
}

func sendZlp() {
	sendUSB0EP0(nil)
}

func sendUSBPacket(ep uint16, data []byte) {
	if ep != 0 {
		panic("not ready to do that")
	}

	sendUSB0EP0(data)
}

func (u *USB) CDC() USBCDC {
	return USBCDC{&u.CDCRXBuffer}
}

type USBCDC struct {
	Buffer *RingBuffer
}

func (u USBCDC) WriteByte(byte) {
	panic("not implemented")
}

const (
	usb_MaxEndpointCount = 16
	usb_MaxBDTSize       = usb_MaxEndpointCount * 2 * 2

	usb_BDT_Own   = 0x80
	usb_BDT_Data1 = 0x40
	usb_BDT_Keep  = 0x20
	nsb_BDT_NoInc = 0x10
	usb_BDT_DTS   = 0x08
	usb_BDT_Stall = 0x04

	usb_BDT_TOK_PID_Mask = 0b00111100
	usb_BDT_TOK_PID_Pos  = 2

	usb_TOK_PID_OUT   = 0x1
	usb_TOK_PID_IN    = 0x9
	usb_TOK_PID_SETUP = 0xD
)

type register8Aligned32 struct {
	volatile.Register8
	_ [3]byte
}

type usbBufferDescriptor struct {
	descriptor volatile.Register32
	address    volatile.Register32
}

type usbBufferDescriptorTable_t struct {
	records *[usb_MaxBDTSize]usbBufferDescriptor
	retain  [usb_MaxBDTSize][]byte
}

type usbBufferDescriptorEntry struct {
	record *usbBufferDescriptor
	retain *[]byte
}

type usbBufferDescriptorOwner bool

const usbBufferOwnedByUSBFS usbBufferDescriptorOwner = true
const usbBufferOwnedByProcessor usbBufferDescriptorOwner = false

func (bdt *usbBufferDescriptorTable_t) Address() uintptr {
	addr := uintptr(unsafe.Pointer(&bdt.records[0]))
	if addr&0x1FF != 0 {
		panic("USB Buffer Descriptor Table is not 512-byte aligned")
	}

	return addr
}

func (bdt *usbBufferDescriptorTable_t) Entry(i int) usbBufferDescriptorEntry {
	return usbBufferDescriptorEntry{&bdt.records[i], &bdt.retain[i]}
}

func (bdt *usbBufferDescriptorTable_t) Get(ep uint16, tx, odd bool) usbBufferDescriptorEntry {
	i := ep << 2
	if tx {
		i |= 2
	}
	if odd {
		i |= 1
	}
	return bdt.Entry(int(i))
}

func (bdt *usbBufferDescriptorTable_t) GetStat(stat uint8) usbBufferDescriptorEntry {
	return bdt.Entry(int(stat >> 2))
}

func (bde usbBufferDescriptorEntry) Reset() {
	bde.record.descriptor.Set(0)
	bde.record.address.Set(0)
	*bde.retain = nil
}

func (bde usbBufferDescriptorEntry) Describe(size uint32, data1 bool) {
	if size >= 1024 {
		panic("USB Buffer Descriptor overflow")
	}

	d := size<<16 | usb_BDT_Own | usb_BDT_DTS
	if data1 {
		d |= usb_BDT_Data1
	}
	bde.record.descriptor.Set(d)
}

func (bde usbBufferDescriptorEntry) Set(data []byte, data1 bool) {
	if len(data) == 0 {
		// send an empty response
		*bde.retain = nil
		bde.record.address.Set(0)
		bde.Describe(0, data1)
		return
	}

	if len(data) >= 1024 {
		panic("USB Buffer Descriptor overflow")
	}

	slice := (*reflect.SliceHeader)(unsafe.Pointer(&data))

	*bde.retain = data
	bde.record.address.Set(uint32(slice.Data))
	bde.Describe(uint32(slice.Len), data1)
}

func (bde usbBufferDescriptorEntry) Data() []byte {
	length := bde.record.descriptor.Get() >> 16

	var data []byte
	slice := (*reflect.SliceHeader)(unsafe.Pointer(&data))
	slice.Len = uintptr(length)
	slice.Cap = uintptr(length)
	slice.Data = uintptr(bde.record.address.Get())
	return data
}

func (bde usbBufferDescriptorEntry) TokenPID() uint32 {
	return (bde.record.descriptor.Get() & usb_BDT_TOK_PID_Mask) >> usb_BDT_TOK_PID_Pos
}

func (bde usbBufferDescriptorEntry) Owner() usbBufferDescriptorOwner {
	if bde.record.descriptor.HasBits(usb_BDT_Own) {
		return usbBufferOwnedByUSBFS
	}
	return usbBufferOwnedByProcessor
}
