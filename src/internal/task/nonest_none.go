// +build !arm !baremetal avr

package task

// nonest is a sync.Locker that blocks nested interrupts while held.
// On non-ARM platforms, this is a no-op.
type nonest struct{}

func (n nonest) Lock()   {}
func (n nonest) Unlock() {}
