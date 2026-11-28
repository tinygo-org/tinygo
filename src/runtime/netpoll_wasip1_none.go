//go:build wasip1 && !scheduler.tasks && !scheduler.asyncify

package runtime

const wasiErrnoIo = 29

//go:linkname runtime_netpoll_wait
func runtime_netpoll_wait(fd uint32, mode uint8, timeout uint64) uint32 {
	var subs [2]__wasi_subscription_t
	var events [2]__wasi_event_t
	subs[0].userData = 1
	subs[0].u.setFDReadWrite(__wasi_eventtype_t(mode), fd)
	nsubs := uint32(1)
	if timeout != 0 {
		subs[1].u.setClock(wasiClockMonotonic, timeout, sleepPrecisionNanoseconds, 0)
		nsubs++
	}
	var nevents uint32
	if errno := poll_oneoff(&subs[0], &events[0], nsubs, &nevents); errno != 0 {
		return uint32(errno)
	}
	for i := uint32(0); i < nevents; i++ {
		if events[i].userData == 1 {
			return uint32(events[i].errno)
		}
		if events[i].errno != 0 {
			return uint32(events[i].errno)
		}
	}
	if nevents == 0 {
		return wasiErrnoIo
	}
	return 0
}
