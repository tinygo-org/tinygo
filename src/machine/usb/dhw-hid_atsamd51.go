//go:build usb.hid && (atsamd51 || atsame5x)
// +build usb.hid
// +build atsamd51 atsame5x

package usb

import "unsafe"

//go:inline
func (d *dhw) descriptorTable() uintptr {
	return uintptr(unsafe.Pointer(&descHID[d.cc.config-1].ed[0]))
}

// endpointDescriptor returns the endpoint descriptor for the given endpoint
// address, encoded as direction D and endpoint number N with the 8-bit mask
// D000NNNN.
//go:inline
func (d *dhw) endpointDescriptor(endpoint uint8) *dhwEPDesc {
	num, dir := unpackEndpoint(endpoint)
	return &descHID[d.cc.config-1].ed[num][dir]
}

//go:inline
func (d *dhw) controlSetupBuffer() uintptr {
	return uintptr(unsafe.Pointer(&descHID[d.cc.config-1].sx[0]))
}

//go:inline
func (d *dhw) controlStatusBuffer(data []uint8) uintptr {
	// reference to class configuration data
	c := descHID[d.cc.config-1]
	for i := range c.cx {
		c.cx[i] = 0 // zero out the control reply buffer
	}
	// copy the given data into control reply buffer
	copy(c.cx[:], data)
	return uintptr(unsafe.Pointer(&c.cx[0]))
}

// =============================================================================
//  [HID] Serial
// =============================================================================

func (d *dhw) serialConfigure() {

	// hid := &descHID[d.cc.config-1]

	// SAMx51 only supports USB full-speed (FS) operation
	// hid.rxSerialSize = descHIDSerialRxFSPacketSize
	// hid.txSerialSize = descHIDSerialTxFSPacketSize

	// Rx and Tx are on same endpoint
	d.endpointEnable(descHIDEndpointSerialRx,
		false, descHIDConfigAttrSerial)

	// d.endpointConfigureRx(descHIDEndpointSerialRx,
	// 	hid.rxSerialSize, false, d.serialNotify)
	// d.endpointConfigureTx(descHIDEndpointSerialTx,
	// 	hid.txSerialSize, false, nil)

	// for i := range hid.rdSerial {
	// 	d.serialReceive(uint8(i))
	// }

	// d.timerConfigure(0, descHIDSerialTxSyncUs, d.serialSync)
}

func (d *dhw) serialReceive(endpoint uint8) {
	hid := &descHID[d.cc.config-1]
	num := uint16(endpoint) & descEndptAddrNumberMsk
	_, _ = hid, num // TODO(ardnew): elaborate stub
}

func (d *dhw) serialTransmit() {
	hid := &descHID[d.cc.config-1]
	_ = hid // TODO(ardnew): elaborate stub
}

func (d *dhw) serialNotify( /* transfer *dhwTransfer */ ) {
	// hid := &descHID[d.cc.config-1]
	// len := hid.rxSerialSize - (uint16(transfer.token>>16) & 0x7FFF)
	// _ = len // TODO(ardnew): elaborate stub
}

// serialFlush discards all buffered input (Rx) data.
func (d *dhw) serialFlush() {
	hid := &descHID[d.cc.config-1]
	_ = hid
}

func (d *dhw) serialSync() {

}

// =============================================================================
//  [HID] Keyboard
// =============================================================================

func (d *dhw) keyboard() *Keyboard { return descHID[d.cc.config-1].keyboard }

func (d *dhw) keyboardConfigure() {

	hid := &descHID[d.cc.config-1]

	// Initialize keyboard
	hid.keyboard.configure(d.dcd, hid)

	// SAMx51 only supports USB full-speed (FS) operation
	hid.txKeyboardSize = descHIDKeyboardTxPacketSize

	// tq := hid.txqKeyboard[:]

	// hid.tqKeyboard.Init(&tq, len(hid.txqKeyboard), QueueFullDiscardFirst)

	d.endpointEnable(txEndpoint(descHIDEndpointKeyboard),
		false, descHIDConfigAttrKeyboard)
	d.endpointEnable(txEndpoint(descHIDEndpointMediaKey),
		false, descHIDConfigAttrMediaKey)

	d.endpointConfigure(txEndpoint(descHIDEndpointKeyboard),
		d.keyboardWriteComplete)
	d.endpointConfigure(txEndpoint(descHIDEndpointMediaKey),
		d.keyboardWriteComplete)
}

