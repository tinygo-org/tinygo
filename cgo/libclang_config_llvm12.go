//go:build !byollvm && llvm12
// +build !byollvm,llvm12

package cgo

/*
#cgo linux			CFLAGS:  -I/usr/lib/llvm-12/include
#cgo darwin	amd64	CFLAGS:  -I/usr/local/opt/llvm@12/include
#cgo darwin	arm64	CFLAGS:  -I/opt/homebrew/opt/llvm@12/include
#cgo freebsd		CFLAGS:  -I/usr/local/llvm12/include
#cgo linux			LDFLAGS: -L/usr/lib/llvm-12/lib -lclang
#cgo darwin	amd64	LDFLAGS: -L/usr/local/opt/llvm@12/lib -lclang -lffi
#cgo darwin	arm64	LDFLAGS: -L/opt/homebrew/opt/llvm@12/lib -lclang -lffi
#cgo freebsd		LDFLAGS: -L/usr/local/llvm12/lib -lclang
*/
import "C"
