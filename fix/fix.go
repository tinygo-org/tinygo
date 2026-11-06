// Package fix runs Go fix analyzers on packages loaded by TinyGo.
package fix

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	internaldiff "github.com/tinygo-org/tinygo/internal/diff"
	"github.com/tinygo-org/tinygo/loader"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/checker"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/hostport"
	"golang.org/x/tools/go/analysis/passes/inline"
	"golang.org/x/tools/go/analysis/passes/modernize"
	"golang.org/x/tools/go/ast/astutil"
)

// Analyzers returns the fixers supported by the pinned x/tools version.
func Analyzers() []*analysis.Analyzer {
	all := append([]*analysis.Analyzer{}, modernize.Suite...)
	all = append(all, buildtag.Analyzer, hostport.Analyzer, inline.Analyzer)
	slices.SortFunc(all, func(a, b *analysis.Analyzer) int { return strings.Compare(a.Name, b.Name) })
	return all
}

// Flags registers analyzer switches and options on a command's flag set.
type Flags struct {
	set     *flag.FlagSet
	enabled map[*analysis.Analyzer]*bool
}

func RegisterFlags(set *flag.FlagSet) *Flags {
	flags := &Flags{set: set, enabled: make(map[*analysis.Analyzer]*bool)}
	for _, a := range Analyzers() {
		flags.enabled[a] = set.Bool(a.Name, false, "enable "+a.Name+" analysis")
		a.Flags.VisitAll(func(f *flag.Flag) {
			set.Var(f.Value, a.Name+"."+f.Name, f.Usage)
		})
	}
	return flags
}

// Selected follows Go's rule: a true switch selects only enabled analyzers.
// Otherwise all analyzers run except those explicitly disabled.
func (f *Flags) Selected() []*analysis.Analyzer {
	only := false
	for _, enabled := range f.enabled {
		only = only || *enabled
	}
	explicit := make(map[string]bool)
	f.set.Visit(func(flag *flag.Flag) { explicit[flag.Name] = true })
	var selected []*analysis.Analyzer
	for _, a := range Analyzers() {
		enabled := *f.enabled[a]
		if enabled || !only && !explicit[a.Name] {
			selected = append(selected, a)
		}
	}
	return selected
}

// Load reloads the same roots with optional source replacements.
type Load func(overlay map[string][]byte) (*loader.Analysis, error)

// Run prepares and type-checks all fixes before it writes files.
// It returns whether fixes were available, including in diff mode.
func Run(load Load, analyzers []*analysis.Analyzer, diff bool, stdout, stderr io.Writer) (bool, error) {
	program, err := load(nil)
	if err != nil {
		return false, err
	}
	for _, name := range program.Skipped {
		if _, err := fmt.Fprintf(stderr, "%s: skipping source rewritten by CGo\n", name); err != nil {
			return false, err
		}
	}
	graph, err := checker.Analyze(analyzers, program.Packages, nil)
	if err != nil {
		return false, err
	}
	for action := range graph.All() {
		if action.Err != nil {
			return false, fmt.Errorf("%s: %w", action, action.Err)
		}
	}
	changes, err := prepare(program, graph, stderr)
	if err != nil {
		return false, err
	}
	if len(changes) == 0 {
		return false, nil
	}
	overlay := make(map[string][]byte, len(changes))
	for _, change := range changes {
		overlay[change.name] = change.after
	}
	if _, err := load(overlay); err != nil {
		return false, fmt.Errorf("fixed packages do not type-check; no files changed: %w", err)
	}
	for name, before := range program.Sources {
		resolved, err := filepath.EvalSymlinks(name)
		if err != nil {
			return false, err
		}
		if resolved != name {
			return false, fmt.Errorf("%s became a symlink during analysis; no files changed", name)
		}
		current, err := os.ReadFile(name)
		if err != nil {
			return false, err
		}
		if !bytes.Equal(current, before) {
			return false, fmt.Errorf("%s changed during analysis; no files changed", name)
		}
	}
	for _, change := range changes {
		if diff {
			if _, err := stdout.Write(internaldiff.Diff(change.name, change.before, change.name, change.after)); err != nil {
				return false, err
			}
		} else {
			if err := write(change); err != nil {
				return false, fmt.Errorf("writing %s (earlier files may have changed): %w", change.name, err)
			}
		}
	}
	return true, nil
}

