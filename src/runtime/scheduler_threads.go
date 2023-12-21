//go:build scheduler.threads

package runtime

func getSystemStackPointer() uintptr {
	return getCurrentStackPointer()
}

func waitForEvents() {
	sleepTicks(nanosecondsToTicks(1000000))
}
