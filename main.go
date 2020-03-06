package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"go/scanner"
	"go/types"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/tinygo-org/tinygo/builder"
	"github.com/tinygo-org/tinygo/compileopts"
	"github.com/tinygo-org/tinygo/goenv"
	"github.com/tinygo-org/tinygo/interp"
	"github.com/tinygo-org/tinygo/loader"
	"tinygo.org/x/go-llvm"

	"go.bug.st/serial"
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

// copyFile copies the given file from src to dst. It copies first to a .tmp
// file which is then moved over a possibly already existing file at the
// destination.
func copyFile(src, dst string) error {
	inf, err := os.Open(src)
	if err != nil {
		return err
	}
	defer inf.Close()
	outpath := dst + ".tmp"
	outf, err := os.Create(outpath)
	if err != nil {
		return err
	}

	_, err = io.Copy(outf, inf)
	if err != nil {
		os.Remove(outpath)
		return err
	}

	err = outf.Close()
	if err != nil {
		return err
	}

	return os.Rename(dst+".tmp", dst)
}

// Build compiles and links the given package and writes it to outpath.
func Build(pkgName, outpath string, options *compileopts.Options) error {
	config, err := builder.NewConfig(options)
	if err != nil {
		return err
	}

	return builder.Build(pkgName, outpath, config, func(tmppath string) error {
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

// Test runs the tests in the given package.
func Test(pkgName string, options *compileopts.Options) error {
	options.TestConfig.CompileTestBinary = true
	config, err := builder.NewConfig(options)
	if err != nil {
		return err
	}

	// Add test build tag. This is incorrect: `go test` only looks at the
	// _test.go file suffix but does not add the test build tag in the process.
	// However, it's a simple fix right now.
	// For details: https://github.com/golang/go/issues/21360
	config.Target.BuildTags = append(config.Target.BuildTags, "test")

	return builder.Build(pkgName, ".elf", config, func(tmppath string) error {
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
	case "native":
		return errors.New("unknown flash method \"native\" - did you miss a -target flag?")
	default:
		return errors.New("unknown flash method: " + flashMethod)
	}

	return builder.Build(pkgName, fileExt, config, func(tmppath string) error {
		// do we need port reset to put MCU into bootloader mode?
		if config.Target.PortReset == "true" {
			if port == "" {
				var err error
				port, err = getDefaultPort()
				if err != nil {
					return err
				}
			}

			err := touchSerialPortAt1200bps(port)
			if err != nil {
				return &commandError{"failed to reset port", tmppath, err}
			}
			// give the target MCU a chance to restart into bootloader
			time.Sleep(3 * time.Second)
		}

		// this flashing method copies the binary data to a Mass Storage Device (msd)
		switch flashMethod {
		case "", "command":
			// Create the command.
			flashCmd := config.Target.FlashCommand
			fileToken := "{" + fileExt[1:] + "}"
			flashCmd = strings.Replace(flashCmd, fileToken, tmppath, -1)

			if port == "" && strings.Contains(flashCmd, "{port}") {
				var err error
				port, err = getDefaultPort()
				if err != nil {
					return err
				}
			}

			flashCmd = strings.Replace(flashCmd, "{port}", port, -1)

			// Execute the command.
			var cmd *exec.Cmd
			switch runtime.GOOS {
			case "windows":
				command := strings.Split(flashCmd, " ")
				if len(command) < 2 {
					return errors.New("invalid flash command")
				}
				cmd = exec.Command(command[0], command[1:]...)
			default:
				cmd = exec.Command("/bin/sh", "-c", flashCmd)
			}

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
				err := flashUF2UsingMSD(config.Target.FlashVolume, tmppath)
				if err != nil {
					return &commandError{"failed to flash", tmppath, err}
				}
				return nil
			case ".hex":
				err := flashHexUsingMSD(config.Target.FlashVolume, tmppath)
				if err != nil {
					return &commandError{"failed to flash", tmppath, err}
				}
				return nil
			default:
				return errors.New("mass storage device flashing currently only supports uf2 and hex")
			}
		case "openocd":
			args, err := config.OpenOCDConfiguration()
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
			return fmt.Errorf("unknown flash method: %s", flashMethod)
		}
	})
}

// FlashGDB compiles and flashes a program to a microcontroller (just like
// Flash) but instead of resetting the target, it will drop into a GDB shell.
// You can then set breakpoints, run the GDB `continue` command to start, hit
// Ctrl+C to break the running program, etc.
//
// Note: this command is expected to execute just before exiting, as it
// modifies global state.
func FlashGDB(pkgName string, ocdOutput bool, options *compileopts.Options) error {
	config, err := builder.NewConfig(options)
	if err != nil {
		return err
	}
	if config.Target.GDB == "" {
		return errors.New("gdb not configured in the target specification")
	}

	return builder.Build(pkgName, "", config, func(tmppath string) error {
		// Find a good way to run GDB.
		gdbInterface, openocdInterface := config.Programmer()
		switch gdbInterface {
		case "msd", "command", "":
			if len(config.Target.Emulator) != 0 {
				// Assume QEMU as an emulator.
				if config.Target.Emulator[0] == "mgba" {
					gdbInterface = "mgba"
				} else {
					gdbInterface = "qemu"
				}
			} else if openocdInterface != "" && config.Target.OpenOCDTarget != "" {
				gdbInterface = "openocd"
			} else if config.Target.JLinkDevice != "" {
				gdbInterface = "jlink"
			}
		}

		// Run the GDB server, if necessary.
		var gdbCommands []string
		var daemon *exec.Cmd
		switch gdbInterface {
		case "native":
			// Run GDB directly.
		case "openocd":
			gdbCommands = append(gdbCommands, "target remote :3333", "monitor halt", "load", "monitor reset halt")

			// We need a separate debugging daemon for on-chip debugging.
			args, err := config.OpenOCDConfiguration()
			if err != nil {
				return err
			}
			daemon = exec.Command("openocd", args...)
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
		case "jlink":
			gdbCommands = append(gdbCommands, "target remote :2331", "load", "monitor reset halt")

			// We need a separate debugging daemon for on-chip debugging.
			daemon = exec.Command("JLinkGDBServer", "-device", config.Target.JLinkDevice)
			if ocdOutput {
				// Make it clear which output is from the daemon.
				w := &ColorWriter{
					Out:    os.Stderr,
					Prefix: "jlink: ",
					Color:  TermColorYellow,
				}
				daemon.Stdout = w
				daemon.Stderr = w
			}
		case "qemu":
			gdbCommands = append(gdbCommands, "target remote :1234")

			// Run in an emulator.
			args := append(config.Target.Emulator[1:], tmppath, "-s", "-S")
			daemon = exec.Command(config.Target.Emulator[0], args...)
			daemon.Stdout = os.Stdout
			daemon.Stderr = os.Stderr
		case "mgba":
			gdbCommands = append(gdbCommands, "target remote :2345")

			// Run in an emulator.
			args := append(config.Target.Emulator[1:], tmppath, "-g")
			daemon = exec.Command(config.Target.Emulator[0], args...)
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
		for _, cmd := range gdbCommands {
			params = append(params, "-ex", cmd)
		}
		cmd := exec.Command(config.Target.GDB, params...)
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

// Run compiles and runs the given program. Depending on the target provided in
// the options, it will run the program directly on the host or will run it in
// an emulator. For example, -target=wasm will cause the binary to be run inside
// of a WebAssembly VM.
func Run(pkgName string, options *compileopts.Options) error {
	config, err := builder.NewConfig(options)
	if err != nil {
		return err
	}

	return builder.Build(pkgName, ".elf", config, func(tmppath string) error {
		if len(config.Target.Emulator) == 0 {
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
			args := append(config.Target.Emulator[1:], tmppath)
			cmd := exec.Command(config.Target.Emulator[0], args...)
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

func touchSerialPortAt1200bps(port string) (err error) {
	retryCount := 3
	for i := 0; i < retryCount; i++ {
		// Open port
		p, e := serial.Open(port, &serial.Mode{BaudRate: 1200})
		if e != nil {
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

func flashUF2UsingMSD(volume, tmppath string) error {
	// find standard UF2 info path
	var infoPath string
	switch runtime.GOOS {
	case "linux", "freebsd":
		infoPath = "/media/*/" + volume + "/INFO_UF2.TXT"
	case "darwin":
		infoPath = "/Volumes/" + volume + "/INFO_UF2.TXT"
	case "windows":
		path, err := windowsFindUSBDrive(volume)
		if err != nil {
			return err
		}
		infoPath = path + "/INFO_UF2.TXT"
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
	var destPath string
	switch runtime.GOOS {
	case "linux", "freebsd":
		destPath = "/media/*/" + volume
	case "darwin":
		destPath = "/Volumes/" + volume
	case "windows":
		path, err := windowsFindUSBDrive(volume)
		if err != nil {
			return err
		}
		destPath = path + "/"
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

func windowsFindUSBDrive(volume string) (string, error) {
	cmd := exec.Command("wmic",
		"PATH", "Win32_LogicalDisk", "WHERE", "VolumeName = '"+volume+"'",
		"get", "DeviceID,VolumeName,FileSystem,DriveType")

	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}

	for _, line := range strings.Split(out.String(), "\n") {
		words := strings.Fields(line)
		if len(words) >= 3 {
			if words[1] == "2" && words[2] == "FAT" {
				return words[0], nil
			}
		}
	}
	return "", errors.New("unable to locate a USB device to be flashed")
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

// getDefaultPort returns the default serial port depending on the operating system.
func getDefaultPort() (port string, err error) {
	var portPath string
	switch runtime.GOOS {
	case "darwin":
		portPath = "/dev/cu.usb*"
	case "linux":
		portPath = "/dev/ttyACM*"
	case "freebsd":
		portPath = "/dev/cuaU*"
	case "windows":
		cmd := exec.Command("wmic",
			"PATH", "Win32_SerialPort", "WHERE", "Caption LIKE 'USB Serial%'", "GET", "DeviceID")

		var out bytes.Buffer
		cmd.Stdout = &out
		err := cmd.Run()
		if err != nil {
			return "", err
		}

		if out.String() == "No Instance(s) Available." {
			return "", errors.New("no serial ports available")
		}

		for _, line := range strings.Split(out.String(), "\n") {
			words := strings.Fields(line)
			if len(words) == 1 {
				if strings.Contains(words[0], "COM") {
					return words[0], nil
				}
			}
		}
		return "", errors.New("unable to locate a serial port")
	default:
		return "", errors.New("unable to search for a default USB device to be flashed on this OS")
	}

	d, err := filepath.Glob(portPath)
	if err != nil {
		return "", err
	}
	if d == nil {
		return "", errors.New("unable to locate a serial port")
	}

	return d[0], nil
}

// runGoList runs the `go list` command but using the configuration used for
// TinyGo.
func runGoList(config *compileopts.Config, flagJSON, flagDeps bool, pkgs []string) error {
	goroot, err := loader.GetCachedGoroot(config)
	if err != nil {
		return err
	}
	args := []string{"list"}
	if flagJSON {
		args = append(args, "-json")
	}
	if flagDeps {
		args = append(args, "-deps")
	}
	if len(config.BuildTags()) != 0 {
		args = append(args, "-tags", strings.Join(config.BuildTags(), " "))
	}
	args = append(args, pkgs...)
	cgoEnabled := "0"
	if config.CgoEnabled() {
		cgoEnabled = "1"
	}
	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "GOROOT="+goroot, "GOOS="+config.GOOS(), "GOARCH="+config.GOARCH(), "CGO_ENABLED="+cgoEnabled)
	cmd.Run()
	return nil
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
	fmt.Fprintln(os.Stderr, "  list:  run go list using the TinyGo root")
	fmt.Fprintln(os.Stderr, "  clean: empty cache directory ("+goenv.Get("GOCACHE")+")")
	fmt.Fprintln(os.Stderr, "  help:  print this help text")
	fmt.Fprintln(os.Stderr, "\nflags:")
	flag.PrintDefaults()
}

// printCompilerError prints compiler errors using the provided logger function
// (similar to fmt.Println).
//
// There is one exception: interp errors may print to stderr unconditionally due
// to limitations in the LLVM bindings.
func printCompilerError(logln func(...interface{}), err error) {
	switch err := err.(type) {
	case types.Error, scanner.Error:
		logln(err)
	case *interp.Error:
		logln("#", err.ImportPath)
		logln(err.Error())
		if !err.Inst.IsNil() {
			err.Inst.Dump()
			logln()
		}
		if len(err.Traceback) > 0 {
			logln("\ntraceback:")
			for _, line := range err.Traceback {
				logln(line.Pos.String() + ":")
				line.Inst.Dump()
				logln()
			}
		}
	case loader.Errors:
		logln("#", err.Pkg.PkgPath)
		for _, err := range err.Errs {
			logln(err)
		}
	case *builder.MultiError:
		for _, err := range err.Errs {
			logln(err)
		}
	default:
		logln("error:", err)
	}
}

func handleCompilerError(err error) {
	if err != nil {
		printCompilerError(func(args ...interface{}) {
			fmt.Fprintln(os.Stderr, args...)
		}, err)
		os.Exit(1)
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "No command-line arguments supplied.")
		usage()
		os.Exit(1)
	}
	command := os.Args[1]

	outpath := flag.String("o", "", "output filename")
	opt := flag.String("opt", "z", "optimization level: 0, 1, 2, s, z")
	gc := flag.String("gc", "", "garbage collector to use (none, leaking, extalloc, conservative)")
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
	port := flag.String("port", "", "flash port")
	programmer := flag.String("programmer", "", "which hardware programmer to use")
	cFlags := flag.String("cflags", "", "additional cflags for compiler")
	ldFlags := flag.String("ldflags", "", "additional ldflags for linker")
	wasmAbi := flag.String("wasm-abi", "js", "WebAssembly ABI conventions: js (no i64 params) or generic")
	heapSize := flag.String("heap-size", "1M", "default heap size in bytes (only supported by WebAssembly)")

	var flagJSON, flagDeps *bool
	if command == "list" || command == "env" {
		flagJSON = flag.Bool("json", false, "print data in JSON format")
	}
	if command == "list" {
		flagDeps = flag.Bool("deps", false, "")
	}

	// Early command processing, before commands are interpreted by the Go flag
	// library.
	switch command {
	case "clang", "ld.lld", "wasm-ld":
		err := builder.RunTool(command, os.Args[2:]...)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	flag.CommandLine.Parse(os.Args[2:])
	options := &compileopts.Options{
		Target:        *target,
		Opt:           *opt,
		GC:            *gc,
		PanicStrategy: *panicStrategy,
		Scheduler:     *scheduler,
		PrintIR:       *printIR,
		DumpSSA:       *dumpSSA,
		VerifyIR:      *verifyIR,
		Debug:         !*nodebug,
		PrintSizes:    *printSize,
		Tags:          *tags,
		WasmAbi:       *wasmAbi,
		Programmer:    *programmer,
	}

	if *cFlags != "" {
		options.CFlags = strings.Split(*cFlags, " ")
	}

	if *ldFlags != "" {
		options.LDFlags = strings.Split(*ldFlags, " ")
	}

	if *panicStrategy != "print" && *panicStrategy != "trap" {
		fmt.Fprintln(os.Stderr, "Panic strategy must be either print or trap.")
		usage()
		os.Exit(1)
	}

	var err error
	if options.HeapSize, err = parseSize(*heapSize); err != nil {
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
		if options.Target == "" && filepath.Ext(*outpath) == ".wasm" {
			options.Target = "wasm"
		}
		err := Build(pkgName, *outpath, options)
		handleCompilerError(err)
	case "build-library":
		// Note: this command is only meant to be used while making a release!
		if *outpath == "" {
			fmt.Fprintln(os.Stderr, "No output filename supplied (-o).")
			usage()
			os.Exit(1)
		}
		if *target == "" {
			fmt.Fprintln(os.Stderr, "No target (-target).")
		}
		if flag.NArg() != 1 {
			fmt.Fprintf(os.Stderr, "Build-library only accepts exactly one library name as argument, %d given\n", flag.NArg())
			usage()
			os.Exit(1)
		}
		var lib *builder.Library
		switch name := flag.Arg(0); name {
		case "compiler-rt":
			lib = &builder.CompilerRT
		case "picolibc":
			lib = &builder.Picolibc
		default:
			fmt.Fprintf(os.Stderr, "Unknown library: %s\n", name)
			os.Exit(1)
		}
		path, err := lib.Load(*target)
		handleCompilerError(err)
		copyFile(path, *outpath)
	case "flash", "gdb":
		if *outpath != "" {
			fmt.Fprintln(os.Stderr, "Output cannot be specified with the flash command.")
			usage()
			os.Exit(1)
		}
		if command == "flash" {
			err := Flash(flag.Arg(0), *port, options)
			handleCompilerError(err)
		} else {
			if !options.Debug {
				fmt.Fprintln(os.Stderr, "Debug disabled while running gdb?")
				usage()
				os.Exit(1)
			}
			err := FlashGDB(flag.Arg(0), *ocdOutput, options)
			handleCompilerError(err)
		}
	case "run":
		if flag.NArg() != 1 {
			fmt.Fprintln(os.Stderr, "No package specified.")
			usage()
			os.Exit(1)
		}
		err := Run(flag.Arg(0), options)
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
		err := Test(pkgName, options)
		handleCompilerError(err)
	case "info":
		if flag.NArg() == 1 {
			options.Target = flag.Arg(0)
		} else if flag.NArg() > 1 {
			fmt.Fprintln(os.Stderr, "only one target name is accepted")
			usage()
			os.Exit(1)
		}
		config, err := builder.NewConfig(options)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			usage()
			os.Exit(1)
		}
		config.GoMinorVersion = 0 // this avoids creating the list of Go1.x build tags.
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Printf("LLVM triple:       %s\n", config.Triple())
		fmt.Printf("GOOS:              %s\n", config.GOOS())
		fmt.Printf("GOARCH:            %s\n", config.GOARCH())
		fmt.Printf("build tags:        %s\n", strings.Join(config.BuildTags(), " "))
		fmt.Printf("garbage collector: %s\n", config.GC())
		fmt.Printf("scheduler:         %s\n", config.Scheduler())
	case "list":
		config, err := builder.NewConfig(options)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			usage()
			os.Exit(1)
		}
		err = runGoList(config, *flagJSON, *flagDeps, flag.Args())
		if err != nil {
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
		usage()
	case "version":
		goversion := "<unknown>"
		if s, err := goenv.GorootVersionString(goenv.Get("GOROOT")); err == nil {
			goversion = s
		}
		fmt.Printf("tinygo version %s %s/%s (using go version %s and LLVM version %s)\n", version, runtime.GOOS, runtime.GOARCH, goversion, llvm.Version)
	case "env":
		if *flagJSON {
			keys := goenv.Keys
			if flag.NArg() != 0 {
				// Show only one (or a few) environment variables.
				keys = flag.Args()
			}
			// Show environment variables in JSON format.
			env := make(map[string]string)
			for _, key := range keys {
				env[key] = goenv.Get(key)
			}
			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "\t")
			encoder.Encode(env)
		} else {
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
		}
	default:
		fmt.Fprintln(os.Stderr, "Unknown command:", command)
		usage()
		os.Exit(1)
	}
}
