package usb

// Implementation of target-agnostic USB device controller driver (dcd).
//
// The types, constants, and methods defined in this unit are applicable to all
// targets. It was designed to complement the device hardware abstraction (dhw)
// implemented for each target, providing common/shared functionality and
// defining a standard interface with which the dhw must adhere.

import (
	"runtime/volatile"
	"unsafe"
)

// dcdCount defines the number of USB cores to configure for device mode. It is
// computed as the sum of all declared device configuration descriptors.
const dcdCount = descCDCCount + descHIDCount

// dcdInstance provides statically-allocated instances of each USB device
// controller configured on this platform.
var dcdInstance [dcdCount]dcd

// dhwInstance provides statically-allocated instances of each USB hardware
// abstraction for ports configured as device on this platform.
var dhwInstance [dcdCount]dhw

// dcd implements a generic USB device controller driver (dcd) for all targets.
type dcd struct {
	*dhw // USB hardware abstraction layer

	core *core // Parent USB core this instance is attached to
	port int   // USB port index
	cc   class // USB device class
	id   int   // USB device controller index

	st volatile.Register8 // USB device state
}

// initDCD initializes and assigns a free device controller instance to the
// given USB port. Returns the initialized device controller or nil if no free
// device controller instances remain.
func initDCD(port int, speed Speed, class class) (*dcd, status) {
	if 0 == dcdCount {
		return nil, statusInvalid // Must have defined device controllers
	}
	switch class.id {
	case classDeviceCDC:
		if 0 == class.config || class.config > descCDCCount {
			return nil, statusInvalid // Must have defined descriptors
		}
	default:
	}
	// Return the first instance whose assigned core is currently nil.
	for i := range dcdInstance {
		if nil == dcdInstance[i].core {
			// Initialize device controller.
			dcdInstance[i].dhw = allocDHW(port, i, speed, &dcdInstance[i])
			dcdInstance[i].core = &coreInstance[port]
			dcdInstance[i].port = port
			dcdInstance[i].cc = class
			dcdInstance[i].id = i
			dcdInstance[i].setState(dcdStateNotReady)
			return &dcdInstance[i], statusOK
		}
	}
	return nil, statusBusy // No free device controller instances available.
}

// class returns the receiver's current device class configuration.
func (d *dcd) class() class { return d.cc }

// dcdSetupSize defines the size (bytes) of a USB standard setup packet.
const dcdSetupSize = unsafe.Sizeof(dcdSetup{}) // 8 bytes

// dcdSetup contains the USB standard setup packet used to configure a device.
type dcdSetup struct {
	bmRequestType uint8
	bRequest      uint8
	wValue        uint16
	wIndex        uint16
	wLength       uint16
}

// setupFrom decodes and returns a USB standard setup packet located at the
// memory address pointed to by addr.
func setupFrom(addr uintptr) (s dcdSetup) {
	var u uint64
	for i := uintptr(0); i < dcdSetupSize; i++ {
		u |= uint64(*(*uint8)(unsafe.Pointer(addr + i))) << (i << 3)
	}
	s.set(u)
	return
}

// setup decodes and returns a USB standard setup packet stored in the given
// byte slice b.
func setup(b []uint8) dcdSetup {
	if len(b) >= int(dcdSetupSize) {
		return dcdSetup{
			bmRequestType: b[0],
			bRequest:      b[1],
			wValue:        packU16(b[2:]),
			wIndex:        packU16(b[4:]),
			wLength:       packU16(b[6:]),
		}
	}
	return dcdSetup{}
}

//go:inline
func (s *dcdSetup) set(u uint64) {
	s.bmRequestType = uint8(u & 0xFF)
	s.bRequest = uint8((u & 0xFF00) >> 8)
	s.wValue = uint16((u & 0xFFFF0000) >> 16)
	s.wIndex = uint16((u & 0xFFFF00000000) >> 32)
	s.wLength = uint16((u & 0xFFFF000000000000) >> 48)
}

// pack returns the receiver USB standard setup packet s encoded as uint64.
//go:inline
func (s dcdSetup) pack() uint64 {
	return ((uint64(s.bmRequestType) & 0xFF) << 0) |
		((uint64(s.bRequest) & 0xFF) << 8) |
		((uint64(s.wValue) & 0xFFFF) << 16) |
		((uint64(s.wIndex) & 0xFFFF) << 32) |
		((uint64(s.wLength) & 0xFFFF) << 48)
}

// direction parses the direction bit from the bmRequestType field of a SETUP
// packet, returning 0 for OUT (Rx) and 1 for IN (Tx) requests.
//go:inline
func (s dcdSetup) direction() uint8 {
	return (s.bmRequestType & descRequestTypeDirMsk) >> descRequestTypeDirPos
}

