package usb

// Implementation of 32-bit target-agnostic USB device controller driver (dcd).

import "unsafe"

func init() {
	if unsafe.Sizeof(uintptr(0)) > 4 {
		panic("USB device controller is only supported on 32-bit systems")
	}
}

// dcdCount defines the number of USB cores to configure for device mode. It is
// computed as the sum of all declared device configuration descriptors.
const dcdCount = descCDCACMCount // + ...

// dcdInstance provides statically-allocated instances of each USB device
// controller configured on this platform.
var dcdInstance [dcdCount]dcd

// dhwInstance provides statically-allocated instances of each USB hardware
// abstraction for ports configured as device on this platform.
var dhwInstance [dcdCount]dhw

// dcd implements a generic USB device controller driver (dcd) for 32-bit ARM
// targets.
type dcd struct {
	*dhw // USB hardware abstraction layer

	core *core // Parent USB core this instance is attached to
	port int   // USB port index
	cc   class // USB device class
	id   int   // USB device controller index
}

// initDCD initializes and assigns a free device controller instance to the
// given USB port. Returns the initialized device controller or nil if no free
// device controller instances remain.
func initDCD(port int, class class) (*dcd, status) {
	if 0 == dcdCount {
		return nil, statusInvalid // Must have defined device controllers
	}
	switch class.id {
	case classDeviceCDCACM:
		if 0 == class.config || class.config > descCDCACMCount {
			return nil, statusInvalid // Must have defined descriptors
		}
	default:
	}
	// Return the first instance whose assigned core is currently nil.
	for i := range dcdInstance {
		if nil == dcdInstance[i].core {
			// Initialize device controller.
			dcdInstance[i].dhw = allocDHW(port, i, &dcdInstance[i])
			dcdInstance[i].core = &coreInstance[port]
			dcdInstance[i].port = port
			dcdInstance[i].cc = class
			dcdInstance[i].id = i
			return &dcdInstance[i], statusOK
		}
	}
	return nil, statusBusy // No free device controller instances available.
}

// class returns the receiver's current device class configuration.
func (d *dcd) class() class { return d.cc }

// dcdSetupSize defines the size (bytes) of a USB standard setup packet.
const dcdSetupSize = 8 // bytes

// dcdSetup contains the USB standard setup packet used to configure a device.
type dcdSetup struct {
	bmRequestType uint8
	bRequest      uint8
	wValue        uint16
	wIndex        uint16
	wLength       uint16
}

// pack returns the receiver setup packet encoded as uint64.
func (s dcdSetup) pack() uint64 {
	return ((uint64(s.bmRequestType) & 0xFF) << 0) |
		((uint64(s.bRequest) & 0xFF) << 8) |
		((uint64(s.wValue) & 0xFFFF) << 16) |
		((uint64(s.wIndex) & 0xFFFF) << 32) |
		((uint64(s.wLength) & 0xFFFF) << 48)
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
	dcdEventInvalid          uint8 = iota // Invalid interrupt
	dcdEventStatusReset                   // USB reset received
	dcdEventStatusRun                     // USB controller entered run state
	dcdEventStatusSuspend                 // USB suspend received
	dcdEventStatusError                   // USB error condition detected on bus
	dcdEventControlSetup                  // USB setup received
	dcdEventPeripheralReady               // USB PHY powered and ready to _go_
	dcdEventTransactComplete              // USB transaction complete
	dcdEventTimer                         // USB (system) timer
)

func (d *dcd) event(ev dcdEvent) {

	switch ev.id {
	case dcdEventInvalid:
	case dcdEventStatusReset:
		d.endpointMask = 0

	case dcdEventPeripheralReady:
		// Configure and enable control endpoint 0
		d.endpointEnable(0, true, 0)

	case dcdEventStatusRun:
	case dcdEventStatusSuspend:
	case dcdEventStatusError:
	case dcdEventControlSetup:
		// On control endpoint 0 setup events, the ev.setup field will be defined
		switch d.controlSetup(ev.setup) {
		case dcdStageSetup:
		case dcdStageData:
		case dcdStageStatus:
			d.controlStatus()
		case dcdStageStall:
			d.controlStall()
		}

	case dcdEventTransactComplete:
	case dcdEventTimer:
	default:
	}
}

