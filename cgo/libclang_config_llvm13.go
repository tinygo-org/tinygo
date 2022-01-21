//go:build !byollvm && !llvm11 && !llvm12
// +build !byollvm,!llvm11,!llvm12

package cgo

/*
#cgo linux        CFLAGS:  -I/usr/lib/llvm-13/include
#cgo darwin amd64 CFLAGS:  -I/usr/local/opt/llvm@13/include
#cgo darwin arm64 CFLAGS:  -I/opt/homebrew/opt/llvm@13/include
#cgo freebsd      CFLAGS:  -I/usr/local/llvm13/include
#cgo linux        LDFLAGS: -L/usr/lib/llvm-13/lib -lclang
#cgo darwin amd64 LDFLAGS: -L/usr/local/opt/llvm@13/lib -lclang -lffi
#cgo darwin arm64 LDFLAGS: -L/opt/homebrew/opt/llvm@13/lib -lclang -lffi
#cgo freebsd      LDFLAGS: -L/usr/local/llvm13/lib -lclang
*/
import "C"
