//go:build tools

// Install tools specified in go.mod.
// See https://marcofranssen.nl/manage-go-tools-via-go-modules for idiom.
package tools

import (
	_ "github.com/ydnar/wasm-tools-go/cmd/wit-bindgen-go"
)

//go:generate go install github.com/ydnar/wasm-tools-go/cmd/wit-bindgen-go
