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
		PeripheralInstances []struct {
			Name          string `xml:"name,attr"`
			Caption       string `xml:"caption,attr"`
			RegisterGroup struct {
				NameInModule string `xml:"name-in-module,attr"`
				Offset       string `xml:"offset,attr"`
			} `xml:"register-group"`
		} `xml:"peripherals>module>instance"`
		Interrupts []*XMLInterrupt `xml:"interrupts>interrupt"`
	} `xml:"devices>device"`
	PeripheralRegisterGroups []struct {
		Name      string `xml:"name,attr"`
		Caption   string `xml:"caption,attr"`
		Registers []struct {
			Name      string `xml:"name,attr"`
			Caption   string `xml:"caption,attr"`
			Offset    string `xml:"offset,attr"`
			Size      uint64 `xml:"size,attr"`
			Bitfields []struct {
				Name    string `xml:"name,attr"`
				Caption string `xml:"caption,attr"`
				Mask    string `xml:"mask,attr"`
			} `xml:"bitfield"`
		} `xml:"register"`
	} `xml:"modules>module>register-group"`
}

type XMLInterrupt struct {
	Index    int    `xml:"index,attr"`
	Name     string `xml:"name,attr"`
	Instance string `xml:"module-instance,attr"`
	Caption  string `xml:"caption,attr"`
}

