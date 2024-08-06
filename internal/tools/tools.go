//go:build tools

// Install tools specified in go.mod.
// See https://marcofranssen.nl/manage-go-tools-via-go-modules for idiom.
package tools

import (
	_ "github.com/golangci/misspell"
	_ "github.com/mgechev/revive"
)

//go:generate go install github.com/client9/misspell/cmd/misspell
//go:generate go install github.com/mgechev/revive
