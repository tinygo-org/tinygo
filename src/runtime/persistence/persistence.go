// Package persistence provides persistent storage across low-power
// and reset states.
//
// The amount of space available is under the control of the application
// logic.
//
// For SRAM and built-in flash a single allocation can be made per
// application (usually in the 'main' code) per persistence type.  I.e.
// RAM and Flash are allocated separately so one application may have
// both a RAM region and a Flash region, but not more than one of each.
//
// Libraries should NOT allocate persistence regions, but should accept
// persistence regions passed in - and may indicate how much space is
// required using public constants (for example).
//
// The built-in regions exposed are deliberately raw - there is no
// native functionality to detect an uninitialized region, measures
// to detect corruption, or functionality to implement transactions.
// These are likely very application specific and left to the consumer
// and/or external libraries.
//
package persistence

// Region enables access to a persistence region.
//
// Alternate implementations of Region provide access to different
// region types (RAM, SRAM, others).  External implementations
// could provide access to battery backed RAM, or SPI flash
// consistent with this interface.
//
type Region interface {
	// Size gets the size of the persistence area (in bytes)
	Size() int64

	// ReadAt reads from the persistence area, it is signature
	// compatible with Go's `io.ReaderAt` interface
	ReadAt(dest []byte, offset int64) (n int, err error)

	// WriterAt reads from the persistence area, it is signature
	// compatible with Go's `io.WriterAt` interface
	WriteAt(src []byte, offset int64) (n int, err error)

	// SubAllocation creates a new persistence object that
	// represents a sub-allocation of the main persistence
	// area
	SubAllocation(offset int, len int) (p Region, err error)
}

// NewRAM is a compiler intrinsic that allocates persistence space in
// RAM.
//
// The state of this area will be preserved in low power states where
// power to RAM is maintained, even when the CPU core and peripherals
// are unpowered, resetting their state.
//
// The size passed in must be a constant value.
//
func NewRAM(size uint32) Region { return GetRAM() }

// NewFlash is a compiler intrinsic that allocates persistence space in
// Flash.
//
// The state of this area will be preserved across hardware reboots.
// Care should be taken not to write excessively to flash, it is best
// used for configuration settings / similar that must be preserved
// when power is removed.
//
// The size passed in must be a constant value.
//
func NewFlash(size uint32) Region
