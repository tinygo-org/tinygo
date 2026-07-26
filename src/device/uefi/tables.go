package uefi

import "unsafe"

type EFI_RUNTIME_SERVICES struct {
	Hdr                  EFI_TABLE_HEADER
	getTime              uintptr
	setTime              uintptr
	getWakeupTime        uintptr
	setWakeupTime        uintptr
	setVirtualAddressMap uintptr
	convertPointer       uintptr
	getVariable          uintptr
	getNextVariableName  uintptr
	setVariable          uintptr
	getNextHighMonoCount uintptr
	resetSystem          uintptr
	updateCapsule        uintptr
	queryCapsuleCaps     uintptr
	queryVariableInfo    uintptr
}

func (p *EFI_RUNTIME_SERVICES) GetTime(time *EFI_TIME, capabilities *EFI_TIME_CAPABILITIES) EFI_STATUS {
	return UefiCall2(p.getTime, uintptr(unsafe.Pointer(time)), uintptr(unsafe.Pointer(capabilities)))
}

type EFI_BOOT_SERVICES struct {
	Hdr                       EFI_TABLE_HEADER
	raiseTPL                  uintptr
	restoreTPL                uintptr
	allocatePages             uintptr
	freePages                 uintptr
	getMemoryMap              uintptr
	allocatePool              uintptr
	freePool                  uintptr
	createEvent               uintptr
	setTimer                  uintptr
	waitForEvent              uintptr
	signalEvent               uintptr
	closeEvent                uintptr
	checkEvent                uintptr
	installProtocolInterface  uintptr
	reinstallProtocolIFace    uintptr
	uninstallProtocolIFace    uintptr
	handleProtocol            uintptr
	reserved                  *VOID
	registerProtocolNotify    uintptr
	locateHandle              uintptr
	locateDevicePath          uintptr
	installConfigurationTable uintptr
	loadImage                 uintptr
	startImage                uintptr
	exit                      uintptr
	unloadImage               uintptr
	exitBootServices          uintptr
	getNextMonotonicCount     uintptr
	stall                     uintptr
	setWatchdogTimer          uintptr
	connectController         uintptr
	disconnectController      uintptr
	openProtocol              uintptr
	closeProtocol             uintptr
	openProtocolInformation   uintptr
	protocolsPerHandle        uintptr
	locateHandleBuffer        uintptr
	locateProtocol            uintptr
}

func (p *EFI_BOOT_SERVICES) AllocatePages(typ EFI_ALLOCATE_TYPE, memoryType EFI_MEMORY_TYPE, pages UINTN, memory *EFI_PHYSICAL_ADDRESS) EFI_STATUS {
	return UefiCall4(p.allocatePages, uintptr(typ), uintptr(memoryType), uintptr(pages), uintptr(unsafe.Pointer(memory)))
}

func (p *EFI_BOOT_SERVICES) FreePages(memory EFI_PHYSICAL_ADDRESS, pages UINTN) EFI_STATUS {
	return UefiCall2(p.freePages, uintptr(memory), uintptr(pages))
}

func (p *EFI_BOOT_SERVICES) CreateEvent(typ EVENT_TYPE, notifyTPL EFI_TPL, notifyFunction unsafe.Pointer, notifyContext unsafe.Pointer, event *EFI_EVENT) EFI_STATUS {
	return UefiCall5(p.createEvent, uintptr(typ), uintptr(notifyTPL), uintptr(notifyFunction), uintptr(notifyContext), uintptr(unsafe.Pointer(event)))
}

func (p *EFI_BOOT_SERVICES) SetTimer(event EFI_EVENT, typ EFI_TIMER_DELAY, triggerTime uint64) EFI_STATUS {
	return UefiCall3(p.setTimer, uintptr(event), uintptr(typ), uintptr(triggerTime))
}

func (p *EFI_BOOT_SERVICES) WaitForEvent(numberOfEvents UINTN, event *EFI_EVENT, index *UINTN) EFI_STATUS {
	return UefiCall3(p.waitForEvent, uintptr(numberOfEvents), uintptr(unsafe.Pointer(event)), uintptr(unsafe.Pointer(index)))
}

func (p *EFI_BOOT_SERVICES) SignalEvent(event EFI_EVENT) EFI_STATUS {
	return UefiCall1(p.signalEvent, uintptr(event))
}

func (p *EFI_BOOT_SERVICES) CloseEvent(event EFI_EVENT) EFI_STATUS {
	return UefiCall1(p.closeEvent, uintptr(event))
}

func (p *EFI_BOOT_SERVICES) CheckEvent(event EFI_EVENT) EFI_STATUS {
	return UefiCall1(p.checkEvent, uintptr(event))
}

func (p *EFI_BOOT_SERVICES) HandleProtocol(handle EFI_HANDLE, protocol *EFI_GUID, iface unsafe.Pointer) EFI_STATUS {
	return UefiCall3(p.handleProtocol, uintptr(handle), uintptr(unsafe.Pointer(protocol)), uintptr(iface))
}

func (p *EFI_BOOT_SERVICES) LocateProtocol(protocol *EFI_GUID, registration *VOID, iface unsafe.Pointer) EFI_STATUS {
	return UefiCall3(p.locateProtocol, uintptr(unsafe.Pointer(protocol)), uintptr(unsafe.Pointer(registration)), uintptr(iface))
}

func (p *EFI_BOOT_SERVICES) Exit(imageHandle EFI_HANDLE, exitStatus EFI_STATUS, exitDataSize UINTN, exitData *CHAR16) EFI_STATUS {
	return UefiCall4(p.exit, uintptr(imageHandle), uintptr(exitStatus), uintptr(exitDataSize), uintptr(unsafe.Pointer(exitData)))
}

func (p *EFI_BOOT_SERVICES) SetWatchdogTimer(timeout UINTN, watchdogCode uint64, dataSize UINTN, watchdogData *CHAR16) EFI_STATUS {
	return UefiCall4(p.setWatchdogTimer, uintptr(timeout), uintptr(watchdogCode), uintptr(dataSize), uintptr(unsafe.Pointer(watchdogData)))
}

type EFI_SYSTEM_TABLE struct {
	Hdr                  EFI_TABLE_HEADER
	FirmwareVendor       *CHAR16
	FirmwareRevision     uint32
	ConsoleInHandle      EFI_HANDLE
	ConIn                *VOID
	ConsoleOutHandle     EFI_HANDLE
	ConOut               *EFI_SIMPLE_TEXT_OUTPUT_PROTOCOL
	StandardErrorHandle  EFI_HANDLE
	StdErr               *EFI_SIMPLE_TEXT_OUTPUT_PROTOCOL
	RuntimeServices      *EFI_RUNTIME_SERVICES
	BootServices         *EFI_BOOT_SERVICES
	NumberOfTableEntries UINTN
	ConfigurationTable   *VOID
}
