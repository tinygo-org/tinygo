package uefi

import (
	"unsafe"
)

// EFI_SPECIFICATION_REVISION_MAJORMINOR combines the major and minor revision into a single value.
// @param major Major revision
// @param minor Minor revision
// @return Combined major and minor revision
func EFI_SPECIFICATION_REVISION_MAJORMINOR(major, minor uint16) uint32 {
	return (uint32(major) << 16) | uint32(minor)
}

// EFI_SPECIFICATION_MAJOR_REVISION represents the major revision of the EFI Specification.
const EFI_SPECIFICATION_MAJOR_REVISION uint16 = 1

// EFI_SPECIFICATION_MINOR_REVISION represents the minor revision of the EFI Specification.
const EFI_SPECIFICATION_MINOR_REVISION uint16 = 2

// EFI_SPECIFICATION_VERSION represents the combined major and minor revision of the EFI Specification.
var EFI_SPECIFICATION_VERSION = EFI_SPECIFICATION_REVISION_MAJORMINOR(EFI_SPECIFICATION_MAJOR_REVISION, EFI_SPECIFICATION_MINOR_REVISION)

// EFI_TABLE_HEADER
// Standard EFI table header
type EFI_TABLE_HEADER struct {
	Signature  uint64
	Revision   uint32
	HeaderSize uint32
	CRC32      uint32
	Reserved   uint32
}

type EFI_PHYSICAL_ADDRESS uint64
type EFI_VIRTUAL_ADDRESS uint64

type EFI_ALLOCATE_TYPE int

const (
	AllocateAnyPages EFI_ALLOCATE_TYPE = iota
	AllocateMaxAddress
	AllocateAddress
	MaxAllocateType
)

// EFI_MEMORY_TYPE is an enumeration of memory types.
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
	EfiUnusableMemory
	EfiACPIReclaimMemory
	EfiACPIMemoryNVS
	EfiMemoryMappedIO
	EfiMemoryMappedIOPortSpace
	EfiPalCode
	EfiPersistentMemory
	EfiUnacceptedMemoryType
	EfiMaxMemoryType
)

// Memory cacheability attribute
const (
	EFI_MEMORY_UC  uint64 = 0x0000000000000001
	EFI_MEMORY_WC  uint64 = 0x0000000000000002
	EFI_MEMORY_WT  uint64 = 0x0000000000000004
	EFI_MEMORY_WB  uint64 = 0x0000000000000008
	EFI_MEMORY_UCE uint64 = 0x0000000000000010
)

// Physical memory protection attribute
const (
	EFI_MEMORY_WP uint64 = 0x0000000000001000
	EFI_MEMORY_RP uint64 = 0x0000000000002000
	EFI_MEMORY_XP uint64 = 0x0000000000004000
	EFI_MEMORY_RO uint64 = 0x0000000000020000
)

// Runtime memory attribute
const (
	EFI_MEMORY_NV      uint64 = 0x0000000000008000
	EFI_MEMORY_RUNTIME uint64 = 0x8000000000000000
)

// Other memory attribute
const (
	EFI_MEMORY_MORE_RELIABLE uint64 = 0x0000000000010000
	EFI_MEMORY_SP            uint64 = 0x0000000000040000
	EFI_MEMORY_CPU_CRYPTO    uint64 = 0x0000000000080000
	EFI_MEMORY_ISA_VALID     uint64 = 0x4000000000000000
	EFI_MEMORY_ISA_MASK      uint64 = 0x0FFFF00000000000
)

const EFI_MEMORY_DESCRIPTOR_VERSION = 1

// EFI_MEMORY_DESCRIPTOR
// Definition of an EFI memory descriptor.
type EFI_MEMORY_DESCRIPTOR struct {
	Type          uint32
	PhysicalStart EFI_PHYSICAL_ADDRESS
	VirtualStart  EFI_VIRTUAL_ADDRESS
	NumberOfPages uint64
	Attribute     uint64
}

// EFI_TIME_CAPABILITIES
// This provides the capabilities of the
// real time clock device as exposed through the EFI interfaces.
type EFI_TIME_CAPABILITIES struct {
	Resolution uint32
	Accuracy   uint32
	SetsToZero BOOLEAN
}

// EFI_OPEN_PROTOCOL_INFORMATION_ENTRY
// EFI Oprn Protocol Information Entry
type EFI_OPEN_PROTOCOL_INFORMATION_ENTRY struct {
	AgentHandle      EFI_HANDLE
	ControllerHandle EFI_HANDLE
	Attributes       uint32
	OpenCount        uint32
}

// EFI_CAPSULE_BLOCK_DESCRIPTOR
// EFI Capsule Block Descriptor
type EFI_CAPSULE_BLOCK_DESCRIPTOR struct {
	Length              uint64
	DataBlock           EFI_PHYSICAL_ADDRESS
	ContinuationPointer EFI_PHYSICAL_ADDRESS
}

// EFI_CAPSULE_HEADER
// EFI Capsule Header.
type EFI_CAPSULE_HEADER struct {
	CapsuleGuid      EFI_GUID
	HeaderSize       uint32
	Flags            uint32
	CapsuleImageSize uint32
}

// EFI_CAPSULE_TABLE
// The EFI System Table entry must point to an array of capsules
// that contain the same CapsuleGuid value. The array must be
// prefixed by a UINT32 that represents the size of the array of capsules.
type EFI_CAPSULE_TABLE struct {
	CapsuleArrayNumber uint32
}

type EFI_LOCATE_SEARCH_TYPE int

const (
	AllHandles       EFI_LOCATE_SEARCH_TYPE = 0
	ByRegisterNotify EFI_LOCATE_SEARCH_TYPE = 1
	ByProtocol       EFI_LOCATE_SEARCH_TYPE = 2
)

// region: EFI Runtime Services Table

type EFI_RESET_TYPE int

const (
	EfiResetCold     EFI_RESET_TYPE = 0
	EfiResetWarm     EFI_RESET_TYPE = 1
	EfiResetShutdown EFI_RESET_TYPE = 2
)

const EFI_RUNTIME_SERVICES_SIGNATURE = 0x56524553544e5552

var (
	EFI_1_02_RUNTIME_SERVICES_REVISION  = EFI_SPECIFICATION_REVISION_MAJORMINOR(1, 02)
	EFI_1_10_RUNTIME_SERVICES_REVISION  = EFI_SPECIFICATION_REVISION_MAJORMINOR(1, 10)
	EFI_2_00_RUNTIME_SERVICES_REVISION  = EFI_SPECIFICATION_REVISION_MAJORMINOR(2, 00)
	EFI_2_10_RUNTIME_SERVICES_REVISION  = EFI_SPECIFICATION_REVISION_MAJORMINOR(2, 10)
	EFI_2_20_RUNTIME_SERVICES_REVISION  = EFI_SPECIFICATION_REVISION_MAJORMINOR(2, 20)
	EFI_2_30_RUNTIME_SERVICES_REVISION  = EFI_SPECIFICATION_REVISION_MAJORMINOR(2, 30)
	EFI_2_31_RUNTIME_SERVICES_REVISION  = EFI_SPECIFICATION_REVISION_MAJORMINOR(2, 31)
	EFI_2_40_RUNTIME_SERVICES_REVISION  = EFI_SPECIFICATION_REVISION_MAJORMINOR(2, 40)
	EFI_2_50_RUNTIME_SERVICES_REVISION  = EFI_SPECIFICATION_REVISION_MAJORMINOR(2, 50)
	EFI_2_60_RUNTIME_SERVICES_REVISION  = EFI_SPECIFICATION_REVISION_MAJORMINOR(2, 60)
	EFI_2_70_RUNTIME_SERVICES_REVISION  = EFI_SPECIFICATION_REVISION_MAJORMINOR(2, 70)
	EFI_2_80_RUNTIME_SERVICES_REVISION  = EFI_SPECIFICATION_REVISION_MAJORMINOR(2, 80)
	EFI_2_90_RUNTIME_SERVICES_REVISION  = EFI_SPECIFICATION_REVISION_MAJORMINOR(2, 90)
	EFI_2_100_RUNTIME_SERVICES_REVISION = EFI_SPECIFICATION_REVISION_MAJORMINOR(2, 100)
	EFI_RUNTIME_SERVICES_REVISION       = EFI_SPECIFICATION_VERSION
)

// EFI_RUNTIME_SERVICES
// EFI Runtime Services Table.
type EFI_RUNTIME_SERVICES struct {
	Hdr                       EFI_TABLE_HEADER
	getTime                   uintptr
	setTime                   uintptr
	getWakeupTime             uintptr
	setWakeupTime             uintptr
	setVirtualAddressMap      uintptr
	convertPointer            uintptr
	getVariable               uintptr
	getNextVariableName       uintptr
	setVariable               uintptr
	getNextHighMonotonicCount uintptr
	resetSystem               uintptr
	updateCapsule             uintptr
	queryCapsuleCapabilities  uintptr
	queryVariableInfo         uintptr
}

// GetTime
// Returns the current time and date information, and the time-keeping capabilities
// of the hardware platform.
// @param[out]  Time             A pointer to storage to receive a snapshot of the current time.
// @param[out]  Capabilities     An optional pointer to a buffer to receive the real time clock
// ..............................device's capabilities.
// @retval EFI_SUCCESS           The operation completed successfully.
// @retval EFI_INVALID_PARAMETER Time is NULL.
// @retval EFI_DEVICE_ERROR      The time could not be retrieved due to hardware error.
func (p *EFI_RUNTIME_SERVICES) GetTime(Time *EFI_TIME, Capabilities *EFI_TIME_CAPABILITIES) EFI_STATUS {
	return UefiCall2(p.getTime, uintptr(unsafe.Pointer(Time)), uintptr(unsafe.Pointer(Capabilities)))
}

// SetTime
// Sets the current local time and date information.
// @param[in]  Time              A pointer to the current time.
// @retval EFI_SUCCESS           The operation completed successfully.
// @retval EFI_INVALID_PARAMETER A time field is out of range.
// @retval EFI_DEVICE_ERROR      The time could not be set due due to hardware error.
func (p *EFI_RUNTIME_SERVICES) SetTime(Time *EFI_TIME) EFI_STATUS {
	return UefiCall1(p.setTime, uintptr(unsafe.Pointer(Time)))
}

// GetWakeupTime
// Returns the current wakeup alarm clock setting.
// @param[out]  Enabled          Indicates if the alarm is currently enabled or disabled.
// @param[out]  Pending          Indicates if the alarm signal is pending and requires acknowledgement.
// @param[out]  Time             The current alarm setting.
// @retval EFI_SUCCESS           The alarm settings were returned.
// @retval EFI_INVALID_PARAMETER Enabled is NULL.
// @retval EFI_INVALID_PARAMETER Pending is NULL.
// @retval EFI_INVALID_PARAMETER Time is NULL.
// @retval EFI_DEVICE_ERROR      The wakeup time could not be retrieved due to a hardware error.
// @retval EFI_UNSUPPORTED       A wakeup timer is not supported on this platform.
func (p *EFI_RUNTIME_SERVICES) GetWakeupTime(Enabled *BOOLEAN, Pending *BOOLEAN, Time *EFI_TIME) EFI_STATUS {
	return UefiCall3(p.getWakeupTime, uintptr(unsafe.Pointer(Enabled)), uintptr(unsafe.Pointer(Pending)), uintptr(unsafe.Pointer(Time)))
}

