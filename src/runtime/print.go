
package runtime

func printstring(s string) {
	for i := 0; i < len(s); i++ {
		putchar(s[i])
	}
}

func printuint(n uint) {
	// TODO: don't recurse, but still be compact (and don't divide/mod
	// more than necessary).
	prevdigits := n / 10
	if prevdigits != 0 {
		printuint(prevdigits)
	}
	putchar(byte((n % 10) + '0'))
}

func printint(n int) {
	// Print integer in signed big-endian base-10 notation, for humans to
	// read.
	if n < 0 {
		putchar('-')
		n = -n
	}
	printuint(uint(n))
}

func printbyte(c uint8) {
	putchar(c)
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
		print("???")
	}
}
