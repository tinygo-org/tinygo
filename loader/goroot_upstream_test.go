package loader

import (
	"testing"

	"github.com/tinygo-org/tinygo/compileopts"
)

func testConfig(goos, goarch string, buildTags ...string) *compileopts.Config {
	return &compileopts.Config{
		Options: &compileopts.Options{},
		Target:  &compileopts.TargetSpec{GOOS: goos, GOARCH: goarch, BuildTags: buildTags},
	}
}

// TestPathsToOverrideUpstreamNetOsPoll checks that on hosted linux/amd64 the
// merged GOROOT takes Go's own net, os and internal/poll (no TinyGo override
// entry), while the other overrides stay in place.
func TestPathsToOverrideUpstreamNetOsPoll(t *testing.T) {
	paths := pathsToOverride(testConfig("linux", "amd64", "linux", "amd64"))
	for _, p := range []string{"net/", "net/http/", "os/", "internal/poll/"} {
		if v, ok := paths[p]; ok {
			t.Errorf("pathsToOverride has %q = %v; want no entry (upstream package)", p, v)
		}
	}
	for p, want := range map[string]bool{"runtime/": false, "internal/": true, "sync/": true, "internal/task/": false} {
		if got, ok := paths[p]; !ok || got != want {
			t.Errorf("pathsToOverride[%q] = %v (present %v); want %v", p, got, ok, want)
		}
	}
}

// TestPathsToOverrideKeepsTinyGoNetOsElsewhere checks that targets other than
// hosted linux/amd64 keep the TinyGo net, os and internal/poll overrides:
// wasm and baremetal targets, non-Linux hosted targets such as darwin, and
// other linux architectures (the epoll netpoller is amd64-only, so upstream
// os/internal/poll would panic at run time there).
func TestPathsToOverrideKeepsTinyGoNetOsElsewhere(t *testing.T) {
	cases := []struct {
		name string
		cfg  *compileopts.Config
	}{
		{"wasm", testConfig("wasip1", "wasm", "wasip1", "wasm", "tinygo.wasm")},
		{"baremetal", testConfig("linux", "arm", "linux", "arm", "baremetal")},
		{"linux/arm64", testConfig("linux", "arm64", "linux", "arm64")},
		{"linux/386", testConfig("linux", "386", "linux", "386")},
		{"darwin", testConfig("darwin", "amd64", "darwin", "amd64")},
		{"windows", testConfig("windows", "amd64", "windows", "amd64")},
	}
	for _, tc := range cases {
		paths := pathsToOverride(tc.cfg)
		// The four entries must still be present (TinyGo's own packages are
		// used); their merge value is whatever the stock table says.
		for _, p := range []string{"net/", "net/http/", "os/", "internal/poll/"} {
			if _, ok := paths[p]; !ok {
				t.Errorf("%s: pathsToOverride has no %q entry; want the TinyGo override kept", tc.name, p)
			}
		}
	}
}
