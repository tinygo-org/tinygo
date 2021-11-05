package runtime

import (
	"unsafe"
)

type stringer interface {
	String() string
}

//go:nobounds
func printstring(s string) {
	for i := 0; i < len(s); i++ {
		putchar(s[i])
	}
}

func printuint8(n uint8) {
	if TargetBits >= 32 {
		printuint32(uint32(n))
	} else {
		prevdigits := n / 10
		if prevdigits != 0 {
			printuint8(prevdigits)
		}
		putchar(byte((n % 10) + '0'))
	}
}

func printint8(n int8) {
	if TargetBits >= 32 {
		printint32(int32(n))
	} else {
		if n < 0 {
			putchar('-')
			n = -n
		}
		printuint8(uint8(n))
	}
}

func printuint16(n uint16) {
	printuint32(uint32(n))
}

func printint16(n int16) {
	printint32(int32(n))
}

func printuint32(n uint32) {
	printuint64(uint64(n))
}

func printint32(n int32) {
	// Print integer in signed big-endian base-10 notation, for humans to
	// read.
	if n < 0 {
		putchar('-')
		n = -n
	}
	printuint32(uint32(n))
}

//go:nobounds
func printuint64(n uint64) {
	digits := [20]byte{} // enough to hold (2^64)-1
	// Fill in all 10 digits.
	firstdigit := 19 // digit index that isn't zero (by default, the last to handle '0' correctly)
	for i := 19; i >= 0; i-- {
		digit := byte(n%10 + '0')
		digits[i] = digit
		if digit != '0' {
			firstdigit = i
		}
		n /= 10
	}
	// Print digits without the leading zeroes.
	for i := firstdigit; i < 20; i++ {
		putchar(digits[i])
	}
}

func printint64(n int64) {
	if n < 0 {
		putchar('-')
		n = -n
	}
	printuint64(uint64(n))
}

// printfloat32() was copied from the relevant source in the original Go
// implementation and modified to work with float32 instead of float64. It is
// copyright by the Go authors, licensed under the same BSD 3-clause license.
// See https://golang.org/LICENSE for details.
//
// It is a near-duplicate of printfloat64. This is done so that printing a
// float32 value doesn't involve float64 routines, which can be unexpected and a
// problem sometimes. It comes with a possible code size reduction if both
// printfloat32 and printfloat64 are used, which seems uncommon.
//
// Source:
// https://github.com/golang/go/blob/master/src/runtime/print.go
func printfloat32(v float32) {
	switch {
	case v != v:
		printstring("NaN")
		return
	case v+v == v && v > 0:
		printstring("+Inf")
		return
	case v+v == v && v < 0:
		printstring("-Inf")
		return
	}

	const n = 7 // digits printed
	var buf [n + 7]byte
	buf[0] = '+'
	e := 0 // exp
	if v == 0 {
		if 1/v < 0 {
			buf[0] = '-'
		}
	} else {
		if v < 0 {
			v = -v
			buf[0] = '-'
		}

		// normalize
		for v >= 10 {
			e++
			v /= 10
		}
		for v < 1 {
			e--
			v *= 10
		}

		// round
		h := float32(5.0)
		for i := 0; i < n; i++ {
			h /= 10
		}
		v += h
		if v >= 10 {
			e++
			v /= 10
		}
	}

	// format +d.dddd+edd
	for i := 0; i < n; i++ {
		s := int(v)
		buf[i+2] = byte(s + '0')
		v -= float32(s)
		v *= 10
	}
	buf[1] = buf[2]
	buf[2] = '.'

	buf[n+2] = 'e'
	buf[n+3] = '+'
	if e < 0 {
		e = -e
		buf[n+3] = '-'
	}

	buf[n+4] = byte(e/100) + '0'
	buf[n+5] = byte(e/10)%10 + '0'
	buf[n+6] = byte(e%10) + '0'
	for _, c := range buf {
		putchar(c)
	}
}

