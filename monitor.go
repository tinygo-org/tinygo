package main

import (
	"bufio"
	"debug/dwarf"
	"debug/elf"
	"debug/macho"
	"debug/pe"
	"errors"
	"fmt"
	"go/token"
	"io"
	"net"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/mattn/go-tty"
	"github.com/tinygo-org/tinygo/compileopts"

	"go.bug.st/serial"
	"go.bug.st/serial/enumerator"
)

// Monitor connects to the given port and reads/writes the serial port.
func Monitor(executable, port string, config *compileopts.Config) error {
	const timeout = time.Second * 3
	var exit func() // function to be called before exiting
	var serialConn io.ReadWriter

	if config.Options.Serial == "rtt" {
		// Use the RTT interface, which is documented (in part) here:
		// https://wiki.segger.com/RTT

		// Try to find the "machine.rttSerialInstance" symbol, which is the RTT
		// control block.
		file, err := elf.Open(executable)
		if err != nil {
			return fmt.Errorf("could not open ELF file to determine RTT control block: %w", err)
		}
		defer file.Close()
		symbols, err := file.Symbols()
		if err != nil {
			return fmt.Errorf("could not read ELF symbol table to determine RTT control block: %w", err)
		}
		var address uint64
		for _, symbol := range symbols {
			if symbol.Name == "machine.rttSerialInstance" {
				address = symbol.Value
				break
			}
		}
		if address == 0 {
			return fmt.Errorf("could not find RTT control block in ELF file")
		}

		// Start an openocd process in the background.
		args, err := config.OpenOCDConfiguration()
		if err != nil {
			return err
		}
		args = append(args,
			"-c", fmt.Sprintf("rtt setup 0x%x 16 \"SEGGER RTT\"", address),
			"-c", "init",
			"-c", "rtt server start 0 0")
		cmd := executeCommand(config.Options, "openocd", args...)
		stderr, err := cmd.StderrPipe()
		if err != nil {
			return err
		}
		cmd.Stdout = os.Stdout
		err = cmd.Start()
		if err != nil {
			return err
		}
		defer cmd.Process.Kill()
		exit = func() {
			// Make sure the openocd process is terminated at exit.
			// This does not happen through the defer above when exiting through
			// os.Exit.
			cmd.Process.Kill()
		}

		// Read the stderr, which logs various important messages we need.
		r := bufio.NewReader(stderr)
		var telnet net.Conn
		var timeoutAt time.Time
		for {
			// Read the next line from the openocd process.
			lineBytes, err := r.ReadBytes('\n')
			if err != nil {
				return err
			}
			line := string(lineBytes)

			if line == "Info : rtt: No control block found\n" {
				// Message that is sent back when OpenOCD can't find the control
				// block after a 'rtt start' message.
				if time.Now().After(timeoutAt) {
					return fmt.Errorf("RTT timeout (could not locate RTT control block at 0x%08x)", address)
				}
				time.Sleep(time.Millisecond * 100)
				telnet.Write([]byte("rtt start\r\n"))
			} else if strings.HasPrefix(line, "Info : Listening on port") {
				// We need two different ports for controlling OpenOCD
				// (typically port 4444) and the RTT channel 0 socket (arbitrary
				// port).
				var port int
				var protocol string
				fmt.Sscanf(line, "Info : Listening on port %d for %s connections\n", &port, &protocol)
				if protocol == "telnet" && telnet == nil {
					// Connect to the "telnet" command line interface.
					telnet, err = net.Dial("tcp4", fmt.Sprintf("localhost:%d", port))
					if err != nil {
						return err
					}
					// Tell OpenOCD to start scanning for the RTT control block.
					telnet.Write([]byte("rtt start\r\n"))
					// Also make sure we will time out if the control block just
					// can't be found.
					timeoutAt = time.Now().Add(timeout)
				} else if protocol == "rtt" {
					// Connect to the RTT channel, for both stdin and stdout.
					conn, err := net.Dial("tcp4", fmt.Sprintf("localhost:%d", port))
					if err != nil {
						return err
					}
					serialConn = conn
				}
			} else if strings.HasPrefix(line, "Info : rtt: Control block found at") {
				// Connection established!
				break
			}
		}
	} else { // -serial=uart or -serial=usb
		var err error
		wait := 300
		for i := 0; i <= wait; i++ {
			port, err = getDefaultPort(port, config.Target.SerialPort)
			if err != nil {
				if i < wait {
					time.Sleep(10 * time.Millisecond)
					continue
				}
				return err
			}
			break
		}

		br := config.Options.BaudRate
		if br <= 0 {
			br = 115200
		}

		wait = 300
		var p serial.Port
		for i := 0; i <= wait; i++ {
			p, err = serial.Open(port, &serial.Mode{BaudRate: br})
			if err != nil {
				if i < wait {
					time.Sleep(10 * time.Millisecond)
					continue
				}
				return err
			}
			serialConn = p
			break
		}
		defer p.Close()
	}

	tty, err := tty.Open()
	if err != nil {
		return err
	}
	defer tty.Close()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	defer signal.Stop(sig)

	go func() {
		<-sig
		tty.Close()
		if exit != nil {
			exit()
		}
		os.Exit(0)
	}()

	fmt.Printf("Connected to %s. Press Ctrl-C to exit.\n", port)

	errCh := make(chan error, 1)

	go func() {
		buf := make([]byte, 100*1024)
		writer := newOutputWriter(os.Stdout, executable)
		for {
			n, err := serialConn.Read(buf)
			if err != nil {
				errCh <- fmt.Errorf("read error: %w", err)
				return
			}
			writer.Write(buf[:n])
		}
	}()

	go func() {
		for {
			r, err := tty.ReadRune()
			if err != nil {
				errCh <- err
				return
			}
			if r == 0 {
				continue
			}
			serialConn.Write([]byte(string(r)))
		}
	}()

	return <-errCh
}

