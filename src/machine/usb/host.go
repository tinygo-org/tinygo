package usb

// Unexported USB device type definitions.
// In device mode, a USB device represents the whole USB port controller.
// In host mode, a USB device represents a peripheral device connected to the
// USB port controller.
type (
	host struct {
		port uint8
	}
)

func (h *host) init(port uint8) (s status) {

	// initialize host
	h.port = port

	return statusSuccess
}

func (h *host) deinit() status {

	// de-initialize host
	return statusSuccess
}