// dcdStage represents the USB transaction stage of a control request.
type dcdStage uint8

// Enumerated constants for all possible USB transaction stages.
const (
	dcdStageSetup  dcdStage = iota // Indicates no stage transition required
	dcdStageData                   // IN/OUT data transfer
	dcdStageStatus                 // Setup request complete
	dcdStageStall                  // Unhandled or invalid request
)

// controlSetup handles setup messages on control endpoint 0.
func (d *dcd) controlSetup(sup dcdSetup) dcdStage {

	// Reset endpoint 0 notify mask
	d.controlMask = 0

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
				d.controlDeviceAddress(sup.wValue)
				d.controlReceive(uintptr(0), 0, false)
				return dcdStageSetup

			// SET CONFIGURATION (0x09):
			case descRequestStandardSetConfiguration:
				d.cc.config = int(sup.wValue)
				if 0 == d.cc.config || d.cc.config > dcdCount {
					// Use default if invalid index received
					d.cc.config = 1
				}

				// Respond based on our device class configuration
				switch d.cc.id {

				// CDC-ACM (single)
				case classDeviceCDCACM:
					d.uartConfigure()
					d.controlReceive(uintptr(0), 0, false)

				default:
					// Unhandled device class
				}
				return dcdStageSetup

			default:
				// Unhandled request
			}

		// --- DEVICE Tx (IN) ---
		case descRequestTypeRecipientDevice | descRequestTypeDirIn:

			// Identify which request was received
			switch sup.bRequest {

			// GET STATUS (0x00):
			case descRequestStandardGetStatus:
				d.controlReply[0] = 0
				d.controlReply[1] = 0
				d.controlTransmit(
					uintptr(unsafe.Pointer(&d.controlReply[0])), 2, false)
				return dcdStageSetup

			// GET DESCRIPTOR (0x06):
			case descRequestStandardGetDescriptor:
				d.controlDescriptor(sup)
				return dcdStageSetup

			// GET CONFIGURATION (0x08):
			case descRequestStandardGetConfiguration:
				d.controlReply[0] = uint8(d.cc.config)
				d.controlTransmit(
					uintptr(unsafe.Pointer(&d.controlReply[0])), 1, false)
				return dcdStageSetup

			default:
				// Unhandled request
			}

		// --- INTERFACE Tx (IN) ---
		case descRequestTypeRecipientInterface | descRequestTypeDirIn:

			// Identify which request was received
			switch sup.bRequest {

			// GET DESCRIPTOR (0x06):
			case descRequestStandardGetDescriptor:
				d.controlDescriptor(sup)
				return dcdStageSetup

			default:
				// Unhandled request
			}

		// --- ENDPOINT Rx (OUT) ---
		case descRequestTypeRecipientEndpoint | descRequestTypeDirOut:

			// Identify which request was received
			switch sup.bRequest {

			// CLEAR FEATURE (0x01):
			case descRequestStandardClearFeature:
				// TODO

			// SET FEATURE (0x03):
			case descRequestStandardSetFeature:
				// TODO

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
				d.controlReply[0] = uint8(status)
				d.controlReply[1] = uint8(status >> 8)
				d.controlTransmit(
					uintptr(unsafe.Pointer(&d.controlReply[0])), 2, false)
				return dcdStageSetup

			default:
				// Unhandled request
			}

		default:
			// Unhandled request recepient or direction
		}

	// === CLASS REQUEST ===
	case descRequestTypeTypeClass:

		// Switch on the recepient and direction of the request
		switch sup.bmRequestType &
			(descRequestTypeRecipientMsk | descRequestTypeDirMsk) {

		// --- INTERFACE Rx (OUT) ---
		case descRequestTypeRecipientInterface | descRequestTypeDirOut:

			// Identify which request was received
			switch sup.bRequest {

			// CDC | SET LINE CODING (0x20):
			case descCDCRequestSetLineCoding:

				// Respond based on our device class configuration
				switch d.cc.id {

				// CDC-ACM (single)
				case classDeviceCDCACM:
					// line coding must contain exactly 7 bytes
					if descCDCACMCodingSize == sup.wLength {
						d.setup = sup
						d.controlReceive(
							uintptr(unsafe.Pointer(&descCDCACM[d.cc.config-1].cx[0])),
							descCDCACMCodingSize, true)
						return dcdStageSetup
					}

				default:
					// Unhandled device class
				}

			// CDC | SET CONTROL LINE STATE (0x22):
			case descCDCRequestSetControlLineState:

				// Respond based on our device class configuration
				switch d.cc.id {

				// CDC-ACM (single)
				case classDeviceCDCACM:

					// Determine interface destination of the notification
					switch sup.wIndex {

					// Control/status interface:
					case descCDCACMInterfaceCtrl:
						d.controlReceive(uintptr(0), 0, false)
						return dcdStageSetup

					default:
						// Unhandled device interface
					}

				default:
					// Unhandled device class
				}

			// CDC | SEND BREAK (0x23):
			case descCDCRequestSendBreak:

				// Respond based on our device class configuration
				switch d.cc.id {

				// CDC-ACM (single)
				case classDeviceCDCACM:
					d.controlReceive(uintptr(0), 0, false)
					return dcdStageSetup

				default:
					// Unhandled device class
				}

			default:
				// Unhandled request
			}

		default:
			// Unhandled request recepient or direction
		}

	case descRequestTypeTypeVendor:
	default:
		// Unhandled request type
	}

	// All successful requests return early. If we reach this point, the request
	// was invalid or unhandled. Stall the endpoint.
	return dcdStageStall
}

