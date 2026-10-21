package fix

import (
	"bytes"
	"flag"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/tinygo-org/tinygo/builder"
	"github.com/tinygo-org/tinygo/compileopts"
	"github.com/tinygo-org/tinygo/compiler"
	"github.com/tinygo-org/tinygo/goenv"
	"github.com/tinygo-org/tinygo/internal/diff"
	"github.com/tinygo-org/tinygo/loader"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/checker"
	"golang.org/x/tools/go/analysis/passes/inline"
	"golang.org/x/tools/go/analysis/passes/modernize"
	"golang.org/x/tools/go/packages"
)

func fixture(t *testing.T, files map[string]string, patterns ...string) (Load, string) {
	return fixtureWithTests(t, files, true, patterns...)
}

func fixtureWithTests(t *testing.T, files map[string]string, includeTests bool, patterns ...string) (Load, string) {
	t.Helper()
	t.Setenv("GOFLAGS", "")
	t.Setenv("GOWORK", "off")
	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("TINYGOROOT", root)
	dir := t.TempDir()
	dir, err = filepath.EvalSymlinks(dir)
	if err != nil {
		t.Fatal(err)
	}
	files["go.mod"] = "module example.com/probe\n\ngo 1.25.0\n"
	for name, contents := range files {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(contents), 0600); err != nil {
			t.Fatal(err)
		}
	}
	t.Chdir(dir)
	config, err := builder.NewConfig(&compileopts.Options{
		GOOS: goenv.Get("GOOS"), GOARCH: goenv.Get("GOARCH"), Opt: "z",
	})
	if err != nil {
		t.Fatal(err)
	}
	machine, err := compiler.NewTargetMachine(&compiler.Config{
		Triple: config.Triple(), CPU: config.CPU(), Features: config.Features(),
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(machine.Dispose)
	sizes := compiler.Sizes(machine)
	return func(overlay map[string][]byte) (*loader.Analysis, error) {
		return loader.LoadForAnalysis(config, patterns, sizes, overlay, includeTests)
	}, dir
}

func read(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestFixPackagesAndTests(t *testing.T) {
	files := map[string]string{
		"probe.go":      "package probe\nimport _ \"example.com/probe/dep\"\nvar X interface{}\n",
		"probe_test.go": "package probe\nvar TestX interface{}\n",
		"external_test.go": "package probe_test\nimport \"example.com/probe\"\n" +
			"var External interface{} = probe.X\n",
		"dep/dep.go": "package dep\nvar Dependency interface{}\n",
	}
	load, dir := fixture(t, files)
	before := read(t, filepath.Join(dir, "probe.go"))
	var output, diagnostics bytes.Buffer
	changed, err := Run(load, []*analysis.Analyzer{modernize.AnyAnalyzer}, true, &output, &diagnostics)
	if err != nil {
		t.Fatal(err)
	}
	if !changed || !strings.Contains(output.String(), "+var X any") ||
		!strings.Contains(output.String(), "+var TestX any") ||
		!strings.Contains(output.String(), "+var External any") {
		t.Fatalf("missing fixes in diff:\n%s\n%s", &output, &diagnostics)
	}
	if !bytes.Equal(before, read(t, filepath.Join(dir, "probe.go"))) {
		t.Fatal("diff wrote source")
	}
	changed, err = Run(load, []*analysis.Analyzer{modernize.AnyAnalyzer}, false, &output, &diagnostics)
	if err != nil || !changed {
		t.Fatalf("fix: changed=%v, err=%v", changed, err)
	}
	if got := string(read(t, filepath.Join(dir, "dep/dep.go"))); got != files["dep/dep.go"] {
		t.Fatalf("dependency changed: %s", got)
	}
	changed, err = Run(load, []*analysis.Analyzer{modernize.AnyAnalyzer}, false, &output, &diagnostics)
	if err != nil || changed {
		t.Fatalf("second fix: changed=%v, err=%v", changed, err)
	}
}

func TestPatternsAndTypeIdentity(t *testing.T) {
	load, _ := fixture(t, map[string]string{
		"a/a.go":      "package a\ntype Value struct{}\n",
		"b/b.go":      "package b\nimport \"example.com/probe/a\"\nvar V a.Value\n",
		"b/b_test.go": "package b\nvar Internal = V\n",
	}, "./a", "./b")
	program, err := load(nil)
	if err != nil {
		t.Fatal(err)
	}
	seen := make(map[string]bool)
	for _, pkg := range program.Packages {
		seen[pkg.PkgPath] = true
		if pkg.PkgPath == "example.com/probe/b" {
			a := pkg.Imports["example.com/probe/a"]
			if a == nil || pkg.Types.Scope().Lookup("V").Type() != a.Types.Scope().Lookup("Value").Type() {
				t.Fatal("imports do not share type identities")
			}
		}
		if pkg.TypesSizes == nil || pkg.TypesInfo == nil || pkg.Fset == nil {
			t.Fatal("incomplete analysis inputs")
		}
	}
	if !seen["example.com/probe/a"] || !seen["example.com/probe/b"] {
		t.Fatalf("missing roots: %v", seen)
	}
}

type packageFact struct{ Value bool }

func (*packageFact) AFact() {}

func TestDependencyFacts(t *testing.T) {
	load, _ := fixture(t, map[string]string{
		"probe.go":   "package probe\nimport _ \"example.com/probe/dep\"\n",
		"dep/dep.go": "package dep\n",
	})
	found := false
	analyzer := &analysis.Analyzer{
		Name: "facts", Doc: "check dependency facts", FactTypes: []analysis.Fact{new(packageFact)},
		Run: func(pass *analysis.Pass) (any, error) {
			for _, imported := range pass.Pkg.Imports() {
				var fact packageFact
				if imported.Path() == "example.com/probe/dep" && pass.ImportPackageFact(imported, &fact) {
					found = fact.Value
				}
			}
			pass.ExportPackageFact(&packageFact{true})
			return nil, nil
		},
	}
	_, err := Run(load, []*analysis.Analyzer{analyzer}, true, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil || !found {
		t.Fatalf("dependency facts: found=%v, err=%v", found, err)
	}
}

func TestLanguageVersion(t *testing.T) {
	load, _ := fixture(t, map[string]string{
		"probe.go": "package probe\nfunc Pointer() *int { v := 1; return &v }\n",
	})
	changed, err := Run(load, []*analysis.Analyzer{modernize.NewExprAnalyzer}, true, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil || changed {
		t.Fatalf("Go 1.25 source received new(expr): changed=%v, err=%v", changed, err)
	}
}

func TestNewImport(t *testing.T) {
	load, dir := fixture(t, map[string]string{
		"probe.go": "package probe\nimport \"sort\"\nfunc Sort(xs []int) { sort.Slice(xs, func(i, j int) bool { return xs[i] < xs[j] }) }\n",
	})
	changed, err := Run(load, []*analysis.Analyzer{modernize.SlicesSortAnalyzer}, false, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil || !changed {
		t.Fatalf("new import: changed=%v, err=%v", changed, err)
	}
	if got := read(t, filepath.Join(dir, "probe.go")); !bytes.Contains(got, []byte("slices.Sort(xs)")) {
		t.Fatalf("sort fix not written:\n%s", got)
	}
}

func TestInlineAcrossPackages(t *testing.T) {
	load, dir := fixture(t, map[string]string{
		"probe.go":   "package probe\nimport \"example.com/probe/dep\"\nvar X = dep.Old(2)\n",
		"dep/dep.go": "package dep\n//go:fix inline\nfunc Old(x int) int { return x + 1 }\n",
	})
	changed, err := Run(load, []*analysis.Analyzer{inline.Analyzer}, false, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil || !changed {
		t.Fatalf("inline: changed=%v, err=%v", changed, err)
	}
	if got := read(t, filepath.Join(dir, "probe.go")); bytes.Contains(got, []byte("dep.Old")) {
		t.Fatalf("call was not inlined:\n%s", got)
	}
}

func TestTinyGoSourceWrite(t *testing.T) {
	load, _ := fixture(t, map[string]string{"probe.go": "package probe\n"}, "internal/itoa")
	originalRoot := goenv.Get("TINYGOROOT")
	root := t.TempDir()
	root, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.CopyFS(filepath.Join(root, "src"), os.DirFS(filepath.Join(originalRoot, "src"))); err != nil {
		t.Fatal(err)
	}
	name := filepath.Join(root, "src", "internal", "itoa", "modernize_probe.go")
	source := []byte("package itoa\nvar modernizeProbe interface{}\n")
	if err := os.WriteFile(name, source, 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("TINYGOROOT", root)
	program, err := load(nil)
	if err != nil {
		t.Fatal(err)
	}
	if !program.Writable[name] {
		t.Fatalf("original source path is not editable: %s", name)
	}
	for _, pkg := range program.Packages {
		for _, version := range pkg.TypesInfo.FileVersions {
			want := "go1." + strconv.Itoa(compileopts.MinimumGoMinorVersion)
			if version != want {
				t.Fatalf("library language version = %s, want %s", version, want)
			}
		}
	}
	changed, err := Run(load, []*analysis.Analyzer{modernize.AnyAnalyzer}, false, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil || !changed {
		t.Fatalf("library write: changed=%v, err=%v", changed, err)
	}
	if !bytes.Contains(read(t, name), []byte("var modernizeProbe any")) {
		t.Fatal("fix did not reach the original TinyGo file")
	}
}

func TestCGoSourceSafety(t *testing.T) {
	for _, cgoEdit := range []bool{false, true} {
		t.Run(map[bool]string{false: "Go file", true: "CGo file"}[cgoEdit], func(t *testing.T) {
			cgoSource := "package probe\n// typedef int value;\nimport \"C\"\nvar Value C.value\n"
			goSource := "package probe\nvar X interface{}\n"
			if cgoEdit {
				cgoSource += "var Y interface{}\n"
			}
			load, dir := fixture(t, map[string]string{"cgo.go": cgoSource, "probe.go": goSource})
			var diagnostics bytes.Buffer
			changed, err := Run(load, []*analysis.Analyzer{modernize.AnyAnalyzer}, false, &bytes.Buffer{}, &diagnostics)
			if err != nil || !changed {
				t.Fatalf("unchanged Go AST: changed=%v, err=%v", changed, err)
			}
			if !strings.Contains(diagnostics.String(), "skipping source rewritten by CGo") {
				t.Fatalf("missing CGo notice: %s", &diagnostics)
			}
			if got := string(read(t, filepath.Join(dir, "cgo.go"))); got != cgoSource {
				t.Fatal("rewritten CGo AST was written to source")
			}
		})
	}
}

func TestAnalyzerSelection(t *testing.T) {
	for _, test := range []struct {
		args []string
		want func(string) bool
	}{
		{nil, func(string) bool { return true }},
		{[]string{"-any"}, func(s string) bool { return s == "any" }},
		{[]string{"-any=false"}, func(s string) bool { return s != "any" }},
		{[]string{"-any", "-rangeint"}, func(s string) bool { return s == "any" || s == "rangeint" }},
	} {
		set := flag.NewFlagSet("test", flag.ContinueOnError)
		flags := RegisterFlags(set)
		if err := set.Parse(test.args); err != nil {
			t.Fatal(err)
		}
		selected := make(map[string]bool)
		for _, a := range flags.Selected() {
			selected[a.Name] = true
		}
		for _, a := range Analyzers() {
			if selected[a.Name] != test.want(a.Name) {
				t.Fatalf("%v: unexpected selection for %s", test.args, a.Name)
			}
		}
	}
}

func TestEditSafety(t *testing.T) {
	for _, test := range []struct {
		name, source, want string
		mutate             bool
	}{
		{"generated", "// Code generated by fixture. DO NOT EDIT.\npackage probe\nvar X interface{}\n", "cannot safely fix", false},
		{"stale", "package probe\nvar X interface{}\n", "changed during analysis", true},
		{"invalid", "package probe\nvar X interface{}\n", "invalid fixed source", false},
		{"conflict", "package probe\nvar X interface{}\n", "conflicting fixes", false},
		{"type error", "package probe\nvar X interface{}\n", "do not type-check", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			load, dir := fixture(t, map[string]string{"probe.go": test.source})
			path := filepath.Join(dir, "probe.go")
			analyzer := &analysis.Analyzer{
				Name: "safety", Doc: "test unsafe edits",
				Run: func(pass *analysis.Pass) (any, error) {
					if test.mutate {
						if err := os.WriteFile(path, []byte(test.source+"\n"), 0600); err != nil {
							return nil, err
						}
					}
					file := pass.Files[0]
					replacement := "package probe\nvar Y any\n"
					if test.name == "invalid" {
						replacement = "invalid {"
					} else if test.name == "type error" {
						replacement = "package probe\nvar X missing\n"
					}
					edits := []analysis.TextEdit{{Pos: file.FileStart, End: file.FileEnd, NewText: []byte(replacement)}}
					pass.Report(analysis.Diagnostic{Pos: file.Pos(), SuggestedFixes: []analysis.SuggestedFix{{TextEdits: edits}}})
					if test.name == "conflict" {
						pass.Report(analysis.Diagnostic{Pos: file.Pos(), SuggestedFixes: []analysis.SuggestedFix{{
							TextEdits: []analysis.TextEdit{{Pos: file.FileStart, End: file.FileEnd, NewText: []byte("package probe\nvar Z any\n")}},
						}}})
					}
					return nil, nil
				},
			}
			_, err := Run(load, []*analysis.Analyzer{analyzer}, false, &bytes.Buffer{}, &bytes.Buffer{})
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("got %v, want %s", err, test.want)
			}
			if !test.mutate && string(read(t, path)) != test.source {
				t.Fatal("failed fix wrote source")
			}
		})
	}
}

func TestAnalyzerError(t *testing.T) {
	load, _ := fixture(t, map[string]string{"probe.go": "package probe\n"})
	analyzer := &analysis.Analyzer{
		Name: "failure", Doc: "fail",
		Run: func(*analysis.Pass) (any, error) { return nil, os.ErrPermission },
	}
	_, err := Run(load, []*analysis.Analyzer{analyzer}, false, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "permission denied") {
		t.Fatalf("analyzer error was lost: %v", err)
	}
}

func TestUnified(t *testing.T) {
	if _, err := exec.LookPath("patch"); err != nil {
		t.Skip("patch is not available")
	}
	for _, test := range []struct{ before, after string }{
		{"", "new\n"}, {"old\n", ""}, {"old", "new"},
		{"a\nb\nc\nd\ne\nf\ng\nh\n", "a\nb\nC\nd\ne\nf\nG\nh\n"},
	} {
		dir := t.TempDir()
		name := filepath.Join(dir, "source.go")
		if err := os.WriteFile(name, []byte(test.before), 0600); err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command("patch", "--quiet", name)
		cmd.Stdin = strings.NewReader(string(diff.Diff(name, []byte(test.before), name, []byte(test.after))))
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("patch: %v\n%s", err, output)
		}
		if got := string(read(t, name)); got != test.after {
			t.Fatalf("patch produced %q, want %q", got, test.after)
		}
	}
}

func TestPrepareRejectsInvalidOffsets(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.go")
	fset := token.NewFileSet()
	file := fset.AddFile(path, -1, 100)
	graph := &checker.Graph{Roots: []*checker.Action{{
		Analyzer: &analysis.Analyzer{Name: "offset"},
		Package:  &packages.Package{ID: "probe", Fset: fset},
		Diagnostics: []analysis.Diagnostic{{SuggestedFixes: []analysis.SuggestedFix{{
			TextEdits: []analysis.TextEdit{{Pos: file.Pos(80), End: file.Pos(90)}},
		}}}},
	}}}
	_, err := prepare(&loader.Analysis{
		Sources:  map[string][]byte{path: []byte("package probe\n")},
		Writable: map[string]bool{path: true},
	}, graph, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "invalid offset") {
		t.Fatalf("got %v, want invalid offset", err)
	}

}

func TestWriteFailure(t *testing.T) {
	dir := t.TempDir()
	name := filepath.Join(dir, "readonly.go")
	if err := os.WriteFile(name, []byte("before"), 0400); err != nil {
		t.Fatal(err)
	}
	if err := write(change{name: name, after: []byte("after"), mode: 0400}); err == nil {
		t.Fatal("read-only file was overwritten")
	}
	if got := string(read(t, name)); got != "before" {
		t.Fatalf("failed write changed source to %q", got)
	}
}
