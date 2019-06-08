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

type channel struct {
	elementSize uint16 // the size of one value in this channel
	state       chanState
	blocked     *coroutine
}

type chanState uint8

const (
	chanStateEmpty chanState = iota
	chanStateRecv
	chanStateSend
	chanStateClosed
)

// chanSelectState is a single channel operation (send/recv) in a select
// statement. The value pointer is either nil (for receives) or points to the
// value to send (for sends).
type chanSelectState struct {
	ch    *channel
	value unsafe.Pointer
}

func deadlockStub()

// chanSend sends a single value over the channel. If this operation can
// complete immediately (there is a goroutine waiting for a value), it sends the
// value and re-activates both goroutines. If not, it sets itself as waiting on
// a value.
func chanSend(sender *coroutine, ch *channel, value unsafe.Pointer) {
	if ch == nil {
		// A nil channel blocks forever. Do not scheduler this goroutine again.
		return
	}
	switch ch.state {
	case chanStateEmpty:
		sender.promise().ptr = value
		ch.state = chanStateSend
		ch.blocked = sender
	case chanStateRecv:
		receiver := ch.blocked
		receiverPromise := receiver.promise()
		memcpy(receiverPromise.ptr, value, uintptr(ch.elementSize))
		receiverPromise.data = 1 // commaOk = true
		ch.blocked = receiverPromise.next
		receiverPromise.next = nil
		activateTask(receiver)
		activateTask(sender)
		if ch.blocked == nil {
			ch.state = chanStateEmpty
		}
	case chanStateClosed:
		runtimePanic("send on closed channel")
	case chanStateSend:
		sender.promise().ptr = value
		sender.promise().next = ch.blocked
		ch.blocked = sender
	}
}

// chanRecv receives a single value over a channel. If there is an available
// sender, it receives the value immediately and re-activates both coroutines.
// If not, it sets itself as available for receiving. If the channel is closed,
// it immediately activates itself with a zero value as the result.
func chanRecv(receiver *coroutine, ch *channel, value unsafe.Pointer) {
	if ch == nil {
		// A nil channel blocks forever. Do not scheduler this goroutine again.
		return
	}
	switch ch.state {
	case chanStateSend:
		sender := ch.blocked
		senderPromise := sender.promise()
		memcpy(value, senderPromise.ptr, uintptr(ch.elementSize))
		receiver.promise().data = 1 // commaOk = true
		ch.blocked = senderPromise.next
		senderPromise.next = nil
		activateTask(receiver)
		activateTask(sender)
		if ch.blocked == nil {
			ch.state = chanStateEmpty
		}
	case chanStateEmpty:
		receiver.promise().ptr = value
		ch.state = chanStateRecv
		ch.blocked = receiver
	case chanStateClosed:
		memzero(value, uintptr(ch.elementSize))
		receiver.promise().data = 0 // commaOk = false
		activateTask(receiver)
	case chanStateRecv:
		receiver.promise().ptr = value
		receiver.promise().next = ch.blocked
		ch.blocked = receiver
	}
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
		// The receiver must be re-activated with a zero value.
		receiverPromise := ch.blocked.promise()
		memzero(receiverPromise.ptr, uintptr(ch.elementSize))
		receiverPromise.data = 0 // commaOk = false
		activateTask(ch.blocked)
		ch.state = chanStateClosed
		ch.blocked = nil
	case chanStateEmpty:
		// Easy case. No available sender or receiver.
		ch.state = chanStateClosed
	}
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
		if state.ch == nil {
			// A nil channel blocks forever, so don't consider it here.
			continue
		}
		if state.value == nil {
			// A receive operation.
			switch state.ch.state {
			case chanStateSend:
				// We can receive immediately.
				sender := state.ch.blocked
				senderPromise := sender.promise()
				memcpy(recvbuf, senderPromise.ptr, uintptr(state.ch.elementSize))
				state.ch.blocked = senderPromise.next
				senderPromise.next = nil
				activateTask(sender)
				if state.ch.blocked == nil {
					state.ch.state = chanStateEmpty
				}
				return uintptr(i), true // commaOk = true
			case chanStateClosed:
				// Receive the zero value.
				memzero(recvbuf, uintptr(state.ch.elementSize))
				return uintptr(i), false // commaOk = false
			}
		} else {
			// A send operation: state.value is not nil.
			switch state.ch.state {
			case chanStateRecv:
				receiver := state.ch.blocked
				receiverPromise := receiver.promise()
				memcpy(receiverPromise.ptr, state.value, uintptr(state.ch.elementSize))
				receiverPromise.data = 1 // commaOk = true
				state.ch.blocked = receiverPromise.next
				receiverPromise.next = nil
				activateTask(receiver)
				if state.ch.blocked == nil {
					state.ch.state = chanStateEmpty
				}
				return uintptr(i), false
			case chanStateClosed:
				runtimePanic("send on closed channel")
			}
		}
	}

	if !blocking {
		return ^uintptr(0), false
	}
	panic("unimplemented: blocking select")
}
