package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strconv"
	"strings"
)

type ruleSet struct {
	packages map[string]*packagePlan
	subtrees []subtreeRule
	allow    []allowRule
	includes []string
	err      error
}

type subtreeRule struct {
	prefix string
	line   int
}

type allowRule struct {
	packages []string
	line     int
}

func ruleLine() int {
	_, _, line, _ := runtime.Caller(2)
	return line
}

func (r *ruleSet) fail(line int, format string, args ...any) {
	if r.err == nil {
		r.err = fmt.Errorf("rules.go:%d: %s", line, fmt.Sprintf(format, args...))
	}
}

func (r *ruleSet) checkPackage(pkg string, line int) {
	if pkg == "" || strings.ContainsAny(pkg, " \t\r\n") || strings.Contains(pkg, "...") {
		r.fail(line, "invalid package name %q", pkg)
	}
}

func (r *ruleSet) packageRule(pkg string, line int) *packagePlan {
	r.checkPackage(pkg, line)
	if r.packages == nil {
		r.packages = make(map[string]*packagePlan)
	}
	p := r.packages[pkg]
	if p == nil {
		p = &packagePlan{name: pkg}
		r.packages[pkg] = p
	}
	p.lines = append(p.lines, line)
	return p
}

// skipPackage accepts pkg/... to skip a package and its descendants.
func (r *ruleSet) skipPackage(pkg string) {
	line := ruleLine()
	if prefix, subtree := strings.CutSuffix(pkg, "/..."); subtree {
		r.checkPackage(prefix, line)
		r.subtrees = append(r.subtrees, subtreeRule{prefix, line})
		return
	}
	r.packageRule(pkg, line).skipped = true
}

func (r *ruleSet) onlyPackages(packages ...string) {
	line := ruleLine()
	if len(packages) == 0 {
		r.fail(line, "expected at least one package")
	}
	for _, pkg := range packages {
		r.checkPackage(pkg, line)
	}
	r.allow = append(r.allow, allowRule{packages, line})
}

func (r *ruleSet) includePackage(pkg string) {
	r.packageRule(pkg, ruleLine())
	r.includes = append(r.includes, pkg)
}

func (r *ruleSet) checkPattern(pattern string, line int) {
	if pattern == "" {
		r.fail(line, "expected a test pattern")
	} else if _, err := regexp.Compile(pattern); err != nil {
		r.fail(line, "%v", err)
	}
}

func (r *ruleSet) skipTest(pkg, pattern string) {
	line := ruleLine()
	p := r.packageRule(pkg, line)
	r.checkPattern(pattern, line)
	p.skips = append(p.skips, pattern)
}

func (r *ruleSet) runTest(pkg, pattern string) {
	line := ruleLine()
	p := r.packageRule(pkg, line)
	r.checkPattern(pattern, line)
	if p.run != "" && p.run != pattern {
		r.fail(line, "%s has conflicting test filters %q and %q", pkg, p.run, pattern)
	}
	p.run = pattern
}

func (r *ruleSet) stackSizeMB(pkg string, size int) {
	line := ruleLine()
	p := r.packageRule(pkg, line)
	if size <= 0 {
		r.fail(line, "%s has invalid stack size %d MB", pkg, size)
	}
	if p.stackMB != 0 && p.stackMB != size {
		r.fail(line, "%s has conflicting stack sizes %d MB and %d MB", pkg, p.stackMB, size)
	}
	p.stackMB = size
}

func (r *ruleSet) short(pkg string) {
	r.packageRule(pkg, ruleLine()).short = true
}

func (r ruleSet) planPackage(name string) packagePlan {
	p := packagePlan{name: name}
	if configured := r.packages[name]; configured != nil {
		p = *configured
		p.lines = slices.Clone(p.lines)
	}
	for _, subtree := range r.subtrees {
		if name == subtree.prefix || strings.HasPrefix(name, subtree.prefix+"/") {
			p.skipped = true
			p.lines = append(p.lines, subtree.line)
		}
	}
	for _, allowlist := range r.allow {
		if !slices.Contains(allowlist.packages, name) {
			p.skipped = true
			p.lines = append(p.lines, allowlist.line)
		}
	}
	slices.Sort(p.lines)
	return p
}

