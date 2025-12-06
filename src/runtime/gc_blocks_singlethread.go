//go:build (gc.conservative || gc.precise) && !scheduler.threads

package runtime

// scanList is a singly linked list of heap objects that have been marked but not scanned.
var scanList *objHeader

func getScanList() **objHeader {
	return &scanList
}

func (b gcBlock) stateAtomic() blockState {
	return b.state()
}

func (b gcBlock) stateByteAtomic() byte {
	return b.stateByte()
}

func (b gcBlock) mark() bool {
	if b.state() == blockStateMark {
		return false
	}
	b.setState(blockStateMark)
	return true
}
