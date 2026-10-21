package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tinygo-org/tinygo/goenv"
)

func fixCommand(t *testing.T, dir string, args ...string) (int, string) {
	t.Helper()
	cmd := exec.Command(os.Args[0], append([]string{"test-main", "fix"}, args...)...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err == nil {
		return 0, string(output)
	}
	if exit, ok := err.(*exec.ExitError); ok {
		return exit.ExitCode(), string(output)
	}
	t.Fatal(err)
	return 0, ""
}

func TestFixCommand(t *testing.T) {
	root, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("TINYGOROOT", root)
	t.Setenv("GOWORK", "off")
	t.Setenv("GOFLAGS", "")
	dir := t.TempDir()
	source := "package probe\n\nvar X interface{}\n"
	path := filepath.Join(dir, "probe.go")
	for name, content := range map[string]string{
		"go.mod": "module example.com/probe\n\ngo 1.25.0\n", "probe.go": source,
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
	}
	status, output := fixCommand(t, root, "-C", dir, "-diff", "-any")
	if status != 1 || !strings.Contains(output, "+var X any") {
		t.Fatalf("diff status=%d:\n%s", status, output)
	}
	before, err := os.ReadFile(path)
	if err != nil || string(before) != source {
		t.Fatalf("diff changed source: %q, %v", before, err)
	}
	status, output = fixCommand(t, dir, "-any")
	if status != 0 {
		t.Fatalf("write status=%d:\n%s", status, output)
	}
	after, err := os.ReadFile(path)
	if err != nil || !bytes.Contains(after, []byte("var X any")) {
		t.Fatalf("fix did not persist: %q, %v", after, err)
	}
	t.Run("go fix parity", func(t *testing.T) {
		_, minor, err := goenv.GetGorootVersion()
		if err != nil {
			t.Fatal(err)
		}
		if minor < 26 {
			t.Skip("Go 1.25 does not have the analyzer-based fix command")
		}
		reference := t.TempDir()
		for name, content := range map[string]string{
			"go.mod": "module example.com/reference\n\ngo 1.25.0\n", "probe.go": source,
		} {
			if err := os.WriteFile(filepath.Join(reference, name), []byte(content), 0600); err != nil {
				t.Fatal(err)
			}
		}
		cmd := exec.Command(filepath.Join(goenv.Get("GOROOT"), "bin", "go"), "fix", "-any")
		cmd.Dir = reference
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("go fix: %v\n%s", err, output)
		}
		want, err := os.ReadFile(filepath.Join(reference, "probe.go"))
		if err != nil || !bytes.Equal(after, want) {
			t.Fatalf("tinygo fix differs from go fix:\ngot %s\nwant %s\n%v", after, want, err)
		}
	})
	status, output = fixCommand(t, dir, "-diff", "-any")
	if status != 0 || output != "" {
		t.Fatalf("second run status=%d:\n%s", status, output)
	}
}

func TestFixHelpAndFlags(t *testing.T) {
	for _, test := range []struct {
		args   []string
		status int
		want   string
	}{
		{[]string{"help"}, 0, "rangeint"},
		{[]string{"help", "any"}, 0, "interface"},
		{[]string{"help", "missing"}, 2, "unknown analyzer"},
		{[]string{"-fixtool=other"}, 2, "flag provided but not defined"},
		{[]string{"-toolexec=other"}, 2, "flag provided but not defined"},
		{[]string{"-tests=false", "-help"}, 0, "include test files"},
		{[]string{"-diff", "./does-not-exist"}, 1, "does-not-exist"},
		{[]string{"-diff", "main.go"}, 1, "select a package"},
	} {
		status, output := fixCommand(t, ".", test.args...)
		if status != test.status || !strings.Contains(output, test.want) {
			t.Fatalf("%v: status=%d, want=%d:\n%s", test.args, status, test.status, output)
		}
	}
}