func write(change change) error {
	if change.mode&0222 == 0 {
		return fmt.Errorf("%s: %w", change.name, os.ErrPermission)
	}
	file, err := os.CreateTemp(filepath.Dir(change.name), ".tinygo-fix-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	if _, err := file.Write(change.after); err != nil {
		file.Close()
		return err
	}
	if err := file.Chmod(change.mode); err != nil {
		file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(file.Name(), change.name)
}

type edit struct {
	start, end int
	text       string
}

type change struct {
	name          string
	before, after []byte
	mode          os.FileMode
}

func prepare(program *loader.Analysis, graph *checker.Graph, stderr io.Writer) ([]change, error) {
	byFile := make(map[string][]edit)
	filePackages := make(map[string]*types.Package)
	roots := slices.Clone(graph.Roots)
	slices.SortFunc(roots, func(a, b *checker.Action) int { return strings.Compare(a.String(), b.String()) })
	for _, action := range roots {
	diagnostics:
		for _, diagnostic := range action.Diagnostics {
			if len(diagnostic.SuggestedFixes) == 0 {
				if _, err := fmt.Fprintf(stderr, "%s: %s\n", action.Package.Fset.Position(diagnostic.Pos), diagnostic.Message); err != nil {
					return nil, err
				}
				continue
			}
			// Suggested fixes are alternatives. Use the first, like Go fix.
			candidate := make(map[string][]edit)
			for _, textEdit := range diagnostic.SuggestedFixes[0].TextEdits {
				file := action.Package.Fset.File(textEdit.Pos)
				if file == nil || action.Package.Fset.File(textEdit.End) != file {
					return nil, errors.New("fix contains an invalid source range")
				}
				name := file.Name()
				filePackages[name] = action.Package.Types
				writable, selected := program.Writable[name]
				if !selected {
					continue diagnostics
				}
				if !writable {
					return nil, fmt.Errorf("%s: cannot safely fix generated or CGo source; no files changed", name)
				}
				start, end := file.Offset(textEdit.Pos), file.Offset(textEdit.End)
				if start > end || end > len(program.Sources[name]) {
					return nil, fmt.Errorf("%s: fix contains an invalid offset", name)
				}
				edits := candidate[name]
				if len(edits) > 0 && edits[len(edits)-1].start == start {
					last := &edits[len(edits)-1]
					last.end = end
					last.text += string(textEdit.NewText)
				} else {
					edits = append(edits, edit{start, end, string(textEdit.NewText)})
				}
				candidate[name] = edits
			}
			for name, edits := range candidate {
				byFile[name] = append(byFile[name], edits...)
			}
		}
	}
	var changes []change
	for name, edits := range byFile {
		slices.SortFunc(edits, func(a, b edit) int {
			if a.start != b.start {
				return a.start - b.start
			}
			if a.end != b.end {
				return a.end - b.end
			}
			return strings.Compare(a.text, b.text)
		})
		edits = slices.Compact(edits)
		before := program.Sources[name]
		var after []byte
		offset := 0
		for i, edit := range edits {
			if edit.start < offset || i > 0 && edit.start == edits[i-1].start {
				return nil, fmt.Errorf("%s: conflicting fixes; run individual analyzers; no files changed", name)
			}
			after = append(after, before[offset:edit.start]...)
			after = append(after, edit.text...)
			offset = edit.end
		}
		after = append(after, before[offset:]...)
		after, err := formatSource(name, after, filePackages[name])
		if err != nil {
			return nil, fmt.Errorf("%s: invalid fixed source; no files changed: %w", name, err)
		}
		if bytes.Equal(before, after) {
			continue
		}
		info, err := os.Stat(name)
		if err != nil {
			return nil, err
		}
		changes = append(changes, change{name, before, after, info.Mode().Perm()})
	}
	slices.SortFunc(changes, func(a, b change) int { return strings.Compare(a.name, b.name) })
	return changes, nil
}

func formatSource(name string, source []byte, pkg *types.Package) ([]byte, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, name, source, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	names := make(map[string]string)
	if pkg != nil {
		for _, imported := range pkg.Imports() {
			names[imported.Path()] = imported.Name()
		}
	}
	used := make(map[string]bool)
	ast.Inspect(file, func(node ast.Node) bool {
		if selector, ok := node.(*ast.SelectorExpr); ok {
			if ident, ok := selector.X.(*ast.Ident); ok && ident.Obj == nil {
				used[ident.Name] = true
			}
		}
		return true
	})
	for _, spec := range slices.Clone(file.Imports) {
		path, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			return nil, err
		}
		explicit := ""
		name := names[path]
		if spec.Name != nil {
			explicit = spec.Name.Name
			name = explicit
		}
		if name != "" && name != "_" && name != "." && !used[name] {
			astutil.DeleteNamedImport(fset, file, explicit, path)
		}
	}
	var output bytes.Buffer
	if err := format.Node(&output, fset, file); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}
