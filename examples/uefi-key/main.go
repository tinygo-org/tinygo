package main

import uefi "machine"

func main() {
	conOut := uefi.ConsoleOut()
	conIn, err := uefi.ConsoleInput()
	if err != nil {
		_, _ = conOut.WriteString("ConsoleInput unavailable\r\n")
		return
	}

	if conIn.HasTextInputEx() {
		_, _ = conOut.WriteString("STIEx available\r\n")
	}
	if conIn.HasTextInput() {
		_, _ = conOut.WriteString("STIP available\r\n")
	}
	_, _ = conOut.WriteString("Press keys, ESC to exit...\r\n")

	for {
		key, source, err := conIn.ReadKeyWithSource()
		if err != nil {
			_, _ = conOut.WriteString("ReadKey failed\r\n")
			_, _ = conOut.WriteString("Source: ")
			_, _ = conOut.WriteString(source.String())
			_, _ = conOut.WriteString("\r\n")
			if statusErr, ok := err.(*uefi.Error); ok {
				_, _ = conOut.WriteString("EFI_STATUS: ")
				writePaddedUint(conOut, uint64(statusErr.Status()), 1)
				_, _ = conOut.WriteString("\r\n")
			}
			return
		}

		_, _ = conOut.WriteString("Source: ")
		_, _ = conOut.WriteString(source.String())
		_, _ = conOut.WriteString("\r\nScanCode: ")
		writePaddedUint(conOut, uint64(key.Key.ScanCode), 1)
		_, _ = conOut.WriteString("\r\nUnicode: ")
		writePaddedUint(conOut, uint64(key.Key.UnicodeChar), 1)
		_, _ = conOut.WriteString("\r\nShiftState: ")
		writeHexUint(conOut, uint64(key.KeyState.KeyShiftState), 8)
		_, _ = conOut.WriteString("\r\nToggleState: ")
		writeHexUint(conOut, uint64(key.KeyState.KeyToggleState), 2)
		_, _ = conOut.WriteString("\r\n\r\n")

		if key.Key.ScanCode == 23 {
			return
		}
	}
}

func writePaddedUint(conOut *uefi.TextOutput, v uint64, width int) {
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
	for len(buf)-i < width {
		i--
		buf[i] = '0'
	}
	_, _ = conOut.WriteString(string(buf[i:]))
}

func writeHexUint(conOut *uefi.TextOutput, v uint64, width int) {
	const digits = "0123456789ABCDEF"
	var buf [32]byte
	i := len(buf)
	for {
		i--
		buf[i] = digits[v&0xF]
		v >>= 4
		if v == 0 {
			break
		}
	}
	for len(buf)-i < width {
		i--
		buf[i] = '0'
	}
	_, _ = conOut.WriteString(string(buf[i:]))
}
