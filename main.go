package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"go/scanner"
	"go/types"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"runtime"
	"runtime/pprof"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/google/shlex"
	"github.com/mattn/go-colorable"
	"github.com/tinygo-org/tinygo/builder"
	"github.com/tinygo-org/tinygo/compileopts"
	"github.com/tinygo-org/tinygo/goenv"
	"github.com/tinygo-org/tinygo/interp"
	"github.com/tinygo-org/tinygo/loader"
	"github.com/tinygo-org/tinygo/transform"
	"tinygo.org/x/go-llvm"

	"go.bug.st/serial"
	"go.bug.st/serial/enumerator"
)

var (
	// This variable is set at build time using -ldflags parameters.
	// See: https://stackoverflow.com/a/11355611
	gitSha1 string
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

	return builder.Build(pkgName, outpath, config, func(result builder.BuildResult) error {
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
		} else {
			// Move was successful.
			return nil
		}
	})
}

// Test runs the tests in the given package. Returns whether the test passed and
// possibly an error if the test failed to run.
func Test(pkgName string, options *compileopts.Options, testCompileOnly, testVerbose, testShort bool, outpath string) (bool, error) {
	options.TestConfig.CompileTestBinary = true
	config, err := builder.NewConfig(options)
	if err != nil {
		return false, err
	}

	passed := true
	err = builder.Build(pkgName, outpath, config, func(result builder.BuildResult) error {
		if testCompileOnly || outpath != "" {
			// Write test binary to the specified file name.
			if outpath == "" {
				// No -o path was given, so create one now.
				// This matches the behavior of go test.
				outpath = filepath.Base(result.MainDir) + ".test"
			}
			copyFile(result.Binary, outpath)
		}
		if testCompileOnly {
			// Do not run the test.
			return nil
		}

		// Run the test.
		start := time.Now()
		var err error
		passed, err = runPackageTest(config, result, testVerbose, testShort)
		if err != nil {
			return err
		}
		duration := time.Since(start)

		// Print the result.
		importPath := strings.TrimSuffix(result.ImportPath, ".test")
		if passed {
			fmt.Printf("ok  \t%s\t%.3fs\n", importPath, duration.Seconds())
		} else {
			fmt.Printf("FAIL\t%s\t%.3fs\n", importPath, duration.Seconds())
		}
		return nil
	})
	if err, ok := err.(loader.NoTestFilesError); ok {
		fmt.Printf("?   \t%s\t[no test files]\n", err.ImportPath)
		// Pretend the test passed - it at least didn't fail.
		return true, nil
	}
	return passed, err
}

