package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"runtime"
	"runtime/pprof"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/shlex"
	"github.com/inhies/go-bytesize"
	"github.com/mattn/go-colorable"
	"github.com/tinygo-org/tinygo/builder"
	"github.com/tinygo-org/tinygo/compileopts"
	"github.com/tinygo-org/tinygo/diagnostics"
	"github.com/tinygo-org/tinygo/goenv"
	"github.com/tinygo-org/tinygo/loader"
	"golang.org/x/tools/go/buildutil"
	"tinygo.org/x/go-llvm"

	"go.bug.st/serial"
	"go.bug.st/serial/enumerator"
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

// moveFile renames the file from src to dst. If renaming doesn't work (for
// example, the rename crosses a filesystem boundary), the file is copied and
// the old file is removed.
func moveFile(src, dst string) error {
	err := os.Rename(src, dst)
	if err == nil {
		// Success!
		return nil
	}
	// Failed to move, probably a different filesystem.
	// Do a copy + remove.
	err = copyFile(src, dst)
	if err != nil {
		return err
	}
	return os.Remove(src)
}

// copyFile copies the given file or directory from src to dst. It can copy over
// a possibly already existing file (but not directory) at the destination.
func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	st, err := source.Stat()
	if err != nil {
		return err
	}

	if st.IsDir() {
		err := os.Mkdir(dst, st.Mode().Perm())
		if err != nil {
			return err
		}
		names, err := source.Readdirnames(0)
		if err != nil {
			return err
		}
		for _, name := range names {
			err := copyFile(filepath.Join(src, name), filepath.Join(dst, name))
			if err != nil {
				return err
			}
		}
		return nil
	} else {
		destination, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, st.Mode())
		if err != nil {
			return err
		}
		defer destination.Close()

		_, err = io.Copy(destination, source)
		return err
	}
}

// executeCommand is a simple wrapper to exec.Cmd
func executeCommand(options *compileopts.Options, name string, arg ...string) *exec.Cmd {
	if options.PrintCommands != nil {
		options.PrintCommands(name, arg...)
	}
	return exec.Command(name, arg...)
}

// printCommand prints a command to stdout while formatting it like a real
// command (escaping characters etc). The resulting command should be easy to
// run directly in a shell, although it is not guaranteed to be a safe shell
// escape. That's not a problem as the primary use case is printing the command,
// not running it.
func printCommand(cmd string, args ...string) {
	command := append([]string{cmd}, args...)
	for i, arg := range command {
		// Source: https://www.oreilly.com/library/view/learning-the-bash/1565923472/ch01s09.html
		const specialChars = "~`#$&*()\\|[]{};'\"<>?! "
		if strings.ContainsAny(arg, specialChars) {
			// See: https://stackoverflow.com/questions/15783701/which-characters-need-to-be-escaped-when-using-bash
			arg = "'" + strings.ReplaceAll(arg, `'`, `'\''`) + "'"
			command[i] = arg
		}
	}
	fmt.Fprintln(os.Stderr, strings.Join(command, " "))
}

// Build compiles and links the given package and writes it to outpath.
func Build(pkgName, outpath string, options *compileopts.Options) error {
	config, err := builder.NewConfig(options)
	if err != nil {
		return err
	}

	if options.PrintJSON {
		b, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			handleCompilerError(err)
		}
		fmt.Printf("%s\n", string(b))
		return nil
	}

	// Create a temporary directory for intermediary files.
	tmpdir, err := os.MkdirTemp("", "tinygo")
	if err != nil {
		return err
	}
	if !options.Work {
		defer os.RemoveAll(tmpdir)
	}

	// Do the build.
	result, err := builder.Build(pkgName, outpath, tmpdir, config)
	if err != nil {
		return err
	}

	if result.Binary != "" {
		// If result.Binary is set, it means there is a build output (elf, hex,
		// etc) that we need to move to the outpath. If it isn't set, it means
		// the build output was a .ll, .bc or .o file that has already been
		// written to outpath and so we don't need to do anything.

		if outpath == "" {
			if strings.HasSuffix(pkgName, ".go") {
				// A Go file was specified directly on the command line.
				// Base the binary name off of it.
				outpath = filepath.Base(pkgName[:len(pkgName)-3]) + config.DefaultBinaryExtension()
			} else {
				// Pick a default output path based on the main directory.
				outpath = filepath.Base(result.MainDir) + config.DefaultBinaryExtension()
			}
		}

		if err := os.Rename(result.Binary, outpath); err != nil {
			// Moving failed. Do a file copy.
			inf, err := os.Open(result.Binary)
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
		}
	}

	// Move was successful.
	return nil
}

// Test runs the tests in the given package. Returns whether the test passed and
// possibly an error if the test failed to run.
func Test(pkgName string, stdout, stderr io.Writer, options *compileopts.Options, outpath string) (bool, error) {
	options.TestConfig.CompileTestBinary = true
	config, err := builder.NewConfig(options)
	if err != nil {
		return false, err
	}

	testConfig := &options.TestConfig

	// Pass test flags to the test binary.
	var flags []string
	if testConfig.Verbose {
		flags = append(flags, "-test.v")
	}
	if testConfig.Short {
		flags = append(flags, "-test.short")
	}
	if testConfig.RunRegexp != "" {
		flags = append(flags, "-test.run="+testConfig.RunRegexp)
	}
	if testConfig.SkipRegexp != "" {
		flags = append(flags, "-test.skip="+testConfig.SkipRegexp)
	}
	if testConfig.BenchRegexp != "" {
		flags = append(flags, "-test.bench="+testConfig.BenchRegexp)
	}
	if testConfig.BenchTime != "" {
		flags = append(flags, "-test.benchtime="+testConfig.BenchTime)
	}
	if testConfig.BenchMem {
		flags = append(flags, "-test.benchmem")
	}
	if testConfig.Count != nil && *testConfig.Count != 1 {
		flags = append(flags, "-test.count="+strconv.Itoa(*testConfig.Count))
	}
	if testConfig.Shuffle != "" {
		flags = append(flags, "-test.shuffle="+testConfig.Shuffle)
	}

	logToStdout := testConfig.Verbose || testConfig.BenchRegexp != ""

	var buf bytes.Buffer
	var output io.Writer = &buf
	// Send the test output to stdout if -v or -bench
	if logToStdout {
		output = os.Stdout
	}

	passed := false
	var duration time.Duration
	result, err := buildAndRun(pkgName, config, output, flags, nil, 0, func(cmd *exec.Cmd, result builder.BuildResult) error {
		if testConfig.CompileOnly || outpath != "" {
			// Write test binary to the specified file name.
			if outpath == "" {
				// No -o path was given, so create one now.
				// This matches the behavior of go test.
				outpath = filepath.Base(result.MainDir) + ".test"
			}
			copyFile(result.Binary, outpath)
		}
		if testConfig.CompileOnly {
			// Do not run the test.
			passed = true
			return nil
		}

		// Tests are always run in the package directory.
		cmd.Dir = result.MainDir

		// Run the test.
		start := time.Now()
		err = cmd.Run()
		duration = time.Since(start)
		passed = err == nil

		// if verbose or benchmarks, then output is already going to stdout
		// However, if we failed and weren't printing to stdout, print the output we accumulated.
		if !passed && !logToStdout {
			buf.WriteTo(stdout)
		}

		if _, ok := err.(*exec.ExitError); ok {
			// Binary exited with a non-zero exit code, which means the test
			// failed. Return nil to avoid printing a useless "exited with
			// error" error message.
			return nil
		}
		return err
	})

	if testConfig.CompileOnly {
		return true, nil
	}

	importPath := strings.TrimSuffix(result.ImportPath, ".test")

	var w io.Writer = stdout
	if logToStdout {
		w = os.Stdout
	}
	if err, ok := err.(loader.NoTestFilesError); ok {
		fmt.Fprintf(w, "?   \t%s\t[no test files]\n", err.ImportPath)
		// Pretend the test passed - it at least didn't fail.
		return true, nil
	} else if passed {
		fmt.Fprintf(w, "ok  \t%s\t%.3fs\n", importPath, duration.Seconds())
	} else {
		fmt.Fprintf(w, "FAIL\t%s\t%.3fs\n", importPath, duration.Seconds())
	}
	return passed, err
}

