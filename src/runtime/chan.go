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
	state   uint8
	blocked *coroutine
}

const (
	chanStateEmpty = iota
	chanStateRecv
	chanStateSend
	chanStateClosed
)

func chanSendStub(caller *coroutine, ch *channel, _ unsafe.Pointer, size uintptr)
func chanRecvStub(caller *coroutine, ch *channel, _ unsafe.Pointer, _ *bool, size uintptr)

// chanSend sends a single value over the channel. If this operation can
// complete immediately (there is a goroutine waiting for a value), it sends the
// value and re-activates both goroutines. If not, it sets itself as waiting on
// a value.
//
// The unsafe.Pointer value is used during lowering. During IR generation, it
// points to the to-be-received value. During coroutine lowering, this value is
// replaced with a read from the coroutine promise.
func chanSend(sender *coroutine, ch *channel, _ unsafe.Pointer, size uintptr) {
	if ch == nil {
		// A nil channel blocks forever. Do not scheduler this goroutine again.
		return
	}
	switch ch.state {
	case chanStateEmpty:
		ch.state = chanStateSend
		ch.blocked = sender
	case chanStateRecv:
		receiver := ch.blocked
		receiverPromise := receiver.promise()
		senderPromise := sender.promise()
		memcpy(unsafe.Pointer(&receiverPromise.data), unsafe.Pointer(&senderPromise.data), size)
		receiverPromise.commaOk = true
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
		sender.promise().next = ch.blocked
		ch.blocked = sender
	}
}

// chanRecv receives a single value over a channel. If there is an available
// sender, it receives the value immediately and re-activates both coroutines.
// If not, it sets itself as available for receiving. If the channel is closed,
// it immediately activates itself with a zero value as the result.
//
// The two unnamed values exist to help during lowering. The unsafe.Pointer
// points to the value, and the *bool points to the comma-ok value. Both are
// replaced by reads from the coroutine promise.
func chanRecv(receiver *coroutine, ch *channel, _ unsafe.Pointer, _ *bool, size uintptr) {
	if ch == nil {
		// A nil channel blocks forever. Do not scheduler this goroutine again.
		return
	}
	switch ch.state {
	case chanStateSend:
		sender := ch.blocked
		receiverPromise := receiver.promise()
		senderPromise := sender.promise()
		memcpy(unsafe.Pointer(&receiverPromise.data), unsafe.Pointer(&senderPromise.data), size)
		receiverPromise.commaOk = true
		ch.blocked = senderPromise.next
		senderPromise.next = nil
		activateTask(receiver)
		activateTask(sender)
		if ch.blocked == nil {
			ch.state = chanStateEmpty
		}
	case chanStateEmpty:
		ch.state = chanStateRecv
		ch.blocked = receiver
	case chanStateClosed:
		receiverPromise := receiver.promise()
		memzero(unsafe.Pointer(&receiverPromise.data), size)
		receiverPromise.commaOk = false
		activateTask(receiver)
	case chanStateRecv:
		receiver.promise().next = ch.blocked
		ch.blocked = receiver
	}
}

// chanClose closes the given channel. If this channel has a receiver or is
// empty, it closes the channel. Else, it panics.
func chanClose(ch *channel, size uintptr) {
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
		memzero(unsafe.Pointer(&receiverPromise.data), size)
		receiverPromise.commaOk = false
		activateTask(ch.blocked)
		ch.state = chanStateClosed
		ch.blocked = nil
	case chanStateEmpty:
		// Easy case. No available sender or receiver.
		ch.state = chanStateClosed
	}
}
