package main

import (
	"errors"
	"flag"
	"fmt"
	"go/types"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	"github.com/tinygo-org/tinygo/compiler"
	"github.com/tinygo-org/tinygo/interp"
	"github.com/tinygo-org/tinygo/loader"
)

// commandError is an error type to wrap os/exec.Command errors. This provides
// some more information regarding what went wrong while running a command.
type commandError struct {
	Msg  string
	File string
	Err  error
}

func (e *commandError) Error() string {
	return e.Msg + " " + e.File + ": " + e.Err.Error()
}

// multiError is a list of multiple errors (actually: diagnostics) returned
// during LLVM IR generation.
type multiError struct {
	Errs []error
}

func (e *multiError) Error() string {
	return e.Errs[0].Error()
}

type BuildConfig struct {
	opt           string
	gc            string
	panicStrategy string
	printIR       bool
	dumpSSA       bool
	debug         bool
	printSizes    string
	cFlags        []string
	ldFlags       []string
	wasmAbi       string
	testConfig    compiler.TestConfig
}

// Helper function for Compiler object.
func Compile(pkgName, outpath string, spec *TargetSpec, config *BuildConfig, action func(string) error) error {
	if config.gc == "" && spec.GC != "" {
		config.gc = spec.GC
	}

	root := sourceDir()

	// Merge and adjust CFlags.
	cflags := append([]string{}, config.cFlags...)
	for _, flag := range spec.CFlags {
		cflags = append(cflags, strings.Replace(flag, "{root}", root, -1))
	}

	// Merge and adjust LDFlags.
	ldflags := append([]string{}, config.ldFlags...)
	for _, flag := range spec.LDFlags {
		ldflags = append(ldflags, strings.Replace(flag, "{root}", root, -1))
	}

	goroot := getGoroot()
	if goroot == "" {
		return errors.New("cannot locate $GOROOT, please set it manually")
	}
	tags := spec.BuildTags
	major, minor, err := getGorootVersion(goroot)
	if err != nil {
		return fmt.Errorf("could not read version from GOROOT (%v): %v", goroot, err)
	}
	if major != 1 {
		return fmt.Errorf("expected major version 1, got go%d.%d", major, minor)
	}
	for i := 1; i <= minor; i++ {
		tags = append(tags, fmt.Sprintf("go1.%d", i))
	}
	compilerConfig := compiler.Config{
		Triple:        spec.Triple,
		CPU:           spec.CPU,
		Features:      spec.Features,
		GOOS:          spec.GOOS,
		GOARCH:        spec.GOARCH,
		GC:            config.gc,
		PanicStrategy: config.panicStrategy,
		CFlags:        cflags,
		LDFlags:       ldflags,
		ClangHeaders:  getClangHeaderPath(root),
		Debug:         config.debug,
		DumpSSA:       config.dumpSSA,
		TINYGOROOT:    root,
		GOROOT:        goroot,
		GOPATH:        getGopath(),
		BuildTags:     tags,
		TestConfig:    config.testConfig,
	}
	c, err := compiler.NewCompiler(pkgName, compilerConfig)
	if err != nil {
		return err
	}

	// Compile Go code to IR.
	errs := c.Compile(pkgName)
	if len(errs) != 0 {
		if len(errs) == 1 {
			return errs[0]
		}
		return &multiError{errs}
	}
	if config.printIR {
		fmt.Println("; Generated LLVM IR:")
		fmt.Println(c.IR())
	}
	if err := c.Verify(); err != nil {
		return errors.New("verification error after IR construction")
	}

	err = interp.Run(c.Module(), c.TargetData(), config.dumpSSA)
	if err != nil {
		return err
	}
	if err := c.Verify(); err != nil {
		return errors.New("verification error after interpreting runtime.initAll")
	}

	if spec.GOOS != "darwin" {
		c.ApplyFunctionSections() // -ffunction-sections
	}
	if err := c.Verify(); err != nil {
		return errors.New("verification error after applying function sections")
	}

	// Browsers cannot handle external functions that have type i64 because it
	// cannot be represented exactly in JavaScript (JS only has doubles). To
	// keep functions interoperable, pass int64 types as pointers to
	// stack-allocated values.
	// Use -wasm-abi=generic to disable this behaviour.
	if config.wasmAbi == "js" && strings.HasPrefix(spec.Triple, "wasm") {
		err := c.ExternalInt64AsPtr()
		if err != nil {
			return err
		}
		if err := c.Verify(); err != nil {
			return errors.New("verification error after running the wasm i64 hack")
		}
	}

	// Optimization levels here are roughly the same as Clang, but probably not
	// exactly.
	switch config.opt {
	case "none:", "0":
		err = c.Optimize(0, 0, 0) // -O0
	case "1":
		err = c.Optimize(1, 0, 0) // -O1
	case "2":
		err = c.Optimize(2, 0, 225) // -O2
	case "s":
		err = c.Optimize(2, 1, 225) // -Os
	case "z":
		err = c.Optimize(2, 2, 5) // -Oz, default
	default:
		err = errors.New("unknown optimization level: -opt=" + config.opt)
	}
	if err != nil {
		return err
	}
	if err := c.Verify(); err != nil {
		return errors.New("verification failure after LLVM optimization passes")
	}

	// On the AVR, pointers can point either to flash or to RAM, but we don't
	// know. As a temporary fix, load all global variables in RAM.
	// In the future, there should be a compiler pass that determines which
	// pointers are flash and which are in RAM so that pointers can have a
	// correct address space parameter (address space 1 is for flash).
	if strings.HasPrefix(spec.Triple, "avr") {
		c.NonConstGlobals()
		if err := c.Verify(); err != nil {
			return errors.New("verification error after making all globals non-constant on AVR")
		}
	}

	// Generate output.
	outext := filepath.Ext(outpath)
	switch outext {
	case ".o":
		return c.EmitObject(outpath)
	case ".bc":
		return c.EmitBitcode(outpath)
	case ".ll":
		return c.EmitText(outpath)
	default:
		// Act as a compiler driver.

		// Create a temporary directory for intermediary files.
		dir, err := ioutil.TempDir("", "tinygo")
		if err != nil {
			return err
		}
		defer os.RemoveAll(dir)

		// Write the object file.
		objfile := filepath.Join(dir, "main.o")
		err = c.EmitObject(objfile)
		if err != nil {
			return err
		}

		// Load builtins library from the cache, possibly compiling it on the
		// fly.
		var librt string
		if spec.RTLib == "compiler-rt" {
			librt, err = loadBuiltins(spec.Triple)
			if err != nil {
				return err
			}
		}

		// Prepare link command.
		executable := filepath.Join(dir, "main")
		tmppath := executable // final file
		ldflags := append(ldflags, "-o", executable, objfile, "-L", root)
		if spec.RTLib == "compiler-rt" {
			ldflags = append(ldflags, librt)
		}

		// Compile extra files.
		for i, path := range spec.ExtraFiles {
			abspath := filepath.Join(root, path)
			outpath := filepath.Join(dir, "extra-"+strconv.Itoa(i)+"-"+filepath.Base(path)+".o")
			cmdNames := []string{spec.Compiler}
			if names, ok := commands[spec.Compiler]; ok {
				cmdNames = names
			}
			err := execCommand(cmdNames, append(cflags, "-c", "-o", outpath, abspath)...)
			if err != nil {
				return &commandError{"failed to build", path, err}
			}
			ldflags = append(ldflags, outpath)
		}

		// Compile C files in packages.
		for i, pkg := range c.Packages() {
			for _, file := range pkg.CFiles {
				path := filepath.Join(pkg.Package.Dir, file)
				outpath := filepath.Join(dir, "pkg"+strconv.Itoa(i)+"-"+file+".o")
				cmdNames := []string{spec.Compiler}
				if names, ok := commands[spec.Compiler]; ok {
					cmdNames = names
				}
				err := execCommand(cmdNames, append(cflags, "-c", "-o", outpath, path)...)
				if err != nil {
					return &commandError{"failed to build", path, err}
				}
				ldflags = append(ldflags, outpath)
			}
		}

		// Link the object files together.
		err = Link(spec.Linker, ldflags...)
		if err != nil {
			return &commandError{"failed to link", executable, err}
		}

		if config.printSizes == "short" || config.printSizes == "full" {
			sizes, err := Sizes(executable)
			if err != nil {
				return err
			}
			if config.printSizes == "short" {
				fmt.Printf("   code    data     bss |   flash     ram\n")
				fmt.Printf("%7d %7d %7d | %7d %7d\n", sizes.Code, sizes.Data, sizes.BSS, sizes.Code+sizes.Data, sizes.Data+sizes.BSS)
			} else {
				fmt.Printf("   code  rodata    data     bss |   flash     ram | package\n")
				for _, name := range sizes.SortedPackageNames() {
					pkgSize := sizes.Packages[name]
					fmt.Printf("%7d %7d %7d %7d | %7d %7d | %s\n", pkgSize.Code, pkgSize.ROData, pkgSize.Data, pkgSize.BSS, pkgSize.Flash(), pkgSize.RAM(), name)
				}
				fmt.Printf("%7d %7d %7d %7d | %7d %7d | (sum)\n", sizes.Sum.Code, sizes.Sum.ROData, sizes.Sum.Data, sizes.Sum.BSS, sizes.Sum.Flash(), sizes.Sum.RAM())
				fmt.Printf("%7d       - %7d %7d | %7d %7d | (all)\n", sizes.Code, sizes.Data, sizes.BSS, sizes.Code+sizes.Data, sizes.Data+sizes.BSS)
			}
		}

		// Get an Intel .hex file or .bin file from the .elf file.
		if outext == ".hex" || outext == ".bin" {
			tmppath = filepath.Join(dir, "main"+outext)
			err := Objcopy(executable, tmppath)
			if err != nil {
				return err
			}
		} else if outext == ".uf2" {
			// Get UF2 from the .elf file.
			tmppath = filepath.Join(dir, "main"+outext)
			err := ConvertELFFileToUF2File(executable, tmppath)
			if err != nil {
				return err
			}
		}
		return action(tmppath)
	}
}