func dirsToModuleRootRel(maindir, modroot string) []string {
	var dirs []string
	last := ".."
	// strip off path elements until we hit the module root
	// adding `..`, `../..`, `../../..` until we're done
	for maindir != modroot {
		dirs = append(dirs, last)
		last = filepath.Join(last, "..")
		maindir = filepath.Dir(maindir)
	}
	dirs = append(dirs, ".")
	return dirs
}

func dirsToModuleRootAbs(maindir, modroot string) []string {
	var dirs = []string{maindir}
	last := filepath.Join(maindir, "..")
	// strip off path elements until we hit the module root
	// adding `..`, `../..`, `../../..` until we're done
	for maindir != modroot {
		dirs = append(dirs, last)
		last = filepath.Join(last, "..")
		maindir = filepath.Dir(maindir)
	}
	return dirs
}

// Flash builds and flashes the built binary to the given serial port.
func Flash(pkgName, port string, options *compileopts.Options) error {
	config, err := builder.NewConfig(options)
	if err != nil {
		return err
	}

	// determine the type of file to compile
	var fileExt string

	flashMethod, _ := config.Programmer()
	switch flashMethod {
	case "command", "":
		switch {
		case strings.Contains(config.Target.FlashCommand, "{hex}"):
			fileExt = ".hex"
		case strings.Contains(config.Target.FlashCommand, "{elf}"):
			fileExt = ".elf"
		case strings.Contains(config.Target.FlashCommand, "{bin}"):
			fileExt = ".bin"
		case strings.Contains(config.Target.FlashCommand, "{uf2}"):
			fileExt = ".uf2"
		case strings.Contains(config.Target.FlashCommand, "{zip}"):
			fileExt = ".zip"
		default:
			return errors.New("invalid target file - did you forget the {hex} token in the 'flash-command' section?")
		}
	case "msd":
		if config.Target.FlashFilename == "" {
			return errors.New("invalid target file: flash-method was set to \"msd\" but no msd-firmware-name was set")
		}
		fileExt = filepath.Ext(config.Target.FlashFilename)
	case "openocd":
		fileExt = ".hex"
	case "bmp":
		fileExt = ".elf"
	case "native":
		return errors.New("unknown flash method \"native\" - did you miss a -target flag?")
	default:
		return errors.New("unknown flash method: " + flashMethod)
	}

	// Create a temporary directory for intermediary files.
	tmpdir, err := os.MkdirTemp("", "tinygo")
	if err != nil {
		return err
	}
	if !options.Work {
		defer os.RemoveAll(tmpdir)
	}

	// Build the binary.
	result, err := builder.Build(pkgName, fileExt, tmpdir, config)
	if err != nil {
		return err
	}

	// do we need port reset to put MCU into bootloader mode?
	if config.Target.PortReset == "true" && flashMethod != "openocd" {
		port, err := getDefaultPort(port, config.Target.SerialPort)
		if err == nil {
			err = touchSerialPortAt1200bps(port)
			if err != nil {
				return &commandError{"failed to reset port", port, err}
			}
			// give the target MCU a chance to restart into bootloader
			time.Sleep(3 * time.Second)
		}
	}

	// Flash the binary to the MCU.
	switch flashMethod {
	case "", "command":
		// Create the command.
		flashCmd := config.Target.FlashCommand
		flashCmdList, err := shlex.Split(flashCmd)
		if err != nil {
			return fmt.Errorf("could not parse flash command %#v: %w", flashCmd, err)
		}

		if strings.Contains(flashCmd, "{port}") {
			var err error
			port, err = getDefaultPort(port, config.Target.SerialPort)
			if err != nil {
				return err
			}
		}

		// Fill in fields in the command template.
		fileToken := "{" + fileExt[1:] + "}"
		for i, arg := range flashCmdList {
			arg = strings.ReplaceAll(arg, fileToken, result.Binary)
			arg = strings.ReplaceAll(arg, "{port}", port)
			flashCmdList[i] = arg
		}

		// Execute the command.
		if len(flashCmdList) < 2 {
			return fmt.Errorf("invalid flash command: %#v", flashCmd)
		}
		cmd := executeCommand(config.Options, flashCmdList[0], flashCmdList[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Dir = goenv.Get("TINYGOROOT")
		err = cmd.Run()
		if err != nil {
			return &commandError{"failed to flash", result.Binary, err}
		}
	case "msd":
		// this flashing method copies the binary data to a Mass Storage Device (msd)
		switch fileExt {
		case ".uf2":
			err := flashUF2UsingMSD(config.Target.FlashVolume, result.Binary, config.Options)
			if err != nil {
				return &commandError{"failed to flash", result.Binary, err}
			}
		case ".hex":
			err := flashHexUsingMSD(config.Target.FlashVolume, result.Binary, config.Options)
			if err != nil {
				return &commandError{"failed to flash", result.Binary, err}
			}
		default:
			return errors.New("mass storage device flashing currently only supports uf2 and hex")
		}
	case "openocd":
		args, err := config.OpenOCDConfiguration()
		if err != nil {
			return err
		}
		exit := " reset exit"
		if config.Target.OpenOCDVerify != nil && *config.Target.OpenOCDVerify {
			exit = " verify" + exit
		}
		args = append(args, "-c", "program "+filepath.ToSlash(result.Binary)+exit)
		cmd := executeCommand(config.Options, "openocd", args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			return &commandError{"failed to flash", result.Binary, err}
		}
	case "bmp":
		gdb, err := config.Target.LookupGDB()
		if err != nil {
			return err
		}
		var bmpGDBPort string
		bmpGDBPort, _, err = getBMPPorts()
		if err != nil {
			return err
		}
		args := []string{"-ex", "target extended-remote " + bmpGDBPort, "-ex", "monitor swdp_scan", "-ex", "attach 1", "-ex", "load", filepath.ToSlash(result.Binary)}
		cmd := executeCommand(config.Options, gdb, args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			return &commandError{"failed to flash", result.Binary, err}
		}
	default:
		return fmt.Errorf("unknown flash method: %s", flashMethod)
	}
	if options.Monitor {
		return Monitor(result.Executable, "", config)
	}
	return nil
}

// Debug compiles and flashes a program to a microcontroller (just like Flash)
// but instead of resetting the target, it will drop into a debug shell like GDB
// or LLDB. You can then set breakpoints, run the `continue` command to start,
// hit Ctrl+C to break the running program, etc.
//
// Note: this command is expected to execute just before exiting, as it
// modifies global state.
func Debug(debugger, pkgName string, ocdOutput bool, options *compileopts.Options) error {
	config, err := builder.NewConfig(options)
	if err != nil {
		return err
	}
	var cmdName string
	switch debugger {
	case "gdb":
		cmdName, err = config.Target.LookupGDB()
	case "lldb":
		cmdName, err = builder.LookupCommand("lldb")
	}
	if err != nil {
		return err
	}

	// Create a temporary directory for intermediary files.
	tmpdir, err := os.MkdirTemp("", "tinygo")
	if err != nil {
		return err
	}
	if !options.Work {
		defer os.RemoveAll(tmpdir)
	}

	// Build the binary to debug.
	format, fileExt := config.EmulatorFormat()
	result, err := builder.Build(pkgName, fileExt, tmpdir, config)
	if err != nil {
		return err
	}

	// Find a good way to run GDB.
	gdbInterface, openocdInterface := config.Programmer()
	switch gdbInterface {
	case "msd", "command", "":
		emulator := config.EmulatorName()
		if emulator != "" {
			if emulator == "mgba" {
				gdbInterface = "mgba"
			} else if emulator == "simavr" {
				gdbInterface = "simavr"
			} else if strings.HasPrefix(emulator, "qemu-system-") {
				gdbInterface = "qemu"
			} else {
				// Assume QEMU as an emulator.
				gdbInterface = "qemu-user"
			}
		} else if openocdInterface != "" && config.Target.OpenOCDTarget != "" {
			gdbInterface = "openocd"
		} else if config.Target.JLinkDevice != "" {
			gdbInterface = "jlink"
		} else {
			gdbInterface = "native"
		}
	}

	// Run the GDB server, if necessary.
	port := ""
	var gdbCommands []string
	var daemon *exec.Cmd
	emulator, err := config.Emulator(format, result.Binary)
	if err != nil {
		return err
	}
	switch gdbInterface {
	case "native":
		// Run GDB directly.
	case "bmp":
		var bmpGDBPort string
		bmpGDBPort, _, err = getBMPPorts()
		if err != nil {
			return err
		}
		port = bmpGDBPort
		gdbCommands = append(gdbCommands, "monitor swdp_scan", "compare-sections", "attach 1", "load")
	case "openocd":
		port = ":3333"
		gdbCommands = append(gdbCommands, "monitor halt", "load", "monitor reset halt")

		// We need a separate debugging daemon for on-chip debugging.
		args, err := config.OpenOCDConfiguration()
		if err != nil {
			return err
		}
		daemon = executeCommand(config.Options, "openocd", args...)
		if ocdOutput {
			// Make it clear which output is from the daemon.
			w := &ColorWriter{
				Out:    colorable.NewColorableStderr(),
				Prefix: "openocd: ",
				Color:  TermColorYellow,
			}
			daemon.Stdout = w
			daemon.Stderr = w
		}
	case "jlink":
		port = ":2331"
		gdbCommands = append(gdbCommands, "load", "monitor reset halt")

		// We need a separate debugging daemon for on-chip debugging.
		daemon = executeCommand(config.Options, "JLinkGDBServer", "-device", config.Target.JLinkDevice)
		if ocdOutput {
			// Make it clear which output is from the daemon.
			w := &ColorWriter{
				Out:    colorable.NewColorableStderr(),
				Prefix: "jlink: ",
				Color:  TermColorYellow,
			}
			daemon.Stdout = w
			daemon.Stderr = w
		}
	case "qemu":
		port = ":1234"
		// Run in an emulator.
		args := append(emulator[1:], "-s", "-S")
		daemon = executeCommand(config.Options, emulator[0], args...)
		daemon.Stdout = os.Stdout
		daemon.Stderr = os.Stderr
	case "qemu-user":
		port = ":1234"
		// Run in an emulator.
		args := append([]string{"-g", "1234"}, emulator[1:]...)
		daemon = executeCommand(config.Options, emulator[0], args...)
		daemon.Stdout = os.Stdout
		daemon.Stderr = os.Stderr
	case "mgba":
		port = ":2345"
		// Run in an emulator.
		args := append(emulator[1:], "-g")
		daemon = executeCommand(config.Options, emulator[0], args...)
		daemon.Stdout = os.Stdout
		daemon.Stderr = os.Stderr
	case "simavr":
		port = ":1234"
		// Run in an emulator.
		args := append(emulator[1:], "-g")
		daemon = executeCommand(config.Options, emulator[0], args...)
		daemon.Stdout = os.Stdout
		daemon.Stderr = os.Stderr
	case "msd":
		return errors.New("gdb is not supported for drag-and-drop programmable devices")
	default:
		return fmt.Errorf("gdb is not supported with interface %#v", gdbInterface)
	}

	if daemon != nil {
		// Make sure the daemon doesn't receive Ctrl-C that is intended for
		// GDB (to break the currently executing program).
		setCommandAsDaemon(daemon)

		// Start now, and kill it on exit.
		err = daemon.Start()
		if err != nil {
			return &commandError{"failed to run", daemon.Path, err}
		}
		defer func() {
			daemon.Process.Signal(os.Interrupt)
			var stopped uint32
			go func() {
				time.Sleep(time.Millisecond * 100)
				if atomic.LoadUint32(&stopped) == 0 {
					daemon.Process.Kill()
				}
			}()
			daemon.Wait()
			atomic.StoreUint32(&stopped, 1)
		}()
	}

	// Ignore Ctrl-C, it must be passed on to GDB.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
		}
	}()

	// Construct and execute a gdb or lldb command.
	// By default: gdb -ex run <binary>
	// Exit the debugger with Ctrl-D.
	params := []string{result.Executable}
	switch debugger {
	case "gdb":
		if port != "" {
			params = append(params, "-ex", "target extended-remote "+port)
		}
		for _, cmd := range gdbCommands {
			params = append(params, "-ex", cmd)
		}
	case "lldb":
		params = append(params, "--arch", config.Triple())
		if port != "" {
			if strings.HasPrefix(port, ":") {
				params = append(params, "-o", "gdb-remote "+port[1:])
			} else {
				return fmt.Errorf("cannot use LLDB over a gdb-remote that isn't a TCP port: %s", port)
			}
		}
		for _, cmd := range gdbCommands {
			if strings.HasPrefix(cmd, "monitor ") {
				params = append(params, "-o", "process plugin packet "+cmd)
			} else if cmd == "load" {
				params = append(params, "-o", "target modules load --load --slide 0")
			} else {
				return fmt.Errorf("don't know how to convert GDB command %#v to LLDB", cmd)
			}
		}
	}
	cmd := executeCommand(config.Options, cmdName, params...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return &commandError{"failed to run " + cmdName + " with", result.Executable, err}
	}
	return nil
}

