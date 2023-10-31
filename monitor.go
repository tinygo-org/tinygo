package main

import (
	"debug/dwarf"
	"debug/elf"
	"debug/macho"
	"debug/pe"
	"errors"
	"fmt"
	"go/token"
	"io"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/mattn/go-tty"
	"github.com/tinygo-org/tinygo/builder"
	"github.com/tinygo-org/tinygo/compileopts"

	"go.bug.st/serial"
	"go.bug.st/serial/enumerator"
)

// Monitor connects to the given port and reads/writes the serial port.
func Monitor(executable, port string, options *compileopts.Options) error {
	config, err := builder.NewConfig(options)
	if err != nil {
		return err
	}

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

	br := options.BaudRate
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
		break
	}
	defer p.Close()

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
		os.Exit(0)
	}()

	fmt.Printf("Connected to %s. Press Ctrl-C to exit.\n", port)

	errCh := make(chan error, 1)

	go func() {
		buf := make([]byte, 100*1024)
		var line []byte
		for {
			n, err := p.Read(buf)
			if err != nil {
				errCh <- fmt.Errorf("read error: %w", err)
				return
			}
			start := 0
			for i, c := range buf[:n] {
				if c == '\n' {
					os.Stdout.Write(buf[start : i+1])
					start = i + 1
					address := extractPanicAddress(line)
					if address != 0 {
						loc, err := addressToLine(executable, address)
						if err == nil && loc.IsValid() {
							fmt.Printf("[tinygo: panic at %s]\n", loc.String())
						}
					}
					line = line[:0]
				} else {
					line = append(line, c)
				}
			}
			os.Stdout.Write(buf[start:n])
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
			p.Write([]byte(string(r)))
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
