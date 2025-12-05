package msc

import (
	"encoding/binary"
	"machine/usb"
	"machine/usb/msc/csw"
	"testing"
)

// MockBlockDevice implements machine.BlockDevice for testing.
type MockBlockDevice struct {
	Data        []byte
	BlockSize   int64
	ReadCount   int
	WriteCount  int
	LastWriteAt int64
}

func (m *MockBlockDevice) ReadAt(p []byte, off int64) (n int, err error) {
	m.ReadCount++
	if off >= int64(len(m.Data)) {
		return 0, nil
	}
	n = copy(p, m.Data[off:])
	return n, nil
}

func (m *MockBlockDevice) WriteAt(p []byte, off int64) (n int, err error) {
	m.WriteCount++
	m.LastWriteAt = off
	if off >= int64(len(m.Data)) {
		// Expand data if needed
		newData := make([]byte, off+int64(len(p)))
		copy(newData, m.Data)
		m.Data = newData
	}
	n = copy(m.Data[off:], p)
	return n, nil
}

func (m *MockBlockDevice) Size() int64 {
	return int64(len(m.Data))
}

func (m *MockBlockDevice) WriteBlockSize() int64 {
	return m.BlockSize
}

func (m *MockBlockDevice) EraseBlockSize() int64 {
	return m.BlockSize
}

func (m *MockBlockDevice) EraseBlocks(start, len int64) error {
	return nil
}

// TestCBWParser verifies that the CBW is correctly parsed.
func TestCBWParser(t *testing.T) {
	// Create a valid CBW
	// Signature: USBC (0x43425355)
	// Tag: 0x12345678
	// Transfer Length: 512 (0x200)
	// Flags: 0x80 (IN)
	// LUN: 0
	// Length: 10
	// CBD: SCSI Read(10) command (dummy)
	cbwData := []byte{
		0x55, 0x53, 0x42, 0x43, // Signature
		0x78, 0x56, 0x34, 0x12, // Tag
		0x00, 0x02, 0x00, 0x00, // Data Transfer Length (512)
		0x80, // Flags (Direction: IN)
		0x00, // LUN
		0x0A, // CBD Length
		// SCSI Command (Read 10) - dummy
		0x28, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Padding
	}

	// Setup mocks
	mockDev := &MockBlockDevice{Data: make([]byte, 1024), BlockSize: 512}
	mockUSB := &MockUSBDevice{InitEndpointDone: true}

	// Initialize MSC
	m := newMSC(mockDev, mockUSB)

	// Manually feed the CBW to the run loop
	// run(b, true) simulates receiving an OUT packet
	ack := m.run(cbwData, true)

	if !ack {
		t.Error("Expected ACK for valid CBW")
	}

	// Check if CBW was parsed correctly
	if m.cbw.Tag() != 0x12345678 {
		t.Errorf("Expected Tag 0x12345678, got 0x%x", m.cbw.Tag())
	}

	if m.transferBytes != 512 {
		t.Errorf("Expected transferBytes 512, got %d", m.transferBytes)
	}

	if m.state != mscStateData {
		t.Errorf("Expected state mscStateData, got %d", m.state)
	}
}

// TestResidueLogic simulates a Short Write scenario and verifies the Residue calculation.
func TestResidueLogic(t *testing.T) {
	// CBW for WRITE (OUT), 512 bytes
	cbwData := []byte{
		0x55, 0x53, 0x42, 0x43, // Signature
		0x11, 0x22, 0x33, 0x44, // Tag
		0x00, 0x02, 0x00, 0x00, // Data Transfer Length (512)
		0x00, // Flags (Direction: OUT)
		0x00, // LUN
		0x0A, // CBD Length
		// SCSI Write(10) command
		0x2A, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Padding
	}

	mockDev := &MockBlockDevice{Data: make([]byte, 1024), BlockSize: 512}
	mockUSB := &MockUSBDevice{InitEndpointDone: true}
	m := newMSC(mockDev, mockUSB)

	// 1. Send CBW
	m.run(cbwData, true)
	if m.state != mscStateData {
		t.Fatalf("Failed to transition to Data state. State: %d", m.state)
	}

	// 2. Host sends 256 bytes (half block)
	dataPacket := make([]byte, 256)
	m.run(dataPacket, true)

	// queuedBytes should be 256. sentBytes (bytes written to block device) should be 0 because we buffer a full block.
	if m.queuedBytes != 256 {
		t.Errorf("Expected queuedBytes 256, got %d", m.queuedBytes)
	}
	if m.sentBytes != 0 {
		t.Errorf("Expected sentBytes 0, got %d", m.sentBytes)
	}

	// Force state to Status to verify Residue calculation
	m.state = mscStateStatus

	// Call run to trigger CSW send
	m.run([]byte{}, false) // IN endpoint event (dummy)

	// Check if CSW was sent
	if len(mockUSB.InPackets) == 0 {
		t.Fatal("No CSW sent")
	}

	// The last packet should be the CSW
	cswPacket := mockUSB.InPackets[len(mockUSB.InPackets)-1]
	if len(cswPacket) != csw.MsgLen {
		t.Errorf("CSW length mismatch. Expected %d, got %d", csw.MsgLen, len(cswPacket))
	}

	// Parse CSW
	// Signature: 0-4
	// Tag: 4-8
	// Residue: 8-12
	// Status: 12

	signature := binary.LittleEndian.Uint32(cswPacket[:4])
	if signature != csw.Signature {
		t.Errorf("Invalid CSW Signature: %x", signature)
	}

	tag := binary.LittleEndian.Uint32(cswPacket[4:8])
	if tag != 0x44332211 { // Little Endian of 11 22 33 44
		t.Errorf("Invalid CSW Tag: %x", tag)
	}

	residue := binary.LittleEndian.Uint32(cswPacket[8:12])
	// Expected Residue = Expected Length (512) - Processed (256) = 256
	if residue != 256 {
		t.Errorf("Incorrect Residue. Expected 256, got %d", residue)
	}

	status := cswPacket[12]
	if status != byte(csw.StatusPassed) {
		t.Errorf("Incorrect Status. Expected %d (Passed), got %d", csw.StatusPassed, status)
	}
}

func TestSetupPacketHandler(t *testing.T) {
	mockDev := &MockBlockDevice{Data: make([]byte, 1024), BlockSize: 512}
	mockUSB := &MockUSBDevice{InitEndpointDone: true}
	m := newMSC(mockDev, mockUSB)

	// Test Get Max LUN (Class Request 0xFE)
	setup := usb.Setup{
		BmRequestType: 0xA1, // Device-to-Host, Class, Interface
		BRequest:      0xFE, // GET MAX LUN
		WValueL:       0,
		WValueH:       0,
		WIndex:        mscInterface,
		WLength:       1,
	}

	handled := m.setupPacketHandler(setup)
	if !handled {
		t.Error("Expected GetMaxLUN to be handled")
	}

	if len(mockUSB.InPackets) != 1 {
		t.Fatalf("Expected 1 IN packet (Max LUN), got %d", len(mockUSB.InPackets))
	}

	if mockUSB.InPackets[0][0] != m.maxLUN {
		t.Errorf("Expected MaxLUN %d, got %d", m.maxLUN, mockUSB.InPackets[0][0])
	}
}