// printfloat64() was copied from the relevant source in the original Go
// implementation. It is copyright by the Go authors, licensed under the same
// BSD 3-clause license. See https://golang.org/LICENSE for details.
//
// Source:
// https://github.com/golang/go/blob/master/src/runtime/print.go
func printfloat64(v float64) {
	switch {
	case v != v:
		printstring("NaN")
		return
	case v+v == v && v > 0:
		printstring("+Inf")
		return
	case v+v == v && v < 0:
		printstring("-Inf")
		return
	}

	const n = 7 // digits printed
	var buf [n + 7]byte
	buf[0] = '+'
	e := 0 // exp
	if v == 0 {
		if 1/v < 0 {
			buf[0] = '-'
		}
	} else {
		if v < 0 {
			v = -v
			buf[0] = '-'
		}

		// normalize
		for v >= 10 {
			e++
			v /= 10
		}
		for v < 1 {
			e--
			v *= 10
		}

		// round
		h := 5.0
		for i := 0; i < n; i++ {
			h /= 10
		}
		v += h
		if v >= 10 {
			e++
			v /= 10
		}
	}

	// format +d.dddd+edd
	for i := 0; i < n; i++ {
		s := int(v)
		buf[i+2] = byte(s + '0')
		v -= float64(s)
		v *= 10
	}
	buf[1] = buf[2]
	buf[2] = '.'

	buf[n+2] = 'e'
	buf[n+3] = '+'
	if e < 0 {
		e = -e
		buf[n+3] = '-'
	}

	buf[n+4] = byte(e/100) + '0'
	buf[n+5] = byte(e/10)%10 + '0'
	buf[n+6] = byte(e%10) + '0'
	for _, c := range buf {
		putchar(c)
	}
}

func printcomplex64(c complex64) {
	putchar('(')
	printfloat32(real(c))
	printfloat32(imag(c))
	printstring("i)")
}

func printcomplex128(c complex128) {
	putchar('(')
	printfloat64(real(c))
	printfloat64(imag(c))
	printstring("i)")
}

func printspace() {
	putchar(' ')
}

func printnl() {
	if baremetal {
		putchar('\r')
	}
	putchar('\n')
}

func printitf(msg interface{}) {
	switch msg := msg.(type) {
	case bool:
		print(msg)
	case int:
		print(msg)
	case int8:
		print(msg)
	case int16:
		print(msg)
	case int32:
		print(msg)
	case int64:
		print(msg)
	case uint:
		print(msg)
	case uint8:
		print(msg)
	case uint16:
		print(msg)
	case uint32:
		print(msg)
	case uint64:
		print(msg)
	case uintptr:
		print(msg)
	case float32:
		print(msg)
	case float64:
		print(msg)
	case complex64:
		print(msg)
	case complex128:
		print(msg)
	case string:
		print(msg)
	case error:
		print(msg.Error())
	case stringer:
		print(msg.String())
	default:
		// cast to underlying type
		itf := *(*_interface)(unsafe.Pointer(&msg))
		putchar('(')
		switch unsafe.Sizeof(itf.typecode) {
		case 2:
			printuint16(uint16(itf.typecode))
		case 4:
			printuint32(uint32(itf.typecode))
		case 8:
			printuint64(uint64(itf.typecode))
		}
		putchar(':')
		print(itf.value)
		putchar(')')
	}
}

func printmap(m *hashmap) {
	print("map[")
	if m == nil {
		print("nil")
	} else {
		print(uint(m.count))
	}
	putchar(']')
}

func printptr(ptr uintptr) {
	if ptr == 0 {
		print("nil")
		return
	}
	putchar('0')
	putchar('x')
	for i := 0; i < int(unsafe.Sizeof(ptr))*2; i++ {
		nibble := byte(ptr >> (unsafe.Sizeof(ptr)*8 - 4))
		if nibble < 10 {
			putchar(nibble + '0')
		} else {
			putchar(nibble - 10 + 'a')
		}
		ptr <<= 4
	}
}

func printbool(b bool) {
	if b {
		printstring("true")
	} else {
		printstring("false")
	}
}
