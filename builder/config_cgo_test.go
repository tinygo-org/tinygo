package builder

import (
	"testing"

	"github.com/tinygo-org/tinygo/compileopts"
)

// TestRequiresCgo checks that targets whose libc comes from the host system
// through cgo (macOS libSystem, Windows mingw-w64) and baremetal targets
// (whose device/machine sources use cgo) are treated as requiring CGo;
// hosted bundled-libc targets (musl, wasi-libc, none) do not.
func TestRequiresCgo(t *testing.T) {
	for _, tc := range []struct {
		name      string
		libc      string
		buildTags []string
		want      bool
	}{
		{"darwin-libSystem", "darwin-libSystem", []string{"darwin", "amd64"}, true},
		{"mingw-w64", "mingw-w64", []string{"windows", "amd64"}, true},
		{"baremetal cortex-m", "picolibc", []string{"cortexm", "baremetal", "linux", "arm"}, true},
		{"baremetal avr", "picolibc", []string{"baremetal", "linux", "avr"}, true},
		{"nintendoswitch", "", []string{"nintendoswitch", "linux", "arm64"}, true},
		{"hosted linux musl", "musl", []string{"linux", "amd64"}, false},
		{"wasip1", "wasi-libc", []string{"wasip1", "wasm"}, false},
		{"wasmbuiltins", "wasmbuiltins", []string{"js", "wasm"}, false},
		{"no libc", "", []string{"linux", "amd64"}, false},
	} {
		if got := requiresCgo(&compileopts.TargetSpec{Libc: tc.libc, BuildTags: tc.buildTags}); got != tc.want {
			t.Errorf("requiresCgo(%s) = %v, want %v", tc.name, got, tc.want)
		}
	}
}
