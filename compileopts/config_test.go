package compileopts

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/tinygo-org/tinygo/goenv"
)

func TestExtraFilesBoehm(t *testing.T) {
	target := &TargetSpec{
		GC: "precise",
		ExtraFiles: []string{
			"src/runtime/asm_tinygowasm.S",
		},
	}
	config := &Config{
		Options: &Options{},
		Target:  target,
	}

	got := config.ExtraFiles()
	want := []string{"src/runtime/asm_tinygowasm.S"}
	if !slices.Equal(got, want) {
		t.Fatalf("unexpected precise GC files: got %v, want %v", got, want)
	}

	config.Options.GC = "boehm"
	got = config.ExtraFiles()
	want = []string{
		"src/runtime/asm_tinygowasm.S",
		"src/runtime/gc_boehm.c",
	}
	if !slices.Equal(got, want) {
		t.Fatalf("unexpected Boehm GC files: got %v, want %v", got, want)
	}
}

func TestValidateOutputExtension(t *testing.T) {
	tests := []struct {
		name      string
		target    TargetSpec
		extension string
		wantErr   string
	}{
		{
			name: "supported firmware extension",
			target: TargetSpec{
				FirmwareExts: []string{".elf", ".hex", ".uf2"},
			},
			extension: ".uf2",
		},
		{
			name: "unsupported firmware extension",
			target: TargetSpec{
				FirmwareExts: []string{".elf", ".hex", ".uf2"},
			},
			extension: ".bin",
			wantErr:   "target pico does not support firmware output extension .bin; use .elf, .hex, or .uf2",
		},
		{
			name: "object output",
			target: TargetSpec{
				FirmwareExts: []string{".elf", ".hex", ".uf2"},
			},
			extension: ".o",
		},
		{
			name: "bitcode output",
			target: TargetSpec{
				FirmwareExts: []string{".elf", ".hex", ".uf2"},
			},
			extension: ".bc",
		},
		{
			name: "LLVM IR output",
			target: TargetSpec{
				FirmwareExts: []string{".elf", ".hex", ".uf2"},
			},
			extension: ".ll",
		},
		{
			name: "unknown extension",
			target: TargetSpec{
				FirmwareExts: []string{".elf", ".hex", ".uf2"},
			},
			extension: ".unknown",
		},
		{
			name: "extensionless output",
			target: TargetSpec{
				FirmwareExts: []string{".elf", ".hex", ".uf2"},
			},
		},
		{
			name: "target without restrictions",
			target: TargetSpec{
				BinaryFormat: "esp32",
			},
			extension: ".bin",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			config := &Config{
				Options: &Options{Target: "pico"},
				Target:  &tc.target,
			}
			err := config.ValidateOutputExtension(tc.extension)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("ValidateOutputExtension(%q) failed: %v", tc.extension, err)
				}
				return
			}
			if err == nil || err.Error() != tc.wantErr {
				t.Fatalf("ValidateOutputExtension(%q) error = %v, want %q", tc.extension, err, tc.wantErr)
			}
		})
	}
}

func TestBinaryFormatFirmwareExtensions(t *testing.T) {
	config := &Config{
		Options: &Options{},
		Target: &TargetSpec{
			BinaryFormat: "esp32",
		},
	}
	for ext, want := range map[string]string{
		".bin": "esp32",
		".hex": "hex",
		".uf2": "uf2",
		".elf": "elf",
	} {
		if got := config.BinaryFormat(ext); got != want {
			t.Errorf("BinaryFormat(%q) = %q, want %q", ext, got, want)
		}
	}
	config.Target.BinaryFormat = ""
	if got := config.BinaryFormat(".bin"); got != "bin" {
		t.Errorf("BinaryFormat(.bin) without a target format = %q, want %q", got, "bin")
	}
}

func TestEmulatorCommand(t *testing.T) {
	const binary = `C:\build dir\firmware.img`
	config := &Config{Target: &TargetSpec{
		Emulator: `'C:\Program Files\qemu.exe' --image={img} --root='{root}/two words' --tmp={tmpDir} --label=a\ b # ignored`,
	}}
	got, err := config.Emulator("img", binary)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		`C:\Program Files\qemu.exe`,
		"--image=" + binary,
		"--root=" + goenv.Get("TINYGOROOT") + "/two words",
		"--tmp=" + os.TempDir(),
		"--label=a b",
	}
	if !slices.Equal(got, want) {
		t.Errorf("Emulator() = %q, want %q", got, want)
	}
}

func TestEmulatorCommandInvalidQuoting(t *testing.T) {
	config := &Config{Target: &TargetSpec{Emulator: `qemu "unfinished`}}
	_, err := config.Emulator("", "image")
	if err == nil || err.Error() != "could not parse emulator command: EOF found when expecting closing quote" {
		t.Fatalf("Emulator() error = %v", err)
	}
}

func TestEmulatorCommandBuiltin(t *testing.T) {
	target, err := LoadTarget(&Options{Target: "wasm"})
	if err != nil {
		t.Fatal(err)
	}
	config := &Config{Target: target}
	const binary = `C:\build dir\program.wasm`
	got, err := config.Emulator("", binary)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"node", goenv.Get("TINYGOROOT") + "/targets/wasm_exec_node.js", binary}
	if !slices.Equal(got, want) {
		t.Errorf("Emulator() = %q, want %q", got, want)
	}
}