type packagePlan struct {
	name    string
	skipped bool
	stackMB int
	short   bool
	skips   []string
	run     string
	lines   []int
}

func (p packagePlan) flags() []string {
	var flags []string
	if p.short {
		flags = append(flags, "-short")
	}
	if p.stackMB != 0 {
		flags = append(flags, fmt.Sprintf("-stack-size=%dMB", p.stackMB))
	}
	if len(p.skips) != 0 {
		flags = append(flags, "-skip="+strings.Join(p.skips, "|"))
	}
	if p.run != "" {
		flags = append(flags, "-run="+p.run)
	}
	return flags
}

func output(cmd *exec.Cmd) ([]byte, error) {
	cmd.Stderr = os.Stderr
	data, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cmd, err)
	}
	return data, nil
}

type targetInfo struct {
	GOOS      string
	GOARCH    string
	GOROOT    string
	BuildTags []string `json:"build_tags"`
}

func configFlags(args []string) []string {
	var result []string
	for i := 0; i < len(args); i++ {
		name, _, hasValue := strings.Cut(strings.TrimLeft(args[i], "-"), "=")
		switch name {
		case "target", "tags", "gc", "scheduler", "panic-unwind", "serial":
			result = append(result, args[i])
			if !hasValue && i+1 < len(args) {
				i++
				result = append(result, args[i])
			}
		}
	}
	return result
}

func discover(goTool, tinygo string, env, args []string) (targetInfo, []string, error) {
	var target targetInfo
	config := configFlags(args)
	info := exec.Command(tinygo, append([]string{"info", "-json"}, config...)...)
	info.Env = env
	data, err := output(info)
	if err != nil {
		return target, nil, err
	}
	if err := json.Unmarshal(data, &target); err != nil {
		return target, nil, fmt.Errorf("decode TinyGo target: %w", err)
	}

	if target.GOROOT == "" || target.GOOS == "" || target.GOARCH == "" {
		return target, nil, errors.New("TinyGo target info is missing GOROOT, GOOS or GOARCH")
	}

	// The std wildcard does not descend into symlinked TinyGo packages.
	stockList := exec.Command(goTool, "list", "std")
	stockList.Env = append(slices.Clone(env), "GOOS="+target.GOOS, "GOARCH="+target.GOARCH, "CGO_ENABLED=0")
	data, err = output(stockList)
	if err != nil {
		return target, nil, err
	}
	names := strings.Fields(string(data))
	stdRoots := make(map[string]bool)
	for _, name := range names {
		if root, _, _ := strings.Cut(name, "/"); root != "vendor" {
			stdRoots[root] = true
		}
	}
	tinygoEnv := exec.Command(tinygo, "env", "TINYGOROOT")
	tinygoEnv.Env = env
	tinygoRoot, err := output(tinygoEnv)
	if err != nil {
		return target, nil, err
	}
	tinygoSrc := filepath.Join(strings.TrimSpace(string(tinygoRoot)), "src")
	err = filepath.WalkDir(tinygoSrc, func(path string, d fs.DirEntry, err error) error {
		if err != nil || path == tinygoSrc {
			return err
		}
		base := d.Name()
		if d.IsDir() {
			if (filepath.Dir(path) == tinygoSrc && !stdRoots[base]) || base == "testdata" || base[0] == '.' || base[0] == '_' {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(base, "_test.go") && base[0] != '.' && base[0] != '_' {
			rel, err := filepath.Rel(tinygoSrc, filepath.Dir(path))
			if err != nil {
				return err
			}
			names = append(names, filepath.ToSlash(rel))
		}
		return nil
	})
	if err != nil {
		return target, nil, err
	}
	slices.Sort(names)
	var candidates []string
	for _, name := range slices.Compact(names) {
		info, err := os.Stat(filepath.Join(target.GOROOT, "src", filepath.FromSlash(name)))
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return target, nil, err
		}
		if info.IsDir() {
			candidates = append(candidates, name)
		}
	}
	if len(candidates) == 0 {
		return target, nil, errors.New("no standard library packages found in TinyGo GOROOT")
	}
	list := []string{"list", "-e", "-json=ImportPath,TestGoFiles,XTestGoFiles", "-tags", strings.Join(target.BuildTags, " ")}
	effectiveList := exec.Command(goTool, append(list, candidates...)...)
	effectiveList.Dir = filepath.Join(target.GOROOT, "src")
	effectiveList.Env = append(stockList.Env, "GOROOT="+target.GOROOT, "CGO_ENABLED=1", "GOWORK=off")
	data, err = output(effectiveList)
	if err != nil {
		return target, nil, err
	}
	var packages []string
	decoder := json.NewDecoder(bytes.NewReader(data))
	for {
		var pkg struct {
			ImportPath                string
			TestGoFiles, XTestGoFiles []string
		}
		if err := decoder.Decode(&pkg); errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return target, nil, fmt.Errorf("decode TinyGo package list: %w", err)
		}
		if len(pkg.TestGoFiles)+len(pkg.XTestGoFiles) > 0 {
			packages = append(packages, pkg.ImportPath)
		}
	}
	if len(packages) == 0 {
		return target, nil, errors.New("no standard library test packages discovered")
	}
	return target, packages, nil
}

