package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/tinygo-org/tinygo/builder"
	"github.com/tinygo-org/tinygo/compileopts"
	"github.com/tinygo-org/tinygo/compiler"
	"github.com/tinygo-org/tinygo/fix"
	"github.com/tinygo-org/tinygo/goenv"
	"github.com/tinygo-org/tinygo/loader"
	"golang.org/x/tools/go/buildutil"
)

const usageFix = `usage: tinygo fix [flags] [packages]

Apply Go fix analyzers using TinyGo's package loader, including package tests.
Use -diff to print changes without writing files. A nonempty diff exits with 1.
Use -tests=false to exclude test files from analysis.
Use -target and -tags to select one target configuration.
Only named packages are changed, not their dependencies or host Go sources.

Run tinygo fix help for analyzer names or tinygo fix help <name> for details.
The -fixtool and -toolexec flags are not supported.
`

func runFix(args []string, stdout, stderr io.Writer) int {
	set := flag.NewFlagSet("fix", flag.ContinueOnError)
	set.SetOutput(stderr)
	set.Usage = func() {
		fmt.Fprint(stderr, usageFix)
		set.PrintDefaults()
	}
	diff := set.Bool("diff", false, "print a unified diff without writing files")
	includeTests := set.Bool("tests", true, "include test files when analyzing packages")
	options := &compileopts.Options{
		GOOS: goenv.Get("GOOS"), GOARCH: goenv.Get("GOARCH"),
		GOARM: goenv.Get("GOARM"), GOMIPS: goenv.Get("GOMIPS"),
		GoCompatibility: true, Opt: "z", PanicStrategy: "print",
	}
	set.StringVar(&options.Target, "target", "", "chip/board name or JSON target specification file")
	set.StringVar(&options.GC, "gc", "", "garbage collector")
	set.StringVar(&options.Scheduler, "scheduler", "", "goroutine scheduler")
	set.StringVar(&options.Serial, "serial", "", "serial output")
	set.StringVar(&options.PanicUnwind, "panic-unwind", "", "panic unwind strategy")
	set.StringVar(&options.LLVMFeatures, "llvm-features", "", "LLVM target features")
	set.StringVar(&options.Directory, "C", "", "change to directory before loading packages")
	var tags buildutil.TagsFlag
	set.Var(&tags, "tags", "build tags")
	analyzerFlags := fix.RegisterFlags(set)
	if err := set.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	patterns := set.Args()
	if len(patterns) > 0 && patterns[0] == "help" {
		for _, a := range fix.Analyzers() {
			if len(patterns) == 1 {
				fmt.Fprintf(stdout, "%-18s %s\n", a.Name, strings.SplitN(a.Doc, "\n", 2)[0])
			} else if len(patterns) == 2 && patterns[1] == a.Name {
				fmt.Fprintln(stdout, a.Doc)
				a.Flags.SetOutput(stdout)
				a.Flags.PrintDefaults()
				return 0
			}
		}
		if len(patterns) == 1 {
			return 0
		}
		fmt.Fprintln(stderr, "unknown analyzer or invalid help arguments")
		return 2
	}
	options.Tags = tags
	if options.Directory != "" {
		if err := os.Chdir(options.Directory); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		options.Directory = ""
	}
	if err := options.Verify(); err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	config, err := builder.NewConfig(options)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	machine, err := compiler.NewTargetMachine(&compiler.Config{
		Triple: config.Triple(), CPU: config.CPU(), Features: config.Features(),
		CodeModel: config.CodeModel(), RelocationModel: config.RelocationModel(),
	})
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	defer machine.Dispose()
	sizes := compiler.Sizes(machine)
	changed, err := fix.Run(func(overlay map[string][]byte) (*loader.Analysis, error) {
		return loader.LoadForAnalysis(config, patterns, sizes, overlay, *includeTests)
	}, analyzerFlags.Selected(), *diff, stdout, stderr)
	if err != nil {
		printBuildOutput(err, false)
		return 1
	}
	if *diff && changed {
		return 1
	}
	return 0
}
