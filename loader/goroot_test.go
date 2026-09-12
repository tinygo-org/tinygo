package loader

import "testing"

func TestNeedsTLSStubPackage(t *testing.T) {
	tests := []struct {
		name      string
		goos      string
		buildTags []string
		want      bool
	}{
		{"hosted linux", "linux", []string{"linux", "amd64"}, false},
		{"hosted darwin", "darwin", []string{"darwin", "arm64"}, false},
		{"windows", "windows", []string{"windows", "amd64"}, true},
		{"wasip1", "wasip1", []string{"wasip1", "tinygo.wasm"}, true},
		{"wasip2", "wasip2", []string{"wasip2", "tinygo.wasm"}, true},
		// A baremetal target reports GOOS=linux, so the build tags have to
		// keep the stub for it.
		{"baremetal", "linux", []string{"linux", "arm", "baremetal"}, true},
		{"nintendoswitch", "linux", []string{"linux", "nintendoswitch"}, true},
		{"wasm_unknown", "linux", []string{"linux", "wasm_unknown"}, true},
	}
	for _, test := range tests {
		if got := needsTLSStubPackage(test.goos, test.buildTags); got != test.want {
			t.Errorf("%s: wanted %v, got %v", test.name, test.want, got)
		}
	}
}
