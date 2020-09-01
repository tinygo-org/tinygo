package main

import (
	"bufio"
	"encoding/xml"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"unicode"
)

var validName = regexp.MustCompile("^[a-zA-Z0-9_]+$")
var enumBitSpecifier = regexp.MustCompile("^#[x01]+$")

type SVDFile struct {
	XMLName     xml.Name        `xml:"device"`
	Name        string          `xml:"name"`
	Description string          `xml:"description"`
	LicenseText string          `xml:"licenseText"`
	Peripherals []SVDPeripheral `xml:"peripherals>peripheral"`
}

type SVDPeripheral struct {
	Name        string `xml:"name"`
	Description string `xml:"description"`
	BaseAddress string `xml:"baseAddress"`
	GroupName   string `xml:"groupName"`
	DerivedFrom string `xml:"derivedFrom,attr"`
	Interrupts  []struct {
		Name  string `xml:"name"`
		Index int    `xml:"value"`
	} `xml:"interrupt"`
	Registers []*SVDRegister `xml:"registers>register"`
	Clusters  []*SVDCluster  `xml:"registers>cluster"`
}

type SVDRegister struct {
	Name          string      `xml:"name"`
	Description   string      `xml:"description"`
	Dim           *string     `xml:"dim"`
	DimIndex      *string     `xml:"dimIndex"`
	DimIncrement  string      `xml:"dimIncrement"`
	Size          *string     `xml:"size"`
	Fields        []*SVDField `xml:"fields>field"`
	Offset        *string     `xml:"offset"`
	AddressOffset *string     `xml:"addressOffset"`
}

type SVDField struct {
	Name             string  `xml:"name"`
	Description      string  `xml:"description"`
	Lsb              *uint32 `xml:"lsb"`
	Msb              *uint32 `xml:"msb"`
	BitOffset        *uint32 `xml:"bitOffset"`
	BitWidth         *uint32 `xml:"bitWidth"`
	BitRange         *string `xml:"bitRange"`
	EnumeratedValues []struct {
		Name        string `xml:"name"`
		Description string `xml:"description"`
		Value       string `xml:"value"`
	} `xml:"enumeratedValues>enumeratedValue"`
}

type SVDCluster struct {
	Dim           *int           `xml:"dim"`
	DimIncrement  string         `xml:"dimIncrement"`
	DimIndex      *string        `xml:"dimIndex"`
	Name          string         `xml:"name"`
	Description   string         `xml:"description"`
	Registers     []*SVDRegister `xml:"register"`
	Clusters      []*SVDCluster  `xml:"cluster"`
	AddressOffset string         `xml:"addressOffset"`
}

type Device struct {
	metadata    map[string]string
	interrupts  []*interrupt
	peripherals []*peripheral
}

type interrupt struct {
	Name            string
	HandlerName     string
	peripheralIndex int
	Value           int // interrupt number
	Description     string
}

type peripheral struct {
	Name        string
	GroupName   string
	BaseAddress uint64
	Description string
	ClusterName string
	registers   []*PeripheralField
	subtypes    []*peripheral
}

// A PeripheralField is a single field in a peripheral type. It may be a full
// peripheral or a cluster within a peripheral.
type PeripheralField struct {
	name        string
	address     uint64
	description string
	registers   []*PeripheralField // contains fields if this is a cluster
	array       int
	elementSize int
	bitfields   []Bitfield
}

type Bitfield struct {
	name        string
	description string
	value       uint32
}

func formatText(text string) string {
	text = regexp.MustCompile(`[ \t\n]+`).ReplaceAllString(text, " ") // Collapse whitespace (like in HTML)
	text = strings.Replace(text, "\\n ", "\n", -1)
	text = strings.TrimSpace(text)
	return text
}

// Replace characters that are not allowed in a symbol name with a '_'. This is
// useful to be able to process SVD files with errors.
func cleanName(text string) string {
	if !validName.MatchString(text) {
		result := make([]rune, 0, len(text))
		for _, c := range text {
			if validName.MatchString(string(c)) {
				result = append(result, c)
			} else {
				result = append(result, '_')
			}
		}
		text = string(result)
	}
	if len(text) != 0 && (text[0] >= '0' && text[0] <= '9') {
		// Identifiers may not start with a number.
		// Add an underscore instead.
		text = "_" + text
	}
	return text
}

