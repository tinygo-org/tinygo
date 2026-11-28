//go:build uefi

package runtime

import "device/uefi"

func putchar(c byte) {
	if c == '\n' {
		buf := [2]uefi.CHAR16{uefi.CHAR16('\r'), 0}
		uefi.ST().ConOut.OutputString(&buf[0])
	}
	buf := [2]uefi.CHAR16{uefi.CHAR16(c), 0}
	uefi.ST().ConOut.OutputString(&buf[0])
}

// stdinBuf holds the UTF-8 bytes of the last key read from ConIn that
// getchar has not returned yet. A key is one UTF-16 code unit, so it takes
// at most 3 bytes.
var (
	stdinBuf  [3]byte
	stdinHead int
	stdinTail int
)

// readKey moves the next waiting key, if any, into stdinBuf. Keys without a
// character, such as the arrow keys, are skipped. Enter arrives as a carriage
// return and is passed on as '\n', the inverse of putchar.
//
// It polls ReadKeyStroke instead of calling CheckEvent on WaitForKey, because
// CheckEvent clears the event when it reports a key, so a later check would
// miss a key that is still waiting.
func readKey() {
	if stdinHead < stdinTail {
		return
	}
	conIn := uefi.ST().ConIn
	if conIn == nil {
		return
	}
	for {
		var key uefi.EFI_INPUT_KEY
		if conIn.ReadKeyStroke(&key) != uefi.EFI_SUCCESS {
			return
		}
		c := uint16(key.UnicodeChar)
		switch {
		case c == 0:
			continue
		case c == '\r':
			stdinBuf[0] = '\n'
			stdinTail = 1
		case c < 0x80:
			stdinBuf[0] = byte(c)
			stdinTail = 1
		case c < 0x800:
			stdinBuf[0] = 0xc0 | byte(c>>6)
			stdinBuf[1] = 0x80 | byte(c)&0x3f
			stdinTail = 2
		case c >= 0xd800 && c < 0xe000:
			// A lone surrogate is not a character.
			continue
		default:
			stdinBuf[0] = 0xe0 | byte(c>>12)
			stdinBuf[1] = 0x80 | byte(c>>6)&0x3f
			stdinBuf[2] = 0x80 | byte(c)&0x3f
			stdinTail = 3
		}
		stdinHead = 0
		return
	}
}

func getchar() byte {
	readKey()
	for stdinHead == stdinTail {
		Gosched()
		readKey()
	}
	c := stdinBuf[stdinHead]
	stdinHead++
	return c
}

func buffered() int {
	readKey()
	return stdinTail - stdinHead
}
