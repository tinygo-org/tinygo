//go:build !((gc.conservative || gc.precise) && scheduler.asyncify)

package runtime

// exitDrainSupported reports whether this build has anything to drain at exit.
const exitDrainSupported = false

// drainIfExited is never called when nothing survives the program.
func drainIfExited() {}