// controlComplete handles the setup completion of control endpoint 0.
func (d *dcd) controlComplete(status uint32) {

	// Reset endpoint 0 notify mask
	d.controlMask = 0

	// First, switch on the type of request (standard, class, or vendor)
	switch d.setup.bmRequestType & descRequestTypeTypeMsk {

	// === CLASS REQUEST ===
	case descRequestTypeTypeClass:

		// Switch on the recepient and direction of the request
		switch d.setup.bmRequestType &
			(descRequestTypeRecipientMsk | descRequestTypeDirMsk) {

		// --- INTERFACE Rx (OUT) ---
		case descRequestTypeRecipientInterface | descRequestTypeDirOut:

			// Identify which request was received
			switch d.setup.bRequest {

			// CDC | SET LINE CODING (0x20):
			case descCDCRequestSetLineCoding:

				// Respond based on our device class configuration
				switch d.cc.id {

				// CDC-ACM (single)
				case classDeviceCDCACM:
					acm := &descCDCACM[d.cc.config-1]

					// Determine interface destination of the notification
					switch d.setup.wIndex {

					// Control/status interface:
					case descCDCACMInterfaceCtrl:

						// Notify PHY to handle triggers like special baud rates, which
						// signal to reboot into bootloader or begin receiving OTA updates
						d.controlLineCoding(descCDCACMLineCoding{
							baud:     packU32(acm.cx[:]),
							stopBits: acm.cx[4],
							parity:   acm.cx[5],
							numBits:  acm.cx[6],
						})

					default:
						// Unhandled device interface
					}

				default:
					// Unhandled device class
				}

			default:
				// Unhandled request
			}

		default:
			// Unhandled recepient or direction
		}

	default:
		// Unhandled request type
	}
}

