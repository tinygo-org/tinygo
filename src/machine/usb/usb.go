package usb

// Hardware abstraction for USB ports configured as either host or device.

// CoreCount defines the total number of USB cores which may be configured in
// device or host mode.
const CoreCount = descCoreCount

// coreInstance provides statically-allocated instances of each USB core
// configured on this platform.
var coreInstance [CoreCount]core

// core represents the core of a USB port configured as either host or device.
type core struct {
	port int
	mode int
	dc   *dcd
	hc   *hcd
}

// Constant definitions for USB core operating modes.
const (
	modeIdle   = 0 // USB port has not been configured
	modeDevice = 1
	modeHost   = 2
)

// Speed represents the configured USB data transfer rate, or bus speed, for
// communication between host and device.
type Speed uint8

// Constant definitions for USB data transfer rates. Note that every transfer
// rate may not be supported by every target due to either hardware or software
// limitations. By far, the most commonly-supported rate is USB 1.1 Full-Speed
// (12 Mbit/sec). If unsure, either use FullSpeed or leave it undefined and let
// the driver use the default speed for your target.
const (
	LowSpeed       Speed = iota + 1 // 1.5 Mbit/sec (USB 1.0)
	FullSpeed                       //  12 Mbit/sec (USB 1.1)
	HighSpeed                       // 480 Mbit/sec (USB 2.0)
	SuperSpeed                      //   5 Gbit/sec (USB 3.0)
	DualSuperSpeed                  //  10 Gbit/sec (USB 3.1, Dual-Lane SS)
)

// initCore initializes a free USB core with given operating mode on the USB
// port at given index, if available. Returns a reference to the initialized
// core or nil if the core is unavailable.
func initCore(port int, speed Speed, class class) (*core, status) {

	if port < 0 || port >= CoreCount || 0 == class.config {
		return nil, statusInvalid
	}

	if modeIdle != coreInstance[port].mode {
		// Check if requested port is already configured as requested class. If so,
		// just return a reference to the existing core instead of an error.
		//
		// This will allow, for instance, TinyGo examples that try to reconfigure
		// the USB (CDC-ACM) UART port (which is already configured by the runtime)
		// to continue without error.
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
		dc, st := initDCD(port, speed, class)
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
		dc.enable(true) // Enable interrupts and enter runtime

	case modeHost:
		// Allocate a free host controller and install interrupts
		hc, st := initHCD(port, speed, class)
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
		hc.enable(true) // Enable interrupts and enter runtime

	default:
		return nil, statusInvalid
	}

	return &coreInstance[port], statusOK
}

// class represents the type of a host/device and its class configuration index.
// The first valid configuration index is 1. Index 0 is reserved and invalid.
type class struct {
	id     int
	config int
}

// Enumerated constants for all supported host/device class configurations.
const (
	classDeviceCDC = 0
	classDeviceHID = 1
)

// mode returns the USB core operating mode of the receiver class c.
//go:inline
func (c class) mode() int {
	switch c.id {
	case classDeviceCDC, classDeviceHID:
		return modeDevice
	default:
		return modeIdle
	}
}

// equals returns true if and only if all fields of the given class are equal to
// those of the receiver c.
//go:inline
func (c class) equals(class class) bool {
	return c.id == class.id && c.config == class.config
}

// status represents the return code of a subroutine.
type status uint8

// Constant definitions for all status codes used within the package.
const (
	statusOK      status = iota // Success
	statusBusy                  // Busy
	statusInvalid               // Invalid argument
	statusFail                  // Failure
)

// ok returns true if and only if the receiver st equals statusOK.
//go:inline
func (s status) ok() bool { return statusOK == s }
