package compileopts

import (
	"encoding/csv"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"
)

func TestLoadTarget(t *testing.T) {
	_, err := LoadTarget(&Options{Target: "arduino"})
	if err != nil {
		t.Error("LoadTarget test failed:", err)
	}

	_, err = LoadTarget(&Options{Target: "notexist"})
	if err == nil {
		t.Error("LoadTarget should have failed with non existing target")
	}

	if !errors.Is(err, fs.ErrNotExist) {
		t.Error("LoadTarget failed for wrong reason:", err)
	}
}

func TestGetTargetSpecs_InheritableOnlyTargetsExcluded(t *testing.T) {
	specs, err := GetTargetSpecs()
	if err != nil {
		t.Fatal("GetTargetSpecs failed:", err)
	}

	// Inheritable-only processor-level targets should not appear in the listing.
	inheritableOnlyTargets := []string{"esp32", "esp32c3", "esp32s3", "esp8266", "rp2040", "rp2350", "rp2350b"}
	for _, name := range inheritableOnlyTargets {
		if _, ok := specs[name]; ok {
			t.Errorf("inheritable-only target %q should not appear in GetTargetSpecs", name)
		}
	}

	// Board targets that inherit from inheritable-only targets should still appear.
	boardTargets := []string{"esp32-coreboard-v2", "pico"}
	for _, name := range boardTargets {
		if _, ok := specs[name]; !ok {
			t.Errorf("board target %q should appear in GetTargetSpecs", name)
		}
	}
}

func TestLoadTarget_InheritableOnlyTargetStillLoadable(t *testing.T) {
	// Inheritable-only targets should still be loadable directly (for building).
	_, err := LoadTarget(&Options{Target: "esp32"})
	if err != nil {
		t.Errorf("LoadTarget should still load inheritable-only target esp32: %v", err)
	}
}

func TestLoadTarget_FirmwareExtensionsInherited(t *testing.T) {
	for _, target := range []string{"pico", "pico2"} {
		t.Run(target, func(t *testing.T) {
			spec, err := LoadTarget(&Options{Target: target})
			if err != nil {
				t.Fatal(err)
			}
			want := []string{".elf", ".hex", ".uf2"}
			if !slices.Equal(spec.FirmwareExts, want) {
				t.Errorf("firmware extensions = %v, want %v", spec.FirmwareExts, want)
			}
			config := &Config{
				Options: &Options{Target: target},
				Target:  spec,
			}
			err = config.ValidateOutputExtension(".bin")
			wantErr := "target " + target + " does not support firmware output extension .bin; use .elf, .hex, or .uf2"
			if err == nil || err.Error() != wantErr {
				t.Errorf("ValidateOutputExtension(.bin) error = %v, want %q", err, wantErr)
			}
		})
	}
}

func TestLoadPY32ConcreteTargets(t *testing.T) {
	file, err := os.Open("../tools/gen-py32-targets/devices.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	records, err := csv.NewReader(file).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	for _, record := range records[1:] {
		part, core := record[0], record[2]
		t.Run(part, func(t *testing.T) {
			spec, err := LoadTarget(&Options{Target: part})
			if err != nil {
				t.Fatal(err)
			}
			if spec.LinkerScript == "" {
				t.Error("concrete target has no linker script")
			}
			wantCPU := "cortex-m0plus"
			if core == "m4" {
				wantCPU = "cortex-m4"
			}
			if spec.CPU != wantCPU {
				t.Errorf("CPU is %q, want %q", spec.CPU, wantCPU)
			}
		})
	}
}

