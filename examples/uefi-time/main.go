package main

import (
	"fmt"
	"machine/uefi"
	"time"
)

func main() {
	conOut := uefi.ConsoleOut()

	writeString(conOut, "UEFI time probe\r\n")

	efiTime, status := uefi.GetTime()
	if status != uefi.EFI_SUCCESS {
		writeString(conOut, "GetTime failed\r\n")
		return
	}

	sec, nsec := efiTime.GetEpoch()
	writeString(conOut, "EFI: ")
	writePaddedUint(conOut, uint64(efiTime.Year), 4)
	writeString(conOut, "-")
	writePaddedUint(conOut, uint64(efiTime.Month), 2)
	writeString(conOut, "-")
	writePaddedUint(conOut, uint64(efiTime.Day), 2)
	writeString(conOut, " ")
	writePaddedUint(conOut, uint64(efiTime.Hour), 2)
	writeString(conOut, ":")
	writePaddedUint(conOut, uint64(efiTime.Minute), 2)
	writeString(conOut, ":")
	writePaddedUint(conOut, uint64(efiTime.Second), 2)
	writeString(conOut, ".")
	writePaddedUint(conOut, uint64(efiTime.Nanosecond), 9)
	writeString(conOut, "\r\n")

	writeString(conOut, "Epoch: ")
	writeInt(conOut, sec)
	writeString(conOut, " s, ")
	writeInt(conOut, int64(nsec))
	writeString(conOut, " ns\r\n")

	now := time.Now()
	writeString(conOut, "Go Unix: ")
	writeInt(conOut, now.Unix())
	writeString(conOut, "\r\n")

	fmt.Fprintln(conOut, "fmt via io.Writer works")
}

func writeString(conOut *uefi.TextOutput, s string) {
	_, _ = conOut.WriteString(s)
}

func writePaddedUint(conOut *uefi.TextOutput, v uint64, width int) {
	var buf [32]byte
	i := len(buf)
	for {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
		if v == 0 {
			break
		}
	}
	for len(buf)-i < width {
		i--
		buf[i] = '0'
	}
	writeString(conOut, string(buf[i:]))
}

func writeInt(conOut *uefi.TextOutput, v int64) {
	if v < 0 {
		writeString(conOut, "-")
		writePaddedUint(conOut, uint64(-v), 1)
		return
	}
	writePaddedUint(conOut, uint64(v), 1)
}
