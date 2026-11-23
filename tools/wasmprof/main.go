// Command wasmprof converts a wasmtime guest profile to a pprof profile with
// file:line and inline frames. See README.md.
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
)

type stringList []string

func (s *stringList) String() string     { return strings.Join(*s, " ") }
func (s *stringList) Set(v string) error { *s = append(*s, v); return nil }

func main() {
	out := flag.String("o", "cpu.pprof", "output pprof file")
	wasmtime := flag.String("wasmtime", "wasmtime", "wasmtime binary; must be the one that produced the profile")
	var compileArgs stringList
	flag.Var(&compileArgs, "compile-arg", "extra argument for `wasmtime compile`, e.g. -O or -W flags given to `wasmtime run` (repeatable)")
	showStats := flag.Bool("stats", false, "print how sample positions were resolved")
	force := flag.Bool("force", false, "write the profile even if function addresses do not match the recompiled module")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "usage: %s [flags] prog.wasm guest-profile.json\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() != 2 {
		flag.Usage()
		os.Exit(2)
	}
	if err := run(flag.Arg(0), flag.Arg(1), *out, *wasmtime, compileArgs, *force, *showStats); err != nil {
		fmt.Fprintln(os.Stderr, "wasmprof:", err)
		os.Exit(1)
	}
}

func run(wasmPath, profPath, outPath, wasmtime string, compileArgs []string, force, showStats bool) error {
	gp, err := readGuestProfile(profPath)
	if err != nil {
		return err
	}
	wf, err := readWasm(wasmPath)
	if err != nil {
		return err
	}
	am, err := compileAddrMap(wasmtime, wasmPath, compileArgs)
	if err != nil {
		return err
	}
	if err := am.check(gp); err != nil {
		if !force {
			return fmt.Errorf("%w\nthe recompiled module does not match the profiled code; use the same wasmtime binary and pass the run's codegen flags with -compile-arg", err)
		}
		fmt.Fprintln(os.Stderr, "wasmprof: warning:", err)
	}
	if !slices.ContainsFunc(wf.modules, func(m *wasmModule) bool { return m.dw != nil }) {
		fmt.Fprintln(os.Stderr, "wasmprof: warning: no DWARF in any module, profile will only have function names")
	}

	var stats Stats
	if showStats {
		stats = Stats{}
	}
	p := buildProfile(gp, am, wf, wasmPath, stats)
	if showStats {
		printStats(stats)
	}
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	if err := p.write(f); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

func printStats(s Stats) {
	total := s["total"]
	if total == 0 {
		return
	}
	keys := make([]string, 0, len(s))
	for k := range s {
		if k != "total" {
			keys = append(keys, k)
		}
	}
	sort.Slice(keys, func(i, j int) bool { return s[keys[i]] > s[keys[j]] })
	fmt.Fprintf(os.Stderr, "leaf position of %d samples:\n", total)
	for _, k := range keys {
		fmt.Fprintf(os.Stderr, "  %6.1f%%  %s\n", 100*float64(s[k])/float64(total), k)
	}
}

// Guest profile (Firefox Profiler "processed" format) subset.

type guestProfile struct {
	Meta struct {
		Interval  float64 `json:"interval"` // milliseconds
		StartTime float64 `json:"startTime"`
		Product   string  `json:"product"`
	} `json:"meta"`
	Libs    []json.RawMessage `json:"libs"`
	Threads []guestThread     `json:"threads"`
	Shared  struct {
		StringArray []string `json:"stringArray"`
	} `json:"shared"`
}

type guestThread struct {
	Name    string `json:"name"`
	Samples struct {
		Stack      []*int    `json:"stack"`
		Weight     []float64 `json:"weight"`
		TimeDeltas []float64 `json:"timeDeltas"`
	} `json:"samples"`
	StackTable struct {
		Prefix []*int `json:"prefix"`
		Frame  []int  `json:"frame"`
	} `json:"stackTable"`
	FrameTable struct {
		Address []int64 `json:"address"`
		Func    []int   `json:"func"`
	} `json:"frameTable"`
	FuncTable struct {
		Name []int `json:"name"`
	} `json:"funcTable"`
	NativeSymbols struct {
		Address []int64 `json:"address"`
		Name    []int   `json:"name"`
	} `json:"nativeSymbols"`
	StringArray []string `json:"stringArray"`
}

func readGuestProfile(path string) (*guestProfile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	gp := &guestProfile{}
	if err := json.Unmarshal(data, gp); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	if len(gp.Threads) == 0 {
		return nil, fmt.Errorf("%s: no threads in profile", path)
	}
	for i := range gp.Threads {
		if gp.Threads[i].StringArray == nil {
			gp.Threads[i].StringArray = gp.Shared.StringArray
		}
	}
	return gp, nil
}

func (t *guestThread) str(i int) string {
	if i < 0 || i >= len(t.StringArray) {
		return ""
	}
	return t.StringArray[i]
}

// Native code address map, from `wasmtime objdump --addrmap`.

type addrMap struct {
	// Sorted by native address. wasm is a wasm file offset, or -1 where the
	// native code has no wasm location.
	entries []addrEntry
	funcs   map[int64]string // native start address -> function name
}

type addrEntry struct {
	native int64
	wasm   int64
}

var (
	reFuncHeader = regexp.MustCompile(`^([0-9a-f]+) (.*):$`)
	reInsn       = regexp.MustCompile(`^\s+([0-9a-f]+):`)
	reAddrmap    = regexp.MustCompile(`addrmap: 0x([0-9a-f]+)`)
	reFuncName   = regexp.MustCompile(`^wasm\[\d+\]::function\[\d+\](?:::(.*))?$`)
)

func compileAddrMap(wasmtime, wasmPath string, compileArgs []string) (*addrMap, error) {
	dir, err := os.MkdirTemp("", "wasmprof")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)
	cwasm := filepath.Join(dir, "module.cwasm")
	args := append([]string{"compile", "-W", "epoch-interruption=y"}, compileArgs...)
	args = append(args, "-o", cwasm, wasmPath)
	cmd := exec.Command(wasmtime, args...)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("%s %s: %w", wasmtime, strings.Join(args, " "), err)
	}
	cmd = exec.Command(wasmtime, "objdump", "--addresses", "--addrmap", "--color", "never", cwasm)
	cmd.Stderr = os.Stderr
	text, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("%s objdump: %w", wasmtime, err)
	}
	return parseObjdump(text)
}

