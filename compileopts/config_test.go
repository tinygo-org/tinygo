package compileopts

import (
	"slices"
	"testing"
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
