//go:build tinygo.unwind.explicit && scheduler.cores

package runtime

var unwindPendingSignal [numCPU]bool

//go:inline
func getUnwindSignal() bool {
	return unwindPendingSignal[currentCPU()]
}

//go:inline
func setUnwindSignal(unwinding bool) {
	unwindPendingSignal[currentCPU()] = unwinding
}