func TestLDFlagsWasmStackSize(t *testing.T) {
	config := &Config{
		Options: &Options{
			Scheduler: "none",
			StackSize: 16 << 20,
		},
		Target: &TargetSpec{
			Linker:    "wasm-ld",
			Scheduler: "asyncify",
		},
	}
	if got := config.LDFlags(); !slices.Contains(got, "stack-size=16777216") {
		t.Fatalf("expected stack-size=16777216 in LDFlags for scheduler=none, got %v", got)
	}

	config.Options.Scheduler = "asyncify"
	if got := config.LDFlags(); slices.Contains(got, "stack-size=16777216") {
		t.Fatalf("unexpected stack-size in LDFlags for scheduler=asyncify, got %v", got)
	}
}

func TestBuildTagsBoardPackage(t *testing.T) {
	t.Run("auto append", func(t *testing.T) {
		config := &Config{
			Options: &Options{},
			Target: &TargetSpec{
				BoardPackage: "example.com/board",
			},
		}
		if got := config.BuildTags(); !slices.Contains(got, "boardpackage") {
			t.Errorf("BuildTags() = %v, missing boardpackage", got)
		}
	})

	t.Run("no duplicate from target or options", func(t *testing.T) {
		for _, tc := range []struct {
			name    string
			tags    []string
			optTags []string
		}{
			{name: "target", tags: []string{"boardpackage"}},
			{name: "options", optTags: []string{"boardpackage"}},
		} {
			t.Run(tc.name, func(t *testing.T) {
				config := &Config{
					Options: &Options{Tags: tc.optTags},
					Target: &TargetSpec{
						BuildTags:    tc.tags,
						BoardPackage: "example.com/board",
					},
				}
				count := 0
				for _, tag := range config.BuildTags() {
					if tag == "boardpackage" {
						count++
					}
				}
				if count != 1 {
					t.Errorf("boardpackage appeared %d times in %v, want 1", count, config.BuildTags())
				}
			})
		}
	})

	t.Run("omitted when empty", func(t *testing.T) {
		config := &Config{
			Options: &Options{},
			Target:  &TargetSpec{},
		}
		if got := config.BuildTags(); slices.Contains(got, "boardpackage") {
			t.Errorf("BuildTags() = %v, unexpected boardpackage", got)
		}
	})
}

func TestLDFlagsAbsoluteLinkerScript(t *testing.T) {
	t.Run("absolute linker script adds directory before TINYGOROOT", func(t *testing.T) {
		absDir := t.TempDir()
		absScript := filepath.Join(absDir, "custom.ld")
		config := &Config{
			Options: &Options{},
			Target: &TargetSpec{
				LinkerScript: absScript,
			},
		}
		flags := config.LDFlags()
		root := goenv.Get("TINYGOROOT")
		extLIdx, rootLIdx, tIdx := -1, -1, -1
		for i := 0; i < len(flags)-1; i++ {
			if flags[i] == "-L" && flags[i+1] == absDir {
				extLIdx = i
			}
			if flags[i] == "-L" && flags[i+1] == root {
				rootLIdx = i
			}
			if flags[i] == "-T" && flags[i+1] == absScript {
				tIdx = i
			}
		}
		if extLIdx < 0 || rootLIdx < 0 || extLIdx > rootLIdx {
			t.Fatalf("LDFlags() = %v, want -L %s before -L %s", flags, absDir, root)
		}
		if tIdx < 0 {
			t.Fatalf("LDFlags() = %v, missing -T %s", flags, absScript)
		}
	})

	t.Run("relative linker script keeps only TINYGOROOT -L", func(t *testing.T) {
		config := &Config{
			Options: &Options{},
			Target: &TargetSpec{
				LinkerScript: "targets/arm.ld",
			},
		}
		flags := config.LDFlags()
		lCount := 0
		hasT := false
		for i := 0; i < len(flags)-1; i++ {
			if flags[i] == "-L" {
				lCount++
			}
			if flags[i] == "-T" && flags[i+1] == "targets/arm.ld" {
				hasT = true
			}
		}
		if lCount != 1 {
			t.Fatalf("LDFlags() = %v, got %d -L flags, want 1", flags, lCount)
		}
		if !hasT {
			t.Fatalf("LDFlags() = %v, missing -T targets/arm.ld", flags)
		}
	})
}

func TestConfigTargetDir(t *testing.T) {
	extDir := t.TempDir()
	extTarget := filepath.Join(extDir, "board.json")
	extConfig := &Config{
		Options: &Options{Target: extTarget},
	}
	if got := extConfig.TargetDir(); got != extDir {
		t.Errorf("TargetDir() for external target = %q, want %q", got, extDir)
	}

	builtinName := &Config{
		Options: &Options{Target: "pico"},
	}
	if got := builtinName.TargetDir(); got != "" {
		t.Errorf("TargetDir() for built-in name = %q, want empty", got)
	}

	builtinPath := filepath.Join(goenv.Get("TINYGOROOT"), "targets", "pico.json")
	builtinPathCfg := &Config{
		Options: &Options{Target: builtinPath},
	}
	if got := builtinPathCfg.TargetDir(); got != "" {
		t.Errorf("TargetDir() for built-in path = %q, want empty", got)
	}
}
