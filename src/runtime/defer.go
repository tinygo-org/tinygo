package runtime

// Defer statements are implemented by transforming the function in the
// following way:
//   * Creating an alloca in the entry block that contains a pointer (initially
//     null) to the linked list of defer frames.
//   * Every time a defer statement is executed, a new defer frame is created
//     using alloca with a pointer to the previous defer frame, and the head
//     pointer in the entry block is replaced with a pointer to this defer
//     frame.
//   * On return, runtime.rundefers is called which calls all deferred functions
//     from the head of the linked list until it has gone through all defer
//     frames.

import "unsafe"

type deferContext unsafe.Pointer

type _defer struct {
	callback func(*_defer)
	next     *_defer
}

func rundefers(stack *_defer) {
	for stack != nil {
		stack.callback(stack)
		stack = stack.next
	}
}
