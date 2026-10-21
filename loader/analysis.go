package loader

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/build"
	"go/types"
	"go/version"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/tinygo-org/tinygo/compileopts"
	"github.com/tinygo-org/tinygo/goenv"
	"golang.org/x/tools/go/packages"
)

// Analysis holds a type-checked package graph and its editable source files.
type Analysis struct {
	Packages []*packages.Package
	Sources  map[string][]byte
	Writable map[string]bool
	Skipped  []string
}

// LoadForAnalysis retains distinct package identities when includeTests is true.
// Overlay keys are absolute original source paths, not cached GOROOT paths.
func LoadForAnalysis(config *compileopts.Config, patterns []string, sizes types.Sizes, overlay map[string][]byte, includeTests bool) (*Analysis, error) {
	if sizes == nil {
		return nil, fmt.Errorf("analysis requires target type sizes")
	}
	if len(patterns) == 0 {
		patterns = []string{"."}
	}
	for _, pattern := range patterns {
		if strings.HasSuffix(pattern, ".go") {
			return nil, fmt.Errorf("named Go files are not supported; select a package instead of %s", pattern)
		}
	}
	patterns, err := analysisPatterns(config, patterns)
	if err != nil {
		return nil, err
	}
	cfg := *config
	cfg.TestConfig.CompileTestBinary = includeTests
	program, err := load(&cfg, patterns, types.Config{Sizes: sizes}, true, overlay)
	if err != nil {
		return nil, err
	}
	if err := program.Parse(); err != nil {
		return nil, err
	}
	result := &Analysis{
		Sources:  program.sources,
		Writable: make(map[string]bool),
	}
	graph := make(map[string]*packages.Package)
	byTypes := make(map[*types.Package]*packages.Package)
	minimumGoVersion := fmt.Sprintf("go1.%d", compileopts.MinimumGoMinorVersion)
	tinySrc, err := analysisPath(filepath.Join(goenv.Get("TINYGOROOT"), "src"))
	if err != nil {
		return nil, fmt.Errorf("resolve TinyGo source root %q: %w", filepath.Join(goenv.Get("TINYGOROOT"), "src"), err)
	}
	for _, pkg := range program.sorted {
		converted := &packages.Package{
			ID: pkg.ImportPath, Name: pkg.Name, PkgPath: pkg.Pkg.Path(),
			Dir: pkg.OriginalDir(), ForTest: pkg.ForTest,
			Fset:  program.fset,
			Types: pkg.Pkg, TypesInfo: &pkg.info, TypesSizes: sizes,
			Imports: make(map[string]*packages.Package),
		}
		if pkg.Module.Path != "" {
			converted.Module = &packages.Module{
				Path: pkg.Module.Path, Version: pkg.Module.Version,
				GoVersion: pkg.Module.GoVersion, Dir: pkg.Module.Dir,
				GoMod: pkg.Module.GoMod, Main: pkg.Module.Main,
			}
		}
		for _, file := range pkg.Files {
			name := program.fset.File(file.FileStart).Name()
			if _, ok := program.sources[name]; ok {
				converted.GoFiles = append(converted.GoFiles, name)
				if !program.rewritten[name] {
					converted.Syntax = append(converted.Syntax, file)
					converted.CompiledGoFiles = append(converted.CompiledGoFiles, name)
				}
			}
			if pkg.Standard || within(tinySrc, name) {
				// Unconstrained library files must use the minimum supported Go version.
				pkg.info.FileVersions[file] = version.Lang(versionMax(minimumGoVersion, file.GoVersion))
			}
		}
		graph[pkg.ImportPath] = converted
		byTypes[pkg.Pkg] = converted
		if pkg.DepOnly || strings.HasSuffix(pkg.ImportPath, ".test") {
			continue
		}
		result.Packages = append(result.Packages, converted)
		var sourceRoot string
		if pkg.Module.Main {
			sourceRoot = pkg.Module.Dir
		} else if pkg.Module.Path == "" && !pkg.Standard && pkg.Root != "" {
			sourceRoot = filepath.Join(pkg.Root, "src")
		}
		if sourceRoot != "" {
			originalRoot := sourceRoot
			sourceRoot, err = analysisPath(sourceRoot)
			if err != nil {
				return nil, fmt.Errorf("resolve analysis package root %q: %w", originalRoot, err)
			}
		}
		for _, file := range pkg.Files {
			name := program.fset.File(file.FileStart).Name()
			if _, exists := program.sources[name]; !exists {
				continue
			}
			if sourceRoot != "" && within(sourceRoot, name) || within(tinySrc, name) {
				// CGo rewrites ASTs before type checking.
				writable := !program.rewritten[name] && !ast.IsGenerated(file)
				if previous, exists := result.Writable[name]; exists {
					writable = writable && previous
				}
				result.Writable[name] = writable
			}
		}
	}
	for _, pkg := range program.sorted {
		converted := graph[pkg.ImportPath]
		for _, file := range pkg.Files {
			for _, spec := range file.Imports {
				path, err := strconv.Unquote(spec.Path.Value)
				if err != nil {
					return nil, err
				}
				if path == "C" {
					continue
				}
				imported, err := pkg.Import(path)
				if err != nil {
					return nil, err
				}
				dependency := byTypes[imported]
				if dependency == nil {
					return nil, fmt.Errorf("%s: missing analysis dependency %s", pkg.ImportPath, path)
				}
				converted.Imports[path] = dependency
			}
		}
	}
	if len(result.Packages) == 0 {
		return nil, fmt.Errorf("no packages matched %s", strings.Join(patterns, " "))
	}
	for name := range result.Writable {
		if program.rewritten[name] {
			result.Skipped = append(result.Skipped, name)
		}
	}
	sort.Strings(result.Skipped)
	return result, nil
}

