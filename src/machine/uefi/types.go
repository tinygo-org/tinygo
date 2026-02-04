//go:build uefi

package uefi

// Handle is an opaque UEFI handle type.
type Handle uintptr

// Status represents a UEFI status code.
type Status uintptr

// UEFI status codes
const (
	Success Status = 0

	// Error codes (high bit set)
	ErrLoadError        Status = 0x8000000000000001
	ErrInvalidParameter Status = 0x8000000000000002
	ErrUnsupported      Status = 0x8000000000000003
	ErrBadBufferSize    Status = 0x8000000000000004
	ErrBufferTooSmall   Status = 0x8000000000000005
	ErrNotReady         Status = 0x8000000000000006
	ErrDeviceError      Status = 0x8000000000000007
	ErrWriteProtected   Status = 0x8000000000000008
	ErrOutOfResources   Status = 0x8000000000000009
	ErrNotFound         Status = 0x800000000000000E
)

// TableHeader is the standard UEFI table header.
type TableHeader struct {
	Signature  uint64
	Revision   uint32
	HeaderSize uint32
	CRC32      uint32
	Reserved   uint32
}

// GUID represents a UEFI Globally Unique Identifier.
type GUID struct {
	Data1 uint32
	Data2 uint16
	Data3 uint16
	Data4 [8]uint8
}
