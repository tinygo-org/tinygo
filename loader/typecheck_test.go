package loader

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"runtime"
	"strings"
	"testing"

	"github.com/tinygo-org/tinygo/compileopts"
	"github.com/tinygo-org/tinygo/goenv"
)

func typecheckProgram(t testing.TB, sources ...string) *Program {
	t.Helper()
	p := &Program{
		analysis:             true,
		fset:                 token.NewFileSet(),
		Packages:             make(map[string]*Package),
		typecheckConcurrency: make(chan struct{}, runtime.GOMAXPROCS(0)),
	}
	for i, source := range sources {
		file, err := parser.ParseFile(p.fset, fmt.Sprintf("pkg%d.go", i), source, parser.ParseComments)
		if err != nil {
			t.Fatal(err)
		}
		pkg := &Package{
			PackageJSON: PackageJSON{ImportPath: file.Name.Name, Name: file.Name.Name},
			program:     p,
			Files:       []*ast.File{file},
			info: types.Info{
				Types:     make(map[ast.Expr]types.TypeAndValue),
				Instances: make(map[*ast.Ident]types.Instance),
				Defs:      make(map[*ast.Ident]types.Object),
				Uses:      make(map[*ast.Ident]types.Object),
			},
			EmbedGlobals: make(map[string][]*EmbedFile),
		}
		pkg.Module.GoVersion = "1.25"
		p.sorted = append(p.sorted, pkg)
		p.Packages[pkg.ImportPath] = pkg
	}
	return p
}

func TestParallelTypecheck(t *testing.T) {
	for _, procs := range []int{1, 2, 8} {
		t.Run(fmt.Sprint(procs), func(t *testing.T) {
			old := runtime.GOMAXPROCS(procs)
			defer runtime.GOMAXPROCS(old)
			p := typecheckProgram(t,
				`package base
				type Number interface { ~int | ~int64 }
				type Box[T Number] struct { Value T }
				func (b Box[T]) Get() T { return b.Value }
				type Alias[T Number] = Box[T]`,
				`package left
				import "base"
				var X = base.Alias[int]{Value: 1}.Get()`,
				`package right
				import b "base"
				var X = b.Box[int64]{Value: 2}.Get()`,
				`package root
				import ("left"; "right"; "unsafe")
				var X = left.X + int(right.X)
				const Size = unsafe.Sizeof(X)`,
			)
			if err := p.Parse(); err != nil {
				t.Fatal(err)
			}
			for _, pkg := range p.sorted {
				if pkg.Pkg == nil || !pkg.Pkg.Complete() {
					t.Fatalf("%s is not fully checked", pkg.ImportPath)
				}
			}
			if err := p.Parse(); err != nil {
				t.Fatalf("second Parse: %v", err)
			}
		})
	}
}

func TestParallelTypecheckErrors(t *testing.T) {
	for _, source := range []string{
		`package first; var X int = "wrong"`,
		`package first
		//go:embed missing
		var X string`,
		`package first; import "missing"`,
	} {
		t.Run(source, func(t *testing.T) {
			old := runtime.GOMAXPROCS(8)
			defer runtime.GOMAXPROCS(old)
			for range 10 {
				p := typecheckProgram(t, source,
					`package second; var X bool = 1`,
					`package dependent; import "first"; var X = first.X`,
				)
				err := p.Parse()
				var errs Errors
				if !errors.As(err, &errs) || errs.Pkg != p.sorted[0] {
					t.Fatalf("Parse error = %v, want errors from first", err)
				}
				if p.sorted[2].Pkg != nil {
					t.Fatal("checked a package with a failed dependency")
				}
			}
		})
	}
}

func TestParallelTypecheckImportMap(t *testing.T) {
	old := runtime.GOMAXPROCS(8)
	defer runtime.GOMAXPROCS(old)
	for _, analysis := range []bool{false, true} {
		t.Run(fmt.Sprint(analysis), func(t *testing.T) {
			p := typecheckProgram(t,
				`package base; const X = 1`,
				`package variant; const X = 2`,
				`package main; import base "base"; const X = base.X`,
			)
			p.analysis = analysis
			variant := p.sorted[1]
			delete(p.Packages, variant.ImportPath)
			variant.ImportPath = "base [base.test]"
			p.Packages[variant.ImportPath] = variant
			p.sorted[2].ImportMap = map[string]string{"base": variant.ImportPath}
			if err := p.Parse(); err != nil {
				t.Fatal(err)
			}
			want := "1"
			if analysis {
				want = "2"
			}
			got := p.MainPkg().Pkg.Scope().Lookup("X").(*types.Const).Val().String()
			if got != want {
				t.Fatalf("X = %s, want %s", got, want)
			}
		})
	}
}

func TestParallelTypecheckSemaphore(t *testing.T) {
	for _, limit := range []int{0, 1, 4} {
		for _, invalid := range []bool{false, true} {
			t.Run(fmt.Sprintf("limit=%d/invalid=%t", limit, invalid), func(t *testing.T) {
				source := "package second; var X = 1"
				if invalid {
					source = `package second; var X int = "wrong"`
				}
				p := typecheckProgram(t, "package first", source)
				semaphore := make(chan struct{}, limit)
				if limit > 1 {
					semaphore <- struct{}{}
				}
				p.typecheckConcurrency = semaphore
				err := p.Parse()
				if limit == 0 {
					if err == nil || err.Error() != "cannot run 0 jobs at a time" {
						t.Fatalf("Parse error = %v, want zero parallelism error", err)
					}
				} else if (err != nil) != invalid {
					t.Fatalf("Parse error = %v, invalid = %t", err, invalid)
				}
				want := 0
				if limit > 1 {
					want = 1
				}
				if len(semaphore) != want {
					t.Fatalf("semaphore holds %d tokens, want %d", len(semaphore), want)
				}
			})
		}
	}
}

