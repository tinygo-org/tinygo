//go:build !go1.23

package cm

// HostLayout marks a struct as using host memory layout.
// See [structs.HostLayout] in Go 1.23 or later.
type HostLayout struct {
	_ hostLayout // prevent accidental conversion with plain struct{}
}

type hostLayout struct{}
