
package runtime

func printstring(s string) {
	for i := 0; i < len(s); i++ {
		putchar(s[i])
	}
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
