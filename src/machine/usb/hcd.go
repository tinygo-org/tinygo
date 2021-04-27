package usb

// Implementation of 32-bit target-agnostic USB host controller driver (hcd).

// hcdCount defines the number of USB cores to configure for host mode. It is
// computed as the sum of all declared host configuration descriptors.
const hcdCount = 0 // + ...

// hcdInstance provides statically-allocated instances of each USB host
// controller configured on this platform.
var hcdInstance [hcdCount]hcd

// hhwInstance provides statically-allocated instances of each USB hardware
// abstraction for ports configured as host on this platform.
var hhwInstance [hcdCount]hhw

// hcd implements USB host controller driver (hcd) interface.
type hcd struct {
	*hhw // USB hardware abstraction layer

	core *core // Parent USB core this instance is attached to
	port int   // USB port index
	cc   class // USB host class
	id   int   // USB host controller index
}

// initHCD initializes and assigns a free host controller instance to the given
// USB port. Returns the initialized host controller or nil if no free host
// controller instances remain.
func initHCD(port int, class class) (*hcd, status) {
	if 0 == hcdCount {
		return nil, statusInvalid // Must have defined host controllers
	}
	switch class.id {
	default:
	}
	// Return the first instance whose assigned core is currently nil.
	for i := range hcdInstance {
		if nil == hcdInstance[i].core {
			// Initialize host controller.
			hcdInstance[i].hhw = allocHHW(port, i, &hcdInstance[i])
			hcdInstance[i].core = &coreInstance[port]
			hcdInstance[i].port = port
			hcdInstance[i].cc = class
			hcdInstance[i].id = i
			return &hcdInstance[i], statusOK
		}
	}
	return nil, statusBusy // No free host controller instances available.
}

// class returns the receiver's current host class configuration.
func (h *hcd) class() class { return h.cc }
