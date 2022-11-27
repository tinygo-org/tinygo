//go:build !byollvm && !llvm14 && !llvm15

package cgo

// As of 2023-05-05, there is a packaging issue on Debian:
// https://github.com/llvm/llvm-project/issues/62199
// A workaround is to fix this locally, using something like this:
//
//   ln -sf ../../x86_64-linux-gnu/libclang-16.so.1 /usr/lib/llvm-16/lib/libclang.so

/*
#cgo linux        CFLAGS:  -I/usr/lib/llvm-16/include
#cgo darwin,amd64 CFLAGS:  -I/usr/local/opt/llvm@16/include
#cgo darwin,arm64 CFLAGS:  -I/opt/homebrew/opt/llvm@16/include
#cgo freebsd      CFLAGS:  -I/usr/local/llvm16/include
#cgo linux        LDFLAGS: -L/usr/lib/llvm-16/lib -lclang
#cgo darwin,amd64 LDFLAGS: -L/usr/local/opt/llvm@16/lib -lclang -lffi
#cgo darwin,arm64 LDFLAGS: -L/opt/homebrew/opt/llvm@16/lib -lclang -lffi
#cgo freebsd      LDFLAGS: -L/usr/local/llvm16/lib -lclang
*/
import "C"
