// +build wasm,wasi

package wasi

//go:wasm-module wasi_unstable
//export clock_time_get
func Clock_time_get(clockid uint32, precision uint64, time *int64) (errno uint16)
