//go:build !(scheduler.tasks && uefi)

package runtime

func schedulerSleepCustom(duration int64) bool {
	return false
}
