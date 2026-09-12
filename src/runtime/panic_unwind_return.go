//go:build tinygo.unwind.explicit || tinygo.unwind.asyncify

package runtime

//go:inline
//go:nobounds
func unwindPending() bool {
	return getUnwindSignal()
}

//go:inline
//go:nobounds
func clearUnwind() {
	setUnwindSignal(false)
	frame := currentDeferFrame()
	if frame != nil {
		frame.PanicState &^= panicUnwinding
	}
}
