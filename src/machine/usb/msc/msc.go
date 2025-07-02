package msc

import (
	"machine"
	"machine/usb"
	"machine/usb/descriptor"
	"machine/usb/msc/csw"
	"machine/usb/msc/scsi"
	"time"
)

type mscState uint8

const (
	mscStateCmd mscState = iota
	mscStateData
	mscStateStatus
	mscStateStatusSent
	mscStateNeedReset
)

const (
	mscInterface = 2
)

var MSC *msc

type msc struct {
	buf           []byte     // Buffer for incoming/outgoing data
	blockCache    []byte     // Buffer for block read/write data
	taskQueued    bool       // Flag to indicate if the buffer has a task queued
	rxStalled     bool       // Flag to indicate if the RX endpoint is stalled
	txStalled     bool       // Flag to indicate if the TX endpoint is stalled
	maxPacketSize uint32     // Maximum packet size for the IN endpoint
	respStatus    csw.Status // Response status for the last command
	sendZLP       bool       // Flag to indicate if a zero-length packet should be sent before sending CSW

	cbw         *CBW   // Last received Command Block Wrapper
	queuedBytes uint32 // Number of bytes queued for sending
	sentBytes   uint32 // Number of bytes sent
	totalBytes  uint32 // Total bytes to send
	cswBuf      []byte // CSW response buffer
	state       mscState

	maxLUN       uint8 // Maximum Logical Unit Number (n-1 for n LUNs)
	dev          machine.BlockDevice
	blockCount   uint32 // Number of blocks in the device
	blockOffset  uint32 // Byte offset of the first block in the device for aligned writes
	blockSizeUSB uint32 // Write block size as presented to the host over USB
	blockSizeRaw uint32 // Write block size of the underlying device hardware
	readOnly     bool

	vendorID   [8]byte  // Max 8 ASCII characters
	productID  [16]byte // Max 16 ASCII characters
	productRev [4]byte  // Max 4 ASCII characters

	senseKey           scsi.Sense
	addlSenseCode      scsi.SenseCode
	addlSenseQualifier uint8
}

// Port returns the USB Mass Storage port
func Port(dev machine.BlockDevice) *msc {
	if MSC == nil {
		MSC = newMSC(dev)
	}
	return MSC
}

func newMSC(dev machine.BlockDevice) *msc {
	// Size our buffer to match the maximum packet size of the IN endpoint
	maxPacketSize := descriptor.EndpointMSCIN.GetMaxPacketSize()
	m := &msc{
		// Some platforms require reads/writes to be aligned to the full underlying hardware block
		blockCache:    make([]byte, dev.WriteBlockSize()),
		blockSizeUSB:  512,
		buf:           make([]byte, dev.WriteBlockSize()),
		cswBuf:        make([]byte, csw.MsgLen),
		cbw:           &CBW{Data: make([]byte, 31)},
		maxPacketSize: uint32(maxPacketSize),
	}
	m.RegisterBlockDevice(dev)

	// Set default inquiry data fields
	m.SetVendorID("TinyGo")
	m.SetProductID("Mass Storage")
	m.SetProductRev("1.0")

	// Initialize the USB Mass Storage Class (MSC) port
	machine.ConfigureUSBEndpoint(descriptor.MSC,
		[]usb.EndpointConfig{
			{
				Index:        usb.MSC_ENDPOINT_IN,
				IsIn:         true,
				Type:         usb.ENDPOINT_TYPE_BULK,
				TxHandler:    txHandler,
				StallHandler: setupPacketHandler,
			},
			{
				Index:          usb.MSC_ENDPOINT_OUT,
				IsIn:           false,
				Type:           usb.ENDPOINT_TYPE_BULK,
				DelayRxHandler: rxHandler,
				StallHandler:   setupPacketHandler,
			},
		},
		[]usb.SetupConfig{
			{
				Index:   mscInterface,
				Handler: setupPacketHandler,
			},
		},
	)

	go m.processTasks()

	return m
}

func (m *msc) processTasks() {
	// Process tasks that cannot be done in an interrupt context
	for {
		if m.taskQueued {
			cmd := m.cbw.SCSICmd()
			switch cmd.CmdType() {
			case scsi.CmdWrite:
				m.scsiWrite(cmd, m.buf)
			case scsi.CmdUnmap:
				m.scsiUnmap(m.buf)
			}

			// Acknowledge the received data from the host
			m.queuedBytes = 0
			m.taskQueued = false
			machine.AckUsbOutTransfer(usb.MSC_ENDPOINT_OUT)
		}
		time.Sleep(100 * time.Microsecond)
	}
}

func (m *msc) ready() bool {
	return m.dev != nil
}

func (m *msc) resetBuffer(length int) {
	// Reset the buffer to the specified length
	m.buf = m.buf[:length]
	for i := 0; i < length; i++ {
		m.buf[i] = 0
	}
}

