//go:build avr || tinygo.wasm

package runtime

import "unsafe"

const hasReturnAddr = false

func returnAddress(level uint32) unsafe.Pointer {
	return nil
}
