package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/pprof/profile"
)

func TestParseObjdump(t *testing.T) {
	text := `
00000020 wasm[0]::function[1]::main.f:
        20: stp     x29, x30, [sp, #-0x10]!
        24: mov     w0, #1
            ╰─╼ addrmap: 0x105
        28: bl      #0x40
            ╰─╼ addrmap: 0x10a
        2c: ret
00000040 wasm[0]::function[2]:
        40: mov     w0, #2
            ╰─╼ addrmap: 0x120
        44: ret
`
	am, err := parseObjdump([]byte(text))
	if err != nil {
		t.Fatal(err)
	}
	if am.funcs[0x20] != "main.f" || am.funcs[0x40] != "" {
		t.Errorf("funcs = %v", am.funcs)
	}
	for _, tc := range []struct{ native, wasm int64 }{
		{0x10, -1}, // before any function
		{0x20, -1}, // prologue, no location yet
		{0x24, 0x105},
		{0x2b, 0x10a}, // return address - 1 of the bl
		{0x2c, 0x10a},
		{0x40, 0x120}, // next function's entry has its own entry
	} {
		if got := am.lookup(tc.native); got != tc.wasm {
			t.Errorf("lookup(%#x) = %#x, want %#x", tc.native, got, tc.wasm)
		}
	}
}

func TestSampleOffset(t *testing.T) {
	m := &wasmModule{data: []byte{
		0x10, 0x05, // call 5
		0x03, 0x40, // loop (empty block type)
		0x20, 0x00, // local.get 0
		0x03, 0x80, 0x01, // loop (type index 128)
		0x20, 0x00,
	}}
	for _, tc := range []struct{ in, want int64 }{
		{0, 0},
		{2, 4},
		{6, 9},
		{-1, -1},
	} {
		if got := m.sampleOffset(tc.in); got != tc.want {
			t.Errorf("sampleOffset(%d) = %d, want %d", tc.in, got, tc.want)
		}
	}
}

func TestIsTombstone(t *testing.T) {
	for addr, want := range map[uint64]bool{0: true, 1: false, 0x1234: false, 0xfffffffe: true, 0xffffffff: true} {
		if got := isTombstone(addr); got != want {
			t.Errorf("isTombstone(%#x) = %v", addr, got)
		}
	}
}

// TestEndToEnd checks that a hot loop is attributed to the right line. It
// needs tinygo and wasmtime on PATH.
func TestEndToEnd(t *testing.T) {
	if testing.Short() {
		t.Skip("short mode")
	}
	for _, tool := range []string{"tinygo", "wasmtime"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("%s not found", tool)
		}
	}
	dir := t.TempDir()
	src := `package main

import "fmt"

//go:noinline
func hot(n int) uint32 {
	x := uint32(1)
	for i := 0; i < n; i++ {
		x = x*1664525 + 1013904223
	}
	return x
}

func main() {
	fmt.Println(hot(400_000_000))
}
`
	const hotLine = 9 // x = x*1664525 + 1013904223
	write := func(name, s string) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(s), 0o644); err != nil {
			t.Fatal(err)
		}
		return p
	}
	mainGo := write("main.go", src)
	wasm := filepath.Join(dir, "main.wasm")
	prof := filepath.Join(dir, "guest.json")
	out := filepath.Join(dir, "cpu.pprof")
	runCmd := func(name string, args ...string) {
		t.Helper()
		cmd := exec.Command(name, args...)
		cmd.Dir = dir
		if b, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%s %s: %v\n%s", name, strings.Join(args, " "), err, b)
		}
	}
	runCmd("tinygo", "build", "-o", wasm, "-target=wasip1", "-buildmode=wasi-legacy", mainGo)
	runCmd("wasmtime", "run", "--profile=guest,"+prof+",1ms", wasm)
	if err := run(wasm, prof, out, "wasmtime", nil, false, false); err != nil {
		t.Fatal(err)
	}

	f, err := os.Open(out)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	p, err := profile.Parse(f)
	if err != nil {
		t.Fatal(err)
	}
	var total, hit int64
	for _, s := range p.Sample {
		total += s.Value[0]
		leaf := s.Location[0].Line[0]
		if leaf.Function.Name == "main.hot" && leaf.Line == hotLine {
			hit += s.Value[0]
		}
	}
	if total < 20 {
		t.Fatalf("only %d samples", total)
	}
	if hit*2 < total {
		t.Errorf("main.hot line %d has %d of %d samples, want most", hotLine, hit, total)
		for _, s := range p.Sample {
			l := s.Location[0].Line[0]
			t.Logf("%d %s:%d", s.Value[0], l.Function.Name, l.Line)
		}
	}
}