func TestOverrideProperties(t *testing.T) {
	baseAutoStackSize := true
	base := &TargetSpec{
		GOOS:             "baseGoos",
		CPU:              "baseCpu",
		CFlags:           []string{"-base-foo", "-base-bar"},
		BuildTags:        []string{"bt1", "bt2"},
		DefaultStackSize: 42,
		AutoStackSize:    &baseAutoStackSize,
	}
	childAutoStackSize := false
	child := &TargetSpec{
		GOOS:             "",
		CPU:              "chlidCpu",
		CFlags:           []string{"-child-foo", "-child-bar"},
		AutoStackSize:    &childAutoStackSize,
		DefaultStackSize: 64,
	}

	base.overrideProperties(child)

	if base.GOOS != "baseGoos" {
		t.Errorf("Overriding failed : got %v", base.GOOS)
	}
	if base.CPU != "chlidCpu" {
		t.Errorf("Overriding failed : got %v", base.CPU)
	}
	if !reflect.DeepEqual(base.CFlags, []string{"-base-foo", "-base-bar", "-child-foo", "-child-bar"}) {
		t.Errorf("Overriding failed : got %v", base.CFlags)
	}
	if !reflect.DeepEqual(base.BuildTags, []string{"bt1", "bt2"}) {
		t.Errorf("Overriding failed : got %v", base.BuildTags)
	}
	if *base.AutoStackSize != false {
		t.Errorf("Overriding failed : got %v", base.AutoStackSize)
	}
	if base.DefaultStackSize != 64 {
		t.Errorf("Overriding failed : got %v", base.DefaultStackSize)
	}

	baseAutoStackSize = true
	base = &TargetSpec{
		AutoStackSize:    &baseAutoStackSize,
		DefaultStackSize: 42,
	}
	child = &TargetSpec{
		AutoStackSize:    nil,
		DefaultStackSize: 0,
	}
	base.overrideProperties(child)
	if *base.AutoStackSize != true {
		t.Errorf("Overriding failed : got %v", base.AutoStackSize)
	}
	if base.DefaultStackSize != 42 {
		t.Errorf("Overriding failed : got %v", base.DefaultStackSize)
	}

}

func TestOverridePropertiesFirmwareExtensions(t *testing.T) {
	base := &TargetSpec{FirmwareExts: []string{".uf2"}}
	child := &TargetSpec{FirmwareExts: []string{".hex"}}

	if err := base.overrideProperties(child); err != nil {
		t.Fatal(err)
	}
	want := []string{".uf2", ".hex"}
	if !slices.Equal(base.FirmwareExts, want) {
		t.Errorf("firmware extensions = %v, want %v", base.FirmwareExts, want)
	}
}