const (
	modeFull = "full"
	modeFast = "fast"
	modeSlow = "slow"
)

type options struct {
	mode, goos, goarch, match           string
	plan, packages, bench, json, report bool
	shard, shardCount                   int
}

func (opts options) environment() []string {
	env := os.Environ()
	if opts.goos != "" {
		env = append(env, "GOOS="+opts.goos)
	}
	if opts.goarch != "" {
		env = append(env, "GOARCH="+opts.goarch)
	}
	return env
}

func planSuite(packages []string, rules ruleSet, opts options) ([]packagePlan, error) {
	if rules.err != nil {
		return nil, rules.err
	}
	packages = append(slices.Clone(packages), rules.includes...)
	slices.Sort(packages)
	packages = slices.Compact(packages)

	var match *regexp.Regexp
	if opts.match != "" {
		var err error
		match, err = regexp.Compile(opts.match)
		if err != nil {
			return nil, fmt.Errorf("invalid package filter: %w", err)
		}
	}
	var plans []packagePlan
	runnable := 0
	for _, name := range packages {
		if match != nil && !match.MatchString(name) {
			continue
		}
		plan := rules.planPackage(name)
		if opts.shardCount > 1 {
			if plan.skipped {
				continue
			}
			shard := runnable%opts.shardCount + 1
			runnable++
			if shard != opts.shard {
				continue
			}
		}
		plans = append(plans, plan)
	}
	return plans, nil
}

func report(goTool, tinygo string) error {
	profiles := []struct {
		label, goos, target string
	}{
		{"darwin", "darwin", ""},
		{"linux", "linux", ""},
		{"wasi", "", "wasip1"},
		{"windows", "windows", ""},
	}
	platforms := make(map[string][]string)
	for _, profile := range profiles {
		opts := options{mode: modeFull, goos: profile.goos, goarch: "amd64"}
		var args []string
		if profile.target != "" {
			args = append(args, "-target="+profile.target)
		}
		target, packages, err := discover(goTool, tinygo, opts.environment(), args)
		if err != nil {
			return fmt.Errorf("%s: %w", profile.label, err)
		}
		plans, err := planSuite(packages, rulesFor(target, opts, runtime.GOOS), opts)
		if err != nil {
			return fmt.Errorf("%s: %w", profile.label, err)
		}
		for _, plan := range plans {
			if !plan.skipped && plan.run == "" && !strings.HasPrefix(plan.name, "./") {
				platforms[plan.name] = append(platforms[plan.name], profile.label)
			}
		}
	}
	names := make([]string, 0, len(platforms))
	for name := range platforms {
		names = append(names, name)
	}
	slices.Sort(names)
	for _, name := range names {
		fmt.Println(name, strings.Join(platforms[name], " "))
	}
	return nil
}

