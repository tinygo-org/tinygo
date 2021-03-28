package usb

type (
	// status is the common error code used internally for package operations.
	status int8
	// mode defines the operating mode of a USB port.
	mode int8

	// port represents a physical USB port, which may be configured as either a
	// device or as a host.
	port struct {
		mode   mode
		device device
		host   host
	}
)

// Constants for unexported types shared across entire package.
const (
	statusSuccess            status = iota // Success
	statusError                            // Failed
	statusBusy                             // Busy
	statusInvalidHandle                    // Invalid handle
	statusInvalidParameter                 // Invalid parameter
	statusInvalidRequest                   // Invalid request
	statusControllerNotFound               // Controller cannot be found
	statusInvalidController                // Invalid controller interface
	statusNotSupported                     // Configuration is not supported
	statusRetry                            // Enumeration get configuration retry
	statusTransferStall                    // Transfer stalled
	statusTransferFailed                   // Transfer failed
	statusAllocFail                        // Allocation failed
	statusLackSwapBuffer                   // Insufficient swap buffer for KHCI
	statusTransferCancel                   // The transfer cancelled
	statusBandwidthFail                    // Allocate bandwidth failed
	statusMSDStatusFail                    // For MSD, the CSW status means fail
	statusEHCIAttached                     // EHCI attached
	statusEHCIDetached                     // EHCI detached
	statusDataOverRun                      // Endpoint data (Rx) exceeds max size
	statusNotImplemented                   // Supported feature not implemented

	modeIdle   mode = iota // USB core idle (unallocated)
	modeDevice             // USB device mode
	modeHost               // USB host mode
)

var (
	// portInstance holds instances for all available ports on the platform.
	portInstance [ConfigPortCount]port
)

func init() {
	// ensure all ports are in idle state by default
	for i := range portInstance {
		portInstance[i].mode = modeIdle
	}
}

func ProcessMessages() (err error) {
	for i := range portInstance {
		s := portInstance[i].process()
		if !s.OK() && nil == err {
			err = s
		}
	}
	return
}

// initPort configures the mode for a given USB port and initializes the
// hardware's port controller. If the port is invalid or not idle (it has
// already been configured), it returns nil and a status code.
func initPort(port uint8, mode mode) (*port, status) {
	if port >= ConfigPortCount || int(port) >= len(portInstance) {
		return nil, statusInvalidController
	}
	if modeIdle != portInstance[port].mode {
		return nil, statusBusy
	}
	portInstance[port].mode = mode
	switch mode {
	case modeDevice:
		if s := portInstance[port].device.init(port); !s.OK() {
			return nil, s
		}
	case modeHost:
		if s := portInstance[port].host.init(port); !s.OK() {
			return nil, s
		}
	}
	return &portInstance[port], statusSuccess
}

// deinit disables the receiver USB port, changing its mode to idle, freeing it
// for reuse or reconfiguration.
func (p *port) deinit() status {
	switch p.mode {
	case modeDevice:
		return p.device.deinit()
	case modeHost:
		return p.host.deinit()
	default:
		return statusInvalidController
	}
}

func (p *port) process() status {
	switch p.mode {
	case modeDevice:
		return p.device.controller.process()
	case modeHost:
		return statusSuccess // TODO: not implemented
	default:
		return statusInvalidController
	}
}

