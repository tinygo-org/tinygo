// +build mimxrt1062

package usb

// descCPUFrequencyHz defines the target CPU frequency (Hz).
const descCPUFrequencyHz = 600000000

// General USB device identification constants.
const (
	descCommonVendorID  = 0x16C0
	descCommonProductID = 0x0483
	descCommonReleaseID = 0x0101 // BCD (1.1)

	descCommonLanguage     = descLanguageEnglish
	descCommonManufacturer = "NXP Semiconductors"
	descCommonProduct      = "TinyGo USB"
	descCommonSerialNumber = "1"
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

	descCDCACMTxTimeoutMs = 120 // millisec
	descCDCACMTxSyncUs    = 75  // microsec

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
var descCDCACM0QH [descCDCACMQHCount]dhwEndpoint

// descCDCACM0CD is the transfer descriptor for data messages transmitted or
// received on the status/control endpoint 0 for the default CDC-ACM (single)
// device class configuration (index 1).
//go:align 32
var descCDCACM0CD dhwTransfer

// descCDCACM0AD is the transfer descriptor for ackowledgement (ACK) messages
// transmitted or received on the status/control endpoint 0 for the default
// CDC-ACM (single) device class configuration (index 1).
//go:align 32
var descCDCACM0AD dhwTransfer

// descCDCACM0RD is an array of transfer descriptors for Rx (OUT) transfers,
// which describe to the device controller the location and quantity of data
// being received for a given transfer, for the default CDC-ACM (single) device
// class configuration (index 1).
//go:align 32
var descCDCACM0RD [descCDCACMRDCount]dhwTransfer

// descCDCACM0TD is an array of transfer descriptors for Tx (IN) transfers,
// which describe to the device controller the location and quantity of data
// being transmitted for a given transfer, for the default CDC-ACM (single)
// device class configuration (index 1).
//go:align 32
var descCDCACM0TD [descCDCACMTDCount]dhwTransfer

// descCDCACM0LineCoding holds the emulated UART line coding for the default
// CDC-ACM (single) device class configuration (index 1).
// i.e., configuration index 1.
//go:align 32
var descCDCACM0LC descCDCACMLineCoding

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

// descCDCACM holds the buffers and control states for all of the CDC-ACM
// (single) device class configurations, ordered by index (offset by -1), for
// iMXRT1062 targets only.
//
// Instances of this type (elements of descCDCACMData) are embedded in elements
// of the common/target-agnostic CDC-ACM class configurations (descCDCACM).
// Methods defined on this type implement target-specific functionality, and
// some of these methods are required by the common device controller driver.
// Thus, this type functions as an additional hardware abstraction layer.
type descCDCACMClassData struct {
	qh *[descCDCACMQHCount]dhwEndpoint // endpoint queue heads

	cd *dhwTransfer                    // control endpoint 0 Rx/Tx data transfer descriptor
	ad *dhwTransfer                    // control endpoint 0 Rx/Tx ACK transfer descriptor
	rd *[descCDCACMRDCount]dhwTransfer // bulk data endpoint Rx (OUT) transfer descriptors
	td *[descCDCACMTDCount]dhwTransfer // bulk data endpoint Tx (IN) transfer descriptors

	cx *[descCDCACMCxCount]uint8    // control endpoint 0 Rx/Tx transfer buffer
	rx *[descCDCACMRxCount]uint8    // bulk data endpoint Rx (OUT) transfer buffer
	tx *[descCDCACMTxCount]uint8    // bulk data endpoint Tx (IN) transfer buffer
	dx *[descCDCACMConfigSize]uint8 // descriptor data Tx (IN) transfer buffer

	cxSize uint16
	rxSize uint16
	txSize uint16

	txHead uint8
	txFree uint16
	txPrev bool

	rxHead uint8
	rxTail uint8
	rxFree uint16

	rxCount *[descCDCACMRDCount]uint16
	rxIndex *[descCDCACMRDCount]uint16
	rxQueue *[descCDCACMRDCount + 1]uint16

	_ [2]uint8
}

// descCDCACMData holds statically-allocated instances for each of the target-
// specific (iMXRT1062) CDC-ACM (single) device class configurations' control
// and data structures, ordered by configuration index (offset by -1). Each
// element is embedded in a corresponding element of descCDCACM.
//go:align 32
var descCDCACMData = [descCDCACMCount]descCDCACMClassData{

	{ // CDC-ACM (single) class configuration index 1 data
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
