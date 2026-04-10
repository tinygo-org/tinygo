// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	"internal/futex"
)

// This file contains stub implementations for internal/poll.
// The official golang implementation states:
//
// "That is, don't think of these as semaphores.
// Think of them as a way to implement sleep and wakeup
// such that every sleep is paired with a single wakeup,
// even if, due to races, the wakeup happens before the sleep."
//
// This is an experimental and probably incomplete implementation of the
// semaphore system, tailed to the network use case. That means, that it does not
// implement the modularity that the semacquire/semacquire1 implementation model
// offers, which in fact is emitted here entirely.
// This means we assume the following constant settings from the golang standard
// library: lifo=false,profile=semaBlock,skipframe=0,reason=waitReasonSemaquire

//go:linkname semacquire internal/poll.runtime_Semacquire
func semacquire(sema *uint32) {
	var semaBlock futex.Futex
	semaBlock.Store(*sema)

	// check if we can acquire the semaphore
	semaBlock.Wait(1)

	// the semaphore is free to use so we can acquire it
	if semaBlock.Swap(0) != 1 {
		panic("semaphore is already acquired, racy")
	}
}

//go:linkname semrelease internal/poll.runtime_Semrelease
func semrelease(sema *uint32) {
	var semaBlock futex.Futex
	semaBlock.Store(*sema)

	// check if we can release the semaphore
	if semaBlock.Swap(1) != 0 {
		panic("semaphore is not acquired, racy")
	}

	// wake up the next waiter
	semaBlock.Wake()
}
