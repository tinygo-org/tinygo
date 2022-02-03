//go:build !go1.16
// +build !go1.16

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package os

func (f *File) Readdirnames(n int) (names []string, err error) {
	return nil, ErrInvalid
}
