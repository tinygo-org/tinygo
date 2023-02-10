// TINYGO: The following is copied and modified from Go 1.19.3 official implementation.

// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package net

const hexDigit = "0123456789abcdef"

// A HardwareAddr represents a physical hardware address.
type HardwareAddr []byte

func (a HardwareAddr) String() string {
	if len(a) == 0 {
		return ""
	}
	buf := make([]byte, 0, len(a)*3-1)
	for i, b := range a {
		if i > 0 {
			buf = append(buf, ':')
		}
		buf = append(buf, hexDigit[b>>4])
		buf = append(buf, hexDigit[b&0xF])
	}
	return string(buf)
}
