// The following is copied from Go 1.16 official implementation.

// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package net

// IPAddr represents the address of an IP end point.
type IPAddr struct {
	IP   IP
	Zone string // IPv6 scoped addressing zone
}
