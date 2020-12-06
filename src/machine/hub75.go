package machine

// Hub75 extends the TinyGo 'machine' package API to support machine-specific
// GPIO operations that are required to efficiently utilize a HUB75 interface
// commonly found on RGB LED matrix panels.
//
// In particular, these methods provide the consumer an ability to set/clear
// multiple HUB75 data/control Pins on a single GPIO port simultaneosly, which
// is necessary to implement bit-banged drivers with sufficient performance.
type Hub75 interface {

	// SetRgb sets/clears each of the 6 RGB data pins.
	SetRgb(bool, bool, bool, bool, bool, bool) // R1, G1, B1, R2, G2, B2

	// SetRgbMask sets/clears each of the 6 RGB data pins from the given bitmask.
	SetRgbMask(uint32)

	// SetRow sets the active pair of data rows with the given index.
	SetRow(int)

	// HoldClkLow sets and holds the clock line low for sufficient time when
	// transferring one bit of data.
	HoldClkLow()

	// HoldClkHigh sets and holds the clock line high for sufficient time when
	// transferring one bit of data.
	HoldClkHigh()

	// HoldLatLow sets and holds the latch line low for sufficient time to stop
	// signalling all shift registers from outputting their content.
	HoldLatLow()

	// HoldLatHigh sets and holds the latch line high for sufficient time to start
	// signalling all shift registers to begin outputting their content.
	HoldLatHigh()

	// GetPinGroupAlignment returns true if and only if all given Pins are on the
	// same GPIO port, and returns the minimum size of the group to which all pins
	// belong (8, 16, or 32 if true, otherwise 0). Returns (true, 0) if no Pins
	// are provided.
	GetPinGroupAlignment(...Pin) (bool, uint8)

	// InitTimer is used to initialize a timer service that fires an interrupt at
	// regular frequency, which is used to signal row data transmission with the
	// given interrupt service routine (ISR). The timer does not begin raising
	// interrupts until StartTimer is called.
	InitTimer(func())

	// ResumeTimer resumes the timer service, with given current value, that
	// signals row data transmission for HUB75 by raising interrupts with given
	// periodicity.
	ResumeTimer(int, int)

	// PauseTimer pauses the timer service that signals row data transmission for
	// HUB75 and returns the current value of the timer.
	PauseTimer() int
}
