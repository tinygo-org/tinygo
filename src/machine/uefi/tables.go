//go:build uefi

package uefi

// SystemTable is the main UEFI system table.
type SystemTable struct {
	Hdr                  TableHeader
	FirmwareVendor       *uint16
	FirmwareRevision     uint32
	ConsoleInHandle      Handle
	ConIn                *SimpleTextInputProtocol
	ConsoleOutHandle     Handle
	ConOut               *SimpleTextOutputProtocol
	StandardErrorHandle  Handle
	StdErr               *SimpleTextOutputProtocol
	RuntimeServices      *RuntimeServices
	BootServices         *BootServices
	NumberOfTableEntries uintptr
	ConfigurationTable   uintptr
}

// BootServices provides boot-time services.
type BootServices struct {
	Hdr TableHeader

	// Task Priority Services
	RaiseTPL   uintptr
	RestoreTPL uintptr

	// Memory Services
	AllocatePages uintptr
	FreePages     uintptr
	GetMemoryMap  uintptr
	AllocatePool  uintptr
	FreePool      uintptr

	// Event & Timer Services
	CreateEvent  uintptr
	SetTimer     uintptr
	WaitForEvent uintptr
	SignalEvent  uintptr
	CloseEvent   uintptr
	CheckEvent   uintptr

	// Protocol Handler Services
	InstallProtocolInterface   uintptr
	ReinstallProtocolInterface uintptr
	UninstallProtocolInterface uintptr
	HandleProtocol             uintptr
	Reserved                   uintptr
	RegisterProtocolNotify     uintptr
	LocateHandle               uintptr
	LocateDevicePath           uintptr
	InstallConfigurationTable  uintptr

	// Image Services
	LoadImage        uintptr
	StartImage       uintptr
	Exit             uintptr
	UnloadImage      uintptr
	ExitBootServices uintptr

	// Miscellaneous Services
	GetNextMonotonicCount uintptr
	Stall                 uintptr
	SetWatchdogTimer      uintptr

	// Driver Support Services
	ConnectController    uintptr
	DisconnectController uintptr

	// Open and Close Protocol Services
	OpenProtocol            uintptr
	CloseProtocol           uintptr
	OpenProtocolInformation uintptr

	// Library Services
	ProtocolsPerHandle                  uintptr
	LocateHandleBuffer                  uintptr
	LocateProtocol                      uintptr
	InstallMultipleProtocolInterfaces   uintptr
	UninstallMultipleProtocolInterfaces uintptr

	// 32-bit CRC Services
	CalculateCrc32 uintptr

	// Miscellaneous Services (cont.)
	CopyMem       uintptr
	SetMem        uintptr
	CreateEventEx uintptr
}

// RuntimeServices provides runtime services.
type RuntimeServices struct {
	Hdr TableHeader

	// Time Services
	GetTime       uintptr
	SetTime       uintptr
	GetWakeupTime uintptr
	SetWakeupTime uintptr

	// Virtual Memory Services
	SetVirtualAddressMap uintptr
	ConvertPointer       uintptr

	// Variable Services
	GetVariable         uintptr
	GetNextVariableName uintptr
	SetVariable         uintptr

	// Miscellaneous Services
	GetNextHighMonotonicCount uintptr
	ResetSystem               uintptr

	// UEFI 2.0 Capsule Services
	UpdateCapsule            uintptr
	QueryCapsuleCapabilities uintptr

	// Miscellaneous UEFI 2.0 Services
	QueryVariableInfo uintptr
}

// Global pointers set by assembly entry point
//
//go:extern efi_image_handle
var imageHandle Handle

//go:extern efi_system_table
var systemTablePtr *SystemTable

// ImageHandle returns the UEFI image handle for this application.
func ImageHandle() Handle {
	return imageHandle
}

// GetSystemTable returns the UEFI system table pointer.
func GetSystemTable() *SystemTable {
	return systemTablePtr
}
