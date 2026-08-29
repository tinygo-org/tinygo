//go:build py32 && (py32f003xx || py32f030xx)

package machine

// PY32F003 datasheet tables 3-6 through 3-8 and PY32F030 datasheet
// tables 3-1 through 3-3.
func uartPinAF(uartNum uint8, pin Pin, tx bool) (uint8, bool) {
	switch uartNum {
	case 1:
		if tx {
			switch pin {
			case PB6, PF3:
				return 0, true
			case PA2, PA9, PA14:
				return 1, true
			case PA7, PA10, PB8, PF1:
				return 8, true
			}
		} else {
			switch pin {
			case PB2, PB7:
				return 0, true
			case PA3, PA10, PA15:
				return 1, true
			case PA8, PA9, PA13, PF0:
				return 8, true
			}
		}
	case 2:
		if tx {
			switch pin {
			case PA2, PA9, PA14, PB6, PB8, PF1, PF3:
				return 4, true
			case PA0, PA4, PA7, PF0:
				return 9, true
			}
		} else {
			switch pin {
			case PB2:
				return 3, true
			case PA3, PA10, PA15, PB7, PF0, PF2:
				return 4, true
			case PA1, PA5, PA8, PF1:
				return 9, true
			}
		}
	}
	return 0, false
}
