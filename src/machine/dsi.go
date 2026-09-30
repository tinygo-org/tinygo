//go:build esp32s3

// Hardware abstraction layer for the MIPI DSI (Display Serial Interface) peripheral.

package machine

import "errors"

var (
	errDSINotImplemented   = errors.New("dsi: operation not implemented")
	errDSIFramebufferError = errors.New("dsi: framebuffer error")
)

// VideoTimings configures the display controller's timing generator.
// This is a rather generic type, so it is kept in a common file.
type VideoTimings struct {
	PCLK        uint32 // Pixel clock frequency (Hz)
	HActive     uint16 // Active horizontal resolution (pixels)
	HFrontPorch uint16 // Horizontal front porch (cycles/pixels)
	HSyncWidth  uint16 // Horizontal sync pulse width
	HBackPorch  uint16 // Horizontal back porch
	VActive     uint16 // Active vertical resolution (lines)
	VFrontPorch uint16 // Vertical front porch (lines)
	VSyncWidth  uint16 // Vertical sync pulse width
	VBackPorch  uint16 // Vertical back porch
}

// PixelFormat defines the RGB pixel encoding on the LCD data bus.
type PixelFormat uint8

// We define two most common DSI pixel formats supported by most displays and DSI host hardwares.
const (
	PixelFormatRGB565 PixelFormat = iota // 5 bits for red, 6 bits for green, and 5 bits for blue (DSI Data Type 0x0E)
	PixelFormatRGB888                    // Standard 24-bit RGB with 8 bits per component (DSI Data Type 0x3E)
)

// Hardware-specific RGBPinConfig describes the physical GPIO-to-LCD-signal wiring.
type RGBPinConfig struct {
	DE    Pin     // LCD data enable
	VSYNC Pin     // LCD vertical sync
	HSYNC Pin     // LCD horizontal sync
	PCLK  Pin     // LCD pixel clock
	Data  [24]Pin // LCD data signals D0..D23 (GPIO pin number each)
}

// Hardware-specific lanes/pins configuration
type LaneConfig struct {
	PixelFormat PixelFormat
	Pins        RGBPinConfig
}

// DSIController handles Control / Command Mode (Panel Configuration).
type DSIController interface {
	// WriteShort sends a 4-byte MIPI DSI Short Packet (Header only, no payload).
	WriteShort(dt uint8, data0, data1 uint8) error

	// WriteLong sends a MIPI DSI Long Packet with a payload buffer.
	WriteLong(dt uint8, payload []byte) error

	// ReadShort initiates a Bus Turnaround (BTA) to read back short response packets.
	ReadShort(dt uint8) (data [2]byte, err error)
}

// DSIStreamer handles Video Pipeline and Framebuffer Management.
type DSIStreamer interface {
	// InitVideo configures D-PHY clocking, active lanes, and video timing parameters.
	InitVideo(lanes LaneConfig, timings VideoTimings) error

	// StartVideo starts continuous DMA transmission over D-PHY lanes.
	StartVideo() error

	// StopVideo halts video packet generation (e.g., prior to panel sleep mode).
	StopVideo() error

	// SetFramebuffer assigns the primary buffer for continuous DMA reading. Must be called prior to StartVideo().
	SetFramebuffer(primary []byte) error

	// SwapFramebuffer queues a new buffer address for the DMA engine.
	// Blocks until the next V-Blanking interval.
	SwapFramebuffer(next []byte) error
}

// If you are getting a compile error on this line please check to see you've correctly implemented the methods
// on the DSI type. They must match the DSIController and DSIStreamer interface method signatures.
// Not all hardwares implement the DSIController, in which case implement stubs that panic() or return an ErrUnimplemented.
var _ DSIController = (*DSI)(nil)
var _ DSIStreamer = (*DSI)(nil)
