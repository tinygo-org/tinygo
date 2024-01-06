//go:build !byollvm && !llvm15 && !llvm16

package cgo

/*
#cgo linux        CFLAGS:  -I/usr/include/llvm-17 -I/usr/include/llvm-c-17 -I/usr/lib/llvm-17/include
#cgo darwin,amd64 CFLAGS:  -I/usr/local/opt/llvm@17/include
#cgo darwin,arm64 CFLAGS:  -I/opt/homebrew/opt/llvm@17/include
#cgo freebsd      CFLAGS:  -I/usr/local/llvm17/include
#cgo linux        LDFLAGS: -L/usr/lib/llvm-17/lib -lclang
#cgo darwin,amd64 LDFLAGS: -L/usr/local/opt/llvm@17/lib -lclang
#cgo darwin,arm64 LDFLAGS: -L/opt/homebrew/opt/llvm@17/lib -lclang
#cgo freebsd      LDFLAGS: -L/usr/local/llvm17/lib -lclang
*/
import "C"
