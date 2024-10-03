// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tls

// A SessionState is a resumable session.
type SessionState struct {
}

// ClientSessionState contains the state needed by a client to
// resume a previous TLS session.
type ClientSessionState struct {
	ticket  []byte
	session *SessionState
}