// SetWakeupTime
// Sets the system wakeup alarm clock time.
// @param[in]  Enable            Enable or disable the wakeup alarm.
// @param[in]  Time              If Enable is TRUE, the time to set the wakeup alarm for.
// ..............................If Enable is FALSE, then this parameter is optional, and may be NULL.
// @retval EFI_SUCCESS           If Enable is TRUE, then the wakeup alarm was enabled. If
// ..............................Enable is FALSE, then the wakeup alarm was disabled.
// @retval EFI_INVALID_PARAMETER A time field is out of range.
// @retval EFI_DEVICE_ERROR      The wakeup time could not be set due to a hardware error.
// @retval EFI_UNSUPPORTED       A wakeup timer is not supported on this platform.
func (p *EFI_RUNTIME_SERVICES) SetWakeupTime(Enable BOOLEAN, Time *EFI_TIME) EFI_STATUS {
	var enable uintptr = 0
	if Enable {
		enable = 1
	}
	return UefiCall2(p.setWakeupTime, enable, uintptr(unsafe.Pointer(Time)))
}

// SetVirtualAddressMap
// Changes the runtime addressing mode of EFI firmware from physical to virtual.
// @param[in]  MemoryMapSize     The size in bytes of VirtualMap.
// @param[in]  DescriptorSize    The size in bytes of an entry in the VirtualMap.
// @param[in]  DescriptorVersion The version of the structure entries in VirtualMap.
// @param[in]  VirtualMap        An array of memory descriptors which contain new virtual
// ..............................address mapping information for all runtime ranges.
// @retval EFI_SUCCESS           The virtual address map has been applied.
// @retval EFI_UNSUPPORTED       EFI firmware is not at runtime, or the EFI firmware is already in
// ..............................virtual address mapped mode.
// @retval EFI_INVALID_PARAMETER DescriptorSize or DescriptorVersion is invalid.
// @retval EFI_NO_MAPPING        A virtual address was not supplied for a range in the memory
// ..............................map that requires a mapping.
// @retval EFI_NOT_FOUND         A virtual address was supplied for an address that is not found
// ..............................in the memory map.
func (p *EFI_RUNTIME_SERVICES) SetVirtualAddressMap(MemoryMapSize UINTN, DescriptorSize UINTN, DescriptorVersion uint32, VirtualMap *EFI_MEMORY_DESCRIPTOR) EFI_STATUS {
	return UefiCall4(p.setVirtualAddressMap, uintptr(MemoryMapSize), uintptr(DescriptorSize), uintptr(DescriptorVersion), uintptr(unsafe.Pointer(VirtualMap)))
}

// ConvertPointer
// Determines the new virtual address that is to be used on subsequent memory accesses.
// @param[in]       DebugDisposition  Supplies type information for the pointer being converted.
// @param[in, out]  Address           A pointer to a pointer that is to be fixed to be the value needed
// .................for the new virtual address mappings being applied.
// @retval EFI_SUCCESS           The pointer pointed to by Address was modified.
// @retval EFI_INVALID_PARAMETER 1) Address is NULL.
// ..............................2) *Address is NULL and DebugDisposition does
// ..............................not have the EFI_OPTIONAL_PTR bit set.
// @retval EFI_NOT_FOUND         The pointer pointed to by Address was not found to be part
// ..............................of the current memory map. This is normally fatal.
func (p *EFI_RUNTIME_SERVICES) ConvertPointer(DebugDisposition UINTN, Address **VOID) EFI_STATUS {
	return UefiCall2(p.convertPointer, uintptr(DebugDisposition), uintptr(unsafe.Pointer(Address)))
}

// GetVariable
// Returns the value of a variable.
// @param[in]       VariableName  A Null-terminated string that is the name of the vendor's
// ...............................variable.
// @param[in]       VendorGuid    A unique identifier for the vendor.
// @param[out]      Attributes    If not NULL, a pointer to the memory location to return the
// ...............................attributes bitmask for the variable.
// @param[in, out]  DataSize      On input, the size in bytes of the return Data buffer.
// .................On output the size of data returned in Data.
// @param[out]      Data          The buffer to return the contents of the variable. May be NULL
// ...............................with a zero DataSize in order to determine the size buffer needed.
// @retval EFI_SUCCESS            The function completed successfully.
// @retval EFI_NOT_FOUND          The variable was not found.
// @retval EFI_BUFFER_TOO_SMALL   The DataSize is too small for the result.
// @retval EFI_INVALID_PARAMETER  VariableName is NULL.
// @retval EFI_INVALID_PARAMETER  VendorGuid is NULL.
// @retval EFI_INVALID_PARAMETER  DataSize is NULL.
// @retval EFI_INVALID_PARAMETER  The DataSize is not too small and Data is NULL.
// @retval EFI_DEVICE_ERROR       The variable could not be retrieved due to a hardware error.
// @retval EFI_SECURITY_VIOLATION The variable could not be retrieved due to an authentication failure.
func (p *EFI_RUNTIME_SERVICES) GetVariable(VariableName *CHAR16, VendorGuid *EFI_GUID, Attributes *uint32, DataSize *UINTN, Data *VOID) EFI_STATUS {
	return UefiCall5(p.getVariable, uintptr(unsafe.Pointer(VariableName)), uintptr(unsafe.Pointer(VendorGuid)), uintptr(unsafe.Pointer(Attributes)), uintptr(unsafe.Pointer(DataSize)), uintptr(unsafe.Pointer(Data)))
}

// GetNextVariableName
// Enumerates the current variable names.
// @param[in, out]  VariableNameSize The size of the VariableName buffer. The size must be large
// .................enough to fit input string supplied in VariableName buffer.
// @param[in, out]  VariableName     On input, supplies the last VariableName that was returned
// .................by GetNextVariableName(). On output, returns the Nullterminated
// .................string of the current variable.
// @param[in, out]  VendorGuid       On input, supplies the last VendorGuid that was returned by
// .................GetNextVariableName(). On output, returns the
// .................VendorGuid of the current variable.
// @retval EFI_SUCCESS           The function completed successfully.
// @retval EFI_NOT_FOUND         The next variable was not found.
// @retval EFI_BUFFER_TOO_SMALL  The VariableNameSize is too small for the result.
// ..............................VariableNameSize has been updated with the size needed to complete the request.
// @retval EFI_INVALID_PARAMETER VariableNameSize is NULL.
// @retval EFI_INVALID_PARAMETER VariableName is NULL.
// @retval EFI_INVALID_PARAMETER VendorGuid is NULL.
// @retval EFI_INVALID_PARAMETER The input values of VariableName and VendorGuid are not a name and
// ..............................GUID of an existing variable.
// @retval EFI_INVALID_PARAMETER Null-terminator is not found in the first VariableNameSize bytes of
// ..............................the input VariableName buffer.
// @retval EFI_DEVICE_ERROR      The variable could not be retrieved due to a hardware error.
func (p *EFI_RUNTIME_SERVICES) GetNextVariableName(VariableNameSize *UINTN, VariableName *CHAR16, VendorGuid *EFI_GUID) EFI_STATUS {
	return UefiCall3(p.getNextVariableName, uintptr(unsafe.Pointer(VariableNameSize)), uintptr(unsafe.Pointer(VariableName)), uintptr(unsafe.Pointer(VendorGuid)))
}

// SetVariable
// Sets the value of a variable.
// @param[in]  VariableName       A Null-terminated string that is the name of the vendor's variable.
// ...............................Each VariableName is unique for each VendorGuid. VariableName must
// ...............................contain 1 or more characters. If VariableName is an empty string,
// ...............................then EFI_INVALID_PARAMETER is returned.
// @param[in]  VendorGuid         A unique identifier for the vendor.
// @param[in]  Attributes         Attributes bitmask to set for the variable.
// @param[in]  DataSize           The size in bytes of the Data buffer. Unless the EFI_VARIABLE_APPEND_WRITE or
// ...............................EFI_VARIABLE_TIME_BASED_AUTHENTICATED_WRITE_ACCESS attribute is set, a size of zero
// ...............................causes the variable to be deleted. When the EFI_VARIABLE_APPEND_WRITE attribute is
// ...............................set, then a SetVariable() call with a DataSize of zero will not cause any change to
// ...............................the variable value (the timestamp associated with the variable may be updated however
// ...............................even if no new data value is provided,see the description of the
// ...............................EFI_VARIABLE_AUTHENTICATION_2 descriptor below. In this case the DataSize will not
// ...............................be zero since the EFI_VARIABLE_AUTHENTICATION_2 descriptor will be populated).
// @param[in]  Data               The contents for the variable.
// @retval EFI_SUCCESS            The firmware has successfully stored the variable and its data as
// ...............................defined by the Attributes.
// @retval EFI_INVALID_PARAMETER  An invalid combination of attribute bits, name, and GUID was supplied, or the
// ...............................DataSize exceeds the maximum allowed.
// @retval EFI_INVALID_PARAMETER  VariableName is an empty string.
// @retval EFI_OUT_OF_RESOURCES   Not enough storage is available to hold the variable and its data.
// @retval EFI_DEVICE_ERROR       The variable could not be retrieved due to a hardware error.
// @retval EFI_WRITE_PROTECTED    The variable in question is read-only.
// @retval EFI_WRITE_PROTECTED    The variable in question cannot be deleted.
// @retval EFI_SECURITY_VIOLATION The variable could not be written due to EFI_VARIABLE_TIME_BASED_AUTHENTICATED_WRITE_ACESS being set,
// ...............................but the AuthInfo does NOT pass the validation check carried out by the firmware.
// @retval EFI_NOT_FOUND          The variable trying to be updated or deleted was not found.
func (p *EFI_RUNTIME_SERVICES) SetVariable(VariableName *CHAR16, VendorGuid *EFI_GUID, Attributes uint32, DataSize UINTN, Data *VOID) EFI_STATUS {
	return UefiCall5(p.setVariable, uintptr(unsafe.Pointer(VariableName)), uintptr(unsafe.Pointer(VendorGuid)), uintptr(Attributes), uintptr(DataSize), uintptr(unsafe.Pointer(Data)))
}

// GetNextHighMonotonicCount
// Returns the next high 32 bits of the platform's monotonic counter.
// @param[out]  HighCount        The pointer to returned value.
// @retval EFI_SUCCESS           The next high monotonic count was returned.
// @retval EFI_INVALID_PARAMETER HighCount is NULL.
// @retval EFI_DEVICE_ERROR      The device is not functioning properly.
func (p *EFI_RUNTIME_SERVICES) GetNextHighMonotonicCount(HighCount *uint32) EFI_STATUS {
	return UefiCall1(p.getNextHighMonotonicCount, uintptr(unsafe.Pointer(HighCount)))
}