// Read ARM SVD files.
func readSVD(path, sourceURL string) (*Device, error) {
	// Open the XML file.
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	decoder := xml.NewDecoder(f)
	device := &SVDFile{}
	err = decoder.Decode(device)
	if err != nil {
		return nil, err
	}

	peripheralDict := map[string]*peripheral{}
	groups := map[string]*peripheral{}

	interrupts := make(map[string]*interrupt)
	var peripheralsList []*peripheral

	// Some SVD files have peripheral elements derived from a peripheral that
	// comes later in the file. To make sure this works, sort the peripherals if
	// needed.
	orderedPeripherals := orderPeripherals(device.Peripherals)

	for _, periphEl := range orderedPeripherals {
		description := formatText(periphEl.Description)
		baseAddress, err := strconv.ParseUint(periphEl.BaseAddress, 0, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid base address: %w", err)
		}
		// Some group names (for example the STM32H7A3x) have an invalid
		// group name. Replace invalid characters with "_".
		groupName := cleanName(periphEl.GroupName)
		if groupName == "" {
			groupName = cleanName(periphEl.Name)
		}

		for _, interrupt := range periphEl.Interrupts {
			addInterrupt(interrupts, interrupt.Name, interrupt.Name, interrupt.Index, description)
			// As a convenience, also use the peripheral name as the interrupt
			// name. Only do that for the nrf for now, as the stm32 .svd files
			// don't always put interrupts in the correct peripheral...
			if len(periphEl.Interrupts) == 1 && strings.HasPrefix(device.Name, "nrf") {
				addInterrupt(interrupts, periphEl.Name, interrupt.Name, interrupt.Index, description)
			}
		}

		if _, ok := groups[groupName]; ok || periphEl.DerivedFrom != "" {
			var derivedFrom *peripheral
			if periphEl.DerivedFrom != "" {
				derivedFrom = peripheralDict[periphEl.DerivedFrom]
			} else {
				derivedFrom = groups[groupName]
			}
			p := &peripheral{
				Name:        periphEl.Name,
				GroupName:   derivedFrom.GroupName,
				Description: description,
				BaseAddress: baseAddress,
			}
			if p.Description == "" {
				p.Description = derivedFrom.Description
			}
			peripheralsList = append(peripheralsList, p)
			peripheralDict[p.Name] = p
			for _, subtype := range derivedFrom.subtypes {
				peripheralsList = append(peripheralsList, &peripheral{
					Name:        periphEl.Name + "_" + subtype.ClusterName,
					GroupName:   subtype.GroupName,
					Description: subtype.Description,
					BaseAddress: baseAddress,
				})
			}
			continue
		}

		p := &peripheral{
			Name:        periphEl.Name,
			GroupName:   groupName,
			Description: description,
			BaseAddress: baseAddress,
			registers:   []*PeripheralField{},
		}
		if p.GroupName == "" {
			p.GroupName = periphEl.Name
		}
		peripheralsList = append(peripheralsList, p)
		peripheralDict[periphEl.Name] = p

		if _, ok := groups[groupName]; !ok && groupName != "" {
			groups[groupName] = p
		}

		for _, register := range periphEl.Registers {
			regName := groupName // preferably use the group name
			if regName == "" {
				regName = periphEl.Name // fall back to peripheral name
			}
			p.registers = append(p.registers, parseRegister(regName, register, baseAddress, "")...)
		}
		for _, cluster := range periphEl.Clusters {
			clusterName := strings.Replace(cluster.Name, "[%s]", "", -1)
			if cluster.DimIndex != nil {
				clusterName = strings.Replace(clusterName, "%s", "", -1)
			}
			clusterPrefix := clusterName + "_"
			clusterOffset, err := strconv.ParseUint(cluster.AddressOffset, 0, 32)
			if err != nil {
				panic(err)
			}
			var dim, dimIncrement int
			if cluster.Dim == nil {
				if clusterOffset == 0 {
					// make this a separate peripheral
					cpRegisters := []*PeripheralField{}
					for _, regEl := range cluster.Registers {
						cpRegisters = append(cpRegisters, parseRegister(groupName, regEl, baseAddress, clusterName+"_")...)
					}
					// handle sub-clusters of registers
					for _, subClusterEl := range cluster.Clusters {
						subclusterName := strings.Replace(subClusterEl.Name, "[%s]", "", -1)
						subclusterPrefix := subclusterName + "_"
						subclusterOffset, err := strconv.ParseUint(subClusterEl.AddressOffset, 0, 32)
						if err != nil {
							panic(err)
						}
						subdim := *subClusterEl.Dim
						subdimIncrement, err := strconv.ParseInt(subClusterEl.DimIncrement, 0, 32)
						if err != nil {
							panic(err)
						}

						if subdim > 1 {
							subcpRegisters := []*PeripheralField{}
							subregSize := 0
							for _, regEl := range subClusterEl.Registers {
								size, err := strconv.ParseInt(*regEl.Size, 0, 32)
								if err != nil {
									panic(err)
								}
								subregSize += int(size)
								subcpRegisters = append(subcpRegisters, parseRegister(groupName, regEl, baseAddress+subclusterOffset, subclusterPrefix)...)
							}
							cpRegisters = append(cpRegisters, &PeripheralField{
								name:        subclusterName,
								address:     baseAddress + subclusterOffset,
								description: subClusterEl.Description,
								registers:   subcpRegisters,
								array:       subdim,
								elementSize: int(subdimIncrement),
							})
						} else {
							for _, regEl := range subClusterEl.Registers {
								cpRegisters = append(cpRegisters, parseRegister(regEl.Name, regEl, baseAddress+subclusterOffset, subclusterPrefix)...)
							}
						}
					}

					sort.SliceStable(cpRegisters, func(i, j int) bool {
						return cpRegisters[i].address < cpRegisters[j].address
					})
					clusterPeripheral := &peripheral{
						Name:        periphEl.Name + "_" + clusterName,
						GroupName:   groupName + "_" + clusterName,
						Description: description + " - " + clusterName,
						ClusterName: clusterName,
						BaseAddress: baseAddress,
						registers:   cpRegisters,
					}
					peripheralsList = append(peripheralsList, clusterPeripheral)
					peripheralDict[clusterPeripheral.Name] = clusterPeripheral
					p.subtypes = append(p.subtypes, clusterPeripheral)
					continue
				}
				dim = -1
				dimIncrement = -1
			} else {
				dim = *cluster.Dim
				if dim == 1 {
					dimIncrement = -1
				} else {
					inc, err := strconv.ParseUint(cluster.DimIncrement, 0, 32)
					if err != nil {
						panic(err)
					}
					dimIncrement = int(inc)
				}
			}
			clusterRegisters := []*PeripheralField{}
			for _, regEl := range cluster.Registers {
				regName := groupName
				if regName == "" {
					regName = periphEl.Name
				}
				clusterRegisters = append(clusterRegisters, parseRegister(regName, regEl, baseAddress+clusterOffset, clusterPrefix)...)
			}
			sort.SliceStable(clusterRegisters, func(i, j int) bool {
				return clusterRegisters[i].address < clusterRegisters[j].address
			})
			if dimIncrement == -1 {
				lastReg := clusterRegisters[len(clusterRegisters)-1]
				lastAddress := lastReg.address
				if lastReg.array != -1 {
					lastAddress = lastReg.address + uint64(lastReg.array*lastReg.elementSize)
				}
				firstAddress := clusterRegisters[0].address
				dimIncrement = int(lastAddress - firstAddress)
			}

			if !unicode.IsUpper(rune(clusterName[0])) && !unicode.IsDigit(rune(clusterName[0])) {
				clusterName = strings.ToUpper(clusterName)
			}

			p.registers = append(p.registers, &PeripheralField{
				name:        clusterName,
				address:     baseAddress + clusterOffset,
				description: cluster.Description,
				registers:   clusterRegisters,
				array:       dim,
				elementSize: dimIncrement,
			})
		}
		sort.SliceStable(p.registers, func(i, j int) bool {
			return p.registers[i].address < p.registers[j].address
		})
	}

	// Make a sorted list of interrupts.
	interruptList := make([]*interrupt, 0, len(interrupts))
	for _, intr := range interrupts {
		interruptList = append(interruptList, intr)
	}
	sort.SliceStable(interruptList, func(i, j int) bool {
		if interruptList[i].Value != interruptList[j].Value {
			return interruptList[i].Value < interruptList[j].Value
		}
		return interruptList[i].peripheralIndex < interruptList[j].peripheralIndex
	})

	// Properly format the license block, with comments.
	licenseBlock := ""
	if text := formatText(device.LicenseText); text != "" {
		licenseBlock = "//     " + strings.Replace(text, "\n", "\n//     ", -1)
		licenseBlock = regexp.MustCompile(`\s+\n`).ReplaceAllString(licenseBlock, "\n")
	}

	return &Device{
		metadata: map[string]string{
			"file":             filepath.Base(path),
			"descriptorSource": sourceURL,
			"name":             device.Name,
			"nameLower":        strings.ToLower(device.Name),
			"description":      strings.TrimSpace(device.Description),
			"licenseBlock":     licenseBlock,
		},
		interrupts:  interruptList,
		peripherals: peripheralsList,
	}, nil
}

