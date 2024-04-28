//go:build tools

// Install linter versions specified in go.mod
// See https://marcofranssen.nl/manage-go-tools-via-go-modules for idom
package tools

import (
	_ "github.com/client9/misspell"
	_ "github.com/mgechev/revive"
)