// ResetSystem
// Resets the entire platform.
// @param[in]  ResetType         The type of reset to perform.
// @param[in]  ResetStatus       The status code for the reset.
// @param[in]  DataSize          The size, in bytes, of ResetData.
// @param[in]  ResetData         For a ResetType of EfiResetCold, EfiResetWarm, or
// ..............................EfiResetShutdown the data buffer starts with a Null-terminated
// ..............................string, optionally followed by additional binary data.
// ..............................The string is a description that the caller may use to further
// ..............................indicate the reason for the system reset.
// ..............................For a ResetType of EfiResetPlatformSpecific the data buffer
// ..............................also starts with a Null-terminated string that is followed
// ..............................by an EFI_GUID that describes the specific type of reset to perform.
func (p *EFI_RUNTIME_SERVICES) ResetSystem(ResetType EFI_RESET_TYPE, ResetStatus EFI_STATUS, DataSize UINTN, ResetData *VOID) EFI_STATUS {
	return UefiCall4(p.resetSystem, uintptr(ResetType), uintptr(ResetStatus), uintptr(DataSize), uintptr(unsafe.Pointer(ResetData)))
}

// UpdateCapsule
// Passes capsules to the firmware with both virtual and physical mapping. Depending on the intended
// consumption, the firmware may process the capsule immediately. If the payload should persist
// across a system reset, the reset value returned from EFI_QueryCapsuleCapabilities must
// be passed into ResetSystem() and will cause the capsule to be processed by the firmware as
// part of the reset process.
// @param[in]  CapsuleHeaderArray Virtual pointer to an array of virtual pointers to the capsules
// ...............................being passed into update capsule.
// @param[in]  CapsuleCount       Number of pointers to EFI_CAPSULE_HEADER in
// ...............................CaspuleHeaderArray.
// @param[in]  ScatterGatherList  Physical pointer to a set of
// ...............................EFI_CAPSULE_BLOCK_DESCRIPTOR that describes the
// ...............................location in physical memory of a set of capsules.
// @retval EFI_SUCCESS           Valid capsule was passed. If
// ..............................CAPSULE_FLAGS_PERSIT_ACROSS_RESET is not set, the
// ..............................capsule has been successfully processed by the firmware.
// @retval EFI_INVALID_PARAMETER CapsuleSize is NULL, or an incompatible set of flags were
// ..............................set in the capsule header.
// @retval EFI_INVALID_PARAMETER CapsuleCount is 0.
// @retval EFI_DEVICE_ERROR      The capsule update was started, but failed due to a device error.
// @retval EFI_UNSUPPORTED       The capsule type is not supported on this platform.
// @retval EFI_OUT_OF_RESOURCES  When ExitBootServices() has been previously called this error indicates the capsule
// ..............................is compatible with this platform but is not capable of being submitted or processed
// ..............................in runtime. The caller may resubmit the capsule prior to ExitBootServices().
// @retval EFI_OUT_OF_RESOURCES  When ExitBootServices() has not been previously called then this error indicates
// ..............................the capsule is compatible with this platform but there are insufficient resources to process.
func (p *EFI_RUNTIME_SERVICES) UpdateCapsule(CapsuleHeaderArray **EFI_CAPSULE_HEADER, CapsuleCount UINTN, ScatterGatherList EFI_PHYSICAL_ADDRESS) EFI_STATUS {
	return UefiCall3(p.updateCapsule, uintptr(unsafe.Pointer(CapsuleHeaderArray)), uintptr(CapsuleCount), uintptr(ScatterGatherList))
}

// QueryCapsuleCapabilities
// Returns if the capsule can be supported via UpdateCapsule().
// @param[in]   CapsuleHeaderArray  Virtual pointer to an array of virtual pointers to the capsules
// .................................being passed into update capsule.
// @param[in]   CapsuleCount        Number of pointers to EFI_CAPSULE_HEADER in
// .................................CaspuleHeaderArray.
// @param[out]  MaxiumCapsuleSize   On output the maximum size that UpdateCapsule() can
// .................................support as an argument to UpdateCapsule() via
// .................................CapsuleHeaderArray and ScatterGatherList.
// @param[out]  ResetType           Returns the type of reset required for the capsule update.
// @retval EFI_SUCCESS           Valid answer returned.
// @retval EFI_UNSUPPORTED       The capsule type is not supported on this platform, and
// ..............................MaximumCapsuleSize and ResetType are undefined.
// @retval EFI_INVALID_PARAMETER MaximumCapsuleSize is NULL.
// @retval EFI_OUT_OF_RESOURCES  When ExitBootServices() has been previously called this error indicates the capsule
// ..............................is compatible with this platform but is not capable of being submitted or processed
// ..............................in runtime. The caller may resubmit the capsule prior to ExitBootServices().
// @retval EFI_OUT_OF_RESOURCES  When ExitBootServices() has not been previously called then this error indicates
// ..............................the capsule is compatible with this platform but there are insufficient resources to process.
func (p *EFI_RUNTIME_SERVICES) QueryCapsuleCapabilities(CapsuleHeaderArray **EFI_CAPSULE_HEADER, CapsuleCount UINTN, MaximumCapsuleSize *uint64, ResetType *EFI_RESET_TYPE) EFI_STATUS {
	return UefiCall4(p.queryCapsuleCapabilities, uintptr(unsafe.Pointer(CapsuleHeaderArray)), uintptr(CapsuleCount), uintptr(unsafe.Pointer(MaximumCapsuleSize)), uintptr(unsafe.Pointer(ResetType)))
}

// QueryVariableInfo
// Returns information about the EFI variables.
// @param[in]   Attributes                   Attributes bitmask to specify the type of variables on
// ..........................................which to return information.
// @param[out]  MaximumVariableStorageSize   On output the maximum size of the storage space
// ..........................................available for the EFI variables associated with the
// ..........................................attributes specified.
// @param[out]  RemainingVariableStorageSize Returns the remaining size of the storage space
// ..........................................available for the EFI variables associated with the
// ..........................................attributes specified.
// @param[out]  MaximumVariableSize          Returns the maximum size of the individual EFI
// ..........................................variables associated with the attributes specified.
// @retval EFI_SUCCESS                  Valid answer returned.
// @retval EFI_INVALID_PARAMETER        An invalid combination of attribute bits was supplied
// @retval EFI_UNSUPPORTED              The attribute is not supported on this platform, and the
// .....................................MaximumVariableStorageSize,
// .....................................RemainingVariableStorageSize, MaximumVariableSize
// .....................................are undefined.
func (p *EFI_RUNTIME_SERVICES) QueryVariableInfo(Attributes uint32, MaximumVariableStorageSize *uint64, RemainingVariableStorageSize *uint64, MaximumVariableSize *uint64) EFI_STATUS {
	return UefiCall4(p.queryVariableInfo, uintptr(Attributes), uintptr(unsafe.Pointer(MaximumVariableStorageSize)), uintptr(unsafe.Pointer(RemainingVariableStorageSize)), uintptr(unsafe.Pointer(MaximumVariableSize)))
}

// endregion

// region: EFI Boot Services Table

type EVENT_TYPE uint32

const (
	EVT_TIMER           EVENT_TYPE = 0x80000000
	EVT_RUNTIME         EVENT_TYPE = 0x40000000
	EVT_RUNTIME_CONTEXT EVENT_TYPE = 0x20000000

	EVT_NOTIFY_WAIT   EVENT_TYPE = 0x00000100
	EVT_NOTIFY_SIGNAL EVENT_TYPE = 0x00000200

	EVT_SIGNAL_EXIT_BOOT_SERVICES     EVENT_TYPE = 0x00000201
	EVT_SIGNAL_VIRTUAL_ADDRESS_CHANGE EVENT_TYPE = 0x60000202

	EVT_EFI_SIGNAL_MASK EVENT_TYPE = 0x000000FF
	EVT_EFI_SIGNAL_MAX  EVENT_TYPE = 4

	EFI_EVENT_TIMER                         = EVT_TIMER
	EFI_EVENT_RUNTIME                       = EVT_RUNTIME
	EFI_EVENT_RUNTIME_CONTEXT               = EVT_RUNTIME_CONTEXT
	EFI_EVENT_NOTIFY_WAIT                   = EVT_NOTIFY_WAIT
	EFI_EVENT_NOTIFY_SIGNAL                 = EVT_NOTIFY_SIGNAL
	EFI_EVENT_SIGNAL_EXIT_BOOT_SERVICES     = EVT_SIGNAL_EXIT_BOOT_SERVICES
	EFI_EVENT_SIGNAL_VIRTUAL_ADDRESS_CHANGE = EVT_SIGNAL_VIRTUAL_ADDRESS_CHANGE
	EFI_EVENT_EFI_SIGNAL_MASK               = EVT_EFI_SIGNAL_MASK
	EFI_EVENT_EFI_SIGNAL_MAX                = EVT_EFI_SIGNAL_MAX
)

const (
	TPL_APPLICATION     EFI_TPL = 4
	TPL_CALLBACK        EFI_TPL = 8
	TPL_NOTIFY          EFI_TPL = 16
	TPL_HIGH_LEVEL      EFI_TPL = 31
	EFI_TPL_APPLICATION         = TPL_APPLICATION
	EFI_TPL_CALLBACK            = TPL_CALLBACK
	EFI_TPL_NOTIFY              = TPL_NOTIFY
	EFI_TPL_HIGH_LEVEL          = TPL_HIGH_LEVEL
)

// EFI_TIMER_DELAY is an enumeration of timer delays.
type EFI_TIMER_DELAY int

const (
	TimerCancel EFI_TIMER_DELAY = iota
	TimerPeriodic
	TimerRelative
	TimerTypeMax
)

// Enumeration for EFI_INTERFACE_TYPE
type EFI_INTERFACE_TYPE uint

const (
	EFI_NATIVE_INTERFACE EFI_INTERFACE_TYPE = iota
	EFI_PCODE_INTERFACE
)

// EFI_BOOT_SERVICES_SIGNATURE
const EFI_BOOT_SERVICES_SIGNATURE = 0x56524553544f4f42

// EFI_BOOT_SERVICES_REVISION
var (
	EFI_1_02_BOOT_SERVICES_REVISION  = EFI_SPECIFICATION_REVISION_MAJORMINOR(1, 02)
	EFI_1_10_BOOT_SERVICES_REVISION  = EFI_SPECIFICATION_REVISION_MAJORMINOR(1, 10)
	EFI_2_00_BOOT_SERVICES_REVISION  = EFI_SPECIFICATION_REVISION_MAJORMINOR(2, 00)
	EFI_2_10_BOOT_SERVICES_REVISION  = EFI_SPECIFICATION_REVISION_MAJORMINOR(2, 10)
	EFI_2_20_BOOT_SERVICES_REVISION  = EFI_SPECIFICATION_REVISION_MAJORMINOR(2, 20)
	EFI_2_30_BOOT_SERVICES_REVISION  = EFI_SPECIFICATION_REVISION_MAJORMINOR(2, 30)
	EFI_2_31_BOOT_SERVICES_REVISION  = EFI_SPECIFICATION_REVISION_MAJORMINOR(2, 31)
	EFI_2_40_BOOT_SERVICES_REVISION  = EFI_SPECIFICATION_REVISION_MAJORMINOR(2, 40)
	EFI_2_50_BOOT_SERVICES_REVISION  = EFI_SPECIFICATION_REVISION_MAJORMINOR(2, 50)
	EFI_2_60_BOOT_SERVICES_REVISION  = EFI_SPECIFICATION_REVISION_MAJORMINOR(2, 60)
	EFI_2_70_BOOT_SERVICES_REVISION  = EFI_SPECIFICATION_REVISION_MAJORMINOR(2, 70)
	EFI_2_80_BOOT_SERVICES_REVISION  = EFI_SPECIFICATION_REVISION_MAJORMINOR(2, 80)
	EFI_2_90_BOOT_SERVICES_REVISION  = EFI_SPECIFICATION_REVISION_MAJORMINOR(2, 90)
	EFI_2_100_BOOT_SERVICES_REVISION = EFI_SPECIFICATION_REVISION_MAJORMINOR(2, 100)
	EFI_BOOT_SERVICES_REVISION       = EFI_SPECIFICATION_VERSION
)