// orderPeripherals sorts the peripherals so that derived peripherals come after
// base peripherals. This is necessary for some SVD files.
func orderPeripherals(input []SVDPeripheral) []*SVDPeripheral {
	var sortedPeripherals []*SVDPeripheral
	var missingBasePeripherals []*SVDPeripheral
	knownBasePeripherals := map[string]struct{}{}
	for i := range input {
		p := &input[i]
		groupName := p.GroupName
		if groupName == "" {
			groupName = p.Name
		}
		knownBasePeripherals[groupName] = struct{}{}
		if p.DerivedFrom != "" {
			if _, ok := knownBasePeripherals[p.DerivedFrom]; !ok {
				missingBasePeripherals = append(missingBasePeripherals, p)
				continue
			}
		}
		sortedPeripherals = append(sortedPeripherals, p)
	}

	// Let's hope all base peripherals are now included.
	sortedPeripherals = append(sortedPeripherals, missingBasePeripherals...)

	return sortedPeripherals
}

func addInterrupt(interrupts map[string]*interrupt, name, interruptName string, index int, description string) {
	if _, ok := interrupts[name]; ok {
		if interrupts[name].Value != index {
			// Note: some SVD files like the one for STM32H7x7 contain mistakes.
			// Instead of throwing an error, simply log it.
			fmt.Fprintf(os.Stderr, "interrupt with the same name has different indexes: %s (%d vs %d)",
				name, interrupts[name].Value, index)
		}
		parts := strings.Split(interrupts[name].Description, " // ")
		hasDescription := false
		for _, part := range parts {
			if part == description {
				hasDescription = true
			}
		}
		if !hasDescription {
			interrupts[name].Description += " // " + description
		}
	} else {
		interrupts[name] = &interrupt{
			Name:            name,
			HandlerName:     interruptName + "_IRQHandler",
			peripheralIndex: len(interrupts),
			Value:           index,
			Description:     description,
		}
	}
}

