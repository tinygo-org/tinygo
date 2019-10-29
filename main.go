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
	"time"

	"github.com/tinygo-org/tinygo/compiler"
	"github.com/tinygo-org/tinygo/goenv"
	"github.com/tinygo-org/tinygo/interp"
	"github.com/tinygo-org/tinygo/loader"

	serial "go.bug.st/serial.v1"
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
	scheduler     string
	printIR       bool
	dumpSSA       bool
	verifyIR      bool
	debug         bool
	printSizes    string
	cFlags        []string
	ldFlags       []string
	tags          string
	wasmAbi       string
	heapSize      int64
	testConfig    compiler.TestConfig
}

// Helper function for Compiler object.
func Compile(pkgName, outpath string, spec *TargetSpec, config *BuildConfig, action func(string) error) error {
	if config.gc == "" && spec.GC != "" {
		config.gc = spec.GC
	}

	root := goenv.Get("TINYGOROOT")

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

	goroot := goenv.Get("GOROOT")
	if goroot == "" {
		return errors.New("cannot locate $GOROOT, please set it manually")
	}
	tags := spec.BuildTags
	major, minor, err := getGorootVersion(goroot)
	if err != nil {
		return fmt.Errorf("could not read version from GOROOT (%v): %v", goroot, err)
	}
	if major != 1 || (minor != 11 && minor != 12 && minor != 13) {
		return fmt.Errorf("requires go version 1.11, 1.12, or 1.13, got go%d.%d", major, minor)
	}
	for i := 1; i <= minor; i++ {
		tags = append(tags, fmt.Sprintf("go1.%d", i))
	}
	if extraTags := strings.Fields(config.tags); len(extraTags) != 0 {
		tags = append(tags, extraTags...)
	}
	scheduler := spec.Scheduler
	if config.scheduler != "" {
		scheduler = config.scheduler
	}
	compilerConfig := compiler.Config{
		Triple:        spec.Triple,
		CPU:           spec.CPU,
		Features:      spec.Features,
		GOOS:          spec.GOOS,
		GOARCH:        spec.GOARCH,
		GC:            config.gc,
		PanicStrategy: config.panicStrategy,
		Scheduler:     scheduler,
		CFlags:        cflags,
		LDFlags:       ldflags,
		ClangHeaders:  getClangHeaderPath(root),
		Debug:         config.debug,
		DumpSSA:       config.dumpSSA,
		VerifyIR:      config.verifyIR,
		TINYGOROOT:    root,
		GOROOT:        goroot,
		GOPATH:        goenv.Get("GOPATH"),
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

	err = interp.Run(c.Module(), config.dumpSSA)
	if err != nil {
		return err
	}
	if err := c.Verify(); err != nil {
		return errors.New("verification error after interpreting runtime.initAll")
	}

	if spec.GOOS != "darwin" {
		c.ApplyFunctionSections() // -ffunction-sections
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
		ldflags = append(ldflags, "-o", executable, objfile, "-L", root)
		if spec.RTLib == "compiler-rt" {
			ldflags = append(ldflags, librt)
		}
		if spec.GOARCH == "wasm" {
			// Round heap size to next multiple of 65536 (the WebAssembly page
			// size).
			heapSize := (config.heapSize + (65536 - 1)) &^ (65536 - 1)
			ldflags = append(ldflags, "--initial-memory="+strconv.FormatInt(heapSize, 10))
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
		if outext == ".hex" || outext == ".bin" || outext == ".gba" {
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

	switch spec.FlashMethod {
	case "command", "":
		switch {
		case strings.Contains(spec.FlashCommand, "{hex}"):
			fileExt = ".hex"
		case strings.Contains(spec.FlashCommand, "{elf}"):
			fileExt = ".elf"
		case strings.Contains(spec.FlashCommand, "{bin}"):
			fileExt = ".bin"
		case strings.Contains(spec.FlashCommand, "{uf2}"):
			fileExt = ".uf2"
		default:
			return errors.New("invalid target file - did you forget the {hex} token in the 'flash-command' section?")
		}
	case "msd":
		if spec.FlashFilename == "" {
			return errors.New("invalid target file: flash-method was set to \"msd\" but no msd-firmware-name was set")
		}
		fileExt = filepath.Ext(spec.FlashFilename)
	case "openocd":
		fileExt = ".hex"
	case "native":
		return errors.New("unknown flash method \"native\" - did you miss a -target flag?")
	default:
		return errors.New("unknown flash method: " + spec.FlashMethod)
	}

	return Compile(pkgName, fileExt, spec, config, func(tmppath string) error {
		// do we need port reset to put MCU into bootloader mode?
		if spec.PortReset == "true" {
			err := touchSerialPortAt1200bps(port)
			if err != nil {
				return &commandError{"failed to reset port", tmppath, err}
			}
			// give the target MCU a chance to restart into bootloader
			time.Sleep(3 * time.Second)
		}

		// this flashing method copies the binary data to a Mass Storage Device (msd)
		switch spec.FlashMethod {
		case "", "command":
			// Create the command.
			flashCmd := spec.FlashCommand
			fileToken := "{" + fileExt[1:] + "}"
			flashCmd = strings.Replace(flashCmd, fileToken, tmppath, -1)
			flashCmd = strings.Replace(flashCmd, "{port}", port, -1)

			// Execute the command.
			cmd := exec.Command("/bin/sh", "-c", flashCmd)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Dir = goenv.Get("TINYGOROOT")
			err := cmd.Run()
			if err != nil {
				return &commandError{"failed to flash", tmppath, err}
			}
			return nil
		case "msd":
			switch fileExt {
			case ".uf2":
				err := flashUF2UsingMSD(spec.FlashVolume, tmppath)
				if err != nil {
					return &commandError{"failed to flash", tmppath, err}
				}
				return nil
			case ".hex":
				err := flashHexUsingMSD(spec.FlashVolume, tmppath)
				if err != nil {
					return &commandError{"failed to flash", tmppath, err}
				}
				return nil
			default:
				return errors.New("mass storage device flashing currently only supports uf2 and hex")
			}
		case "openocd":
			args, err := spec.OpenOCDConfiguration()
			if err != nil {
				return err
			}
			args = append(args, "-c", "program "+tmppath+" reset exit")
			cmd := exec.Command("openocd", args...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err = cmd.Run()
			if err != nil {
				return &commandError{"failed to flash", tmppath, err}
			}
			return nil
		default:
			return fmt.Errorf("unknown flash method: %s", spec.FlashMethod)
		}
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
		// Find a good way to run GDB.
		gdbInterface := spec.FlashMethod
		switch gdbInterface {
		case "msd", "command", "":
			if gdbInterface == "" {
				gdbInterface = "command"
			}
			if spec.OpenOCDInterface != "" && spec.OpenOCDTarget != "" {
				gdbInterface = "openocd"
			}
		}

		// Run the GDB server, if necessary.
		var gdbCommands []string
		switch gdbInterface {
		case "native":
			// Run GDB directly.
		case "openocd":
			gdbCommands = append(gdbCommands, "target remote :3333", "monitor halt", "load", "monitor reset halt")

			// We need a separate debugging daemon for on-chip debugging.
			args, err := spec.OpenOCDConfiguration()
			if err != nil {
				return err
			}
			daemon := exec.Command("openocd", args...)
			if ocdOutput {
				// Make it clear which output is from the daemon.
				w := &ColorWriter{
					Out:    os.Stderr,
					Prefix: "openocd: ",
					Color:  TermColorYellow,
				}
				daemon.Stdout = w
				daemon.Stderr = w
			}
			// Make sure the daemon doesn't receive Ctrl-C that is intended for
			// GDB (to break the currently executing program).
			setCommandAsDaemon(daemon)
			// Start now, and kill it on exit.
			daemon.Start()
			defer func() {
				daemon.Process.Signal(os.Interrupt)
				// Maybe we should send a .Kill() after x seconds?
				daemon.Wait()
			}()
		case "msd":
			return errors.New("gdb is not supported for drag-and-drop programmable devices")
		default:
			return fmt.Errorf("gdb is not supported with interface %#v", gdbInterface)
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
		for _, cmd := range gdbCommands {
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

func touchSerialPortAt1200bps(port string) error {
	// Open port
	p, err := serial.Open(port, &serial.Mode{BaudRate: 1200})
	if err != nil {
		return fmt.Errorf("opening port: %s", err)
	}
	defer p.Close()

	p.SetDTR(false)
	return nil
}

func flashUF2UsingMSD(volume, tmppath string) error {
	// find standard UF2 info path
	infoPath := "/media/*/" + volume + "/INFO_UF2.TXT"
	if runtime.GOOS == "darwin" {
		infoPath = "/Volumes/" + volume + "/INFO_UF2.TXT"
	}

	d, err := filepath.Glob(infoPath)
	if err != nil {
		return err
	}
	if d == nil {
		return errors.New("unable to locate UF2 device: " + volume)
	}

	return moveFile(tmppath, filepath.Dir(d[0])+"/flash.uf2")
}

func flashHexUsingMSD(volume, tmppath string) error {
	// find expected volume path
	destPath := "/media/*/" + volume
	if runtime.GOOS == "darwin" {
		destPath = "/Volumes/" + volume
	}

	d, err := filepath.Glob(destPath)
	if err != nil {
		return err
	}
	if d == nil {
		return errors.New("unable to locate device: " + volume)
	}

	return moveFile(tmppath, d[0]+"/flash.hex")
}

// parseSize converts a human-readable size (with k/m/g suffix) into a plain
// number.
func parseSize(s string) (int64, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	if len(s) == 0 {
		return 0, errors.New("no size provided")
	}
	multiply := int64(1)
	switch s[len(s)-1] {
	case 'k':
		multiply = 1 << 10
	case 'm':
		multiply = 1 << 20
	case 'g':
		multiply = 1 << 30
	}
	if multiply != 1 {
		s = s[:len(s)-1]
	}
	n, err := strconv.ParseInt(s, 0, 64)
	n *= multiply
	return n, err
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
	fmt.Fprintln(os.Stderr, "  env:   list environment variables used during build")
	fmt.Fprintln(os.Stderr, "  clean: empty cache directory ("+goenv.Get("GOCACHE")+")")
	fmt.Fprintln(os.Stderr, "  help:  print this help text")
	fmt.Fprintln(os.Stderr, "\nflags:")
	flag.PrintDefaults()
}

func handleCompilerError(err error) {
	if err != nil {
		switch err := err.(type) {
		case *interp.Unsupported:
			// hit an unknown/unsupported instruction
			fmt.Fprintln(os.Stderr, "unsupported instruction during init evaluation:")
			err.Inst.Dump()
			fmt.Fprintln(os.Stderr)
		case types.Error:
			fmt.Fprintln(os.Stderr, err)
		case loader.Errors:
			fmt.Fprintln(os.Stderr, "#", err.Pkg.ImportPath)
			for _, err := range err.Errs {
				fmt.Fprintln(os.Stderr, err)
			}
		case *multiError:
			for _, err := range err.Errs {
				fmt.Fprintln(os.Stderr, err)
			}
		default:
			fmt.Fprintln(os.Stderr, "error:", err)
		}
		os.Exit(1)
	}
}

func main() {
	outpath := flag.String("o", "", "output filename")
	opt := flag.String("opt", "z", "optimization level: 0, 1, 2, s, z")
	gc := flag.String("gc", "", "garbage collector to use (none, leaking, conservative)")
	panicStrategy := flag.String("panic", "print", "panic strategy (print, trap)")
	scheduler := flag.String("scheduler", "", "which scheduler to use (coroutines, tasks)")
	printIR := flag.Bool("printir", false, "print LLVM IR")
	dumpSSA := flag.Bool("dumpssa", false, "dump internal Go SSA")
	verifyIR := flag.Bool("verifyir", false, "run extra verification steps on LLVM IR")
	tags := flag.String("tags", "", "a space-separated list of extra build tags")
	target := flag.String("target", "", "LLVM target | .json file with TargetSpec")
	printSize := flag.String("size", "", "print sizes (none, short, full)")
	nodebug := flag.Bool("no-debug", false, "disable DWARF debug symbol generation")
	ocdOutput := flag.Bool("ocd-output", false, "print OCD daemon output during debug")
	port := flag.String("port", "/dev/ttyACM0", "flash port")
	cFlags := flag.String("cflags", "", "additional cflags for compiler")
	ldFlags := flag.String("ldflags", "", "additional ldflags for linker")
	wasmAbi := flag.String("wasm-abi", "js", "WebAssembly ABI conventions: js (no i64 params) or generic")
	heapSize := flag.String("heap-size", "1M", "default heap size in bytes (only supported by WebAssembly)")

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
		scheduler:     *scheduler,
		printIR:       *printIR,
		dumpSSA:       *dumpSSA,
		verifyIR:      *verifyIR,
		debug:         !*nodebug,
		printSizes:    *printSize,
		tags:          *tags,
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

	var err error
	if config.heapSize, err = parseSize(*heapSize); err != nil {
		fmt.Fprintln(os.Stderr, "Could not read heap size:", *heapSize)
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
		err := os.RemoveAll(goenv.Get("GOCACHE"))
		if err != nil {
			fmt.Fprintln(os.Stderr, "cannot clean cache:", err)
			os.Exit(1)
		}
	case "help":
		usage()
	case "version":
		goversion := "<unknown>"
		if s, err := getGorootVersionString(goenv.Get("GOROOT")); err == nil {
			goversion = s
		}
		fmt.Printf("tinygo version %s %s/%s (using go version %s)\n", version, runtime.GOOS, runtime.GOARCH, goversion)
	case "env":
		if flag.NArg() == 0 {
			// Show all environment variables.
			for _, key := range goenv.Keys {
				fmt.Printf("%s=%#v\n", key, goenv.Get(key))
			}
		} else {
			// Show only one (or a few) environment variables.
			for i := 0; i < flag.NArg(); i++ {
				fmt.Println(goenv.Get(flag.Arg(i)))
			}
		}
	default:
		fmt.Fprintln(os.Stderr, "Unknown command:", command)
		usage()
		os.Exit(1)
	}
}