// initCDCACM applies a CDC-ACM configuration to the receiver port p and then
// returns the configured deviceClassDriver and deviceClass that were assigned
// to the receiver.
//
// The given deviceClassEventHandler is called for any USB device-level event
// notifications received, which allows an upper-layer CDC-ACM driver (such as
// a UART interface implementation) the opportunity to handle device events.
func (p *port) initCDCACM(id uint8, handler deviceClassEventHandler) (*deviceCDCACM, *deviceClass) {

	// verify a valid port was provided
	if nil == p || nil == p.device.controller || p.mode != modeDevice {
		return nil, nil
	}

	if 0 == id || int(id) > len(configDeviceCDCACM[p.device.port]) {
		return nil, nil
	}

	// get a reference to each of the class interfaces
	comm := &deviceCDCACMConfigInstance[p.device.port][id-1].info.interfaceList[0]
	data := &deviceCDCACMConfigInstance[p.device.port][id-1].info.interfaceList[1]

	// configDeviceCDCACM must be defined per package API. these settings will be
	// platform-specific, and will probably be implemented in a build tag-
	// constrained source file. the length of this array corresponds to the number
	// of USB CDC-ACM ports that are being created, and the index of each element
	// corresponds to the physical USB port (core index). Each element is a slice
	// of alternate device configurations that may be selected for a given port.

	// CDC-ACM Communication/control interface
	comm.interfaceNumber =
		configDeviceCDCACM[p.device.port][id-1].commInterfaceIndex

	comm.deviceInterface[0].endpoint[0].address =
		configDeviceCDCACM[p.device.port][id-1].commInterruptInEndpoint |
			specDescriptorEndpointAddressDirectionIn

	comm.deviceInterface[0].endpoint[0].maxPacketSize =
		configDeviceCDCACM[p.device.port][id-1].commInterruptInPacketSize

	comm.deviceInterface[0].endpoint[0].interval =
		configDeviceCDCACM[p.device.port][id-1].commInterruptInInterval

	// CDC-ACM Data interface
	data.interfaceNumber =
		configDeviceCDCACM[p.device.port][id-1].dataInterfaceIndex

	data.deviceInterface[0].endpoint[0].address =
		configDeviceCDCACM[p.device.port][id-1].dataBulkInEndpoint |
			specDescriptorEndpointAddressDirectionIn

	data.deviceInterface[0].endpoint[0].maxPacketSize =
		configDeviceCDCACM[p.device.port][id-1].dataBulkInPacketSize

	data.deviceInterface[0].endpoint[1].address =
		configDeviceCDCACM[p.device.port][id-1].dataBulkOutEndpoint |
			specDescriptorEndpointAddressDirectionOut

	data.deviceInterface[0].endpoint[1].maxPacketSize =
		configDeviceCDCACM[p.device.port][id-1].dataBulkOutPacketSize

	// assign our configured CDC-ACM class to the receiver's device and call its
	// class initialization routine(s).
	cls := p.device.initClass(id, deviceCDCACMConfigInstance[p.device.port], handler)
	acm := cls.config[0].driver.(*deviceCDCACM)

	return acm, cls
}

// OK returns true if and only if the receiver s is equal to statusSuccess.
//go:inline
func (s status) OK() bool { return statusSuccess == s }

// Error returns a simple descriptive error string of the receiver s.
func (s status) Error() string {
	switch s {
	case statusSuccess:
		return ""
	case statusError:
		return "failed"
	case statusBusy:
		return "busy"
	case statusInvalidHandle:
		return "invalid handle"
	case statusInvalidParameter:
		return "invalid parameter"
	case statusInvalidRequest:
		return "invalid request"
	case statusControllerNotFound:
		return "controller not found"
	case statusInvalidController:
		return "invalid controller interface"
	case statusNotSupported:
		return "configuration not supported"
	case statusRetry:
		return "retry enumeration"
	case statusTransferStall:
		return "transfer stalled"
	case statusTransferFailed:
		return "transfer failed"
	case statusAllocFail:
		return "allocation failed"
	case statusLackSwapBuffer:
		return "insufficient swap buffer"
	case statusTransferCancel:
		return "transfer cancelled"
	case statusBandwidthFail:
		return "bandwidth allocation failed"
	case statusMSDStatusFail:
		return "mass-storage device failed"
	case statusEHCIAttached:
		return "host attached"
	case statusEHCIDetached:
		return "host detached"
	case statusDataOverRun:
		return "data overrun"
	case statusNotImplemented:
		return "feature not implemented"
	default:
		return "unknown error"
	}
}

// leU64 returns a slice containing 8 bytes from the given uint64 u.
//
// The returned bytes have little-endian ordering; that is, the first element
// at index 0 is the least-significant byte in u and index 7 is the most-
// significant byte.
//go:inline
func leU64(u uint64) []uint8 {
	if u == 0 {
		// skip all processing for the common case (u = 0)
		return []uint8{0, 0, 0, 0, 0, 0, 0, 0}
	}
	return []uint8{
		uint8(u), uint8(u >> 8), uint8(u >> 16), uint8(u >> 24),
		uint8(u >> 32), uint8(u >> 40), uint8(u >> 48), uint8(u >> 56),
	}
}

// leU32 returns a slice containing 4 bytes from the given uint32 u.
//
// The returned bytes have little-endian ordering; that is, the first element
// at index 0 is the least-significant byte in u and index 3 is the most-
// significant byte.
//go:inline
func leU32(u uint32) []uint8 {
	if u == 0 {
		// skip all processing for the common case (u = 0)
		return []uint8{0, 0, 0, 0}
	}
	return []uint8{
		uint8(u), uint8(u >> 8), uint8(u >> 16), uint8(u >> 24),
	}
}

// leU16 returns a slice containing 2 bytes from the given uint16 u.
//
// The returned bytes have little-endian ordering; that is, the first element
// at index 0 is the least-significant byte in u and index 1 is the most-
// significant byte.
//go:inline
func leU16(u uint16) []uint8 {
	if u == 0 {
		// skip all processing for the common case (u = 0)
		return []uint8{0, 0}
	}
	return []uint8{
		uint8(u), uint8(u >> 8),
	}
}

