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
