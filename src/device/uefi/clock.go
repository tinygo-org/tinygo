//go:build uefi

package uefi

import "sync"

var calibrateMutex sync.Mutex
var calculatedFrequency uint64

func TicksFrequency() uint64 {
	frequency := getTSCFrequency()
	if frequency > 0 {
		return frequency
	}

	calibrateMutex.Lock()
	defer calibrateMutex.Unlock()
	if calculatedFrequency > 0 {
		return calculatedFrequency
	}

	var event EFI_EVENT
	var index UINTN
	if BS().CreateEvent(EVT_TIMER, TPL_CALLBACK, nil, nil, &event) != EFI_SUCCESS {
		return 0
	}
	defer BS().CloseEvent(event)

	start := Ticks()
	if BS().SetTimer(event, TimerPeriodic, 250*10000) != EFI_SUCCESS {
		return 0
	}
	if BS().WaitForEvent(1, &event, &index) != EFI_SUCCESS {
		return 0
	}

	calculatedFrequency = (Ticks() - start) * 4
	return calculatedFrequency
}
