package runtime

// This file implements the 'chan' type and send/receive/select operations.

// A channel can be in one of the following states:
//     empty:
//       No goroutine is waiting on a send or receive operation. The 'blocked'
//       member is nil.
//     recv:
//       A goroutine tries to receive from the channel. This goroutine is stored
//       in the 'blocked' member.
//     send:
//       The reverse of send. A goroutine tries to send to the channel. This
//       goroutine is stored in the 'blocked' member.
//     closed:
//       The channel is closed. Sends will panic, receives will get a zero value
//       plus optionally the indication that the channel is zero (with the
//       comma-ok value in the task).
//
// A send/recv transmission is completed by copying from the data element of the
// sending task to the data element of the receiving task, and setting
// the 'comma-ok' value to true.
// A receive operation on a closed channel is completed by zeroing the data
// element of the receiving task and setting the 'comma-ok' value to false.

import (
	"internal/task"
	"runtime/interrupt"
	"unsafe"
)

func chanDebug(ch *channel) {
	if schedulerDebug {
		if ch.bufSize > 0 {
			println("--- channel update:", ch, ch.state.String(), ch.bufSize, ch.bufUsed)
		} else {
			println("--- channel update:", ch, ch.state.String())
		}
	}
}

// channelBlockedList is a list of channel operations on a specific channel which are currently blocked.
type channelBlockedList struct {
	// next is a pointer to the next blocked channel operation on the same channel.
	next *channelBlockedList

	// t is the task associated with this channel operation.
	// If this channel operation is not part of a select, then the pointer field of the state holds the data buffer.
	// If this channel operation is part of a select, then the pointer field of the state holds the recieve buffer.
	// If this channel operation is a receive, then the data field should be set to zero when resuming due to channel closure.
	t *task.Task

	// s is a pointer to the channel select state corresponding to this operation.
	// This will be nil if and only if this channel operation is not part of a select statement.
	// If this is a send operation, then the send buffer can be found in this select state.
	s *chanSelectState

	// allSelectOps is a slice containing all of the channel operations involved with this select statement.
	// Before resuming the task, all other channel operations on this select statement should be canceled by removing them from their corresponding lists.
	allSelectOps []channelBlockedList
}

// remove takes the current list of blocked channel operations and removes the specified operation.
// This returns the resulting list, or nil if the resulting list is empty.
// A nil receiver is treated as an empty list.
func (b *channelBlockedList) remove(old *channelBlockedList) *channelBlockedList {
	if b == old {
		return b.next
	}
	c := b
	for ; c != nil && c.next != old; c = c.next {
	}
	if c != nil {
		c.next = old.next
	}
	return b
}

// detatch removes all other channel operations that are part of the same select statement.
// If the input is not part of a select statement, this is a no-op.
// This must be called before resuming any task blocked on a channel operation in order to ensure that it is not placed on the runqueue twice.
func (b *channelBlockedList) detach() {
	if b.allSelectOps == nil {
		// nothing to do
		return
	}
	for i, v := range b.allSelectOps {
		// cancel all other channel operations that are part of this select statement
		switch {
		case &b.allSelectOps[i] == b:
			// This entry is the one that was already detatched.
			continue
		case v.t == nil:
			// This entry is not used (nil channel).
			continue
		}
		v.s.ch.blocked = v.s.ch.blocked.remove(&b.allSelectOps[i])
		if v.s.ch.blocked == nil {
			if v.s.value == nil {
				// recv operation
				if v.s.ch.state != chanStateClosed {
					v.s.ch.state = chanStateEmpty
				}
			} else {
				// send operation
				if v.s.ch.bufUsed == 0 {
					// unbuffered channel
					v.s.ch.state = chanStateEmpty
				} else {
					// buffered channel
					v.s.ch.state = chanStateBuf
				}
			}
		}
		chanDebug(v.s.ch)
	}
}

type channel struct {
	elementSize uintptr // the size of one value in this channel
	bufSize     uintptr // size of buffer (in elements)
	state       chanState
	blocked     *channelBlockedList
	bufHead     uintptr        // head index of buffer (next push index)
	bufTail     uintptr        // tail index of buffer (next pop index)
	bufUsed     uintptr        // number of elements currently in buffer
	buf         unsafe.Pointer // pointer to first element of buffer
}

// chanMake creates a new channel with the given element size and buffer length in number of elements.
// This is a compiler intrinsic.
func chanMake(elementSize uintptr, bufSize uintptr) *channel {
	return &channel{
		elementSize: elementSize,
		bufSize:     bufSize,
		buf:         alloc(elementSize*bufSize, nil),
	}
}

