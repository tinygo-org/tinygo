package usb2

// Hardware abstraction for USB ports configured as either host or device.

// core represents the core of a USB port configured as either host or device.
type core struct {
	port int
	mode int
	dc   dci
	hc   hci
}

// Constant definitions for USB core operating modes.
const (
	modeIdle   = 0
	modeDevice = 1
	modeHost   = 2
)

// CoreCount defines the total number of USB cores to configure in device or
// host mode.
const CoreCount = dciCount + hciCount

// coreInstance provides statically-allocated instances of each USB core
// configured on this platform.
var coreInstance [CoreCount]core

// status represents the return code of a subroutine.
type status uint8

// Constant definitions for all status codes used within the package.
const (
	statusOK              status = iota // Success
	statusBusy                          // Busy
	statusRetry                         // Retry
	statusInvalidArgument               // Invalid argument
)

func (st status) ok() bool { return statusOK == st }

// initCore initializes a free USB core with given operating mode on the USB
// port at given index, if available. Returns a reference to the initialized
// core or nil if the core is unavailable.
func initCore(port, mode int) (*core, status) {

	if port < 0 || port >= CoreCount {
		return nil, statusInvalidArgument
	}
	if modeIdle != coreInstance[port].mode {
		return nil, statusBusy
	}

	switch mode {
	case modeDevice:
		// Allocate a free device controller and install interrupts
		dc, st := initDCI(port)
		if !st.ok() {
			return nil, st
		}
		// Initialize buffers and device descriptors
		if st = dc.init(); !st.ok() {
			return nil, st
		}
		coreInstance[port].port = port
		coreInstance[port].mode = mode
		coreInstance[port].dc = dc
		// Enable interrupts and enter runtime
		if st = dc.enable(true); !st.ok() {
			coreInstance[port].mode = modeIdle
			coreInstance[port].dc = nil
			return nil, st
		}

	case modeHost:
		// Allocate a free host controller and install interrupts
		hc, st := initHCI(port)
		if !st.ok() {
			return nil, st
		}
		// Initialize buffers and device descriptors
		if st = hc.init(); !st.ok() {
			return nil, st
		}
		coreInstance[port].port = port
		coreInstance[port].mode = mode
		coreInstance[port].hc = hc
		// Enable interrupts and enter runtime
		if st = hc.enable(true); !st.ok() {
			coreInstance[port].mode = modeIdle
			coreInstance[port].hc = nil
			return nil, st
		}

	default:
		return nil, statusInvalidArgument
	}

	return &coreInstance[port], statusOK
}