func TestConfigLinkerFlavor(t *testing.T) {
	tests := []struct {
		name   string
		target *TargetSpec
		goos   string
		want   string
	}{
		{
			name:   "default gnu",
			target: &TargetSpec{},
			goos:   "linux",
			want:   "gnu",
		},
		{
			name:   "default coff",
			target: &TargetSpec{},
			goos:   "windows",
			want:   "coff",
		},
		{
			name:   "default darwin",
			target: &TargetSpec{},
			goos:   "darwin",
			want:   "darwin",
		},
		{
			name: "target override",
			target: &TargetSpec{
				LinkerFlavor: "coff",
			},
			goos: "linux",
			want: "coff",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.target.GOOS = tc.goos
			config := &Config{
				Options: &Options{},
				Target:  tc.target,
			}
			if got := config.LinkerFlavor(); got != tc.want {
				t.Fatalf("LinkerFlavor() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestConfigPanicUnwind(t *testing.T) {
	tests := []struct {
		name    string
		options Options
		target  TargetSpec
		want    string
	}{
		{
			name:   "native defaults to setjmp",
			target: TargetSpec{Triple: "x86_64-unknown-linux"},
			want:   "setjmp",
		},
		{
			name:   "riscv64 defaults to setjmp",
			target: TargetSpec{Triple: "riscv64-unknown-unknown"},
			want:   "setjmp",
		},
		{
			name:   "xtensa defaults to setjmp",
			target: TargetSpec{Triple: "xtensa-unknown-unknown"},
			want:   "setjmp",
		},
		{
			name:    "explicit command line opt in",
			options: Options{PanicUnwind: "explicit"},
			target:  TargetSpec{Triple: "riscv64-unknown-unknown"},
			want:    "explicit",
		},
		{
			name:    "auto overrides target opt in",
			options: Options{PanicUnwind: "auto"},
			target:  TargetSpec{Triple: "riscv64-unknown-unknown", PanicUnwind: "explicit"},
			want:    "setjmp",
		},
		{
			name:    "asyncify enables unwinding",
			options: Options{Scheduler: "asyncify"},
			target:  TargetSpec{Triple: "wasm32-unknown-unknown", PanicUnwind: "auto"},
			want:    "asyncify",
		},
		{
			name:    "none disables asyncify unwinding",
			options: Options{Scheduler: "asyncify", PanicUnwind: "none"},
			target:  TargetSpec{Triple: "wasm32-unknown-unknown"},
			want:    "none",
		},
		{
			name:    "none disables setjmp unwinding",
			options: Options{PanicUnwind: "none"},
			target:  TargetSpec{Triple: "x86_64-unknown-linux"},
			want:    "none",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			config := Config{Options: &tc.options, Target: &tc.target}
			if got := config.PanicUnwind(); got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestLoadTargetLibc(t *testing.T) {
	tests := []struct {
		name    string
		options Options
	}{
		{"glibc with a target", Options{Target: "arduino", Libc: "glibc"}},
		{"musl with a target", Options{Target: "wasip1", Libc: "musl"}},
		{"glibc on windows", Options{GOOS: "windows", GOARCH: "amd64", Libc: "glibc"}},
		{"glibc on 386", Options{GOOS: "linux", GOARCH: "386", Libc: "glibc"}},
		{"glibc on arm", Options{GOOS: "linux", GOARCH: "arm", GOARM: "6", Libc: "glibc"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := LoadTarget(&tc.options); err == nil {
				t.Errorf("LoadTarget should reject -libc=%s", tc.options.Libc)
			}
		})
	}

	if _, err := LoadTarget(&Options{GOOS: "linux", GOARCH: "amd64", Libc: "musl"}); err != nil {
		t.Errorf("LoadTarget should accept -libc=musl on linux/amd64: %v", err)
	}
}

func TestGlibcLibDirs(t *testing.T) {
	tests := []struct {
		goarch, hostGOARCH string
		want               []string
	}{
		{"amd64", "amd64", []string{"/usr/lib/x86_64-linux-gnu", "/usr/lib64", "/usr/lib"}},
		{"arm64", "arm64", []string{"/usr/lib/aarch64-linux-gnu", "/usr/lib64", "/usr/lib"}},
		{"arm64", "amd64", []string{"/usr/lib/aarch64-linux-gnu"}},
		{"amd64", "arm64", []string{"/usr/lib/x86_64-linux-gnu"}},
	}
	for _, tc := range tests {
		got := glibcLibDirs(tc.goarch, tc.hostGOARCH)
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("glibcLibDirs(%q, %q) = %v, want %v", tc.goarch, tc.hostGOARCH, got, tc.want)
		}
	}
}

func TestGlibcLinkLibraries(t *testing.T) {
	for _, goarch := range []string{"amd64", "arm64"} {
		t.Run(goarch, func(t *testing.T) {
			if _, err := GlibcLibDir(goarch); err != nil {
				t.Skip(err)
			}
			target, err := LoadTarget(&Options{GOOS: "linux", GOARCH: goarch, Libc: "glibc"})
			if err != nil {
				t.Fatal(err)
			}
			for _, library := range []string{"-lm", "-lpthread"} {
				if !slices.Contains(target.LDFlags, library) {
					t.Errorf("glibc linker flags %v do not contain %s", target.LDFlags, library)
				}
			}
		})
	}
}

func TestLoadTargetExternalRelativePaths(t *testing.T) {
	boardJSON := filepath.Join("testdata", "exttarget", "board", "board.json")
	spec, err := LoadTarget(&Options{Target: boardJSON})
	if err != nil {
		t.Fatal(err)
	}

	absTestdata, err := filepath.Abs(filepath.Join("testdata", "exttarget"))
	if err != nil {
		t.Fatal(err)
	}
	wantLinker := filepath.Join(absTestdata, "board", "board.ld")
	if spec.LinkerScript != wantLinker {
		t.Errorf("LinkerScript = %q, want %q", spec.LinkerScript, wantLinker)
	}
	if spec.BoardPackage != "example.com/myboard" {
		t.Errorf("BoardPackage = %q, want %q", spec.BoardPackage, "example.com/myboard")
	}
	for _, wantFile := range []string{
		filepath.Join(absTestdata, "parent", "parent_init.c"),
		filepath.Join(absTestdata, "parent", "common.c"),
		"src/runtime/fallback_only.c",
		filepath.Join(absTestdata, "board", "board_init.c"),
		filepath.Join(absTestdata, "board", "common.c"),
	} {
		if !slices.Contains(spec.ExtraFiles, wantFile) {
			t.Errorf("ExtraFiles = %v, missing %q", spec.ExtraFiles, wantFile)
		}
	}

	t.Run("inherit board package", func(t *testing.T) {
		dir := t.TempDir()
		childPath := filepath.Join(dir, "child.json")
		parentAbs := filepath.Join(absTestdata, "parent", "base.json")
		childJSON := `{"inherits": [` + `"` + filepath.ToSlash(parentAbs) + `"` + `]}`
		if err := os.WriteFile(childPath, []byte(childJSON), 0o600); err != nil {
			t.Fatal(err)
		}
		childSpec, err := LoadTarget(&Options{Target: childPath})
		if err != nil {
			t.Fatal(err)
		}
		if childSpec.BoardPackage != "example.com/parentboard" {
			t.Errorf("BoardPackage = %q, want %q", childSpec.BoardPackage, "example.com/parentboard")
		}
	})

	t.Run("external wins over tinygoroot", func(t *testing.T) {
		dir := t.TempDir()
		extLinker := filepath.Join(dir, "targets", "pico.ld")
		if err := os.MkdirAll(filepath.Dir(extLinker), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(extLinker, []byte("/* external */"), 0o600); err != nil {
			t.Fatal(err)
		}
		targetPath := filepath.Join(dir, "board.json")
		targetJSON := `{"inherits": ["pico"], "linkerscript": "targets/pico.ld", "extra-files": ["targets/pico.ld"]}`
		if err := os.WriteFile(targetPath, []byte(targetJSON), 0o600); err != nil {
			t.Fatal(err)
		}
		extSpec, err := LoadTarget(&Options{Target: targetPath})
		if err != nil {
			t.Fatal(err)
		}
		if extSpec.LinkerScript != extLinker {
			t.Errorf("LinkerScript = %q, want %q", extSpec.LinkerScript, extLinker)
		}
		if !slices.Contains(extSpec.ExtraFiles, extLinker) {
			t.Errorf("ExtraFiles = %v, missing %q", extSpec.ExtraFiles, extLinker)
		}
	})

	t.Run("tinygoroot fallback", func(t *testing.T) {
		dir := t.TempDir()
		targetPath := filepath.Join(dir, "board.json")
		targetJSON := `{"inherits": ["pico"], "linkerscript": "targets/pico.ld", "extra-files": ["src/runtime/missing_rel.c"]}`
		if err := os.WriteFile(targetPath, []byte(targetJSON), 0o600); err != nil {
			t.Fatal(err)
		}
		extSpec, err := LoadTarget(&Options{Target: targetPath})
		if err != nil {
			t.Fatal(err)
		}
		if extSpec.LinkerScript != "targets/pico.ld" {
			t.Errorf("LinkerScript = %q, want %q", extSpec.LinkerScript, "targets/pico.ld")
		}
		if !slices.Contains(extSpec.ExtraFiles, "src/runtime/missing_rel.c") {
			t.Errorf("ExtraFiles = %v, missing fallback path", extSpec.ExtraFiles)
		}
	})

	t.Run("cwd fallback for json inherits", func(t *testing.T) {
		cwdDir := t.TempDir()
		targetDir := t.TempDir()
		cwdParent := filepath.Join(cwdDir, "cwd_parent.json")
		if err := os.WriteFile(cwdParent, []byte(`{"inherits": ["cortex-m"], "board-package": "example.com/cwd"}`), 0o600); err != nil {
			t.Fatal(err)
		}
		childPath := filepath.Join(targetDir, "child.json")
		if err := os.WriteFile(childPath, []byte(`{"inherits": ["cwd_parent.json"]}`), 0o600); err != nil {
			t.Fatal(err)
		}
		t.Chdir(cwdDir)
		extSpec, err := LoadTarget(&Options{Target: childPath})
		if err != nil {
			t.Fatal(err)
		}
		if extSpec.BoardPackage != "example.com/cwd" {
			t.Errorf("BoardPackage = %q, want %q", extSpec.BoardPackage, "example.com/cwd")
		}
	})
}
