//go:build !byollvm && llvm23

package cgo

/*
#cgo linux        CFLAGS:  -I/usr/include/llvm-23 -I/usr/include/llvm-c-23 -I/usr/lib/llvm-23/include -I/usr/lib64/llvm23/include
#cgo darwin,amd64 CFLAGS:  -I/usr/local/opt/llvm@23/include
#cgo darwin,arm64 CFLAGS:  -I/opt/homebrew/opt/llvm@23/include
#cgo freebsd      CFLAGS:  -I/usr/local/llvm23/include
#cgo linux        LDFLAGS: -L/usr/lib/llvm-23/lib -lclang
#cgo darwin,amd64 LDFLAGS: -L/usr/local/opt/llvm@23/lib -lclang
#cgo darwin,arm64 LDFLAGS: -L/opt/homebrew/opt/llvm@23/lib -lclang
#cgo freebsd      LDFLAGS: -L/usr/local/llvm23/lib -lclang
*/
import "C"
