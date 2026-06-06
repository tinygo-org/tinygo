//go:build uefi

package uefi

import "unsafe"

// Memory allocation types for AllocatePages
type AllocateType uint32

const (
	AllocateAnyPages AllocateType = iota
	AllocateMaxAddress
	AllocateAddress
)

// Memory types for AllocatePages
type MemoryType uint32

const (
	EfiReservedMemoryType MemoryType = iota
	EfiLoaderCode
	EfiLoaderData
	EfiBootServicesCode
	EfiBootServicesData
	EfiRuntimeServicesCode
	EfiRuntimeServicesData
	EfiConventionalMemory
	EfiUnusableMemory
	EfiACPIReclaimMemory
	EfiACPIMemoryNVS
	EfiMemoryMappedIO
	EfiMemoryMappedIOPortSpace
	EfiPalCode
	EfiPersistentMemory
	EfiMaxMemoryType
)

// PageSize is the UEFI page size (4KB)
const PageSize = 4096

// AllocatePages allocates memory pages from UEFI.
// Returns the physical address of the allocated memory, or 0 on failure.
// For AllocateAddress, pass the desired address as addr.
func AllocatePages(allocType AllocateType, memType MemoryType, pages uintptr, addr *uintptr) uintptr {
	st := GetSystemTable()
	if st == nil || st.BootServices == nil || st.BootServices.AllocatePages == 0 {
		return 0
	}

	status := Call(
		st.BootServices.AllocatePages,
		uintptr(allocType),
		uintptr(memType),
		pages,
		uintptr(unsafe.Pointer(addr)),
	)
	if status != Success {
		return 0
	}

	return *addr
}

// FreePages frees memory pages previously allocated with AllocatePages.
func FreePages(memory uintptr, pages uintptr) Status {
	st := GetSystemTable()
	if st == nil || st.BootServices == nil || st.BootServices.FreePages == 0 {
		return ErrUnsupported
	}
	return Call(st.BootServices.FreePages, memory, pages)
}

// MemoryDescriptor describes a region in the UEFI memory map.
// Note: the actual descriptor size returned by GetMemoryMap may be larger
// than this struct due to firmware extensions. Always use the returned
// descriptorSize to iterate.
type MemoryDescriptor struct {
	Type          uint32
	_             uint32 // padding
	PhysicalStart uint64
	VirtualStart  uint64
	NumberOfPages uint64
	Attribute     uint64
	_             uint64 // padding
}

// GetMemoryMap retrieves the current UEFI memory map.
// Returns the number of entries for iteration.
// Use MemMapEntry(buf, i, descSize) to access entries.
func GetMemoryMap(memMapBuffer []byte, memMapSize *uintptr, memDescSize *uintptr) int {
	st := GetSystemTable()
	if st == nil || st.BootServices == nil || st.BootServices.GetMemoryMap == 0 {
		return 0
	}

	status := Call(st.BootServices.GetMemoryMap,
		uintptr(unsafe.Pointer(memMapSize)),
		uintptr(unsafe.Pointer(&memMapBuffer[0])),
		uintptr(0),
		uintptr(unsafe.Pointer(memDescSize)),
		uintptr(0),
	)
	if status != Success {
		return 0
	} else if *memDescSize < unsafe.Sizeof(MemoryDescriptor{}) {
		return 0
	}

	return int(*memMapSize) / int(*memDescSize)
}

// MemMapEntry returns the i-th memory descriptor from the last GetMemoryMap call.
func MemMapEntry(memMapBuffer []byte, i int, descSize uintptr) *MemoryDescriptor {
	base := unsafe.Pointer(&memMapBuffer[0])
	return (*MemoryDescriptor)(unsafe.Add(base, i*int(descSize)))
}