// Run compiles and runs the given program. Depending on the target provided in
// the options, it will run the program directly on the host or will run it in
// an emulator. For example, -target=wasm will cause the binary to be run inside
// of a WebAssembly VM.
func Run(pkgName string, options *compileopts.Options, cmdArgs []string) error {
	config, err := builder.NewConfig(options)
	if err != nil {
		return err
	}

	_, err = buildAndRun(pkgName, config, os.Stdout, cmdArgs, nil, 0, func(cmd *exec.Cmd, result builder.BuildResult) error {
		return cmd.Run()
	})
	return err
}

// buildAndRun builds and runs the given program, writing output to stdout and
// errors to os.Stderr. It takes care of emulators (qemu, wasmtime, etc) and
// passes command line arguments and environment variables in a way appropriate
// for the given emulator.
func buildAndRun(pkgName string, config *compileopts.Config, stdout io.Writer, cmdArgs, environmentVars []string, timeout time.Duration, run func(cmd *exec.Cmd, result builder.BuildResult) error) (builder.BuildResult, error) {

	isSingleFile := strings.HasSuffix(pkgName, ".go")

	// Determine whether we're on a system that supports environment variables
	// and command line parameters (operating systems, WASI) or not (baremetal,
	// WebAssembly in the browser). If we're on a system without an environment,
	// we need to pass command line arguments and environment variables through
	// global variables (built into the binary directly) instead of the
	// conventional way.
	needsEnvInVars := config.GOOS() == "js"
	for _, tag := range config.BuildTags() {
		if tag == "baremetal" {
			needsEnvInVars = true
		}
	}
	var args, emuArgs, env []string
	var extraCmdEnv []string
	if needsEnvInVars {
		runtimeGlobals := make(map[string]string)
		if len(cmdArgs) != 0 {
			runtimeGlobals["osArgs"] = strings.Join(cmdArgs, "\x00")
		}
		if len(environmentVars) != 0 {
			runtimeGlobals["osEnv"] = strings.Join(environmentVars, "\x00")
		}
		if len(runtimeGlobals) != 0 {
			// This sets the global variables like they would be set with
			// `-ldflags="-X=runtime.osArgs=first\x00second`.
			// The runtime package has two variables (osArgs and osEnv) that are
			// both strings, from which the parameters and environment variables
			// are read.
			config.Options.GlobalValues = map[string]map[string]string{
				"runtime": runtimeGlobals,
			}
		}
	} else if config.EmulatorName() == "wasmtime" {
		for _, v := range environmentVars {
			emuArgs = append(emuArgs, "--env", v)
		}

		// Use of '--' argument no longer necessary as of Wasmtime v14:
		// https://github.com/bytecodealliance/wasmtime/pull/6946
		// args = append(args, "--")
		args = append(args, cmdArgs...)

		// Set this for nicer backtraces during tests, but don't override the user.
		if _, ok := os.LookupEnv("WASMTIME_BACKTRACE_DETAILS"); !ok {
			extraCmdEnv = append(extraCmdEnv, "WASMTIME_BACKTRACE_DETAILS=1")
		}
	} else {
		// Pass environment variables and command line parameters as usual.
		// This also works on qemu-aarch64 etc.
		args = cmdArgs
		env = environmentVars
	}

	// Create a temporary directory for intermediary files.
	tmpdir, err := os.MkdirTemp("", "tinygo")
	if err != nil {
		return builder.BuildResult{}, err
	}
	if !config.Options.Work {
		defer os.RemoveAll(tmpdir)
	}

	// Build the binary to be run.
	format, fileExt := config.EmulatorFormat()
	result, err := builder.Build(pkgName, fileExt, tmpdir, config)
	if err != nil {
		return result, err
	}

	// If needed, set a timeout on the command. This is done in tests so
	// they don't waste resources on a stalled test.
	var ctx context.Context
	if timeout != 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), timeout)
		defer cancel()
	}

	// Set up the command.
	var name string
	if config.Target.Emulator == "" {
		name = result.Binary
	} else {
		emulator, err := config.Emulator(format, result.Binary)
		if err != nil {
			return result, err
		}

		name = emulator[0]

		// wasmtime is a WebAssembly runtime CLI with WASI enabled by default.
		// By default, only stdio is allowed. For example, while STDOUT routes
		// to the host, other files don't. It also does not inherit environment
		// variables from the host. Some tests read testdata files, often from
		// outside the package directory. Other tests require temporary
		// writeable directories. We allow this by adding wasmtime flags below.
		if name == "wasmtime" {
			// Below adds additional wasmtime flags in case a test reads files
			// outside its directory, like "../testdata/e.txt". This allows any
			// relative directory up to the module root, even if the test never
			// reads any files.
			if config.TestConfig.CompileTestBinary {
				// Add relative dirs (../, ../..) up to module root (for wasip1)
				dirs := dirsToModuleRootRel(result.MainDir, result.ModuleRoot)

				// Add absolute dirs up to module root (for wasip2)
				dirs = append(dirs, dirsToModuleRootAbs(result.MainDir, result.ModuleRoot)...)

				for _, d := range dirs {
					emuArgs = append(emuArgs, "--dir="+d)
				}
			}

			dir := result.MainDir
			if isSingleFile {
				dir, _ = os.Getwd()
			}
			emuArgs = append(emuArgs, "--dir=.")
			emuArgs = append(emuArgs, "--dir="+dir)
			emuArgs = append(emuArgs, "--env=PWD="+dir)
		}

		emuArgs = append(emuArgs, emulator[1:]...)
		args = append(emuArgs, args...)
	}
	var cmd *exec.Cmd
	if ctx != nil {
		cmd = exec.CommandContext(ctx, name, args...)
	} else {
		cmd = exec.Command(name, args...)
	}
	cmd.Env = append(cmd.Env, env...)
	cmd.Env = append(cmd.Env, extraCmdEnv...)

	// Configure stdout/stderr. The stdout may go to a buffer, not a real
	// stdout.
	cmd.Stdout = newOutputWriter(stdout, result.Executable)
	cmd.Stderr = os.Stderr
	if config.EmulatorName() == "simavr" {
		cmd.Stdout = nil // don't print initial load commands
		cmd.Stderr = stdout
	}

	// If this is a test, reserve CPU time for it so that increased
	// parallelism doesn't blow up memory usage. If this isn't a test but
	// simply `tinygo run`, then it is practically a no-op.
	config.Options.Semaphore <- struct{}{}
	defer func() {
		<-config.Options.Semaphore
	}()

	// Run binary.
	if config.Options.PrintCommands != nil {
		config.Options.PrintCommands(cmd.Path, cmd.Args...)
	}
	err = run(cmd, result)
	if err != nil {
		if ctx != nil && ctx.Err() == context.DeadlineExceeded {
			fmt.Fprintf(stdout, "--- timeout of %s exceeded, terminating...\n", timeout)
			err = ctx.Err()
		}
		return result, &commandError{"failed to run compiled binary", result.Binary, err}
	}
	return result, nil
}

