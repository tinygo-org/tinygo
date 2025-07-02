package msc

import (
	"machine"
	"machine/usb"
	"machine/usb/msc/csw"
)

func setupPacketHandler(setup usb.Setup) bool {
	if MSC != nil {
		return MSC.setupPacketHandler(setup)
	}
	return false
}

func (m *msc) setupPacketHandler(setup usb.Setup) bool {
	ok := false
	wValue := (uint16(setup.WValueH) << 8) | uint16(setup.WValueL)
	switch setup.BRequest {
	case usb.CLEAR_FEATURE:
		ok = m.handleClearFeature(setup, wValue)
	case usb.GET_MAX_LUN:
		ok = m.handleGetMaxLun(setup, wValue)
	case usb.MSC_RESET:
		ok = m.handleReset(setup, wValue)
	}
	return ok
}

// Handles the CLEAR_FEATURE request for clearing ENDPOINT_HALT/stall
func (m *msc) handleClearFeature(setup usb.Setup, wValue uint16) bool {
	ok := false
	// wValue is the feature selector (0x00 for ENDPOINT_HALT)
	// We aren't handling any other feature selectors
	// https://wiki.osdev.org/Universal_Serial_Bus#CLEAR_FEATURE
	if wValue != 0 {
		return ok
	}
	// Clearing the stall is not enough, continue stalling until a reset is received first
	// 6.6.1 CBW Not Valid
	// If the CBW is not valid, the device shall STALL the Bulk-In pipe. Also, the device
	// shall either STALL the Bulk-Out pipe, or the device shall accept and discard any
	// Bulk-Out data. The device shall maintain this state until a Reset Recovery
	// For Reset Recovery the host shall issue in the following order: :
	// (a) a Bulk-Only Mass Storage Reset (handleReset())
	// (b) a Clear Feature HALT to the Bulk-In endpoint (clear stall IN)
	// (c) a Clear Feature HALT to the Bulk-Out endpoint (clear stall OUT)
	// https://usb.org/sites/default/files/usbmassbulk_10.pdf
	if m.state == mscStateNeedReset {
		wIndex := setup.WIndex & 0x7F // Clear the direction bit from the endpoint address for comparison
		if wIndex == usb.MSC_ENDPOINT_IN {
			m.stallEndpoint(usb.MSC_ENDPOINT_IN)
		} else if wIndex == usb.MSC_ENDPOINT_OUT {
			m.stallEndpoint(usb.MSC_ENDPOINT_OUT)
		}
		return ok
	}

	// Clear the direction bit from the endpoint address for comparison
	wIndex := setup.WIndex & 0x7F

	// Clear the IN/OUT stalls if addressed to the endpoint, or both if addressed to the interface
	if wIndex == usb.MSC_ENDPOINT_IN || wIndex == mscInterface {
		m.clearStallEndpoint(usb.MSC_ENDPOINT_IN)
		ok = true
	}
	if wIndex == usb.MSC_ENDPOINT_OUT || wIndex == mscInterface {
		m.clearStallEndpoint(usb.MSC_ENDPOINT_OUT)
		ok = true
	}
	// Send a CSW if needed to resume after the IN endpoint stall is cleared
	if m.state == mscStateStatus && wIndex == usb.MSC_ENDPOINT_IN {
		m.sendCSW(csw.StatusPassed)
		ok = true
	}

	if ok {
		machine.SendZlp()
	}
	return ok
}

// 3.2 Get Max LUN
// https://usb.org/sites/default/files/usbmassbulk_10.pdf
func (m *msc) handleGetMaxLun(setup usb.Setup, wValue uint16) bool {
	if setup.WIndex != mscInterface || setup.WLength != 1 || wValue != 0 {
		return false
	}
	// Send the maximum LUN ID number (zero-indexed, so n-1) supported by the device
	m.resetBuffer(1) // Shrink buffer to 1 byte
	m.buf[0] = m.maxLUN
	return machine.SendUSBInPacket(usb.CONTROL_ENDPOINT, m.buf)
}

// 3.1 Bulk-Only Mass Storage Reset
// https://usb.org/sites/default/files/usbmassbulk_10.pdf
func (m *msc) handleReset(setup usb.Setup, wValue uint16) bool {
	if setup.WIndex != mscInterface || setup.WLength != 0 || wValue != 0 {
		return false
	}
	// Reset to command waiting state
	m.state = mscStateCmd

	// Reset transfer state
	m.resetBuffer(0)
	m.senseKey = 0
	m.addlSenseCode = 0
	m.addlSenseQualifier = 0

	// Send a zero-length packet (ZLP) to indicate the reset is complete
	machine.SendZlp()

	// Return true to indicate successful reset
	return true
}

func (m *msc) stallEndpoint(ep uint8) {
	if ep == usb.MSC_ENDPOINT_IN {
		m.txStalled = true
		machine.USBDev.SetStallEPIn(usb.MSC_ENDPOINT_IN)
	} else if ep == usb.MSC_ENDPOINT_OUT {
		m.rxStalled = true
		machine.USBDev.SetStallEPOut(usb.MSC_ENDPOINT_OUT)
	} else if ep == usb.CONTROL_ENDPOINT {
		machine.USBDev.SetStallEPIn(usb.CONTROL_ENDPOINT)
	}
}

func (m *msc) clearStallEndpoint(ep uint8) {
	if ep == usb.MSC_ENDPOINT_IN {
		machine.USBDev.ClearStallEPIn(usb.MSC_ENDPOINT_IN)
		m.txStalled = false
	} else if ep == usb.MSC_ENDPOINT_OUT {
		machine.USBDev.ClearStallEPOut(usb.MSC_ENDPOINT_OUT)
		m.rxStalled = false
	}
}

func (m *msc) setStringField(field []byte, value string) {
	copy(field, []byte(value))
	for i := len(value); i < len(field); i++ {
		field[i] = 0x20 // Fill remaining bytes with spaces
	}
}

func (m *msc) SetVendorID(vendorID string) {
	m.setStringField(m.vendorID[:], vendorID)
}

func (m *msc) SetProductID(productID string) {
	m.setStringField(m.productID[:], productID)
}

func (m *msc) SetProductRev(productRev string) {
	m.setStringField(m.productRev[:], productRev)
}

func SetVendorID(vendorID string) {
	if MSC != nil {
		MSC.SetVendorID(vendorID)
	}
}
func SetProductID(productID string) {
	if MSC != nil {
		MSC.SetProductID(productID)
	}
}
func SetProductRev(productRev string) {
	if MSC != nil {
		MSC.SetProductRev(productRev)
	}
}