func run(goTool, tinygo string, extra []string, opts options) error {
	goTool, err := exec.LookPath(goTool)
	if err != nil {
		return err
	}
	goTool, err = filepath.Abs(goTool)
	if err != nil {
		return err
	}
	if opts.report {
		return report(goTool, tinygo)
	}
	env := opts.environment()
	target, packages, err := discover(goTool, tinygo, env, extra)
	if err != nil {
		return err
	}
	plans, err := planSuite(packages, rulesFor(target, opts, runtime.GOOS), opts)
	if err != nil {
		return err
	}
	var defaults []string
	var commands [][]string
	selected := 0
	for _, plan := range plans {
		name := plan.name
		if opts.plan && !opts.packages && len(plan.lines) != 0 {
			fmt.Printf("# %s: rules.go %v", name, plan.lines)
			if plan.skipped {
				fmt.Print(" (skipped)")
			}
			fmt.Println()
		}
		if plan.skipped {
			continue
		}
		selected++
		if opts.packages {
			if !strings.HasPrefix(name, "./") {
				fmt.Println(name)
			}
			continue
		}
		flags := plan.flags()
		if len(flags) != 0 || opts.json {
			commands = append(commands, append(flags, name))
		} else {
			defaults = append(defaults, name)
		}
	}
	if len(defaults) != 0 {
		commands = append([][]string{defaults}, commands...)
	}
	if selected == 0 {
		return errors.New("no standard library test packages selected")
	}
	var failures []error
	for _, command := range commands {
		args := append(append([]string{"test"}, extra...), command...)
		if opts.bench {
			args = append([]string{"test", "-bench=."}, args[1:]...)
		}
		executable := tinygo
		if opts.json {
			tinygoPath, err := exec.LookPath(tinygo)
			if err != nil {
				return err
			}
			pkg := command[len(command)-1]
			args = append([]string{"tool", "test2json", "-t", "-p", pkg, tinygoPath, "test", "-v"}, args[1:]...)
			executable = goTool
		}
		if opts.plan {
			fmt.Printf("%q", executable)
			for _, arg := range args {
				fmt.Printf(" %q", arg)
			}
			fmt.Println()
			continue
		}
		cmd := exec.Command(executable, args...)
		cmd.Env = env
		cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
		if err := cmd.Run(); err != nil {
			err = fmt.Errorf("%s: %w", cmd, err)
			if !opts.json {
				return err
			}
			failures = append(failures, err)
		}
	}
	return errors.Join(failures...)
}

func main() {
	var opts options
	var shard string
	flag.BoolVar(&opts.plan, "plan", false, "print matching rule lines and test commands")
	flag.BoolVar(&opts.packages, "packages", false, "print selected standard library packages")
	flag.BoolVar(&opts.bench, "bench", false, "run benchmarks")
	flag.BoolVar(&opts.json, "json", false, "emit test2json events")
	flag.BoolVar(&opts.report, "report", false, "report tested packages across platforms")
	flag.StringVar(&opts.match, "match", "", "restrict selected packages by regular expression")
	flag.StringVar(&shard, "shard", "1/1", "run one shard of selected packages (index/count, starting at 1)")
	flag.StringVar(&opts.mode, "mode", modeFull, "suite mode: full, fast or slow")
	flag.StringVar(&opts.goos, "goos", "", "set GOOS for TinyGo")
	flag.StringVar(&opts.goarch, "goarch", "", "set GOARCH for TinyGo")
	flag.Parse()
	if flag.NArg() < 2 {
		fmt.Fprintln(os.Stderr, "usage: stdlibtest [-plan] GO TINYGO [TINYGO FLAGS...]")
		os.Exit(2)
	}
	if opts.mode != modeFull && opts.mode != modeFast && opts.mode != modeSlow {
		fmt.Fprintln(os.Stderr, "invalid suite mode:", opts.mode)
		os.Exit(2)
	}
	index, count, ok := strings.Cut(shard, "/")
	var indexErr, countErr error
	opts.shard, indexErr = strconv.Atoi(index)
	opts.shardCount, countErr = strconv.Atoi(count)
	if !ok || indexErr != nil || countErr != nil || opts.shard < 1 || opts.shard > opts.shardCount {
		fmt.Fprintln(os.Stderr, "invalid shard (expected 1 <= index <= count):", shard)
		os.Exit(2)
	}
	if opts.report && opts.shardCount > 1 {
		fmt.Fprintln(os.Stderr, "-report cannot be combined with sharding")
		os.Exit(2)
	}
	if err := run(flag.Arg(0), flag.Arg(1), flag.Args()[2:], opts); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