func analysisPatterns(config *compileopts.Config, patterns []string) ([]string, error) {
	root, err := analysisPath(filepath.Join(goenv.Get("TINYGOROOT"), "src"))
	if err != nil {
		return nil, fmt.Errorf("resolve TinyGo source root %q: %w", filepath.Join(goenv.Get("TINYGOROOT"), "src"), err)
	}
	var result []string
	for _, pattern := range patterns {
		if !filepath.IsAbs(pattern) && !strings.HasPrefix(pattern, ".") {
			result = append(result, pattern)
			continue
		}
		patternPath := filepath.FromSlash(pattern)
		recursive := strings.HasSuffix(patternPath, string(filepath.Separator)+"...")
		base := strings.TrimSuffix(patternPath, string(filepath.Separator)+"...")
		if strings.Contains(base, "...") {
			result = append(result, pattern)
			continue
		}
		if config.Options.Directory != "" && !filepath.IsAbs(base) {
			base = filepath.Join(config.Options.Directory, base)
		}
		base, err = filepath.Abs(base)
		if err != nil {
			return nil, err
		}
		originalBase := base
		base, err = analysisPath(base)
		if err != nil {
			return nil, fmt.Errorf("resolve analysis pattern %q: %w", originalBase, err)
		}
		if !within(root, base) {
			result = append(result, pattern)
			continue
		}
		rel, err := filepath.Rel(root, base)
		if err != nil {
			return nil, err
		}
		if recursive {
			dirs, err := analysisPackageDirs(base, config)
			if err != nil {
				return nil, err
			}
			matched := false
			for _, dir := range dirs {
				rel, err := filepath.Rel(root, dir)
				if err != nil {
					return nil, err
				}
				if rel != "." {
					result = append(result, filepath.ToSlash(rel))
					matched = true
				}
			}
			if !matched {
				return nil, fmt.Errorf("no packages matched %s", pattern)
			}
			continue
		}
		if rel == "." {
			return nil, fmt.Errorf("no Go package in TinyGo src root; select a package or use ./src/...")
		}
		result = append(result, filepath.ToSlash(rel))
	}
	return result, nil
}

func analysisPackageDirs(root string, config *compileopts.Config) ([]string, error) {
	context := build.Default
	context.GOOS = config.GOOS()
	context.GOARCH = config.GOARCH()
	context.CgoEnabled = true
	context.BuildTags = append(append([]string(nil), context.BuildTags...), config.BuildTags()...)
	found := make(map[string]bool)
	err := filepath.WalkDir(root, func(name string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if name != root {
				base := entry.Name()
				if strings.HasPrefix(base, ".") || strings.HasPrefix(base, "_") ||
					base == "testdata" || base == "vendor" {
					return filepath.SkipDir
				}
			}
			return nil
		}
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			return nil
		}
		match, err := context.MatchFile(filepath.Dir(name), entry.Name())
		if err != nil {
			return err
		}
		if match {
			found[filepath.Dir(name)] = true
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk TinyGo source tree %q: %w", root, err)
	}
	dirs := make([]string, 0, len(found))
	for dir := range found {
		dirs = append(dirs, dir)
	}
	sort.Strings(dirs)
	return dirs, nil
}

func analysisImportPath(path string) string {
	if index := strings.Index(path, " ["); index >= 0 {
		return path[:index]
	}
	return path
}

func versionMax(a, b string) string {
	if version.Compare(a, b) < 0 {
		return b
	}
	return a
}

func within(dir, name string) bool {
	rel, err := filepath.Rel(dir, name)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func analysisOverlay(goroot string, overlay map[string][]byte) (func(), []string, error) {
	if len(overlay) == 0 {
		return func() {}, nil, nil
	}
	dir, err := os.MkdirTemp("", "tinygo-analysis-")
	if err != nil {
		return nil, nil, err
	}
	cleanup := func() { os.RemoveAll(dir) }
	replacements := make(map[string]string)
	for name, data := range overlay {
		file, err := os.CreateTemp(dir, "*.go")
		if err != nil {
			cleanup()
			return nil, nil, err
		}
		if _, err := file.Write(data); err != nil {
			file.Close()
			cleanup()
			return nil, nil, err
		}
		if err := file.Close(); err != nil {
			cleanup()
			return nil, nil, err
		}
		replacements[name] = file.Name()
		for _, root := range []string{goenv.Get("TINYGOROOT"), goenv.Get("GOROOT")} {
			sourceRoot, err := analysisPath(filepath.Join(root, "src"))
			if err != nil {
				cleanup()
				return nil, nil, fmt.Errorf("resolve analysis overlay source root %q: %w", filepath.Join(root, "src"), err)
			}
			if within(sourceRoot, name) {
				rel, err := filepath.Rel(sourceRoot, name)
				if err != nil {
					cleanup()
					return nil, nil, err
				}
				replacements[filepath.Join(goroot, "src", rel)] = file.Name()
			}
		}
	}
	data, err := json.Marshal(struct{ Replace map[string]string }{replacements})
	if err == nil {
		err = os.WriteFile(filepath.Join(dir, "overlay.json"), data, 0600)
	}
	if err != nil {
		cleanup()
		return nil, nil, err
	}
	return cleanup, []string{"-overlay=" + filepath.Join(dir, "overlay.json")}, nil
}
