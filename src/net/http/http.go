// TINYGO: The following is copied and modified from Go 1.19.3 official implementation.

// TINYGO: Removed HTTP/2 Pusher interface

// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"io"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// incomparable is a zero-width, non-comparable type. Adding it to a struct
// makes that struct also non-comparable, and generally doesn't add
// any size (as long as it's first).
type incomparable [0]func()

// maxInt64 is the effective "infinite" value for the Server and
// Transport's byte-limiting readers.
const maxInt64 = 1<<63 - 1

// aLongTimeAgo is a non-zero time, far in the past, used for
// immediate cancellation of network operations.
var aLongTimeAgo = time.Unix(1, 0)

// contextKey is a value for use with context.WithValue. It's used as
// a pointer so it fits in an interface{} without allocation.
type contextKey struct {
	name string
}

func (k *contextKey) String() string { return "net/http context value " + k.name }

// Given a string of the form "host", "host:port", or "[ipv6::address]:port",
// return true if the string includes a port.
func hasPort(s string) bool { return strings.LastIndex(s, ":") > strings.LastIndex(s, "]") }

// removeEmptyPort strips the empty port in ":port" to ""
// as mandated by RFC 3986 Section 6.2.3.
func removeEmptyPort(host string) string {
	if hasPort(host) {
		return strings.TrimSuffix(host, ":")
	}
	return host
}

func isTokenRune(r rune) bool {
	i := int(r)
	return i < len(isTokenTable) && isTokenTable[i]
}

func isNotToken(r rune) bool {
	return !isTokenRune(r)
}

// stringContainsCTLByte reports whether s contains any ASCII control character.
func stringContainsCTLByte(s string) bool {
	for i := 0; i < len(s); i++ {
		b := s[i]
		if b < ' ' || b == 0x7f {
			return true
		}
	}
	return false
}

func hexEscapeNonASCII(s string) string {
	newLen := 0
	for i := 0; i < len(s); i++ {
		if s[i] >= utf8.RuneSelf {
			newLen += 3
		} else {
			newLen++
		}
	}
	if newLen == len(s) {
		return s
	}
	b := make([]byte, 0, newLen)
	for i := 0; i < len(s); i++ {
		if s[i] >= utf8.RuneSelf {
			b = append(b, '%')
			b = strconv.AppendInt(b, int64(s[i]), 16)
		} else {
			b = append(b, s[i])
		}
	}
	return string(b)
}

// NoBody is an io.ReadCloser with no bytes. Read always returns EOF
// and Close always returns nil. It can be used in an outgoing client
// request to explicitly signal that a request has zero bytes.
// An alternative, however, is to simply set Request.Body to nil.
var NoBody = noBody{}

type noBody struct{}

func (noBody) Read([]byte) (int, error)         { return 0, io.EOF }
func (noBody) Close() error                     { return nil }
func (noBody) WriteTo(io.Writer) (int64, error) { return 0, nil }