func parseBitfields(groupName, regName string, fieldEls []*SVDField, bitfieldPrefix string) []Bitfield {
	var fields []Bitfield
	for _, fieldEl := range fieldEls {
		// Some bitfields (like the STM32H7x7) contain invalid bitfield
		// names like "CNT[31]". Replace invalid characters with "_" when
		// needed.
		fieldName := cleanName(fieldEl.Name)
		if !unicode.IsUpper(rune(fieldName[0])) && !unicode.IsDigit(rune(fieldName[0])) {
			fieldName = strings.ToUpper(fieldName)
		}

		// Find the lsb/msb that is encoded in various ways.
		// Standards are great, that's why there are so many to choose from!
		var lsb, msb uint32
		if fieldEl.Lsb != nil && fieldEl.Msb != nil {
			// try to use lsb/msb tags
			lsb = *fieldEl.Lsb
			msb = *fieldEl.Msb
		} else if fieldEl.BitOffset != nil && fieldEl.BitWidth != nil {
			// try to use bitOffset/bitWidth tags
			lsb = *fieldEl.BitOffset
			msb = *fieldEl.BitWidth + lsb - 1
		} else if fieldEl.BitRange != nil {
			// try use bitRange
			// example string: "[20:16]"
			parts := strings.Split(strings.Trim(*fieldEl.BitRange, "[]"), ":")
			l, err := strconv.ParseUint(parts[1], 0, 32)
			if err != nil {
				panic(err)
			}
			lsb = uint32(l)
			m, err := strconv.ParseUint(parts[0], 0, 32)
			if err != nil {
				panic(err)
			}
			msb = uint32(m)
		} else {
			// this is an error. what to do?
			fmt.Fprintln(os.Stderr, "unable to find lsb/msb in field:", fieldName)
			continue
		}

		fields = append(fields, Bitfield{
			name:        fmt.Sprintf("%s_%s%s_%s_Pos", groupName, bitfieldPrefix, regName, fieldName),
			description: fmt.Sprintf("Position of %s field.", fieldName),
			value:       lsb,
		})
		fields = append(fields, Bitfield{
			name:        fmt.Sprintf("%s_%s%s_%s_Msk", groupName, bitfieldPrefix, regName, fieldName),
			description: fmt.Sprintf("Bit mask of %s field.", fieldName),
			value:       (0xffffffff >> (31 - (msb - lsb))) << lsb,
		})
		if lsb == msb { // single bit
			fields = append(fields, Bitfield{
				name:        fmt.Sprintf("%s_%s%s_%s", groupName, bitfieldPrefix, regName, fieldName),
				description: fmt.Sprintf("Bit %s.", fieldName),
				value:       1 << lsb,
			})
		}
		for _, enumEl := range fieldEl.EnumeratedValues {
			enumName := enumEl.Name
			if strings.EqualFold(enumName, "reserved") || !validName.MatchString(enumName) {
				continue
			}
			if !unicode.IsUpper(rune(enumName[0])) && !unicode.IsDigit(rune(enumName[0])) {
				enumName = strings.ToUpper(enumName)
			}
			enumDescription := strings.Replace(enumEl.Description, "\n", " ", -1)
			enumValue, err := strconv.ParseUint(enumEl.Value, 0, 32)
			if err != nil {
				if enumBitSpecifier.MatchString(enumEl.Value) {
					// NXP SVDs use the form #xx1x, #x0xx, etc for values
					enumValue, err = strconv.ParseUint(strings.Replace(enumEl.Value[1:], "x", "0", -1), 2, 32)
					if err != nil {
						panic(err)
					}
				} else {
					panic(err)
				}
			}
			fields = append(fields, Bitfield{
				name:        fmt.Sprintf("%s_%s%s_%s_%s", groupName, bitfieldPrefix, regName, fieldName, enumName),
				description: enumDescription,
				value:       uint32(enumValue),
			})
		}
	}
	return fields
}

