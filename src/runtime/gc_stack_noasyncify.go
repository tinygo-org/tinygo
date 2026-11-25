//go:build !scheduler.asyncify

package runtime

func stackReached(header uintptr) {}