// EFI_BOOT_SERVICES
// EFI Boot Services Table.
type EFI_BOOT_SERVICES struct {
	Hdr                                 EFI_TABLE_HEADER
	raiseTPL                            uintptr
	restoreTPL                          uintptr
	allocatePages                       uintptr
	freePages                           uintptr
	getMemoryMap                        uintptr
	allocatePool                        uintptr
	freePool                            uintptr
	createEvent                         uintptr
	setTimer                            uintptr
	waitForEvent                        uintptr
	signalEvent                         uintptr
	closeEvent                          uintptr
	checkEvent                          uintptr
	installProtocolInterface            uintptr
	reinstallProtocolInterface          uintptr
	uninstallProtocolInterface          uintptr
	handleProtocol                      uintptr
	Reserved                            *VOID
	registerProtocolNotify              uintptr
	locateHandle                        uintptr
	locateDevicePath                    uintptr
	installConfigurationTable           uintptr
	loadImage                           uintptr
	startImage                          uintptr
	exit                                uintptr
	unloadImage                         uintptr
	exitBootServices                    uintptr
	getNextMonotonicCount               uintptr
	stall                               uintptr
	setWatchdogTimer                    uintptr
	connectController                   uintptr
	disconnectController                uintptr
	openProtocol                        uintptr
	closeProtocol                       uintptr
	openProtocolInformation             uintptr
	protocolsPerHandle                  uintptr
	locateHandleBuffer                  uintptr
	locateProtocol                      uintptr
	installMultipleProtocolInterfaces   uintptr
	uninstallMultipleProtocolInterfaces uintptr
	calculateCrc32                      uintptr
	copyMem                             uintptr
	setMem                              uintptr
	createEventEx                       uintptr
}

// RaiseTPL
// Raises a task's priority level and returns its previous level.
// @param[in]  NewTpl          The new task priority level.
// @return Previous task priority level
func (p *EFI_BOOT_SERVICES) RaiseTPL(NewTpl EFI_TPL) EFI_STATUS {
	return UefiCall1(p.raiseTPL, uintptr(NewTpl))
}

// RestoreTPL
// Restores a task's priority level to its previous value.
// @param[in]  OldTpl          The previous task priority level to restore.
func (p *EFI_BOOT_SERVICES) RestoreTPL(OldTpl EFI_TPL) EFI_STATUS {
	return UefiCall1(p.restoreTPL, uintptr(OldTpl))
}

// AllocatePages
// Allocates memory pages from the system.
// @param[in]       Type         The type of allocation to perform.
// @param[in]       MemoryType   The type of memory to allocate.
// ..............................MemoryType values in the range 0x70000000..0x7FFFFFFF
// ..............................are reserved for OEM use. MemoryType values in the range
// ..............................0x80000000..0xFFFFFFFF are reserved for use by UEFI OS loaders
// ..............................that are provided by operating system vendors.
// @param[in]       Pages        The number of contiguous 4 KB pages to allocate.
// @param[in, out]  Memory       The pointer to a physical address. On input, the way in which the address is
// .................used depends on the value of Type.
// @retval EFI_SUCCESS           The requested pages were allocated.
// @retval EFI_INVALID_PARAMETER 1) Type is not AllocateAnyPages or
// ..............................AllocateMaxAddress or AllocateAddress.
// ..............................2) MemoryType is in the range
// ..............................EfiMaxMemoryType..0x6FFFFFFF.
// ..............................3) Memory is NULL.
// ..............................4) MemoryType is EfiPersistentMemory.
// @retval EFI_OUT_OF_RESOURCES  The pages could not be allocated.
// @retval EFI_NOT_FOUND         The requested pages could not be found.
func (p *EFI_BOOT_SERVICES) AllocatePages(Type EFI_ALLOCATE_TYPE, MemoryType EFI_MEMORY_TYPE, Pages UINTN, Memory *EFI_PHYSICAL_ADDRESS) EFI_STATUS {
	return UefiCall4(p.allocatePages, uintptr(Type), uintptr(MemoryType), uintptr(Pages), uintptr(unsafe.Pointer(Memory)))
}

// FreePages
// Frees memory pages.
// @param[in]  Memory      The base physical address of the pages to be freed.
// @param[in]  Pages       The number of contiguous 4 KB pages to free.
// @retval EFI_SUCCESS           The requested pages were freed.
// @retval EFI_INVALID_PARAMETER Memory is not a page-aligned address or Pages is invalid.
// @retval EFI_NOT_FOUND         The requested memory pages were not allocated with
// ..............................AllocatePages().
func (p *EFI_BOOT_SERVICES) FreePages(Memory EFI_PHYSICAL_ADDRESS, Pages UINTN) EFI_STATUS {
	return UefiCall2(p.freePages, uintptr(Memory), uintptr(Pages))
}

// GetMemoryMap
// Returns the current memory map.
// @param[in, out]  MemoryMapSize         A pointer to the size, in bytes, of the MemoryMap buffer.
// .................On input, this is the size of the buffer allocated by the caller.
// .................On output, it is the size of the buffer returned by the firmware if
// .................the buffer was large enough, or the size of the buffer needed to contain
// .................the map if the buffer was too small.
// @param[out]      MemoryMap             A pointer to the buffer in which firmware places the current memory
// .......................................map.
// @param[out]      MapKey                A pointer to the location in which firmware returns the key for the
// .......................................current memory map.
// @param[out]      DescriptorSize        A pointer to the location in which firmware returns the size, in bytes, of
// .......................................an individual EFI_MEMORY_DESCRIPTOR.
// @param[out]      DescriptorVersion     A pointer to the location in which firmware returns the version number
// .......................................associated with the EFI_MEMORY_DESCRIPTOR.
// @retval EFI_SUCCESS           The memory map was returned in the MemoryMap buffer.
// @retval EFI_BUFFER_TOO_SMALL  The MemoryMap buffer was too small. The current buffer size
// ..............................needed to hold the memory map is returned in MemoryMapSize.
// @retval EFI_INVALID_PARAMETER 1) MemoryMapSize is NULL.
// ..............................2) The MemoryMap buffer is not too small and MemoryMap is
// ..............................NULL.
func (p *EFI_BOOT_SERVICES) GetMemoryMap(MemoryMapSize *UINTN, MemoryMap *EFI_MEMORY_DESCRIPTOR, MapKey *UINTN, DescriptorSize *UINTN, DescriptorVersion *uint32) EFI_STATUS {
	return UefiCall5(p.getMemoryMap, uintptr(unsafe.Pointer(MemoryMapSize)), uintptr(unsafe.Pointer(MemoryMap)), uintptr(unsafe.Pointer(MapKey)), uintptr(unsafe.Pointer(DescriptorSize)), uintptr(unsafe.Pointer(DescriptorVersion)))
}

// AllocatePool
// Allocates pool memory.
// @param[in]   PoolType         The type of pool to allocate.
// ..............................MemoryType values in the range 0x70000000..0x7FFFFFFF
// ..............................are reserved for OEM use. MemoryType values in the range
// ..............................0x80000000..0xFFFFFFFF are reserved for use by UEFI OS loaders
// ..............................that are provided by operating system vendors.
// @param[in]   Size             The number of bytes to allocate from the pool.
// @param[out]  Buffer           A pointer to a pointer to the allocated buffer if the call succeeds;
// ..............................undefined otherwise.
// @retval EFI_SUCCESS           The requested number of bytes was allocated.
// @retval EFI_OUT_OF_RESOURCES  The pool requested could not be allocated.
// @retval EFI_INVALID_PARAMETER Buffer is NULL.
// ..............................PoolType is in the range EfiMaxMemoryType..0x6FFFFFFF.
// ..............................PoolType is EfiPersistentMemory.
func (p *EFI_BOOT_SERVICES) AllocatePool(PoolType EFI_MEMORY_TYPE, Size UINTN, Buffer **VOID) EFI_STATUS {
	return UefiCall3(p.allocatePool, uintptr(PoolType), uintptr(Size), uintptr(unsafe.Pointer(Buffer)))
}

// FreePool
// Returns pool memory to the system.
// @param[in]  Buffer            The pointer to the buffer to free.
// @retval EFI_SUCCESS           The memory was returned to the system.
// @retval EFI_INVALID_PARAMETER Buffer was invalid.
func (p *EFI_BOOT_SERVICES) FreePool(Buffer *VOID) EFI_STATUS {
	return UefiCall1(p.freePool, uintptr(unsafe.Pointer(Buffer)))
}

// CreateEvent
// Creates an event.
// @param[in]   Type             The type of event to create and its mode and attributes.
// @param[in]   NotifyTpl        The task priority level of event notifications, if needed.
// @param[in]   NotifyFunction   The pointer to the event's notification function, if any.
// @param[in]   NotifyContext    The pointer to the notification function's context; corresponds to parameter
// ..............................Context in the notification function.
// @param[out]  Event            The pointer to the newly created event if the call succeeds; undefined
// ..............................otherwise.
// @retval EFI_SUCCESS           The event structure was created.
// @retval EFI_INVALID_PARAMETER One or more parameters are invalid.
// @retval EFI_OUT_OF_RESOURCES  The event could not be allocated.
func (p *EFI_BOOT_SERVICES) CreateEvent(Type EVENT_TYPE, NotifyTpl EFI_TPL, NotifyFunction unsafe.Pointer, NotifyContext unsafe.Pointer, Event *EFI_EVENT) EFI_STATUS {
	return UefiCall5(p.createEvent, uintptr(Type), uintptr(NotifyTpl), uintptr(NotifyFunction), uintptr(NotifyContext), uintptr(unsafe.Pointer(Event)))
}

// SetTimer
// Sets the type of timer and the trigger time for a timer event.
// @param[in]  Event             The timer event that is to be signaled at the specified time.
// @param[in]  Type              The type of time that is specified in TriggerTime.
// @param[in]  TriggerTime       The number of 100ns units until the timer expires.
// ..............................A TriggerTime of 0 is legal.
// ..............................If Type is TimerRelative and TriggerTime is 0, then the timer
// ..............................event will be signaled on the next timer tick.
// ..............................If Type is TimerPeriodic and TriggerTime is 0, then the timer
// ..............................event will be signaled on every timer tick.
// @retval EFI_SUCCESS           The event has been set to be signaled at the requested time.
// @retval EFI_INVALID_PARAMETER Event or Type is not valid.
func (p *EFI_BOOT_SERVICES) SetTimer(Event EFI_EVENT, Type EFI_TIMER_DELAY, TriggerTime uint64) EFI_STATUS {
	return UefiCall3(p.setTimer, uintptr(Event), uintptr(Type), uintptr(TriggerTime))
}

