package main

import (
	"bytes"
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"github.com/tinygo-org/tinygo/builder"
	"github.com/tinygo-org/tinygo/goenv"
	"github.com/tinygo-org/tinygo/interp"
)

func TestRP2PLLTable(t *testing.T) {
	tmpdir := t.TempDir()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filepath.Join(goenv.Get("TINYGOROOT"), "src/machine/machine_rp2_pll.go"), nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var source bytes.Buffer
	source.WriteString("package main\nimport \"math\"\n")
	names := map[string]bool{
		"abs": false, "pdTable": false,
		"genTable": false, "genTableEntry": false,
	}
	for _, decl := range file.Decls {
		var name string
		switch decl := decl.(type) {
		case *ast.FuncDecl:
			name = decl.Name.Name
		case *ast.GenDecl:
			if len(decl.Specs) != 1 {
				continue
			}
			switch spec := decl.Specs[0].(type) {
			case *ast.TypeSpec:
				name = spec.Name.Name
			case *ast.ValueSpec:
				if len(spec.Names) == 1 {
					name = spec.Names[0].Name
				}
			}
		}
		if _, ok := names[name]; !ok {
			continue
		}
		if err := printer.Fprint(&source, fset, decl); err != nil {
			t.Fatal(err)
		}
		source.WriteByte('\n')
		names[name] = true
	}
	for name, found := range names {
		if !found {
			t.Fatalf("missing PLL declaration %s", name)
		}
	}
	fixture, err := os.ReadFile("testdata/rp2-pll.go")
	if err != nil {
		t.Fatal(err)
	}
	source.Write(bytes.TrimPrefix(fixture, []byte("package main\n")))
	path := filepath.Join(tmpdir, "main.go")
	if err := os.WriteFile(path, source.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}
	options := optionsFromTarget("wasip1", sema)
	options.Opt = "1"
	options.Scheduler = "none"
	options.InterpMaxLoopIterations = interp.DefaultMaxInterpBlockEntries
	config, err := builder.NewConfig(&options)
	if err != nil {
		t.Fatal(err)
	}
	irPath := filepath.Join(tmpdir, "main.ll")
	if _, err := builder.Build(path, irPath, tmpdir, config); err != nil {
		t.Fatal(err)
	}
	ir, err := os.ReadFile(irPath)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(ir, []byte("@main.checksum")) {
		t.Fatal("PLL checksum was not pre-evaluated")
	}
	result, err := builder.Build(path, filepath.Join(tmpdir, "main.wasm"), tmpdir, config)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(result.Binary)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	r := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfigInterpreter())
	defer r.Close(ctx)
	if _, err := wasi_snapshot_preview1.Instantiate(ctx, r); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	_, err = r.InstantiateWithConfig(ctx, data, wazero.NewModuleConfig().WithStdout(&output).WithStderr(&output))
	if err != nil {
		t.Fatalf("PLL table check failed: %v\n%s", err, &output)
	}
	if output.String() != "961\n" {
		t.Fatalf("unexpected PLL checksum: %q", &output)
	}
}

func TestRP2PLLPreEvaluation(t *testing.T) {
	for _, tc := range []struct {
		target  string
		pllSys  uint32
		fbdiv   uint32
		postdiv uint32
	}{
		{"pico", 0x40028000, 100, 6<<16 | 1<<12},
		{"pico2", 0x40050000, 125, 5<<16 | 2<<12},
	} {
		t.Run(tc.target, func(t *testing.T) {
			tmpdir := t.TempDir()
			options := optionsFromTarget(tc.target, sema)
			options.Opt = "1"
			options.InterpMaxLoopIterations = interp.DefaultMaxInterpBlockEntries
			config, err := builder.NewConfig(&options)
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(tmpdir, "blinky.ll")
			if _, err := builder.Build("examples/blinky1", path, tmpdir, config); err != nil {
				t.Fatal(err)
			}
			ir, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			for _, name := range []string{
				"machine.pdTable", "machine.pllsysFB", "machine.pllsysPD1", "machine.pllsysPD2",
				"machine.genTable", "machine.pllFreqOutPostdiv",
			} {
				if strings.Contains(string(ir), "@"+name) {
					t.Errorf("PLL calculation was not pre-evaluated: %s remains", name)
				}
			}
			for _, reg := range []struct {
				offset uint32
				value  uint32
			}{{0, 1}, {8, tc.fbdiv}, {12, tc.postdiv}} {
				store := fmt.Sprintf("store volatile i32 %d, ptr inttoptr (i32 %d to ptr)", reg.value, tc.pllSys+reg.offset)
				if !strings.Contains(string(ir), store) {
					t.Errorf("missing PLL_SYS register write: %s", store)
				}
			}
		})
	}
}
