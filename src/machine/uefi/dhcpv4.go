package uefi

import (
	"unsafe"
)

//---------------------------------------------------------------------------
// GUIDs
//---------------------------------------------------------------------------

// EFI_DHCP4_SERVICE_BINDING_PROTOCOL_GUID — §29.2.1
var EFI_DHCP4_SERVICE_BINDING_PROTOCOL_GUID = EFI_GUID{
	0x9d9a39d8, 0xbd42, 0x4a73,
	[8]uint8{0xa4, 0xd5, 0x8e, 0xe9, 0x4b, 0xe1, 0x13, 0x80},
}

// EFI_DHCP4_PROTOCOL_GUID — §29.2.2
var EFI_DHCP4_PROTOCOL_GUID = EFI_GUID{
	0x8a219718, 0x4ef5, 0x4761,
	[8]uint8{0x91, 0xc8, 0xc0, 0xf0, 0x4b, 0xda, 0x9e, 0x56},
}

//---------------------------------------------------------------------------
// Enums / States
//---------------------------------------------------------------------------

// EFI_DHCP4_STATE — DHCP state machine §29.2.3
const (
	Dhcp4Stopped = iota
	Dhcp4Init
	Dhcp4Selecting
	Dhcp4Requesting
	Dhcp4Bound
	Dhcp4Renewing
	Dhcp4Rebinding
	Dhcp4InitReboot
	Dhcp4Rebooting
)

// EFI_DHCP4_EVENT tracks events in the DHCP process. §29.2.4
const (
	Dhcp4SendDiscover = iota + 1
	Dhcp4RcvdOffer
	Dhcp4SelectOffer
	Dhcp4SendRequest
	Dhcp4RcvdAck
	Dhcp4RcvdNak
	Dhcp4SendDecline
	Dhcp4BoundCompleted
	Dhcp4EnterRenewing
	Dhcp4EnterRebinding
	Dhcp4AddressLost
	Dhcp4Fail
)

// EFI_DHCP4_PACKET_OPTION 29.2.4
type EFI_DHCP4_PACKET_OPTION struct {
	OpCode uint8
	Length uint8
	Data   [1]uint8
}

//---------------------------------------------------------------------------
// EFI_DHCP4_SERVICE_BINDING_PROTOCOL
//---------------------------------------------------------------------------

type EFI_DHCP4_SERVICE_BINDING_PROTOCOL struct {
	createChild  uintptr // (*this, *childHandle)
	destroyChild uintptr // (*this, childHandle)
}

func (p *EFI_DHCP4_SERVICE_BINDING_PROTOCOL) CreateChild(childHandle *EFI_HANDLE) EFI_STATUS {
	return UefiCall2(
		p.createChild,
		uintptr(unsafe.Pointer(p)),
		uintptr(unsafe.Pointer(childHandle)),
	)
}

func (p *EFI_DHCP4_SERVICE_BINDING_PROTOCOL) DestroyChild(childHandle EFI_HANDLE) EFI_STATUS {
	return UefiCall2(
		p.destroyChild,
		uintptr(unsafe.Pointer(p)),
		uintptr(childHandle),
	)
}

//---------------------------------------------------------------------------
// EFI_DHCP4_PROTOCOL
//---------------------------------------------------------------------------

// EFI_DHCP4_CONFIG_DATA is for configuring a DHCP request. §29.2.4
type EFI_DHCP4_CONFIG_DATA struct {
	DiscoverTryCount uint32
	DiscoverTimeout  *uint32
	RequestTryCount  uint32
	RequestTimeout   *uint32
	ClientAddress    EFI_IPv4_ADDRESS
	Dhcp4Callback    EFI_DHCP4_CALLBACK
	CallbackContext  unsafe.Pointer // not sure about this
	OptionCount      uint32
	OptionList       **EFI_DHCP4_PACKET_OPTION
}

// EFI_DHCP4_CALLBACK is not yet supported. Needs a PE+ -> SysV trampoline.
type EFI_DHCP4_CALLBACK uintptr

type EFI_DHCP4_MODE_DATA struct {
	State         uint32
	ConfigData    EFI_DHCP4_CONFIG_DATA
	ClientAddress EFI_IPv4_ADDRESS
	ClientMac     EFI_MAC_ADDRESS
	ServerAddress EFI_IPv4_ADDRESS
	RouterAddress EFI_IPv4_ADDRESS
	SubnetMask    EFI_IPv4_ADDRESS
}

// EFI_DHCP4_PROTOCOL function table §29.2.2
type EFI_DHCP4_PROTOCOL struct {
	getModeData     uintptr
	configure       uintptr
	start           uintptr
	renewRebind     uintptr
	release         uintptr
	stop            uintptr
	build           uintptr
	transmitReceive uintptr
	parse           uintptr
}

