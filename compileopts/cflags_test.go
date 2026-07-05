package compileopts

import (
	"path/filepath"
	"testing"
)

func TestDirectoryPrefixMaps(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{
			name: "file",
			got:  FilePrefixMap(filepath.Join("tmp", "source"), "/_/module"),
			want: "-ffile-prefix-map=" + filepath.Join("tmp", "source") + string(filepath.Separator) + "=/_/module/",
		},
		{
			name: "debug",
			got:  DebugPrefixMap(string(filepath.Separator), "/_/"),
			want: "-fdebug-prefix-map=" + string(filepath.Separator) + "=/_/",
		},
		{
			name: "existing-boundaries",
			got:  FilePrefixMap(filepath.Join("tmp", "source")+string(filepath.Separator), "/_/module/"),
			want: "-ffile-prefix-map=" + filepath.Join("tmp", "source") + string(filepath.Separator) + "=/_/module/",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Fatalf("prefix map = %q, want %q", tc.got, tc.want)
			}
		})
	}
}

func TestWASILibcHeaderPrefixMaps(t *testing.T) {
	for _, preview := range []string{"wasip1", "wasip2"} {
		t.Run(preview, func(t *testing.T) {
			name := "wasi-libc"
			if preview == "wasip2" {
				name += "-wasip2"
			}
			config := &Config{
				Options: &Options{TrimPath: true},
				Target: &TargetSpec{
					Triple:    "wasm32-unknown-wasi",
					Libc:      "wasi-libc",
					BuildTags: []string{preview},
				},
				LibraryKeys: map[string]string{name: "content-key"},
			}
			maps := config.LibcHeaderPrefixMaps()
			want := FilePrefixMap(config.LibraryPath(name), config.CSourcePath("tinygo-cache/"+name))
			if len(maps) != 1 || maps[0] != want {
				t.Fatalf("header prefix maps = %v, want %s", maps, want)
			}
		})
	}
}
