package uefi

import "unsafe"

//---------------------------------------------------------------------------
//  GUID
//---------------------------------------------------------------------------

// {A19832B9-AC25-11D3-9A2D-0090273FC14D}
var EFI_SIMPLE_NETWORK_PROTOCOL_GUID = EFI_GUID{
	0xA19832B9, 0xAC25, 0x11D3,
	[8]uint8{0x9A, 0x2D, 0x00, 0x90, 0x27, 0x3F, 0xC1, 0x4D},
}

//---------------------------------------------------------------------------
//  Enums and Constants – §10.4
//---------------------------------------------------------------------------

const (
	EFI_SIMPLE_NETWORK_STOPPED     = 0
	EFI_SIMPLE_NETWORK_STARTED     = 1
	EFI_SIMPLE_NETWORK_INITIALIZED = 2
)

// ---------------------------------------------------------------------------
//
//	Types – §10.4.3
//
// ---------------------------------------------------------------------------
type EFI_SIMPLE_NETWORK_MODE struct {
	State                 uint32
	HwAddressSize         uint32
	MediaHeaderSize       uint32
	MaxPacketSize         uint32
	NvRamSize             uint32
	NvRamAccessSize       uint32
	ReceiveFilterMask     uint32
	ReceiveFilterSetting  uint32
	MaxMCastFilterCount   uint32
	MCastFilterCount      uint32
	MCastFilter           [32]EFI_MAC_ADDRESS
	CurrentAddress        EFI_MAC_ADDRESS
	BroadcastAddress      EFI_MAC_ADDRESS
	PermanentAddress      EFI_MAC_ADDRESS
	IfType                uint8
	MacAddressChangeable  bool
	MultipleTxSupported   bool
	MediaPresentSupported bool
	MediaPresent          bool
}

//---------------------------------------------------------------------------
//  Protocol – §10.4.2
//---------------------------------------------------------------------------

type EFI_SIMPLE_NETWORK_PROTOCOL struct {
	Revision       uint64
	start          uintptr // (this)
	stop           uintptr // (this)
	initialize     uintptr // (this)
	reset          uintptr // (this, extVerify)
	shutdown       uintptr // (this)
	receiveFilters uintptr // (this, enable, disable, resetMCast, mCastCount, mCastFilter)
	stationAddress uintptr // (this, reset, newAddress)
	statistics     uintptr // (this, reset, statsSize, stats)
	mCastIpToMac   uintptr // (this, ipv4, ipv6, mac)
	nvData         uintptr // (this, readWrite, offset, bufferSize, buffer)
	getStatus      uintptr // (this, intStatus, txBuf)
	transmit       uintptr // (this, headerSize, bufferSize, buffer, srcAddr, dstAddr, proto)
	receive        uintptr // (this, headerSize, bufferSize, buffer, srcAddr, dstAddr, proto)
	WaitForPacket  EFI_EVENT
	Mode           *EFI_SIMPLE_NETWORK_MODE
}

func (snp *EFI_SIMPLE_NETWORK_PROTOCOL) Start() EFI_STATUS {
	return UefiCall1(snp.start, uintptr(unsafe.Pointer(snp)))
}

func (snp *EFI_SIMPLE_NETWORK_PROTOCOL) Stop() EFI_STATUS {
	return UefiCall1(snp.stop, uintptr(unsafe.Pointer(snp)))
}

func (snp *EFI_SIMPLE_NETWORK_PROTOCOL) Initialize() EFI_STATUS {
	return UefiCall1(snp.initialize, uintptr(unsafe.Pointer(snp)))
}

func (snp *EFI_SIMPLE_NETWORK_PROTOCOL) Shutdown() EFI_STATUS {
	return UefiCall1(snp.shutdown, uintptr(unsafe.Pointer(snp)))
}

type SimpleNetworkProtocol struct {
	*EFI_SIMPLE_NETWORK_PROTOCOL
}

func EnumerateSNP() (snps []*SimpleNetworkProtocol, err error) {
	var (
		handleCount  UINTN
		handleBuffer *EFI_HANDLE
	)
	status := BS().LocateHandleBuffer(ByProtocol, &EFI_SIMPLE_NETWORK_PROTOCOL_GUID, nil, &handleCount, &handleBuffer)
	if status != EFI_SUCCESS {
		return nil, StatusError(status)
	}
	// if none were found, we should have gotten EFI_NOT_FOUND

	//turn handleBuffer into a slice of EFI_HANDLEs
	handleSlice := unsafe.Slice((*EFI_HANDLE)(unsafe.Pointer(handleBuffer)), int(handleCount))

	for i := range handleSlice {
		BS().ConnectController(handleSlice[i], nil, nil, true)
		var snp *EFI_SIMPLE_NETWORK_PROTOCOL
		status := BS().HandleProtocol(
			handleSlice[i],
			&EFI_DHCP4_SERVICE_BINDING_PROTOCOL_GUID,
			unsafe.Pointer(&snp),
		)
		if status != EFI_SUCCESS {
			// just skip, or error out entirely?? Hmm..
			continue
		}
		snps = append(snps, &SimpleNetworkProtocol{EFI_SIMPLE_NETWORK_PROTOCOL: snp})
	}

	return snps, nil
}
