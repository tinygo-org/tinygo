//go:build linux && !baremetal && !tinygo.wasm && !nintendoswitch

package os_test

import (
	"os"
	"strings"
	"testing"
)

func TestHostname(t *testing.T) {
	name, err := os.Hostname()
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile("/proc/sys/kernel/hostname")
	if err != nil {
		t.Skip(err)
	}
	if w := strings.TrimSpace(string(want)); name != w {
		t.Errorf("Hostname() = %q, want %q", name, w)
	}
}
