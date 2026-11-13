//go:build scheduler.threads && (darwin || (linux && !baremetal && !wasm_unknown && !wasip2 && !nintendoswitch))

package runtime

func SleepTicksNs(ns int64) {
	sleepTicks(nanosecondsToTicks(ns))
}

func SetSignalFutex(v uint32) {
	signalFutex.Store(v)
}

func GetSignalFutex() uint32 {
	return signalFutex.Load()
}
