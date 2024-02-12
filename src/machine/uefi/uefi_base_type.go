package uefi

// EFI_TIME
// EFI Time Abstraction:
// Year:       1900 - 9999
// Month:      1 - 12
// Day:        1 - 31
// Hour:       0 - 23
// Minute:     0 - 59
// Second:     0 - 59
// Nanosecond: 0 - 999,999,999
// TimeZone:   -1440 to 1440 or 2047
type EFI_TIME struct {
	Year       uint16
	Month      byte
	Day        byte
	Hour       byte
	Minute     byte
	Second     byte
	Pad1       byte
	Nanosecond uint32
	TimeZone   int16
	Daylight   byte
	Pad2       byte
}
