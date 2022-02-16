package interp

import (
	"io/ioutil"
	"os"
	"testing"

	"tinygo.org/x/go-llvm"
)

func TestCompile(t *testing.T) {
	t.Parallel()

	// Read the input IR.
	ctx := llvm.NewContext()
	buf, err := llvm.NewMemoryBufferFromFile("testdata/testcompile.ll")
	os.Stat("testcompile.ll") // make sure this file is tracked by `go test` caching
	if err != nil {
		t.Fatalf("could not read file: %v", err)
	}
	mod, err := ctx.ParseIR(buf)
	if err != nil {
		t.Fatalf("could not load module:\n%v", err)
	}

	// Compile everything.
	cc := &constParser{
		td:     llvm.NewTargetData(mod.DataLayout()),
		tCache: make(map[llvm.Type]typ),
		vCache: make(map[llvm.Value]value),
		fCache: make(map[llvm.Type]fnTyInfo),
	}
	err = cc.mapGlobals(mod)
	if err != nil {
		t.Errorf("failed to map globals: %s", err.Error())
		return
	}
	for fn := mod.FirstFunction(); !fn.IsNil(); fn = llvm.NextFunction(fn) {
		if fn.IsDeclaration() {
			continue
		}
		name := fn.Name()
		t.Run(name, func(t *testing.T) {
			sig, err := cc.parseFuncSignature(fn)
			if err != nil {
				t.Errorf("failed to parse signature: %s", err.Error())
				return
			}
			r, err := compile(cc, sig, fn)
			if err != nil {
				t.Errorf("compilation failed: %s", err.Error())
				return
			}
			got := r.String()
			t.Parallel()
			path := "testdata/testcompile_" + name + ".txt"
			if *update {
				err = ioutil.WriteFile(path, []byte(got), 0644)
				if err != nil {
					t.Errorf("failed to write updated output: %s", err.Error())
				}
			} else {
				data, err := ioutil.ReadFile(path)
				if err != nil {
					t.Errorf("failed to read expected output file: %s", err.Error())
					return
				}
				expected := string(data)
				if got != expected {
					t.Error("unexpected result")
					t.Error(got)
				}
			}
		})
	}
}
