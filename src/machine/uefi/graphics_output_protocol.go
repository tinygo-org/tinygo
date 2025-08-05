// Graphics Output Protocol (GOP) – §12.9 UEFI 2.10
package uefi

import (
	"unsafe"
)

//---------------------------------------------------------------------------
//  GUID                                                                   //
//---------------------------------------------------------------------------

// {9042A9DE-23DC-4A38-96FB-7ADED080516A}
var GraphicsOutputProtocolGUID = EFI_GUID{
	0x9042a9de, 0x23dc, 0x4a38,
	[8]uint8{0x96, 0xfb, 0x7a, 0xde, 0xd0, 0x80, 0x51, 0x6a},
}

//---------------------------------------------------------------------------
//  Pixel formats & blt operations                                          //
//---------------------------------------------------------------------------

const (
	PixelRedGreenBlueReserved8BitPerColor = iota
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


// 12.9.4   EFI_PIXEL_BITMASK
type EFI_PIXEL_BITMASK struct {
	RedMask, GreenMask, BlueMask, ReservedMask uint32
}

// 12.9.3   EFI_GRAPHICS_OUTPUT_MODE_INFORMATION
type EFI_GRAPHICS_OUTPUT_MODE_INFORMATION struct {
	Version              uint32
	HorizontalResolution uint32
	VerticalResolution   uint32
	PixelFormat          uint32            // values above
	PixelInformation     EFI_PIXEL_BITMASK // only valid when PixelFormat==PixelBitMask
	PixelsPerScanLine    uint32
}

// 12.9.5   EFI_GRAPHICS_OUTPUT_BLT_PIXEL
type EFI_GRAPHICS_OUTPUT_BLT_PIXEL struct {
	Blue, Green, Red, Reserved uint8
}

// 12.9.2   EFI_GRAPHICS_OUTPUT_PROTOCOL_MODE
type EFI_GRAPHICS_OUTPUT_PROTOCOL_MODE struct {
	MaxMode         uint32                                // total available modes
	Mode            uint32                                // current mode
	Info            *EFI_GRAPHICS_OUTPUT_MODE_INFORMATION // mode info for current mode
	SizeOfInfo      UINTN
	FrameBufferBase EFI_PHYSICAL_ADDRESS
	FrameBufferSize UINTN
}

//---------------------------------------------------------------------------
//  EFI_GRAPHICS_OUTPUT_PROTOCOL itself                                     //
//---------------------------------------------------------------------------

type EFI_GRAPHICS_OUTPUT_PROTOCOL struct {
	queryMode uintptr
	setMode   uintptr
	blt       uintptr
	Mode      *EFI_GRAPHICS_OUTPUT_PROTOCOL_MODE
}

// QueryMode – returns a filled-in MODE_INFORMATION for ModeNumber.
func (p *EFI_GRAPHICS_OUTPUT_PROTOCOL) QueryMode(
	ModeNumber uint32,
	SizeOfInfo *UINTN,
	Info **EFI_GRAPHICS_OUTPUT_MODE_INFORMATION,
) EFI_STATUS {
	return UefiCall4(
		p.queryMode,
		uintptr(unsafe.Pointer(p)),
		uintptr(ModeNumber),
		uintptr(unsafe.Pointer(SizeOfInfo)),
		uintptr(unsafe.Pointer(Info)),
	)
}

// SetMode – switch to the requested graphics mode.
func (p *EFI_GRAPHICS_OUTPUT_PROTOCOL) SetMode(ModeNumber uint32) EFI_STATUS {
	return UefiCall2(
		p.setMode,
		uintptr(unsafe.Pointer(p)),
		uintptr(ModeNumber),
	)
}

// Blt – the work-horse pixel pump.
//
//   - BltBuffer    may be nil when the operation is VideoFill or VideoToVideo.
//   - Delta        is bytes per scan line in BltBuffer (0 ⇒ tightly packed).
func (p *EFI_GRAPHICS_OUTPUT_PROTOCOL) Blt(
	BltBuffer *EFI_GRAPHICS_OUTPUT_BLT_PIXEL,
	BltOperation EFI_GRAPHICS_OUTPUT_BLT_OPERATION,
	SourceX, SourceY,
	DestinationX, DestinationY,
	Width, Height,
	Delta UINTN,
) EFI_STATUS {
	return UefiCall10(
		p.blt,
		uintptr(unsafe.Pointer(p)),
		uintptr(unsafe.Pointer(BltBuffer)),
		uintptr(BltOperation),
		uintptr(SourceX), uintptr(SourceY),
		uintptr(DestinationX), uintptr(DestinationY),
		uintptr(Width), uintptr(Height),
		uintptr(Delta),
	)
}

func GraphicsOutputProtocol() (*EFI_GRAPHICS_OUTPUT_PROTOCOL, error) {
	st := ST()
	var iFace unsafe.Pointer
	status := (*st).BootServices.LocateProtocol(
		&GraphicsOutputProtocolGUID,
		nil,
		unsafe.Pointer(&iFace))

	if status == EFI_SUCCESS {
		gop := (*EFI_GRAPHICS_OUTPUT_PROTOCOL)(iFace)
		return gop, nil
	}

	return nil, StatusError(status)
}

// Init finds the highest resolution mode and uses it as the
// current mode. This is not strictly necessary, but it can help
// make some poorly behaved firmware work better.
//
// Highest resolution is calculated by total number of pixels.
func (p *EFI_GRAPHICS_OUTPUT_PROTOCOL) Init() (info *EFI_GRAPHICS_OUTPUT_MODE_INFORMATION) {
	var (
		highestModeInfo EFI_GRAPHICS_OUTPUT_MODE_INFORMATION
		highestMode     uint32
		pixelMax        uint32
	)

	info = new(EFI_GRAPHICS_OUTPUT_MODE_INFORMATION)
	for i := uint32(0); i < p.Mode.MaxMode; i++ {
		var size UINTN
		status := p.QueryMode(i, &size, &info)
		if status != EFI_SUCCESS {
			// silently skip
			continue
		}
		if pixelCnt := info.HorizontalResolution * info.VerticalResolution; pixelCnt > pixelMax {
			highestMode = i
			highestModeInfo = *info
			pixelMax = pixelCnt
		}
	}

	// set the mode we found
	//if p.Mode.Mode != highestMode { // on the fence if this is better or not
	p.SetMode(highestMode)
	//}

	return &highestModeInfo
}
