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
	blocked     *task
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

// chanSend sends a single value over the channel. If this operation can
// complete immediately (there is a goroutine waiting for a value), it sends the
// value and re-activates both goroutines. If not, it sets itself as waiting on
// a value.
func chanSend(sender *task, ch *channel, value unsafe.Pointer) {
	if ch == nil {
		// A nil channel blocks forever. Do not scheduler this goroutine again.
		chanYield()
		return
	}
	switch ch.state {
	case chanStateEmpty:
		scheduleLogChan("  send: chan is empty    ", ch, sender)
		sender.state().ptr = value
		ch.state = chanStateSend
		ch.blocked = sender
		chanYield()
	case chanStateRecv:
		scheduleLogChan("  send: chan in recv mode", ch, sender)
		receiver := ch.blocked
		receiverState := receiver.state()
		memcpy(receiverState.ptr, value, uintptr(ch.elementSize))
		receiverState.data = 1 // commaOk = true
		ch.blocked = receiverState.next
		receiverState.next = nil
		activateTask(receiver)
		reactivateParent(sender)
		if ch.blocked == nil {
			ch.state = chanStateEmpty
		}
	case chanStateClosed:
		runtimePanic("send on closed channel")
	case chanStateSend:
		scheduleLogChan("  send: chan in send mode", ch, sender)
		sender.state().ptr = value
		sender.state().next = ch.blocked
		ch.blocked = sender
		chanYield()
	}
}

// chanRecv receives a single value over a channel. If there is an available
// sender, it receives the value immediately and re-activates both coroutines.
// If not, it sets itself as available for receiving. If the channel is closed,
// it immediately activates itself with a zero value as the result.
func chanRecv(receiver *task, ch *channel, value unsafe.Pointer) {
	if ch == nil {
		// A nil channel blocks forever. Do not scheduler this goroutine again.
		chanYield()
		return
	}
	switch ch.state {
	case chanStateSend:
		scheduleLogChan("  recv: chan in send mode", ch, receiver)
		sender := ch.blocked
		senderState := sender.state()
		memcpy(value, senderState.ptr, uintptr(ch.elementSize))
		receiver.state().data = 1 // commaOk = true
		ch.blocked = senderState.next
		senderState.next = nil
		reactivateParent(receiver)
		activateTask(sender)
		if ch.blocked == nil {
			ch.state = chanStateEmpty
		}
	case chanStateEmpty:
		scheduleLogChan("  recv: chan is empty    ", ch, receiver)
		receiver.state().ptr = value
		ch.state = chanStateRecv
		ch.blocked = receiver
		chanYield()
	case chanStateClosed:
		scheduleLogChan("  recv: chan is closed   ", ch, receiver)
		memzero(value, uintptr(ch.elementSize))
		receiver.state().data = 0 // commaOk = false
		reactivateParent(receiver)
	case chanStateRecv:
		scheduleLogChan("  recv: chan in recv mode", ch, receiver)
		receiver.state().ptr = value
		receiver.state().next = ch.blocked
		ch.blocked = receiver
		chanYield()
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
		receiverState := ch.blocked.state()
		memzero(receiverState.ptr, uintptr(ch.elementSize))
		receiverState.data = 0 // commaOk = false
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
				senderState := sender.state()
				memcpy(recvbuf, senderState.ptr, uintptr(state.ch.elementSize))
				state.ch.blocked = senderState.next
				senderState.next = nil
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
				receiverState := receiver.state()
				memcpy(receiverState.ptr, state.value, uintptr(state.ch.elementSize))
				receiverState.data = 1 // commaOk = true
				state.ch.blocked = receiverState.next
				receiverState.next = nil
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
