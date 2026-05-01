package uefi

type UINTN uintptr
type EFI_STATUS UINTN
type EFI_TPL UINTN
type EFI_HANDLE uintptr
type EFI_EVENT uintptr
type EFI_PHYSICAL_ADDRESS uint64

type CHAR16 uint16
type BOOLEAN bool
type VOID byte

type EFI_TABLE_HEADER struct {
	Signature  uint64
	Revision   uint32
	HeaderSize uint32
	CRC32      uint32
	Reserved   uint32
}

type EFI_ALLOCATE_TYPE int

const (
	AllocateAnyPages EFI_ALLOCATE_TYPE = iota
	AllocateMaxAddress
	AllocateAddress
)

type EFI_MEMORY_TYPE int

const (
	EfiReservedMemoryType EFI_MEMORY_TYPE = iota
	EfiLoaderCode
	EfiLoaderData
	EfiBootServicesCode
	EfiBootServicesData
	EfiRuntimeServicesCode
	EfiRuntimeServicesData
	EfiConventionalMemory
)

type EVENT_TYPE uint32

const (
	EVT_TIMER EVENT_TYPE = 0x80000000
)

const (
	TPL_CALLBACK EFI_TPL = 8
)

type EFI_TIMER_DELAY int

const (
	TimerCancel EFI_TIMER_DELAY = iota
	TimerPeriodic
	TimerRelative
)
