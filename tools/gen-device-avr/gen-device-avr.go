package main

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"html/template"
	"math/bits"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type AVRToolsDeviceFile struct {
	XMLName xml.Name `xml:"avr-tools-device-file"`
	Devices []struct {
		Name          string `xml:"name,attr"`
		Architecture  string `xml:"architecture,attr"`
		Family        string `xml:"family,attr"`
		AddressSpaces []struct {
			Name           string `xml:"name,attr"`
			Size           string `xml:"size,attr"`
			MemorySegments []struct {
				Name  string `xml:"name,attr"`
				Start string `xml:"start,attr"`
				Size  string `xml:"size,attr"`
			} `xml:"memory-segment"`
		} `xml:"address-spaces>address-space"`
		Interrupts []*XMLInterrupt `xml:"interrupts>interrupt"`
	} `xml:"devices>device"`
	Modules []struct {
		Name          string `xml:"name,attr"`
		Caption       string `xml:"caption,attr"`
		RegisterGroup struct {
			Name      string `xml:"name,attr"`
			Caption   string `xml:"caption,attr"`
			Registers []struct {
				Name      string `xml:"name,attr"`
				Caption   string `xml:"caption,attr"`
				Offset    string `xml:"offset,attr"`
				Size      int    `xml:"size,attr"`
				Bitfields []struct {
					Name    string `xml:"name,attr"`
					Caption string `xml:"caption,attr"`
					Mask    string `xml:"mask,attr"`
				} `xml:"bitfield"`
			} `xml:"register"`
		} `xml:"register-group"`
	} `xml:"modules>module"`
}

type XMLInterrupt struct {
	Index    int    `xml:"index,attr"`
	Name     string `xml:"name,attr"`
	Instance string `xml:"module-instance,attr"`
	Caption  string `xml:"caption,attr"`
}

type Device struct {
	metadata    map[string]interface{}
	interrupts  []Interrupt
	peripherals []*Peripheral
}

// AddressSpace is the Go version of an XML element like the following:
//
//	<address-space endianness="little" name="data" id="data" start="0x0000" size="0x0900">
//
// It describes one address space in an AVR microcontroller. One address space
// may have multiple memory segments.
type AddressSpace struct {
	Size     string
	Segments map[string]MemorySegment
}

// MemorySegment is the Go version of an XML element like the following:
//
//	<memory-segment name="IRAM" start="0x0100" size="0x0800" type="ram" external="false"/>
//
// It describes a single contiguous area of memory in a particular address space
// (see AddressSpace).
type MemorySegment struct {
	start int64
	size  int64
}

type Interrupt struct {
	Index   int
	Name    string
	Caption string
}

type Peripheral struct {
	Name      string
	Caption   string
	Registers []*Register
}

type Register struct {
	Caption    string
	Variants   []RegisterVariant
	Bitfields  []Bitfield
	peripheral *Peripheral
}

type RegisterVariant struct {
	Name    string
	Address int64
}

type Bitfield struct {
	Name    string
	Caption string
	Mask    uint
}

