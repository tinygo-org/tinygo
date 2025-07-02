package runtime

// This file implements the 'chan' type and send/receive/select operations.
//
// Every channel has a list of senders and a list of receivers, and possibly a
// queue. There is no 'channel state', the state is inferred from the available
// senders/receivers and values in the buffer.
//
// - A sender will first try to send the value to a waiting receiver if there is
//   one, but only if there is nothing in the queue (to keep the values flowing
//   in the correct order). If it can't, it will add the value in the queue and
//   possibly wait as a sender if there's no space available.
// - A receiver will first try to read a value from the queue, but if there is
//   none it will try to read from a sender in the list. It will block if it
//   can't proceed.
//
// State is kept in various ways:
//
// - The sender value is stored in the sender 'channelOp', which is really a
//   queue entry. This works for both senders and select operations: a select
//   operation has a separate value to send for each case.
// - The receiver value is stored inside Task.Ptr. This works for receivers, and
//   importantly also works for select which has a single buffer for every
//   receive operation.
// - The `Task.Data` value stores how the channel operation proceeded. For
//   normal send/receive operations, it starts at chanOperationWaiting and then
//   is changed to chanOperationOk or chanOperationClosed depending on whether
//   the send/receive proceeded normally or because it was closed. For a select
//   operation, it also stores the 'case' index in the upper bits (zero for
//   non-select operations) so that the select operation knows which case did
//   proceed.
//   The value is at the same time also a way that goroutines can be the first
//   (and only) goroutine to 'take' a channel operation using an atomic CAS
//   operation to change it from 'waiting' to any other value. This is important
//   for the select statement because multiple goroutines could try to let
//   different channels in the select statement proceed at the same time. By
//   using Task.Data, only a single channel operation in the select statement
//   can proceed.
// - It is possible for the channel queues to contain already-processed senders
//   or receivers. This can happen when the select statement managed to proceed
//   but the goroutine doing the select has not yet cleaned up the stale queue
//   entries before returning. This should therefore only happen for a short
//   period.

import (
	"internal/task"
	"runtime/interrupt"
	"unsafe"
)

// The runtime implementation of the Go 'chan' type.
type channel struct {
	closed       bool
	selectLocked bool
	elementSize  uintptr
	bufCap       uintptr // 'cap'
	bufLen       uintptr // 'len'
	bufHead      uintptr
	bufTail      uintptr
	senders      chanQueue
	receivers    chanQueue
	lock         task.PMutex
	buf          unsafe.Pointer
}

const (
	chanOperationWaiting = 0b00 // waiting for a send/receive operation to continue
	chanOperationOk      = 0b01 // successfully sent or received (not closed)
	chanOperationClosed  = 0b10 // channel was closed, the value has been zeroed
	chanOperationMask    = 0b11
)

type chanQueue struct {
	first *channelOp
}

// Pus the next channel operation to the queue. All appropriate fields must have
// been initialized already.
// This function must be called with interrupts disabled and the channel lock
// held.
func (q *chanQueue) push(node *channelOp) {
	node.next = q.first
	q.first = node
}

// Pop the next waiting channel from the queue. Channels that are no longer
// waiting (for example, when they're part of a select operation) will be
// skipped.
// This function must be called with interrupts disabled.
func (q *chanQueue) pop(chanOp uint32) *channelOp {
	for {
		if q.first == nil {
			return nil
		}

		// Pop next from the queue.
		popped := q.first
		q.first = q.first.next

		// The new value for the 'data' field will be a combination of the
		// channel operation and the select index. (The select index is 0 for
		// non-select channel operations).
		newDataValue := chanOp | popped.index<<2

		// Try to be the first to proceed with this goroutine.
		swapped := popped.task.DataAtomicUint32().CompareAndSwap(0, newDataValue)
		if swapped {
			return popped
		}
	}
}