// Return the number of entries in this chan, called from the len builtin.
// A nil chan is defined as having length 0.
//
//go:inline
func chanLen(c *channel) int {
	if c == nil {
		return 0
	}
	return int(c.bufUsed)
}

// wrapper for use in reflect
func chanLenUnsafePointer(p unsafe.Pointer) int {
	c := (*channel)(p)
	return chanLen(c)
}

// Return the capacity of this chan, called from the cap builtin.
// A nil chan is defined as having capacity 0.
//
//go:inline
func chanCap(c *channel) int {
	if c == nil {
		return 0
	}
	return int(c.bufSize)
}

// wrapper for use in reflect
func chanCapUnsafePointer(p unsafe.Pointer) int {
	c := (*channel)(p)
	return chanCap(c)
}

// resumeRX resumes the next receiver and returns the destination pointer.
// If the ok value is true, then the caller is expected to store a value into this pointer.
func (ch *channel) resumeRX(ok bool) unsafe.Pointer {
	// pop a blocked goroutine off the stack
	var b *channelBlockedList
	b, ch.blocked = ch.blocked, ch.blocked.next

	// get destination pointer
	dst := b.t.Ptr

	if !ok {
		// the result value is zero
		memzero(dst, ch.elementSize)
		b.t.Data = 0
	}

	if b.s != nil {
		// tell the select op which case resumed
		b.t.Ptr = unsafe.Pointer(b.s)

		// detach associated operations
		b.detach()
	}

	// push task onto runqueue
	runqueue.Push(b.t)

	return dst
}

// resumeTX resumes the next sender and returns the source pointer.
// The caller is expected to read from the value in this pointer before yielding.
func (ch *channel) resumeTX() unsafe.Pointer {
	// pop a blocked goroutine off the stack
	var b *channelBlockedList
	b, ch.blocked = ch.blocked, ch.blocked.next

	// get source pointer
	src := b.t.Ptr

	if b.s != nil {
		// use state's source pointer
		src = b.s.value

		// tell the select op which case resumed
		b.t.Ptr = unsafe.Pointer(b.s)

		// detach associated operations
		b.detach()
	}

	// push task onto runqueue
	runqueue.Push(b.t)

	return src
}

// push value to end of channel if space is available
// returns whether there was space for the value in the buffer
func (ch *channel) push(value unsafe.Pointer) bool {
	// immediately return false if the channel is not buffered
	if ch.bufSize == 0 {
		return false
	}

	// ensure space is available
	if ch.bufUsed == ch.bufSize {
		return false
	}

	// copy value to buffer
	memcpy(
		unsafe.Add(ch.buf, // pointer to the base of the buffer + offset = pointer to destination element
			ch.elementSize*ch.bufHead), // element size * equivalent slice index = offset
		value,
		ch.elementSize,
	)

	// update buffer state
	ch.bufUsed++
	ch.bufHead++
	if ch.bufHead == ch.bufSize {
		ch.bufHead = 0
	}

	return true
}

// pop value from channel buffer if one is available
// returns whether a value was popped or not
// result is stored into value pointer
func (ch *channel) pop(value unsafe.Pointer) bool {
	// channel is empty
	if ch.bufUsed == 0 {
		return false
	}

	// compute address of source
	addr := unsafe.Add(ch.buf, (ch.elementSize * ch.bufTail))

	// copy value from buffer
	memcpy(
		value,
		addr,
		ch.elementSize,
	)

	// zero buffer element to allow garbage collection of value
	memzero(
		addr,
		ch.elementSize,
	)

	// update buffer state
	ch.bufUsed--

	// move tail up
	ch.bufTail++
	if ch.bufTail == ch.bufSize {
		ch.bufTail = 0
	}

	return true
}

// try to send a value to a channel, without actually blocking
// returns whether the value was sent
// will panic if channel is closed
func (ch *channel) trySend(value unsafe.Pointer) bool {
	if ch == nil {
		// send to nil channel blocks forever
		// this is non-blocking, so just say no
		return false
	}

	i := interrupt.Disable()

	switch ch.state {
	case chanStateEmpty, chanStateBuf:
		// try to dump the value directly into the buffer
		if ch.push(value) {
			ch.state = chanStateBuf
			interrupt.Restore(i)
			return true
		}
		interrupt.Restore(i)
		return false
	case chanStateRecv:
		// unblock reciever
		dst := ch.resumeRX(true)

		// copy value to reciever
		memcpy(dst, value, ch.elementSize)

		// change state to empty if there are no more receivers
		if ch.blocked == nil {
			ch.state = chanStateEmpty
		}

		interrupt.Restore(i)
		return true
	case chanStateSend:
		// something else is already waiting to send
		interrupt.Restore(i)
		return false
	case chanStateClosed:
		interrupt.Restore(i)
		runtimePanic("send on closed channel")
	default:
		interrupt.Restore(i)
		runtimePanic("invalid channel state")
	}

	interrupt.Restore(i)
	return false
}

