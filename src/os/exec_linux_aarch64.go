// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build linux && !baremetal && !tinygo.wasm && aarch64

package os

import "errors"

// On the aarch64 architecture, the fork system call is not available.
// Therefore, the fork function is implemented to return an error.
func startProcess(name string, argv []string, attr *ProcAttr) (p *Process, err error) {
	return nil, errors.New("fork not yet supported on aarch64")
}