// Remove the given to-be-removed node from the queue if it is part of the
// queue. If there are multiple, only one will be removed.
// This function must be called with interrupts disabled and the channel lock
// held.
func (q *chanQueue) remove(remove *channelOp) {
	n := &q.first
	for *n != nil {
		if *n == remove {
			*n = (*n).next
			return
		}
		n = &((*n).next)
	}
}

type channelOp struct {
	next  *channelOp
	task  *task.Task
	index uint32         // select index, 0 for non-select operation
	value unsafe.Pointer // if this is a sender, this is the value to send
}

type chanSelectState struct {
	ch    *channel
	value unsafe.Pointer
}

func chanMake(elementSize uintptr, bufSize uintptr) *channel {
	return &channel{
		elementSize: elementSize,
		bufCap:      bufSize,
		buf:         alloc(elementSize*bufSize, nil),
	}
}

// Return the number of entries in this chan, called from the len builtin.
// A nil chan is defined as having length 0.
func chanLen(c *channel) int {
	if c == nil {
		return 0
	}
	return int(c.bufLen)
}

// Return the capacity of this chan, called from the cap builtin.
// A nil chan is defined as having capacity 0.
func chanCap(c *channel) int {
	if c == nil {
		return 0
	}
	return int(c.bufCap)
}

// Push the value to the channel buffer array, for a send operation.
// This function may only be called when interrupts are disabled, the channel is
// locked and it is known there is space available in the buffer.
func (ch *channel) bufferPush(value unsafe.Pointer) {
	elemAddr := unsafe.Add(ch.buf, ch.bufHead*ch.elementSize)
	ch.bufLen++
	ch.bufHead++
	if ch.bufHead == ch.bufCap {
		ch.bufHead = 0
	}

	memcpy(elemAddr, value, ch.elementSize)
}

// Pop a value from the channel buffer and store it in the 'value' pointer, for
// a receive operation.
// This function may only be called when interrupts are disabled, the channel is
// locked and it is known there is at least one value available in the buffer.
func (ch *channel) bufferPop(value unsafe.Pointer) {
	elemAddr := unsafe.Add(ch.buf, ch.bufTail*ch.elementSize)
	ch.bufLen--
	ch.bufTail++
	if ch.bufTail == ch.bufCap {
		ch.bufTail = 0
	}

	memcpy(value, elemAddr, ch.elementSize)

	// Zero the value to allow the GC to collect it.
	memzero(elemAddr, ch.elementSize)
}

// Try to proceed with this send operation without blocking, and return whether
// the send succeeded. Interrupts must be disabled and the lock must be held
// when calling this function.
func (ch *channel) trySend(value unsafe.Pointer) bool {
	// To make sure we send values in the correct order, we can only send
	// directly to a receiver when there are no values in the buffer.

	// Do not allow sending on a closed channel.
	if ch.closed {
		// Note: we cannot currently recover from this panic.
		// There's some state in the select statement especially that would be
		// corrupted if we allowed recovering from this panic.
		runtimePanic("send on closed channel")
	}

	// There is no value in the buffer and we have a receiver available. Copy
	// the value directly into the receiver.
	if ch.bufLen == 0 {
		if receiver := ch.receivers.pop(chanOperationOk); receiver != nil {
			memcpy(receiver.task.Ptr, value, ch.elementSize)
			scheduleTask(receiver.task)
			return true
		}
	}

	// If there is space in the buffer (if this is a buffered channel), we can
	// store the value in the buffer and continue.
	if ch.bufLen < ch.bufCap {
		ch.bufferPush(value)
		return true
	}
	return false
}

