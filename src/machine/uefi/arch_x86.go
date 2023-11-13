//go:build i386 || amd64

package uefi

import "device/x86"

func Ticks() uint64 {
	return x86.AsmReadRdtsc()
}

func CpuPause() {
	x86.AsmPause()
}

func getTscFrequency() uint64 {
	return x86.InternalGetPerformanceCounterFrequency()
}
