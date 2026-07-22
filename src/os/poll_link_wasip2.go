//go:build wasip2

package os

import (
	// Pulls internal/poll into the build for wasip2 so its TCP/pollable
	// surface (WasipNFD, DialTCPWasip2, ListenTCPWasip2) is linkable from
	// user code via //go:linkname. Once wasip2 net.Listen/Dial land in
	// the stdlib this blank import will be replaced by a real consumer.
	_ "internal/poll"
)
