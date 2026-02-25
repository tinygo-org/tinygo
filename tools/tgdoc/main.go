package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/doc"
	"go/format"
	"go/token"
	"log"
	"os"
	"sort"
	"strings"
)

// Result holds the extracted documentation for a single target.
type Result struct {
	Target Target
	Pkg    *doc.Package
	Fset   *token.FileSet
}

// JSON-serializable API types.

type PackageAPI struct {
	Target    string     `json:"target"`
	Package   string     `json:"package"`
	Doc       string     `json:"doc,omitempty"`
	Constants []APIValue `json:"constants,omitempty"`
	Variables []APIValue `json:"variables,omitempty"`
	Types     []APIType  `json:"types,omitempty"`
	Functions []APIFunc  `json:"functions,omitempty"`
}

type APIValue struct {
	Doc   string   `json:"doc,omitempty"`
	Names []string `json:"names"`
	Decl  string   `json:"decl"`
}

type APIType struct {
	Doc       string     `json:"doc,omitempty"`
	Name      string     `json:"name"`
	Decl      string     `json:"decl"`
	Constants []APIValue `json:"constants,omitempty"`
	Variables []APIValue `json:"variables,omitempty"`
	Functions []APIFunc  `json:"functions,omitempty"`
	Methods   []APIFunc  `json:"methods,omitempty"`
}

type APIFunc struct {
	Doc  string `json:"doc,omitempty"`
	Name string `json:"name"`
	Decl string `json:"decl"`
	Recv string `json:"recv,omitempty"`
}

func formatNode(fset *token.FileSet, node ast.Node) string {
	var buf bytes.Buffer
	format.Node(&buf, fset, node)
	return buf.String()
}

func convertValues(fset *token.FileSet, vals []*doc.Value) []APIValue {
	out := make([]APIValue, 0, len(vals))
	for _, v := range vals {
		out = append(out, APIValue{
			Doc:   v.Doc,
			Names: v.Names,
			Decl:  formatNode(fset, v.Decl),
		})
	}
	return out
}

func convertFuncs(fset *token.FileSet, funcs []*doc.Func) []APIFunc {
	out := make([]APIFunc, 0, len(funcs))
	for _, f := range funcs {
		out = append(out, APIFunc{
			Doc:  f.Doc,
			Name: f.Name,
			Decl: formatNode(fset, f.Decl),
			Recv: f.Recv,
		})
	}
	return out
}

func convertPackage(targetName string, pkg *doc.Package, fset *token.FileSet) *PackageAPI {
	api := &PackageAPI{
		Target:    targetName,
		Package:   pkg.Name,
		Doc:       pkg.Doc,
		Constants: convertValues(fset, pkg.Consts),
		Variables: convertValues(fset, pkg.Vars),
		Functions: convertFuncs(fset, pkg.Funcs),
	}
	for _, t := range pkg.Types {
		api.Types = append(api.Types, APIType{
			Doc:       t.Doc,
			Name:      t.Name,
			Decl:      formatNode(fset, t.Decl),
			Constants: convertValues(fset, t.Consts),
			Variables: convertValues(fset, t.Vars),
			Functions: convertFuncs(fset, t.Funcs),
			Methods:   convertFuncs(fset, t.Methods),
		})
	}
	return api
}

func main() {
	jsonOutput := flag.Bool("json", false, "output JSON to stdout")
	httpAddr := flag.String("http", "", "serve HTTP documentation (e.g. :8080)")
	allDecls := flag.Bool("all", false, "include unexported identifiers")
	includeBase := flag.Bool("base", false, "include base/parent targets")
	flag.Parse()

	args := flag.Args()
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: tgdoc [flags] <targets-dir> <package-dir>\n")
		fmt.Fprintf(os.Stderr, "\nexample: tgdoc -http :8080 ./targets ./src/machine\n")
		fmt.Fprintf(os.Stderr, "\nFlags:\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	targetsDir := args[0]
	pkgDir := args[1]

	if *httpAddr != "" && !strings.Contains(*httpAddr, ":") {
		fmt.Fprintf(os.Stderr, "error: -http value %q doesn't look like an address (expected e.g. :8080)\n", *httpAddr)
		os.Exit(1)
	}

	targets, err := LoadTargets(targetsDir, *includeBase)
	if err != nil {
		log.Fatal(err)
	}

	var results []Result
	for _, t := range targets {
		pkg, fset, err := ExtractDocs(t, pkgDir, *allDecls)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: %s: %v\n", t.Name, err)
			continue
		}
		results = append(results, Result{t, pkg, fset})
	}

	sort.Slice(results, func(i, j int) bool { return results[i].Target.Name < results[j].Target.Name })

	if *httpAddr != "" {
		serve(*httpAddr, results)
		return
	}

	// Default: JSON output.
	_ = jsonOutput
	apis := make([]*PackageAPI, 0, len(results))
	for _, r := range results {
		apis = append(apis, convertPackage(r.Target.Name, r.Pkg, r.Fset))
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(apis); err != nil {
		log.Fatal(err)
	}
}