// WaitForEvent
// Stops execution until an event is signaled.
// @param[in]   NumberOfEvents   The number of events in the Event array.
// @param[in]   Event            An array of EFI_EVENT.
// @param[out]  Index            The pointer to the index of the event which satisfied the wait condition.
// @retval EFI_SUCCESS           The event indicated by Index was signaled.
// @retval EFI_INVALID_PARAMETER 1) NumberOfEvents is 0.
// ..............................2) The event indicated by Index is of type
// ..............................EVT_NOTIFY_SIGNAL.
// @retval EFI_UNSUPPORTED       The current TPL is not TPL_APPLICATION.
func (p *EFI_BOOT_SERVICES) WaitForEvent(NumberOfEvents UINTN, Event *EFI_EVENT, Index *UINTN) EFI_STATUS {
	return UefiCall3(p.waitForEvent, uintptr(NumberOfEvents), uintptr(unsafe.Pointer(Event)), uintptr(unsafe.Pointer(Index)))
}

// SignalEvent
// Signals an event.
// @param[in]  Event             The event to signal.
// @retval EFI_SUCCESS           The event has been signaled.
func (p *EFI_BOOT_SERVICES) SignalEvent(Event EFI_EVENT) EFI_STATUS {
	return UefiCall1(p.signalEvent, uintptr(Event))
}

// CloseEvent
// Closes an event.
// @param[in]  Event             The event to close.
// @retval EFI_SUCCESS           The event has been closed.
func (p *EFI_BOOT_SERVICES) CloseEvent(Event EFI_EVENT) EFI_STATUS {
	return UefiCall1(p.closeEvent, uintptr(Event))
}

// CheckEvent
// Checks whether an event is in the signaled state.
// @param[in]  Event             The event to check.
// @retval EFI_SUCCESS           The event is in the signaled state.
// @retval EFI_NOT_READY         The event is not in the signaled state.
// @retval EFI_INVALID_PARAMETER Event is of type EVT_NOTIFY_SIGNAL.
func (p *EFI_BOOT_SERVICES) CheckEvent(Event EFI_EVENT) EFI_STATUS {
	return UefiCall1(p.checkEvent, uintptr(Event))
}

// InstallProtocolInterface
// Installs a protocol interface on a device handle. If the handle does not exist, it is created and added
// to the list of handles in the system. InstallMultipleProtocolInterfaces() performs
// more error checking than InstallProtocolInterface(), so it is recommended that
// InstallMultipleProtocolInterfaces() be used in place of
// InstallProtocolInterface()
// @param[in, out]  Handle         A pointer to the EFI_HANDLE on which the interface is to be installed.
// @param[in]       Protocol       The numeric ID of the protocol interface.
// @param[in]       InterfaceType  Indicates whether Interface is supplied in native form.
// @param[in]       Interface      A pointer to the protocol interface.
// @retval EFI_SUCCESS           The protocol interface was installed.
// @retval EFI_OUT_OF_RESOURCES  Space for a new handle could not be allocated.
// @retval EFI_INVALID_PARAMETER Handle is NULL.
// @retval EFI_INVALID_PARAMETER Protocol is NULL.
// @retval EFI_INVALID_PARAMETER InterfaceType is not EFI_NATIVE_INTERFACE.
// @retval EFI_INVALID_PARAMETER Protocol is already installed on the handle specified by Handle.
func (p *EFI_BOOT_SERVICES) InstallProtocolInterface(Handle *EFI_HANDLE, Protocol *EFI_GUID, InterfaceType EFI_INTERFACE_TYPE, Interface *VOID) EFI_STATUS {
	return UefiCall4(p.installProtocolInterface, uintptr(unsafe.Pointer(Handle)), uintptr(unsafe.Pointer(Protocol)), uintptr(InterfaceType), uintptr(unsafe.Pointer(Interface)))
}

// ReinstallProtocolInterface
// Reinstalls a protocol interface on a device handle.
// @param[in]  Handle            Handle on which the interface is to be reinstalled.
// @param[in]  Protocol          The numeric ID of the interface.
// @param[in]  OldInterface      A pointer to the old interface. NULL can be used if a structure is not
// ..............................associated with Protocol.
// @param[in]  NewInterface      A pointer to the new interface.
// @retval EFI_SUCCESS           The protocol interface was reinstalled.
// @retval EFI_NOT_FOUND         The OldInterface on the handle was not found.
// @retval EFI_ACCESS_DENIED     The protocol interface could not be reinstalled,
// ..............................because OldInterface is still being used by a
// ..............................driver that will not release it.
// @retval EFI_INVALID_PARAMETER Handle is NULL.
// @retval EFI_INVALID_PARAMETER Protocol is NULL.
func (p *EFI_BOOT_SERVICES) ReinstallProtocolInterface(Handle EFI_HANDLE, Protocol *EFI_GUID, OldInterface *VOID, NewInterface *VOID) EFI_STATUS {
	return UefiCall4(p.reinstallProtocolInterface, uintptr(Handle), uintptr(unsafe.Pointer(Protocol)), uintptr(unsafe.Pointer(OldInterface)), uintptr(unsafe.Pointer(NewInterface)))
}

// UninstallProtocolInterface
// Removes a protocol interface from a device handle. It is recommended that
// UninstallMultipleProtocolInterfaces() be used in place of
// UninstallProtocolInterface().
// @param[in]  Handle            The handle on which the interface was installed.
// @param[in]  Protocol          The numeric ID of the interface.
// @param[in]  Interface         A pointer to the interface.
// @retval EFI_SUCCESS           The interface was removed.
// @retval EFI_NOT_FOUND         The interface was not found.
// @retval EFI_ACCESS_DENIED     The interface was not removed because the interface
// ..............................is still being used by a driver.
// @retval EFI_INVALID_PARAMETER Handle is NULL.
// @retval EFI_INVALID_PARAMETER Protocol is NULL.
func (p *EFI_BOOT_SERVICES) UninstallProtocolInterface(Handle EFI_HANDLE, Protocol *EFI_GUID, Interface *VOID) EFI_STATUS {
	return UefiCall3(p.uninstallProtocolInterface, uintptr(Handle), uintptr(unsafe.Pointer(Protocol)), uintptr(unsafe.Pointer(Interface)))
}

// HandleProtocol
// Queries a handle to determine if it supports a specified protocol.
// @param[in]   Handle           The handle being queried.
// @param[in]   Protocol         The published unique identifier of the protocol.
// @param[out]  Interface        Supplies the address where a pointer to the corresponding Protocol
// ..............................Interface is returned.
// @retval EFI_SUCCESS           The interface information for the specified protocol was returned.
// @retval EFI_UNSUPPORTED       The device does not support the specified protocol.
// @retval EFI_INVALID_PARAMETER Handle is NULL.
// @retval EFI_INVALID_PARAMETER Protocol is NULL.
// @retval EFI_INVALID_PARAMETER Interface is NULL.
func (p *EFI_BOOT_SERVICES) HandleProtocol(Handle EFI_HANDLE, Protocol *EFI_GUID, Interface unsafe.Pointer) EFI_STATUS {
	return UefiCall3(p.handleProtocol, uintptr(Handle), uintptr(unsafe.Pointer(Protocol)), uintptr(Interface))
}

// RegisterProtocolNotify
// Creates an event that is to be signaled whenever an interface is installed for a specified protocol.
// @param[in]   Protocol         The numeric ID of the protocol for which the event is to be registered.
// @param[in]   Event            Event that is to be signaled whenever a protocol interface is registered
// ..............................for Protocol.
// @param[out]  Registration     A pointer to a memory location to receive the registration value.
// @retval EFI_SUCCESS           The notification event has been registered.
// @retval EFI_OUT_OF_RESOURCES  Space for the notification event could not be allocated.
// @retval EFI_INVALID_PARAMETER Protocol is NULL.
// @retval EFI_INVALID_PARAMETER Event is NULL.
// @retval EFI_INVALID_PARAMETER Registration is NULL.
func (p *EFI_BOOT_SERVICES) RegisterProtocolNotify(Protocol *EFI_GUID, Event EFI_EVENT, Registration **VOID) EFI_STATUS {
	return UefiCall3(p.registerProtocolNotify, uintptr(unsafe.Pointer(Protocol)), uintptr(Event), uintptr(unsafe.Pointer(Registration)))
}

// LocateHandle
// Returns an array of handles that support a specified protocol.
// @param[in]       SearchType   Specifies which handle(s) are to be returned.
// @param[in]       Protocol     Specifies the protocol to search by.
// @param[in]       SearchKey    Specifies the search key.
// @param[in, out]  BufferSize   On input, the size in bytes of Buffer. On output, the size in bytes of
// .................the array returned in Buffer (if the buffer was large enough) or the
// .................size, in bytes, of the buffer needed to obtain the array (if the buffer was
// .................not large enough).
// @param[out]      Buffer       The buffer in which the array is returned.
// @retval EFI_SUCCESS           The array of handles was returned.
// @retval EFI_NOT_FOUND         No handles match the search.
// @retval EFI_BUFFER_TOO_SMALL  The BufferSize is too small for the result.
// @retval EFI_INVALID_PARAMETER SearchType is not a member of EFI_LOCATE_SEARCH_TYPE.
// @retval EFI_INVALID_PARAMETER SearchType is ByRegisterNotify and SearchKey is NULL.
// @retval EFI_INVALID_PARAMETER SearchType is ByProtocol and Protocol is NULL.
// @retval EFI_INVALID_PARAMETER One or more matches are found and BufferSize is NULL.
// @retval EFI_INVALID_PARAMETER BufferSize is large enough for the result and Buffer is NULL.
func (p *EFI_BOOT_SERVICES) LocateHandle(SearchType EFI_LOCATE_SEARCH_TYPE, Protocol *EFI_GUID, SearchKey *VOID, BufferSize *UINTN, Buffer *EFI_HANDLE) EFI_STATUS {
	return UefiCall5(p.locateHandle, uintptr(SearchType), uintptr(unsafe.Pointer(Protocol)), uintptr(unsafe.Pointer(SearchKey)), uintptr(unsafe.Pointer(BufferSize)), uintptr(unsafe.Pointer(Buffer)))
}

// LocateDevicePath
// Locates the handle to a device on the device path that supports the specified protocol.
// @param[in]       Protocol     Specifies the protocol to search for.
// @param[in, out]  DevicePath   On input, a pointer to a pointer to the device path. On output, the device
// .................path pointer is modified to point to the remaining part of the device
// .................path.
// @param[out]      Device       A pointer to the returned device handle.
// @retval EFI_SUCCESS           The resulting handle was returned.
// @retval EFI_NOT_FOUND         No handles match the search.
// @retval EFI_INVALID_PARAMETER Protocol is NULL.
// @retval EFI_INVALID_PARAMETER DevicePath is NULL.
// @retval EFI_INVALID_PARAMETER A handle matched the search and Device is NULL.
func (p *EFI_BOOT_SERVICES) LocateDevicePath(Protocol *EFI_GUID, DevicePath **EFI_DEVICE_PATH_PROTOCOL, Device *EFI_HANDLE) EFI_STATUS {
	return UefiCall3(p.locateDevicePath, uintptr(unsafe.Pointer(Protocol)), uintptr(unsafe.Pointer(DevicePath)), uintptr(unsafe.Pointer(Device)))
}

