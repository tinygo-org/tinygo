//go:build scheduler.tasks && uefi

package runtime

//go:linkname gosched runtime.Gosched
func gosched()

func schedulerSleepCustom(duration int64) bool {
	deadline := ticks() + nanosecondsToTicks(duration)
	for ticks() < deadline {
		gosched()
	}
	return true
}