// try to recieve a value from a channel, without really blocking
// returns whether a value was recieved
// second return is the comma-ok value
func (ch *channel) tryRecv(value unsafe.Pointer) (bool, bool) {
	if ch == nil {
		// recieve from nil channel blocks forever
		// this is non-blocking, so just say no
		return false, false
	}

	i := interrupt.Disable()

	switch ch.state {
	case chanStateBuf, chanStateSend:
		// try to pop the value directly from the buffer
		if ch.pop(value) {
			// unblock next sender if applicable
			if ch.blocked != nil {
				src := ch.resumeTX()

				// push sender's value into buffer
				ch.push(src)

				if ch.blocked == nil {
					// last sender unblocked - update state
					ch.state = chanStateBuf
				}
			}

			if ch.bufUsed == 0 {
				// channel empty - update state
				ch.state = chanStateEmpty
			}

			interrupt.Restore(i)
			return true, true
		} else if ch.blocked != nil {
			// unblock next sender if applicable
			src := ch.resumeTX()

			// copy sender's value
			memcpy(value, src, ch.elementSize)

			if ch.blocked == nil {
				// last sender unblocked - update state
				ch.state = chanStateEmpty
			}

			interrupt.Restore(i)
			return true, true
		}
		interrupt.Restore(i)
		return false, false
	case chanStateRecv, chanStateEmpty:
		// something else is already waiting to recieve
		interrupt.Restore(i)
		return false, false
	case chanStateClosed:
		if ch.pop(value) {
			interrupt.Restore(i)
			return true, true
		}

		// channel closed - nothing to recieve
		memzero(value, ch.elementSize)
		interrupt.Restore(i)
		return true, false
	default:
		runtimePanic("invalid channel state")
	}

	runtimePanic("unreachable")
	return false, false
}

type chanState uint8

const (
	chanStateEmpty  chanState = iota // nothing in channel, no senders/recievers
	chanStateRecv                    // nothing in channel, recievers waiting
	chanStateSend                    // senders waiting, buffer full if present
	chanStateBuf                     // buffer not empty, no senders waiting
	chanStateClosed                  // channel closed
)

func (s chanState) String() string {
	switch s {
	case chanStateEmpty:
		return "empty"
	case chanStateRecv:
		return "recv"
	case chanStateSend:
		return "send"
	case chanStateBuf:
		return "buffered"
	case chanStateClosed:
		return "closed"
	default:
		return "invalid"
	}
}

// chanSelectState is a single channel operation (send/recv) in a select
// statement. The value pointer is either nil (for receives) or points to the
// value to send (for sends).
type chanSelectState struct {
	ch    *channel
	value unsafe.Pointer
}

// chanSend sends a single value over the channel.
// This operation will block unless a value is immediately available.
// May panic if the channel is closed.
func chanSend(ch *channel, value unsafe.Pointer, blockedlist *channelBlockedList) {
	i := interrupt.Disable()

	if ch.trySend(value) {
		// value immediately sent
		chanDebug(ch)
		interrupt.Restore(i)
		return
	}

	if ch == nil {
		// A nil channel blocks forever. Do not schedule this goroutine again.
		interrupt.Restore(i)
		deadlock()
	}

	// wait for reciever
	sender := task.Current()
	ch.state = chanStateSend
	sender.Ptr = value
	*blockedlist = channelBlockedList{
		next: ch.blocked,
		t:    sender,
	}
	ch.blocked = blockedlist
	chanDebug(ch)
	interrupt.Restore(i)
	task.Pause()
	sender.Ptr = nil
}

