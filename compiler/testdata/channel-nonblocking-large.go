package main

type largeChannelValue struct {
	data [3000]int
}

func selectNonBlockingLargeSend(ch chan largeChannelValue, value largeChannelValue) bool {
	select {
	case ch <- value:
		return true
	default:
		return false
	}
}