func parseObjdump(text []byte) (*addrMap, error) {
	am := &addrMap{funcs: map[int64]string{}}
	var last int64 = -1
	sc := bufio.NewScanner(bytes.NewReader(text))
	sc.Buffer(nil, 1<<20)
	for sc.Scan() {
		line := sc.Text()
		if m := reFuncHeader.FindStringSubmatch(line); m != nil {
			addr, _ := strconv.ParseInt(m[1], 16, 64)
			name := m[2]
			if fm := reFuncName.FindStringSubmatch(name); fm != nil {
				name = fm[1]
			}
			am.funcs[addr] = name
			// No wasm location until the first addrmap entry of this function.
			am.entries = append(am.entries, addrEntry{addr, -1})
			last = addr
			continue
		}
		if m := reInsn.FindStringSubmatch(line); m != nil {
			last, _ = strconv.ParseInt(m[1], 16, 64)
			continue
		}
		if m := reAddrmap.FindStringSubmatch(line); m != nil && last >= 0 {
			pos, _ := strconv.ParseInt(m[1], 16, 64)
			am.entries = append(am.entries, addrEntry{last, pos})
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	if len(am.funcs) == 0 {
		return nil, errors.New("wasmtime objdump printed no functions")
	}
	sort.SliceStable(am.entries, func(i, j int) bool { return am.entries[i].native < am.entries[j].native })
	return am, nil
}

// lookup returns the wasm file offset for a native code offset, or -1.
func (am *addrMap) lookup(native int64) int64 {
	i := sort.Search(len(am.entries), func(i int) bool { return am.entries[i].native > native }) - 1
	if i < 0 {
		return -1
	}
	return am.entries[i].wasm
}

// check verifies that every function the profile knows about is at the same
// native address in the recompiled module.
func (am *addrMap) check(gp *guestProfile) error {
	bad, total := 0, 0
	var first string
	for i := range gp.Threads {
		t := &gp.Threads[i]
		for j, addr := range t.NativeSymbols.Address {
			name := t.str(t.NativeSymbols.Name[j])
			total++
			// wasmtime objdump truncates long names, so compare as a prefix.
			if got, ok := am.funcs[addr]; !ok || got != "" && !strings.HasPrefix(name, got) {
				bad++
				if first == "" {
					first = fmt.Sprintf("%s at %#x in profile, recompiled module has %q there", name, addr, got)
				}
			}
		}
	}
	if bad > 0 {
		return fmt.Errorf("%d of %d profiled functions are not where the profile says (first: %s)", bad, total, first)
	}
	return nil
}
