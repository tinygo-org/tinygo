//go:build uefi

package uefi

import "unsafe"

// Time represents a UEFI time value.
type Time struct {
	Year       uint16
	Month      uint8
	Day        uint8
	Hour       uint8
	Minute     uint8
	Second     uint8
	Pad1       uint8
	Nanosecond uint32
	TimeZone   int16
	Daylight   uint8
	Pad2       uint8
}

// Timestamp converts a UEFI Time to Unix timestamp (seconds since 1970).
func (t Time) Timestamp() int64 {
	daysInMonth := [12]int{31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}

	year := int(t.Year)
	month := int(t.Month)
	day := int(t.Day)

	days := int64(0)

	for y := 1970; y < year; y++ {
		if isLeapYear(y) {
			days += 366
		} else {
			days += 365
		}
	}

	for m := 1; m < month; m++ {
		days += int64(daysInMonth[m-1])
		if m == 2 && isLeapYear(year) {
			days++
		}
	}

	days += int64(day - 1)

	sec := days * 24 * 60 * 60
	sec += int64(t.Hour) * 60 * 60
	sec += int64(t.Minute) * 60
	sec += int64(t.Second)

	if t.TimeZone != 2047 && t.TimeZone >= -1440 && t.TimeZone <= 1440 {
		sec -= int64(t.TimeZone) * 60
	}

	return sec
}

func isLeapYear(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

// GetTime retrieves the current time from UEFI runtime services.
// Returns the status code (0 = success, non-zero = error).
func GetTime(time *Time) Status {
	st := GetSystemTable()
	if st == nil || st.RuntimeServices == nil || st.RuntimeServices.GetTime == 0 {
		return ErrUnsupported
	}
	return Call(st.RuntimeServices.GetTime, uintptr(unsafe.Pointer(time)), 0)

}

// Stall delays execution for the specified number of microseconds.
// Note: This blocks all execution including goroutines.
func Stall(microseconds uint64) {
	st := GetSystemTable()
	if st == nil || st.BootServices == nil || st.BootServices.Stall == 0 {
		return
	}
	Call(st.BootServices.Stall, uintptr(microseconds))
}