func (p *EFI_DHCP4_PROTOCOL) GetModeData(modeData *EFI_DHCP4_MODE_DATA) EFI_STATUS {
	return UefiCall2(
		p.getModeData,
		uintptr(unsafe.Pointer(p)),
		uintptr(unsafe.Pointer(modeData)),
	)
}

func (p *EFI_DHCP4_PROTOCOL) Configure(cfg *EFI_DHCP4_CONFIG_DATA) EFI_STATUS {
	return UefiCall2(
		p.configure,
		uintptr(unsafe.Pointer(p)),
		uintptr(unsafe.Pointer(cfg)),
	)
}

func (p *EFI_DHCP4_PROTOCOL) Start(asyncEvent *EFI_EVENT) EFI_STATUS {
	return UefiCall2(
		p.start,
		uintptr(unsafe.Pointer(p)),
		uintptr(unsafe.Pointer(asyncEvent)),
	)
}

func (p *EFI_DHCP4_PROTOCOL) RenewRebind(asyncEvent *EFI_EVENT) EFI_STATUS {
	return UefiCall2(p.renewRebind, uintptr(unsafe.Pointer(p)), uintptr(unsafe.Pointer(asyncEvent)))
}

func (p *EFI_DHCP4_PROTOCOL) Release(asyncEvent *EFI_EVENT) EFI_STATUS {
	return UefiCall2(p.release, uintptr(unsafe.Pointer(p)), uintptr(unsafe.Pointer(asyncEvent)))
}

func (p *EFI_DHCP4_PROTOCOL) Stop() EFI_STATUS {
	return UefiCall1(p.stop, uintptr(unsafe.Pointer(p)))
}

func (p *EFI_DHCP4_PROTOCOL) Build(packetBuffer unsafe.Pointer) EFI_STATUS {
	return UefiCall2(p.build, uintptr(unsafe.Pointer(p)), uintptr(packetBuffer))
}

func (p *EFI_DHCP4_PROTOCOL) TransmitReceive(token unsafe.Pointer) EFI_STATUS {
	return UefiCall2(p.transmitReceive, uintptr(unsafe.Pointer(p)), uintptr(token))
}

func (p *EFI_DHCP4_PROTOCOL) Parse(packetBuffer unsafe.Pointer, parseResult unsafe.Pointer) EFI_STATUS {
	return UefiCall3(p.parse, uintptr(unsafe.Pointer(p)), uintptr(packetBuffer), uintptr(parseResult))
}

// DHCPv4 wraps EFI_DHCP4_PROTOCOL to provide a more go idiomatic API for handling DHCPv4.
type DHCPv4 struct {
	*EFI_DHCP4_PROTOCOL
}

func EnumerateDHCPv4() ([]*DHCPv4, error) {
	var (
		handleCount  UINTN
		handleBuffer *EFI_HANDLE
	)
	status := BS().LocateHandleBuffer(ByProtocol, &EFI_DHCP4_SERVICE_BINDING_PROTOCOL_GUID, nil, &handleCount, &handleBuffer)
	if status != EFI_SUCCESS {
		return nil, StatusError(status)
	}
	// if none were found, we should have gotten EFI_NOT_FOUND

	//turn handleBuffer into a slice of EFI_HANDLEs
	handleSlice := unsafe.Slice((*EFI_HANDLE)(unsafe.Pointer(handleBuffer)), int(handleCount))

	dhcpv4s := make([]*DHCPv4, 0, int(handleCount))

	// Turn Binding handles into Protocol
	for i := range handleSlice {
		var binding *EFI_DHCP4_SERVICE_BINDING_PROTOCOL
		status := BS().HandleProtocol(
			handleSlice[i],
			&EFI_DHCP4_SERVICE_BINDING_PROTOCOL_GUID,
			unsafe.Pointer(&binding),
		)
		if status != EFI_SUCCESS {
			// just skip, or error out entirely?? Hmm..
			continue
		}
		var bindChild EFI_HANDLE
		// TODO: track and clean up after the children
		status = binding.CreateChild(&bindChild)
		if status != EFI_SUCCESS {
			continue
		}
		var dhcpp *EFI_DHCP4_PROTOCOL
		status = BS().HandleProtocol(
			bindChild,
			&EFI_DHCP4_PROTOCOL_GUID,
			unsafe.Pointer(&dhcpp),
		)
		if status != EFI_SUCCESS {
			continue
		}
		dhcpv4s = append(dhcpv4s, &DHCPv4{EFI_DHCP4_PROTOCOL: dhcpp})
	}

	return dhcpv4s, nil
}
