//go:build !byollvm && llvm22

package cgo

/*
#cgo linux        CFLAGS:  -I/usr/include/llvm-22 -I/usr/include/llvm-c-22 -I/usr/lib/llvm-22/include -I/usr/lib64/llvm22/include
#cgo darwin,amd64 CFLAGS:  -I/usr/local/opt/llvm@22/include
#cgo darwin,arm64 CFLAGS:  -I/opt/homebrew/opt/llvm@22/include
#cgo freebsd      CFLAGS:  -I/usr/local/llvm22/include
#cgo linux        LDFLAGS: -L/usr/lib/llvm-22/lib -lclang
#cgo darwin,amd64 LDFLAGS: -L/usr/local/opt/llvm@22/lib -lclang
#cgo darwin,arm64 LDFLAGS: -L/opt/homebrew/opt/llvm@22/lib -lclang
#cgo freebsd      LDFLAGS: -L/usr/local/llvm22/lib -lclang
*/
import "C"
