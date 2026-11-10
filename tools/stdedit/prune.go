package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"slices"
)

// Prune removes local variables left unused by earlier passes, then every
// top-level declaration not reachable from exported names, init, keep, and
// the TinyGo files next to the output.
func Prune(keep ...string) Pass {
	return func(p *Package) error { return prune(p, true, keep) }
}

// Extract is Prune with only names, init and the TinyGo files as roots, so
// the output has just the part of the package TinyGo needs.
func Extract(names ...string) Pass {
	return func(p *Package) error { return prune(p, false, names) }
}

func prune(p *Package, exported bool, keep []string) error {
	for {
		unused, err := p.check()
		if err != nil {
			return err
		}
		if len(unused) == 0 {
			break
		}
		dead := map[types.Object]bool{}
		for id, obj := range p.Info.Defs {
			if obj != nil && slices.ContainsFunc(unused, func(e types.Error) bool { return e.Pos == id.Pos() }) {
				dead[obj] = true
			}
		}
		for _, f := range p.Upstream {
			removeDead(p, f, dead)
		}
	}

	declOf := map[types.Object]ast.Decl{}
	methods := map[types.Object][]ast.Decl{}
	var work []ast.Node
	for _, f := range p.Upstream {
		for _, d := range f.Decls {
			if fd, ok := d.(*ast.FuncDecl); ok && fd.Recv != nil {
				recv := recvType(p.Info, fd)
				methods[recv] = append(methods[recv], d)
				continue
			}
			for _, id := range declIdents(d) {
				obj := p.Info.Defs[id]
				if obj == nil {
					continue
				}
				declOf[obj] = d
				if (exported && obj.Exported()) || obj.Name() == "init" {
					work = append(work, d)
				}
			}
		}
	}
	for _, name := range keep {
		obj := p.Types.Scope().Lookup(name)
		if obj == nil || declOf[obj] == nil {
			return fmt.Errorf("prune: keep %s not found in %s %s", name, p.GoVersion, p.Path)
		}
		work = append(work, declOf[obj])
	}
	for _, f := range p.Local {
		work = append(work, f)
	}

	kept := map[ast.Node]bool{}
	for len(work) > 0 {
		n := work[len(work)-1]
		work = work[:len(work)-1]
		if kept[n] {
			continue
		}
		kept[n] = true
		if d, ok := n.(ast.Decl); ok {
			for _, id := range declIdents(d) {
				for _, m := range methods[p.Info.Defs[id]] {
					work = append(work, m)
				}
			}
		}
		for _, obj := range uses(p.Info, n) {
			if d := declOf[obj]; d != nil {
				work = append(work, d)
			}
		}
	}

	// A comment between two declarations goes with the one after it, and
	// a comment on the last line of a declaration goes with it.
	for _, f := range p.Upstream {
		tf := p.Fset.File(f.Pos())
		lineEnd := func(pos token.Pos) token.Pos {
			if l := tf.Line(pos); l < tf.LineCount() {
				return tf.LineStart(l+1) - 1
			}
			return token.Pos(tf.Base() + tf.Size())
		}
		prev := f.Name.End()
		var decls []ast.Decl
		for _, d := range f.Decls {
			if g, ok := d.(*ast.GenDecl); ok && g.Tok == token.IMPORT {
				decls = append(decls, d)
				continue
			}
			start := min(lineEnd(prev)+1, d.Pos())
			prev = d.End()
			if kept[d] {
				decls = append(decls, d)
				continue
			}
			p.Remove(start, lineEnd(d.End()))
		}
		f.Decls = decls
	}
	return p.Check()
}

// recvType returns the named type a method is declared on.
func recvType(info *types.Info, fd *ast.FuncDecl) types.Object {
	t := info.Defs[fd.Name].Type().(*types.Signature).Recv().Type()
	if p, ok := t.(*types.Pointer); ok {
		t = p.Elem()
	}
	return t.(*types.Named).Origin().Obj()
}

// uses returns the objects referred to inside n.
func uses(info *types.Info, n ast.Node) []types.Object {
	var objs []types.Object
	ast.Inspect(n, func(n ast.Node) bool {
		id, ok := n.(*ast.Ident)
		if !ok {
			return true
		}
		switch obj := info.Uses[id].(type) {
		case nil:
		case *types.Func:
			objs = append(objs, obj.Origin())
		case *types.Var:
			objs = append(objs, obj.Origin())
		default:
			objs = append(objs, obj)
		}
		return true
	})
	return objs
}

// removeDead drops the declarations of and assignments to the dead local
// variables in f.
func removeDead(p *Package, f *ast.File, dead map[types.Object]bool) {
	isDead := func(x ast.Expr) bool {
		id, ok := x.(*ast.Ident)
		if !ok {
			return false
		}
		if obj := p.Info.Defs[id]; obj != nil {
			return dead[obj]
		}
		return dead[p.Info.Uses[id]]
	}
	filter := func(list []ast.Stmt) []ast.Stmt {
		return slices.DeleteFunc(list, func(s ast.Stmt) bool {
			var lhs, rhs []ast.Expr
			switch s := s.(type) {
			case *ast.AssignStmt:
				lhs, rhs = s.Lhs, s.Rhs
			case *ast.DeclStmt:
				g := s.Decl.(*ast.GenDecl)
				if g.Tok != token.VAR || len(g.Specs) != 1 {
					return false
				}
				vs := g.Specs[0].(*ast.ValueSpec)
				for _, n := range vs.Names {
					lhs = append(lhs, n)
				}
				rhs = vs.Values
			default:
				return false
			}
			if !slices.ContainsFunc(lhs, func(x ast.Expr) bool { return !isDead(x) }) && !hasCall(rhs) {
				p.Remove(s.Pos(), s.End())
				return true
			}
			return false
		})
	}
	ast.Inspect(f, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.BlockStmt:
			n.List = filter(n.List)
		case *ast.CaseClause:
			n.Body = filter(n.Body)
		}
		return true
	})
}

func hasCall(exprs []ast.Expr) bool {
	found := false
	for _, x := range exprs {
		ast.Inspect(x, func(n ast.Node) bool {
			if _, ok := n.(*ast.CallExpr); ok {
				found = true
			}
			return !found
		})
	}
	return found
}
