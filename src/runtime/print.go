
package runtime

// #include <stdio.h>
import "C"

func printstring(s string) {
	for i := 0; i < len(s); i++ {
		C.putchar(C.int(s[i]))
	}
}

func printuint(n uint) {
	// TODO: don't recurse, but still be compact (and don't divide/mod
	// more than necessary).
	prevdigits := n / 10
	if prevdigits != 0 {
		printuint(prevdigits)
	}
	C.putchar(C.int((n % 10) + '0'))
}

func printint(n int) {
	// Print integer in signed big-endian base-10 notation, for humans to
	// read.
	if n < 0 {
		C.putchar('-')
		n = -n
	}
	printuint(uint(n))
}

func printbyte(c uint8) {
	C.putchar(C.int(c))
}

func printspace() {
	C.putchar(' ')
}

func printnl() {
	C.putchar('\n')
}

func printitf(msg interface{}) {
	switch msg := msg.(type) {
	case string:
		print(msg)
	default:
		print("???")
	}
}
