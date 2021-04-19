// +build mimxrt1062

package usb2

// descCPUFrequencyHz defines the target CPU frequency (Hz).
const descCPUFrequencyHz = 600000000

// General USB device identification constants.
const (
	descVendorID  = 0x16C0
	descProductID = 0x0483
	descReleaseID = 0x0101

	descManufacturer = "NXP Semiconductors"
	descProduct      = "TinyGo USB"
	descSerialNumber = "0000000000"
)

// Constants for USB CDC-ACM device classes.
const (
	// descCDCACMCount defines the number of USB cores that will be configured as
	// CDC-ACM (single) devices.
	descCDCACMCount = 1

	descCDCACMQHCount = 2 * (descCDCACMEndpointCount + 1)
	descCDCACMRDCount = 2 * descCDCACMEndpointCount
	descCDCACMTDCount = descCDCACMEndpointCount
	descCDCACMRxCount = descCDCACMRxSize * descCDCACMRDCount
	descCDCACMTxCount = descCDCACMTxSize * descCDCACMTDCount
	descCDCACMCxCount = 8

	descCDCACMMaxPower = 50 // 100 mA

	descCDCACMStatusPacketSize = 16
	descCDCACMDataRxPacketSize = descCDCACMDataRxHSPacketSize // high-speed
	descCDCACMDataTxPacketSize = descCDCACMDataTxHSPacketSize // high-speed
	descCDCACMRxSize           = descCDCACMDataRxPacketSize
	descCDCACMTxSize           = 4 * descCDCACMDataTxPacketSize

	descCDCACMDataRxFSPacketSize = 64  // full-speed
	descCDCACMDataTxFSPacketSize = 64  // full-speed
	descCDCACMDataRxHSPacketSize = 512 // high-speed
	descCDCACMDataTxHSPacketSize = 512 // high-speed
)

// descCDCACM0QH is an array of endpoint queue heads, which is where all
// transfers for a given endpoint are managed, for the default CDC-ACM (single)
// device class configuration (index 1).
//
// From the iMXRT1062 Reference Manual:
//
//   Software must ensure that no interface data structure reachable
//   by the Device Controller spans a 4K-page boundary.
//
//   The [queue head] is a 48-byte data structure, but must be aligned on
//   64-byte boundaries.
//
//   Endpoint queue heads are arranged in an array in a continuous area of
//   memory pointed to by the USB.ENDPOINTLISTADDR pointer. The even-numbered
//   device queue heads in the list support receive endpoints (OUT/SETUP) and
//   the odd-numbered queue heads in the list are used for transmit endpoints
//   (IN/INTERRUPT). The device controller will index into this array based upon
//   the endpoint number received from the USB bus. All information necessary to
//   respond to transactions for all primed transfers is contained in this list
//   so the Device Controller can readily respond to incoming requests without
//   having to traverse a linked list.
//go:align 4096
var descCDCACM0QH [descCDCACMQHCount]dcdEndpoint

// descCDCACM0CD is the transfer descriptor for data messages transmitted or
// received on the status/control endpoint 0 for the default CDC-ACM (single)
// device class configuration (index 1).
//go:align 32
var descCDCACM0CD dcdTransfer

// descCDCACM0AD is the transfer descriptor for ackowledgement (ACK) messages
// transmitted or received on the status/control endpoint 0 for the default
// CDC-ACM (single) device class configuration (index 1).
//go:align 32
var descCDCACM0AD dcdTransfer

// descCDCACM0RD is an array of transfer descriptors for Rx (OUT) transfers,
// which describe to the device controller the location and quantity of data
// being received for a given transfer, for the default CDC-ACM (single) device
// class configuration (index 1).
//go:align 32
var descCDCACM0RD [descCDCACMRDCount]dcdTransfer

// descCDCACM0TD is an array of transfer descriptors for Tx (IN) transfers,
// which describe to the device controller the location and quantity of data
// being transmitted for a given transfer, for the default CDC-ACM (single)
// device class configuration (index 1).
//go:align 32
var descCDCACM0TD [descCDCACMTDCount]dcdTransfer

