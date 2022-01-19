package usb

// Implementation of target-agnostic USB device controller driver (dcd).
//
// The types, constants, and methods defined in this unit are applicable to all
// targets. It was designed to complement the device hardware abstraction (dhw)
// implemented for each target, providing common/shared functionality and
// defining a standard interface with which the dhw must adhere.

import "unsafe"

// dcdCount defines the number of USB cores to configure for device mode. It is
// computed as the sum of all declared device configuration descriptors.
const dcdCount = descCDCACMCount + descHIDCount

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
}

// initDCD initializes and assigns a free device controller instance to the
// given USB port. Returns the initialized device controller or nil if no free
// device controller instances remain.
func initDCD(port int, speed Speed, class class) (*dcd, status) {
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
			dcdInstance[i].dhw = allocDHW(port, i, speed, &dcdInstance[i])
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
func setupFrom(addr uintptr) dcdSetup {
	var u uint64
	for i := uintptr(0); i < 8; i++ {
		u |= uint64(*(*uint8)(unsafe.Pointer(addr + i))) << (i << 3)
	}
	return dcdSetup{
		bmRequestType: uint8(u & 0xFF),
		bRequest:      uint8((u & 0xFF00) >> 8),
		wValue:        uint16((u & 0xFFFF0000) >> 16),
		wIndex:        uint16((u & 0xFFFF00000000) >> 32),
		wLength:       uint16((u & 0xFFFF000000000000) >> 48),
	}
}

// setup decodes and returns a USB standard setup packet stored in the given
// byte slice b.
func setup(b []uint8) dcdSetup {
	if len(b) >= 8 {
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

// pack returns the receiver USB standard setup packet s encoded as uint64.
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
		d.stage = d.controlSetup(ev.setup)
		switch d.stage {
		case dcdStageSetup:
		case dcdStageDataIn:
		case dcdStageDataOut:
		case dcdStageStatusIn:
		case dcdStageStatusOut:
		case dcdStageStall:
			d.controlStall()
		}

	case dcdEventTransactComplete:
	case dcdEventTimer:
	default:
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
				d.setDeviceAddress(sup.wValue)
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

				// HID
				case classDeviceHID:
					d.serialConfigure()
					d.keyboardConfigure()
					d.mouseConfigure()
					d.joystickConfigure()
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

				// Respond based on our device class configuration
				switch d.cc.id {

				// CDC-ACM (single)
				case classDeviceCDCACM:
					d.controlDescriptorCDCACM(sup)
					return dcdStageSetup

				// HID
				case classDeviceHID:
					d.controlDescriptorHID(sup)
					return dcdStageSetup

				default:
					// Unhandled device class
				}

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

				// Respond based on our device class configuration
				switch d.cc.id {

				// CDC-ACM (single)
				case classDeviceCDCACM:
					d.controlDescriptorCDCACM(sup)
					return dcdStageSetup

				// HID
				case classDeviceHID:
					d.controlDescriptorHID(sup)
					return dcdStageSetup

				default:
					// Unhandled device class
				}

			// GET HID REPORT (0x01):
			case descHIDRequestGetReport:

				// Respond based on our device class configuration
				switch d.cc.id {

				// HID
				case classDeviceHID:
					d.controlDescriptorHID(sup)
					return dcdStageSetup

				default:
					// Unhandled device class
				}

			default:
				// Unhandled request
			}

		// --- ENDPOINT Rx (OUT) ---
		case descRequestTypeRecipientEndpoint | descRequestTypeDirOut:

			// Identify which request was received
			switch sup.bRequest {

			// CLEAR FEATURE (0x01):
			case descRequestStandardClearFeature:
				d.endpointClearFeature(uint8(sup.wIndex))
				d.controlReceive(uintptr(0), 0, false)
				return dcdStageSetup

			// SET FEATURE (0x03):
			case descRequestStandardSetFeature:
				d.endpointSetFeature(uint8(sup.wIndex))
				d.controlReceive(uintptr(0), 0, false)
				return dcdStageSetup

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
						// CDC Line Coding packet receipt handling occurs in method
						// controlComplete().
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

					// Determine interface destination of the request
					switch sup.wIndex {

					// Control/status interface:
					case descCDCACMInterfaceCtrl:
						// DTR is bit 0 (mask 0x01), RTS is bit 1 (mask 0x02)
						d.uartSetLineState(0 != sup.wValue&0x01, 0 != sup.wValue&0x02)
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

			// HID | SET REPORT (0x09)
			case descHIDRequestSetReport:

				// Respond based on our device class configuration
				switch d.cc.id {

				// HID
				case classDeviceHID:
					if sup.wLength <= descHIDCxSize {
						d.setup = sup
						descHID[d.cc.config-1].cx[0] = 0xE9
						d.controlReceive(
							uintptr(unsafe.Pointer(&descHID[d.cc.config-1].cx[0])),
							uint32(sup.wLength), true)
						return dcdStageSetup
					}

				default:
					// Unhandled device class
				}

			// HID | SET IDLE (0x0A)
			case descHIDRequestSetIdle:

				// Respond based on our device class configuration
				switch d.cc.id {

				// HID
				case classDeviceHID:
					idleRate := sup.wValue >> 8
					_ = idleRate
					d.controlReceive(uintptr(0), 0, false)
					return dcdStageSetup

				default:
					// Unhandled device class
				}

			default:
				// Unhandled request
			}

		// --- INTERFACE Tx (IN) ---
		case descRequestTypeRecipientInterface | descRequestTypeDirIn:

			// Identify which request was received
			switch sup.bRequest {

			// HID | GET REPORT (0x01)
			case descHIDRequestGetReport:

				// Respond based on our device class configuration
				switch d.cc.id {

				// HID
				case classDeviceHID:
					reportType := uint8(sup.wValue >> 8)
					reportID := uint8(sup.wValue)
					_, _ = reportType, reportID
					d.controlReply[0] = 0
					d.controlReply[1] = 0
					d.controlTransmit(
						uintptr(unsafe.Pointer(&d.controlReply[0])), 2, false)
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

					// Determine interface destination of the request
					switch d.setup.wIndex {

					// CDC-ACM Control Interface:
					case descCDCACMInterfaceCtrl:
						// Notify PHY to handle triggers like special baud rates, which
						// signal to reboot into bootloader or begin receiving OTA updates
						d.uartSetLineCoding(descCDCACMLineCoding{
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

			// HID | SET REPORT (0x09)
			case descHIDRequestSetReport:

				// Respond based on our device class configuration
				switch d.cc.id {

				// HID
				case classDeviceHID:
					hid := &descHID[d.cc.config-1]

					// Determine interface destination of the request
					switch d.setup.wIndex {

					// HID Keyboard Interface
					case descHIDInterfaceKeyboard:

						// Determine the type of descriptor being requested
						switch d.setup.wValue >> 8 {

						// Configuration descriptor
						case descTypeConfigure:
							if 1 == d.setup.wLength {
								hid.keyboard.led = hid.cx[0]
								d.controlTransmit(uintptr(0), 0, false)
							}

						default:
							// Unhandled descriptor type
						}

					// HID Serial Interface
					case descHIDInterfaceSerial:

						// Determine the type of descriptor being requested
						switch d.setup.wValue >> 8 {

						// String descriptor
						case descTypeString:
							if d.setup.wLength >= 4 && 0x68C245A9 == packU32(hid.cx[0:4]) {
								d.enableSOF(true, descHIDInterfaceCount)
							}

						default:
							// Unhandled descriptor type
						}

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

func (d *dcd) controlDescriptorCDCACM(sup dcdSetup) {

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
							s = descCommonProduct + " CDC-ACM"
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
}

func (d *dcd) controlDescriptorHID(sup dcdSetup) {

	hid := &descHID[d.cc.config-1]
	dxn := uint8(0)
	pos := uint8(0)

	// Determine the type of descriptor being requested
	switch sup.wValue >> 8 {

	// Device descriptor
	case descTypeDevice:
		dxn = descLengthDevice
		_ = copy(hid.dx[:], hid.device[:dxn])

	// Configuration descriptor
	case descTypeConfigure:
		dxn = uint8(descHIDConfigSize)
		_ = copy(hid.dx[:], hid.config[:dxn])

	// String descriptor
	case descTypeString:
		if 0 == len(hid.locale) {
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
			if code >= len(hid.locale) {
				code = 0
			}
			sd = hid.locale[code].descriptor[sup.wValue&0xFF][:]

		} else {

			// setup.wIndex now contains a language code, which we specified in a
			// previous request (above: setup.wValue = [0x03]00). We need to locate
			// the set of strings whose language matches the language code given in
			// this new setup.wIndex.
			for code := range hid.locale {
				if sup.wIndex == hid.locale[code].language {
					// Found language, check if string descriptor at given index exists
					if int(sup.wValue&0xFF) < len(hid.locale[code].descriptor) {

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
							s = descCommonProduct + " HID"
						case 3:
							s = descCommonSerialNumber
						}

						// Construct a string descriptor dynamically to be transmitted on
						// the serial bus.
						sd = hid.locale[code].descriptor[int(sup.wValue&0xFF)][:]
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
			_ = copy(hid.dx[:], sd[:dxn])
		}

	// Device qualification descriptor
	case descTypeQualification:
		dxn = descLengthQualification
		_ = copy(hid.dx[:], hid.qualif[:dxn])

	// Alternate configuration descriptor
	case descTypeOtherSpeedConfiguration:
		// TODO

	// HID descriptor
	case descTypeHID:

		// Determine interface destination of the request
		switch sup.wIndex {
		case descHIDInterfaceKeyboard:
			pos = descHIDConfigKeyboardPos

		case descHIDInterfaceMouse:
			pos = descHIDConfigMousePos

		case descHIDInterfaceSerial:
			pos = descHIDConfigSerialPos

		case descHIDInterfaceJoystick:
			pos = descHIDConfigJoystickPos

		case descHIDInterfaceMediaKey:
			pos = descHIDConfigMediaKeyPos

		default:
			// Unhandled HID interface
		}

		if 0 != pos {
			dxn = descLengthInterface
			_ = copy(hid.dx[:], hid.config[pos:pos+dxn])
		}

	// HID report descriptor
	case descTypeHIDReport:

		// Determine interface destination of the request
		switch sup.wIndex {
		case descHIDInterfaceKeyboard:
			dxn = uint8(len(descHIDReportKeyboard))
			_ = copy(hid.dx[:], descHIDReportKeyboard[:])

		case descHIDInterfaceMouse:
			dxn = uint8(len(descHIDReportMouse))
			_ = copy(hid.dx[:], descHIDReportMouse[:])

		case descHIDInterfaceSerial:
			dxn = uint8(len(descHIDReportSerial))
			_ = copy(hid.dx[:], descHIDReportSerial[:])

		case descHIDInterfaceJoystick:
			dxn = uint8(len(descHIDReportJoystick))
			_ = copy(hid.dx[:], descHIDReportJoystick[:])

		case descHIDInterfaceMediaKey:
			dxn = uint8(len(descHIDReportMediaKey))
			_ = copy(hid.dx[:], descHIDReportMediaKey[:])

		default:
			// Unhandled HID interface
		}

	default:
		// Unhandled descriptor type
	}

	if dxn > 0 {
		if dxn > uint8(sup.wLength) {
			dxn = uint8(sup.wLength)
		}
		flushCache(
			uintptr(unsafe.Pointer(&hid.dx[0])), uintptr(dxn))
		d.controlTransmit(
			uintptr(unsafe.Pointer(&hid.dx[0])), uint32(dxn), false)
	}
}
