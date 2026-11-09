package main

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"reflect"
	"slices"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
)

// RewriteSpec describes how to generate one TinyGo package from upstream.
type RewriteSpec struct {
	Path  string // import path in GOROOT
	Steps Pass
}

// Pass transforms the whole package.
type Pass func(p *Package) error

// Rewrite looks at one node and may replace, delete or insert around it.
// It returns true when it changed something.
type Rewrite func(p *Package, c *Cursor) bool

// Cursor wraps an astutil.Cursor so that Node follows Replace, and rewrites
// in a Chain see what earlier ones did.
type Cursor struct {
	c    *astutil.Cursor
	node ast.Node
}

func (c *Cursor) Node() ast.Node { return c.node }

func (c *Cursor) Parent() ast.Node { return c.c.Parent() }

func (c *Cursor) Replace(n ast.Node) {
	c.c.Replace(n)
	c.node = n
}

// Package is the working copy of a package while its RewriteSpec is applied.
type Package struct {
	Fset      *token.FileSet
	Path      string
	GoVersion string
	Upstream  []*ast.File
	Local     []*ast.File // TinyGo files next to the output
	Info      *types.Info

	removed []span // code whose comments are dropped
	err     error
}

// Fail records an error in a rewrite. Walk returns it.
func (p *Package) Fail(err error) {
	if p.err == nil {
		p.err = err
	}
}

type span struct {
	pos, end token.Pos
}

// Remove records that the code in n was removed, so its comments go too.
func (p *Package) Remove(n ast.Node) {
	p.removed = append(p.removed, span{n.Pos(), n.End()})
}

// ParseDecl parses a single top-level declaration. Its positions are all set
// to pos, so it prints where the code it replaces was.
func (p *Package) ParseDecl(src string, pos token.Pos) (ast.Decl, error) {
	f, err := parser.ParseFile(p.Fset, "", "package p\n"+src, 0)
	if err != nil {
		return nil, err
	}
	if len(f.Decls) != 1 {
		return nil, fmt.Errorf("want one declaration in %q", src)
	}
	setPos(f.Decls[0], pos)
	return f.Decls[0], nil
}

// ParseStmts parses a list of statements with positions set to pos.
func (p *Package) ParseStmts(src string, pos token.Pos) ([]ast.Stmt, error) {
	f, err := parser.ParseFile(p.Fset, "", "package p\nfunc _() {\n"+src+"\n}", 0)
	if err != nil {
		return nil, err
	}
	list := f.Decls[0].(*ast.FuncDecl).Body.List
	for _, s := range list {
		setPos(s, pos)
	}
	return list, nil
}

func setPos(n ast.Node, pos token.Pos) {
	posType := reflect.TypeFor[token.Pos]()
	ast.Inspect(n, func(n ast.Node) bool {
		if n == nil {
			return false
		}
		v := reflect.ValueOf(n).Elem()
		for i := range v.NumField() {
			if f := v.Field(i); f.Type() == posType && f.Int() != 0 {
				f.Set(reflect.ValueOf(pos))
			}
		}
		return true
	})
}

// Check type checks the upstream and local files and refreshes p.Info. Unused
// imports are not errors, the output only imports what it uses.
func (p *Package) Check() error {
	p.Info = &types.Info{
		Defs:      map[*ast.Ident]types.Object{},
		Uses:      map[*ast.Ident]types.Object{},
		Implicits: map[ast.Node]types.Object{},
	}
	var firstErr error
	conf := types.Config{
		GoVersion: p.GoVersion,
		Importer:  importer.ForCompiler(p.Fset, "source", nil),
		Error: func(err error) {
			if !strings.Contains(err.(types.Error).Msg, "imported and not used") && firstErr == nil {
				firstErr = err
			}
		},
	}
	conf.Check(p.Path, p.Fset, append(slices.Clone(p.Upstream), p.Local...), p.Info)
	return firstErr
}

// Pipeline runs passes in order. Each pass sees the result of the one before.
func Pipeline(ps ...Pass) Pass {
	return func(p *Package) error {
		for i, pass := range ps {
			if err := pass(p); err != nil {
				return fmt.Errorf("step %d: %w", i+1, err)
			}
		}
		return nil
	}
}

// Walk applies r to every node of the upstream files. It fails if r changed
// nothing, which means upstream no longer has what r expects.
func Walk(r Rewrite) Pass {
	return func(p *Package) error {
		changed := false
		for _, f := range p.Upstream {
			apply(p, f, r, &changed)
		}
		if p.err != nil {
			return p.err
		}
		if !changed {
			return fmt.Errorf("no change in %s %s", p.GoVersion, p.Path)
		}
		return nil
	}
}

func apply(p *Package, root ast.Node, r Rewrite, changed *bool) ast.Node {
	return astutil.Apply(root, func(ac *astutil.Cursor) bool {
		if r(p, &Cursor{ac, ac.Node()}) {
			*changed = true
		}
		return true
	}, nil)
}

// Chain runs rewrites on the same node in one walk.
func Chain(rs ...Rewrite) Rewrite {
	return func(p *Package, c *Cursor) bool {
		changed := false
		for _, r := range rs {
			if r(p, c) {
				changed = true
			}
		}
		return changed
	}
}

// In limits r to the top-level declaration called name.
func In(name string, r Rewrite) Rewrite {
	return func(p *Package, c *Cursor) bool {
		d, ok := c.Node().(ast.Decl)
		if !ok || !slices.ContainsFunc(declIdents(d), func(id *ast.Ident) bool { return id.Name == name }) {
			return false
		}
		if _, ok := c.Parent().(*ast.File); !ok {
			return false
		}
		changed := false
		if n := apply(p, d, r, &changed); n != d {
			c.Replace(n)
		}
		return changed
	}
}

// declIdents returns the identifiers a top-level declaration defines.
func declIdents(d ast.Decl) []*ast.Ident {
	switch d := d.(type) {
	case *ast.FuncDecl:
		if d.Recv == nil {
			return []*ast.Ident{d.Name}
		}
	case *ast.GenDecl:
		var ids []*ast.Ident
		for _, s := range d.Specs {
			switch s := s.(type) {
			case *ast.TypeSpec:
				ids = append(ids, s.Name)
			case *ast.ValueSpec:
				ids = append(ids, s.Names...)
			}
		}
		return ids
	}
	return nil
}