// InstallConfigurationTable
// Adds, updates, or removes a configuration table entry from the EFI System Table.
// @param[in]  Guid              A pointer to the GUID for the entry to add, update, or remove.
// @param[in]  Table             A pointer to the configuration table for the entry to add, update, or
// ..............................remove. May be NULL.
// @retval EFI_SUCCESS           The (Guid, Table) pair was added, updated, or removed.
// @retval EFI_NOT_FOUND         An attempt was made to delete a nonexistent entry.
// @retval EFI_INVALID_PARAMETER Guid is NULL.
// @retval EFI_OUT_OF_RESOURCES  There is not enough memory available to complete the operation.
func (p *EFI_BOOT_SERVICES) InstallConfigurationTable(Guid *EFI_GUID, Table *VOID) EFI_STATUS {
	return UefiCall2(p.installConfigurationTable, uintptr(unsafe.Pointer(Guid)), uintptr(unsafe.Pointer(Table)))
}

// LoadImage
// Loads an EFI image into memory.
// @param[in]   BootPolicy        If TRUE, indicates that the request originates from the boot
// ...............................manager, and that the boot manager is attempting to load
// ...............................FilePath as a boot selection. Ignored if SourceBuffer is
// ...............................not NULL.
// @param[in]   ParentImageHandle The caller's image handle.
// @param[in]   DevicePath        The DeviceHandle specific file path from which the image is
// ...............................loaded.
// @param[in]   SourceBuffer      If not NULL, a pointer to the memory location containing a copy
// ...............................of the image to be loaded.
// @param[in]   SourceSize        The size in bytes of SourceBuffer. Ignored if SourceBuffer is NULL.
// @param[out]  ImageHandle       The pointer to the returned image handle that is created when the
// ...............................image is successfully loaded.
// @retval EFI_SUCCESS            Image was loaded into memory correctly.
// @retval EFI_NOT_FOUND          Both SourceBuffer and DevicePath are NULL.
// @retval EFI_INVALID_PARAMETER  One or more parametes are invalid.
// @retval EFI_UNSUPPORTED        The image type is not supported.
// @retval EFI_OUT_OF_RESOURCES   Image was not loaded due to insufficient resources.
// @retval EFI_LOAD_ERROR         Image was not loaded because the image format was corrupt or not
// ...............................understood.
// @retval EFI_DEVICE_ERROR       Image was not loaded because the device returned a read error.
// @retval EFI_ACCESS_DENIED      Image was not loaded because the platform policy prohibits the
// ...............................image from being loaded. NULL is returned in *ImageHandle.
// @retval EFI_SECURITY_VIOLATION Image was loaded and an ImageHandle was created with a
// ...............................valid EFI_LOADED_IMAGE_PROTOCOL. However, the current
// ...............................platform policy specifies that the image should not be started.
func (p *EFI_BOOT_SERVICES) LoadImage(BootPolicy BOOLEAN, ParentImageHandle EFI_HANDLE, DevicePath *EFI_DEVICE_PATH_PROTOCOL, SourceBuffer *VOID, SourceSize UINTN, ImageHandle *EFI_HANDLE) EFI_STATUS {
	return UefiCall6(p.loadImage, convertBoolean(BootPolicy), uintptr(ParentImageHandle), uintptr(unsafe.Pointer(DevicePath)), uintptr(unsafe.Pointer(SourceBuffer)), uintptr(SourceSize), uintptr(unsafe.Pointer(ImageHandle)))
}

// StartImage
// Transfers control to a loaded image's entry point.
// @param[in]   ImageHandle       Handle of image to be started.
// @param[out]  ExitDataSize      The pointer to the size, in bytes, of ExitData.
// @param[out]  ExitData          The pointer to a pointer to a data buffer that includes a Null-terminated
// ...............................string, optionally followed by additional binary data.
// @retval EFI_INVALID_PARAMETER  ImageHandle is either an invalid image handle or the image
// ...............................has already been initialized with StartImage.
// @retval EFI_SECURITY_VIOLATION The current platform policy specifies that the image should not be started.
// @return Exit code from image
func (p *EFI_BOOT_SERVICES) StartImage(ImageHandle EFI_HANDLE, ExitDataSize *UINTN, ExitData **CHAR16) EFI_STATUS {
	return UefiCall3(p.startImage, uintptr(ImageHandle), uintptr(unsafe.Pointer(ExitDataSize)), uintptr(unsafe.Pointer(ExitData)))
}

// Exit
// Terminates a loaded EFI image and returns control to boot services.
// @param[in]  ImageHandle       Handle that identifies the image. This parameter is passed to the
// ..............................image on entry.
// @param[in]  ExitStatus        The image's exit code.
// @param[in]  ExitDataSize      The size, in bytes, of ExitData. Ignored if ExitStatus is EFI_SUCCESS.
// @param[in]  ExitData          The pointer to a data buffer that includes a Null-terminated string,
// ..............................optionally followed by additional binary data. The string is a
// ..............................description that the caller may use to further indicate the reason
// ..............................for the image's exit. ExitData is only valid if ExitStatus
// ..............................is something other than EFI_SUCCESS. The ExitData buffer
// ..............................must be allocated by calling AllocatePool().
// @retval EFI_SUCCESS           The image specified by ImageHandle was unloaded.
// @retval EFI_INVALID_PARAMETER The image specified by ImageHandle has been loaded and
// ..............................started with LoadImage() and StartImage(), but the
// ..............................image is not the currently executing image.
func (p *EFI_BOOT_SERVICES) Exit(ImageHandle EFI_HANDLE, ExitStatus EFI_STATUS, ExitDataSize UINTN, ExitData *CHAR16) EFI_STATUS {
	return UefiCall4(p.exit, uintptr(ImageHandle), uintptr(ExitStatus), uintptr(ExitDataSize), uintptr(unsafe.Pointer(ExitData)))
}

// UnloadImage
// Unloads an image.
// @param[in]  ImageHandle       Handle that identifies the image to be unloaded.
// @retval EFI_SUCCESS           The image has been unloaded.
// @retval EFI_INVALID_PARAMETER ImageHandle is not a valid image handle.
func (p *EFI_BOOT_SERVICES) UnloadImage(ImageHandle EFI_HANDLE) EFI_STATUS {
	return UefiCall1(p.unloadImage, uintptr(ImageHandle))
}

// ExitBootServices
// Terminates all boot services.
// @param[in]  ImageHandle       Handle that identifies the exiting image.
// @param[in]  MapKey            Key to the latest memory map.
// @retval EFI_SUCCESS           Boot services have been terminated.
// @retval EFI_INVALID_PARAMETER MapKey is incorrect.
func (p *EFI_BOOT_SERVICES) ExitBootServices(ImageHandle EFI_HANDLE, MapKey UINTN) EFI_STATUS {
	return UefiCall2(p.exitBootServices, uintptr(ImageHandle), uintptr(MapKey))
}

// GetNextMonotonicCount
// Returns a monotonically increasing count for the platform.
// @param[out]  Count            The pointer to returned value.
// @retval EFI_SUCCESS           The next monotonic count was returned.
// @retval EFI_INVALID_PARAMETER Count is NULL.
// @retval EFI_DEVICE_ERROR      The device is not functioning properly.
func (p *EFI_BOOT_SERVICES) GetNextMonotonicCount(Count *uint64) EFI_STATUS {
	return UefiCall1(p.getNextMonotonicCount, uintptr(unsafe.Pointer(Count)))
}

// Stall
// Induces a fine-grained stall.
// @param[in]  Microseconds      The number of microseconds to stall execution.
// @retval EFI_SUCCESS           Execution was stalled at least the requested number of
// ..............................Microseconds.
func (p *EFI_BOOT_SERVICES) Stall(Microseconds UINTN) EFI_STATUS {
	return UefiCall1(p.stall, uintptr(Microseconds))
}

// SetWatchdogTimer
// Sets the system's watchdog timer.
// @param[in]  Timeout           The number of seconds to set the watchdog timer to.
// @param[in]  WatchdogCode      The numeric code to log on a watchdog timer timeout event.
// @param[in]  DataSize          The size, in bytes, of WatchdogData.
// @param[in]  WatchdogData      A data buffer that includes a Null-terminated string, optionally
// ..............................followed by additional binary data.
// @retval EFI_SUCCESS           The timeout has been set.
// @retval EFI_INVALID_PARAMETER The supplied WatchdogCode is invalid.
// @retval EFI_UNSUPPORTED       The system does not have a watchdog timer.
// @retval EFI_DEVICE_ERROR      The watchdog timer could not be programmed due to a hardware
// ..............................error.
func (p *EFI_BOOT_SERVICES) SetWatchdogTimer(Timeout UINTN, WatchdogCode uint64, DataSize UINTN, WatchdogData *CHAR16) EFI_STATUS {
	return UefiCall4(p.setWatchdogTimer, uintptr(Timeout), uintptr(WatchdogCode), uintptr(DataSize), uintptr(unsafe.Pointer(WatchdogData)))
}

// ConnectController
// Connects one or more drivers to a controller.
// @param[in]  ControllerHandle      The handle of the controller to which driver(s) are to be connected.
// @param[in]  DriverImageHandle     A pointer to an ordered list handles that support the
// ..................................EFI_DRIVER_BINDING_PROTOCOL.
// @param[in]  RemainingDevicePath   A pointer to the device path that specifies a child of the
// ..................................controller specified by ControllerHandle.
// @param[in]  Recursive             If TRUE, then ConnectController() is called recursively
// ..................................until the entire tree of controllers below the controller specified
// ..................................by ControllerHandle have been created. If FALSE, then
// ..................................the tree of controllers is only expanded one level.
// @retval EFI_SUCCESS           1) One or more drivers were connected to ControllerHandle.
// ..............................2) No drivers were connected to ControllerHandle, but
// ..............................RemainingDevicePath is not NULL, and it is an End Device
// ..............................Path Node.
// @retval EFI_INVALID_PARAMETER ControllerHandle is NULL.
// @retval EFI_NOT_FOUND         1) There are no EFI_DRIVER_BINDING_PROTOCOL instances
// ..............................present in the system.
// ..............................2) No drivers were connected to ControllerHandle.
// ..............................@retval EFI_SECURITY_VIOLATION
// ..............................The user has no permission to start UEFI device drivers on the device path
// ..............................associated with the ControllerHandle or specified by the RemainingDevicePath.
func (p *EFI_BOOT_SERVICES) ConnectController(ControllerHandle EFI_HANDLE, DriverImageHandle *EFI_HANDLE, RemainingDevicePath *EFI_DEVICE_PATH_PROTOCOL, Recursive BOOLEAN) EFI_STATUS {
	return UefiCall4(p.connectController, uintptr(ControllerHandle), uintptr(unsafe.Pointer(DriverImageHandle)), uintptr(unsafe.Pointer(RemainingDevicePath)), convertBoolean(Recursive))
}

