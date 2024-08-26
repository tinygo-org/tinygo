//go:build !byollvm && !llvm15 && !llvm16 && !llvm17 && !llvm18

package cgo

/*
#cgo linux        CFLAGS:  -I/usr/include/llvm-19 -I/usr/include/llvm-c-19 -I/usr/lib/llvm-19/include
#cgo darwin,amd64 CFLAGS:  -I/usr/local/opt/llvm@19/include
#cgo darwin,arm64 CFLAGS:  -I/opt/homebrew/opt/llvm@19/include
#cgo freebsd      CFLAGS:  -I/usr/local/llvm19/include
#cgo linux        LDFLAGS: -L/usr/lib/llvm-19/lib -lclang
#cgo darwin,amd64 LDFLAGS: -L/usr/local/opt/llvm@19/lib -lclang
#cgo darwin,arm64 LDFLAGS: -L/opt/homebrew/opt/llvm@19/lib -lclang
#cgo freebsd      LDFLAGS: -L/usr/local/llvm19/lib -lclang
*/
import "C"
