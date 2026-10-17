package builder

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"tinygo.org/x/go-llvm"
)

func TestCGoHeaderCompileIsPathIndependent(t *testing.T) {
	var outputs [][]byte
	for _, dirName := range []string{"first", "second"} {
		dir := filepath.Join(t.TempDir(), dirName)
		if err := os.Mkdir(dir, 0o777); err != nil {
			t.Fatal(err)
		}
		source := filepath.Join(dir, "snippet.c")
		output := filepath.Join(dir, "snippet.bc")
		if err := os.WriteFile(source, []byte("const char *sourceName = __FILE__;\n"), 0o666); err != nil {
			t.Fatal(err)
		}
		flags := cgoHeaderCompileArgs(source, output, []string{
			"-gdwarf-4",
			"--target=x86_64-unknown-linux-gnu",
		})
		if err := runCCompiler(flags...); err != nil {
			t.Fatal(err)
		}

		ctx := llvm.NewContext()
		mod := ctx.NewModule("package")
		headerMod, err := ctx.ParseBitcodeFile(output)
		if err != nil {
			t.Fatal(err)
		}
		if err := llvm.LinkModules(mod, headerMod); err != nil {
			t.Fatal(err)
		}
		buf := llvm.WriteBitcodeToMemoryBuffer(mod)
		outputs = append(outputs, bytes.Clone(buf.Bytes()))
		buf.Dispose()
		mod.Dispose()
		ctx.Dispose()
	}
	if !bytes.Equal(outputs[0], outputs[1]) {
		t.Fatal("CGo header bitcode depends on its temporary source path")
	}
}
