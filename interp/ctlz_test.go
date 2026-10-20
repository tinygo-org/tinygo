package interp

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"tinygo.org/x/go-llvm"
)

func TestInterpCTLZ(t *testing.T) {
	for _, width := range []int{8, 16, 32, 64} {
		highest := fmt.Sprint(uint64(1) << (width - 1))
		interior := fmt.Sprint(uint64(1) << (width / 2))
		for _, test := range []struct {
			name       string
			input      string
			zeroPoison bool
			want       uint64
			runtime    bool
		}{
			{"zero", "0", false, uint64(width), false},
			{"zero", "0", true, 0, true},
			{"one", "1", false, uint64(width - 1), false},
			{"one", "1", true, uint64(width - 1), false},
			{"two", "2", false, uint64(width - 2), false},
			{"two", "2", true, uint64(width - 2), false},
			{"highest", highest, false, 0, false},
			{"highest", highest, true, 0, false},
			{"all", "-1", false, 0, false},
			{"all", "-1", true, 0, false},
			{"interior", interior, false, uint64(width/2 - 1), false},
			{"interior", interior, true, uint64(width/2 - 1), false},
			{"unknown", "%input", false, 0, true},
			{"unknown", "%input", true, 0, true},
		} {
			for _, entrypoint := range []string{"runtime.initAll", "main.init"} {
				t.Run(fmt.Sprintf("%s/i%d/%s/%t", entrypoint, width, test.name, test.zeroPoison), func(t *testing.T) {
					t.Parallel()
					runCTLZTest(t, entrypoint, width, test.input, test.zeroPoison, test.want, test.runtime)
				})
			}
		}
	}
	for _, entrypoint := range []string{"runtime.initAll", "main.init"} {
		t.Run(entrypoint+"/i1", func(t *testing.T) {
			t.Parallel()
			runCTLZTest(t, entrypoint, 1, "1", false, 0, true)
		})
	}
}

func runCTLZTest(t *testing.T, entrypoint string, width int, input string, zeroPoison bool, want uint64, wantRuntime bool) {
	t.Helper()
	inputCall := ""
	if input == "%input" {
		inputCall = fmt.Sprintf("%%input = call i%d @unknown()", width)
	}
	ir := fmt.Sprintf(`
target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"
@result = global i%[1]d -1
define void @runtime.initAll() {
	call void @main.init()
	ret void
}
define internal void @main.init() {
	%[2]s
	%%count = call i%[1]d @llvm.ctlz.i%[1]d(i%[1]d %[3]s, i1 %[4]t)
	store i%[1]d %%count, ptr @result
	ret void
}
declare i%[1]d @llvm.ctlz.i%[1]d(i%[1]d, i1 immarg)
declare i%[1]d @unknown()
`, width, inputCall, input, zeroPoison)

	ctx := llvm.NewContext()
	defer ctx.Dispose()
	path := filepath.Join(t.TempDir(), "ctlz.ll")
	if err := os.WriteFile(path, []byte(ir), 0o600); err != nil {
		t.Fatal(err)
	}
	buf, err := llvm.NewMemoryBufferFromFile(path)
	if err != nil {
		t.Fatal(err)
	}
	mod, err := ctx.ParseIR(buf)
	if err != nil {
		t.Fatal(err)
	}
	defer mod.Dispose()
	if err := llvm.VerifyModule(mod, llvm.ReturnStatusAction); err != nil {
		t.Fatal(err)
	}
	if entrypoint == "runtime.initAll" {
		err = Run(mod, time.Minute, DefaultMaxInterpBlockEntries, false)
	} else {
		err = RunFunc(mod.NamedFunction(entrypoint), time.Minute, DefaultMaxInterpBlockEntries, false)
	}
	if err != nil {
		t.Fatal(err)
	}
	if err := llvm.VerifyModule(mod, llvm.ReturnStatusAction); err != nil {
		t.Fatal(err)
	}

	fn := mod.NamedFunction(entrypoint)
	body := fn.String()
	if wantRuntime {
		if !strings.Contains(body, fmt.Sprintf("@llvm.ctlz.i%d(", width)) {
			t.Fatalf("expected ctlz at runtime:\n%s", body)
		}
		if !strings.Contains(body, fmt.Sprintf("i1 %t)", zeroPoison)) {
			t.Fatalf("expected unchanged zero poison flag:\n%s", body)
		}
		if !strings.Contains(body, "store ") {
			t.Fatalf("expected result stored at runtime:\n%s", body)
		}
		return
	}
	if got := mod.NamedGlobal("result").Initializer().ZExtValue(); got != want {
		t.Fatalf("got %d, want %d:\n%s", got, want, mod.String())
	}
	if inst := fn.EntryBasicBlock().FirstInstruction(); inst.IsAReturnInst().IsNil() {
		t.Fatalf("expected initializer to contain only a return:\n%s", body)
	}
}