func TestFixExcludesUnbuildableTargetTests(t *testing.T) {
	root, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("TINYGOROOT", root)
	t.Setenv("GOWORK", "off")
	t.Setenv("GOFLAGS", "")
	dir := t.TempDir()
	files := map[string]string{
		"go.mod":        "module example.com/probe\n\ngo 1.25.0\n",
		"probe.go":      "package probe\nvar X interface{}\n",
		"probe_test.go": "//go:build tinygo\n\npackage probe\nvar _ = missingTargetAPI\n",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
	}
	status, output := fixCommand(t, dir, "-diff", "-any")
	if status != 1 || !strings.Contains(output, "missingTargetAPI") {
		t.Fatalf("default test analysis status=%d, want target test type error:\n%s", status, output)
	}
	status, output = fixCommand(t, dir, "-tests=false", "-diff", "-any")
	if status != 1 || !strings.Contains(output, "+var X any") ||
		strings.Contains(output, "missingTargetAPI") {
		t.Fatalf("fix without tests status=%d:\n%s", status, output)
	}
}

func TestFixTinyGoSource(t *testing.T) {
	for _, args := range [][]string{
		{"-diff", "-any", "./src/internal/itoa"},
		{"-diff", "internal/itoa"},
		{"-target=cortex-m-qemu", "-diff", "-any", "internal/itoa"},
	} {
		status, output := fixCommand(t, ".", args...)
		if status != 0 {
			t.Fatalf("%v: status=%d:\n%s", args, status, output)
		}
	}
	status, output := fixCommand(t, ".", "-diff", "runtime")
	if status != 1 || !strings.Contains(output, "--- ") || !strings.Contains(output, "/src/runtime/") {
		t.Fatalf("runtime default suite: status=%d:\n%s", status, output)
	}
}

func TestFixTargetSelection(t *testing.T) {
	root, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("TINYGOROOT", root)
	t.Setenv("GOWORK", "off")
	t.Setenv("GOFLAGS", "")
	dir := t.TempDir()
	files := map[string]string{
		"go.mod":      "module example.com/target\n\ngo 1.25.0\n",
		"common.go":   "package probe\nvar Common interface{}\n",
		"selected.go": "//go:build selected\n\npackage probe\nvar Selected interface{}\n",
		"other.go":    "//go:build !selected\n\npackage probe\nvar Other interface{}\n",
		"avr.go": "//go:build avr\n\npackage probe\nimport \"unsafe\"\n" +
			"const IntSize = unsafe.Sizeof(int(0))\nconst PointerSize = unsafe.Sizeof(uintptr(0))\n" +
			"var _ [4-IntSize]byte\nvar _ [IntSize-4]byte\nvar _ [2-PointerSize]byte\nvar _ [PointerSize-2]byte\n" +
			"var AVR interface{}\n",
	}
	for name, source := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(source), 0600); err != nil {
			t.Fatal(err)
		}
	}
	for _, args := range [][]string{{"-tags=selected", "-any"}, {"-target=arduino", "-any"}} {
		status, output := fixCommand(t, dir, args...)
		if status != 0 {
			t.Fatalf("%v: status=%d:\n%s", args, status, output)
		}
		selected, err := os.ReadFile(filepath.Join(dir, "selected.go"))
		if err != nil || !bytes.Contains(selected, []byte("Selected any")) {
			t.Fatalf("selected file was not fixed: %s\n%v", selected, err)
		}
		other, err := os.ReadFile(filepath.Join(dir, "other.go"))
		if err != nil {
			t.Fatal(err)
		}
		if args[0] == "-tags=selected" && string(other) != files["other.go"] {
			t.Fatal("file excluded by build tags was changed")
		}
	}
	avr, err := os.ReadFile(filepath.Join(dir, "avr.go"))
	if err != nil || !bytes.Contains(avr, []byte("AVR any")) {
		t.Fatalf("AVR source was not fixed: %s\n%v", avr, err)
	}
}
