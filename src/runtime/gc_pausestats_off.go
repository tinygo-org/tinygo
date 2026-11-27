//go:build (gc.conservative || gc.precise) && !gcexp_pausestats

package runtime

const gcPauseStats = false
