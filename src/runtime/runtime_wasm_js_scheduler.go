//go:build wasm && !wasi && !scheduler.none && !wasip1 && !wasip2 && !wasm_unknown

package runtime

//export resume
func resume() {
	go func() {
		handleEvent()
	}()

	scheduler(false)
	exitIfMainReturned()
}

//export go_scheduler
func go_scheduler() {
	scheduler(false)
	exitIfMainReturned()
}

// exitIfMainReturned exits when main returned during this run of the
// scheduler. _start only checks for that before the program first waits.
func exitIfMainReturned() {
	if mainExited {
		mainReturnExit()
	}
}
