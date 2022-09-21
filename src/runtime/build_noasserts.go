//go:build !runtime_asserts

package runtime

// disable assertions for the garbage collector
const gcAsserts = false

// disable assertions for the scheduler
const schedulerAsserts = false
