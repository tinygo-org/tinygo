//go:build !scheduler.cores

package task

// Affinity is only used by the cores scheduler.
type Affinity struct{}
