//go:build !scheduler.threads

package runtime

func newTimerNode() *timerNode { return new(timerNode) }

func releaseTimerNode(*timerNode) {}
