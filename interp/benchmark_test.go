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

// BenchmarkRecoverableErrors runs package initializers that each fail with a
// recoverable error, in a module with many functions and debug info.
func BenchmarkRecoverableErrors(b *testing.B) {
	for _, size := range []struct{ inits, funcs int }{{50, 2000}, {200, 2000}, {200, 8000}} {
		b.Run(fmt.Sprintf("inits=%d/funcs=%d", size.inits, size.funcs), func(b *testing.B) {
			path := filepath.Join(b.TempDir(), "recoverable.ll")
			err := os.WriteFile(path, []byte(recoverableErrorsIR(size.inits, size.funcs)), 0o666)
			if err != nil {
				b.Fatal(err)
			}
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				ctx := llvm.NewContext()
				buf, err := llvm.NewMemoryBufferFromFile(path)
				if err != nil {
					b.Fatal(err)
				}
				mod, err := ctx.ParseIR(buf)
				if err != nil {
					b.Fatal(err)
				}
				b.StartTimer()
				err = Run(mod, time.Minute, 10, false)
				b.StopTimer()
				if err != nil {
					b.Fatal(err)
				}
				mod.Dispose()
				ctx.Dispose()
			}
		})
	}
}

// recoverableErrorsIR returns a module whose package initializers each call a
// loop that runs longer than the loop limit of 10 passed to Run.
func recoverableErrorsIR(inits, funcs int) string {
	var s strings.Builder
	s.WriteString(`target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

!llvm.dbg.cu = !{!0}
!llvm.module.flags = !{!1, !2}
!0 = distinct !DICompileUnit(language: DW_LANG_C99, file: !3, producer: "interp benchmark", isOptimized: true, runtimeVersion: 0, emissionKind: FullDebug)
!1 = !{i32 2, !"Debug Info Version", i32 3}
!2 = !{i32 7, !"Dwarf Version", i32 4}
!3 = !DIFile(filename: "recoverable.go", directory: "/")
!4 = !DISubroutineType(types: !{})
`)
	md := 5
	sub := func(name string, line int) (sp, loc int) {
		sp, loc = md, md+1
		md += 2
		fmt.Fprintf(&s, "!%d = distinct !DISubprogram(name: %q, scope: !3, file: !3, line: %d, type: !4, scopeLine: %d, spFlags: DISPFlagDefinition, unit: !0)\n", sp, name, line, line)
		fmt.Fprintf(&s, "!%d = !DILocation(line: %d, column: 1, scope: !%d)\n", loc, line, sp)
		return sp, loc
	}

	sp, loc := sub("fill", 1)
	fmt.Fprintf(&s, `
define internal i64 @fill(i64 %%n) !dbg !%d {
entry:
  br label %%loop, !dbg !%d
loop:
  %%i = phi i64 [ 0, %%entry ], [ %%next, %%loop ], !dbg !%d
  %%next = add i64 %%i, 1, !dbg !%d
  %%done = icmp eq i64 %%next, %%n, !dbg !%d
  br i1 %%done, label %%exit, label %%loop, !dbg !%d
exit:
  ret i64 %%next, !dbg !%d
}
`, sp, loc, loc, loc, loc, loc, loc)

	s.WriteString("\ndefine void @runtime.initAll() unnamed_addr {\nentry:\n")
	for i := 0; i < inits; i++ {
		fmt.Fprintf(&s, "  call void @pkg%d.init(ptr undef)\n", i)
	}
	s.WriteString("  ret void\n}\n")
	for i := 0; i < inits; i++ {
		sp, loc := sub(fmt.Sprintf("pkg%d.init", i), 10+i)
		fmt.Fprintf(&s, `
@pkg%d.v = global i64 0

define internal void @pkg%d.init(ptr %%context) unnamed_addr !dbg !%d {
entry:
  %%v = call i64 @fill(i64 100), !dbg !%d
  store i64 %%v, ptr @pkg%d.v, align 8, !dbg !%d
  ret void, !dbg !%d
}
`, i, i, sp, loc, i, loc, loc)
	}
	for i := 0; i < funcs; i++ {
		sp, loc := sub(fmt.Sprintf("filler%d", i), 10000+i)
		fmt.Fprintf(&s, `
define i64 @filler%d(i64 %%x) !dbg !%d {
entry:
  %%a = add i64 %%x, %d, !dbg !%d
  %%b = mul i64 %%a, 3, !dbg !%d
  ret i64 %%b, !dbg !%d
}
`, i, sp, i, loc, loc, loc)
	}
	return s.String()
}
