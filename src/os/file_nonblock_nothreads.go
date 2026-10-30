//go:build !scheduler.threads

package os

// pollBlocks is false where goroutines share a thread, so waitIO polls and
// sleeps instead of blocking it.
const pollBlocks = false
