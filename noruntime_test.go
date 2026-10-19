package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tinygo-org/tinygo/builder"
	"github.com/tinygo-org/tinygo/interp"
)

func TestNoRuntime(t *testing.T) {
	for _, tc := range []struct {
		name   string
		source string
		fail   bool
	}{
		{"evaluated", `var result = calc(12)
//go:noruntime
func calc(n int) int { return n * 7 }
func main() { println(result) }`, false},
		{"loop-limit", `var result = calc(1001)
//go:noruntime
func calc(n int) int { sum := 0; for i := 0; i < n; i++ { sum += i }; return sum }
func main() { println(result) }`, true},
		{"loop-below-limit", `var result = calc(999)
//go:noruntime
func calc(n int) int { sum := 0; for i := 0; i < n; i++ { sum += i }; return sum }
func main() { println(result) }`, false},
		{"runtime", `//go:noruntime
func calc(n int) int { return n * 7 }
func main() { println(calc(12)) }`, true},
		{"inline-hint", `//go:inline
//go:noruntime
func calc(n int) int { return n * 7 }
func main() { println(calc(12)) }`, true},
		{"linkname-before-definition", `import _ "unsafe"
//go:linkname alias main.calc
func alias(n int) int
//go:noruntime
func calc(n int) int { return n * 7 }
func main() { println(alias(12)) }`, true},
		{"linkname-after-definition", `import _ "unsafe"
//go:noruntime
func calc(n int) int { return n * 7 }
//go:linkname alias main.calc
func alias(n int) int
func main() { println(alias(12)) }`, true},
		{"evaluated-linkname", `import _ "unsafe"
//go:linkname alias main.calc
func alias(n int) int
var result = alias(12)
//go:noruntime
func calc(n int) int { return n * 7 }
func main() { println(result) }`, false},
		{"unused-linkname", `import _ "unsafe"
//go:linkname alias main.calc
func alias(n int) int
//go:noruntime
func calc(n int) int { return n * 7 }
func main() {}`, false},
		{"unused", `//go:noruntime
func calc(n int) int { return n * 7 }
func main() {}`, false},
		{"dead-caller", `//go:noruntime
func calc(n int) int { return n * 7 }
func dead() int { return calc(12) }
func main() {}`, false},
		{"function-value", `var callback = calc
//go:noruntime
func calc(n int) int { return n * 7 }
func main() { println(callback(12)) }`, true},
		{"evaluated-function-value", `var result = func() int { f := calc; return f(12) }()
//go:noruntime
func calc(n int) int { return n * 7 }
func main() { println(result) }`, false},
		{"unused-function-value", `var callback = calc
//go:noruntime
func calc(n int) int { return n * 7 }
func main() {}`, false},
		{"unsupported", `import "runtime/volatile"
var reg volatile.Register32
var result = calc(0)
//go:noruntime
func calc(n int) uint32 { return reg.Get() }
func main() { println(result) }`, true},
		{"dead-aggregate", `type large [100]int
var result = calc(large{12})
//go:noruntime
func calc(n large) int { return n[0] * 7 }
func main() { println(result) }`, false},
		{"generic", `var result = calc(12)
//go:noruntime
func calc[T ~int](n T) T { return n * 7 }
func main() { println(result) }`, false},
		{"generic-runtime", `//go:noruntime
func calc[T ~int](n T) T { return n * 7 }
func main() { println(calc(12)) }`, true},
		{"method-value", `type number int
var callback = number(12).calc
//go:noruntime
func (n number) calc() int { return int(n) * 7 }
func main() { println(callback()) }`, true},
		{"interface-call", `type number int
type calculator interface { calc() int }
var callback calculator = number(12)
//go:noruntime
func (n number) calc() int { return int(n) * 7 }
func main() { println(callback.calc()) }`, true},
		{"exported", `//go:export calc
//go:noruntime
func calc(n int) int { return n * 7 }
func main() {}`, true},
		{"retained-method-table", `type number int
type observer interface { other() int }
var callback observer = number(12)
//go:noruntime
func (n number) calc() int { return int(n) * 7 }
func (n number) other() int { return int(n) }
func main() { println(callback.other()) }`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, opt := range []string{"0", "2"} {
				t.Run(opt, func(t *testing.T) {
					tmpdir := t.TempDir()
					path := filepath.Join(tmpdir, "main.go")
					if err := os.WriteFile(path, []byte("package main\n"+tc.source+"\n"), 0o600); err != nil {
						t.Fatal(err)
					}
					options := optionsFromTarget("wasip1", sema)
					options.Opt = opt
					options.Scheduler = "none"
					options.InterpMaxLoopIterations = interp.DefaultMaxInterpBlockEntries
					config, err := builder.NewConfig(&options)
					if err != nil {
						t.Fatal(err)
					}
					out := filepath.Join(tmpdir, "main.ll")
					_, err = builder.Build(path, out, tmpdir, config)
					if tc.fail {
						if err == nil || !strings.Contains(err.Error(), "//go:noruntime function ") || !strings.Contains(err.Error(), " remains reachable at runtime") {
							t.Fatalf("expected noruntime error, got %v", err)
						}
						return
					}
					if err != nil {
						t.Fatal(err)
					}
					data, err := os.ReadFile(out)
					if err != nil {
						t.Fatal(err)
					}
					if strings.Contains(string(data), "tinygo-noruntime") {
						t.Fatal("noruntime function remains in output")
					}
				})
			}
		})
	}
}
