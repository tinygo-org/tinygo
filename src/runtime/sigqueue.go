// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !plan9
// +build !plan9

package runtime

// Stub definitions copied from upstream golang
// TODO: implement a minimal functional version of these functions

// go: linkname os/signal.signal_disable signal_disable
func signal_disable(_ uint32) {}

// go: linkname os/signal.signal_enable signal_enable
func signal_enable(_ uint32) {}

// go: linkname os/signal.signal_ignore signal_ignore
func signal_ignore(_ uint32) {}

// go: linkname os/signal.signal_ignored signal_ignored
func signal_ignored(_ uint32) bool { return true }

// go: linkname os/signal.signal_recv signal_recv
func signal_recv() uint32 { return 0 }
