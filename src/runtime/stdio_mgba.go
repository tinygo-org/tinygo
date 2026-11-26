//go:build gameboyadvance && mgbadebug

package runtime

func putchar(c byte) {
	if c == '\n' || index >= len(textBuffer) {
		debugFlags.Set(logLevel)
		index = 0

		// mGBA automatically prints a new line so we can ignore it
		if c == '\n' {
			return
		}
	}

	textBuffer[index] = c
	index++
}

func getchar() byte {
	// dummy, TODO
	return 0
}

func buffered() int {
	// dummy, TODO
	return 0
}

func initStdio() {
}