//go:inline
func (s dcdSetup) equals(t dcdSetup) bool {
	return s.bmRequestType == t.bmRequestType && s.bRequest == t.bRequest &&
		s.wValue == t.wValue && s.wIndex == t.wIndex && s.wLength == t.wLength
}

// dcdState defines the current state of the device class driver.
type dcdState uint8

const (
	dcdStateNotReady   dcdState = iota // initial state, before END_OF_RESET
	dcdStateDefault                    // after END_OF_RESET, before SET_ADDRESS
	dcdStateAddressed                  // after SET_ADDRESS, before SET_CONFIGURATION
	dcdStateConfigured                 // after SET_CONFIGURATION, operational state
	dcdStateSuspended                  // while operational, after SUSPEND
)

func (d *dcd) state() dcdState { return dcdState(d.st.Get()) }

func (d *dcd) setState(state dcdState) (ok bool) {
	curr := d.state()
	switch state {
	case dcdStateNotReady:
		ok = true
	case dcdStateDefault:
		ok = curr == dcdStateNotReady || curr == dcdStateDefault
	case dcdStateAddressed:
		ok = curr == dcdStateDefault
	case dcdStateConfigured:
		ok = curr == dcdStateAddressed || curr == dcdStateConfigured || curr == dcdStateSuspended
	case dcdStateSuspended:
		ok = curr == dcdStateAddressed || curr == dcdStateConfigured || curr == dcdStateSuspended
	default:
		ok = false
	}
	if ok {
		d.st.Set(uint8(state))
	}
	return
}

// dcdEvent is used to describe virtual interrupts on the USB bus to a device
// controller.
//
// Since the device controller software is intended for use with multiple TinyGo
// targets, all of which may not have exactly the same USB bus interrupts, a
// "virtual interrupt" is defined that is common to all targets. The target's
// hardware implementation (type dhw) is responsible for translating real system
// interrupts it receives into the appropriate virtual interrupt code, defined
// below, and notifying the device controller via method (*dcd).event(dcdEvent).
type dcdEvent struct {
	id    uint8
	setup dcdSetup
	mask  uint32
}

// Enumerated constants for all possible USB device controller interrupt codes.
const (
	dcdEventInvalid             uint8 = iota // Invalid interrupt
	dcdEventStatusReset                      // USB RESET received
	dcdEventStatusResume                     // USB RESUME condition
	dcdEventStatusSuspend                    // USB SUSPEND received
	dcdEventStatusError                      // USB error condition detected on bus
	dcdEventDeviceReady                      // USB PHY powered and ready to _go_
	dcdEventDeviceAddress                    // USB device SET_ADDRESS complete
	dcdEventDeviceConfiguration              // USB device SET_CONFIGURATION complete
	dcdEventControlSetup                     // USB SETUP received
	dcdEventControlComplete                  // USB control request complete
	dcdEventTransferComplete                 // USB data transfer complete
	dcdEventTimer                            // USB (system) timer
)

func (d *dcd) event(ev dcdEvent) {

	switch ev.id {

	case dcdEventStatusReset:
		d.setState(dcdStateNotReady)

	case dcdEventStatusResume:
		d.setState(dcdStateConfigured)

	case dcdEventStatusSuspend:
		d.setState(dcdStateSuspended)

	case dcdEventDeviceReady:
		if d.setState(dcdStateDefault) {
			// Configure and enable control endpoint 0
			d.endpointEnable(0, true, 0)
		}

	case dcdEventDeviceAddress:
		// -- ** IMPORTANT ** --
		// dcdEventDeviceAddress must be triggered by the target driver, because
		// different MCUs require setting the device address at different times
		// during the enumeration process.
		d.setState(dcdStateAddressed)

	case dcdEventDeviceConfiguration:
		d.setState(dcdStateConfigured)

	case dcdEventControlSetup:
		// On control endpoint 0 setup events, the ev.setup field will be defined.
		// We overwrite the receiver's setup field, leaving it unmodified throughout
		// all transactions of a control transfer. It is only cleared once the
		// completion event dcdEventControlComplete has been called and finished
		// processing, or if its initial processing fails due to error.
		d.setup = ev.setup
		d.stage = d.controlSetup(ev.setup)
		switch d.stage {
		case dcdStageDataIn, dcdStageDataOut:
			// TBD: control endpoint data transfer

		case dcdStageStatusIn, dcdStageStatusOut:
			// TBD: control endpoint status transfer

		case dcdStageStall:
			d.controlStall(true, ev.setup.direction())

		case dcdStageSetup:
			fallthrough
		default:
			// TBD: no stage transition occurred
		}

	case dcdEventControlComplete:
		d.controlComplete()
		// clear the active SETUP packet once the control transfer completes.
		d.setup = dcdSetup{}

	case dcdEventTransferComplete:
		// TBD: data endpoint transfer complete

	case dcdEventInvalid, dcdEventStatusError, dcdEventTimer:
		fallthrough
	default:
		// TBD: unhandled events
	}
}

