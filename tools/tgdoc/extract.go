package main

import (
	"bufio"
	"bytes"
	"fmt"
	"go/ast"
	"go/build/constraint"
	"go/doc"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// buildTagSet creates the full set of tags that should evaluate to true
// for a given target, matching TinyGo's Config.BuildTags() behavior.
func buildTagSet(t Target) map[string]bool {
	tags := make(map[string]bool)
	for _, tag := range t.BuildTags {
		tags[tag] = true
	}
	if t.GOOS != "" {
		tags[t.GOOS] = true
	}
	if t.GOARCH != "" {
		tags[t.GOARCH] = true
	}
	tags["tinygo"] = true
	tags["purego"] = true
	tags["osusergo"] = true
	tags["math_big_pure_go"] = true
	if t.GC != "" {
		tags["gc."+t.GC] = true
	}
	if t.Scheduler != "" {
		tags["scheduler."+t.Scheduler] = true
	}
	if t.Serial != "" {
		tags["serial."+t.Serial] = true
	}
	// Go version tags — TinyGo currently tracks Go 1.22.
	for i := 1; i <= 22; i++ {
		tags[fmt.Sprintf("go1.%d", i)] = true
	}
	return tags
}

// extractBuildConstraint reads the file header (before the package clause)
// looking for a //go:build constraint line.
func extractBuildConstraint(data []byte) (constraint.Expr, error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "package ") {
			break
		}
		if constraint.IsGoBuild(trimmed) {
			return constraint.Parse(trimmed)
		}
	}
	return nil, nil
}

// ExtractDocs filters .go files in pkgDir by the target's build tags,
// parses them, and returns a *doc.Package via go/doc.NewFromFiles.
func ExtractDocs(t Target, pkgDir string, allDecls bool) (*doc.Package, *token.FileSet, error) {
	tags := buildTagSet(t)

	entries, err := os.ReadDir(pkgDir)
	if err != nil {
		return nil, nil, err
	}

	fset := token.NewFileSet()
	var files []*ast.File

	for _, e := range entries {
		name := e.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}

		path := filepath.Join(pkgDir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, nil, err
		}

		expr, err := extractBuildConstraint(data)
		if err != nil {
			continue // skip files with unparseable constraints
		}

		if expr != nil && !expr.Eval(func(tag string) bool { return tags[tag] }) {
			continue
		}

		f, err := parser.ParseFile(fset, path, data, parser.ParseComments)
		if err != nil {
			continue // skip files with syntax errors
		}

		files = append(files, f)
	}

	if len(files) == 0 {
		return nil, nil, fmt.Errorf("no matching files for target %s", t.Name)
	}

	var opts []any
	if allDecls {
		opts = append(opts, doc.AllDecls)
	}

	docPkg, err := doc.NewFromFiles(fset, files, "machine", opts...)
	if err != nil {
		return nil, nil, err
	}

	return docPkg, fset, nil
}