// SerialPortInfo is a structure that holds information about the port and its
// associated TargetSpec.
type SerialPortInfo struct {
	Name   string
	IsUSB  bool
	VID    string
	PID    string
	Target string
	Spec   *compileopts.TargetSpec
}

// ListSerialPort returns serial port information and any detected TinyGo
// target.
func ListSerialPorts() ([]SerialPortInfo, error) {
	maps, err := compileopts.GetTargetSpecs()
	if err != nil {
		return nil, err
	}

	portsList, err := enumerator.GetDetailedPortsList()
	if err != nil {
		return nil, err
	}

	serialPortInfo := []SerialPortInfo{}
	for _, p := range portsList {
		info := SerialPortInfo{
			Name:  p.Name,
			IsUSB: p.IsUSB,
			VID:   p.VID,
			PID:   p.PID,
		}
		vid := strings.ToLower(p.VID)
		pid := strings.ToLower(p.PID)
		for k, v := range maps {
			usbInterfaces := v.SerialPort
			for _, s := range usbInterfaces {
				parts := strings.Split(s, ":")
				if len(parts) != 2 {
					continue
				}
				if vid == strings.ToLower(parts[0]) && pid == strings.ToLower(parts[1]) {
					info.Target = k
					info.Spec = v
				}
			}
		}
		serialPortInfo = append(serialPortInfo, info)
	}

	return serialPortInfo, nil
}

var addressMatch = regexp.MustCompile(`^panic: runtime error at 0x([0-9a-f]+): `)

// Extract the address from the "panic: runtime error at" message.
func extractPanicAddress(line []byte) uint64 {
	matches := addressMatch.FindSubmatch(line)
	if matches != nil {
		address, err := strconv.ParseUint(string(matches[1]), 16, 64)
		if err == nil {
			return address
		}
	}
	return 0
}

// Convert an address in the binary to a source address location.
func addressToLine(executable string, address uint64) (token.Position, error) {
	data, err := readDWARF(executable)
	if err != nil {
		return token.Position{}, err
	}
	r := data.Reader()

	for {
		e, err := r.Next()
		if err != nil {
			return token.Position{}, err
		}
		if e == nil {
			break
		}
		switch e.Tag {
		case dwarf.TagCompileUnit:
			r.SkipChildren()
			lr, err := data.LineReader(e)
			if err != nil {
				return token.Position{}, err
			}
			var lineEntry = dwarf.LineEntry{
				EndSequence: true,
			}
			for {
				// Read the next .debug_line entry.
				prevLineEntry := lineEntry
				err := lr.Next(&lineEntry)
				if err != nil {
					if err == io.EOF {
						break
					}
					return token.Position{}, err
				}

				if prevLineEntry.EndSequence && lineEntry.Address == 0 {
					// Tombstone value. This symbol has been removed, for
					// example by the --gc-sections linker flag. It is still
					// here in the debug information because the linker can't
					// just remove this reference.
					// Read until the next EndSequence so that this sequence is
					// skipped.
					// For more details, see (among others):
					// https://reviews.llvm.org/D84825
					for {
						err := lr.Next(&lineEntry)
						if err != nil {
							return token.Position{}, err
						}
						if lineEntry.EndSequence {
							break
						}
					}
				}

				if !prevLineEntry.EndSequence {
					// The chunk describes the code from prevLineEntry to
					// lineEntry.
					if prevLineEntry.Address <= address && lineEntry.Address > address {
						return token.Position{
							Filename: prevLineEntry.File.Name,
							Line:     prevLineEntry.Line,
							Column:   prevLineEntry.Column,
						}, nil
					}
				}
			}
		}
	}

	return token.Position{}, nil // location not found
}

// Read the DWARF debug information from a given file (in various formats).
func readDWARF(executable string) (*dwarf.Data, error) {
	f, err := os.Open(executable)
	if err != nil {
		return nil, err
	}
	if file, err := elf.NewFile(f); err == nil {
		return file.DWARF()
	} else if file, err := macho.NewFile(f); err == nil {
		return file.DWARF()
	} else if file, err := pe.NewFile(f); err == nil {
		return file.DWARF()
	} else {
		return nil, errors.New("unknown binary format")
	}
}

type outputWriter struct {
	out        io.Writer
	executable string
	line       []byte
}

// newOutputWriter returns an io.Writer that will intercept panic addresses and
// will try to insert a source location in the output if the source location can
// be found in the executable.
func newOutputWriter(out io.Writer, executable string) *outputWriter {
	return &outputWriter{
		out:        out,
		executable: executable,
	}
}

func (w *outputWriter) Write(p []byte) (n int, err error) {
	start := 0
	for i, c := range p {
		if c == '\n' {
			w.out.Write(p[start : i+1])
			start = i + 1
			address := extractPanicAddress(w.line)
			if address != 0 {
				loc, err := addressToLine(w.executable, address)
				if err == nil && loc.Filename != "" {
					fmt.Printf("[tinygo: panic at %s]\n", loc.String())
				}
			}
			w.line = w.line[:0]
		} else {
			w.line = append(w.line, c)
		}
	}
	w.out.Write(p[start:])
	n = len(p)
	return
}
