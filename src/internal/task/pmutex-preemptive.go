//go:build !tinygo.unicore

package task

// PMutex is a real mutex on systems that can be either preemptive or threaded,
// and a dummy lock on other (purely cooperative) systems.
//
// It is mainly useful for short operations that need a lock when threading may
// be involved, but which do not need a lock with a purely cooperative
// scheduler.
type PMutex = Mutex