// descCDCACM0Cx is the buffer for control/status data received on endpoint 0 of
// the default CDC-ACM (single) device class configuration (index 1).
//go:align 32
var descCDCACM0Cx [descCDCACMCxCount]uint8

// descCDCACM0Rx is the receive (Rx) buffer of data endpoints for the default
// CDC-ACM (single) device class configuration (index 1).
//go:align 32
var descCDCACM0Rx [descCDCACMRxCount]uint8

// descCDCACM0Tx is the transmit (Tx) buffer of data endpoints for the default
// CDC-ACM (single) device class configuration (index 1).
//go:align 32
var descCDCACM0Tx [descCDCACMTxCount]uint8

// descCDCACM0Dx is the transmit (Tx) buffer of descriptor data for the default
// CDC-ACM (single) device class configuration (index 1).
var descCDCACM0Dx [descCDCACMConfigSize]uint8

var descCDCACM0RDNum [descCDCACMRDCount]uint16
var descCDCACM0RDIdx [descCDCACMRDCount]uint16
var descCDCACM0RDQue [descCDCACMRDCount + 1]uint16

type descCDCACMClass struct {
	locstr *[descCDCACMLanguageCount]descLocalStrings // string descriptors
	device *[descLengthDevice]uint8                   // device descriptor
	qualif *[descLengthQualification]uint8            // device qualification descriptor
	config *[descCDCACMConfigSize]uint8               // configuration descriptor

	coding *[descCDCACMCodingSize]uint8 // UART line coding
	cticks int64
	rtsdtr uint8

	qh *[descCDCACMQHCount]dcdEndpoint // endpoint queue heads

	cd *dcdTransfer                    // control endpoint 0 Rx/Tx data transfer descriptor
	ad *dcdTransfer                    // control endpoint 0 Rx/Tx ACK transfer descriptor
	rd *[descCDCACMRDCount]dcdTransfer // bulk data endpoint Rx (OUT) transfer descriptors
	td *[descCDCACMTDCount]dcdTransfer // bulk data endpoint Tx (IN) transfer descriptors

	cx *[descCDCACMCxCount]uint8    // control endpoint 0 Rx/Tx transfer buffer
	rx *[descCDCACMRxCount]uint8    // bulk data endpoint Rx (OUT) transfer buffer
	tx *[descCDCACMTxCount]uint8    // bulk data endpoint Tx (IN) transfer buffer
	dx *[descCDCACMConfigSize]uint8 // descriptor data Tx (IN) transfer buffer

	cxSize uint16
	rxSize uint16
	txSize uint16

	txHead uint8
	txFree uint16

	rxHead uint8
	rxTail uint8
	rxFree uint16

	rxCount *[descCDCACMRDCount]uint16
	rxIndex *[descCDCACMRDCount]uint16
	rxQueue *[descCDCACMRDCount + 1]uint16
}

// descCDCACM holds the configuration, endpoint, and transfer descriptors, along
// with the buffers and control states, for all of the CDC-ACM (single) device
// class configurations, ordered by configuration index (offset by -1).
var descCDCACM = [descCDCACMCount]descCDCACMClass{
	{
		locstr: &descCDCACM0String,
		device: &descCDCACM0Device,
		qualif: &descCDCACM0Qualif,
		config: &descCDCACM0Config,

		coding: &descCDCACM0Coding,

		qh: &descCDCACM0QH,

		cd: &descCDCACM0CD,
		ad: &descCDCACM0AD,
		rd: &descCDCACM0RD,
		td: &descCDCACM0TD,

		cx: &descCDCACM0Cx,
		rx: &descCDCACM0Rx,
		tx: &descCDCACM0Tx,
		dx: &descCDCACM0Dx,

		cxSize: descCDCACMStatusPacketSize,
		rxSize: descCDCACMDataRxPacketSize,
		txSize: descCDCACMDataTxPacketSize,

		rxCount: &descCDCACM0RDNum,
		rxIndex: &descCDCACM0RDIdx,
		rxQueue: &descCDCACM0RDQue,
	},
}
