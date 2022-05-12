//go:build usb.cdc
// +build usb.cdc

package usb

import "unsafe"

//go:inline
func (d *dcd) endpointMaxPacketSize(endpoint uint8) uint32 {
	switch endpointNumber(endpoint) {
	case descCDCEndpointCtrl:
		return descControlPacketSize
	case descCDCEndpointStatus:
		return descCDCStatusPacketSize
	case descCDCEndpointDataRx:
		return descCDCDataRxPacketSize
	case descCDCEndpointDataTx:
		return descCDCDataTxPacketSize
	}
	return descControlPacketSize
}

//go:inline
func (d *dcd) controlEndpoint() uint8 {
	return descCDCEndpointCtrl
}

func (d *dcd) controlSetConfiguration() {
	d.cdcConfigure()
}

func (d *dcd) controlClassRequest(sup dcdSetup) dcdStage {

	// Switch on the recepient and direction of the request
	switch sup.bmRequestType &
		(descRequestTypeRecipientMsk | descRequestTypeDirMsk) {

	// --- INTERFACE Rx (OUT) ---
	case descRequestTypeRecipientInterface | descRequestTypeDirOut:

		// Identify which request was received
		switch sup.bRequest {

		// CDC | SET LINE CODING (0x20):
		case descCDCRequestSetLineCoding:
			// line coding must contain exactly 7 bytes
			if uint16(descCDCLineCodingSize) == sup.wLength {
				d.controlReceive(
					uintptr(unsafe.Pointer(&descCDC[d.cc.config-1].cx[0])),
					uint32(descCDCLineCodingSize), true)
				// CDC Line Coding packet receipt handling occurs in method
				// controlComplete().
				return dcdStageDataOut
			}

		// CDC | SET CONTROL LINE STATE (0x22):
		case descCDCRequestSetControlLineState:
			// Determine interface destination of the request
			switch sup.wIndex {
			// Control/status interface:
			case descCDCInterfaceCtrl:
				// CDC Control Line State packet receipt handling occurs in method
				// controlComplete().
				d.controlReceive(uintptr(0), 0, false)
				return dcdStageStatusOut

			default:
				// Unhandled device interface
			}

		// CDC | SEND BREAK (0x23):
		case descCDCRequestSendBreak:
			d.controlReceive(uintptr(0), 0, false)
			return dcdStageStatusOut

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
	}
	return false
}

func (d *dcd) controlGetDescriptor(sup dcdSetup) {

	acm := &descCDC[d.cc.config-1]
	dxn := uint8(0)

	// Determine the type of descriptor being requested
	switch sup.wValue >> 8 {

	// Device descriptor
	case descTypeDevice:
		dxn = descLengthDevice
		_ = copy(acm.dx[:], acm.device[:dxn])

	// Configuration descriptor
	case descTypeConfigure:
		dxn = uint8(descCDCConfigSize)
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

			// CDC | SET LINE CODING (0x20):
			case descCDCRequestSetLineCoding:

				acm := &descCDC[d.cc.config-1]

				// Determine interface destination of the request
				switch d.setup.wIndex {

				// CDC-ACM Control Interface:
				case descCDCInterfaceCtrl:
					// Notify PHY to handle triggers like special baud rates, which
					// signal to reboot into bootloader or begin receiving OTA updates
					d.cdcSetLineCoding(acm.cx[:])

				default:
					// Unhandled device interface
				}

				// CDC | SET CONTROL LINE STATE (0x22):
			case descCDCRequestSetControlLineState:

				// Determine interface destination of the request
				switch d.setup.wIndex {

				// Control/status interface:
				case descCDCInterfaceCtrl:
					// DTR is bit 0 (mask 0x01), RTS is bit 1 (mask 0x02)
					d.cdcSetLineState(d.setup.wValue)

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
