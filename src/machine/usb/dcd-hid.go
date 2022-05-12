//go:build baremetal && usb.hid
// +build baremetal,usb.hid

package usb

import "unsafe"

//go:inline
func (d *dcd) endpointMaxPacketSize(endpoint uint8) uint32 {
	switch endpointNumber(endpoint) {
	case descHIDEndpointCtrl:
		return descControlPacketSize
	case descHIDEndpointKeyboard:
		return descHIDKeyboardTxPacketSize
	case descHIDEndpointMouse:
		return descHIDMouseTxPacketSize
	case descHIDEndpointSerialRx: // == descHIDEndpointSerialTx
		switch endpoint {
		case rxEndpoint(endpoint):
			return descHIDSerialRxPacketSize
		case txEndpoint(endpoint):
			return descHIDSerialTxPacketSize
		}
	case descHIDEndpointJoystick:
		return descHIDJoystickTxPacketSize
	case descHIDEndpointMediaKey:
		return descHIDMediaKeyTxPacketSize
	}
	return descControlPacketSize
}

//go:inline
func (d *dcd) controlEndpoint() uint8 {
	return descHIDEndpointCtrl
}

//go:inline
func (d *dcd) controlSetConfiguration() {
	d.serialConfigure()
	d.keyboardConfigure()
	d.mouseConfigure()
	d.joystickConfigure()
}

func (d *dcd) controlClassRequest(sup dcdSetup) dcdStage {

	// Switch on the recepient and direction of the request
	switch sup.bmRequestType &
		(descRequestTypeRecipientMsk | descRequestTypeDirMsk) {

	// --- INTERFACE Rx (OUT) ---
	case descRequestTypeRecipientInterface | descRequestTypeDirOut:

		// Identify which request was received
		switch sup.bRequest {

		// HID | SET REPORT (0x09)
		case descHIDRequestSetReport:
			if sup.wLength <= descHIDSxSize {
				descHID[d.cc.config-1].cx[0] = 0xE9
				d.controlReceive(
					uintptr(unsafe.Pointer(&descHID[d.cc.config-1].cx[0])),
					uint32(sup.wLength), true)
				return dcdStageDataOut
			}

		// HID | SET IDLE (0x0A)
		case descHIDRequestSetIdle:
			idleRate := sup.wValue >> 8
			// TBD: do we need to handle this request? wIndex contains the target
			// interface of the request.
			_ = idleRate
			d.controlReceive(uintptr(0), 0, false)
			return dcdStageStatusOut

		default:
			// Unhandled request
		}

	// --- INTERFACE Tx (IN) ---
	case descRequestTypeRecipientInterface | descRequestTypeDirIn:

		// Identify which request was received
		switch sup.bRequest {

		// HID | GET REPORT (0x01)
		case descHIDRequestGetReport:

			reportType := uint8(sup.wValue >> 8)
			reportID := uint8(sup.wValue)
			// TBD: do we need to handle this request? wIndex contains the target
			// interface of the request.
			_, _ = reportType, reportID
			d.controlTransmit(
				d.controlStatusBuffer([]uint8{0, 0}),
				2, false)
			return dcdStageDataIn

		default:
			// Unhandled request
		}

	default:
		// Unhandled request recepient or direction
	}

	return dcdStageStall
}

func (d *dcd) controlGetInterfaceDescriptor(sup dcdSetup) bool {
	switch sup.bRequest {
	// GET DESCRIPTOR (0x06):
	case descRequestStandardGetDescriptor:
		d.controlGetDescriptor(sup)
		return true
	// GET HID REPORT (0x01):
	case descHIDRequestGetReport:
		d.controlGetDescriptor(sup)
		return true
	}
	return false
}

func (d *dcd) controlGetDescriptor(sup dcdSetup) {

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

// controlComplete handles the setup completion of control endpoint 0.
func (d *dcd) controlComplete() {

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

			// HID | SET REPORT (0x09)
			case descHIDRequestSetReport:

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
				// Unhandled request
			}

		default:
			// Unhandled recepient or direction
		}

	default:
		// Unhandled request type
	}
}
