package runtime

import "internal/task"

// Stub for NumCgoCall, does not return the real value
func NumCgoCall() int {
	return 0
}

func NumGoroutine() int {
	return task.NumGoroutine()
}

// Stub for Breakpoint, does not do anything.
func Breakpoint() {
	panic("Breakpoint not supported")
}

// Stub for SetBlockProfileRate, does not do anything. TinyGo has no block
// profiler.
func SetBlockProfileRate(rate int) {
}

// Stub for SetMutexProfileFraction, does not do anything and always reports the
// previous rate as 0. TinyGo has no mutex profiler.
func SetMutexProfileFraction(rate int) int {
	return 0
}

// A StackRecord describes a single execution stack.
type StackRecord struct {
	Stack0 [32]uintptr // stack trace for this record, ends at first 0 entry
}

// Stack returns the stack trace associated with the record, a prefix of
// r.Stack0.
func (r *StackRecord) Stack() []uintptr {
	for i, v := range r.Stack0 {
		if v == 0 {
			return r.Stack0[0:i]
		}
	}
	return r.Stack0[0:]
}

// Stub for ThreadCreateProfile, always reports an empty profile. TinyGo has no
// thread creation profiler.
func ThreadCreateProfile(p []StackRecord) (n int, ok bool) {
	return 0, true
}
