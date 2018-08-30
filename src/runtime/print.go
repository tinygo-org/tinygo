package runtime

import (
	"unsafe"
)

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

func printint16(n uint16) {
	printint32(int32(n))
}

func printuint32(n uint32) {
	// TODO: don't recurse, but still be compact (and don't divide/mod
	// more than necessary).
	prevdigits := n / 10
	if prevdigits != 0 {
		printuint32(prevdigits)
	}
	putchar(byte((n % 10) + '0'))
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

func printuint64(n uint64) {
	prevdigits := n / 10
	if prevdigits != 0 {
		printuint64(prevdigits)
	}
	putchar(byte((n % 10) + '0'))
}

func printint64(n int64) {
	if n < 0 {
		putchar('-')
		n = -n
	}
	printuint64(uint64(n))
}

func printspace() {
	putchar(' ')
}

func printnl() {
	putchar('\r')
	putchar('\n')
}

func printitf(msg interface{}) {
	switch msg := msg.(type) {
	case string:
		print(msg)
	default:
		// cast to underlying type
		itf := *(*_interface)(unsafe.Pointer(&msg))
		putchar('(')
		print(itf.typecode)
		putchar(':')
		print(itf.value)
		putchar(')')
	}
}

func printptr(ptr uintptr) {
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
