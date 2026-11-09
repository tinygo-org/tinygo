package main

import (
	"fmt"
	"go/ast"
)

// ReplaceDecl replaces the top-level declaration of name with src. In a
// grouped declaration only the spec for name is replaced.
func ReplaceDecl(name, src string) Rewrite {
	return func(p *Package, c *Cursor) bool {
		var old ast.Node
		switch n := c.Node().(type) {
		case *ast.FuncDecl:
			if n.Recv == nil && n.Name.Name == name {
				old = n
			}
		case *ast.GenDecl:
			if n.Lparen == 0 && definesOnly(n.Specs[0], name) {
				old = n
			}
		case *ast.ValueSpec, *ast.TypeSpec:
			if g, ok := c.Parent().(*ast.GenDecl); ok && g.Lparen != 0 && definesOnly(n.(ast.Spec), name) {
				old = n
			}
		}
		if old == nil {
			return false
		}
		d, err := p.ParseDecl(src, old.Pos())
		if err != nil {
			p.Fail(err)
			return false
		}
		if !definesOnlyDecl(d, name) {
			p.Fail(fmt.Errorf("replacement for %s defines %s", name, declNames(d)))
			return false
		}
		p.Remove(old)
		if _, ok := old.(ast.Spec); ok {
			g, ok := d.(*ast.GenDecl)
			if !ok || g.Tok != c.Parent().(*ast.GenDecl).Tok {
				p.Fail(fmt.Errorf("replacement for %s is not a %s spec", name, c.Parent().(*ast.GenDecl).Tok))
				return false
			}
			c.Replace(g.Specs[0])
			return true
		}
		c.Replace(d)
		return true
	}
}

func definesOnly(s ast.Spec, name string) bool {
	switch s := s.(type) {
	case *ast.TypeSpec:
		return s.Name.Name == name
	case *ast.ValueSpec:
		return len(s.Names) == 1 && s.Names[0].Name == name
	}
	return false
}

func definesOnlyDecl(d ast.Decl, name string) bool {
	ids := declIdents(d)
	return len(ids) == 1 && ids[0].Name == name
}

func declNames(d ast.Decl) []string {
	var names []string
	for _, id := range declIdents(d) {
		names = append(names, id.Name)
	}
	return names
}

// InsertFirst puts the statements in src at the start of the body of the
// function name.
func InsertFirst(name, src string) Rewrite {
	return func(p *Package, c *Cursor) bool {
		fd, ok := c.Node().(*ast.FuncDecl)
		if !ok || fd.Recv != nil || fd.Name.Name != name || fd.Body == nil {
			return false
		}
		stmts, err := p.ParseStmts(src, fd.Body.Lbrace)
		if err != nil {
			p.Fail(err)
			return false
		}
		fd.Body.List = append(stmts, fd.Body.List...)
		return true
	}
}
