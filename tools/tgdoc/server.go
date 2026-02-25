package main

import (
	"bytes"
	"go/ast"
	"go/doc"
	"go/format"
	"go/token"
	"html/template"
	"log"
	"net/http"
	"sort"
	"strings"
)

// IdentIndex is the precomputed identifier index across all targets.
type IdentIndex struct {
	Entries []*IdentEntry
	ByKey   map[string]*IdentEntry
	Total   int // total number of targets
}

// IdentEntry is one identifier (type, func, method, const, or var) across all targets.
type IdentEntry struct {
	Name   string
	Kind   string // "type", "func", "method", "const", "var"
	Count  int
	Key    string // URL key: "kind/name"
	Groups []IdentGroup
}

// IdentGroup is a set of targets with identical declaration for an identifier.
type IdentGroup struct {
	Decl    string   // representative formatted declaration
	Targets []string // sorted target names
}

func buildIdentIndex(results []Result) *IdentIndex {
	type occ struct {
		target string
		decl   string // grouping key + display
	}
	type ikey struct {
		name, kind string
	}

	idx := make(map[ikey][]occ)

	addFunc := func(target string, fset *token.FileSet, f *doc.Func, prefix string) {
		name := f.Name
		if prefix != "" {
			name = prefix + "." + f.Name
		}
		kind := "func"
		if f.Recv != "" {
			kind = "method"
		}
		decl := fmtNode(fset, f.Decl)
		idx[ikey{name, kind}] = append(idx[ikey{name, kind}], occ{target, decl})
	}

	addValues := func(target string, fset *token.FileSet, vals []*doc.Value, kind string) {
		for _, v := range vals {
			for _, name := range v.Names {
				typeKey := resolveValueType(fset, v.Decl, name)
				idx[ikey{name, kind}] = append(idx[ikey{name, kind}], occ{target, typeKey})
			}
		}
	}

	for _, r := range results {
		t := r.Target.Name
		fset := r.Fset
		pkg := r.Pkg

		for _, ty := range pkg.Types {
			decl := fmtNode(fset, ty.Decl)
			idx[ikey{ty.Name, "type"}] = append(idx[ikey{ty.Name, "type"}], occ{t, decl})

			for _, m := range ty.Methods {
				addFunc(t, fset, m, ty.Name)
			}
			for _, f := range ty.Funcs {
				addFunc(t, fset, f, "")
			}
			addValues(t, fset, ty.Consts, "const")
			addValues(t, fset, ty.Vars, "var")
		}

		for _, f := range pkg.Funcs {
			addFunc(t, fset, f, "")
		}
		addValues(t, fset, pkg.Consts, "const")
		addValues(t, fset, pkg.Vars, "var")
	}

	index := &IdentIndex{
		ByKey: make(map[string]*IdentEntry),
		Total: len(results),
	}

	for key, occs := range idx {
		groups := make(map[string][]string) // decl -> targets
		for _, o := range occs {
			groups[o.decl] = append(groups[o.decl], o.target)
		}

		var gs []IdentGroup
		for decl, targets := range groups {
			sort.Strings(targets)
			gs = append(gs, IdentGroup{Decl: decl, Targets: targets})
		}
		sort.Slice(gs, func(i, j int) bool { return len(gs[i].Targets) > len(gs[j].Targets) })

		urlKey := key.kind + "/" + key.name
		entry := &IdentEntry{
			Name:   key.name,
			Kind:   key.kind,
			Count:  len(occs),
			Key:    urlKey,
			Groups: gs,
		}
		index.Entries = append(index.Entries, entry)
		index.ByKey[urlKey] = entry
	}

	sort.Slice(index.Entries, func(i, j int) bool {
		if index.Entries[i].Count != index.Entries[j].Count {
			return index.Entries[i].Count > index.Entries[j].Count
		}
		return index.Entries[i].Name < index.Entries[j].Name
	})

	return index
}

// resolveValueType returns the effective type string for a const/var name
// within a GenDecl, handling iota type inheritance for constants.
func resolveValueType(fset *token.FileSet, gd *ast.GenDecl, name string) string {
	isConst := gd.Tok == token.CONST
	var lastType ast.Expr
	for _, spec := range gd.Specs {
		vs := spec.(*ast.ValueSpec)
		if vs.Type != nil {
			lastType = vs.Type
		}
		for _, n := range vs.Names {
			if n.Name != name {
				continue
			}
			effType := vs.Type
			if effType == nil && isConst {
				effType = lastType
			}
			if effType != nil {
				return fmtNode(fset, effType)
			}
			return "(untyped)"
		}
	}
	return "(unknown)"
}