func chanSend(ch *channel, value unsafe.Pointer, op *channelOp) {
	if ch == nil {
		// A nil channel blocks forever. Do not schedule this goroutine again.
		deadlock()
	}

	mask := interrupt.Disable()
	ch.lock.Lock()

	// See whether we can proceed immediately, and if so, return early.
	if ch.trySend(value) {
		ch.lock.Unlock()
		interrupt.Restore(mask)
		return
	}

	// Can't proceed. Add us to the list of senders and wait until we're awoken.
	t := task.Current()
	t.SetDataUint32(chanOperationWaiting)
	op.task = t
	op.index = 0
	op.value = value
	ch.senders.push(op)
	ch.lock.Unlock()
	interrupt.Restore(mask)

	// Wait until this goroutine is resumed.
	// It might be resumed after Unlock() and before Pause(). In that case,
	// because we use semaphores, the Pause() will continue immediately.
	task.Pause()

	// Check whether the sent happened normally (not because the channel was
	// closed while sending).
	if t.DataUint32() == chanOperationClosed {
		// Oops, this channel was closed while sending!
		runtimePanic("send on closed channel")
	}
}

// Try to proceed with this receive operation without blocking, and return
// whether the receive operation succeeded. Interrupts must be disabled and the
// lock must be held when calling this function.
func (ch *channel) tryRecv(value unsafe.Pointer) (received, ok bool) {
	// To make sure we keep the values in the channel in the correct order, we
	// first have to read values from the buffer before we can look at the
	// senders.

	// If there is a value available in the buffer, we can pull it out and
	// proceed immediately.
	if ch.bufLen > 0 {
		ch.bufferPop(value)

		// Check for the next sender available and push it to the buffer.
		if sender := ch.senders.pop(chanOperationOk); sender != nil {
			ch.bufferPush(sender.value)
			scheduleTask(sender.task)
		}

		return true, true
	}

	if ch.closed {
		// Channel is closed, so proceed immediately.
		memzero(value, ch.elementSize)
		return true, false
	}

	// If there is a sender, we can proceed with the channel operation
	// immediately.
	if sender := ch.senders.pop(chanOperationOk); sender != nil {
		memcpy(value, sender.value, ch.elementSize)
		scheduleTask(sender.task)
		return true, true
	}

	return false, false
}

func chanRecv(ch *channel, value unsafe.Pointer, op *channelOp) bool {
	if ch == nil {
		// A nil channel blocks forever. Do not schedule this goroutine again.
		deadlock()
	}

	mask := interrupt.Disable()
	ch.lock.Lock()

	if received, ok := ch.tryRecv(value); received {
		ch.lock.Unlock()
		interrupt.Restore(mask)
		return ok
	}

	// We can't proceed, so we add ourselves to the list of receivers and wait
	// until we're awoken.
	t := task.Current()
	t.Ptr = value
	t.SetDataUint32(chanOperationWaiting)
	op.task = t
	op.index = 0
	ch.receivers.push(op)
	ch.lock.Unlock()
	interrupt.Restore(mask)

	// Wait until the goroutine is resumed.
	task.Pause()

	// Return whether the receive happened from a closed channel.
	return t.DataUint32() != chanOperationClosed
}

// chanClose closes the given channel. If this channel has a receiver or is
// empty, it closes the channel. Else, it panics.
func chanClose(ch *channel) {
	if ch == nil {
		// Not allowed by the language spec.
		runtimePanic("close of nil channel")
	}

	mask := interrupt.Disable()
	ch.lock.Lock()

	if ch.closed {
		// Not allowed by the language spec.
		ch.lock.Unlock()
		interrupt.Restore(mask)
		runtimePanic("close of closed channel")
	}

	// Proceed all receiving operations that are blocked.
	for {
		receiver := ch.receivers.pop(chanOperationClosed)
		if receiver == nil {
			// Processed all receivers.
			break
		}

		// Zero the value that the receiver is getting.
		memzero(receiver.task.Ptr, ch.elementSize)

		// Wake up the receiving goroutine.
		scheduleTask(receiver.task)
	}

	// Let all senders panic.
	for {
		sender := ch.senders.pop(chanOperationClosed)
		if sender == nil {
			break // processed all senders
		}

		// Wake up the sender.
		scheduleTask(sender.task)
	}

	ch.closed = true

	ch.lock.Unlock()
	interrupt.Restore(mask)
}

