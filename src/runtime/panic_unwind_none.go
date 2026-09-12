//go:build tinygo.unwind.none

package runtime

func startUnwind(frame *deferFrame) bool {
	return false
}

func clearPanicReplay() {
}

func rewindPanic() bool {
	return false
}