func (d *dhw) keyboardSendKeys(consumer bool) bool {

	hid := &descHID[d.cc.config-1]
	data := [8]uint8{}

	if !consumer {

		data[0] = hid.keyboard.mod
		data[1] = 0
		data[2] = hid.keyboard.key[0]
		data[3] = hid.keyboard.key[1]
		data[4] = hid.keyboard.key[2]
		data[5] = hid.keyboard.key[3]
		data[6] = hid.keyboard.key[4]
		data[7] = hid.keyboard.key[5]

		return d.keyboardWrite(txEndpoint(descHIDEndpointKeyboard), data[:])

	} else {

		// 44444444 44333333 33332222 22222211 11111111  [ word ]
		// 98765432 10987654 32109876 54321098 76543210  [ index ]  (right-to-left)

		data[1] = uint8((hid.keyboard.con[1] << 2) | ((hid.keyboard.con[0] >> 8) & 0x03))
		data[2] = uint8((hid.keyboard.con[2] << 4) | ((hid.keyboard.con[1] >> 6) & 0x0F))
		data[3] = uint8((hid.keyboard.con[3] << 6) | ((hid.keyboard.con[2] >> 4) & 0x3F))
		data[4] = uint8(hid.keyboard.con[3] >> 2)
		data[5] = hid.keyboard.sys[0]
		data[6] = hid.keyboard.sys[1]
		data[7] = hid.keyboard.sys[2]

		return d.keyboardWrite(txEndpoint(descHIDEndpointMediaKey), data[:])

	}
}

func (d *dhw) keyboardWriteComplete(endpoint uint8, size uint32) {
	hid := &descHID[d.cc.config-1]
	num := uint16(endpoint) & descEndptAddrNumberMsk
	if size > 0 && size%uint32(hid.txKeyboardSize) == 0 {
		// Send ZLP if transfer length is a non-zero multiple of max packet size.
		d.endpointTransfer(endpoint, 0, 0)
	}
	d.ep[num][descDirTx].setActiveTransfer(nil)
}

func (d *dhw) keyboardWrite(endpoint uint8, data []uint8) bool {

	hid := &descHID[d.cc.config-1]
	num := uint16(endpoint) & descEndptAddrNumberMsk

	for off := 0; off < len(data); off += int(hid.txKeyboardSize) {

		cnt := len(data[off:])
		if cnt > int(hid.txKeyboardSize) {
			cnt = int(hid.txKeyboardSize)
		}

		for d.ep[num][descDirTx].hasActiveTransfer() {
		}

		ready, _ := d.ep[num][descDirTx].scheduleTransfer(
			uintptr(unsafe.Pointer(&data[0])), uint32(cnt))
		if ready {
			if xfer, ok := d.ep[num][descDirTx].pendingTransfer(); ok {
				d.ep[num][descDirTx].setActiveTransfer(xfer)
				d.endpointTransfer(endpoint, xfer.data, xfer.size)
			}
		}
	}

	// size := uint16(len(data))
	// xfer := &hid.tdKeyboard[hid.txKeyboardHead]
	// when := ticks()
	// for {
	// 	if 0 == xfer.token&0x80 {
	// 		if 0 != xfer.token&0x68 {
	// 			// TODO: token contains error, how to handle?
	// 		}
	// 		hid.txKeyboardPrev = false
	// 		break
	// 	}
	// 	if hid.txKeyboardPrev {
	// 		return false
	// 	}
	// 	if ticks()-when > descHIDKeyboardTxTimeoutMs {
	// 		// Waited too long, assume host connection dropped
	// 		hid.txKeyboardPrev = true
	// 		return false
	// 	}
	// }
	// // Without this delay, the order packets are transmitted is seriously screwy.
	// udelay(60)
	// buff := hid.txKeyboard[hid.txKeyboardHead*descHIDKeyboardTxSize:]
	// _ = copy(buff, data)
	// d.transferPrepare(xfer, &buff[0], size, 0)
	// flushCache(uintptr(unsafe.Pointer(&buff[0])), descHIDKeyboardTxSize)
	// d.endpointTransmit(endpoint, xfer)
	// hid.txKeyboardHead += 1
	// if hid.txKeyboardHead >= descHIDKeyboardTDCount {
	// 	hid.txKeyboardHead = 0
	// }
	return true
}

// =============================================================================
//  [HID] Mouse
// =============================================================================

func (d *dhw) mouseConfigure() {

	// hid := &descHID[d.cc.config-1]

	// SAMx51 only supports USB full-speed (FS) operation
	// hid.txMouseSize = descHIDMouseTxFSPacketSize

	d.endpointEnable(descHIDEndpointMouse,
		false, descHIDConfigAttrMouse)

	// d.endpointConfigureTx(descHIDEndpointMouse,
	// 	hid.txMouseSize, false, nil)
}

// =============================================================================
//  [HID] Joystick
// =============================================================================

func (d *dhw) joystickConfigure() {

	// hid := &descHID[d.cc.config-1]

	// SAMx51 only supports USB full-speed (FS) operation
	// hid.txJoystickSize = descHIDJoystickTxFSPacketSize

	d.endpointEnable(descHIDEndpointJoystick,
		false, descHIDConfigAttrJoystick)

	// d.endpointConfigureTx(descHIDEndpointJoystick,
	// 	hid.txJoystickSize, false, nil)
}
