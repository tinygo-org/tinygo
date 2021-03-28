package usb

// hostController provides hardware abstraction over a platform's physical
// USB host port controller (e.g., an EHCI-compliant peripheral).
//
// To connect the USB 2.0 host driver stack to a particular platform, the
// following method must be implemented:
//
//		func (h *host) initController() hostController
//
// This method shall return an implementation of hostController, or nil if a
// controller cannot be allocated for the given port.
type hostController interface{}