// fmtNode formats an AST node to Go source. Separate from formatNode in main.go
// to make it clear this is the server's version (same logic).
func fmtNode(fset *token.FileSet, node ast.Node) string {
	var buf bytes.Buffer
	format.Node(&buf, fset, node)
	return buf.String()
}

// --- HTTP Server ---

func serve(addr string, results []Result) {
	index := buildIdentIndex(results)

	byTarget := make(map[string]Result, len(results))
	for _, r := range results {
		byTarget[r.Target.Name] = r
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		data := pageData{Index: index, Results: results}
		pageTmpl.Execute(w, data)
	})

	mux.HandleFunc("/id/", func(w http.ResponseWriter, r *http.Request) {
		key := strings.TrimPrefix(r.URL.Path, "/id/")
		entry, ok := index.ByKey[key]
		if !ok {
			http.NotFound(w, r)
			return
		}
		data := pageData{Index: index, Results: results, Selected: entry, SelectedKey: key}
		pageTmpl.Execute(w, data)
	})

	mux.HandleFunc("/target/", func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/target/")
		res, ok := byTarget[name]
		if !ok {
			http.NotFound(w, r)
			return
		}
		api := convertPackage(res.Target.Name, res.Pkg, res.Fset)
		targetTmpl.Execute(w, api)
	})

	log.Printf("serving documentation on http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

type pageData struct {
	Index       *IdentIndex
	Results     []Result
	Selected    *IdentEntry
	SelectedKey string
}

// --- Templates ---

var pageTmpl = template.Must(template.New("page").Funcs(template.FuncMap{
	"kindBadge": func(kind string) string {
		switch kind {
		case "type":
			return "T"
		case "func":
			return "F"
		case "method":
			return "M"
		case "const":
			return "C"
		case "var":
			return "V"
		}
		return "?"
	},
	"pre": func(s string) template.HTML {
		return template.HTML("<pre>" + template.HTMLEscapeString(s) + "</pre>")
	},
	"kindClass": func(kind string) string { return "kind-" + kind },
}).Parse(`<!DOCTYPE html>
<html><head>
<title>tgdoc{{if .Selected}} — {{.Selected.Name}}{{end}}</title>
<style>
* { box-sizing: border-box; margin: 0; padding: 0; }
body { font-family: system-ui, -apple-system, sans-serif; }
.layout { display: flex; height: 100vh; }

/* Sidebar */
.sidebar { width: 320px; min-width: 320px; border-right: 1px solid #ddd; display: flex; flex-direction: column; background: #fafafa; }
.sidebar-header { padding: 12px; border-bottom: 1px solid #ddd; }
.sidebar-header h2 { font-size: 1.1em; margin-bottom: 8px; }
.sidebar-header input { width: 100%; padding: 6px 10px; border: 1px solid #ccc; border-radius: 4px; font-size: 0.9em; }
.sidebar-header .filter-row { display: flex; gap: 4px; margin-top: 6px; flex-wrap: wrap; }
.sidebar-header .filter-btn { font-size: 0.7em; padding: 2px 6px; border: 1px solid #ccc; border-radius: 3px; background: white; cursor: pointer; }
.sidebar-header .filter-btn.active { background: #333; color: white; border-color: #333; }
.ident-list { overflow-y: auto; flex: 1; }
.ident-list a { display: flex; align-items: center; padding: 3px 12px; text-decoration: none; color: #333; font-size: 0.85em; border-left: 3px solid transparent; }
.ident-list a:hover { background: #f0f0f0; }
.ident-list a.active { background: #e8e8f4; border-left-color: #4444cc; }
.ident-list .badge { display: inline-block; width: 18px; height: 18px; line-height: 18px; text-align: center; border-radius: 3px; font-size: 0.65em; font-weight: bold; margin-right: 6px; color: white; flex-shrink: 0; }
.ident-list .count { margin-left: auto; color: #999; font-size: 0.8em; padding-left: 8px; flex-shrink: 0; }
.ident-list .ngroups { color: salmon; font-size: 0.8em; padding-left: 2px; flex-shrink: 0; }
.ident-list .name { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.kind-type .badge { background: #2196F3; }
.kind-func .badge { background: #4CAF50; }
.kind-method .badge { background: #9C27B0; }
.kind-const .badge { background: #FF9800; }
.kind-var .badge { background: #795548; }

/* Main content */
main { flex: 1; overflow-y: auto; padding: 2em; }
main h1 { font-size: 1.5em; border-bottom: 2px solid #333; padding-bottom: 0.3em; margin-bottom: 0.5em; }
main .subtitle { color: #666; margin-bottom: 1.5em; }
.group { margin-bottom: 1.5em; border: 1px solid #e0e0e0; border-radius: 6px; overflow: hidden; }
.group-header { background: #f5f5f5; padding: 10px 14px; font-size: 0.9em; border-bottom: 1px solid #e0e0e0; }
.group-header strong { font-size: 1.05em; }
.group pre { margin: 0; padding: 14px; background: #fcfcfc; overflow-x: auto; font-size: 0.85em; border-bottom: 1px solid #eee; }
.group .targets { padding: 10px 14px; display: flex; flex-wrap: wrap; gap: 4px; }
.group .targets a { background: #e8e8e8; padding: 2px 8px; border-radius: 3px; font-size: 0.78em; text-decoration: none; color: #333; }
.group .targets a:hover { background: #d0d0d0; }
.welcome nav a { margin-right: 1em; }
</style>
</head><body>
<div class="layout">
<aside class="sidebar">
<div class="sidebar-header">
<h2><a href="/" style="text-decoration:none;color:inherit">tgdoc</a></h2>
<input type="text" id="filter" placeholder="Filter identifiers..." oninput="filterList()">
<div class="filter-row">
<button class="filter-btn active" data-kind="all" onclick="toggleKind(this)">All</button>
<button class="filter-btn active" data-kind="type" onclick="toggleKind(this)">T</button>
<button class="filter-btn active" data-kind="func" onclick="toggleKind(this)">F</button>
<button class="filter-btn active" data-kind="method" onclick="toggleKind(this)">M</button>
<button class="filter-btn active" data-kind="const" onclick="toggleKind(this)">C</button>
<button class="filter-btn active" data-kind="var" onclick="toggleKind(this)">V</button>
</div>
</div>
<div class="ident-list">
{{range .Index.Entries}}
<a href="/id/{{.Key}}" class="{{kindClass .Kind}}{{if eq .Key $.SelectedKey}} active{{end}}" data-name="{{.Name}}" data-kind="{{.Kind}}">
<span class="badge">{{kindBadge .Kind}}</span>
<span class="name">{{.Name}}</span>
<span class="count">{{.Count}}{{if gt (len .Groups) 1}}<span class="ngroups">/{{len .Groups}}</span>{{end}}</span>
</a>
{{end}}
</div>
</aside>
<main>
{{if .Selected}}
<h1><span class="badge" style="display:inline-block;padding:2px 8px;border-radius:4px;font-size:0.6em;vertical-align:middle;color:white;background:{{if eq .Selected.Kind "type"}}#2196F3{{else if eq .Selected.Kind "func"}}#4CAF50{{else if eq .Selected.Kind "method"}}#9C27B0{{else if eq .Selected.Kind "const"}}#FF9800{{else}}#795548{{end}}">{{.Selected.Kind}}</span> {{.Selected.Name}}</h1>
<p class="subtitle">Present in {{.Selected.Count}} / {{.Index.Total}} targets
{{if eq (len .Selected.Groups) 1}}— identical across all targets{{else}}— {{len .Selected.Groups}} distinct signatures{{end}}</p>

{{range .Selected.Groups}}
<div class="group">
<div class="group-header"><strong>{{len .Targets}} target{{if ne (len .Targets) 1}}s{{end}}</strong></div>
{{pre .Decl}}
<div class="targets">{{range .Targets}}<a href="/target/{{.}}">{{.}}</a> {{end}}</div>
</div>
{{end}}

{{else}}
<div class="welcome">
<h1>tgdoc</h1>
<p class="subtitle">{{len .Index.Entries}} identifiers across {{.Index.Total}} targets</p>
<p>Select an identifier from the sidebar to see its appearances across targets.</p>
<p style="margin-top:1em">Browse by target:</p>
<nav style="margin-top:0.5em;column-count:3">
{{range .Results}}<a href="/target/{{.Target.Name}}" style="display:block;padding:2px 0">{{.Target.Name}}</a>
{{end}}
</nav>
</div>
{{end}}
</main>
</div>
<script>
var activeKinds = new Set(["type","func","method","const","var"]);
function filterList() {
  var q = document.getElementById("filter").value.toLowerCase();
  document.querySelectorAll(".ident-list a").forEach(function(a) {
    var name = a.dataset.name.toLowerCase();
    var kind = a.dataset.kind;
    var textMatch = !q || name.includes(q);
    var kindMatch = activeKinds.has(kind);
    a.style.display = (textMatch && kindMatch) ? "" : "none";
  });
}
function toggleKind(btn) {
  var kind = btn.dataset.kind;
  if (kind === "all") {
    var allActive = activeKinds.size === 5;
    document.querySelectorAll(".filter-btn").forEach(function(b) {
      if (allActive) { b.classList.remove("active"); activeKinds.clear(); }
      else { b.classList.add("active"); activeKinds.add(b.dataset.kind); }
    });
    if (!allActive) activeKinds.delete("all");
  } else {
    btn.classList.toggle("active");
    if (activeKinds.has(kind)) activeKinds.delete(kind); else activeKinds.add(kind);
    var allBtn = document.querySelector('[data-kind="all"]');
    if (activeKinds.size === 5) allBtn.classList.add("active"); else allBtn.classList.remove("active");
  }
  filterList();
}
</script>
</body></html>`))

var targetTmpl = template.Must(template.New("target").Funcs(template.FuncMap{
	"pre": func(s string) template.HTML {
		return template.HTML("<pre>" + template.HTMLEscapeString(s) + "</pre>")
	},
}).Parse(`<!DOCTYPE html>
<html><head>
<title>{{.Target}} — tgdoc</title>
<style>
body { font-family: system-ui, sans-serif; max-width: 900px; margin: 2em auto; padding: 0 1em; }
h1 { border-bottom: 2px solid #333; padding-bottom: 0.3em; }
h2 { margin-top: 2em; color: #444; }
h3 { margin-top: 1.5em; }
pre { background: #f5f5f5; padding: 0.8em; overflow-x: auto; font-size: 0.9em; border-radius: 4px; }
.doc { color: #555; margin-bottom: 0.5em; white-space: pre-wrap; }
.section { margin-left: 1em; }
a { text-decoration: none; }
a:hover { text-decoration: underline; }
</style>
</head><body>
<p><a href="/">← identifier index</a></p>
<h1>{{.Target}}</h1>
<p>package {{.Package}}</p>
{{if .Doc}}<div class="doc">{{.Doc}}</div>{{end}}

{{if .Constants}}
<h2>Constants</h2>
{{range .Constants}}
{{if .Doc}}<div class="doc">{{.Doc}}</div>{{end}}
{{pre .Decl}}
{{end}}
{{end}}

{{if .Variables}}
<h2>Variables</h2>
{{range .Variables}}
{{if .Doc}}<div class="doc">{{.Doc}}</div>{{end}}
{{pre .Decl}}
{{end}}
{{end}}

{{if .Functions}}
<h2>Functions</h2>
{{range .Functions}}
<h3>func {{.Name}}</h3>
{{if .Doc}}<div class="doc">{{.Doc}}</div>{{end}}
{{pre .Decl}}
{{end}}
{{end}}

{{if .Types}}
<h2>Types</h2>
{{range .Types}}
<h3>type {{.Name}}</h3>
{{if .Doc}}<div class="doc">{{.Doc}}</div>{{end}}
{{pre .Decl}}
{{if .Constants}}<div class="section"><h4>Associated Constants</h4>
{{range .Constants}}{{pre .Decl}}{{end}}</div>{{end}}
{{if .Variables}}<div class="section"><h4>Associated Variables</h4>
{{range .Variables}}{{pre .Decl}}{{end}}</div>{{end}}
{{if .Functions}}<div class="section"><h4>Constructors</h4>
{{range .Functions}}
{{if .Doc}}<div class="doc">{{.Doc}}</div>{{end}}
{{pre .Decl}}
{{end}}</div>{{end}}
{{if .Methods}}<div class="section"><h4>Methods</h4>
{{range .Methods}}
{{if .Doc}}<div class="doc">{{.Doc}}</div>{{end}}
{{pre .Decl}}
{{end}}</div>{{end}}
{{end}}
{{end}}

</body></html>`))