// runPackageTest runs a test binary that was previously built. The return
// values are whether the test passed and any errors encountered while trying to
// run the binary.
func runPackageTest(config *compileopts.Config, result builder.BuildResult, testVerbose, testShort bool) (bool, error) {
	var cmd *exec.Cmd
	if len(config.Target.Emulator) == 0 {
		// Run directly.
		var flags []string
		if testVerbose {
			flags = append(flags, "-test.v")
		}
		if testShort {
			flags = append(flags, "-test.short")
		}
		cmd = executeCommand(config.Options, result.Binary, flags...)
		cmd.Dir = result.MainDir
	} else {
		// Run in an emulator.
		// TODO: pass the -test.v flag if needed.
		args := append(config.Target.Emulator[1:], result.Binary)
		cmd = executeCommand(config.Options, config.Target.Emulator[0], args...)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			// Binary exited with a non-zero exit code, which means the test
			// failed.
			return false, nil
		}
		return false, &commandError{"failed to run compiled binary", result.Binary, err}
	}
	return true, nil
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

	return builder.Build(pkgName, fileExt, config, func(result builder.BuildResult) error {
		// do we need port reset to put MCU into bootloader mode?
		if config.Target.PortReset == "true" && flashMethod != "openocd" {
			port, err := getDefaultPort(port, config.Target.SerialPort)
			if err != nil {
				return err
			}

			err = touchSerialPortAt1200bps(port)
			if err != nil {
				return &commandError{"failed to reset port", result.Binary, err}
			}
			// give the target MCU a chance to restart into bootloader
			time.Sleep(3 * time.Second)
		}

		// this flashing method copies the binary data to a Mass Storage Device (msd)
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
			return nil
		case "msd":
			switch fileExt {
			case ".uf2":
				err := flashUF2UsingMSD(config.Target.FlashVolume, result.Binary, config.Options)
				if err != nil {
					return &commandError{"failed to flash", result.Binary, err}
				}
				return nil
			case ".hex":
				err := flashHexUsingMSD(config.Target.FlashVolume, result.Binary, config.Options)
				if err != nil {
					return &commandError{"failed to flash", result.Binary, err}
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
			args = append(args, "-c", "program "+filepath.ToSlash(result.Binary)+" reset exit")
			cmd := executeCommand(config.Options, "openocd", args...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err = cmd.Run()
			if err != nil {
				return &commandError{"failed to flash", result.Binary, err}
			}
			return nil
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
			return nil
		default:
			return fmt.Errorf("unknown flash method: %s", flashMethod)
		}
	})
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

	return builder.Build(pkgName, "", config, func(result builder.BuildResult) error {
		// Find a good way to run GDB.
		gdbInterface, openocdInterface := config.Programmer()
		switch gdbInterface {
		case "msd", "command", "":
			if len(config.Target.Emulator) != 0 {
				if config.Target.Emulator[0] == "mgba" {
					gdbInterface = "mgba"
				} else if config.Target.Emulator[0] == "simavr" {
					gdbInterface = "simavr"
				} else if strings.HasPrefix(config.Target.Emulator[0], "qemu-system-") {
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
			args := append(config.Target.Emulator[1:], result.Binary, "-s", "-S")
			daemon = executeCommand(config.Options, config.Target.Emulator[0], args...)
			daemon.Stdout = os.Stdout
			daemon.Stderr = os.Stderr
		case "qemu-user":
			port = ":1234"

			// Run in an emulator.
			args := append(config.Target.Emulator[1:], "-g", "1234", result.Binary)
			daemon = executeCommand(config.Options, config.Target.Emulator[0], args...)
			daemon.Stdout = os.Stdout
			daemon.Stderr = os.Stderr
		case "mgba":
			port = ":2345"

			// Run in an emulator.
			args := append(config.Target.Emulator[1:], result.Binary, "-g")
			daemon = executeCommand(config.Options, config.Target.Emulator[0], args...)
			daemon.Stdout = os.Stdout
			daemon.Stderr = os.Stderr
		case "simavr":
			port = ":1234"

			// Run in an emulator.
			args := append(config.Target.Emulator[1:], "-g", result.Binary)
			daemon = executeCommand(config.Options, config.Target.Emulator[0], args...)
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
		params := []string{result.Binary}
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
		err := cmd.Run()
		if err != nil {
			return &commandError{"failed to run " + cmdName + " with", result.Binary, err}
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

	return builder.Build(pkgName, ".elf", config, func(result builder.BuildResult) error {
		if len(config.Target.Emulator) == 0 {
			// Run directly.
			cmd := executeCommand(config.Options, result.Binary)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err := cmd.Run()
			if err != nil {
				if err, ok := err.(*exec.ExitError); ok && err.Exited() {
					// Workaround for QEMU which always exits with an error.
					return nil
				}
				return &commandError{"failed to run compiled binary", result.Binary, err}
			}
			return nil
		} else {
			// Run in an emulator.
			args := append(config.Target.Emulator[1:], result.Binary)
			cmd := executeCommand(config.Options, config.Target.Emulator[0], args...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err := cmd.Run()
			if err != nil {
				if err, ok := err.(*exec.ExitError); ok && err.Exited() {
					// Workaround for QEMU which always exits with an error.
					return nil
				}
				return &commandError{"failed to run emulator with", result.Binary, err}
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

const maxMSDRetries = 10

func flashUF2UsingMSD(volume, tmppath string, options *compileopts.Options) error {
	// find standard UF2 info path
	var infoPath string
	switch runtime.GOOS {
	case "linux", "freebsd":
		infoPath = "/media/*/" + volume + "/INFO_UF2.TXT"
	case "darwin":
		infoPath = "/Volumes/" + volume + "/INFO_UF2.TXT"
	case "windows":
		path, err := windowsFindUSBDrive(volume, options)
		if err != nil {
			return err
		}
		infoPath = path + "/INFO_UF2.TXT"
	}

	d, err := locateDevice(volume, infoPath)
	if err != nil {
		return err
	}

	return moveFile(tmppath, filepath.Dir(d)+"/flash.uf2")
}

func flashHexUsingMSD(volume, tmppath string, options *compileopts.Options) error {
	// find expected volume path
	var destPath string
	switch runtime.GOOS {
	case "linux", "freebsd":
		destPath = "/media/*/" + volume
	case "darwin":
		destPath = "/Volumes/" + volume
	case "windows":
		path, err := windowsFindUSBDrive(volume, options)
		if err != nil {
			return err
		}
		destPath = path + "/"
	}

	d, err := locateDevice(volume, destPath)
	if err != nil {
		return err
	}

	return moveFile(tmppath, d+"/flash.hex")
}

func locateDevice(volume, path string) (string, error) {
	var d []string
	var err error
	for i := 0; i < maxMSDRetries; i++ {
		d, err = filepath.Glob(path)
		if err != nil {
			return "", err
		}
		if d != nil {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	if d == nil {
		return "", errors.New("unable to locate device: " + volume)
	}
	return d[0], nil
}

func windowsFindUSBDrive(volume string, options *compileopts.Options) (string, error) {
	cmd := executeCommand(options, "wmic",
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
			if len(parts) != 3 || (parts[0] != "acm" && parts[0] == "usb") {
				// acm and usb are the two types of serial ports recognized
				// under Linux (ttyACM*, ttyUSB*). Other operating systems don't
				// generally make this distinction. If this is not one of the
				// given USB devices, don't try to parse the USB IDs.
				continue
			}
			vid, err := strconv.ParseUint(parts[1], 16, 16)
			if err != nil {
				return "", fmt.Errorf("could not parse USB vendor ID %q: %w", parts[1], err)
			}
			pid, err := strconv.ParseUint(parts[2], 16, 16)
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

		if len(ports) == 0 {
			// fallback
			switch runtime.GOOS {
			case "darwin":
				ports, err = filepath.Glob("/dev/cu.usb*")
			case "linux":
				ports, err = filepath.Glob("/dev/ttyACM*")
			case "windows":
				ports, err = serial.GetPortsList()
			}
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
		if len(ports) == 1 {
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

func usage() {
	fmt.Fprintln(os.Stderr, "TinyGo is a Go compiler for small places.")
	fmt.Fprintln(os.Stderr, "version:", goenv.Version)
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

// try to make the path relative to the current working directory. If any error
// occurs, this error is ignored and the absolute path is returned instead.
func tryToMakePathRelative(dir string) string {
	wd, err := os.Getwd()
	if err != nil {
		return dir
	}
	relpath, err := filepath.Rel(wd, dir)
	if err != nil {
		return dir
	}
	return relpath
}

// printCompilerError prints compiler errors using the provided logger function
// (similar to fmt.Println).
//
// There is one exception: interp errors may print to stderr unconditionally due
// to limitations in the LLVM bindings.
func printCompilerError(logln func(...interface{}), err error) {
	switch err := err.(type) {
	case types.Error:
		printCompilerError(logln, scanner.Error{
			Pos: err.Fset.Position(err.Pos),
			Msg: err.Msg,
		})
	case scanner.Error:
		if !strings.HasPrefix(err.Pos.Filename, filepath.Join(goenv.Get("GOROOT"), "src")) && !strings.HasPrefix(err.Pos.Filename, filepath.Join(goenv.Get("TINYGOROOT"), "src")) {
			// This file is not from the standard library (either the GOROOT or
			// the TINYGOROOT). Make the path relative, for easier reading.
			// Ignore any errors in the process (falling back to the absolute
			// path).
			err.Pos.Filename = tryToMakePathRelative(err.Pos.Filename)
		}
		logln(err)
	case scanner.ErrorList:
		for _, scannerErr := range err {
			printCompilerError(logln, *scannerErr)
		}
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
	case transform.CoroutinesError:
		logln(err.Pos.String() + ": " + err.Msg)
		logln("\ntraceback:")
		for _, line := range err.Traceback {
			logln(line.Name)
			if line.Position.IsValid() {
				logln("\t" + line.Position.String())
			}
		}
	case loader.Errors:
		logln("#", err.Pkg.ImportPath)
		for _, err := range err.Errs {
			printCompilerError(logln, err)
		}
	case loader.Error:
		logln(err.Err.Error())
		logln("package", err.ImportStack[0])
		for _, pkgPath := range err.ImportStack[1:] {
			logln("\timports", pkgPath)
		}
	case *builder.MultiError:
		for _, err := range err.Errs {
			printCompilerError(logln, err)
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
func parseGoLinkFlag(flagsString string) (map[string]map[string]string, error) {
	set := flag.NewFlagSet("link", flag.ExitOnError)
	globalVarValues := make(globalValuesFlag)
	set.Var(globalVarValues, "X", "Set the value of the string variable to the given value.")
	flags, err := shlex.Split(flagsString)
	if err != nil {
		return nil, err
	}
	err = set.Parse(flags)
	if err != nil {
		return nil, err
	}
	return map[string]map[string]string(globalVarValues), nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "No command-line arguments supplied.")
		usage()
		os.Exit(1)
	}
	command := os.Args[1]

	opt := flag.String("opt", "z", "optimization level: 0, 1, 2, s, z")
	interpPass := flag.String("interp-pass", "all", "interp pass: none, all, package, program")
	gc := flag.String("gc", "", "garbage collector to use (none, leaking, extalloc, conservative)")
	panicStrategy := flag.String("panic", "print", "panic strategy (print, trap)")
	scheduler := flag.String("scheduler", "", "which scheduler to use (none, coroutines, tasks)")
	serial := flag.String("serial", "", "which serial output to use (none, uart, usb)")
	printIR := flag.Bool("printir", false, "print LLVM IR")
	dumpSSA := flag.Bool("dumpssa", false, "dump internal Go SSA")
	verifyIR := flag.Bool("verifyir", false, "run extra verification steps on LLVM IR")
	tags := flag.String("tags", "", "a space-separated list of extra build tags")
	target := flag.String("target", "", "chip/board name or JSON target specification file")
	printSize := flag.String("size", "", "print sizes (none, short, full)")
	printStacks := flag.Bool("print-stacks", false, "print stack sizes of goroutines")
	printAllocsString := flag.String("print-allocs", "", "regular expression of functions for which heap allocations should be printed")
	printCommands := flag.Bool("x", false, "Print commands")
	parallelism := flag.Int("p", runtime.GOMAXPROCS(0), "the number of build jobs that can run in parallel")
	nodebug := flag.Bool("no-debug", false, "strip debug information")
	ocdCommandsString := flag.String("ocd-commands", "", "OpenOCD commands, overriding target spec (can specify multiple separated by commas)")
	ocdOutput := flag.Bool("ocd-output", false, "print OCD daemon output during debug")
	port := flag.String("port", "", "flash port (can specify multiple candidates separated by commas)")
	programmer := flag.String("programmer", "", "which hardware programmer to use")
	ldflags := flag.String("ldflags", "", "Go link tool compatible ldflags")
	wasmAbi := flag.String("wasm-abi", "", "WebAssembly ABI conventions: js (no i64 params) or generic")
	llvmFeatures := flag.String("llvm-features", "", "comma separated LLVM features to enable")
	cpuprofile := flag.String("cpuprofile", "", "cpuprofile output")

	var flagJSON, flagDeps, flagTest *bool
	if command == "help" || command == "list" {
		flagJSON = flag.Bool("json", false, "print data in JSON format")
		flagDeps = flag.Bool("deps", false, "supply -deps flag to go list")
		flagTest = flag.Bool("test", false, "supply -test flag to go list")
	}
	var outpath string
	if command == "help" || command == "build" || command == "build-library" || command == "test" {
		flag.StringVar(&outpath, "o", "", "output filename")
	}
	var testCompileOnlyFlag, testVerboseFlag, testShortFlag *bool
	if command == "help" || command == "test" {
		testCompileOnlyFlag = flag.Bool("c", false, "compile the test binary but do not run it")
		testVerboseFlag = flag.Bool("v", false, "verbose: print additional output")
		testShortFlag = flag.Bool("short", false, "short: run smaller test suite to save time")
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
	globalVarValues, err := parseGoLinkFlag(*ldflags)
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
		Target:          *target,
		Opt:             *opt,
		InterpPass:      *interpPass,
		GC:              *gc,
		PanicStrategy:   *panicStrategy,
		Scheduler:       *scheduler,
		Serial:          *serial,
		PrintIR:         *printIR,
		DumpSSA:         *dumpSSA,
		VerifyIR:        *verifyIR,
		Parallelism:     *parallelism,
		Debug:           !*nodebug,
		PrintSizes:      *printSize,
		PrintStacks:     *printStacks,
		PrintAllocs:     printAllocs,
		Tags:            *tags,
		GlobalValues:    globalVarValues,
		WasmAbi:         *wasmAbi,
		Programmer:      *programmer,
		OpenOCDCommands: ocdCommands,
		LLVMFeatures:    *llvmFeatures,
	}
	if *printCommands {
		options.PrintCommands = printCommand
	}

	err = options.Verify()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		usage()
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
		if outpath == "" {
			fmt.Fprintln(os.Stderr, "No output filename supplied (-o).")
			usage()
			os.Exit(1)
		}
		pkgName := "."
		if flag.NArg() == 1 {
			pkgName = filepath.ToSlash(flag.Arg(0))
		} else if flag.NArg() > 1 {
			fmt.Fprintln(os.Stderr, "build only accepts a single positional argument: package name, but multiple were specified")
			usage()
			os.Exit(1)
		}
		if options.Target == "" && filepath.Ext(outpath) == ".wasm" {
			options.Target = "wasm"
		}

		err := Build(pkgName, outpath, options)
		handleCompilerError(err)
	case "build-library":
		// Note: this command is only meant to be used while making a release!
		if outpath == "" {
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
		tmpdir, err := ioutil.TempDir("", "tinygo*")
		if err != nil {
			handleCompilerError(err)
		}
		defer os.RemoveAll(tmpdir)
		config := &compileopts.Config{
			Options: options,
			Target: &compileopts.TargetSpec{
				Triple: *target,
			},
		}
		path, err := lib.Load(config, tmpdir)
		handleCompilerError(err)
		err = copyFile(path, outpath)
		if err != nil {
			handleCompilerError(err)
		}
	case "flash", "gdb", "lldb":
		pkgName := filepath.ToSlash(flag.Arg(0))
		if command == "flash" {
			err := Flash(pkgName, *port, options)
			handleCompilerError(err)
		} else {
			if !options.Debug {
				fmt.Fprintln(os.Stderr, "Debug disabled while running debugger?")
				usage()
				os.Exit(1)
			}
			err := Debug(command, pkgName, *ocdOutput, options)
			handleCompilerError(err)
		}
	case "run":
		if flag.NArg() != 1 {
			fmt.Fprintln(os.Stderr, "No package specified.")
			usage()
			os.Exit(1)
		}
		pkgName := filepath.ToSlash(flag.Arg(0))
		err := Run(pkgName, options)
		handleCompilerError(err)
	case "test":
		var pkgNames []string
		for i := 0; i < flag.NArg(); i++ {
			pkgNames = append(pkgNames, filepath.ToSlash(flag.Arg(i)))
		}
		if len(pkgNames) == 0 {
			pkgNames = []string{"."}
		}
		allTestsPassed := true
		for _, pkgName := range pkgNames {
			// TODO: parallelize building the test binaries
			passed, err := Test(pkgName, options, *testCompileOnlyFlag, *testVerboseFlag, *testShortFlag, outpath)
			handleCompilerError(err)
			if !passed {
				allTestsPassed = false
			}
		}
		if !allTestsPassed {
			fmt.Println("FAIL")
			os.Exit(1)
		}
	case "targets":
		dir := filepath.Join(goenv.Get("TINYGOROOT"), "targets")
		entries, err := ioutil.ReadDir(dir)
		if err != nil {
			fmt.Fprintln(os.Stderr, "could not list targets:", err)
			os.Exit(1)
			return
		}
		for _, entry := range entries {
			if !entry.Mode().IsRegular() || !strings.HasSuffix(entry.Name(), ".json") {
				// Only inspect JSON files.
				continue
			}
			path := filepath.Join(dir, entry.Name())
			spec, err := compileopts.LoadTarget(&compileopts.Options{Target: path})
			if err != nil {
				fmt.Fprintln(os.Stderr, "could not list target:", err)
				os.Exit(1)
				return
			}
			if spec.FlashMethod == "" && spec.FlashCommand == "" && spec.Emulator == nil {
				// This doesn't look like a regular target file, but rather like
				// a parent target (such as targets/cortex-m.json).
				continue
			}
			name := entry.Name()
			name = name[:len(name)-5]
			fmt.Println(name)
		}
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
		cachedGOROOT, err := loader.GetCachedGoroot(config)
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
		fmt.Printf("cached GOROOT:     %s\n", cachedGOROOT)
	case "list":
		config, err := builder.NewConfig(options)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			usage()
			os.Exit(1)
		}
		var extraArgs []string
		if *flagJSON {
			extraArgs = append(extraArgs, "-json")
		}
		if *flagDeps {
			extraArgs = append(extraArgs, "-deps")
		}
		if *flagTest {
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
		usage()
	case "version":
		goversion := "<unknown>"
		if s, err := goenv.GorootVersionString(goenv.Get("GOROOT")); err == nil {
			goversion = s
		}
		version := goenv.Version
		if strings.HasSuffix(goenv.Version, "-dev") && gitSha1 != "" {
			version += "-" + gitSha1
		}
		fmt.Printf("tinygo version %s %s/%s (using go version %s and LLVM version %s)\n", version, runtime.GOOS, runtime.GOARCH, goversion, llvm.Version)
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