func (d *dcd) controlDescriptor(sup dcdSetup) {

	// Respond based on our device class configuration
	switch d.cc.id {

	// CDC-ACM (single)
	case classDeviceCDCACM:
		acm := &descCDCACM[d.cc.config-1]
		dxn := uint8(0)

		// Determine the type of descriptor being requested
		switch sup.wValue >> 8 {

		// Device descriptor
		case descTypeDevice:
			dxn = descLengthDevice
			_ = copy(acm.dx[:], acm.device[:dxn])

		// Configuration descriptor
		case descTypeConfigure:
			dxn = uint8(descCDCACMConfigSize)
			_ = copy(acm.dx[:], acm.config[:dxn])

		// String descriptor
		case descTypeString:
			if 0 == len(acm.locale) {
				break // No string descriptors defined!
			}
			var sd []uint8
			if 0 == uint8(sup.wValue) {

				// setup.wIndex contains an arbitrary index referring to a collection of
				// strings in some given language. This case (setup.wValue = [0x03]00)
				// is a string request from the host to determine what that language is.
				//
				// In subsequent string requests, the host will populate setup.wIndex
				// with the language code we return here in this string descriptor.
				//
				// This way all strings returned to the host are in the same language,
				// whatever language that may be.
				code := int(sup.wIndex)
				if code >= len(acm.locale) {
					code = 0
				}
				sd = acm.locale[code].descriptor[sup.wValue&0xFF][:]

			} else {

				// setup.wIndex now contains a language code, which we specified in a
				// previous request (above: setup.wValue = [0x03]00). We need to locate
				// the set of strings whose language matches the language code given in
				// this new setup.wIndex.
				for code := range acm.locale {
					if sup.wIndex == acm.locale[code].language {
						// Found language, check if string descriptor at given index exists
						if int(sup.wValue&0xFF) < len(acm.locale[code].descriptor) {

							// Found language with a string defined at the requested index.
							//
							// TODO: Add API methods to device controller that allows the user
							//       to provide these strings at/before driver initialization.
							//
							// For now, we just always use the descCommon* strings.
							var s string
							switch uint8(sup.wValue) {
							case 1:
								s = descCommonManufacturer
							case 2:
								s = descCommonProduct
							case 3:
								s = descCommonSerialNumber
							}

							// Construct a string descriptor dynamically to be transmitted on
							// the serial bus.
							sd = acm.locale[code].descriptor[int(sup.wValue&0xFF)][:]
							// String descriptor format is 2-byte header + 2-bytes per rune
							sd[0] = uint8(2 + 2*len(s)) // header[0] = descriptor length
							sd[1] = descTypeString      // header[1] = descriptor type
							// Copy UTF-8 string into string descriptor as UTF-16
							for n, c := range s {
								if 2+2*n >= len(sd) {
									break
								}
								sd[2+2*n] = uint8(c)
								sd[3+2*n] = 0
							}
							break // end search for matching language code
						}
					}
				}
			}
			// Copy string descriptor into descriptor transmit buffer
			if nil != sd && len(sd) >= 0 {
				dxn = sd[0]
				_ = copy(acm.dx[:], sd[:dxn])
			}

		// Device qualification descriptor
		case descTypeQualification:
			dxn = descLengthQualification
			_ = copy(acm.dx[:], acm.qualif[:dxn])

		// Alternate configuration descriptor
		case descTypeOtherSpeedConfiguration:
			// TODO

		default:
			// Unhandled descriptor type
		}

		if dxn > 0 {
			if dxn > uint8(sup.wLength) {
				dxn = uint8(sup.wLength)
			}
			flushCache(
				uintptr(unsafe.Pointer(&acm.dx[0])), uintptr(dxn))
			d.controlTransmit(
				uintptr(unsafe.Pointer(&acm.dx[0])), uint32(dxn), false)
		}

	default:
		// Unhandled device class
	}

}
