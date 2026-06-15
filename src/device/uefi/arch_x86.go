//go:build i386 || amd64

package uefi

import "device/amd64"

func Ticks() uint64 {
	return amd64.AsmReadRdtsc()
}

func CpuPause() {
	amd64.AsmPause()
}

func getTSCFrequency() uint64 {
	return amd64.InternalGetPerformanceCounterFrequency()
}