type Register struct {
	element     *SVDRegister
	baseAddress uint64
}

func NewRegister(element *SVDRegister, baseAddress uint64) *Register {
	return &Register{
		element:     element,
		baseAddress: baseAddress,
	}
}

func (r *Register) name() string {
	return strings.Replace(r.element.Name, "[%s]", "", -1)
}

func (r *Register) description() string {
	return strings.Replace(r.element.Description, "\n", " ", -1)
}

func (r *Register) address() uint64 {
	offsetString := r.element.Offset
	if offsetString == nil {
		offsetString = r.element.AddressOffset
	}
	addr, err := strconv.ParseUint(*offsetString, 0, 32)
	if err != nil {
		panic(err)
	}
	return r.baseAddress + addr
}

func (r *Register) dim() int {
	if r.element.Dim == nil {
		return -1 // no dim elements
	}
	dim, err := strconv.ParseInt(*r.element.Dim, 0, 32)
	if err != nil {
		panic(err)
	}
	return int(dim)
}

func (r *Register) dimIndex() []string {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("register", r.name())
			panic(err)
		}
	}()

	dim := r.dim()
	if r.element.DimIndex == nil {
		if dim <= 0 {
			return nil
		}

		idx := make([]string, dim)
		for i := range idx {
			idx[i] = strconv.FormatInt(int64(i), 10)
		}
		return idx
	}

	t := strings.Split(*r.element.DimIndex, "-")
	if len(t) == 2 {
		x, err := strconv.ParseInt(t[0], 0, 32)
		if err != nil {
			panic(err)
		}
		y, err := strconv.ParseInt(t[1], 0, 32)
		if err != nil {
			panic(err)
		}

		if x < 0 || y < x || y-x != int64(dim-1) {
			panic("invalid dimIndex")
		}

		idx := make([]string, dim)
		for i := x; i <= y; i++ {
			idx[i-x] = strconv.FormatInt(i, 10)
		}
		return idx
	} else if len(t) > 2 {
		panic("invalid dimIndex")
	}

	s := strings.Split(*r.element.DimIndex, ",")
	if len(s) != dim {
		panic("invalid dimIndex")
	}

	return s
}

func (r *Register) size() int {
	if r.element.Size != nil {
		size, err := strconv.ParseInt(*r.element.Size, 0, 32)
		if err != nil {
			panic(err)
		}
		return int(size) / 8
	}
	return 4
}

