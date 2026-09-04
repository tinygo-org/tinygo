package pprof

// TinyGo does not implement pprof goroutine labels, but callers such as
// google.golang.org/grpc reference the label API. Provide inert stubs so they
// compile: labels are simply not attached.

import "context"

// LabelSet is a set of profiling labels (unused under TinyGo).
type LabelSet struct{}

// Labels returns a LabelSet from key/value pairs. Inert under TinyGo.
func Labels(args ...string) LabelSet { return LabelSet{} }

// Label returns the value of the named label; always ("", false) under TinyGo.
func Label(ctx context.Context, key string) (string, bool) { return "", false }

// ForLabels iterates the labels on ctx; a no-op under TinyGo.
func ForLabels(ctx context.Context, f func(key, value string) bool) {}

// WithLabels returns ctx unchanged under TinyGo (labels are not stored).
func WithLabels(ctx context.Context, labels LabelSet) context.Context { return ctx }

// SetGoroutineLabels is a no-op under TinyGo.
func SetGoroutineLabels(ctx context.Context) {}

// Do calls f with ctx; labels are ignored under TinyGo.
func Do(ctx context.Context, labels LabelSet, f func(context.Context)) { f(ctx) }
