package usb2

// Hardware abstraction for USB ports configured as either host or device.

import "unsafe"

func init() {
	if unsafe.Sizeof(uintptr(0)) > 4 {
		panic("USB is only supported on 32-bit systems")
	}
}

// core represents the core of a USB port configured as either host or device.
type core struct {
	port int
	mode int
	dc   dcd
	hc   hcd
}

// Constant definitions for USB core operating modes.
const (
	modeIdle   = 0
	modeDevice = 1
	modeHost   = 2
)

// class represents the type of a host/device and its class configuration index.
// The first valid configuration index is 1. Index 0 is reserved and invalid.
type class struct {
	id     int
	config int
}

// Constant definitions for all host/device classes.
const (
	classDeviceCDCACM = 0 // The only currently-supported class (CDC-ACM)
)

// mode returns the USB core operating mode of the receiver class cl.
//go:inline
func (cl class) mode() int {
	switch cl.id {
	case classDeviceCDCACM:
		return modeDevice
	default:
		return modeIdle
	}
}

// equals returns true if and only if all fields of the given class are equal to
// those of the receiver cl.
//go:inline
func (cl class) equals(class class) bool {
	return cl.id == class.id && cl.config == class.config
}

// CoreCount defines the total number of USB cores to configure in device or
// host mode.
const CoreCount = dcdCount + hcdCount

// coreInstance provides statically-allocated instances of each USB core
// configured on this platform.
var coreInstance [CoreCount]core

// status represents the return code of a subroutine.
type status uint8

// Constant definitions for all status codes used within the package.
const (
	statusOK      status = iota // Success
	statusBusy                  // Busy
	statusRetry                 // Retry
	statusInvalid               // Invalid argument
)

// ok returns true if and only if the receiver st equals statusOK.
//go:inline
func (st status) ok() bool { return statusOK == st }

// initCore initializes a free USB core with given operating mode on the USB
// port at given index, if available. Returns a reference to the initialized
// core or nil if the core is unavailable.
func initCore(port int, class class) (*core, status) {

	if port < 0 || port >= CoreCount || 0 == class.config {
		return nil, statusInvalid
	}

	if modeIdle != coreInstance[port].mode {
		// Check if requested port is already configured as requested class. If so,
		// just return a reference to the existing core instead of an error.
		// For instance, this will allow TinyGo examples that try to reconfigure the
		// USB (CDC-ACM) UART port (which is already configured by the runtime) to
		// continue without error.
		if coreInstance[port].mode == class.mode() {
			switch class.mode() {
			case modeDevice:
				if coreInstance[port].dc.class().equals(class) {
					return &coreInstance[port], statusOK
				}
			case modeHost:
				if coreInstance[port].hc.class().equals(class) {
					return &coreInstance[port], statusOK
				}
			}
		}
		return nil, statusBusy
	}

	switch class.mode() {
	case modeDevice:
		// Allocate a free device controller and install interrupts
		dc, st := initDCD(port, class)
		if !st.ok() {
			return nil, st
		}
		// Initialize buffers and device descriptors
		if st = dc.init(); !st.ok() {
			return nil, st
		}
		coreInstance[port].port = port
		coreInstance[port].mode = modeDevice
		coreInstance[port].dc = dc
		// Enable interrupts and enter runtime
		if st = dc.enable(true); !st.ok() {
			coreInstance[port].mode = modeIdle
			coreInstance[port].dc = nil
			return nil, st
		}

	case modeHost:
		// Allocate a free host controller and install interrupts
		hc, st := initHCD(port, class)
		if !st.ok() {
			return nil, st
		}
		// Initialize buffers and device descriptors
		if st = hc.init(); !st.ok() {
			return nil, st
		}
		coreInstance[port].port = port
		coreInstance[port].mode = modeHost
		coreInstance[port].hc = hc
		// Enable interrupts and enter runtime
		if st = hc.enable(true); !st.ok() {
			coreInstance[port].mode = modeIdle
			coreInstance[port].hc = nil
			return nil, st
		}

	default:
		return nil, statusInvalid
	}

	return &coreInstance[port], statusOK
}