// dcdStage represents the stage of a USB control transfer.
type dcdStage uint8

// Enumerated constants for all possible USB control transfer stages.
const (
	dcdStageSetup     dcdStage = iota // Indicates no stage transition required
	dcdStageDataIn                    // IN data transfer
	dcdStageDataOut                   // OUT data transfer
	dcdStageStatusIn                  // IN status request
	dcdStageStatusOut                 // OUT status request
	dcdStageStall                     // Unhandled or invalid request
)

// controlSetup handles setup messages on control endpoint 0.
func (d *dcd) controlSetup(sup dcdSetup) dcdStage {

	// First, switch on the type of request (standard, class, or vendor)
	switch sup.bmRequestType & descRequestTypeTypeMsk {

	// === STANDARD REQUEST ===
	case descRequestTypeTypeStandard:

		// Switch on the recepient and direction of the request
		switch sup.bmRequestType &
			(descRequestTypeRecipientMsk | descRequestTypeDirMsk) {

		// --- DEVICE Rx (OUT) ---
		case descRequestTypeRecipientDevice | descRequestTypeDirOut:

			// Identify which request was received
			switch sup.bRequest {

			// SET ADDRESS (0x05):
			case descRequestStandardSetAddress:
				d.setDeviceAddress(sup.wValue)
				d.controlReceive(uintptr(0), 0, false)
				return dcdStageStatusOut

			// SET CONFIGURATION (0x09):
			case descRequestStandardSetConfiguration:
				d.cc.config = int(sup.wValue)
				if 0 == d.cc.config || d.cc.config > dcdCount {
					// Use default if invalid index received
					d.cc.config = 1
				}
				d.event(dcdEvent{id: dcdEventDeviceConfiguration})
				d.controlSetConfiguration()
				d.controlReceive(uintptr(0), 0, false)
				return dcdStageStatusOut

			default:
				// Unhandled request
			}

		// --- DEVICE Tx (IN) ---
		case descRequestTypeRecipientDevice | descRequestTypeDirIn:

			// Identify which request was received
			switch sup.bRequest {

			// GET STATUS (0x00):
			case descRequestStandardGetStatus:
				d.controlTransmit(
					d.controlStatusBuffer([]uint8{0, 0}),
					2, false)
				return dcdStageDataIn

			// GET DESCRIPTOR (0x06):
			case descRequestStandardGetDescriptor:
				d.controlGetDescriptor(sup)
				return dcdStageDataIn

			// GET CONFIGURATION (0x08):
			case descRequestStandardGetConfiguration:
				d.controlTransmit(
					d.controlStatusBuffer([]uint8{
						uint8(d.cc.config),
					}),
					1, false)
				return dcdStageDataIn

			default:
				// Unhandled request
			}

		// --- INTERFACE Tx (IN) ---
		case descRequestTypeRecipientInterface | descRequestTypeDirIn:
			if d.controlGetInterfaceDescriptor(sup) {
				return dcdStageDataIn
			}

		// --- ENDPOINT Rx (OUT) ---
		case descRequestTypeRecipientEndpoint | descRequestTypeDirOut:

			// Identify which request was received
			switch sup.bRequest {

			// CLEAR FEATURE (0x01):
			case descRequestStandardClearFeature:
				d.endpointClearFeature(uint8(sup.wIndex))
				d.controlReceive(uintptr(0), 0, false)
				return dcdStageStatusOut

			// SET FEATURE (0x03):
			case descRequestStandardSetFeature:
				d.endpointSetFeature(uint8(sup.wIndex))
				d.controlReceive(uintptr(0), 0, false)
				return dcdStageStatusOut

			default:
				// Unhandled request
			}

		// --- ENDPOINT Tx (IN) ---
		case descRequestTypeRecipientEndpoint | descRequestTypeDirIn:

			// Identify which request was received
			switch sup.bRequest {

			// GET STATUS (0x00):
			case descRequestStandardGetStatus:
				status := d.endpointStatus(uint8(sup.wIndex))
				d.controlTransmit(
					d.controlStatusBuffer([]uint8{
						uint8(status),
						uint8(status >> 8),
					}),
					2, false)
				return dcdStageDataIn

			default:
				// Unhandled request
			}

		default:
			// Unhandled request recepient or direction
		}

	// === CLASS REQUEST ===
	case descRequestTypeTypeClass:

		// Forward all class requests to the device class implementation.
		return d.controlClassRequest(sup)

	case descRequestTypeTypeVendor:
	default:
		// Unhandled request type
	}

	// All successful requests return early. If we reach this point, the request
	// was invalid or unhandled. Stall the endpoint.
	return dcdStageStall
}
