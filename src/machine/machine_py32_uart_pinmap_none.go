//go:build py32 && !py32f002bxx && !py32f003xx && !py32f030xx

package machine

func uartPinAF(uartNum uint8, pin Pin, tx bool) (uint8, bool) {
	return 0, false
}