type countingSizes struct {
	types.Sizes
	calls   int
	offsets []int64
}

func (s *countingSizes) Sizeof(typ types.Type) int64 {
	s.calls++
	runtime.Gosched()
	return s.Sizes.Sizeof(typ)
}

func (s *countingSizes) Alignof(typ types.Type) int64 {
	s.calls++
	runtime.Gosched()
	return s.Sizes.Alignof(typ)
}

func (s *countingSizes) Offsetsof(fields []*types.Var) []int64 {
	s.calls++
	runtime.Gosched()
	s.offsets = append(s.offsets[:0], s.Sizes.Offsetsof(fields)...)
	return s.offsets
}

func TestLockedSizesOffsets(t *testing.T) {
	sizes := &lockedSizes{Sizes: &countingSizes{Sizes: types.SizesFor("gc", "amd64")}}
	field := types.NewVar(token.NoPos, nil, "x", types.Typ[types.Byte])
	offsets := sizes.Offsetsof([]*types.Var{field, field})
	field = types.NewVar(token.NoPos, nil, "x", types.Typ[types.Int])
	sizes.Offsetsof([]*types.Var{field, field})
	if len(offsets) != 2 || offsets[0] != 0 || offsets[1] != 1 {
		t.Fatalf("previous offsets changed to %v, want [0 1]", offsets)
	}
}

func TestParallelTypecheckSizes(t *testing.T) {
	old := runtime.GOMAXPROCS(8)
	defer runtime.GOMAXPROCS(old)
	var sources []string
	for i := range 20 {
		sources = append(sources, fmt.Sprintf(`package pkg%d
		import "unsafe"
		const Size = unsafe.Sizeof(int(0))
		var V struct { A byte; B int }
		const Align = unsafe.Alignof(V)
		const Offset = unsafe.Offsetof(V.B)`, i))
	}
	p := typecheckProgram(t, sources...)
	sizes := &countingSizes{Sizes: types.SizesFor("gc", "amd64")}
	p.typeChecker.Sizes = sizes
	if err := p.Parse(); err != nil {
		t.Fatal(err)
	}
	if sizes.calls < 20 {
		t.Fatalf("size callbacks called %d times, want at least 20", sizes.calls)
	}
	if p.typeChecker.Sizes != sizes {
		t.Fatal("Parse changed the typechecker configuration")
	}
}

func TestParallelTypecheckErrorOrder(t *testing.T) {
	old := runtime.GOMAXPROCS(8)
	defer runtime.GOMAXPROCS(old)
	sources := []string{
		"package base\n" + strings.Repeat("var _ = 1\n", 2000),
		`package first
		import _ "base"
		var X int = "wrong"
		var Y bool = 1`,
		`package second; var X int = "wrong"`,
	}
	baseline := typecheckProgram(t, sources...)
	if err := baseline.sorted[0].Check(); err != nil {
		t.Fatal(err)
	}
	var want Errors
	if err := baseline.sorted[1].Check(); !errors.As(err, &want) {
		t.Fatalf("sequential Check error = %v, want package errors", err)
	}
	for range 10 {
		p := typecheckProgram(t, sources...)
		var got Errors
		if err := p.Parse(); !errors.As(err, &got) || got.Pkg != p.sorted[1] {
			t.Fatalf("Parse error = %v, want errors from first", err)
		}
		if len(got.Errs) != len(want.Errs) {
			t.Fatalf("got %d errors, want %d", len(got.Errs), len(want.Errs))
		}
		for i, err := range got.Errs {
			if err.Error() != want.Errs[i].Error() {
				t.Errorf("error %d = %v, want %v", i, err, want.Errs[i])
			}
		}
	}
}

func BenchmarkTypecheck(b *testing.B) {
	_, minor, err := goenv.GetGorootVersion()
	if err != nil {
		b.Fatal(err)
	}
	config := &compileopts.Config{
		Options: &compileopts.Options{},
		Target: &compileopts.TargetSpec{
			Triple: "x86_64-unknown-linux", GOOS: "linux", GOARCH: "amd64",
			BuildTags: []string{"linux", "amd64"},
		},
		GoMinorVersion: minor,
	}
	for _, input := range []string{"fmt", "encoding/json"} {
		b.Run(strings.ReplaceAll(input, "/", "_"), func(b *testing.B) {
			for _, procs := range []int{1, 4, 8} {
				for _, parallel := range []bool{false, true} {
					b.Run(fmt.Sprintf("procs=%d/parallel=%t", procs, parallel), func(b *testing.B) {
						old := runtime.GOMAXPROCS(procs)
						defer runtime.GOMAXPROCS(old)
						b.ReportAllocs()
						for b.Loop() {
							b.StopTimer()
							p, err := load(config, []string{input}, types.Config{
								Sizes: types.SizesFor("gc", "amd64"),
							}, true, nil)
							if err != nil {
								b.Fatal(err)
							}
							for _, pkg := range p.sorted {
								if err := pkg.Parse(); err != nil {
									b.Fatal(err)
								}
							}
							b.StartTimer()
							if parallel {
								if err := p.check(); err != nil {
									b.Fatal(err)
								}
							} else {
								for _, pkg := range p.sorted {
									if err := pkg.Check(); err != nil {
										b.Fatal(err)
									}
								}
							}
						}
					})
				}
			}
		})
	}
}
