package runtime

// Some helper types for the defer statement.
// See compiler/defer.go for details.

type _defer struct {
	callback uintptr // callback number
	next     *_defer
}