type Device struct {
	metadata   map[string]interface{}
	interrupts []Interrupt
	types      []*PeripheralType
	instances  []*PeripheralInstance
	oldStyle   bool
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

// Peripheral instance, for example PORTB
type PeripheralInstance struct {
	Name    string
	Caption string
	Address uint64
	Type    *PeripheralType
}

// Peripheral type, for example PORT (if it's shared between different
// instances, which is the case for new-style ATDF files).
type PeripheralType struct {
	Name      string
	Caption   string
	Registers []*Register
	Instances []*PeripheralInstance
}

// Single register or struct field in a peripheral type.
type Register struct {
	Caption   string
	Name      string
	Type      string
	Offset    uint64 // offset, only for old-style ATDF files
	Bitfields []Bitfield
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

	// There appear to be two kinds of devices and ATDF files: those before
	// ~2017 and those introduced as part of the tinyAVR (1 and 2 series).
	// The newer devices are structured slightly differently, with peripherals
	// laid out more like Cortex-M chips and one or more instances per chip.
	// Older designs basically just have a bunch of registers with little
	// structure in them.
	// The code generated for these chips is quite different:
	//   * For old-style chips we'll generate a bunch of registers without
	//     peripherals (e.g. PORTB, DDRB, etc).
	//   * For new-style chips we'll generate proper peripheral structs like we
	//     do for Cortex-M chips.
	oldStyle := true
	for _, instanceEl := range device.PeripheralInstances {
		if instanceEl.RegisterGroup.NameInModule == "" {
			continue
		}
		offset, err := strconv.ParseUint(instanceEl.RegisterGroup.Offset, 0, 16)
		if err != nil {
			return nil, fmt.Errorf("failed to parse offset %#v of peripheral %s: %v", instanceEl.RegisterGroup.Offset, instanceEl.Name, err)
		}
		if offset != 0 {
			oldStyle = false
		}
	}

	// Read all peripheral types.
	var types []*PeripheralType
	typeMap := make(map[string]*PeripheralType)
	allRegisters := map[string]*Register{}
	for _, registerGroupEl := range xml.PeripheralRegisterGroups {
		var regs []*Register
		regEls := registerGroupEl.Registers
		if !oldStyle {
			// We only need to sort registers when we're generating peripheral
			// structs.
			sort.SliceStable(regEls, func(i, j int) bool {
				return regEls[i].Offset < regEls[j].Offset
			})
		}
		addReg := func(reg *Register) {
			if oldStyle {
				// Check for duplicate registers (they happen).
				if reg2 := allRegisters[reg.Name]; reg2 != nil {
					return
				}
				allRegisters[reg.Name] = reg
			}
			regs = append(regs, reg)
		}
		offset := uint64(0)
		for _, regEl := range regEls {
			regOffset, err := strconv.ParseUint(regEl.Offset, 0, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse offset %#v of register %s: %v", regEl.Offset, regEl.Name, err)
			}
			if !oldStyle {
				// Add some padding to the gap in the struct, if needed.
				if offset < regOffset {
					regs = append(regs, &Register{
						Name: "_",
						Type: fmt.Sprintf("[%d]volatile.Register8", regOffset-offset),
					})
					offset = regOffset
				}

				// Check for overlapping registers.
				if offset > regOffset {
					return nil, fmt.Errorf("register %s in peripheral %s overlaps with another register", regEl.Name, registerGroupEl.Name)
				}
			}

			var bitfields []Bitfield
			for _, bitfieldEl := range regEl.Bitfields {
				maskString := bitfieldEl.Mask
				if len(maskString) == 2 {
					// Two devices (ATtiny102 and ATtiny104) appear to have an
					// error in the bitfields, leaving out the '0x' prefix.
					maskString = "0x" + maskString
				}
				mask, err := strconv.ParseUint(maskString, 0, 32)
				if err != nil {
					return nil, fmt.Errorf("failed to parse mask %#v of bitfield %s: %v", maskString, bitfieldEl.Name, err)
				}
				name := regEl.Name + "_" + bitfieldEl.Name
				if !oldStyle {
					name = registerGroupEl.Name + "_" + name
				}
				bitfields = append(bitfields, Bitfield{
					Name:    name,
					Caption: bitfieldEl.Caption,
					Mask:    uint(mask),
				})
			}

			switch regEl.Size {
			case 1:
				addReg(&Register{
					Name:      regEl.Name,
					Type:      "volatile.Register8",
					Caption:   regEl.Caption,
					Offset:    regOffset,
					Bitfields: bitfields,
				})
			case 2:
				addReg(&Register{
					Name:    regEl.Name + "L",
					Type:    "volatile.Register8",
					Caption: regEl.Caption + " (lower bits)",
					Offset:  regOffset + 0,
				})
				addReg(&Register{
					Name:    regEl.Name + "H",
					Type:    "volatile.Register8",
					Caption: regEl.Caption + " (upper bits)",
					Offset:  regOffset + 1,
				})
			default:
				panic("todo: unknown size")
			}
			offset += regEl.Size
		}
		periphType := &PeripheralType{
			Name:      registerGroupEl.Name,
			Caption:   registerGroupEl.Caption,
			Registers: regs,
		}
		types = append(types, periphType)
		typeMap[periphType.Name] = periphType
	}

	// Read all peripheral instances.
	var instances []*PeripheralInstance
	for _, instanceEl := range device.PeripheralInstances {
		if instanceEl.RegisterGroup.NameInModule == "" {
			continue
		}
		offset, err := strconv.ParseUint(instanceEl.RegisterGroup.Offset, 0, 16)
		if err != nil {
			return nil, fmt.Errorf("failed to parse offset %#v of peripheral %s: %v", instanceEl.RegisterGroup.Offset, instanceEl.Name, err)
		}
		periphType := typeMap[instanceEl.RegisterGroup.NameInModule]
		instance := &PeripheralInstance{
			Name:    instanceEl.Name,
			Caption: instanceEl.Caption,
			Address: offset,
			Type:    periphType,
		}
		instances = append(instances, instance)
		periphType.Instances = append(periphType.Instances, instance)
	}

	ramStart := int64(0)
	ramSize := int64(0) // for devices with no RAM
	for _, ramSegmentName := range []string{"IRAM", "INTERNAL_SRAM", "SRAM"} {
		if segment, ok := memorySizes["data"].Segments[ramSegmentName]; ok {
			ramStart = segment.start
			ramSize = segment.size
		}
	}

	// Flash that is mapped into the data address space (attiny10, attiny1616,
	// etc).
	mappedFlashStart := int64(0)
	for _, name := range []string{"MAPPED_PROGMEM", "MAPPED_FLASH"} {
		if segment, ok := memorySizes["data"].Segments[name]; ok {
			mappedFlashStart = segment.start
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
			"mappedFlashStart": mappedFlashStart,
			"numInterrupts":    len(device.Interrupts),
		},
		interrupts: interrupts,
		types:      types,
		instances:  instances,
		oldStyle:   oldStyle,
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

{{if .oldStyle -}}
// Peripherals.
var (
{{- range .instances}}
	// {{.Caption}}
	{{range .Type.Registers -}}
	{{if ne .Name "_" -}}
	{{.Name}} = (*{{.Type}})(unsafe.Pointer(uintptr(0x{{printf "%x" .Offset}})))
	{{end -}}
	{{end -}}
{{end}})
{{else}}
// Peripherals instances.
var (
{{- range .instances -}}
	{{.Name}} = (*{{.Type.Name}}_Type)(unsafe.Pointer(uintptr(0x{{printf "%x" .Address}})))
	{{- if .Caption}}// {{.Caption}}{{end}}
{{end -}}
)

// Peripheral type definitions.

{{range .types}}
type {{.Name}}_Type struct {
{{range .Registers -}}
	{{.Name}} {{.Type}} {{if .Caption}} // {{.Caption}} {{end}}
{{end -}}
}
{{end}}
{{end}}
`))
	err = t.Execute(w, map[string]interface{}{
		"metadata":     device.metadata,
		"pkgName":      filepath.Base(strings.TrimRight(outdir, "/")),
		"interrupts":   device.interrupts,
		"interruptMax": maxInterruptNum,
		"instances":    device.instances,
		"types":        device.types,
		"oldStyle":     device.oldStyle,
	})
	if err != nil {
		return err
	}

	// Write bitfields.
	for _, peripheral := range device.types {
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
			fmt.Fprintf(w, "\n\t// %s", register.Name)
			if register.Caption != "" {
				fmt.Fprintf(w, ": %s", register.Caption)
			}
			fmt.Fprintf(w, "\n")
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
{{if .mappedFlashStart -}}
__mapped_flash_start = 0x{{printf "%x" .mappedFlashStart}};
{{end -}}
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