func (m *msc) sendUSBPacket(b []byte) {
	if machine.USBDev.InitEndpointComplete {
		// Send the USB packet
		machine.SendUSBInPacket(usb.MSC_ENDPOINT_IN, b)
	}
}

func (m *msc) sendCSW(status csw.Status) {
	// Generate CSW packet into m.cswBuf and send it
	residue := uint32(0)
	if m.totalBytes >= m.sentBytes {
		residue = m.totalBytes - m.sentBytes
	}
	m.cbw.CSW(status, residue, m.cswBuf)
	m.state = mscStateStatusSent
	m.sendUSBPacket(m.cswBuf)
}

func txHandler() {
	if MSC != nil {
		MSC.txHandler()
	}
}

func (m *msc) txHandler() {
	m.run([]byte{}, false)
}

func rxHandler(b []byte) bool {
	ack := true
	if MSC != nil {
		ack = MSC.run(b, true)
	}
	return ack
}

/*
	Connection Happy Path Overview:

0. MSC starts out in mscStateCmd status.

1. Host sends CBW (Command Block Wrapper) packet to MSC.
  - CBW contains the SCSI command to be executed, the length of the data to be transferred, etc.

2. MSC receives CBW.
  - CBW is validated and saved.
  - State is changed to mscStateData.
  - MSC routes the command to the appropriate SCSI command handler.

3. The MSC SCSI command handler responds with the initial data packet (if applicable).
  - If no data packet is needed, state is changed to mscStateStatus and step 4 is skipped.

4. The host acks the data packet and MSC calls m.scsiDataTransfer() to continue sending (or
receiving) data.
  - This cycle continues until all data requested in the CBW is sent/received.
  - State is changed to mscStateStatus.
  - MSC waits for the host to ACK the final data packet.

5. MSC then sends a CSW (Command Status Wrapper) to the host to report the final status of the
command execution and moves to mscStateStatusSent.

6. The host ACKs the CSW and the MSC moves back to mscStateCmd, waiting for the next CBW.
*/
func (m *msc) run(b []byte, isEpOut bool) bool {
	ack := true

	switch m.state {
	case mscStateCmd:
		// Receiving a new command block wrapper (CBW)

		// IN endpoint transfer complete confirmation, no action needed
		if !isEpOut {
			return ack
		}

		// Create a temporary CBW wrapper to validate the incoming data. Has to be temporary
		// to avoid it escaping into the heap since we're in interrupt context
		cbw := CBW{Data: b}

		// Verify size and signature
		if !cbw.validLength() || !cbw.validSignature() {
			// 6.6.1 CBW Not Valid
			// https://usb.org/sites/default/files/usbmassbulk_10.pdf
			m.state = mscStateNeedReset
			m.stallEndpoint(usb.MSC_ENDPOINT_IN)
			m.stallEndpoint(usb.MSC_ENDPOINT_OUT)
			m.stallEndpoint(usb.CONTROL_ENDPOINT)
			return ack
		}

		// Save the validated CBW for later reference
		copy(m.cbw.Data, b)

		// Move on to the data transfer phase next go around (after sending the first message)
		m.state = mscStateData
		m.totalBytes = cbw.transferLength()
		m.queuedBytes = 0
		m.sentBytes = 0
		m.respStatus = csw.StatusPassed

		m.scsiCmdBegin()

	case mscStateData:
		// Transfer data
		ack = m.scsiDataTransfer(b)

	case mscStateStatus:
	// Sending CSW status response
	// Placed after the switch statement so we can send the CSW without having to send a packet
	// to cycle back through this block, e.g. with TEST UNIT READY which sends only a CSW after
	// setting the sense key/add'l code/qualifier internally

	case mscStateStatusSent:
		// Wait for the status phase to complete
		if !isEpOut && m.queuedBytes == csw.MsgLen {
			// Status confirmed sent, wait for next CBW
			m.state = mscStateCmd
		} else {
			// We're not expecting any data here, ignore it. Original log line:
			// TU_LOG1("  Warning expect SCSI Status but received unknown data\r\n");
		}

	case mscStateNeedReset:
		// Received an invalid CBW message, stop everything until we get reset
	}

	// Send CSW status response
	// Placed after the switch statement so we can send the CSW without having to send a packet
	// to cycle back through this block, e.g. with TEST UNIT READY which sends only a CSW after
	// setting the sense key/add'l code/qualifier internally
	if m.state == mscStateStatus && !m.txStalled {
		if m.totalBytes > m.sentBytes && m.cbw.isIn() {
			// 6.7.2 The Thirteen Cases - Case 5 (Hi > Di): STALL before status
			m.stallEndpoint(usb.MSC_ENDPOINT_IN)
		} else if m.sendZLP {
			// Send a zero-length packet to force the end of the transfer before we send a CSW
			m.queuedBytes = 0
			m.sendZLP = false
			m.sendUSBPacket(m.buf[:0])
		} else {
			m.sendCSW(m.respStatus)
			m.state = mscStateCmd
		}
	}

	return ack
}
