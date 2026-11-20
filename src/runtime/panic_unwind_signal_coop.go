//go:build tinygo.unwind.explicit && scheduler.threads && wasip3

package runtime

// Cooperative threads only switch at blocking calls. The signal is only set
// while returning synchronously to the defer frame, so one global is enough.
var unwindPendingSignal bool

//go:inline
func getUnwindSignal() bool {
	return unwindPendingSignal
}

//go:inline
func setUnwindSignal(unwinding bool) {
	unwindPendingSignal = unwinding
}
