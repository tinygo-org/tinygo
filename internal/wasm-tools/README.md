# wasm-tools directory

This directory has a separate `go.mod` file because the `wasm-tools-go` module requires Go 1.22, while TinyGo itself supports Go 1.19.

When the minimum Go version for TinyGo is 1.22, this directory can be folded into `internal/tools` and the `go.mod` and `go.sum` files deleted.