func readATDF(path string) (*Device, error) {
	// Read Atmel device descriptor files.
	// See: http://packs.download.atmel.com

	// Open the XML file.
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	decoder := xml.NewDecoder(f)
	xml := &AVRToolsDeviceFile{}
	err = decoder.Decode(xml)
	if err != nil {
		return nil, err
	}

	device := xml.Devices[0]

	memorySizes := make(map[string]*AddressSpace, len(device.AddressSpaces))
	for _, el := range device.AddressSpaces {
		memorySizes[el.Name] = &AddressSpace{
			Size:     el.Size,
			Segments: make(map[string]MemorySegment),
		}
		for _, segmentEl := range el.MemorySegments {
			start, err := strconv.ParseInt(segmentEl.Start, 0, 32)
			if err != nil {
				return nil, err
			}
			size, err := strconv.ParseInt(segmentEl.Size, 0, 32)
			if err != nil {
				return nil, err
			}
			memorySizes[el.Name].Segments[segmentEl.Name] = MemorySegment{
				start: start,
				size:  size,
			}
		}
	}

	allRegisters := map[string]*Register{}

	var peripherals []*Peripheral
	for _, el := range xml.Modules {
		peripheral := &Peripheral{
			Name:    el.Name,
			Caption: el.Caption,
		}
		peripherals = append(peripherals, peripheral)

		regElGroup := el.RegisterGroup
		for _, regEl := range regElGroup.Registers {
			regOffset, err := strconv.ParseInt(regEl.Offset, 0, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse offset %#v of register %s: %v", regEl.Offset, regEl.Name, err)
			}
			reg := &Register{
				Caption:    regEl.Caption,
				peripheral: peripheral,
			}
			switch regEl.Size {
			case 1:
				reg.Variants = []RegisterVariant{
					{
						Name:    regEl.Name,
						Address: regOffset,
					},
				}
			case 2:
				reg.Variants = []RegisterVariant{
					{
						Name:    regEl.Name + "L",
						Address: regOffset,
					},
					{
						Name:    regEl.Name + "H",
						Address: regOffset + 1,
					},
				}
			default:
				// TODO
				continue
			}

			for _, bitfieldEl := range regEl.Bitfields {
				mask := bitfieldEl.Mask
				if len(mask) == 2 {
					// Two devices (ATtiny102 and ATtiny104) appear to have an
					// error in the bitfields, leaving out the '0x' prefix.
					mask = "0x" + mask
				}
				maskInt, err := strconv.ParseUint(mask, 0, 32)
				if err != nil {
					return nil, fmt.Errorf("failed to parse mask %#v of bitfield %s: %v", mask, bitfieldEl.Name, err)
				}
				reg.Bitfields = append(reg.Bitfields, Bitfield{
					Name:    regEl.Name + "_" + bitfieldEl.Name,
					Caption: bitfieldEl.Caption,
					Mask:    uint(maskInt),
				})
			}

			if firstReg, ok := allRegisters[regEl.Name]; ok {
				// merge bit fields with previous register
				merged := append(firstReg.Bitfields, reg.Bitfields...)
				firstReg.Bitfields = make([]Bitfield, 0, len(merged))
				m := make(map[string]interface{})
				for _, field := range merged {
					if _, ok := m[field.Name]; !ok {
						m[field.Name] = nil
						firstReg.Bitfields = append(firstReg.Bitfields, field)
					}
				}
				continue
			} else {
				allRegisters[regEl.Name] = reg
			}

			peripheral.Registers = append(peripheral.Registers, reg)
		}
	}

	ramStart := int64(0)
	ramSize := int64(0) // for devices with no RAM
	for _, ramSegmentName := range []string{"IRAM", "INTERNAL_SRAM", "SRAM"} {
		if segment, ok := memorySizes["data"].Segments[ramSegmentName]; ok {
			ramStart = segment.start
			ramSize = segment.size
		}
	}

	flashSize, err := strconv.ParseInt(memorySizes["prog"].Size, 0, 32)
	if err != nil {
		return nil, err
	}

	// Process the interrupts to clean up inconsistencies between ATDF files.
	var interrupts []Interrupt
	hasResetInterrupt := false
	for _, intr := range device.Interrupts {
		name := intr.Name
		if intr.Instance != "" {
			// ATDF files for newer chips also have an instance name, which must
			// be specified to make the interrupt name unique.
			name = intr.Instance + "_" + name
		}
		if name == "RESET" {
			hasResetInterrupt = true
		}
		interrupts = append(interrupts, Interrupt{
			Index:   intr.Index,
			Name:    name,
			Caption: intr.Caption,
		})
	}
	if !hasResetInterrupt {
		interrupts = append(interrupts, Interrupt{
			Index: 0,
			Name:  "RESET",
		})
	}
	sort.SliceStable(interrupts, func(i, j int) bool {
		return interrupts[i].Index < interrupts[j].Index
	})

	return &Device{
		metadata: map[string]interface{}{
			"file":             filepath.Base(path),
			"descriptorSource": "http://packs.download.atmel.com/",
			"name":             device.Name,
			"nameLower":        strings.ToLower(device.Name),
			"description":      fmt.Sprintf("Device information for the %s.", device.Name),
			"arch":             device.Architecture,
			"family":           device.Family,
			"flashSize":        int(flashSize),
			"ramStart":         ramStart,
			"ramSize":          ramSize,
			"numInterrupts":    len(device.Interrupts),
		},
		interrupts:  interrupts,
		peripherals: peripherals,
	}, nil
}