// chanRecv receives a single value over a channel.
// It blocks if there is no available value to recieve.
// The recieved value is copied into the value pointer.
// Returns the comma-ok value.
func chanRecv(ch *channel, value unsafe.Pointer, blockedlist *channelBlockedList) bool {
	i := interrupt.Disable()

	if rx, ok := ch.tryRecv(value); rx {
		// value immediately available
		chanDebug(ch)
		interrupt.Restore(i)
		return ok
	}

	if ch == nil {
		// A nil channel blocks forever. Do not schedule this goroutine again.
		interrupt.Restore(i)
		deadlock()
	}

	// wait for a value
	receiver := task.Current()
	ch.state = chanStateRecv
	receiver.Ptr, receiver.Data = value, 1
	*blockedlist = channelBlockedList{
		next: ch.blocked,
		t:    receiver,
	}
	ch.blocked = blockedlist
	chanDebug(ch)
	interrupt.Restore(i)
	task.Pause()
	ok := receiver.Data == 1
	receiver.Ptr, receiver.Data = nil, 0
	return ok
}

// chanClose closes the given channel. If this channel has a receiver or is
// empty, it closes the channel. Else, it panics.
func chanClose(ch *channel) {
	if ch == nil {
		// Not allowed by the language spec.
		runtimePanic("close of nil channel")
	}
	i := interrupt.Disable()
	switch ch.state {
	case chanStateClosed:
		// Not allowed by the language spec.
		interrupt.Restore(i)
		runtimePanic("close of closed channel")
	case chanStateSend:
		// This panic should ideally on the sending side, not in this goroutine.
		// But when a goroutine tries to send while the channel is being closed,
		// that is clearly invalid: the send should have been completed already
		// before the close.
		interrupt.Restore(i)
		runtimePanic("close channel during send")
	case chanStateRecv:
		// unblock all receivers with the zero value
		ch.state = chanStateClosed
		for ch.blocked != nil {
			ch.resumeRX(false)
		}
	case chanStateEmpty, chanStateBuf:
		// Easy case. No available sender or receiver.
	}
	ch.state = chanStateClosed
	interrupt.Restore(i)
	chanDebug(ch)
}

// chanSelect is the runtime implementation of the select statement. This is
// perhaps the most complicated statement in the Go spec. It returns the
// selected index and the 'comma-ok' value.
//
// TODO: do this in a round-robin fashion (as specified in the Go spec) instead
// of picking the first one that can proceed.
func chanSelect(recvbuf unsafe.Pointer, states []chanSelectState, ops []channelBlockedList) (uintptr, bool) {
	istate := interrupt.Disable()

	if selected, ok := tryChanSelect(recvbuf, states); selected != ^uintptr(0) {
		// one channel was immediately ready
		interrupt.Restore(istate)
		return selected, ok
	}

	// construct blocked operations
	for i, v := range states {
		if v.ch == nil {
			// A nil channel receive will never complete.
			// A nil channel send would have panicked during tryChanSelect.
			ops[i] = channelBlockedList{}
			continue
		}

		ops[i] = channelBlockedList{
			next:         v.ch.blocked,
			t:            task.Current(),
			s:            &states[i],
			allSelectOps: ops,
		}
		v.ch.blocked = &ops[i]
		if v.value == nil {
			// recv
			switch v.ch.state {
			case chanStateEmpty:
				v.ch.state = chanStateRecv
			case chanStateRecv:
				// already in correct state
			default:
				interrupt.Restore(istate)
				runtimePanic("invalid channel state")
			}
		} else {
			// send
			switch v.ch.state {
			case chanStateEmpty:
				v.ch.state = chanStateSend
			case chanStateSend:
				// already in correct state
			case chanStateBuf:
				// already in correct state
			default:
				interrupt.Restore(istate)
				runtimePanic("invalid channel state")
			}
		}
		chanDebug(v.ch)
	}

	// expose rx buffer
	t := task.Current()
	t.Ptr = recvbuf
	t.Data = 1

	// wait for one case to fire
	interrupt.Restore(istate)
	task.Pause()

	// figure out which one fired and return the ok value
	return (uintptr(t.Ptr) - uintptr(unsafe.Pointer(&states[0]))) / unsafe.Sizeof(chanSelectState{}), t.Data != 0
}

// tryChanSelect is like chanSelect, but it does a non-blocking select operation.
func tryChanSelect(recvbuf unsafe.Pointer, states []chanSelectState) (uintptr, bool) {
	istate := interrupt.Disable()

	// See whether we can receive from one of the channels.
	for i, state := range states {
		if state.value == nil {
			// A receive operation.
			if rx, ok := state.ch.tryRecv(recvbuf); rx {
				chanDebug(state.ch)
				interrupt.Restore(istate)
				return uintptr(i), ok
			}
		} else {
			// A send operation: state.value is not nil.
			if state.ch.trySend(state.value) {
				chanDebug(state.ch)
				interrupt.Restore(istate)
				return uintptr(i), true
			}
		}
	}

	interrupt.Restore(istate)
	return ^uintptr(0), false
}
