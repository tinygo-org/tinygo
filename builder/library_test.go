package builder

import (
	"testing"

	"github.com/tinygo-org/tinygo/compileopts"
)

func TestRISCVMarchUsesLastFlag(t *testing.T) {
	config := &compileopts.Config{
		Target: &compileopts.TargetSpec{
			CFlags: []string{"-march=rv32imac", "-march=rv32imafc"},
		},
	}
	if got := riscvMarch(config, "rv32imac"); got != "rv32imafc" {
		t.Errorf("riscvMarch returned %q, want rv32imafc", got)
	}
}
