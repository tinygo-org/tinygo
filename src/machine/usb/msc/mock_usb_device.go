package msc

import (
	"machine/usb"
	"machine/usb/descriptor"
)

// MockUSBDevice implements usb.Controller for testing.
type MockUSBDevice struct {
	InPackets        [][]byte
	OutAcked         bool
	StallIn          bool
	StallOut         bool
	InitEndpointDone bool
	ZlpSent          bool
}

func (m *MockUSBDevice) ConfigureUSBEndpoint(desc descriptor.Descriptor, epSettings []usb.EndpointConfig, setup []usb.SetupConfig) {
}

func (m *MockUSBDevice) SendUSBInPacket(ep uint32, data []byte) bool {
	// Copy data to avoid modification issues if buffer is reused
	packet := make([]byte, len(data))
	copy(packet, data)
	m.InPackets = append(m.InPackets, packet)
	return true
}

func (m *MockUSBDevice) AckUsbOutTransfer(ep uint32) {
	m.OutAcked = true
}

func (m *MockUSBDevice) SendZlp() {
	m.ZlpSent = true
	m.SendUSBInPacket(0, []byte{})
}

func (m *MockUSBDevice) IsInitEndpointComplete() bool {
	return m.InitEndpointDone
}

func (m *MockUSBDevice) SetStallEPIn(ep uint32) {
	m.StallIn = true
}

func (m *MockUSBDevice) SetStallEPOut(ep uint32) {
	m.StallOut = true
}

func (m *MockUSBDevice) ClearStallEPIn(ep uint32) {
	m.StallIn = false
}

func (m *MockUSBDevice) ClearStallEPOut(ep uint32) {
	m.StallOut = false
}

func (m *MockUSBDevice) Enable() {
}

func (m *MockUSBDevice) ReceiveUSBControlPacket() ([7]byte, error) {
	return [7]byte{}, nil
}
