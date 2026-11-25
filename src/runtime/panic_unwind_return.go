//go:build tinygo.unwind.explicit || tinygo.unwind.asyncify

package runtime

//go:inline
func unwindPending() bool {
	return getUnwindSignal()
}

//go:inline
func clearUnwind() {
	setUnwindSignal(false)
	frame := currentDeferFrame()
	if frame != nil {
		frame.PanicState &^= panicUnwinding
	}
}
