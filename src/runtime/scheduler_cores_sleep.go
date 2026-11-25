//go:build scheduler.cores && !virt

package runtime

// Currently sleeping core, or 0xff.
// Must only be accessed with the scheduler lock held.
var sleepingCore uint8 = 0xff

// Return whether another core is sleeping.
// May only be called with the scheduler lock held.
func hasSleepingCore() bool {
	return sleepingCore != 0xff
}
