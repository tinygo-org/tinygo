// Command gen-py32-targets generates PY32 target and linker definitions from
// normalized CMSIS-Pack device metadata.
package main

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type device struct {
	Part       string
	Family     string
	Core       string
	FlashStart uint64
	FlashSize  uint64
	RAMStart   uint64
	RAMSize    uint64
}

type target struct {
	Inherits     []string `json:"inherits"`
	BuildTags    []string `json:"build-tags,omitempty"`
	LinkerScript string   `json:"linkerscript,omitempty"`
	ExtraFiles   []string `json:"extra-files,omitempty"`
	FlashCommand string   `json:"flash-command,omitempty"`
}

var orphanFamilies = map[string]string{
	// This SVD is present in PY32F0xx_DFP, but that pack has no matching PDSC
	// device entry from which a concrete memory configuration can be derived.
	"py32f001xx": "m0",
}

var noGPIOAFRH = map[string]bool{
	"py32f001cxx": true,
	"py32f002bxx": true,
	"py32f002cxx": true,
	"py32f002xxx": true,
	"py32f002zxx": true,
	"py32l020xx":  true,
	"py32m010xx":  true,
}

var gpioOSPDDER = familySet("py32e407xx", "py32f410xx", "py32l090xx", "py32t090xx", "py32t092xx")
var gpioClockAHB2 = familySet("py32e407xx", "py32f403xx")
var gpioClockAHB = familySet("py32f410xx")
var hsiFSOP = familySet("py32l090xx", "py32t090xx", "py32t092xx")
var hsiFSLiteral4 = familySet("py32f001cxx", "py32f002cxx")
var noHSIFS = familySet("py32e407xx", "py32f403xx", "py32f410xx")
var uartType = familySet("py32t020xx")
var usartSplitData = familySet("py32e407xx")
var usartUnnumbered = familySet("py32f410xx")
var uartClockAPB2 = familySet("py32e407xx", "py32f403xx", "py32f410xx")
var uartNoInterrupt = familySet("py32f410xx")

func familySet(names ...string) map[string]bool {
	set := make(map[string]bool, len(names))
	for _, name := range names {
		set[name] = true
	}
	return set
}

func familyBuildTags(family string) []string {
	tags := []string{family}
	features := []struct {
		name string
		set  map[string]bool
	}{
		{"py32_no_gpio_afrh", noGPIOAFRH},
		{"py32_gpio_ospdder", gpioOSPDDER},
		{"py32_gpio_clock_ahb2", gpioClockAHB2},
		{"py32_gpio_clock_ahb", gpioClockAHB},
		{"py32_hsi_fs_op", hsiFSOP},
		{"py32_hsi_fs_literal4", hsiFSLiteral4},
		{"py32_no_hsi_fs", noHSIFS},
		{"py32_uart_type", uartType},
		{"py32_usart_split_data", usartSplitData},
		{"py32_usart_unnumbered", usartUnnumbered},
		{"py32_uart_clock_apb2", uartClockAPB2},
		{"py32_uart_no_interrupt", uartNoInterrupt},
	}
	for _, feature := range features {
		if feature.set[family] {
			tags = append(tags, feature.name)
		}
	}
	return tags
}

var flashCommands = map[string]string{
	"py32f002ax5": "pyocd load -t py32f002ax5 {bin}",
	"py32f002bx5": "pyocd load -t py32f002bx5 {bin}",
	"py32f003x4":  "pyocd load -t py32f003x4 {bin}",
	"py32f003x6":  "pyocd load -t py32f003x6 {bin}",
	"py32f003x7":  "pyocd load -t py32f003x7 {bin}",
	"py32f003x8":  "pyocd load -t py32f003x8 {bin}",
	"py32f030x4":  "pyocd load -t py32f030x4 {bin}",
	"py32f030x6":  "pyocd load -t py32f030x6 {bin}",
	"py32f030x7":  "pyocd load -t py32f030x7 {bin}",
	"py32f030x8":  "pyocd load -t py32f030x8 {bin}",
}

