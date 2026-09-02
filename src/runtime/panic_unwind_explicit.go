//go:build tinygo.unwind.explicit

package runtime

func startUnwind(frame *deferFrame) bool {
	frame.PanicState |= panicUnwinding
	setUnwindSignal(true)
	return true
}

func clearPanicReplay() {
}

func rewindPanic() bool {
	return false
}