func parseRegister(groupName string, regEl *SVDRegister, baseAddress uint64, bitfieldPrefix string) []*PeripheralField {
	reg := NewRegister(regEl, baseAddress)

	if reg.dim() != -1 {
		dimIncrement, err := strconv.ParseUint(regEl.DimIncrement, 0, 32)
		if err != nil {
			panic(err)
		}
		if strings.Contains(reg.name(), "%s") {
			// a "spaced array" of registers, special processing required
			// we need to generate a separate register for each "element"
			var results []*PeripheralField
			for i, j := range reg.dimIndex() {
				regAddress := reg.address() + (uint64(i) * dimIncrement)
				results = append(results, &PeripheralField{
					name:        strings.ToUpper(strings.Replace(reg.name(), "%s", j, -1)),
					address:     regAddress,
					description: reg.description(),
					array:       -1,
					elementSize: reg.size(),
				})
			}
			// set first result bitfield
			shortName := strings.ToUpper(strings.Replace(strings.Replace(reg.name(), "_%s", "", -1), "%s", "", -1))
			results[0].bitfields = parseBitfields(groupName, shortName, regEl.Fields, bitfieldPrefix)
			return results
		}
	}
	regName := reg.name()
	if !unicode.IsUpper(rune(regName[0])) && !unicode.IsDigit(rune(regName[0])) {
		regName = strings.ToUpper(regName)
	}
	regName = cleanName(regName)

	bitfields := parseBitfields(groupName, regName, regEl.Fields, bitfieldPrefix)
	return []*PeripheralField{&PeripheralField{
		name:        regName,
		address:     reg.address(),
		description: reg.description(),
		bitfields:   bitfields,
		array:       reg.dim(),
		elementSize: reg.size(),
	}}
}

