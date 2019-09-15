package compiler

// This file manages symbols, that is, functions and globals. It reads their
// pragmas, determines the link name, etc.

import (
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"github.com/tinygo-org/tinygo/loader"
	"golang.org/x/tools/go/ssa"
	"tinygo.org/x/go-llvm"
)

// globalInfo contains some information about a specific global. By default,
// linkName is equal to .RelString(nil) on a global and extern is false, but for
// some symbols this is different (due to //go:extern for example).
type globalInfo struct {
	linkName string // go:extern
	extern   bool   // go:extern
}

// loadASTComments loads comments on globals from the AST, for use later in the
// program. In particular, they are required for //go:extern pragmas on globals.
func (c *Compiler) loadASTComments(lprogram *loader.Program) {
	c.astComments = map[string]*ast.CommentGroup{}
	for _, pkgInfo := range lprogram.Sorted() {
		for _, file := range pkgInfo.Files {
			for _, decl := range file.Decls {
				switch decl := decl.(type) {
				case *ast.GenDecl:
					switch decl.Tok {
					case token.VAR:
						if len(decl.Specs) != 1 {
							continue
						}
						for _, spec := range decl.Specs {
							switch spec := spec.(type) {
							case *ast.ValueSpec: // decl.Tok == token.VAR
								for _, name := range spec.Names {
									id := pkgInfo.Pkg.Path() + "." + name.Name
									c.astComments[id] = decl.Doc
								}
							}
						}
					}
				}
			}
		}
	}
}

// getGlobal returns a LLVM IR global value for a Go SSA global. It is added to
// the LLVM IR if it has not been added already.
func (c *Compiler) getGlobal(g *ssa.Global) llvm.Value {
	info := c.getGlobalInfo(g)
	llvmGlobal := c.mod.NamedGlobal(info.linkName)
	if llvmGlobal.IsNil() {
		llvmType := c.getLLVMType(g.Type().(*types.Pointer).Elem())
		llvmGlobal = llvm.AddGlobal(c.mod, llvmType, info.linkName)
		if !info.extern {
			llvmGlobal.SetInitializer(llvm.ConstNull(llvmType))
			llvmGlobal.SetLinkage(llvm.InternalLinkage)
		}
	}
	return llvmGlobal
}

// getGlobalInfo returns some information about a specific global.
func (c *Compiler) getGlobalInfo(g *ssa.Global) globalInfo {
	info := globalInfo{}
	if strings.HasPrefix(g.Name(), "C.") {
		// Created by CGo: such a name cannot be created by regular C code.
		info.linkName = g.Name()[2:]
		info.extern = true
	} else {
		// Pick the default linkName.
		info.linkName = g.RelString(nil)
		// Check for //go: pragmas, which may change the link name (among
		// others).
		doc := c.astComments[info.linkName]
		if doc != nil {
			info.parsePragmas(doc)
		}
	}
	return info
}

// Parse //go: pragma comments from the source. In particular, it parses the
// //go:extern pragma on globals.
func (info *globalInfo) parsePragmas(doc *ast.CommentGroup) {
	for _, comment := range doc.List {
		if !strings.HasPrefix(comment.Text, "//go:") {
			continue
		}
		parts := strings.Fields(comment.Text)
		switch parts[0] {
		case "//go:extern":
			info.extern = true
			if len(parts) == 2 {
				info.linkName = parts[1]
			}
		}
	}
}
