package xtensa

import "C"

// DisableInterrupts disables all interrupts, and returns the old interrupt
// state.
//
//export DisableInterrupts
func DisableInterrupts() uintptr

// EnableInterrupts enables all interrupts again. The value passed in must be
// the mask returned by DisableInterrupts.
//
//export EnableInterrupts
func EnableInterrupts(mask uintptr)
