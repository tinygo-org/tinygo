// Package sync implements synchronization primitives similar to those provided by the standard Go implementation.
// These are not safe to access from within interrupts, or from another thread.
// The primitives also lack any fairness guarantees, similar to channels and the scheduler.
package sync
