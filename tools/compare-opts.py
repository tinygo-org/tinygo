#!/usr/bin/env python3
"""Compare what TinyGo and the Go compiler leave behind after optimization.

Bounds checks:  tinygo -print-bounds   vs  go -gcflags=-d=ssa/check_bce
Heap values:    tinygo -print-allocs   vs  go -gcflags="-m -m"

Usage:
    tools/compare-opts.py [options] <package-dir> [-- extra tinygo args]

The package directory is a Go package (for example ./mypkg). If it is
a main package it is built with `go build` / `tinygo build`. Otherwise its
test binary is built (`go test -c` / `tinygo test -c`) and results in _test.go
files are ignored. Use --tinygo-main to build a main package with tinygo
instead of the tests.

Results are compared per file:line, since the two compilers report different
columns. Each category has three sections: reported by both, only by Go and
only by TinyGo. Use --only bounds or --only allocs to select one category.

Known differences that are not missed optimizations:
- Go reports "escapes to heap" for constants converted to an interface, such
  as an int based error constant. These use static data and do not allocate.
  TinyGo stores pointer sized values directly in the interface.
- TinyGo checks len <= cap in make([]T, len, cap) inline and reports it as
  IsSliceInBounds. Go does this check inside runtime.makeslice.
"""
import argparse
import os
import re
import subprocess
import sys
import tempfile

CATEGORIES = ("bounds", "allocs")


def run(cmd, cwd):
    return subprocess.run(cmd, cwd=cwd, capture_output=True, text=True)


def norm(path, cwd):
    return os.path.realpath(path if os.path.isabs(path) else os.path.join(cwd, path))


def is_main(pkgdir):
    p = run(["go", "list", "-f", "{{.Name}}", "."], pkgdir)
    if p.returncode != 0:
        sys.exit("go list failed: " + p.stderr)
    return p.stdout.strip() == "main"


def add(res, cat, path, line, text):
    res[cat].setdefault((path, int(line)), set()).add(text)


def go_results(pkgdir, extra, test):
    """Run go once and collect bounds checks and heap escapes."""
    # Go replays cached compiler output, so no -a is needed.
    cmd = ["go", "test", "-c", "-o", os.devnull] if test else ["go", "build", "-o", os.devnull]
    cmd += ["-gcflags=-d=ssa/check_bce/debug=1 -m -m"] + extra + ["."]
    p = run(cmd, pkgdir)
    res = {c: {} for c in CATEGORIES}
    for line in (p.stdout + p.stderr).splitlines():
        m = re.match(r"^(.+?\.go):(\d+):(\d+): Found (\w+)", line)
        if m:
            add(res, "bounds", norm(m.group(1), pkgdir), m.group(2), m.group(4))
            continue
        # "x escapes to heap" and "moved to heap: x". With -m -m the first
        # form is repeated with a trailing colon and flow details.
        m = re.match(r"^(.+?\.go):(\d+):(\d+): (.+?) escapes to heap\b", line)
        if not m:
            m = re.match(r"^(.+?\.go):(\d+):(\d+): (moved to heap: \S+)", line)
            text = m and m.group(4)
        else:
            text = m.group(4) + " escapes"
        if m:
            add(res, "allocs", norm(m.group(1), pkgdir), m.group(2), text)
        elif p.returncode != 0 and line.strip() and not line.startswith("#"):
            print("go build:", line, file=sys.stderr)
    if p.returncode != 0:
        sys.exit("go build failed")
    return res


def tinygo_results(pkgdir, main, tinygo, extra, test):
    """Run tinygo once and collect bounds checks and heap allocations."""
    # The alloc regexp matches function names, and inlining makes those
    # unreliable, so match everything and filter by file below.
    regex = re.escape(pkgdir) + r"/[^/]*\.go$"
    with tempfile.TemporaryDirectory() as tmp:
        cmd = [tinygo, "test", "-c"] if test else [tinygo, "build"]
        cmd += ["-o", os.path.join(tmp, "out"), "-print-bounds", regex,
                "-print-allocs", "."] + extra + ["."]
        # tinygo must run inside the module of the main package.
        p = run(cmd, main)
    res = {c: {} for c in CATEGORIES}
    for line in p.stderr.splitlines():
        m = re.match(r"^(.+?\.go):(\d+):(\d+): Found (\w+)$", line)
        if m:
            add(res, "bounds", norm(m.group(1), "/"), m.group(2), m.group(4))
            continue
        m = re.match(r"^(.+?\.go):(\d+):(\d+): object allocated on the heap: (.*)$", line)
        if m:
            add(res, "allocs", norm(m.group(1), "/"), m.group(2), m.group(4))
        elif p.returncode != 0 and line.strip():
            print("tinygo:", line, file=sys.stderr)
    if p.returncode != 0:
        sys.exit("tinygo build failed")
    return res


