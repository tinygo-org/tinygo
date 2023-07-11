//go:build !byollvm && llvm14

package cgo

/*
#cgo linux        CFLAGS:  -I/usr/lib/llvm-14/include
#cgo darwin,amd64 CFLAGS:  -I/usr/local/opt/llvm@14/include
#cgo darwin,arm64 CFLAGS:  -I/opt/homebrew/opt/llvm@14/include
#cgo freebsd      CFLAGS:  -I/usr/local/llvm14/include
#cgo linux        LDFLAGS: -L/usr/lib/llvm-14/lib -lclang
#cgo darwin,amd64 LDFLAGS: -L/usr/local/opt/llvm@14/lib -lclang -lffi
#cgo darwin,arm64 LDFLAGS: -L/opt/homebrew/opt/llvm@14/lib -lclang -lffi
#cgo freebsd      LDFLAGS: -L/usr/local/llvm14/lib -lclang
*/
import "C"