func touchSerialPortAt1200bps(port string) (err error) {
	retryCount := 3
	for i := 0; i < retryCount; i++ {
		// Open port
		p, e := serial.Open(port, &serial.Mode{BaudRate: 1200})
		if e != nil {
			if runtime.GOOS == `windows` {
				se, ok := e.(*serial.PortError)
				if ok && se.Code() == serial.InvalidSerialPort {
					// InvalidSerialPort error occurs when transitioning to boot
					return nil
				}
			}
			time.Sleep(1 * time.Second)
			err = e
			continue
		}
		defer p.Close()

		p.SetDTR(false)
		return nil
	}
	return fmt.Errorf("opening port: %s", err)
}

func flashUF2UsingMSD(volumes []string, tmppath string, options *compileopts.Options) error {
	for start := time.Now(); time.Since(start) < options.Timeout; {
		// Find a UF2 mount point.
		mounts, err := findFATMounts(options)
		if err != nil {
			return err
		}
		for _, mount := range mounts {
			for _, volume := range volumes {
				if mount.name != volume {
					continue
				}
				if _, err := os.Stat(filepath.Join(mount.path, "INFO_UF2.TXT")); err != nil {
					// No INFO_UF2.TXT found, which is expected on a UF2
					// filesystem.
					continue
				}
				// Found the filesystem, so flash the device!
				return moveFile(tmppath, filepath.Join(mount.path, "flash.uf2"))
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	return errors.New("unable to locate any volume: [" + strings.Join(volumes, ",") + "]")
}

func flashHexUsingMSD(volumes []string, tmppath string, options *compileopts.Options) error {
	for start := time.Now(); time.Since(start) < options.Timeout; {
		// Find all mount points.
		mounts, err := findFATMounts(options)
		if err != nil {
			return err
		}
		for _, mount := range mounts {
			for _, volume := range volumes {
				if mount.name != volume {
					continue
				}
				// Found the filesystem, so flash the device!
				return moveFile(tmppath, filepath.Join(mount.path, "flash.hex"))
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	return errors.New("unable to locate any volume: [" + strings.Join(volumes, ",") + "]")
}

type mountPoint struct {
	name string
	path string
}

// Find all the mount points on the system that use the FAT filesystem.
func findFATMounts(options *compileopts.Options) ([]mountPoint, error) {
	var points []mountPoint
	switch runtime.GOOS {
	case "darwin":
		list, err := os.ReadDir("/Volumes")
		if err != nil {
			return nil, fmt.Errorf("could not list mount points: %w", err)
		}
		for _, elem := range list {
			// TODO: find a way to check for the filesystem type.
			// (Only return FAT filesystems).
			points = append(points, mountPoint{
				name: elem.Name(),
				path: filepath.Join("/Volumes", elem.Name()),
			})
		}
		sort.Slice(points, func(i, j int) bool {
			return points[i].path < points[j].name
		})
		return points, nil
	case "linux":
		tab, err := os.ReadFile("/proc/mounts") // symlink to /proc/self/mounts on my system
		if err != nil {
			return nil, fmt.Errorf("could not list mount points: %w", err)
		}
		for _, line := range strings.Split(string(tab), "\n") {
			fields := strings.Fields(line)
			if len(fields) <= 2 {
				continue
			}
			fstype := fields[2]
			// chromeos bind mounts use 9p
			if !(fstype == "vfat" || fstype == "9p") {
				continue
			}
			fspath := strings.ReplaceAll(fields[1], "\\040", " ")
			points = append(points, mountPoint{
				name: filepath.Base(fspath),
				path: fspath,
			})
		}
		return points, nil
	case "windows":
		// Obtain a list of all currently mounted volumes.
		cmd := executeCommand(options, "wmic",
			"PATH", "Win32_LogicalDisk",
			"get", "DeviceID,VolumeName,FileSystem,DriveType")
		var out bytes.Buffer
		cmd.Stdout = &out
		err := cmd.Run()
		if err != nil {
			return nil, fmt.Errorf("could not list mount points: %w", err)
		}

		// Extract data to convert to a []mountPoint slice.
		for _, line := range strings.Split(out.String(), "\n") {
			words := strings.Fields(line)
			if len(words) < 3 {
				continue
			}
			if words[1] != "2" || words[2] != "FAT" {
				// - DriveType 2 is removable (which we're looking for).
				// - We only want to return FAT filesystems.
				continue
			}
			points = append(points, mountPoint{
				name: words[3],
				path: words[0],
			})
		}
		return points, nil
	default:
		return nil, fmt.Errorf("unknown GOOS for listing mount points: %s", runtime.GOOS)
	}
}

// getDefaultPort returns the default serial port depending on the operating system.
func getDefaultPort(portFlag string, usbInterfaces []string) (port string, err error) {
	portCandidates := strings.FieldsFunc(portFlag, func(c rune) bool { return c == ',' })
	if len(portCandidates) == 1 {
		return portCandidates[0], nil
	}

	var ports []string
	switch runtime.GOOS {
	case "freebsd":
		ports, err = filepath.Glob("/dev/cuaU*")
	case "darwin", "linux", "windows":
		var portsList []*enumerator.PortDetails
		portsList, err = enumerator.GetDetailedPortsList()
		if err != nil {
			return "", err
		}

		var preferredPortIDs [][2]uint16
		for _, s := range usbInterfaces {
			parts := strings.Split(s, ":")
			if len(parts) != 2 {
				return "", fmt.Errorf("could not parse USB VID/PID pair %q", s)
			}
			vid, err := strconv.ParseUint(parts[0], 16, 16)
			if err != nil {
				return "", fmt.Errorf("could not parse USB vendor ID %q: %w", parts[1], err)
			}
			pid, err := strconv.ParseUint(parts[1], 16, 16)
			if err != nil {
				return "", fmt.Errorf("could not parse USB product ID %q: %w", parts[1], err)
			}
			preferredPortIDs = append(preferredPortIDs, [2]uint16{uint16(vid), uint16(pid)})
		}

		var primaryPorts []string   // ports picked from preferred USB VID/PID
		var secondaryPorts []string // other ports (as a fallback)
		for _, p := range portsList {
			if !p.IsUSB {
				continue
			}
			if p.VID != "" && p.PID != "" {
				foundPort := false
				vid, vidErr := strconv.ParseUint(p.VID, 16, 16)
				pid, pidErr := strconv.ParseUint(p.PID, 16, 16)
				if vidErr == nil && pidErr == nil {
					for _, id := range preferredPortIDs {
						if uint16(vid) == id[0] && uint16(pid) == id[1] {
							primaryPorts = append(primaryPorts, p.Name)
							foundPort = true
							continue
						}
					}
				}
				if foundPort {
					continue
				}
			}

			secondaryPorts = append(secondaryPorts, p.Name)
		}
		if len(primaryPorts) == 1 {
			// There is exactly one match in the set of preferred ports. Use
			// this port, even if there may be others available. This allows
			// flashing a specific board even if there are multiple available.
			return primaryPorts[0], nil
		} else if len(primaryPorts) > 1 {
			// There are multiple preferred ports, probably because more than
			// one device of the same type are connected (e.g. two Arduino
			// Unos).
			ports = primaryPorts
		} else {
			// No preferred ports found. Fall back to other serial ports
			// available in the system.
			ports = secondaryPorts
		}
	default:
		return "", errors.New("unable to search for a default USB device to be flashed on this OS")
	}

	if err != nil {
		return "", err
	} else if ports == nil {
		return "", errors.New("unable to locate a serial port")
	} else if len(ports) == 0 {
		return "", errors.New("no serial ports available")
	}

	if len(portCandidates) == 0 {
		if len(usbInterfaces) > 0 {
			return "", errors.New("unable to search for a default USB device - use -port flag, available ports are " + strings.Join(ports, ", "))
		} else if len(ports) == 1 {
			return ports[0], nil
		} else {
			return "", errors.New("multiple serial ports available - use -port flag, available ports are " + strings.Join(ports, ", "))
		}
	}

	for _, ps := range portCandidates {
		for _, p := range ports {
			if p == ps {
				return p, nil
			}
		}
	}

	return "", errors.New("port you specified '" + strings.Join(portCandidates, ",") + "' does not exist, available ports are " + strings.Join(ports, ", "))
}

// getBMPPorts returns BlackMagicProbe's serial ports if any
func getBMPPorts() (gdbPort, uartPort string, err error) {
	var portsList []*enumerator.PortDetails
	portsList, err = enumerator.GetDetailedPortsList()
	if err != nil {
		return "", "", err
	}
	var ports []string
	for _, p := range portsList {
		if !p.IsUSB {
			continue
		}
		if p.VID != "" && p.PID != "" {
			vid, vidErr := strconv.ParseUint(p.VID, 16, 16)
			pid, pidErr := strconv.ParseUint(p.PID, 16, 16)
			if vidErr == nil && pidErr == nil && vid == 0x1d50 && pid == 0x6018 {
				ports = append(ports, p.Name)
			}
		}
	}
	if len(ports) == 2 {
		return ports[0], ports[1], nil
	} else if len(ports) == 0 {
		return "", "", errors.New("no BMP detected")
	} else {
		return "", "", fmt.Errorf("expected 2 BMP serial ports, found %d - did you perhaps connect more than one BMP?", len(ports))
	}
}

const (
	usageBuild = `Build compiles the packages named by the import paths, along with their
dependencies, but it does not install the results. The output binary is
specified using the -o parameter. The generated file type depends on the
extension:

	.o:
			Create a relocatable object file. You can use this option if you
			don't want to use the TinyGo build system or want to do other custom
			things.

	.ll:
			Create textual LLVM IR, after optimization. This is mainly useful
			for debugging.

	.bc:
			Create LLVM bitcode, after optimization. This may be useful for
			debugging or for linking into other programs using LTO.

	.hex:
			Create an Intel HEX file to flash it to a microcontroller.

	.bin:
			Similar, but create a binary file.

	.wasm:
			Compile and link a WebAssembly file.

(all other) Compile and link the program into a regular executable. For
microcontrollers, it is common to use the .elf file extension to indicate a
linked ELF file is generated. For Linux, it is common to build binaries with no
extension at all.`

	usageRun = `Run the program, either directly on the host or in an emulated environment 
(depending on -target).`

	usageFlash = `Flash the program to a microcontroller. Some common flags are described below.

	-target={name}: 
			Specifies the type of microcontroller that is used. The name of the
			microcontroller is given on the individual pages for each board type
			listed under Microcontrollers
			(https://tinygo.org/docs/reference/microcontrollers/).
			Examples: "arduino-nano", "d1mini", "xiao".

	-monitor: 
			Start the serial monitor (see below) immediately after
			flashing. However, some microcontrollers need a split second
			or two to configure the serial port after flashing, and
			using the "-monitor" flag can fail because the serial
			monitor starts too quickly. In that case, use the "tinygo
			monitor" command explicitly.`

	usageMonitor = `Start the serial monitor on the serial port that is connected to the
microcontroller. If there is only a single board attached to the host computer,
the default values for various options should be sufficient. In other
situations, particularly if you have multiple microcontrollers attached, some
parameters may need to be overridden using the following flags:

	-port={port}:
			If there are multiple microcontroller attached, an error
			message will display a list of potential serial ports. The
			appropriate port can be specified by this flag. On Linux,
			the port will be something like /dev/ttyUSB0 or /dev/ttyACM1.
			On MacOS, the port will look like /dev/cu.usbserial-1420. On
			Windows, the port will be something like COM1 or COM31.

	-baudrate={rate}:
			The default baud rate is 115200. Boards using the AVR
			processor (e.g. Arduino Nano, Arduino Mega 2560) use 9600
			instead.

	-target={name}:
			If you have more than one microcontrollers attached, you can
			sometimes just specify the target name and let tinygo
			monitor figure out the port. Sometimes, this does not work
			and you have to explicitly use the -port flag.

The serial monitor intercepts several control characters for its own use instead of sending them
to the microcontroller:

	Control-C: terminates the tinygo monitor
	Control-Z: suspends the tinygo monitor and drops back into shell
	Control-\: terminates the tinygo monitor with a stack trace
	Control-S: flow control, suspends output to the console
	Control-Q: flow control, resumes output to the console
	Control-@: thrown away by tinygo monitor

Note: If you are using os.Stdin on the microcontroller, you may find that a CR
character on the host computer (also known as Enter, ^M, or \r) is transmitted
to the microcontroller without conversion, so os.Stdin returns a \r character
instead of the expected \n (also known as ^J, NL, or LF) to indicate
end-of-line. You may be able to get around this problem by hitting Control-J in
tinygo monitor to transmit the \n end-of-line character.`

	usageGdb = `Build the program, optionally flash it to a microcontroller if it is a remote 
target, and drop into a GDB shell. From there you can set breakpoints, start the
program with "run" or "continue" ("run" for a local program, continue for
on-chip debugging), single-step, show a backtrace, break and resume the program
with Ctrl-C/"continue", etc. You may need to install extra tools (like openocd
and arm-none-eabi-gdb) to be able to do this. Also, you may need a dedicated
debugger to be able to debug certain boards if no debugger is integrated. Some
boards (like the BBC micro:bit and most professional evaluation boards) have an
integrated debugger.`

	usageClean = `Clean the cache directory, normally stored in $HOME/.cache/tinygo. This is not
normally needed.`

	usageHelp    = `Print a short summary of the available commands, plus a list of command flags.`
	usageVersion = `Print the version of the command and the version of the used $GOROOT.`
	usageEnv     = `Print a list of environment variables that affect TinyGo (as a shell script).
If one or more variable names are given as arguments, env prints the value of
each on a new line.`

	usageDefault = `TinyGo is a Go compiler for small places.
version: %s
usage: %s <command> [arguments]
commands:
		build:		compile packages and dependencies
		run:		compile and run immediately
		test:		test packages
		flash:		compile and flash to the device
		gdb:		run/flash and immediately enter GDB
		lldb:		run/flash and immediately enter LLDB
		monitor:	open communication port
		ports:		list available serial ports
		env:		list environment variables used during build
		list:		run go list using the TinyGo root
		clean:		empty cache directory (%s)
		targets:	list targets
		info:		show info for specified target
		version:	show version
		help:		print this help text`
)

var (
	commandHelp = map[string]string{
		"build":   usageBuild,
		"run":     usageRun,
		"flash":   usageFlash,
		"monitor": usageMonitor,
		"gdb":     usageGdb,
		"clean":   usageClean,
		"help":    usageHelp,
		"version": usageVersion,
		"env":     usageEnv,
	}
)

func usage(command string) {
	val, ok := commandHelp[command]
	if !ok {
		fmt.Fprintf(os.Stderr, usageDefault, goenv.Version(), os.Args[0], goenv.Get("GOCACHE"))
		if flag.Parsed() {
			fmt.Fprintln(os.Stderr, "\nflags:")
			flag.PrintDefaults()
		}

		fmt.Fprintln(os.Stderr, "\nfor more details, see https://tinygo.org/docs/reference/usage/")
	} else {
		fmt.Fprintln(os.Stderr, val)
	}

}

func handleCompilerError(err error) {
	if err != nil {
		wd, getwdErr := os.Getwd()
		if getwdErr != nil {
			wd = ""
		}
		diagnostics.CreateDiagnostics(err).WriteTo(os.Stderr, wd)
		os.Exit(1)
	}
}

// This is a special type for the -X flag to parse the pkgpath.Var=stringVal
// format. It has to be a special type to allow multiple variables to be defined
// this way.
type globalValuesFlag map[string]map[string]string

func (m globalValuesFlag) String() string {
	return "pkgpath.Var=value"
}

func (m globalValuesFlag) Set(value string) error {
	equalsIndex := strings.IndexByte(value, '=')
	if equalsIndex < 0 {
		return errors.New("expected format pkgpath.Var=value")
	}
	pathAndName := value[:equalsIndex]
	pointIndex := strings.LastIndexByte(pathAndName, '.')
	if pointIndex < 0 {
		return errors.New("expected format pkgpath.Var=value")
	}
	path := pathAndName[:pointIndex]
	name := pathAndName[pointIndex+1:]
	stringValue := value[equalsIndex+1:]
	if m[path] == nil {
		m[path] = make(map[string]string)
	}
	m[path][name] = stringValue
	return nil
}

// parseGoLinkFlag parses the -ldflags parameter. Its primary purpose right now
// is the -X flag, for setting the value of global string variables.
func parseGoLinkFlag(flagsString string) (map[string]map[string]string, string, error) {
	set := flag.NewFlagSet("link", flag.ExitOnError)
	globalVarValues := make(globalValuesFlag)
	set.Var(globalVarValues, "X", "Set the value of the string variable to the given value.")
	extLDFlags := set.String("extldflags", "", "additional flags to pass to external linker")
	flags, err := shlex.Split(flagsString)
	if err != nil {
		return nil, "", err
	}
	err = set.Parse(flags)
	if err != nil {
		return nil, "", err
	}
	return map[string]map[string]string(globalVarValues), *extLDFlags, nil
}

// getListOfPackages returns a standard list of packages for a given list that might
// include wildards using `go list`.
// For example [./...] => ["pkg1", "pkg1/pkg12", "pkg2"]
func getListOfPackages(pkgs []string, options *compileopts.Options) ([]string, error) {
	config, err := builder.NewConfig(options)
	if err != nil {
		return nil, err
	}
	cmd, err := loader.List(config, nil, pkgs)
	if err != nil {
		return nil, fmt.Errorf("failed to run `go list`: %w", err)
	}
	outputBuf := bytes.NewBuffer(nil)
	cmd.Stdout = outputBuf
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return nil, err
	}

	var pkgNames []string
	sc := bufio.NewScanner(outputBuf)
	for sc.Scan() {
		pkgNames = append(pkgNames, sc.Text())
	}

	return pkgNames, nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "No command-line arguments supplied.")
		usage("")
		os.Exit(1)
	}
	command := os.Args[1]

	opt := flag.String("opt", "z", "optimization level: 0, 1, 2, s, z")
	gc := flag.String("gc", "", "garbage collector to use (none, leaking, conservative)")
	panicStrategy := flag.String("panic", "print", "panic strategy (print, trap)")
	scheduler := flag.String("scheduler", "", "which scheduler to use (none, tasks, asyncify)")
	serial := flag.String("serial", "", "which serial output to use (none, uart, usb, rtt)")
	work := flag.Bool("work", false, "print the name of the temporary build directory and do not delete this directory on exit")
	interpTimeout := flag.Duration("interp-timeout", 180*time.Second, "interp optimization pass timeout")
	var tags buildutil.TagsFlag
	flag.Var(&tags, "tags", "a space-separated list of extra build tags")
	target := flag.String("target", "", "chip/board name or JSON target specification file")
	buildMode := flag.String("buildmode", "", "build mode to use (default, c-shared)")
	var stackSize uint64
	flag.Func("stack-size", "goroutine stack size (if unknown at compile time)", func(s string) error {
		size, err := bytesize.Parse(s)
		stackSize = uint64(size)
		return err
	})
	printSize := flag.String("size", "", "print sizes (none, short, full)")
	printStacks := flag.Bool("print-stacks", false, "print stack sizes of goroutines")
	printAllocsString := flag.String("print-allocs", "", "regular expression of functions for which heap allocations should be printed")
	printCommands := flag.Bool("x", false, "Print commands")
	parallelism := flag.Int("p", runtime.GOMAXPROCS(0), "the number of build jobs that can run in parallel")
	nodebug := flag.Bool("no-debug", false, "strip debug information")
	ocdCommandsString := flag.String("ocd-commands", "", "OpenOCD commands, overriding target spec (can specify multiple separated by commas)")
	ocdOutput := flag.Bool("ocd-output", false, "print OCD daemon output during debug")
	port := flag.String("port", "", "flash port (can specify multiple candidates separated by commas)")
	timeout := flag.Duration("timeout", 20*time.Second, "the length of time to retry locating the MSD volume to be used for flashing")
	programmer := flag.String("programmer", "", "which hardware programmer to use")
	ldflags := flag.String("ldflags", "", "Go link tool compatible ldflags")
	llvmFeatures := flag.String("llvm-features", "", "comma separated LLVM features to enable")
	cpuprofile := flag.String("cpuprofile", "", "cpuprofile output")
	monitor := flag.Bool("monitor", false, "enable serial monitor")
	baudrate := flag.Int("baudrate", 115200, "baudrate of serial monitor")

	// Internal flags, that are only intended for TinyGo development.
	printIR := flag.Bool("internal-printir", false, "print LLVM IR")
	dumpSSA := flag.Bool("internal-dumpssa", false, "dump internal Go SSA")
	verifyIR := flag.Bool("internal-verifyir", false, "run extra verification steps on LLVM IR")
	// Don't generate debug information in the IR, to make IR more readable.
	// You generally want debug information in IR for various features, like
	// stack size calculation and features like -size=short, -print-allocs=,
	// etc. The -no-debug flag is used to strip it at link time. But for TinyGo
	// development it can be useful to not emit debug information at all.
	skipDwarf := flag.Bool("internal-nodwarf", false, "internal flag, use -no-debug instead")

	var flagJSON, flagDeps, flagTest bool
	if command == "help" || command == "list" || command == "info" || command == "build" {
		flag.BoolVar(&flagJSON, "json", false, "print data in JSON format")
	}
	if command == "help" || command == "list" {
		flag.BoolVar(&flagDeps, "deps", false, "supply -deps flag to go list")
		flag.BoolVar(&flagTest, "test", false, "supply -test flag to go list")
	}
	var outpath string
	if command == "help" || command == "build" || command == "test" {
		flag.StringVar(&outpath, "o", "", "output filename")
	}

	var witPackage, witWorld string
	if command == "help" || command == "build" || command == "test" || command == "run" {
		flag.StringVar(&witPackage, "wit-package", "", "wit package for wasm component embedding")
		flag.StringVar(&witWorld, "wit-world", "", "wit world for wasm component embedding")
	}

	var testConfig compileopts.TestConfig
	if command == "help" || command == "test" {
		flag.BoolVar(&testConfig.CompileOnly, "c", false, "compile the test binary but do not run it")
		flag.BoolVar(&testConfig.Verbose, "v", false, "verbose: print additional output")
		flag.BoolVar(&testConfig.Short, "short", false, "short: run smaller test suite to save time")
		flag.StringVar(&testConfig.RunRegexp, "run", "", "run: regexp of tests to run")
		flag.StringVar(&testConfig.SkipRegexp, "skip", "", "skip: regexp of tests to skip")
		testConfig.Count = flag.Int("count", 1, "count: number of times to run tests/benchmarks `count` times")
		flag.StringVar(&testConfig.BenchRegexp, "bench", "", "bench: regexp of benchmarks to run")
		flag.StringVar(&testConfig.BenchTime, "benchtime", "", "run each benchmark for duration `d`")
		flag.BoolVar(&testConfig.BenchMem, "benchmem", false, "show memory stats for benchmarks")
		flag.StringVar(&testConfig.Shuffle, "shuffle", "", "shuffle the order the tests and benchmarks run")
	}

	// Early command processing, before commands are interpreted by the Go flag
	// library.
	handleChdirFlag()
	switch command {
	case "clang", "ld.lld", "wasm-ld":
		err := builder.RunTool(command, os.Args[2:]...)
		if err != nil {
			// The tool should have printed an error message already.
			// Don't print another error message here.
			os.Exit(1)
		}
		os.Exit(0)
	}

	flag.CommandLine.Parse(os.Args[2:])
	globalVarValues, extLDFlags, err := parseGoLinkFlag(*ldflags)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var printAllocs *regexp.Regexp
	if *printAllocsString != "" {
		printAllocs, err = regexp.Compile(*printAllocsString)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	var ocdCommands []string
	if *ocdCommandsString != "" {
		ocdCommands = strings.Split(*ocdCommandsString, ",")
	}

	options := &compileopts.Options{
		GOOS:            goenv.Get("GOOS"),
		GOARCH:          goenv.Get("GOARCH"),
		GOARM:           goenv.Get("GOARM"),
		GOMIPS:          goenv.Get("GOMIPS"),
		Target:          *target,
		BuildMode:       *buildMode,
		StackSize:       stackSize,
		Opt:             *opt,
		GC:              *gc,
		PanicStrategy:   *panicStrategy,
		Scheduler:       *scheduler,
		Serial:          *serial,
		Work:            *work,
		InterpTimeout:   *interpTimeout,
		PrintIR:         *printIR,
		DumpSSA:         *dumpSSA,
		VerifyIR:        *verifyIR,
		SkipDWARF:       *skipDwarf,
		Semaphore:       make(chan struct{}, *parallelism),
		Debug:           !*nodebug,
		PrintSizes:      *printSize,
		PrintStacks:     *printStacks,
		PrintAllocs:     printAllocs,
		Tags:            []string(tags),
		TestConfig:      testConfig,
		GlobalValues:    globalVarValues,
		Programmer:      *programmer,
		OpenOCDCommands: ocdCommands,
		LLVMFeatures:    *llvmFeatures,
		PrintJSON:       flagJSON,
		Monitor:         *monitor,
		BaudRate:        *baudrate,
		Timeout:         *timeout,
		WITPackage:      witPackage,
		WITWorld:        witWorld,
		ExtLDFlags:      extLDFlags,
	}
	if *printCommands {
		options.PrintCommands = printCommand
	}

	err = options.Verify()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		usage(command)
		os.Exit(1)
	}

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			fmt.Fprintln(os.Stderr, "could not create CPU profile: ", err)
			os.Exit(1)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			fmt.Fprintln(os.Stderr, "could not start CPU profile: ", err)
			os.Exit(1)
		}
		defer pprof.StopCPUProfile()
	}

	switch command {
	case "build":
		pkgName := "."
		if flag.NArg() == 1 {
			pkgName = filepath.ToSlash(flag.Arg(0))
		} else if flag.NArg() > 1 {
			fmt.Fprintln(os.Stderr, "build only accepts a single positional argument: package name, but multiple were specified")
			usage(command)
			os.Exit(1)
		}

		err := Build(pkgName, outpath, options)
		handleCompilerError(err)
	case "flash", "gdb", "lldb":
		pkgName := filepath.ToSlash(flag.Arg(0))
		if command == "flash" {
			err := Flash(pkgName, *port, options)
			handleCompilerError(err)
		} else {
			if !options.Debug {
				fmt.Fprintln(os.Stderr, "Debug disabled while running debugger?")
				usage(command)
				os.Exit(1)
			}
			err := Debug(command, pkgName, *ocdOutput, options)
			handleCompilerError(err)
		}
	case "run":
		if flag.NArg() < 1 {
			fmt.Fprintln(os.Stderr, "No package specified.")
			usage(command)
			os.Exit(1)
		}
		pkgName := filepath.ToSlash(flag.Arg(0))
		err := Run(pkgName, options, flag.Args()[1:])
		handleCompilerError(err)
	case "test":
		var pkgNames []string
		for i := 0; i < flag.NArg(); i++ {
			pkgNames = append(pkgNames, filepath.ToSlash(flag.Arg(i)))
		}
		if len(pkgNames) == 0 {
			pkgNames = []string{"."}
		}

		explicitPkgNames, err := getListOfPackages(pkgNames, options)
		if err != nil {
			fmt.Printf("cannot resolve packages: %v\n", err)
			os.Exit(1)
		}

		if outpath != "" && len(explicitPkgNames) > 1 {
			fmt.Println("cannot use -o flag with multiple packages")
			os.Exit(1)
		}

		fail := make(chan struct{}, 1)
		var wg sync.WaitGroup
		bufs := make([]testOutputBuf, len(explicitPkgNames))
		for i := range bufs {
			bufs[i].done = make(chan struct{})
		}

		wg.Add(1)
		go func() {
			defer wg.Done()

			// Flush the output one test at a time.
			// This ensures that outputs from different tests are not mixed together.
			for i := range bufs {
				err := bufs[i].flush(os.Stdout, os.Stderr)
				if err != nil {
					// There was an error writing to stdout or stderr, so we probably cannot print this.
					select {
					case fail <- struct{}{}:
					default:
					}
				}
			}
		}()

		// Build and run the tests concurrently.
		// This uses an additional semaphore to reduce the memory usage.
		testSema := make(chan struct{}, cap(options.Semaphore))
		for i, pkgName := range explicitPkgNames {
			pkgName := pkgName
			buf := &bufs[i]
			testSema <- struct{}{}
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer func() { <-testSema }()
				defer close(buf.done)
				stdout := (*testStdout)(buf)
				stderr := (*testStderr)(buf)
				passed, err := Test(pkgName, stdout, stderr, options, outpath)
				if err != nil {
					wd, getwdErr := os.Getwd()
					if getwdErr != nil {
						wd = ""
					}
					diagnostics.CreateDiagnostics(err).WriteTo(os.Stderr, wd)
				}
				if !passed {
					select {
					case fail <- struct{}{}:
					default:
					}
				}
			}()
		}

		// Wait for all tests to finish.
		wg.Wait()
		close(fail)
		if _, fail := <-fail; fail {
			os.Exit(1)
		}
	case "monitor":
		config, err := builder.NewConfig(options)
		handleCompilerError(err)
		err = Monitor("", *port, config)
		handleCompilerError(err)
	case "ports":
		serialPortInfo, err := ListSerialPorts()
		handleCompilerError(err)
		if len(serialPortInfo) == 0 {
			fmt.Println("No serial ports found.")
		}
		fmt.Printf("%-20s %-9s %s\n", "Port", "ID", "Boards")
		for _, s := range serialPortInfo {
			fmt.Printf("%-20s %4s:%4s %s\n", s.Name, s.VID, s.PID, s.Target)
		}
	case "targets":
		specs, err := compileopts.GetTargetSpecs()
		if err != nil {
			fmt.Fprintln(os.Stderr, "could not list targets:", err)
			os.Exit(1)
			return
		}
		names := []string{}
		for key := range specs {
			names = append(names, key)
		}
		sort.Strings(names)
		for _, name := range names {
			fmt.Println(name)
		}
	case "info":
		if flag.NArg() == 1 {
			options.Target = flag.Arg(0)
		} else if flag.NArg() > 1 {
			fmt.Fprintln(os.Stderr, "only one target name is accepted")
			usage(command)
			os.Exit(1)
		}
		config, err := builder.NewConfig(options)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			usage(command)
			os.Exit(1)
		}
		config.GoMinorVersion = 0 // this avoids creating the list of Go1.x build tags.
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		cachedGOROOT, err := loader.GetCachedGoroot(config)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if flagJSON {
			json, _ := json.MarshalIndent(struct {
				Target     *compileopts.TargetSpec `json:"target"`
				GOROOT     string                  `json:"goroot"`
				GOOS       string                  `json:"goos"`
				GOARCH     string                  `json:"goarch"`
				GOARM      string                  `json:"goarm"`
				GOMIPS     string                  `json:"gomips"`
				BuildTags  []string                `json:"build_tags"`
				GC         string                  `json:"garbage_collector"`
				Scheduler  string                  `json:"scheduler"`
				LLVMTriple string                  `json:"llvm_triple"`
			}{
				Target:     config.Target,
				GOROOT:     cachedGOROOT,
				GOOS:       config.GOOS(),
				GOARCH:     config.GOARCH(),
				GOARM:      config.GOARM(),
				GOMIPS:     config.GOMIPS(),
				BuildTags:  config.BuildTags(),
				GC:         config.GC(),
				Scheduler:  config.Scheduler(),
				LLVMTriple: config.Triple(),
			}, "", "  ")
			fmt.Println(string(json))
		} else {
			fmt.Printf("LLVM triple:       %s\n", config.Triple())
			fmt.Printf("GOOS:              %s\n", config.GOOS())
			fmt.Printf("GOARCH:            %s\n", config.GOARCH())
			fmt.Printf("build tags:        %s\n", strings.Join(config.BuildTags(), " "))
			fmt.Printf("garbage collector: %s\n", config.GC())
			fmt.Printf("scheduler:         %s\n", config.Scheduler())
			fmt.Printf("cached GOROOT:     %s\n", cachedGOROOT)
		}
	case "list":
		config, err := builder.NewConfig(options)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			usage(command)
			os.Exit(1)
		}
		var extraArgs []string
		if flagJSON {
			extraArgs = append(extraArgs, "-json")
		}
		if flagDeps {
			extraArgs = append(extraArgs, "-deps")
		}
		if flagTest {
			extraArgs = append(extraArgs, "-test")
		}
		cmd, err := loader.List(config, extraArgs, flag.Args())
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed to run `go list`:", err)
			os.Exit(1)
		}
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				os.Exit(exitErr.ExitCode())
			}
			fmt.Fprintln(os.Stderr, "failed to run `go list`:", err)
			os.Exit(1)
		}
	case "clean":
		// remove cache directory
		err := os.RemoveAll(goenv.Get("GOCACHE"))
		if err != nil {
			fmt.Fprintln(os.Stderr, "cannot clean cache:", err)
			os.Exit(1)
		}
	case "help":
		command := ""
		if flag.NArg() >= 1 {
			command = flag.Arg(0)
		}
		usage(command)
	case "version":
		goversion := "<unknown>"
		if s, err := goenv.GorootVersionString(); err == nil {
			goversion = s
		}
		fmt.Printf("tinygo version %s %s/%s (using go version %s and LLVM version %s)\n", goenv.Version(), runtime.GOOS, runtime.GOARCH, goversion, llvm.Version)
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
		usage("")
		os.Exit(1)
	}
}