func writeGo(outdir string, device *Device) error {
	// The Go module for this device.
	outf, err := os.Create(outdir + "/" + device.metadata["nameLower"].(string) + ".go")
	if err != nil {
		return err
	}
	defer outf.Close()
	w := bufio.NewWriter(outf)

	maxInterruptNum := 0
	for _, intr := range device.interrupts {
		if intr.Index > maxInterruptNum {
			maxInterruptNum = intr.Index
		}
	}

	t := template.Must(template.New("go").Parse(`// Automatically generated file. DO NOT EDIT.
// Generated by gen-device-avr.go from {{.metadata.file}}, see {{.metadata.descriptorSource}}

//go:build {{.pkgName}} && {{.metadata.nameLower}}

// {{.metadata.description}}
package {{.pkgName}}

import (
	"runtime/volatile"
	"unsafe"
)

// Some information about this device.
const (
	DEVICE = "{{.metadata.name}}"
	ARCH   = "{{.metadata.arch}}"
	FAMILY = "{{.metadata.family}}"
)

// Interrupts
const ({{range .interrupts}}
	IRQ_{{.Name}} = {{.Index}} // {{.Caption}}{{end}}
	IRQ_max = {{.interruptMax}} // Highest interrupt number on this device.
)

// Pseudo function call that is replaced by the compiler with the actual
// functions registered through interrupt.New.
//go:linkname callHandlers runtime/interrupt.callHandlers
func callHandlers(num int)

{{- range .interrupts}}
//export __vector_{{.Name}}
//go:interrupt
func interrupt{{.Name}}() {
	callHandlers(IRQ_{{.Name}})
}
{{- end}}

// Peripherals.
var ({{range .peripherals}}
	// {{.Caption}}
{{range .Registers}}{{range .Variants}}	{{.Name}} = (*volatile.Register8)(unsafe.Pointer(uintptr(0x{{printf "%x" .Address}})))
{{end}}{{end}}{{end}})
`))
	err = t.Execute(w, map[string]interface{}{
		"metadata":     device.metadata,
		"pkgName":      filepath.Base(strings.TrimRight(outdir, "/")),
		"interrupts":   device.interrupts,
		"interruptMax": maxInterruptNum,
		"peripherals":  device.peripherals,
	})
	if err != nil {
		return err
	}

	// Write bitfields.
	for _, peripheral := range device.peripherals {
		// Only write bitfields when there are any.
		numFields := 0
		for _, r := range peripheral.Registers {
			numFields += len(r.Bitfields)
		}
		if numFields == 0 {
			continue
		}

		fmt.Fprintf(w, "\n// Bitfields for %s: %s\nconst(", peripheral.Name, peripheral.Caption)
		for _, register := range peripheral.Registers {
			if len(register.Bitfields) == 0 {
				continue
			}
			for _, variant := range register.Variants {
				fmt.Fprintf(w, "\n\t// %s", variant.Name)
				if register.Caption != "" {
					fmt.Fprintf(w, ": %s", register.Caption)
				}
				fmt.Fprintf(w, "\n")
			}
			allBits := map[string]interface{}{}
			for _, bitfield := range register.Bitfields {
				if bits.OnesCount(bitfield.Mask) == 1 {
					fmt.Fprintf(w, "\t%s = 0x%x", bitfield.Name, bitfield.Mask)
					if len(bitfield.Caption) != 0 {
						fmt.Fprintf(w, " // %s", bitfield.Caption)
					}
					fmt.Fprintf(w, "\n")
					fmt.Fprintf(w, "\t%s_Msk = 0x%x", bitfield.Name, bitfield.Mask)
					if len(bitfield.Caption) != 0 {
						fmt.Fprintf(w, " // %s", bitfield.Caption)
					}
					fmt.Fprintf(w, "\n")
					allBits[bitfield.Name] = nil
				} else {
					n := 0
					for i := uint(0); i < 8; i++ {
						if (bitfield.Mask>>i)&1 == 0 {
							continue
						}
						name := fmt.Sprintf("%s%d", bitfield.Name, n)
						if _, ok := allBits[name]; !ok {
							fmt.Fprintf(w, "\t%s = 0x%x", name, 1<<i)
							if len(bitfield.Caption) != 0 {
								fmt.Fprintf(w, " // %s", bitfield.Caption)
							}
							fmt.Fprintf(w, "\n")
							allBits[name] = nil
						}
						n++
					}
					fmt.Fprintf(w, "\t%s_Msk = 0x%x", bitfield.Name, bitfield.Mask)
					if len(bitfield.Caption) != 0 {
						fmt.Fprintf(w, " // %s", bitfield.Caption)
					}
					fmt.Fprintf(w, "\n")
				}
			}
		}
		fmt.Fprintf(w, ")\n")
	}
	return w.Flush()
}