// We currently use a global select lock to avoid deadlocks while locking each
// individual channel in the select. Without this global lock, two select
// operations that have a different order of the same channels could end up in a
// deadlock. This global lock is inefficient if there are many select operations
// happening in parallel, but gets the job done.
//
// If this becomes a performance issue, we can see how the Go runtime does this.
// I think it does this by sorting all states by channel address and then
// locking them in that order to avoid this deadlock.
var chanSelectLock task.PMutex

// Lock all channels (taking care to skip duplicate channels).
func lockAllStates(states []chanSelectState) {
	if !hasParallelism {
		return
	}
	for _, state := range states {
		if state.ch != nil && !state.ch.selectLocked {
			state.ch.lock.Lock()
			state.ch.selectLocked = true
		}
	}
}

// Unlock all channels (taking care to skip duplicate channels).
func unlockAllStates(states []chanSelectState) {
	if !hasParallelism {
		return
	}
	for _, state := range states {
		if state.ch != nil && state.ch.selectLocked {
			state.ch.lock.Unlock()
			state.ch.selectLocked = false
		}
	}
}

// chanSelect implements blocking or non-blocking select operations.
// The 'ops' slice must be set if (and only if) this is a blocking select.
func chanSelect(recvbuf unsafe.Pointer, states []chanSelectState, ops []channelOp) (uint32, bool) {
	mask := interrupt.Disable()

	// Lock everything.
	chanSelectLock.Lock()
	lockAllStates(states)

	const selectNoIndex = ^uint32(0)
	selectIndex := selectNoIndex
	selectOk := true

	// Iterate over each state, and see if it can proceed.
	// TODO: start from a random index.
	for i, state := range states {
		if state.ch == nil {
			// A nil channel blocks forever, so it won't take part of the select
			// operation.
			continue
		}

		if state.value == nil { // chan receive
			if received, ok := state.ch.tryRecv(recvbuf); received {
				selectIndex = uint32(i)
				selectOk = ok
				break
			}
		} else { // chan send
			if state.ch.trySend(state.value) {
				selectIndex = uint32(i)
				break
			}
		}
	}

	// If this select can immediately proceed, or is a non-blocking select,
	// return early.
	blocking := len(ops) != 0
	if selectIndex != selectNoIndex || !blocking {
		unlockAllStates(states)
		chanSelectLock.Unlock()
		interrupt.Restore(mask)
		return selectIndex, selectOk
	}

	// The select is blocking and no channel operation can proceed, so things
	// become more complicated.
	// We add ourselves as a sender/receiver to every channel, and wait for the
	// first one to complete. Only one will successfully complete, because
	// senders and receivers use a compare-and-exchange atomic operation on
	// t.Data so that only one will be able to "take" this select operation.
	t := task.Current()
	t.Ptr = recvbuf
	t.SetDataUint32(chanOperationWaiting)
	for i, state := range states {
		if state.ch == nil {
			continue
		}
		op := &ops[i]
		op.task = t
		op.index = uint32(i)
		if state.value == nil { // chan receive
			state.ch.receivers.push(op)
		} else { // chan send
			op.value = state.value
			state.ch.senders.push(op)
		}
	}

	// Now we wait until one of the send/receive operations can proceed.
	unlockAllStates(states)
	chanSelectLock.Unlock()
	interrupt.Restore(mask)
	task.Pause()

	// Resumed, so one channel operation must have progressed.

	// Make sure all channel ops are removed from the senders/receivers
	// queue before we return and the memory of them becomes invalid.
	chanSelectLock.Lock()
	lockAllStates(states)
	for i, state := range states {
		if state.ch == nil {
			continue
		}
		op := &ops[i]
		mask := interrupt.Disable()
		if state.value == nil {
			state.ch.receivers.remove(op)
		} else {
			state.ch.senders.remove(op)
		}
		interrupt.Restore(mask)
	}
	unlockAllStates(states)
	chanSelectLock.Unlock()

	// Pull the return values out of t.Data (which contains two bitfields).
	selectIndex = t.DataUint32() >> 2
	selectOk = t.DataUint32()&chanOperationMask != chanOperationClosed

	return selectIndex, selectOk
}