// DisconnectController
// Disconnects one or more drivers from a controller.
// @param[in]  ControllerHandle      The handle of the controller from which driver(s) are to be disconnected.
// @param[in]  DriverImageHandle     The driver to disconnect from ControllerHandle.
// ..................................If DriverImageHandle is NULL, then all the drivers currently managing
// ..................................ControllerHandle are disconnected from ControllerHandle.
// @param[in]  ChildHandle           The handle of the child to destroy.
// ..................................If ChildHandle is NULL, then all the children of ControllerHandle are
// ..................................destroyed before the drivers are disconnected from ControllerHandle.
// @retval EFI_SUCCESS           1) One or more drivers were disconnected from the controller.
// ..............................2) On entry, no drivers are managing ControllerHandle.
// ..............................3) DriverImageHandle is not NULL, and on entry
// ..............................DriverImageHandle is not managing ControllerHandle.
// @retval EFI_INVALID_PARAMETER 1) ControllerHandle is NULL.
// ..............................2) DriverImageHandle is not NULL, and it is not a valid EFI_HANDLE.
// ..............................3) ChildHandle is not NULL, and it is not a valid EFI_HANDLE.
// ..............................4) DriverImageHandle does not support the EFI_DRIVER_BINDING_PROTOCOL.
// @retval EFI_OUT_OF_RESOURCES  There are not enough resources available to disconnect any drivers from
// ..............................ControllerHandle.
// @retval EFI_DEVICE_ERROR      The controller could not be disconnected because of a device error.
func (p *EFI_BOOT_SERVICES) DisconnectController(ControllerHandle EFI_HANDLE, DriverImageHandle EFI_HANDLE, ChildHandle EFI_HANDLE) EFI_STATUS {
	return UefiCall3(p.disconnectController, uintptr(ControllerHandle), uintptr(DriverImageHandle), uintptr(ChildHandle))
}

// OpenProtocol
// Queries a handle to determine if it supports a specified protocol. If the protocol is supported by the
// handle, it opens the protocol on behalf of the calling agent.
// @param[in]   Handle           The handle for the protocol interface that is being opened.
// @param[in]   Protocol         The published unique identifier of the protocol.
// @param[out]  Interface        Supplies the address where a pointer to the corresponding Protocol
// ..............................Interface is returned.
// @param[in]   AgentHandle      The handle of the agent that is opening the protocol interface
// ..............................specified by Protocol and Interface.
// @param[in]   ControllerHandle If the agent that is opening a protocol is a driver that follows the
// ..............................UEFI Driver Model, then this parameter is the controller handle
// ..............................that requires the protocol interface. If the agent does not follow
// ..............................the UEFI Driver Model, then this parameter is optional and may
// ..............................be NULL.
// @param[in]   Attributes       The open mode of the protocol interface specified by Handle
// ..............................and Protocol.
// @retval EFI_SUCCESS           An item was added to the open list for the protocol interface, and the
// ..............................protocol interface was returned in Interface.
// @retval EFI_UNSUPPORTED       Handle does not support Protocol.
// @retval EFI_INVALID_PARAMETER One or more parameters are invalid.
// @retval EFI_ACCESS_DENIED     Required attributes can't be supported in current environment.
// @retval EFI_ALREADY_STARTED   Item on the open list already has requierd attributes whose agent
// ..............................handle is the same as AgentHandle.
func (p *EFI_BOOT_SERVICES) OpenProtocol(Handle EFI_HANDLE, Protocol *EFI_GUID, Interface **VOID, AgentHandle EFI_HANDLE, ControllerHandle EFI_HANDLE, Attributes uint32) EFI_STATUS {
	return UefiCall6(p.openProtocol, uintptr(Handle), uintptr(unsafe.Pointer(Protocol)), uintptr(unsafe.Pointer(Interface)), uintptr(AgentHandle), uintptr(ControllerHandle), uintptr(Attributes))
}

// CloseProtocol
// Closes a protocol on a handle that was opened using OpenProtocol().
// @param[in]  Handle            The handle for the protocol interface that was previously opened
// ..............................with OpenProtocol(), and is now being closed.
// @param[in]  Protocol          The published unique identifier of the protocol.
// @param[in]  AgentHandle       The handle of the agent that is closing the protocol interface.
// @param[in]  ControllerHandle  If the agent that opened a protocol is a driver that follows the
// ..............................UEFI Driver Model, then this parameter is the controller handle
// ..............................that required the protocol interface.
// @retval EFI_SUCCESS           The protocol instance was closed.
// @retval EFI_INVALID_PARAMETER 1) Handle is NULL.
// ..............................2) AgentHandle is NULL.
// ..............................3) ControllerHandle is not NULL and ControllerHandle is not a valid EFI_HANDLE.
// ..............................4) Protocol is NULL.
// @retval EFI_NOT_FOUND         1) Handle does not support the protocol specified by Protocol.
// ..............................2) The protocol interface specified by Handle and Protocol is not
// ..............................currently open by AgentHandle and ControllerHandle.
func (p *EFI_BOOT_SERVICES) CloseProtocol(Handle EFI_HANDLE, Protocol *EFI_GUID, AgentHandle EFI_HANDLE, ControllerHandle EFI_HANDLE) EFI_STATUS {
	return UefiCall4(p.closeProtocol, uintptr(Handle), uintptr(unsafe.Pointer(Protocol)), uintptr(AgentHandle), uintptr(ControllerHandle))
}

// OpenProtocolInformation
// Retrieves the list of agents that currently have a protocol interface opened.
// @param[in]   Handle           The handle for the protocol interface that is being queried.
// @param[in]   Protocol         The published unique identifier of the protocol.
// @param[out]  EntryBuffer      A pointer to a buffer of open protocol information in the form of
// ..............................EFI_OPEN_PROTOCOL_INFORMATION_ENTRY structures.
// @param[out]  EntryCount       A pointer to the number of entries in EntryBuffer.
// @retval EFI_SUCCESS           The open protocol information was returned in EntryBuffer, and the
// ..............................number of entries was returned EntryCount.
// @retval EFI_OUT_OF_RESOURCES  There are not enough resources available to allocate EntryBuffer.
// @retval EFI_NOT_FOUND         Handle does not support the protocol specified by Protocol.
func (p *EFI_BOOT_SERVICES) OpenProtocolInformation(Handle EFI_HANDLE, Protocol *EFI_GUID, EntryBuffer **EFI_OPEN_PROTOCOL_INFORMATION_ENTRY, EntryCount *UINTN) EFI_STATUS {
	return UefiCall4(p.openProtocolInformation, uintptr(Handle), uintptr(unsafe.Pointer(Protocol)), uintptr(unsafe.Pointer(EntryBuffer)), uintptr(unsafe.Pointer(EntryCount)))
}

// ProtocolsPerHandle
// Retrieves the list of protocol interface GUIDs that are installed on a handle in a buffer allocated
// from pool.
// @param[in]   Handle              The handle from which to retrieve the list of protocol interface
// .................................GUIDs.
// @param[out]  ProtocolBuffer      A pointer to the list of protocol interface GUID pointers that are
// .................................installed on Handle.
// @param[out]  ProtocolBufferCount A pointer to the number of GUID pointers present in
// .................................ProtocolBuffer.
// @retval EFI_SUCCESS           The list of protocol interface GUIDs installed on Handle was returned in
// ..............................ProtocolBuffer. The number of protocol interface GUIDs was
// ..............................returned in ProtocolBufferCount.
// @retval EFI_OUT_OF_RESOURCES  There is not enough pool memory to store the results.
// @retval EFI_INVALID_PARAMETER Handle is NULL.
// @retval EFI_INVALID_PARAMETER Handle is not a valid EFI_HANDLE.
// @retval EFI_INVALID_PARAMETER ProtocolBuffer is NULL.
// @retval EFI_INVALID_PARAMETER ProtocolBufferCount is NULL.
func (p *EFI_BOOT_SERVICES) ProtocolsPerHandle(Handle EFI_HANDLE, ProtocolBuffer ***EFI_GUID, ProtocolBufferCount *UINTN) EFI_STATUS {
	return UefiCall3(p.protocolsPerHandle, uintptr(Handle), uintptr(unsafe.Pointer(ProtocolBuffer)), uintptr(unsafe.Pointer(ProtocolBufferCount)))
}

// LocateHandleBuffer
// Returns an array of handles that support the requested protocol in a buffer allocated from pool.
// @param[in]       SearchType   Specifies which handle(s) are to be returned.
// @param[in]       Protocol     Provides the protocol to search by.
// ..............................This parameter is only valid for a SearchType of ByProtocol.
// @param[in]       SearchKey    Supplies the search key depending on the SearchType.
// @param[out]      NoHandles    The number of handles returned in Buffer.
// @param[out]      Buffer       A pointer to the buffer to return the requested array of handles that
// ..............................support Protocol.
// @retval EFI_SUCCESS           The array of handles was returned in Buffer, and the number of
// ..............................handles in Buffer was returned in NoHandles.
// @retval EFI_NOT_FOUND         No handles match the search.
// @retval EFI_OUT_OF_RESOURCES  There is not enough pool memory to store the matching results.
// @retval EFI_INVALID_PARAMETER NoHandles is NULL.
// @retval EFI_INVALID_PARAMETER Buffer is NULL.
func (p *EFI_BOOT_SERVICES) LocateHandleBuffer(SearchType EFI_LOCATE_SEARCH_TYPE, Protocol *EFI_GUID, SearchKey *VOID, NoHandles *UINTN, Buffer **EFI_HANDLE) EFI_STATUS {
	return UefiCall5(p.locateHandleBuffer, uintptr(SearchType), uintptr(unsafe.Pointer(Protocol)), uintptr(unsafe.Pointer(SearchKey)), uintptr(unsafe.Pointer(NoHandles)), uintptr(unsafe.Pointer(Buffer)))
}

// LocateProtocol
// Returns the first protocol instance that matches the given protocol.
// @param[in]  Protocol          Provides the protocol to search for.
// @param[in]  Registration      Optional registration key returned from
// ..............................RegisterProtocolNotify().
// @param[out]  Interface        On return, a pointer to the first interface that matches Protocol and
// ..............................Registration.
// @retval EFI_SUCCESS           A protocol instance matching Protocol was found and returned in
// ..............................Interface.
// @retval EFI_NOT_FOUND         No protocol instances were found that match Protocol and
// ..............................Registration.
// @retval EFI_INVALID_PARAMETER Interface is NULL.
// ..............................Protocol is NULL.
func (p *EFI_BOOT_SERVICES) LocateProtocol(Protocol *EFI_GUID, Registration *VOID, InterfacePtr unsafe.Pointer) EFI_STATUS {
	return UefiCall3(p.locateProtocol, uintptr(unsafe.Pointer(Protocol)), uintptr(unsafe.Pointer(Registration)), uintptr(InterfacePtr))
}

