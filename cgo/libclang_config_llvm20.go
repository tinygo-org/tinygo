//go:build !byollvm && !llvm15 && !llvm16 && !llvm17 && !llvm18 && !llvm19

package cgo

/*
#cgo linux        CFLAGS:  -I/usr/include/llvm-20 -I/usr/include/llvm-c-20 -I/usr/lib/llvm-20/include
#cgo darwin,amd64 CFLAGS:  -I/usr/local/opt/llvm@20/include
#cgo darwin,arm64 CFLAGS:  -I/opt/homebrew/opt/llvm@20/include
#cgo freebsd      CFLAGS:  -I/usr/local/llvm20/include
#cgo linux        LDFLAGS: -L/usr/lib/llvm-20/lib -lclang
#cgo darwin,amd64 LDFLAGS: -L/usr/local/opt/llvm@20/lib -lclang
#cgo darwin,arm64 LDFLAGS: -L/opt/homebrew/opt/llvm@20/lib -lclang
#cgo freebsd      LDFLAGS: -L/usr/local/llvm20/lib -lclang
*/
import "C"