func main() {
	table := flag.String("table", "tools/gen-py32-targets/devices.csv", "normalized device table")
	out := flag.String("out", "targets", "target output directory")
	flag.Parse()

	devices, err := readDevices(*table)
	if err != nil {
		fatal(err)
	}
	if err := generate(*out, devices); err != nil {
		fatal(err)
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "gen-py32-targets:", err)
	os.Exit(1)
}

func readDevices(path string) ([]device, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	header, err := reader.Read()
	if err != nil {
		return nil, err
	}
	wantHeader := []string{"part", "family", "core", "flash_start", "flash_size", "ram_start", "ram_size"}
	if strings.Join(header, ",") != strings.Join(wantHeader, ",") {
		return nil, fmt.Errorf("unexpected CSV header %q", strings.Join(header, ","))
	}

	var devices []device
	seen := make(map[string]bool)
	for line := 2; ; line++ {
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", line, err)
		}
		if seen[record[0]] {
			return nil, fmt.Errorf("line %d: duplicate part %q", line, record[0])
		}
		seen[record[0]] = true
		if record[2] != "m0" && record[2] != "m4" {
			return nil, fmt.Errorf("line %d: unsupported core %q", line, record[2])
		}
		values := make([]uint64, 4)
		for i, field := range record[3:] {
			values[i], err = strconv.ParseUint(field, 0, 64)
			if err != nil || values[i] == 0 {
				return nil, fmt.Errorf("line %d: invalid memory value %q", line, field)
			}
		}
		devices = append(devices, device{
			Part: record[0], Family: record[1], Core: record[2],
			FlashStart: values[0], FlashSize: values[1],
			RAMStart: values[2], RAMSize: values[3],
		})
	}
	return devices, nil
}

func systemStackSize(ramSize uint64) uint64 {
	switch {
	case ramSize <= 4*1024:
		return 1024
	case ramSize <= 16*1024:
		return 2 * 1024
	default:
		return 4 * 1024
	}
}

func generate(out string, devices []device) error {
	if err := os.MkdirAll(out, 0o755); err != nil {
		return err
	}
	families := make(map[string]string, len(orphanFamilies))
	for name, core := range orphanFamilies {
		families[name] = core
	}
	for _, device := range devices {
		if core, ok := families[device.Family]; ok && core != device.Core {
			return fmt.Errorf("family %q mixes %s and %s cores", device.Family, core, device.Core)
		}
		families[device.Family] = device.Core
	}

	familyNames := make([]string, 0, len(families))
	for name := range families {
		familyNames = append(familyNames, name)
	}
	sort.Strings(familyNames)
	for _, family := range familyNames {
		base := "py32"
		if families[family] == "m4" {
			base = "py32-m4"
		}
		spec := target{
			Inherits:   []string{base},
			BuildTags:  familyBuildTags(family),
			ExtraFiles: []string{"src/device/py32/" + family + ".s"},
		}
		if err := writeJSON(filepath.Join(out, family+".json"), spec); err != nil {
			return err
		}
	}

	for _, device := range devices {
		spec := target{
			Inherits:     []string{device.Family},
			BuildTags:    []string{device.Part},
			LinkerScript: "targets/" + device.Part + ".ld",
			FlashCommand: flashCommands[device.Part],
		}
		if err := writeJSON(filepath.Join(out, device.Part+".json"), spec); err != nil {
			return err
		}
		stackSize := systemStackSize(device.RAMSize)
		linker := fmt.Sprintf("MEMORY\n{\n  FLASH_TEXT (rx) : ORIGIN = %#x, LENGTH = %#x\n  RAM (xrw)       : ORIGIN = %#x, LENGTH = %#x\n}\n\n_stack_size = %dK;\n\nINCLUDE \"targets/arm.ld\"\n", device.FlashStart, device.FlashSize, device.RAMStart, device.RAMSize, stackSize/1024)
		if err := os.WriteFile(filepath.Join(out, device.Part+".ld"), []byte(linker), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "    ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}
