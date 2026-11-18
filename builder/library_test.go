package builder

import (
	"testing"

	"github.com/tinygo-org/tinygo/compileopts"
)

func TestRISCVMarch(t *testing.T) {
	tests := []struct {
		name   string
		cflags []string
		want   string
	}{
		{"default", nil, "rv32imac"},
		{"single flag", []string{"-march=rv32imc"}, "rv32imc"},
		{"last flag wins", []string{"-march=rv32imac", "-march=rv32imafc"}, "rv32imafc"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			config := &compileopts.Config{
				Target: &compileopts.TargetSpec{CFlags: tc.cflags},
			}
			if got := riscvMarch(config, "rv32imac"); got != tc.want {
				t.Errorf("riscvMarch() = %q, want %q", got, tc.want)
			}
		})
	}
}
