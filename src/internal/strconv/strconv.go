// Copyright 2026 The TinyGo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Float conversion uses the exact slow paths only, so the pow10 tables are
// not linked in. See https://github.com/tinygo-org/tinygo/issues/6039.
package strconv

//go:generate go run ../../../tools/gen-strconv
