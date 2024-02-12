//go:build uefi

package uefi

import (
	"sync"
)

var calibrateMutex sync.Mutex
var calculatedFrequency uint64

func TicksFrequency() uint64 {
	frequency := getTscFrequency()
	if frequency > 0 {
		return frequency
	}

	var event EFI_EVENT
	var status EFI_STATUS
	var index UINTN

	calibrateMutex.Lock()
	defer calibrateMutex.Unlock()

	freq := calculatedFrequency
	if freq > 0 {
		return freq
	}

	bs := systemTable.BootServices

	status = bs.CreateEvent(EVT_TIMER, TPL_CALLBACK, nil, nil, &event)
	if status != EFI_SUCCESS {
		DebugPrint("GetTscFrequency) CreateEvent Failed", uint64(status))
		return 0
	}
	defer bs.CloseEvent(event)

	st := Ticks()
	status = bs.SetTimer(event, TimerPeriodic, 250*10000)
	if status != EFI_SUCCESS {
		DebugPrint("GetTscFrequency) SetTimer Failed", uint64(status))
		return 0
	}
	status = bs.WaitForEvent(1, &event, &index)
	diff := Ticks() - st

	calculatedFrequency = diff * 4

	return calculatedFrequency
}

func GetTime() (EFI_TIME, EFI_STATUS) {
	var status EFI_STATUS
	var time EFI_TIME

	status = systemTable.RuntimeServices.GetTime(&time, nil)

	return time, status
}

func (t *EFI_TIME) GetEpoch() (sec int64, nsec int32) {
	year := int(t.Year)
	month := int(t.Month) - 1

	// Compute days since the absolute epoch.
	d := daysSinceEpoch(year)

	// Add in days before this month.
	d += uint64(daysBefore[month-1])
	if isLeap(year) && month >= 3 {
		d++ // February 29
	}

	// Add in days before today.
	d += uint64(t.Day - 1)

	// Add in time elapsed today.
	abs := d * secondsPerDay
	abs += uint64(uint64(t.Hour)*uint64(secondsPerHour) + uint64(t.Minute)*uint64(secondsPerMinute) + uint64(t.Second))

	sec = int64(abs) + (absoluteToInternal + internalToUnix)
	nsec = int32(t.Nanosecond)

	return
}

// region: time utils
const (
	secondsPerMinute = 60
	secondsPerHour   = 60 * secondsPerMinute
	secondsPerDay    = 24 * secondsPerHour
	daysPer400Years  = 365*400 + 97
	daysPer100Years  = 365*100 + 24
	daysPer4Years    = 365*4 + 1

	// The unsigned zero year for internal calculations.
	// Must be 1 mod 400, and times before it will not compute correctly,
	// but otherwise can be changed at will.
	absoluteZeroYear = -292277022399

	// The year of the zero Time.
	// Assumed by the unixToInternal computation below.
	internalYear = 1

	// Offsets to convert between internal and absolute or Unix times.
	absoluteToInternal int64 = (absoluteZeroYear - internalYear) * 365.2425 * secondsPerDay

	unixToInternal int64 = (1969*365 + 1969/4 - 1969/100 + 1969/400) * secondsPerDay
	internalToUnix int64 = -unixToInternal
)

// daysBefore[m] counts the number of days in a non-leap year
// before month m begins. There is an entry for m=12, counting
// the number of days before January of next year (365).
var daysBefore = [...]int32{
	0,
	31,
	31 + 28,
	31 + 28 + 31,
	31 + 28 + 31 + 30,
	31 + 28 + 31 + 30 + 31,
	31 + 28 + 31 + 30 + 31 + 30,
	31 + 28 + 31 + 30 + 31 + 30 + 31,
	31 + 28 + 31 + 30 + 31 + 30 + 31 + 31,
	31 + 28 + 31 + 30 + 31 + 30 + 31 + 31 + 30,
	31 + 28 + 31 + 30 + 31 + 30 + 31 + 31 + 30 + 31,
	31 + 28 + 31 + 30 + 31 + 30 + 31 + 31 + 30 + 31 + 30,
	31 + 28 + 31 + 30 + 31 + 30 + 31 + 31 + 30 + 31 + 30 + 31,
}

// daysSinceEpoch takes a year and returns the number of days from
// the absolute epoch to the start of that year.
// This is basically (year - zeroYear) * 365, but accounting for leap days.
func daysSinceEpoch(year int) uint64 {
	y := uint64(int64(year) - absoluteZeroYear)

	// Add in days from 400-year cycles.
	n := y / 400
	y -= 400 * n
	d := daysPer400Years * n

	// Add in 100-year cycles.
	n = y / 100
	y -= 100 * n
	d += daysPer100Years * n

	// Add in 4-year cycles.
	n = y / 4
	y -= 4 * n
	d += daysPer4Years * n

	// Add in non-leap years.
	n = y
	d += 365 * n

	return d
}

func isLeap(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

//endregion
