// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package signal

import (
	"os"
)

// Just stubbing the functions for now since signal handling is not yet implemented in tinygo
func Reset(sig ...os.Signal)                      {}
func Ignore(sig ...os.Signal)                     {}
func Notify(c chan<- os.Signal, sig ...os.Signal) {}
