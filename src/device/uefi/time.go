package uefi

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

type EFI_TIME_CAPABILITIES struct {
	Resolution uint32
	Accuracy   uint32
	SetsToZero BOOLEAN
}

func GetTime() (EFI_TIME, EFI_STATUS) {
	var time EFI_TIME
	status := ST().RuntimeServices.GetTime(&time, nil)
	return time, status
}

func (t *EFI_TIME) GetEpoch() (sec int64, nsec int32) {
	if t.TimeZone != 0x07FF { // EFI_UNSPECIFIED_TIMEZONE
		sec -= int64(t.TimeZone) * 60
	}
	year := int(t.Year)
	month := int(t.Month)

	d := daysSinceEpoch(year)
	d += uint64(daysBefore[month-1])
	if isLeap(year) && month > 2 {
		d++
	}
	d += uint64(t.Day - 1)

	abs := d * secondsPerDay
	abs += uint64(uint64(t.Hour)*uint64(secondsPerHour) + uint64(t.Minute)*uint64(secondsPerMinute) + uint64(t.Second))

	sec = int64(abs) + (absoluteToInternal + internalToUnix)
	nsec = int32(t.Nanosecond)
	return
}

const (
	secondsPerMinute = 60
	secondsPerHour   = 60 * secondsPerMinute
	secondsPerDay    = 24 * secondsPerHour
	daysPer400Years  = 365*400 + 97
	daysPer100Years  = 365*100 + 24
	daysPer4Years    = 365*4 + 1

	absoluteZeroYear = -292277022399
	internalYear     = 1

	absoluteToInternal int64 = (absoluteZeroYear - internalYear) * 365.2425 * secondsPerDay
	unixToInternal     int64 = (1969*365 + 1969/4 - 1969/100 + 1969/400) * secondsPerDay
	internalToUnix     int64 = -unixToInternal
)

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

func daysSinceEpoch(year int) uint64 {
	y := uint64(int64(year) - absoluteZeroYear)

	n := y / 400
	y -= 400 * n
	d := daysPer400Years * n

	n = y / 100
	y -= 100 * n
	d += daysPer100Years * n

	n = y / 4
	y -= 4 * n
	d += daysPer4Years * n

	d += 365 * y
	return d
}

func isLeap(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}
