// The following is copied from Go 1.16 official implementation.

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package net

// A ParseError is the error type of literal network address parsers.
type ParseError struct {
	// Type is the type of string that was expected, such as
	// "IP address", "CIDR address".
	Type string

	// Text is the malformed text string.
	Text string
}

func (e *ParseError) Error() string { return "invalid " + e.Type + ": " + e.Text }

type AddrError struct {
	Err  string
	Addr string
}

func (e *AddrError) Error() string {
	if e == nil {
		return "<nil>"
	}
	s := e.Err
	if e.Addr != "" {
		s = "address " + e.Addr + ": " + s
	}
	return s
}

func (e *AddrError) Timeout() bool   { return false }
func (e *AddrError) Temporary() bool { return false }
