package rpi3

// for computing the current time
var startTime uint32
var startTicks uint64

//SetStartTime should be called exactly once, at boot time of a program that
//is loaded by the bootloader.
func SetStartTime(nowUnix uint32) {
	startTime = nowUnix
	startTicks = SysTimer()
}

// UART0TimeDateString is a convenience routine that can print the current time
// to UART0 if SetStartTime has been called by the bootloader.  If you use
// the bootloader normally, this should work.
func UART0TimeDateString(t uint32) {
	//years
	current := t
	yr := 1970
	for current > lengthOfYear(yr) {
		current -= lengthOfYear(yr)
		yr++
	}
	//we have less than a year of secs left
	monIndex := 0
	length := uint32(0)
	var ok bool
	for _, m := range monthSeq {
		switch m {
		case "sep":
			length = 30
		case "feb":
			length = lengthOfFebruary(yr)
		default:
			length, ok = monthLen[m]
			if !ok {
				print("bad month: '", m, "'\n")
				Abort()
			}
		}
		length = length * 24 * 60 * 60
		if current < length {
			break
		}
		current -= length
		monIndex++
	}
	// we have less than a month of secs left, and we dealt with the length of different months
	day := 1
	for current > 24*60*60 {
		current -= (24 * 60 * 60)
		day++
	}
	// we are down to hrs, mins, secs
	hour := 0
	for current > 60*60 {
		current -= (60 * 60)
		hour++
	}
	min := 0
	for current > 60 {
		current -= 60
		min++
	}
	//whats left is secs
	print("UTC: ", yr, " ", monthSeq[monIndex], " ", day, " ", hour, ":", min, ":", current)
}

var yearNormal = uint32(365 * 24 * 60 * 60)
var yearLeap = uint32(366 * 24 * 60 * 60)

var monthLen = map[string]uint32{
	"jan": 31,
	"feb": 28,
	"mar": 31,
	"apr": 30,
	"may": 31,
	"jun": 30,
	"jul": 31,
	"aug": 31,
	"sep": 30,
	"oct": 31,
	"nov": 31,
	"dec": 31,
}
var monthSeq = []string{"jan", "feb", "mar", "apr", "may", "jun", "jul", "aug", "sep", "oct", "nov", "dec"}

func lengthOfYear(yr int) uint32 {
	length := yearNormal
	if yr%4 == 0 && (yr%100 != 0 || yr%400 == 0) {
		length = yearLeap
	}
	return length
}

func lengthOfFebruary(yr int) uint32 {
	length := uint32(28)
	if yr%4 == 0 && (yr%100 != 0 || yr%400 == 0) {
		length = uint32(29)
	}
	return length
}
