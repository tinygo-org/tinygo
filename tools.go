//go:build tools

// Install tools specified in go.mod.
// See https://marcofranssen.nl/manage-go-tools-via-go-modules for idiom.
package main

import (
	_ "github.com/golangci/misspell"
	_ "github.com/mgechev/revive"
	_ "go.bytecodealliance.org/cm"
	_ "go.bytecodealliance.org/cmd/wit-bindgen-go"
)

//go:generate go install github.com/golangci/misspell/cmd/misspell
//go:generate go install github.com/mgechev/revive
//go:generate go install go.bytecodealliance.org/cmd/wit-bindgen-go