func writeAsm(outdir string, device *Device) error {
	// The interrupt vector, which is hard to write directly in Go.
	out, err := os.Create(outdir + "/" + device.metadata["nameLower"].(string) + ".s")
	if err != nil {
		return err
	}
	defer out.Close()
	t := template.Must(template.New("asm").Parse(
		`; Automatically generated file. DO NOT EDIT.
; Generated by gen-device-avr.go from {{.file}}, see {{.descriptorSource}}

; This is the default handler for interrupts, if triggered but not defined.
; Sleep inside so that an accidentally triggered interrupt won't drain the
; battery of a battery-powered device.
.section .text.__vector_default
.global  __vector_default
__vector_default:
    sleep
    rjmp __vector_default

; Avoid the need for repeated .weak and .set instructions.
.macro IRQ handler
    .weak  \handler
    .set   \handler, __vector_default
.endm

; The interrupt vector of this device. Must be placed at address 0 by the linker.
.section .vectors, "a", %progbits
.global  __vectors
`))
	err = t.Execute(out, device.metadata)
	if err != nil {
		return err
	}
	num := 0
	for _, intr := range device.interrupts {
		jmp := "jmp"
		if device.metadata["flashSize"].(int) <= 8*1024 {
			// When a device has 8kB or less flash, rjmp (2 bytes) must be used
			// instead of jmp (4 bytes).
			// https://www.avrfreaks.net/forum/rjmp-versus-jmp
			jmp = "rjmp"
		}
		if intr.Index < num {
			// Some devices have duplicate interrupts, probably for historical
			// reasons.
			continue
		}
		for intr.Index > num {
			fmt.Fprintf(out, "    %s __vector_default\n", jmp)
			num++
		}
		num++
		fmt.Fprintf(out, "    %s __vector_%s\n", jmp, intr.Name)
	}

	fmt.Fprint(out, `
    ; Define default implementations for interrupts, redirecting to
    ; __vector_default when not implemented.
`)
	for _, intr := range device.interrupts {
		fmt.Fprintf(out, "    IRQ __vector_%s\n", intr.Name)
	}
	return nil
}

func writeLD(outdir string, device *Device) error {
	// Variables for the linker script.
	out, err := os.Create(outdir + "/" + device.metadata["nameLower"].(string) + ".ld")
	if err != nil {
		return err
	}
	defer out.Close()
	t := template.Must(template.New("ld").Parse(`/* Automatically generated file. DO NOT EDIT. */
/* Generated by gen-device-avr.go from {{.file}}, see {{.descriptorSource}} */

__flash_size = 0x{{printf "%x" .flashSize}};
__ram_start = 0x{{printf "%x" .ramStart}};
__ram_size   = 0x{{printf "%x" .ramSize}};
__num_isrs   = {{.numInterrupts}};
`))
	return t.Execute(out, device.metadata)
}

func processFile(filepath, outdir string) error {
	device, err := readATDF(filepath)
	if err != nil {
		return err
	}
	err = writeGo(outdir, device)
	if err != nil {
		return err
	}
	err = writeAsm(outdir, device)
	if err != nil {
		return err
	}
	return writeLD(outdir, device)
}

func generate(indir, outdir string) error {
	// Read list of ATDF files to process.
	matches, err := filepath.Glob(indir + "/*.atdf")
	if err != nil {
		return err
	}

	// Start worker goroutines.
	var wg sync.WaitGroup
	workChan := make(chan string)
	errChan := make(chan error, 1)
	for i := 0; i < runtime.NumCPU(); i++ {
		go func() {
			for filepath := range workChan {
				err := processFile(filepath, outdir)
				wg.Done()
				if err != nil {
					// Store error to errChan if no error was stored before.
					select {
					case errChan <- err:
					default:
					}
				}
			}
		}()
	}

	// Submit all jobs to the goroutines.
	wg.Add(len(matches))
	for _, filepath := range matches {
		fmt.Println(filepath)
		workChan <- filepath
	}
	close(workChan)

	// Wait until all workers have finished.
	wg.Wait()

	// Check for an error.
	select {
	case err := <-errChan:
		return err
	default:
		return nil
	}
}

func main() {
	indir := os.Args[1]  // directory with register descriptor files (*.atdf)
	outdir := os.Args[2] // output directory
	err := generate(indir, outdir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
