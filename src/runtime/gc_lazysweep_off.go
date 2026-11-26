//go:build (gc.conservative || gc.precise) && !gcexp_lazysweep

package runtime

const lazySweep = false