// InstallMultipleProtocolInterfaces
// Installs one or more protocol interfaces into the boot services environment.
// @param[in, out]  Handle       The pointer to a handle to install the new protocol interfaces on,
// .................or a pointer to NULL if a new handle is to be allocated.
// @param  ...                   A variable argument list containing pairs of protocol GUIDs and protocol
// ..............................interfaces.
// @retval EFI_SUCCESS           All the protocol interface was installed.
// @retval EFI_OUT_OF_RESOURCES  There was not enough memory in pool to install all the protocols.
// @retval EFI_ALREADY_STARTED   A Device Path Protocol instance was passed in that is already present in
// ..............................the handle database.
// @retval EFI_INVALID_PARAMETER Handle is NULL.
// @retval EFI_INVALID_PARAMETER Protocol is already installed on the handle specified by Handle.
func (p *EFI_BOOT_SERVICES) InstallMultipleProtocolInterfaces(Handle *EFI_HANDLE) EFI_STATUS {
	return UefiCall1(p.installMultipleProtocolInterfaces, uintptr(unsafe.Pointer(Handle)))
}

// UninstallMultipleProtocolInterfaces
// Removes one or more protocol interfaces into the boot services environment.
// @param[in]  Handle            The handle to remove the protocol interfaces from.
// @param  ...                   A variable argument list containing pairs of protocol GUIDs and
// ..............................protocol interfaces.
// @retval EFI_SUCCESS           All the protocol interfaces were removed.
// @retval EFI_INVALID_PARAMETER One of the protocol interfaces was not previously installed on Handle.
func (p *EFI_BOOT_SERVICES) UninstallMultipleProtocolInterfaces(Handle EFI_HANDLE) EFI_STATUS {
	return UefiCall1(p.uninstallMultipleProtocolInterfaces, uintptr(Handle))
}

// CalculateCrc32
// Computes and returns a 32-bit CRC for a data buffer.
// @param[in]   Data             A pointer to the buffer on which the 32-bit CRC is to be computed.
// @param[in]   DataSize         The number of bytes in the buffer Data.
// @param[out]  Crc32            The 32-bit CRC that was computed for the data buffer specified by Data
// ..............................and DataSize.
// @retval EFI_SUCCESS           The 32-bit CRC was computed for the data buffer and returned in
// ..............................Crc32.
// @retval EFI_INVALID_PARAMETER Data is NULL.
// @retval EFI_INVALID_PARAMETER Crc32 is NULL.
// @retval EFI_INVALID_PARAMETER DataSize is 0.
func (p *EFI_BOOT_SERVICES) CalculateCrc32(Data *VOID, DataSize UINTN, Crc32 *uint32) EFI_STATUS {
	return UefiCall3(p.calculateCrc32, uintptr(unsafe.Pointer(Data)), uintptr(DataSize), uintptr(unsafe.Pointer(Crc32)))
}

// CopyMem
// Copies the contents of one buffer to another buffer.
// @param[in]  Destination       The pointer to the destination buffer of the memory copy.
// @param[in]  Source            The pointer to the source buffer of the memory copy.
// @param[in]  Length            Number of bytes to copy from Source to Destination.
func (p *EFI_BOOT_SERVICES) CopyMem(Destination *VOID, Source *VOID, Length UINTN) EFI_STATUS {
	return UefiCall3(p.copyMem, uintptr(unsafe.Pointer(Destination)), uintptr(unsafe.Pointer(Source)), uintptr(Length))
}

// SetMem
// The SetMem() function fills a buffer with a specified value.
// @param[in]  Buffer            The pointer to the buffer to fill.
// @param[in]  Size              Number of bytes in Buffer to fill.
// @param[in]  Value             Value to fill Buffer with.
func (p *EFI_BOOT_SERVICES) SetMem(Buffer *VOID, Size UINTN, Value byte) EFI_STATUS {
	return UefiCall3(p.setMem, uintptr(unsafe.Pointer(Buffer)), uintptr(Size), uintptr(Value))
}

// CreateEventEx
// Creates an event in a group.
// @param[in]   Type             The type of event to create and its mode and attributes.
// @param[in]   NotifyTpl        The task priority level of event notifications,if needed.
// @param[in]   NotifyFunction   The pointer to the event's notification function, if any.
// @param[in]   NotifyContext    The pointer to the notification function's context; corresponds to parameter
// ..............................Context in the notification function.
// @param[in]   EventGroup       The pointer to the unique identifier of the group to which this event belongs.
// ..............................If this is NULL, then the function behaves as if the parameters were passed
// ..............................to CreateEvent.
// @param[out]  Event            The pointer to the newly created event if the call succeeds; undefined
// ..............................otherwise.
// @retval EFI_SUCCESS           The event structure was created.
// @retval EFI_INVALID_PARAMETER One or more parameters are invalid.
// @retval EFI_OUT_OF_RESOURCES  The event could not be allocated.
func (p *EFI_BOOT_SERVICES) CreateEventEx(Type uint32, NotifyTpl EFI_TPL, NotifyFunction unsafe.Pointer, NotifyContext uintptr, EventGroup *EFI_GUID, Event *EFI_EVENT) EFI_STATUS {
	return UefiCall6(p.createEventEx, uintptr(Type), uintptr(NotifyTpl), uintptr(NotifyFunction), NotifyContext, uintptr(unsafe.Pointer(EventGroup)), uintptr(unsafe.Pointer(Event)))
}

// endregion

// EFI_GUID definitions
var (
	MPS_TABLE_GUID = EFI_GUID{
		Data1: 0xeb9d2d2f,
		Data2: 0x2d88,
		Data3: 0x11d3,
		Data4: [8]uint8{0x9a, 0x16, 0x00, 0x90, 0x27, 0x3f, 0xc1, 0x4d},
	}

	ACPI_TABLE_GUID = EFI_GUID{
		Data1: 0xeb9d2d30,
		Data2: 0x2d88,
		Data3: 0x11d3,
		Data4: [8]uint8{0x9a, 0x16, 0x00, 0x90, 0x27, 0x3f, 0xc1, 0x4d},
	}

	ACPI_20_TABLE_GUID = EFI_GUID{
		Data1: 0x8868e871,
		Data2: 0xe4f1,
		Data3: 0x11d3,
		Data4: [8]uint8{0xbc, 0x22, 0x00, 0x80, 0xc7, 0x3c, 0x88, 0x81},
	}

	SMBIOS_TABLE_GUID = EFI_GUID{
		Data1: 0xeb9d2d31,
		Data2: 0x2d88,
		Data3: 0x11d3,
		Data4: [8]uint8{0x9a, 0x16, 0x00, 0x90, 0x27, 0x3f, 0xc1, 0x4d},
	}

	SMBIOS3_TABLE_GUID = EFI_GUID{
		Data1: 0xf2fd1544,
		Data2: 0x9794,
		Data3: 0x4a2c,
		Data4: [8]uint8{0x99, 0x2e, 0xe5, 0xbb, 0xcf, 0x20, 0xe3, 0x94},
	}

	SAL_SYSTEM_TABLE_GUID = EFI_GUID{
		Data1: 0xeb9d2d32,
		Data2: 0x2d88,
		Data3: 0x11d3,
		Data4: [8]uint8{0x9a, 0x16, 0x00, 0x90, 0x27, 0x3f, 0xc1, 0x4d},
	}

	EFI_DTB_TABLE_GUID = EFI_GUID{
		Data1: 0xb1b621d5,
		Data2: 0xf19c,
		Data3: 0x41a5,
		Data4: [8]uint8{0x83, 0x0b, 0xd9, 0x15, 0x2c, 0x69, 0xaa, 0xe0},
	}
)

// EFI_CONFIGURATION_TABLE
// Contains a set of GUID/pointer pairs comprised of the ConfigurationTable field in the
// EFI System Table.
type EFI_CONFIGURATION_TABLE struct {
	VendorGuid  EFI_GUID
	VendorTable *VOID
}

// region: EFI System Table

// EFI_SYSTEM_TABLE_SIGNATURE is the signature for the EFI system table.
const EFI_SYSTEM_TABLE_SIGNATURE = 0x5453595320494249

var (
	EFI1_02_SYSTEM_TABLE_REVISION  = EFI_SPECIFICATION_REVISION_MAJORMINOR(1, 02)
	EFI1_10_SYSTEM_TABLE_REVISION  = EFI_SPECIFICATION_REVISION_MAJORMINOR(1, 10)
	EFI2_00_SYSTEM_TABLE_REVISION  = EFI_SPECIFICATION_REVISION_MAJORMINOR(2, 00)
	EFI2_10_SYSTEM_TABLE_REVISION  = EFI_SPECIFICATION_REVISION_MAJORMINOR(2, 10)
	EFI2_20_SYSTEM_TABLE_REVISION  = EFI_SPECIFICATION_REVISION_MAJORMINOR(2, 20)
	EFI2_30_SYSTEM_TABLE_REVISION  = EFI_SPECIFICATION_REVISION_MAJORMINOR(2, 30)
	EFI2_31_SYSTEM_TABLE_REVISION  = EFI_SPECIFICATION_REVISION_MAJORMINOR(2, 31)
	EFI2_40_SYSTEM_TABLE_REVISION  = EFI_SPECIFICATION_REVISION_MAJORMINOR(2, 40)
	EFI2_50_SYSTEM_TABLE_REVISION  = EFI_SPECIFICATION_REVISION_MAJORMINOR(2, 50)
	EFI2_60_SYSTEM_TABLE_REVISION  = EFI_SPECIFICATION_REVISION_MAJORMINOR(2, 60)
	EFI2_70_SYSTEM_TABLE_REVISION  = EFI_SPECIFICATION_REVISION_MAJORMINOR(2, 70)
	EFI2_80_SYSTEM_TABLE_REVISION  = EFI_SPECIFICATION_REVISION_MAJORMINOR(2, 80)
	EFI2_90_SYSTEM_TABLE_REVISION  = EFI_SPECIFICATION_REVISION_MAJORMINOR(2, 90)
	EFI2_100_SYSTEM_TABLE_REVISION = EFI_SPECIFICATION_REVISION_MAJORMINOR(2, 100)
	EFISYSTEM_TABLE_REVISION       = EFI_SPECIFICATION_VERSION
)

// EFI_SYSTEM_TABLE
// EFI System Table
type EFI_SYSTEM_TABLE struct {
	Hdr                  EFI_TABLE_HEADER
	FirmwareVendor       *CHAR16
	FirmwareRevision     uint32
	ConsoleInHandle      EFI_HANDLE
	ConIn                *EFI_SIMPLE_TEXT_INPUT_PROTOCOL
	ConsoleOutHandle     EFI_HANDLE
	ConOut               *EFI_SIMPLE_TEXT_OUTPUT_PROTOCOL
	StandardErrorHandle  EFI_HANDLE
	StdErr               *EFI_SIMPLE_TEXT_OUTPUT_PROTOCOL
	RuntimeServices      *EFI_RUNTIME_SERVICES
	BootServices         *EFI_BOOT_SERVICES
	NumberOfTableEntries UINTN
	ConfigurationTable   *EFI_CONFIGURATION_TABLE
}

// endregion

type EFI_LOAD_OPTION struct {
	Attributes         uint32
	FilePathListLength uint16
}

//type EFI_KEY_OPTION struct {
//	KeyData       EFI_BOOT_KEY_DATA
//	BootOptionCrc uint32
//	BootOption    uint16
//}
