package builder

import (
	"os"
	"os/exec"
	"strconv"
)

// BinaryenConfig is a configuration for binaryen's wasm-opt tool.
type BinaryenConfig struct {
	// OptLevel is the optimization level to use.
	// This must be between 0 (no optimization) and 2 (maximum optimization).
	OptLevel int

	// ShrinkLevel is the size optimization level to use.
	// This must be between 0 (no size optimization) and 2 (maximum size optimization).
	ShrinkLevel int

	// SpecialPasses are special transform/optimization passes to explicitly run before the main optimizer.
	SpecialPasses []string
}

func wasmOptCmd(src, dst string, cfg BinaryenConfig) error {
	args := []string{
		src,
		"--output", dst,
		"--optimize-level", strconv.Itoa(cfg.OptLevel),
		"--shrink-level", strconv.Itoa(cfg.ShrinkLevel),
	}
	for _, pass := range cfg.SpecialPasses {
		args = append(args, "--"+pass)
	}

	cmd := exec.Command("wasm-opt", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