func Build(pkgName, outpath, target string, config *BuildConfig) error {
	spec, err := LoadTarget(target)
	if err != nil {
		return err
	}

	return Compile(pkgName, outpath, spec, config, func(tmppath string) error {
		if err := os.Rename(tmppath, outpath); err != nil {
			// Moving failed. Do a file copy.
			inf, err := os.Open(tmppath)
			if err != nil {
				return err
			}
			defer inf.Close()
			outf, err := os.OpenFile(outpath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
			if err != nil {
				return err
			}

			// Copy data to output file.
			_, err = io.Copy(outf, inf)
			if err != nil {
				return err
			}

			// Check whether file writing was successful.
			return outf.Close()
		} else {
			// Move was successful.
			return nil
		}
	})
}

func Test(pkgName, target string, config *BuildConfig) error {
	spec, err := LoadTarget(target)
	if err != nil {
		return err
	}

	spec.BuildTags = append(spec.BuildTags, "test")
	config.testConfig.CompileTestBinary = true
	return Compile(pkgName, ".elf", spec, config, func(tmppath string) error {
		cmd := exec.Command(tmppath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			// Propagate the exit code
			if err, ok := err.(*exec.ExitError); ok {
				if status, ok := err.Sys().(syscall.WaitStatus); ok {
					os.Exit(status.ExitStatus())
				}
				os.Exit(1)
			}
			return &commandError{"failed to run compiled binary", tmppath, err}
		}
		return nil
	})
}

func Flash(pkgName, target, port string, config *BuildConfig) error {
	spec, err := LoadTarget(target)
	if err != nil {
		return err
	}

	// determine the type of file to compile
	var fileExt string

	switch {
	case strings.Contains(spec.Flasher, "{hex}"):
		fileExt = ".hex"
	case strings.Contains(spec.Flasher, "{elf}"):
		fileExt = ".elf"
	case strings.Contains(spec.Flasher, "{bin}"):
		fileExt = ".bin"
	case strings.Contains(spec.Flasher, "{uf2}"):
		fileExt = ".uf2"
	default:
		return errors.New("invalid target file - did you forget the {hex} token in the 'flash' section?")
	}

	return Compile(pkgName, fileExt, spec, config, func(tmppath string) error {
		if spec.Flasher == "" {
			return errors.New("no flash command specified - did you miss a -target flag?")
		}

		// Create the command.
		flashCmd := spec.Flasher
		fileToken := "{" + fileExt[1:] + "}"
		flashCmd = strings.Replace(flashCmd, fileToken, tmppath, -1)
		flashCmd = strings.Replace(flashCmd, "{port}", port, -1)

		// Execute the command.
		cmd := exec.Command("/bin/sh", "-c", flashCmd)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Dir = sourceDir()
		err := cmd.Run()
		if err != nil {
			return &commandError{"failed to flash", tmppath, err}
		}
		return nil
	})
}

// Flash a program on a microcontroller and drop into a GDB shell.
//
// Note: this command is expected to execute just before exiting, as it
// modifies global state.
func FlashGDB(pkgName, target, port string, ocdOutput bool, config *BuildConfig) error {
	spec, err := LoadTarget(target)
	if err != nil {
		return err
	}

	if spec.GDB == "" {
		return errors.New("gdb not configured in the target specification")
	}

	return Compile(pkgName, "", spec, config, func(tmppath string) error {
		if len(spec.OCDDaemon) != 0 {
			// We need a separate debugging daemon for on-chip debugging.
			daemon := exec.Command(spec.OCDDaemon[0], spec.OCDDaemon[1:]...)
			if ocdOutput {
				// Make it clear which output is from the daemon.
				w := &ColorWriter{
					Out:    os.Stderr,
					Prefix: spec.OCDDaemon[0] + ": ",
					Color:  TermColorYellow,
				}
				daemon.Stdout = w
				daemon.Stderr = w
			}
			// Make sure the daemon doesn't receive Ctrl-C that is intended for
			// GDB (to break the currently executing program).
			// https://stackoverflow.com/a/35435038/559350
			daemon.SysProcAttr = &syscall.SysProcAttr{
				Setpgid: true,
				Pgid:    0,
			}
			// Start now, and kill it on exit.
			daemon.Start()
			defer func() {
				daemon.Process.Signal(os.Interrupt)
				// Maybe we should send a .Kill() after x seconds?
				daemon.Wait()
			}()
		}

		// Ignore Ctrl-C, it must be passed on to GDB.
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		go func() {
			for range c {
			}
		}()

		// Construct and execute a gdb command.
		// By default: gdb -ex run <binary>
		// Exit GDB with Ctrl-D.
		params := []string{tmppath}
		for _, cmd := range spec.GDBCmds {
			params = append(params, "-ex", cmd)
		}
		cmd := exec.Command(spec.GDB, params...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			return &commandError{"failed to run gdb with", tmppath, err}
		}
		return nil
	})
}

// Compile and run the given program, directly or in an emulator.
func Run(pkgName, target string, config *BuildConfig) error {
	spec, err := LoadTarget(target)
	if err != nil {
		return err
	}

	return Compile(pkgName, ".elf", spec, config, func(tmppath string) error {
		if len(spec.Emulator) == 0 {
			// Run directly.
			cmd := exec.Command(tmppath)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err := cmd.Run()
			if err != nil {
				if err, ok := err.(*exec.ExitError); ok && err.Exited() {
					// Workaround for QEMU which always exits with an error.
					return nil
				}
				return &commandError{"failed to run compiled binary", tmppath, err}
			}
			return nil
		} else {
			// Run in an emulator.
			args := append(spec.Emulator[1:], tmppath)
			cmd := exec.Command(spec.Emulator[0], args...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err := cmd.Run()
			if err != nil {
				if err, ok := err.(*exec.ExitError); ok && err.Exited() {
					// Workaround for QEMU which always exits with an error.
					return nil
				}
				return &commandError{"failed to run emulator with", tmppath, err}
			}
			return nil
		}
	})
}

func usage() {
	fmt.Fprintln(os.Stderr, "TinyGo is a Go compiler for small places.")
	fmt.Fprintln(os.Stderr, "version:", version)
	fmt.Fprintf(os.Stderr, "usage: %s command [-printir] [-target=<target>] -o <output> <input>\n", os.Args[0])
	fmt.Fprintln(os.Stderr, "\ncommands:")
	fmt.Fprintln(os.Stderr, "  build: compile packages and dependencies")
	fmt.Fprintln(os.Stderr, "  run:   compile and run immediately")
	fmt.Fprintln(os.Stderr, "  test:  test packages")
	fmt.Fprintln(os.Stderr, "  flash: compile and flash to the device")
	fmt.Fprintln(os.Stderr, "  gdb:   run/flash and immediately enter GDB")
	fmt.Fprintln(os.Stderr, "  clean: empty cache directory ("+cacheDir()+")")
	fmt.Fprintln(os.Stderr, "  help:  print this help text")
	fmt.Fprintln(os.Stderr, "\nflags:")
	flag.PrintDefaults()
}

func handleCompilerError(err error) {
	if err != nil {
		if errUnsupported, ok := err.(*interp.Unsupported); ok {
			// hit an unknown/unsupported instruction
			fmt.Fprintln(os.Stderr, "unsupported instruction during init evaluation:")
			errUnsupported.Inst.Dump()
			fmt.Fprintln(os.Stderr)
		} else if errCompiler, ok := err.(types.Error); ok {
			fmt.Fprintln(os.Stderr, errCompiler)
		} else if errLoader, ok := err.(loader.Errors); ok {
			fmt.Fprintln(os.Stderr, "#", errLoader.Pkg.ImportPath)
			for _, err := range errLoader.Errs {
				fmt.Fprintln(os.Stderr, err)
			}
		} else if errMulti, ok := err.(*multiError); ok {
			for _, err := range errMulti.Errs {
				fmt.Fprintln(os.Stderr, err)
			}
		} else {
			fmt.Fprintln(os.Stderr, "error:", err)
		}
		os.Exit(1)
	}
}

func main() {
	outpath := flag.String("o", "", "output filename")
	opt := flag.String("opt", "z", "optimization level: 0, 1, 2, s, z")
	gc := flag.String("gc", "", "garbage collector to use (none, dumb, marksweep)")
	panicStrategy := flag.String("panic", "print", "panic strategy (abort, trap)")
	printIR := flag.Bool("printir", false, "print LLVM IR")
	dumpSSA := flag.Bool("dumpssa", false, "dump internal Go SSA")
	target := flag.String("target", "", "LLVM target")
	printSize := flag.String("size", "", "print sizes (none, short, full)")
	nodebug := flag.Bool("no-debug", false, "disable DWARF debug symbol generation")
	ocdOutput := flag.Bool("ocd-output", false, "print OCD daemon output during debug")
	port := flag.String("port", "/dev/ttyACM0", "flash port")
	cFlags := flag.String("cflags", "", "additional cflags for compiler")
	ldFlags := flag.String("ldflags", "", "additional ldflags for linker")
	wasmAbi := flag.String("wasm-abi", "js", "WebAssembly ABI conventions: js (no i64 params) or generic")

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "No command-line arguments supplied.")
		usage()
		os.Exit(1)
	}
	command := os.Args[1]

	flag.CommandLine.Parse(os.Args[2:])
	config := &BuildConfig{
		opt:           *opt,
		gc:            *gc,
		panicStrategy: *panicStrategy,
		printIR:       *printIR,
		dumpSSA:       *dumpSSA,
		debug:         !*nodebug,
		printSizes:    *printSize,
		wasmAbi:       *wasmAbi,
	}

	if *cFlags != "" {
		config.cFlags = strings.Split(*cFlags, " ")
	}

	if *ldFlags != "" {
		config.ldFlags = strings.Split(*ldFlags, " ")
	}

	if *panicStrategy != "print" && *panicStrategy != "trap" {
		fmt.Fprintln(os.Stderr, "Panic strategy must be either print or trap.")
		usage()
		os.Exit(1)
	}

	os.Setenv("CC", "clang -target="+*target)

	switch command {
	case "build":
		if *outpath == "" {
			fmt.Fprintln(os.Stderr, "No output filename supplied (-o).")
			usage()
			os.Exit(1)
		}
		pkgName := "."
		if flag.NArg() == 1 {
			pkgName = flag.Arg(0)
		} else if flag.NArg() > 1 {
			fmt.Fprintln(os.Stderr, "build only accepts a single positional argument: package name, but multiple were specified")
			usage()
			os.Exit(1)
		}
		target := *target
		if target == "" && filepath.Ext(*outpath) == ".wasm" {
			target = "wasm"
		}
		err := Build(pkgName, *outpath, target, config)
		handleCompilerError(err)
	case "build-builtins":
		// Note: this command is only meant to be used while making a release!
		if *outpath == "" {
			fmt.Fprintln(os.Stderr, "No output filename supplied (-o).")
			usage()
			os.Exit(1)
		}
		if *target == "" {
			fmt.Fprintln(os.Stderr, "No target (-target).")
		}
		err := compileBuiltins(*target, func(path string) error {
			return moveFile(path, *outpath)
		})
		handleCompilerError(err)
	case "flash", "gdb":
		if *outpath != "" {
			fmt.Fprintln(os.Stderr, "Output cannot be specified with the flash command.")
			usage()
			os.Exit(1)
		}
		if command == "flash" {
			err := Flash(flag.Arg(0), *target, *port, config)
			handleCompilerError(err)
		} else {
			if !config.debug {
				fmt.Fprintln(os.Stderr, "Debug disabled while running gdb?")
				usage()
				os.Exit(1)
			}
			err := FlashGDB(flag.Arg(0), *target, *port, *ocdOutput, config)
			handleCompilerError(err)
		}
	case "run":
		if flag.NArg() != 1 {
			fmt.Fprintln(os.Stderr, "No package specified.")
			usage()
			os.Exit(1)
		}
		err := Run(flag.Arg(0), *target, config)
		handleCompilerError(err)
	case "test":
		pkgName := "."
		if flag.NArg() == 1 {
			pkgName = flag.Arg(0)
		} else if flag.NArg() > 1 {
			fmt.Fprintln(os.Stderr, "test only accepts a single positional argument: package name, but multiple were specified")
			usage()
			os.Exit(1)
		}
		err := Test(pkgName, *target, config)
		handleCompilerError(err)
	case "clean":
		// remove cache directory
		dir := cacheDir()
		err := os.RemoveAll(dir)
		if err != nil {
			fmt.Fprintln(os.Stderr, "cannot clean cache:", err)
			os.Exit(1)
		}
	case "help":
		usage()
	case "version":
		fmt.Printf("tinygo version %s %s/%s\n", version, runtime.GOOS, runtime.GOARCH)
	default:
		fmt.Fprintln(os.Stderr, "Unknown command:", command)
		usage()
		os.Exit(1)
	}
}
