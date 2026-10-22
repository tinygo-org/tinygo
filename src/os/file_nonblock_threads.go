//go:build scheduler.threads

package os

// pollBlocks lets waitIO block its thread in ppoll, as each goroutine has one.
const pollBlocks = true
