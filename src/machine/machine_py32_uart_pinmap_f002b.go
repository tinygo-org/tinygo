//go:build py32 && py32f002bxx

package machine

// PY32F002B datasheet tables 3-4 and 3-5.
func uartPinAF(uartNum uint8, pin Pin, tx bool) (uint8, bool) {
	if uartNum != 1 {
		return 0, false
	}
	if tx {
		switch pin {
		case PA3, PA6, PA7, PB4, PB6:
			return 1, true
		}
	} else {
		switch pin {
		case PA2, PA4, PB5:
			return 1, true
		case PA7:
			return 3, true
		}
	}
	return 0, false
}
