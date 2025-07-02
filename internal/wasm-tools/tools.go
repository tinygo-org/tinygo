//go:build tools

// Install tools specified in go.mod.
// See https://marcofranssen.nl/manage-go-tools-via-go-modules for idiom.
package tools

import (
	_ "go.bytecodealliance.org/cm"
	_ "go.bytecodealliance.org/cmd/wit-bindgen-go"
)

//go:generate go install go.bytecodealliance.org/cmd/wit-bindgen-go
