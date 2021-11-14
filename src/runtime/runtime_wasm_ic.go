//go:build tinygo.wasm && ic
// +build tinygo.wasm,ic

package runtime

// Implementations of Canister System API

//go:wasm-module ic0
//export msg_reply
func msg_reply()

//go:wasm-module ic0
//export msg_reply_data_append
func msg_reply_data_append(src *uint32, size *uint32)