// The Go module for this device.
func writeGo(outdir string, device *Device, interruptSystem string) error {
	outf, err := os.Create(filepath.Join(outdir, device.metadata["nameLower"]+".go"))
	if err != nil {
		return err
	}
	defer outf.Close()
	w := bufio.NewWriter(outf)

	maxInterruptValue := 0
	for _, intr := range device.interrupts {
		if intr.Value > maxInterruptValue {
			maxInterruptValue = intr.Value
		}
	}

	t := template.Must(template.New("go").Parse(`// Automatically generated file. DO NOT EDIT.
// Generated by gen-device-svd.go from {{.metadata.file}}, see {{.metadata.descriptorSource}}

// +build {{.pkgName}},{{.metadata.nameLower}}

// {{.metadata.description}}
//
{{.metadata.licenseBlock}}
package {{.pkgName}}

import (
{{if eq .interruptSystem "hardware"}}"runtime/interrupt"{{end}}
	"runtime/volatile"
	"unsafe"
)

// Some information about this device.
const (
	DEVICE	 = "{{.metadata.name}}"
)

// Interrupt numbers.
const ({{range .interrupts}}
	IRQ_{{.Name}} = {{.Value}} // {{.Description}}{{end}}
	IRQ_max = {{.interruptMax}} // Highest interrupt number on this device.
)

{{if eq .interruptSystem "hardware"}}
// Map interrupt numbers to function names.
// These aren't real calls, they're removed by the compiler.
var ({{range .interrupts}}
	_ = interrupt.Register(IRQ_{{.Name}}, "{{.HandlerName}}"){{end}}
)
{{end}}

// Peripherals.
var (
{{range .peripherals}}	{{.Name}} = (*{{.GroupName}}_Type)(unsafe.Pointer(uintptr(0x{{printf "%x" .BaseAddress}}))) // {{.Description}}
{{end}})
`))
	err = t.Execute(w, map[string]interface{}{
		"metadata":        device.metadata,
		"interrupts":      device.interrupts,
		"peripherals":     device.peripherals,
		"pkgName":         filepath.Base(strings.TrimRight(outdir, "/")),
		"interruptMax":    maxInterruptValue,
		"interruptSystem": interruptSystem,
	})
	if err != nil {
		return err
	}

	// Define peripheral struct types.
	for _, peripheral := range device.peripherals {
		if peripheral.registers == nil {
			// This peripheral was derived from another peripheral. No new type
			// needs to be defined for it.
			continue
		}
		fmt.Fprintf(w, "\n// %s\ntype %s_Type struct {\n", peripheral.Description, peripheral.GroupName)
		address := peripheral.BaseAddress
		for _, register := range peripheral.registers {
			if register.registers == nil && address > register.address {
				// In Nordic SVD files, these registers are deprecated or
				// duplicates, so can be ignored.
				//fmt.Fprintf(os.Stderr, "skip: %s.%s 0x%x - 0x%x %d\n", peripheral.Name, register.name, address, register.address, register.elementSize)
				continue
			}

			var regType string
			switch register.elementSize {
			case 8:
				regType = "volatile.Register64"
			case 4:
				regType = "volatile.Register32"
			case 2:
				regType = "volatile.Register16"
			case 1:
				regType = "volatile.Register8"
			default:
				regType = "volatile.Register32"
			}

			// insert padding, if needed
			if address < register.address {
				bytesNeeded := register.address - address
				if bytesNeeded == 1 {
					w.WriteString("\t_ byte\n")
				} else {
					fmt.Fprintf(w, "\t_ [%d]byte\n", bytesNeeded)
				}
				address = register.address
			}

			lastCluster := false
			if register.registers != nil {
				// This is a cluster, not a register. Create the cluster type.
				regType = "struct {\n"
				subaddress := register.address
				for _, subregister := range register.registers {
					var subregType string
					switch subregister.elementSize {
					case 8:
						subregType = "volatile.Register64"
					case 4:
						subregType = "volatile.Register32"
					case 2:
						subregType = "volatile.Register16"
					case 1:
						subregType = "volatile.Register8"
					}
					if subregType == "" {
						panic("unknown element size")
					}

					if subregister.array != -1 {
						subregType = fmt.Sprintf("[%d]%s", subregister.array, subregType)
					}
					if subaddress != subregister.address {
						bytesNeeded := subregister.address - subaddress
						if bytesNeeded == 1 {
							regType += "\t\t_ byte\n"
						} else {
							regType += fmt.Sprintf("\t\t_ [%d]byte\n", bytesNeeded)
						}
						subaddress += bytesNeeded
					}
					var subregSize uint64
					if subregister.array != -1 {
						subregSize = uint64(subregister.array * subregister.elementSize)
					} else {
						subregSize = uint64(subregister.elementSize)
					}
					subaddress += subregSize
					regType += fmt.Sprintf("\t\t%s %s\n", subregister.name, subregType)
				}
				if register.array != -1 {
					if subaddress != register.address+uint64(register.elementSize) {
						bytesNeeded := (register.address + uint64(register.elementSize)) - subaddress
						if bytesNeeded == 1 {
							regType += "\t_ byte\n"
						} else {
							regType += fmt.Sprintf("\t_ [%d]byte\n", bytesNeeded)
						}
					}
				} else {
					lastCluster = true
				}
				regType += "\t}"
				address = subaddress
			}

			if register.array != -1 {
				regType = fmt.Sprintf("[%d]%s", register.array, regType)
			}
			fmt.Fprintf(w, "\t%s %s // 0x%X\n", register.name, regType, register.address-peripheral.BaseAddress)

			// next address
			if lastCluster {
				lastCluster = false
			} else if register.array != -1 {
				address = register.address + uint64(register.elementSize*register.array)
			} else {
				address = register.address + uint64(register.elementSize)
			}
		}
		w.WriteString("}\n")
	}

	// Define bitfields.
	for _, peripheral := range device.peripherals {
		if peripheral.registers == nil {
			// This peripheral was derived from another peripheral. Bitfields are
			// already defined.
			continue
		}
		fmt.Fprintf(w, "\n// Bitfields for %s: %s\nconst(", peripheral.Name, peripheral.Description)
		for _, register := range peripheral.registers {
			if len(register.bitfields) != 0 {
				writeGoRegisterBitfields(w, register, register.name)
			}
			if register.registers == nil {
				continue
			}
			for _, subregister := range register.registers {
				writeGoRegisterBitfields(w, subregister, register.name+"."+subregister.name)
			}
		}
		w.WriteString(")\n")
	}

	return w.Flush()
}

func writeGoRegisterBitfields(w *bufio.Writer, register *PeripheralField, name string) {
	w.WriteString("\n\t// " + name)
	if register.description != "" {
		w.WriteString(": " + register.description)
	}
	w.WriteByte('\n')
	for _, bitfield := range register.bitfields {
		fmt.Fprintf(w, "\t%s = 0x%x", bitfield.name, bitfield.value)
		if bitfield.description != "" {
			w.WriteString(" // " + bitfield.description)
		}
		w.WriteByte('\n')
	}
}