def show(title, keys, go, tg, cwd, cat):
    print(f"\n== {title} ({len(keys)}) ==")
    cache = {}
    for f, l in sorted(keys):
        if f not in cache:
            try:
                cache[f] = open(f).read().splitlines()
            except OSError:
                cache[f] = []
        src = cache[f][l - 1].strip() if 0 < l <= len(cache[f]) else ""
        g, t = go.get((f, l)) or set(), tg.get((f, l)) or set()
        if cat == "bounds":
            note = ",".join(sorted(g | t))
        else:
            note = "; ".join(
                [f"go: {', '.join(sorted(g))}"] * bool(g) + [f"tinygo: {', '.join(sorted(t))}"] * bool(t))
        rel = os.path.relpath(f, cwd)
        if rel.startswith(".."):
            rel = f
        print(f"{rel}:{l}: [{note}] {src}")


def own_files(res, pkgdir):
    """Keep results in the package's own non-test, non-generated files."""
    def own(path):
        name = os.path.basename(path)
        return (os.path.dirname(path) == pkgdir and not name.endswith("_test.go")
                and not name.startswith("_"))
    return {k: v for k, v in res.items() if own(k[0])}


def compare(cat, go, tg, cwd):
    noun = {"bounds": "check", "allocs": "heap allocation"}[cat]
    print(f"\n######## {cat} ########")
    both = go.keys() & tg.keys()
    only_go = go.keys() - tg.keys()
    only_tg = tg.keys() - go.keys()
    show(f"both report a {noun}", both, go, tg, cwd, cat)
    show(f"only Go reports a {noun}", only_go, go, tg, cwd, cat)
    show(f"only TinyGo reports a {noun}", only_tg, go, tg, cwd, cat)
    print(f"\nsummary {cat}: go={len(go)} tinygo={len(tg)} both={len(both)} "
          f"only-go={len(only_go)} only-tinygo={len(only_tg)}")


def main():
    argv = sys.argv[1:]
    extra_tg = []
    if "--" in argv:
        i = argv.index("--")
        argv, extra_tg = argv[:i], argv[i + 1:]
    ap = argparse.ArgumentParser(description=__doc__.split("\n\n")[0])
    ap.add_argument("pkgdir")
    ap.add_argument("--tinygo", default="tinygo", help="tinygo binary (default: tinygo)")
    ap.add_argument("--tinygo-main", help="directory of the main package to build with tinygo (default: pkgdir)")
    ap.add_argument("--tags", default="",
                    help='build tags for both compilers, space or comma separated (e.g. "purego noasm")')
    ap.add_argument("--go-args", default="", help="extra args for go build, space separated")
    ap.add_argument("--only", choices=CATEGORIES, help="report only one category")
    args = ap.parse_args(argv)

    pkgdir = os.path.realpath(args.pkgdir)
    main_pkg = os.path.realpath(args.tinygo_main or pkgdir)
    if os.sep in args.tinygo:
        args.tinygo = os.path.realpath(args.tinygo)
    # A non-main package is built as its test binary (unless --tinygo-main
    # names a main package for tinygo). Results in _test.go files are ignored.
    test = not is_main(pkgdir)
    tg_test = test and not args.tinygo_main
    tags = [t for t in re.split(r"[ ,]+", args.tags) if t]
    go_extra = args.go_args.split()
    if tags:
        go_extra.append("-tags=" + ",".join(tags))
        extra_tg = ["-tags", " ".join(tags)] + extra_tg
    go = go_results(pkgdir, go_extra, test)
    tg = tinygo_results(pkgdir, main_pkg, args.tinygo, extra_tg, tg_test)

    cwd = os.getcwd()
    for cat in CATEGORIES:
        if args.only and cat != args.only:
            continue
        g, t = (own_files(r[cat], pkgdir) for r in (go, tg))
        compare(cat, g, t, cwd)


if __name__ == "__main__":
    main()