// beU64 returns a slice containing 8 bytes from the given uint64 u.
//
// The returned bytes have big-endian ordering; that is, the first element at
// index 0 is the most-significant byte in u and index 7 is the least-
// significant byte.
//go:inline
func beU64(u uint64) []uint8 {
	if u == 0 {
		// skip all processing for the common case (u = 0)
		return []uint8{0, 0, 0, 0, 0, 0, 0, 0}
	}
	return []uint8{
		uint8(u >> 56), uint8(u >> 48), uint8(u >> 40), uint8(u >> 32),
		uint8(u >> 24), uint8(u >> 16), uint8(u >> 8), uint8(u),
	}
}

// beU32 returns a slice containing 4 bytes from the given uint32 u.
//
// The returned bytes have big-endian ordering; that is, the first element at
// index 0 is the most-significant byte in u and index 3 is the least-
// significant byte.
//go:inline
func beU32(u uint32) []uint8 {
	if u == 0 {
		// skip all processing for the common case (u = 0)
		return []uint8{0, 0, 0, 0}
	}
	return []uint8{
		uint8(u >> 24), uint8(u >> 16), uint8(u >> 8), uint8(u),
	}
}

// beU16 returns a slice containing 2 bytes from the given uint16 u.
//
// The returned bytes have big-endian ordering; that is, the first element at
// index 0 is the most-significant byte in u and index 1 is the least-
// significant byte.
//go:inline
func beU16(u uint16) []uint8 {
	if u == 0 {
		// skip all processing for the common case (u = 0)
		return []uint8{0, 0}
	}
	return []uint8{
		uint8(u >> 8), uint8(u),
	}
}

// revU64 returns the given uint64 u with bytes in the reverse order.
//go:inline
func revU64(u uint64) uint64 {
	if u == 0 {
		// skip all processing for the common case (u = 0)
		return 0
	}
	return ((u & 0x00000000000000FF) << 56) |
		((u & 0x000000000000FF00) << 40) |
		((u & 0x0000000000FF0000) << 24) |
		((u & 0x00000000FF000000) << 8) |
		((u & 0x000000FF00000000) >> 8) |
		((u & 0x0000FF0000000000) >> 24) |
		((u & 0x00FF000000000000) >> 40) |
		((u & 0xFF00000000000000) >> 56)
}

// revU32 returns the given uint32 u with bytes in the reverse order.
//go:inline
func revU32(u uint32) uint32 {
	if u == 0 {
		// skip all processing for the common case (u = 0)
		return 0
	}
	return ((u & 0x000000FF) << 24) | ((u & 0x0000FF00) << 8) |
		((u & 0x00FF0000) >> 8) | ((u & 0xFF000000) >> 24)
}

// revU16 returns the given uint16 u with bytes in the reverse order.
//go:inline
func revU16(u uint16) uint16 {
	if u == 0 {
		// skip all processing for the common case (u = 0)
		return 0
	}
	return ((u & 0x00FF) << 8) | ((u & 0xFF00) >> 8)
}

// packU64 returns a uint64 constructed by concatenating the bytes in slice b.
//
// The least-significant byte in the returned value is the first element at
// index 0 in b and the most significant byte is index 7, if given. If fewer
// than 8 elements are given in b, the corresponding bytes in the returned value
// are all 0.
//go:inline
func packU64(b []uint8) (u uint64) {
	for i := 0; i < 8 && i < len(b); i++ {
		u |= uint64(b[i]) << (i * 8)
	}
	return
}

// packU32 returns a uint32 constructed by concatenating the bytes in slice b.
//
// The least-significant byte in the returned value is the first element at
// index 0 in b and the most significant byte is index 3, if given. If fewer
// than 4 elements are given in b, the corresponding bytes in the returned value
// are all 0.
//go:inline
func packU32(b []uint8) (u uint32) {
	for i := 0; i < 4 && i < len(b); i++ {
		u |= uint32(b[i]) << (i * 8)
	}
	return
}

// packU16 returns a uint16 constructed by concatenating the bytes in slice b.
//
// The least-significant byte in the returned value is the first element at
// index 0 in b and the most significant byte is index 1, if given. If fewer
// than 2 elements are given in b, the corresponding bytes in the returned value
// are all 0.
//go:inline
func packU16(b []uint8) (u uint16) {
	for i := 0; i < 2 && i < len(b); i++ {
		u |= uint16(b[i]) << (i * 8)
	}
	return
}
