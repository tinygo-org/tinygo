//go:build rp2040
// +build rp2040

package machine

import (
	"device/rp"
	"runtime/volatile"
	"unsafe"
)

type psmType struct {
	frceOn  volatile.Register32
	frceOff volatile.Register32
	wdsel   volatile.Register32
	done    volatile.Register32
}

var psm = (*psmType)(unsafe.Pointer(rp.PSM))
