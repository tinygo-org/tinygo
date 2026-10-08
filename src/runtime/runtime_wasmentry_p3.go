//go:build wasip3

package runtime

// A task that runs _initialize cannot block, but package initializers may
// sleep or wait for goroutines. Run them from wasi:cli/run instead.
const initInRun = true