// testOutputBuf is used to buffer the output of concurrent tests.
type testOutputBuf struct {
	mu             sync.Mutex
	output         []outputEntry
	stdout, stderr io.Writer
	outerr, errerr error
	done           chan struct{}
}

// flush the output to stdout and stderr.
// This waits until done is closed.
func (b *testOutputBuf) flush(stdout, stderr io.Writer) error {
	b.mu.Lock()

	var err error
	b.stdout = stdout
	b.stderr = stderr
	for _, e := range b.output {
		var w io.Writer
		var errDst *error
		if e.stderr {
			w = stderr
			errDst = &b.errerr
		} else {
			w = stdout
			errDst = &b.outerr
		}
		if *errDst != nil {
			continue
		}

		_, werr := w.Write(e.data)
		if werr != nil {
			if err == nil {
				err = werr
			}
			*errDst = err
		}
	}

	b.mu.Unlock()

	<-b.done

	return err
}

// testStdout writes stdout from a test to the output buffer.
type testStdout testOutputBuf

func (out *testStdout) Write(data []byte) (int, error) {
	buf := (*testOutputBuf)(out)
	buf.mu.Lock()

	if buf.stdout != nil {
		// Write the output directly.
		err := out.outerr
		buf.mu.Unlock()
		if err != nil {
			return 0, err
		}
		return buf.stdout.Write(data)
	}

	defer buf.mu.Unlock()

	// Append the output.
	if len(buf.output) == 0 || buf.output[len(buf.output)-1].stderr {
		buf.output = append(buf.output, outputEntry{
			stderr: false,
		})
	}
	last := &buf.output[len(buf.output)-1]
	last.data = append(last.data, data...)

	return len(data), nil
}

