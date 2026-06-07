package uefi

import "unsafe"

var GraphicsOutputProtocolGUID = EFI_GUID{
	0x9042a9de, 0x23dc, 0x4a38,
	[8]byte{0x96, 0xfb, 0x7a, 0xde, 0xd0, 0x80, 0x51, 0x6a},
}

type EFI_GRAPHICS_PIXEL_FORMAT uint32

const (
	PixelRedGreenBlueReserved8BitPerColor EFI_GRAPHICS_PIXEL_FORMAT = iota
	PixelBlueGreenRedReserved8BitPerColor
	PixelBitMask
	PixelBltOnly
	PixelFormatMax
)

type EFI_GRAPHICS_OUTPUT_BLT_OPERATION uint32

const (
	BltVideoFill EFI_GRAPHICS_OUTPUT_BLT_OPERATION = iota
	BltVideoToBltBuffer
	BltBufferToVideo
	BltVideoToVideo
	BltOperationMax
)

type EFI_PIXEL_BITMASK struct {
	RedMask      uint32
	GreenMask    uint32
	BlueMask     uint32
	ReservedMask uint32
}

type EFI_GRAPHICS_OUTPUT_MODE_INFORMATION struct {
	Version              uint32
	HorizontalResolution uint32
	VerticalResolution   uint32
	PixelFormat          EFI_GRAPHICS_PIXEL_FORMAT
	PixelInformation     EFI_PIXEL_BITMASK
	PixelsPerScanLine    uint32
}

type EFI_GRAPHICS_OUTPUT_BLT_PIXEL struct {
	Blue     uint8
	Green    uint8
	Red      uint8
	Reserved uint8
}

type EFI_GRAPHICS_OUTPUT_PROTOCOL_MODE struct {
	MaxMode         uint32
	Mode            uint32
	Info            *EFI_GRAPHICS_OUTPUT_MODE_INFORMATION
	SizeOfInfo      UINTN
	FrameBufferBase EFI_PHYSICAL_ADDRESS
	FrameBufferSize UINTN
}

type EFI_GRAPHICS_OUTPUT_PROTOCOL struct {
	queryMode uintptr
	setMode   uintptr
	blt       uintptr
	Mode      *EFI_GRAPHICS_OUTPUT_PROTOCOL_MODE
}

func (p *EFI_GRAPHICS_OUTPUT_PROTOCOL) QueryMode(
	modeNumber uint32,
	sizeOfInfo *UINTN,
	info **EFI_GRAPHICS_OUTPUT_MODE_INFORMATION,
) EFI_STATUS {
	return UefiCall4(
		p.queryMode,
		uintptr(unsafe.Pointer(p)),
		uintptr(modeNumber),
		uintptr(unsafe.Pointer(sizeOfInfo)),
		uintptr(unsafe.Pointer(info)),
	)
}

func (p *EFI_GRAPHICS_OUTPUT_PROTOCOL) SetMode(modeNumber uint32) EFI_STATUS {
	return UefiCall2(
		p.setMode,
		uintptr(unsafe.Pointer(p)),
		uintptr(modeNumber),
	)
}

func (p *EFI_GRAPHICS_OUTPUT_PROTOCOL) Blt(
	bltBuffer *EFI_GRAPHICS_OUTPUT_BLT_PIXEL,
	bltOperation EFI_GRAPHICS_OUTPUT_BLT_OPERATION,
	sourceX UINTN,
	sourceY UINTN,
	destinationX UINTN,
	destinationY UINTN,
	width UINTN,
	height UINTN,
	delta UINTN,
) EFI_STATUS {
	return UefiCall10(
		p.blt,
		uintptr(unsafe.Pointer(p)),
		uintptr(unsafe.Pointer(bltBuffer)),
		uintptr(bltOperation),
		uintptr(sourceX),
		uintptr(sourceY),
		uintptr(destinationX),
		uintptr(destinationY),
		uintptr(width),
		uintptr(height),
		uintptr(delta),
	)
}

func GraphicsOutputProtocol() (*EFI_GRAPHICS_OUTPUT_PROTOCOL, EFI_STATUS) {
	var iface *EFI_GRAPHICS_OUTPUT_PROTOCOL
	status := BS().LocateProtocol(&GraphicsOutputProtocolGUID, nil, unsafe.Pointer(&iface))
	if status != EFI_SUCCESS {
		return nil, status
	}
	return iface, EFI_SUCCESS
}
