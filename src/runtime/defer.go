package runtime

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
