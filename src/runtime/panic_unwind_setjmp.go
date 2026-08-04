//go:build tinygo.unwind.setjmp

package runtime

func startUnwind(frame *deferFrame) bool {
	tinygo_longjmp(frame)
	return false
}

func clearPanicReplay() {
}

func rewindPanic() bool {
	return false
}
