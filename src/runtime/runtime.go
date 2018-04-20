
package runtime

// #include <stdio.h>
// #include <stdlib.h>
import "C"

const Compiler = "tgo"

func printstring(s string) {
	for i := 0; i < len(s); i++ {
		C.putchar(C.int(s[i]))
	}
}

func printint(n int) {
	// Print integer in signed big-endian base-10 notation, for humans to
	// read.
	// TODO: don't recurse, but still be compact (and don't divide/mod
	// more than necessary).
	if n < 0 {
		C.putchar('-')
		n = -n
	}
	prevdigits := n / 10
	if prevdigits != 0 {
		printint(prevdigits)
	}
	C.putchar(C.int((n % 10) + '0'))
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

func _panic(message interface{}) {
	printstring("panic: ")
	printitf(message)
	printnl()
	C.exit(1)
}
