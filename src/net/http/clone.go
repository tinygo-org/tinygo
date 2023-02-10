// TINYGO: The following is copied and modified from Go 1.19.3 official implementation.

// TINYGO: Removed multipart stuff

// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"net/url"
)

func cloneURLValues(v url.Values) url.Values {
	if v == nil {
		return nil
	}
	// http.Header and url.Values have the same representation, so temporarily
	// treat it like http.Header, which does have a clone:
	return url.Values(Header(v).Clone())
}

func cloneURL(u *url.URL) *url.URL {
	if u == nil {
		return nil
	}
	u2 := new(url.URL)
	*u2 = *u
	if u.User != nil {
		u2.User = new(url.Userinfo)
		*u2.User = *u.User
	}
	return u2
}

// cloneOrMakeHeader invokes Header.Clone but if the
// result is nil, it'll instead make and return a non-nil Header.
func cloneOrMakeHeader(hdr Header) Header {
	clone := hdr.Clone()
	if clone == nil {
		clone = make(Header)
	}
	return clone
}
