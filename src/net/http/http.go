// The following is copied from Go 1.16 official implementation.

// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import "golang.org/x/net/http/httpguts"

func isNotToken(r rune) bool {
	return !httpguts.IsTokenRune(r)
}