// testStderr writes stderr from a test to the output buffer.
type testStderr testOutputBuf

func (out *testStderr) Write(data []byte) (int, error) {
	buf := (*testOutputBuf)(out)
	buf.mu.Lock()

	if buf.stderr != nil {
		// Write the output directly.
		err := out.errerr
		buf.mu.Unlock()
		if err != nil {
			return 0, err
		}
		return buf.stderr.Write(data)
	}

	defer buf.mu.Unlock()

	// Append the output.
	if len(buf.output) == 0 || !buf.output[len(buf.output)-1].stderr {
		buf.output = append(buf.output, outputEntry{
			stderr: true,
		})
	}
	last := &buf.output[len(buf.output)-1]
	last.data = append(last.data, data...)

	return len(data), nil
}

type outputEntry struct {
	stderr bool
	data   []byte
}

// handleChdirFlag handles the -C flag before doing anything else.
// The -C flag must be the first flag on the command line, to make it easy to find
// even with commands that have custom flag parsing.
// handleChdirFlag handles the flag by chdir'ing to the directory
// and then removing that flag from the command line entirely.
//
// We have to handle the -C flag this way for two reasons:
//
//  1. Toolchain selection needs to be in the right directory to look for go.mod and go.work.
//
//  2. A toolchain switch later on reinvokes the new go command with the same arguments.
//     The parent toolchain has already done the chdir; the child must not try to do it again.

func handleChdirFlag() {
	used := 2 // b.c. command at os.Args[1]
	if used >= len(os.Args) {
		return
	}

	var dir string
	switch a := os.Args[used]; {
	default:
		return

	case a == "-C", a == "--C":
		if used+1 >= len(os.Args) {
			return
		}
		dir = os.Args[used+1]
		os.Args = slicesDelete(os.Args, used, used+2)

	case strings.HasPrefix(a, "-C="), strings.HasPrefix(a, "--C="):
		_, dir, _ = strings.Cut(a, "=")
		os.Args = slicesDelete(os.Args, used, used+1)
	}

	if err := os.Chdir(dir); err != nil {
		fmt.Fprintln(os.Stderr, "cannot chdir:", err)
		os.Exit(1)
	}
}

// go1.19 compatibility: lacks slices package
func slicesDelete[S ~[]E, E any](s S, i, j int) S {
	_ = s[i:j:len(s)] // bounds check

	if i == j {
		return s
	}

	return append(s[:i], s[j:]...)
}
