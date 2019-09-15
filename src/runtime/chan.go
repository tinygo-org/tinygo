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
//       commao-ok value in the coroutine).
//
// A send/recv transmission is completed by copying from the data element of the
// sending coroutine to the data element of the receiving coroutine, and setting
// the 'comma-ok' value to true.
// A receive operation on a closed channel is completed by zeroing the data
// element of the receiving coroutine and setting the 'comma-ok' value to false.

import (
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

type channel struct {
	elementSize uintptr // the size of one value in this channel
	bufSize     uintptr // size of buffer (in elements)
	state       chanState
	blocked     *task
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
		buf:         alloc(elementSize * bufSize),
	}
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
		unsafe.Pointer( // pointer to the base of the buffer + offset = pointer to destination element
			uintptr(ch.buf)+
				uintptr( // element size * equivalent slice index = offset
					ch.elementSize* // element size (bytes)
						ch.bufHead, // index of first available buffer entry
				),
		),
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
	addr := unsafe.Pointer(uintptr(ch.buf) + (ch.elementSize * ch.bufTail))

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

	switch ch.state {
	case chanStateEmpty, chanStateBuf:
		// try to dump the value directly into the buffer
		if ch.push(value) {
			ch.state = chanStateBuf
			return true
		}
		return false
	case chanStateRecv:
		// unblock reciever
		receiver := unblockChain(&ch.blocked, nil)

		// copy value to reciever
		receiverState := receiver.state()
		memcpy(receiverState.ptr, value, ch.elementSize)
		receiverState.data = 1 // commaOk = true

		// change state to empty if there are no more receivers
		if ch.blocked == nil {
			ch.state = chanStateEmpty
		}

		return true
	case chanStateSend:
		// something else is already waiting to send
		return false
	case chanStateClosed:
		runtimePanic("send on closed channel")
	default:
		runtimePanic("invalid channel state")
	}

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

	switch ch.state {
	case chanStateBuf, chanStateSend:
		// try to pop the value directly from the buffer
		if ch.pop(value) {
			// unblock next sender if applicable
			if sender := unblockChain(&ch.blocked, nil); sender != nil {
				// push sender's value into buffer
				ch.push(sender.state().ptr)

				if ch.blocked == nil {
					// last sender unblocked - update state
					ch.state = chanStateBuf
				}
			}

			if ch.bufUsed == 0 {
				// channel empty - update state
				ch.state = chanStateEmpty
			}

			return true, true
		} else if sender := unblockChain(&ch.blocked, nil); sender != nil {
			// unblock next sender if applicable
			// copy sender's value
			memcpy(value, sender.state().ptr, ch.elementSize)

			if ch.blocked == nil {
				// last sender unblocked - update state
				ch.state = chanStateEmpty
			}

			return true, true
		}
		return false, false
	case chanStateRecv, chanStateEmpty:
		// something else is already waiting to recieve
		return false, false
	case chanStateClosed:
		if ch.pop(value) {
			return true, true
		}

		// channel closed - nothing to recieve
		memzero(value, ch.elementSize)
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
func chanSend(ch *channel, value unsafe.Pointer) {
	if ch.trySend(value) {
		// value immediately sent
		chanDebug(ch)
		return
	}

	if ch == nil {
		// A nil channel blocks forever. Do not schedule this goroutine again.
		deadlock()
	}

	// wait for reciever
	sender := getCoroutine()
	ch.state = chanStateSend
	senderState := sender.state()
	senderState.ptr = value
	ch.blocked, senderState.next = sender, ch.blocked
	chanDebug(ch)
	yield()
	senderState.ptr = nil
}

// chanRecv receives a single value over a channel.
// It blocks if there is no available value to recieve.
// The recieved value is copied into the value pointer.
// Returns the comma-ok value.
func chanRecv(ch *channel, value unsafe.Pointer) bool {
	if rx, ok := ch.tryRecv(value); rx {
		// value immediately available
		chanDebug(ch)
		return ok
	}

	if ch == nil {
		// A nil channel blocks forever. Do not schedule this goroutine again.
		deadlock()
	}

	// wait for a value
	receiver := getCoroutine()
	ch.state = chanStateRecv
	receiverState := receiver.state()
	receiverState.ptr, receiverState.data = value, 0
	ch.blocked, receiverState.next = receiver, ch.blocked
	chanDebug(ch)
	yield()
	ok := receiverState.data == 1
	receiverState.ptr, receiverState.data = nil, 0
	return ok
}

// chanClose closes the given channel. If this channel has a receiver or is
// empty, it closes the channel. Else, it panics.
func chanClose(ch *channel) {
	if ch == nil {
		// Not allowed by the language spec.
		runtimePanic("close of nil channel")
	}
	switch ch.state {
	case chanStateClosed:
		// Not allowed by the language spec.
		runtimePanic("close of closed channel")
	case chanStateSend:
		// This panic should ideally on the sending side, not in this goroutine.
		// But when a goroutine tries to send while the channel is being closed,
		// that is clearly invalid: the send should have been completed already
		// before the close.
		runtimePanic("close channel during send")
	case chanStateRecv:
		// unblock all receivers with the zero value
		for rx := unblockChain(&ch.blocked, nil); rx != nil; rx = unblockChain(&ch.blocked, nil) {
			// get receiver state
			state := rx.state()

			// store the zero value
			memzero(state.ptr, ch.elementSize)

			// set the comma-ok value to false (channel closed)
			state.data = 0
		}
	case chanStateEmpty, chanStateBuf:
		// Easy case. No available sender or receiver.
	}
	ch.state = chanStateClosed
	chanDebug(ch)
}

// chanSelect is the runtime implementation of the select statement. This is
// perhaps the most complicated statement in the Go spec. It returns the
// selected index and the 'comma-ok' value.
//
// TODO: do this in a round-robin fashion (as specified in the Go spec) instead
// of picking the first one that can proceed.
func chanSelect(recvbuf unsafe.Pointer, states []chanSelectState, blocking bool) (uintptr, bool) {
	// See whether we can receive from one of the channels.
	for i, state := range states {
		if state.value == nil {
			// A receive operation.
			if rx, ok := state.ch.tryRecv(recvbuf); rx {
				chanDebug(state.ch)
				return uintptr(i), ok
			}
		} else {
			// A send operation: state.value is not nil.
			if state.ch.trySend(state.value) {
				chanDebug(state.ch)
				return uintptr(i), true
			}
		}
	}

	if !blocking {
		return ^uintptr(0), false
	}
	panic("unimplemented: blocking select")
}
