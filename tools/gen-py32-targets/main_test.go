package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDeviceTable(t *testing.T) {
	devices, err := readDevices("devices.csv")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(devices), 87; got != want {
		t.Fatalf("got %d concrete devices, want %d", got, want)
	}

	families := make(map[string]bool)
	for family := range orphanFamilies {
		families[family] = true
	}
	m4 := 0
	for _, device := range devices {
		families[device.Family] = true
		if device.Core == "m4" {
			m4++
		}
	}
	if got, want := len(families), 41; got != want {
		t.Fatalf("got %d SVD families, want %d", got, want)
	}
	if got, want := m4, 7; got != want {
		t.Fatalf("got %d Cortex-M4 devices, want %d", got, want)
	}
}

func TestSystemStackSize(t *testing.T) {
	tests := []struct {
		ramSize uint64
		want    uint64
	}{
		{2 * 1024, 1024},
		{4 * 1024, 1024},
		{4*1024 + 1, 2 * 1024},
		{16 * 1024, 2 * 1024},
		{16*1024 + 1, 4 * 1024},
		{144 * 1024, 4 * 1024},
	}
	for _, test := range tests {
		if got := systemStackSize(test.ramSize); got != test.want {
			t.Errorf("systemStackSize(%d) = %d, want %d", test.ramSize, got, test.want)
		}
	}
}

func TestGenerate(t *testing.T) {
	devices, err := readDevices("devices.csv")
	if err != nil {
		t.Fatal(err)
	}
	out := t.TempDir()
	if err := generate(out, devices); err != nil {
		t.Fatal(err)
	}

	entries, err := os.ReadDir(out)
	if err != nil {
		t.Fatal(err)
	}
	// 41 family JSON files, 77 generated concrete JSON files, and linker
	// scripts for all 87 concrete targets.
	if got, want := len(entries), 205; got != want {
		t.Fatalf("generated %d files, want %d", got, want)
	}

	checkTarget(t, filepath.Join(out, "py32e407xc.json"), "py32e407xx", "targets/py32e407xc.ld")
	checkTarget(t, filepath.Join(out, "py32f001xx.json"), "py32", "")
	checkTarget(t, filepath.Join(out, "py32f410xx.json"), "py32-m4", "")
	checkBuildTags(t, filepath.Join(out, "py32e407xx.json"), "py32e407xx", "py32_gpio_ospdder", "py32_gpio_clock_ahb2", "py32_no_hsi_fs", "py32_usart_split_data", "py32_uart_clock_apb2")
	checkBuildTags(t, filepath.Join(out, "py32f001cxx.json"), "py32f001cxx", "no_gpio_afrh", "py32_hsi_fs_literal4", "py32_usart1_clock_literal")
	checkBuildTags(t, filepath.Join(out, "py32f002cxx.json"), "py32f002cxx", "no_gpio_afrh", "py32_hsi_fs_literal4")
	checkBuildTags(t, filepath.Join(out, "py32f032xx.json"), "py32f032xx", "py32_hsi_fs_literal3")
	checkBuildTags(t, filepath.Join(out, "py32f410xx.json"), "py32f410xx", "py32_gpio_ospdder", "py32_gpio_clock_ahb", "py32_no_hsi_fs", "py32_usart_unnumbered", "py32_uart_clock_apb2", "py32_uart_no_interrupt")
	checkBuildTags(t, filepath.Join(out, "py32t020xx.json"), "py32t020xx", "py32_uart_type")

	data, err := os.ReadFile(filepath.Join(out, "py32e407xc.ld"))
	if err != nil {
		t.Fatal(err)
	}
	want := "ORIGIN = 0x20000000, LENGTH = 0x1c000"
	if !contains(string(data), want) {
		t.Fatalf("E407 linker script does not contain combined contiguous RAM %q", want)
	}
	if _, err := os.Stat(filepath.Join(out, "py32f030x8.json")); !os.IsNotExist(err) {
		t.Fatalf("legacy target was generated: %v", err)
	}
	checkFileContains(t, filepath.Join(out, "py32f002ax5.ld"), "_stack_size = 1K;")
	checkFileContains(t, filepath.Join(out, "py32f030x8.ld"), "_stack_size = 2K;")
	checkFileContains(t, filepath.Join(out, "py32l090xc.ld"), "_stack_size = 4K;")
}

func checkFileContains(t *testing.T, path, want string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), want) {
		t.Errorf("%s does not contain %q", path, want)
	}
}

func checkBuildTags(t *testing.T, path string, want ...string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var spec target
	if err := json.Unmarshal(data, &spec); err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(spec.BuildTags, ","); got != strings.Join(want, ",") {
		t.Errorf("%s build tags are %q, want %q", path, got, strings.Join(want, ","))
	}
}

func checkTarget(t *testing.T, path, inherited, linker string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var spec target
	if err := json.Unmarshal(data, &spec); err != nil {
		t.Fatal(err)
	}
	if len(spec.Inherits) != 1 || spec.Inherits[0] != inherited {
		t.Errorf("%s inherits %v, want %q", path, spec.Inherits, inherited)
	}
	if spec.LinkerScript != linker {
		t.Errorf("%s linker script is %q, want %q", path, spec.LinkerScript, linker)
	}
}

func contains(value, substring string) bool {
	for i := 0; i+len(substring) <= len(value); i++ {
		if value[i:i+len(substring)] == substring {
			return true
		}
	}
	return false
}
