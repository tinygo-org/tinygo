package main

import "machine/uefi"

func main() {
	conOut := uefi.ConsoleOut()
	conIn, _ := uefi.ConsoleInput()

	gop, status := uefi.GraphicsOutputProtocol()
	if status != uefi.EFI_SUCCESS {
		_, _ = conOut.WriteString("GraphicsOutputProtocol unavailable\r\n")
		_, _ = conOut.WriteString("EFI_STATUS: ")
		writeUint(conOut, uint64(status))
		_, _ = conOut.WriteString("\r\n")
		return
	}

	if gop.Mode == nil || gop.Mode.Info == nil {
		_, _ = conOut.WriteString("GOP mode info unavailable\r\n")
		return
	}

	info := gop.Mode.Info
	_, _ = conOut.WriteString("GOP current mode\r\n")
	_, _ = conOut.WriteString("Mode: ")
	writeUint(conOut, uint64(gop.Mode.Mode))
	_, _ = conOut.WriteString("\r\nResolution: ")
	writeUint(conOut, uint64(info.HorizontalResolution))
	_, _ = conOut.WriteString("x")
	writeUint(conOut, uint64(info.VerticalResolution))
	_, _ = conOut.WriteString("\r\nPixelFormat: ")
	writeUint(conOut, uint64(info.PixelFormat))
	_, _ = conOut.WriteString("\r\nFramebuffer bytes: ")
	writeUint(conOut, uint64(gop.Mode.FrameBufferSize))
	_, _ = conOut.WriteString("\r\nPress ESC to exit after draw\r\n")

	width := uefi.UINTN(info.HorizontalResolution)
	height := uefi.UINTN(info.VerticalResolution)
	halfWidth := width / 2
	halfHeight := height / 2

	fillRect(gop, 0, 0, halfWidth, halfHeight, uefi.EFI_GRAPHICS_OUTPUT_BLT_PIXEL{Red: 0xC0})
	fillRect(gop, halfWidth, 0, width-halfWidth, halfHeight, uefi.EFI_GRAPHICS_OUTPUT_BLT_PIXEL{Green: 0xC0})
	fillRect(gop, 0, halfHeight, halfWidth, height-halfHeight, uefi.EFI_GRAPHICS_OUTPUT_BLT_PIXEL{Blue: 0xC0})
	fillRect(gop, halfWidth, halfHeight, width-halfWidth, height-halfHeight, uefi.EFI_GRAPHICS_OUTPUT_BLT_PIXEL{Red: 0xC0, Green: 0xC0, Blue: 0xC0})

	if conIn == nil {
		return
	}
	for {
		key, _, err := conIn.ReadKeyWithSource()
		if err != nil {
			return
		}
		if key.Key.ScanCode == 23 {
			return
		}
	}
}

func fillRect(gop *uefi.EFI_GRAPHICS_OUTPUT_PROTOCOL, x, y, width, height uefi.UINTN, color uefi.EFI_GRAPHICS_OUTPUT_BLT_PIXEL) {
	if gop.Blt(&color, uefi.BltVideoFill, 0, 0, x, y, width, height, 0) != uefi.EFI_SUCCESS {
		return
	}
}

func writeUint(conOut *uefi.TextOutput, v uint64) {
	var buf [32]byte
	i := len(buf)
	for {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
		if v == 0 {
			break
		}
	}
	_, _ = conOut.WriteString(string(buf[i:]))
}
