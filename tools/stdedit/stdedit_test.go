package main

import (
	"bytes"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func newPackage(t *testing.T, upstream string, local ...string) *Package {
	t.Helper()
	p := &Package{Fset: token.NewFileSet(), Path: "p", GoVersion: "go1.26"}
	f, err := parser.ParseFile(p.Fset, "p.go", upstream, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	p.Upstream = []*ast.File{f}
	for _, src := range local {
		f, err := parser.ParseFile(p.Fset, "local.go", src, parser.ParseComments)
		if err != nil {
			t.Fatal(err)
		}
		p.Local = append(p.Local, f)
	}
	if err := p.Check(); err != nil {
		t.Fatal(err)
	}
	return p
}

func generateTest(t *testing.T, p *Package, pass Pass) string {
	t.Helper()
	if err := pass(p); err != nil {
		t.Fatal(err)
	}
	if err := p.Check(); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if _, err := p.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	_, body, _ := strings.Cut(out, "package p\n")
	return strings.TrimSpace(body)
}

func TestReplaceDecl(t *testing.T) {
	src := `package p

// A is kept.
const A = 1

const (
	B = 2
	C = 3
)

// F is replaced.
func F() int {
	// gone
	return A
}
`
	p := newPackage(t, src)
	got := generateTest(t, p, Pipeline(
		Walk(ReplaceDecl("A", "const A = 10")),
		Walk(ReplaceDecl("C", "const C = 30")),
		Walk(ReplaceDecl("F", "func F() int { return B }")),
	))
	want := `// A is kept.
const A = 10

const (
	B = 2
	C = 30
)

// F is replaced.
func F() int { return B }`
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestInsertFirst(t *testing.T) {
	p := newPackage(t, "package p\n\nfunc F(x int) int {\n\treturn x\n}\n")
	got := generateTest(t, p, Walk(InsertFirst("F", "if x < 0 { return 0 }")))
	want := "func F(x int) int {\n\tif x < 0 {\n\t\treturn 0\n\t}\n\treturn x\n}"
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestChainSeesReplacement(t *testing.T) {
	p := newPackage(t, "package p\n\nfunc F() int {\n\treturn 1\n}\n")
	got := generateTest(t, p, Walk(Chain(
		ReplaceDecl("F", "func F() int { return 2 }"),
		InsertFirst("F", "println()"),
	)))
	want := "func F() int { println(); return 2 }"
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestIn(t *testing.T) {
	p := newPackage(t, "package p\n\nfunc F() { g() }\n\nfunc G() { g() }\n\nfunc g() {}\n")
	insert := func(p *Package, c *Cursor) bool {
		b, ok := c.Node().(*ast.BlockStmt)
		if !ok {
			return false
		}
		stmts, _ := p.ParseStmts("println()", b.Lbrace)
		b.List = append(stmts, b.List...)
		return true
	}
	got := generateTest(t, p, Walk(In("G", insert)))
	want := "func F() { g() }\n\nfunc G() { println(); g() }\n\nfunc g() {}"
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestErrors(t *testing.T) {
	ran := false
	for _, tc := range []struct {
		name string
		pass Pass
		want string
	}{
		{"no change", Walk(ReplaceDecl("Missing", "const Missing = 1")), "no change in go1.26 p"},
		{"bad source", Walk(ReplaceDecl("A", "const A =")), "expected"},
		{"wrong name", Walk(ReplaceDecl("A", "const B = 1")), "replacement for A defines [B]"},
		{"pipeline stops", Pipeline(Walk(InsertFirst("Missing", "")), func(*Package) error { ran = true; return nil }), "step 1: no change"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := newPackage(t, "package p\n\nconst A = 1\n")
			err := tc.pass(p)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("got %v, want %q", err, tc.want)
			}
		})
	}
	if ran {
		t.Errorf("pipeline ran the step after a failure")
	}
}

func TestParseUpstream(t *testing.T) {
	for _, tc := range []struct {
		name  string
		files map[string]string
		want  string
	}{
		{"plain", map[string]string{"a.go": "package p\n", "a_test.go": "package p\n", "gen.go": "//go:build ignore\n\npackage main\n"}, ""},
		{"build tag", map[string]string{"a.go": "//go:build !purego\n\npackage p\n"}, "p/a.go has //go:build !purego"},
		{"os suffix", map[string]string{"a_linux.go": "package p\n"}, "p/a_linux.go only builds for some GOOS or GOARCH"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			dir := filepath.Join(root, "src", "p")
			if err := os.MkdirAll(dir, 0o755); err != nil {
				t.Fatal(err)
			}
			for name, src := range tc.files {
				if err := os.WriteFile(filepath.Join(dir, name), []byte(src), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			old := build.Default.GOROOT
			build.Default.GOROOT = root
			defer func() { build.Default.GOROOT = old }()
			files, err := parseUpstream(token.NewFileSet(), "p")
			switch {
			case tc.want == "" && err != nil:
				t.Fatal(err)
			case tc.want == "" && len(files) != 1:
				t.Errorf("got %d files, want 1", len(files))
			case tc.want != "" && (err == nil || err.Error() != tc.want):
				t.Errorf("got %v, want %q", err, tc.want)
			}
		})
	}
}