// The interrupt vector, which is hard to write directly in Go.
func writeAsm(outdir string, device *Device) error {
	outf, err := os.Create(filepath.Join(outdir, device.metadata["nameLower"]+".s"))
	if err != nil {
		return err
	}
	defer outf.Close()
	w := bufio.NewWriter(outf)

	t := template.Must(template.New("go").Parse(`// Automatically generated file. DO NOT EDIT.
// Generated by gen-device-svd.go from {{.file}}, see {{.descriptorSource}}

// {{.description}}
//
{{.licenseBlock}}

.syntax unified

// This is the default handler for interrupts, if triggered but not defined.
.section .text.Default_Handler
.global  Default_Handler
.type    Default_Handler, %function
Default_Handler:
    wfe
    b    Default_Handler
.size Default_Handler, .-Default_Handler

// Avoid the need for repeated .weak and .set instructions.
.macro IRQ handler
    .weak  \handler
    .set   \handler, Default_Handler
.endm

// Must set the "a" flag on the section:
// https://svnweb.freebsd.org/base/stable/11/sys/arm/arm/locore-v4.S?r1=321049&r2=321048&pathrev=321049
// https://sourceware.org/binutils/docs/as/Section.html#ELF-Version
.section .isr_vector, "a", %progbits
.global  __isr_vector
    // Interrupt vector as defined by Cortex-M, starting with the stack top.
    // On reset, SP is initialized with *0x0 and PC is loaded with *0x4, loading
    // _stack_top and Reset_Handler.
    .long _stack_top
    .long Reset_Handler
    .long NMI_Handler
    .long HardFault_Handler
    .long MemoryManagement_Handler
    .long BusFault_Handler
    .long UsageFault_Handler
    .long 0
    .long 0
    .long 0
    .long 0
    .long SVC_Handler
    .long DebugMon_Handler
    .long 0
    .long PendSV_Handler
    .long SysTick_Handler

    // Extra interrupts for peripherals defined by the hardware vendor.
`))
	err = t.Execute(w, device.metadata)
	if err != nil {
		return err
	}
	num := 0
	for _, intr := range device.interrupts {
		if intr.Value == num-1 {
			continue
		}
		if intr.Value < num {
			panic("interrupt numbers are not sorted")
		}
		for intr.Value > num {
			w.WriteString("    .long 0\n")
			num++
		}
		num++
		fmt.Fprintf(w, "    .long %s\n", intr.HandlerName)
	}

	w.WriteString(`
    // Define default implementations for interrupts, redirecting to
    // Default_Handler when not implemented.
    IRQ NMI_Handler
    IRQ HardFault_Handler
    IRQ MemoryManagement_Handler
    IRQ BusFault_Handler
    IRQ UsageFault_Handler
    IRQ SVC_Handler
    IRQ DebugMon_Handler
    IRQ PendSV_Handler
    IRQ SysTick_Handler
`)
	for _, intr := range device.interrupts {
		fmt.Fprintf(w, "    IRQ %s_IRQHandler\n", intr.Name)
	}
	return w.Flush()
}

func generate(indir, outdir, sourceURL, interruptSystem string) error {
	if _, err := os.Stat(indir); os.IsNotExist(err) {
		fmt.Fprintln(os.Stderr, "cannot find input directory:", indir)
		os.Exit(1)
	}
	os.MkdirAll(outdir, 0777)

	infiles, err := filepath.Glob(filepath.Join(indir, "*.svd"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "could not read .svd files:", err)
		os.Exit(1)
	}
	sort.Strings(infiles)
	for _, infile := range infiles {
		fmt.Println(infile)
		device, err := readSVD(infile, sourceURL)
		if err != nil {
			return fmt.Errorf("failed to read: %w", err)
		}
		err = writeGo(outdir, device, interruptSystem)
		if err != nil {
			return fmt.Errorf("failed to write Go file: %w", err)
		}
		switch interruptSystem {
		case "software":
			// Nothing to do.
		case "hardware":
			err = writeAsm(outdir, device)
			if err != nil {
				return fmt.Errorf("failed to write assembly file: %w", err)
			}
		default:
			return fmt.Errorf("unknown interrupt system: %s", interruptSystem)
		}
	}
	return nil
}

func main() {
	sourceURL := flag.String("source", "<unknown>", "source SVD file")
	interruptSystem := flag.String("interrupts", "hardware", "interrupt system in use (software, hardware)")
	flag.Parse()
	if flag.NArg() != 2 {
		fmt.Fprintln(os.Stderr, "provide exactly two arguments: input directory (with .svd files) and output directory for generated files")
		flag.PrintDefaults()
		return
	}
	indir := flag.Arg(0)
	outdir := flag.Arg(1)
	err := generate(indir, outdir, *sourceURL, *interruptSystem)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
