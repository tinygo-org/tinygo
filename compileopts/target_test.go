package compileopts

import (
	"encoding/csv"
	"errors"
	"io/fs"
	"os"
	"reflect"
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
