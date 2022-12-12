package main

import (
	"bufio"
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"io/fs"
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
	XMLName     xml.Name `xml:"device"`
	Name        string   `xml:"name"`
	Description string   `xml:"description"`
	LicenseText string   `xml:"licenseText"`
	CPU         *struct {
		Name         string `xml:"name"`
		FPUPresent   bool   `xml:"fpuPresent"`
		NVICPrioBits int    `xml:"nvicPrioBits"`
	} `xml:"cpu"`
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
	EnumeratedValues struct {
		DerivedFrom     string `xml:"derivedFrom,attr"`
		Name            string `xml:"name"`
		EnumeratedValue []struct {
			Name        string `xml:"name"`
			Description string `xml:"description"`
			Value       string `xml:"value"`
		} `xml:"enumeratedValue"`
	} `xml:"enumeratedValues"`
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
	Metadata       *Metadata
	Interrupts     []*Interrupt
	Peripherals    []*Peripheral
	PeripheralDict map[string]*Peripheral
}

type Metadata struct {
	File             string
	DescriptorSource string
	Name             string
	NameLower        string
	Description      string
	LicenseBlock     string

	HasCPUInfo   bool // set if the following fields are populated
	CPUName      string
	FPUPresent   bool
	NVICPrioBits int
}

type Interrupt struct {
	Name            string
	HandlerName     string
	PeripheralIndex int
	Value           int // interrupt number
	Description     string
}

type Peripheral struct {
	Name        string
	GroupName   string
	BaseAddress uint64
	Description string
	ClusterName string
	Registers   []*PeripheralField
	Subtypes    []*Peripheral
}

// A PeripheralField is a single field in a peripheral type. It may be a full
// peripheral or a cluster within a peripheral.
type PeripheralField struct {
	Name         string
	Address      uint64
	Description  string
	Registers    []*PeripheralField // contains fields if this is a cluster
	Array        int
	ElementSize  int
	Constants    []Constant
	ShortName    string     // name stripped of "spaced array" suffix
	Bitfields    []Bitfield // set of bit-fields provided by this
	HasBitfields bool       // set true when Bitfields was set for a first PeripheralField of "spaced array".
}

type Constant struct {
	Name        string
	Description string
	Value       uint64
}

type Bitfield struct {
	Name   string
	Offset uint32
	Mask   uint32
}

func formatText(text string) string {
	text = regexp.MustCompile(`[ \t\n]+`).ReplaceAllString(text, " ") // Collapse whitespace (like in HTML)
	text = strings.ReplaceAll(text, "\\n ", "\n")
	text = strings.TrimSpace(text)
	return text
}

func isMultiline(s string) bool {
	return strings.Index(s, "\n") >= 0
}

func splitLine(s string) []string {
	return strings.Split(s, "\n")
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

func processSubCluster(p *Peripheral, cluster *SVDCluster, clusterOffset uint64, clusterName string, peripheralDict map[string]*Peripheral) []*Peripheral {
	var peripheralsList []*Peripheral
	clusterPrefix := clusterName + "_"
	cpRegisters := []*PeripheralField{}

	for _, regEl := range cluster.Registers {
		cpRegisters = append(cpRegisters, parseRegister(p.GroupName, regEl, p.BaseAddress+clusterOffset, clusterPrefix)...)
	}
	// handle sub-clusters of registers
	for _, subClusterEl := range cluster.Clusters {
		subclusterName := strings.ReplaceAll(subClusterEl.Name, "[%s]", "")
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
				if regEl.Size == nil {
					continue
				}
				size, err := strconv.ParseInt(*regEl.Size, 0, 32)
				if err != nil {
					panic(err)
				}
				subregSize += int(size)
				subcpRegisters = append(subcpRegisters, parseRegister(p.GroupName, regEl, p.BaseAddress+clusterOffset+subclusterOffset, subclusterPrefix)...)
			}
			peripheralsList = append(peripheralsList, processSubCluster(p, subClusterEl, clusterOffset+subclusterOffset, subclusterName, peripheralDict)...)

			cpRegisters = append(cpRegisters, &PeripheralField{
				Name:        subclusterName,
				Address:     p.BaseAddress + clusterOffset + subclusterOffset,
				Description: subClusterEl.Description,
				Registers:   subcpRegisters,
				Array:       subdim,
				ElementSize: int(subdimIncrement),
				ShortName:   clusterPrefix + subclusterName,
			})
		} else {
			for _, regEl := range subClusterEl.Registers {
				cpRegisters = append(cpRegisters, parseRegister(regEl.Name, regEl, p.BaseAddress+clusterOffset+subclusterOffset, subclusterPrefix)...)
			}
		}
	}

	sort.SliceStable(cpRegisters, func(i, j int) bool {
		return cpRegisters[i].Address < cpRegisters[j].Address
	})
	clusterPeripheral := &Peripheral{
		Name:        p.Name + "_" + clusterName,
		GroupName:   p.GroupName + "_" + clusterName,
		Description: p.Description + " - " + clusterName,
		ClusterName: clusterName,
		BaseAddress: p.BaseAddress + clusterOffset,
		Registers:   cpRegisters,
	}
	peripheralsList = append(peripheralsList, clusterPeripheral)
	peripheralDict[clusterPeripheral.Name] = clusterPeripheral
	p.Subtypes = append(p.Subtypes, clusterPeripheral)

	return peripheralsList
}

func processCluster(p *Peripheral, clusters []*SVDCluster, peripheralDict map[string]*Peripheral) []*Peripheral {
	var peripheralsList []*Peripheral
	for _, cluster := range clusters {
		clusterName := strings.ReplaceAll(cluster.Name, "[%s]", "")
		if cluster.DimIndex != nil {
			clusterName = strings.ReplaceAll(clusterName, "%s", "")
		}
		clusterPrefix := clusterName + "_"
		clusterOffset, err := strconv.ParseUint(cluster.AddressOffset, 0, 32)
		if err != nil {
			panic(err)
		}
		var dim, dimIncrement int
		if cluster.Dim == nil {
			// Nordic SVD have sub-clusters with another sub-clusters.
			if clusterOffset == 0 || len(cluster.Clusters) > 0 {
				peripheralsList = append(peripheralsList, processSubCluster(p, cluster, clusterOffset, clusterName, peripheralDict)...)
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
			regName := p.GroupName
			if regName == "" {
				regName = p.Name
			}
			clusterRegisters = append(clusterRegisters, parseRegister(regName, regEl, p.BaseAddress+clusterOffset, clusterPrefix)...)
		}
		sort.SliceStable(clusterRegisters, func(i, j int) bool {
			return clusterRegisters[i].Address < clusterRegisters[j].Address
		})
		if dimIncrement == -1 && len(clusterRegisters) > 0 {
			lastReg := clusterRegisters[len(clusterRegisters)-1]
			lastAddress := lastReg.Address
			if lastReg.Array != -1 {
				lastAddress = lastReg.Address + uint64(lastReg.Array*lastReg.ElementSize)
			}
			firstAddress := clusterRegisters[0].Address
			dimIncrement = int(lastAddress - firstAddress)
		}

		if !unicode.IsUpper(rune(clusterName[0])) && !unicode.IsDigit(rune(clusterName[0])) {
			clusterName = strings.ToUpper(clusterName)
		}

		p.Registers = append(p.Registers, &PeripheralField{
			Name:        clusterName,
			Address:     p.BaseAddress + clusterOffset,
			Description: cluster.Description,
			Registers:   clusterRegisters,
			Array:       dim,
			ElementSize: dimIncrement,
			ShortName:   clusterName,
		})
	}
	sort.SliceStable(p.Registers, func(i, j int) bool {
		return p.Registers[i].Address < p.Registers[j].Address
	})
	return peripheralsList
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

	peripheralDict := map[string]*Peripheral{}
	groups := map[string]*Peripheral{}

	interrupts := make(map[string]*Interrupt)
	var peripheralsList []*Peripheral

	// Some SVD files have peripheral elements derived from a peripheral that
	// comes later in the file. To make sure this works, sort the peripherals if
	// needed.
	orderedPeripherals := orderPeripherals(device.Peripherals)

	for _, periphEl := range orderedPeripherals {
		description := formatText(periphEl.Description)
		baseAddress, err := strconv.ParseUint(periphEl.BaseAddress, 0, 64)
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
			var derivedFrom *Peripheral
			if periphEl.DerivedFrom != "" {
				derivedFrom = peripheralDict[periphEl.DerivedFrom]
			} else {
				derivedFrom = groups[groupName]
			}
			p := &Peripheral{
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
			for _, subtype := range derivedFrom.Subtypes {
				peripheralsList = append(peripheralsList, &Peripheral{
					Name:        periphEl.Name + "_" + subtype.ClusterName,
					GroupName:   subtype.GroupName,
					Description: subtype.Description,
					BaseAddress: baseAddress,
				})
			}
			continue
		}

		p := &Peripheral{
			Name:        periphEl.Name,
			GroupName:   groupName,
			Description: description,
			BaseAddress: baseAddress,
			Registers:   []*PeripheralField{},
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
			p.Registers = append(p.Registers, parseRegister(regName, register, baseAddress, "")...)
		}
		peripheralsList = append(peripheralsList, processCluster(p, periphEl.Clusters, peripheralDict)...)
	}

	// Make a sorted list of interrupts.
	interruptList := make([]*Interrupt, 0, len(interrupts))
	for _, intr := range interrupts {
		interruptList = append(interruptList, intr)
	}
	sort.SliceStable(interruptList, func(i, j int) bool {
		if interruptList[i].Value != interruptList[j].Value {
			return interruptList[i].Value < interruptList[j].Value
		}
		return interruptList[i].PeripheralIndex < interruptList[j].PeripheralIndex
	})

	// Properly format the license block, with comments.
	licenseBlock := ""
	if text := formatText(device.LicenseText); text != "" {
		licenseBlock = "//     " + strings.ReplaceAll(text, "\n", "\n//     ")
		licenseBlock = regexp.MustCompile(`\s+\n`).ReplaceAllString(licenseBlock, "\n")
	}

	// Remove "-" characters from the device name because such characters cannot
	// be used in build tags. Necessary for the ESP32-C3 for example.
	nameLower := strings.ReplaceAll(strings.ToLower(device.Name), "-", "")
	metadata := &Metadata{
		File:             filepath.Base(path),
		DescriptorSource: sourceURL,
		Name:             device.Name,
		NameLower:        nameLower,
		Description:      strings.TrimSpace(device.Description),
		LicenseBlock:     licenseBlock,
	}
	if device.CPU != nil {
		metadata.HasCPUInfo = true
		metadata.CPUName = device.CPU.Name
		metadata.FPUPresent = device.CPU.FPUPresent
		metadata.NVICPrioBits = device.CPU.NVICPrioBits
	}
	return &Device{
		Metadata:       metadata,
		Interrupts:     interruptList,
		Peripherals:    peripheralsList,
		PeripheralDict: peripheralDict,
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

func addInterrupt(interrupts map[string]*Interrupt, name, interruptName string, index int, description string) {
	if _, ok := interrupts[name]; ok {
		if interrupts[name].Value != index {
			// Note: some SVD files like the one for STM32H7x7 contain mistakes.
			// Instead of throwing an error, simply log it.
			fmt.Fprintf(os.Stderr, "interrupt with the same name has different indexes: %s (%d vs %d)\n",
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
		interrupts[name] = &Interrupt{
			Name:            name,
			HandlerName:     interruptName + "_IRQHandler",
			PeripheralIndex: len(interrupts),
			Value:           index,
			Description:     description,
		}
	}
}

func parseBitfields(groupName, regName string, fieldEls []*SVDField, bitfieldPrefix string) ([]Constant, []Bitfield) {
	var fields []Constant
	var bitfields []Bitfield
	enumSeen := map[string]int64{}
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

		// The enumerated values can be the same as another field, so to avoid
		// duplication SVD files can simply refer to another set of enumerated
		// values in the same register.
		// See: https://www.keil.com/pack/doc/CMSIS/SVD/html/elem_registers.html#elem_enumeratedValues
		enumeratedValues := fieldEl.EnumeratedValues
		if enumeratedValues.DerivedFrom != "" {
			parts := strings.Split(enumeratedValues.DerivedFrom, ".")
			if len(parts) == 1 {
				found := false
				for _, otherFieldEl := range fieldEls {
					if otherFieldEl.EnumeratedValues.Name == parts[0] {
						found = true
						enumeratedValues = otherFieldEl.EnumeratedValues
					}
				}
				if !found {
					fmt.Fprintf(os.Stderr, "Warning: could not find enumeratedValue.derivedFrom of %s for register field %s\n", enumeratedValues.DerivedFrom, fieldName)
				}
			} else {
				// The derivedFrom attribute may also point to enumerated values
				// in other registers and even peripherals, but this feature
				// isn't often used in SVD files.
				fmt.Fprintf(os.Stderr, "TODO: enumeratedValue.derivedFrom to a different register: %s\n", enumeratedValues.DerivedFrom)
			}
		}

		bitfields = append(bitfields, Bitfield{
			Name:   fieldName,
			Offset: lsb,
			Mask:   (0xffffffff >> (31 - (msb - lsb))) << lsb,
		})
		fields = append(fields, Constant{
			Name:        fmt.Sprintf("%s_%s%s_%s_Pos", groupName, bitfieldPrefix, regName, fieldName),
			Description: fmt.Sprintf("Position of %s field.", fieldName),
			Value:       uint64(lsb),
		})
		fields = append(fields, Constant{
			Name:        fmt.Sprintf("%s_%s%s_%s_Msk", groupName, bitfieldPrefix, regName, fieldName),
			Description: fmt.Sprintf("Bit mask of %s field.", fieldName),
			Value:       (0xffffffffffffffff >> (63 - (msb - lsb))) << lsb,
		})
		if lsb == msb { // single bit
			fields = append(fields, Constant{
				Name:        fmt.Sprintf("%s_%s%s_%s", groupName, bitfieldPrefix, regName, fieldName),
				Description: fmt.Sprintf("Bit %s.", fieldName),
				Value:       1 << lsb,
			})
		}
		for _, enumEl := range enumeratedValues.EnumeratedValue {
			enumName := enumEl.Name
			if strings.EqualFold(enumName, "reserved") || !validName.MatchString(enumName) {
				continue
			}
			if !unicode.IsUpper(rune(enumName[0])) && !unicode.IsDigit(rune(enumName[0])) {
				enumName = strings.ToUpper(enumName)
			}
			enumDescription := formatText(enumEl.Description)
			var enumValue uint64
			var err error
			if strings.HasPrefix(enumEl.Value, "0b") {
				val := strings.TrimPrefix(enumEl.Value, "0b")
				enumValue, err = strconv.ParseUint(val, 2, 64)
			} else {
				enumValue, err = strconv.ParseUint(enumEl.Value, 0, 64)
			}
			if err != nil {
				if enumBitSpecifier.MatchString(enumEl.Value) {
					// NXP SVDs use the form #xx1x, #x0xx, etc for values
					enumValue, err = strconv.ParseUint(strings.ReplaceAll(enumEl.Value[1:], "x", "0"), 2, 64)
					if err != nil {
						panic(err)
					}
				} else {
					panic(err)
				}
			}
			enumName = fmt.Sprintf("%s_%s%s_%s_%s", groupName, bitfieldPrefix, regName, fieldName, enumName)

			// Avoid duplicate values. Duplicate names with the same value are
			// allowed, but the same name with a different value is not. Instead
			// of trying to work around those cases, remove the value entirely
			// as there is probably not one correct answer in such a case.
			// For example, SVD files from NXP have enums limited to 20
			// characters, leading to lots of duplicates when these enum names
			// are long. Nothing here can really fix those cases.
			previousEnumValue, seenBefore := enumSeen[enumName]
			if seenBefore {
				if previousEnumValue < 0 {
					// There was a mismatch before, ignore all equally named fields.
					continue
				}
				if int64(enumValue) != previousEnumValue {
					// There is a mismatch. Mark it as such, and remove the
					// existing enum bitfield value.
					enumSeen[enumName] = -1
					for i, field := range fields {
						if field.Name == enumName {
							fields = append(fields[:i], fields[i+1:]...)
							break
						}
					}
				}
				continue
			}
			enumSeen[enumName] = int64(enumValue)

			fields = append(fields, Constant{
				Name:        enumName,
				Description: enumDescription,
				Value:       enumValue,
			})
		}
	}
	return fields, bitfields
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
	return strings.ReplaceAll(r.element.Name, "[%s]", "")
}

func (r *Register) description() string {
	return formatText(r.element.Description)
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
			shortName := strings.ToUpper(strings.ReplaceAll(strings.ReplaceAll(reg.name(), "_%s", ""), "%s", ""))
			for i, j := range reg.dimIndex() {
				regAddress := reg.address() + (uint64(i) * dimIncrement)
				results = append(results, &PeripheralField{
					Name:        strings.ToUpper(strings.ReplaceAll(reg.name(), "%s", j)),
					Address:     regAddress,
					Description: reg.description(),
					Array:       -1,
					ElementSize: reg.size(),
					ShortName:   shortName,
				})
			}
			// set first result bitfield
			results[0].Constants, results[0].Bitfields = parseBitfields(groupName, shortName, regEl.Fields, bitfieldPrefix)
			results[0].HasBitfields = len(results[0].Bitfields) > 0
			for i := 1; i < len(results); i++ {
				results[i].Bitfields = results[0].Bitfields
				results[i].HasBitfields = results[0].HasBitfields
			}
			return results
		}
	}
	regName := reg.name()
	if !unicode.IsUpper(rune(regName[0])) && !unicode.IsDigit(rune(regName[0])) {
		regName = strings.ToUpper(regName)
	}
	regName = cleanName(regName)

	constants, bitfields := parseBitfields(groupName, regName, regEl.Fields, bitfieldPrefix)
	return []*PeripheralField{&PeripheralField{
		Name:         regName,
		Address:      reg.address(),
		Description:  reg.description(),
		Constants:    constants,
		Array:        reg.dim(),
		ElementSize:  reg.size(),
		ShortName:    regName,
		Bitfields:    bitfields,
		HasBitfields: len(bitfields) > 0,
	}}
}

// The Go module for this device.
func writeGo(outdir string, device *Device, interruptSystem string) error {
	outf, err := os.Create(filepath.Join(outdir, device.Metadata.NameLower+".go"))
	if err != nil {
		return err
	}
	defer outf.Close()
	w := bufio.NewWriter(outf)

	maxInterruptValue := 0
	for _, intr := range device.Interrupts {
		if intr.Value > maxInterruptValue {
			maxInterruptValue = intr.Value
		}
	}

	interruptHandlerMap := make(map[string]*Interrupt)
	var interruptHandlers []*Interrupt
	for _, intr := range device.Interrupts {
		if _, ok := interruptHandlerMap[intr.HandlerName]; !ok {
			interruptHandlerMap[intr.HandlerName] = intr
			interruptHandlers = append(interruptHandlers, intr)
		}
	}

	t := template.Must(template.New("go").Funcs(template.FuncMap{
		"bytesNeeded": func(i, j uint64) uint64 { return j - i },
		"isMultiline": isMultiline,
		"splitLine":   splitLine,
	}).Parse(`// Automatically generated file. DO NOT EDIT.
// Generated by gen-device-svd.go from {{.device.Metadata.File}}, see {{.device.Metadata.DescriptorSource}}

// +build {{.pkgName}},{{.device.Metadata.NameLower}}

// {{.device.Metadata.Description}}
//
{{.device.Metadata.LicenseBlock}}
package {{.pkgName}}

import (
	"runtime/volatile"
	"unsafe"
)

// Some information about this device.
const (
	Device       = "{{.device.Metadata.Name}}"
{{- if .device.Metadata.HasCPUInfo }}
	CPU          = "{{.device.Metadata.CPUName}}"
	FPUPresent   = {{.device.Metadata.FPUPresent}}
	NVICPrioBits = {{.device.Metadata.NVICPrioBits}}
{{- end }}
)

// Interrupt numbers.
const (
{{- range .device.Interrupts}}
	{{- if .Description}}
		{{- range .Description|splitLine}}
	// {{.}}
		{{- end}}
	{{- end}}
	IRQ_{{.Name}} = {{.Value}}
	{{- "\n"}}
{{- end}}
	// Highest interrupt number on this device.
	IRQ_max = {{.interruptMax}} 
)

// Pseudo function call that is replaced by the compiler with the actual
// functions registered through interrupt.New.
//go:linkname callHandlers runtime/interrupt.callHandlers
func callHandlers(num int)

{{- if eq .interruptSystem "hardware"}}
{{- range .interruptHandlers}}
//export {{.HandlerName}}
func interrupt{{.Name}}() {
	callHandlers(IRQ_{{.Name}})
}
{{- end}}
{{- end}}

{{- if eq .interruptSystem "software"}}
func HandleInterrupt(num int) {
	switch num {
	{{- range .interruptHandlers}}
	case IRQ_{{.Name}}:
		callHandlers(IRQ_{{.Name}})
	{{- end}}
	}
}
{{- end}}

// Peripherals.
var (
{{- range .device.Peripherals}}
	{{- if .Description}}
		{{- range .Description|splitLine}}
	// {{.}}
		{{- end}}
	{{- end}}
	{{.Name}} = (*{{.GroupName}}_Type)(unsafe.Pointer(uintptr(0x{{printf "%x" .BaseAddress}})))
	{{- "\n"}}
{{- end}}
)

`))
	err = t.Execute(w, map[string]interface{}{
		"device":            device,
		"pkgName":           filepath.Base(strings.TrimRight(outdir, "/")),
		"interruptMax":      maxInterruptValue,
		"interruptSystem":   interruptSystem,
		"interruptHandlers": interruptHandlers,
	})
	if err != nil {
		return err
	}

	// Define peripheral struct types.
	for _, peripheral := range device.Peripherals {
		if peripheral.Registers == nil {
			// This peripheral was derived from another peripheral. No new type
			// needs to be defined for it.
			continue
		}
		fmt.Fprintln(w)
		if peripheral.Description != "" {
			for _, l := range splitLine(peripheral.Description) {
				fmt.Fprintf(w, "// %s\n", l)
			}
		}
		fmt.Fprintf(w, "type %s_Type struct {\n", peripheral.GroupName)

		address := peripheral.BaseAddress
		type clusterInfo struct {
			name        string
			description string
			address     uint64
			size        uint64
			registers   []*PeripheralField
		}
		clusters := []clusterInfo{}
		for _, register := range peripheral.Registers {
			if register.Registers == nil && address > register.Address {
				// In Nordic SVD files, these registers are deprecated or
				// duplicates, so can be ignored.
				//fmt.Fprintf(os.Stderr, "skip: %s.%s 0x%x - 0x%x %d\n", peripheral.Name, register.name, address, register.address, register.elementSize)
				// remove bit fields from such register
				register.Bitfields = nil
				continue
			}

			var regType string
			switch register.ElementSize {
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
			if address < register.Address {
				bytesNeeded := register.Address - address
				if bytesNeeded == 1 {
					w.WriteString("\t_ byte\n")
				} else {
					fmt.Fprintf(w, "\t_ [%d]byte\n", bytesNeeded)
				}
				address = register.Address
			}

			lastCluster := false
			if register.Registers != nil {
				// This is a cluster, not a register. Create the cluster type.
				regType = peripheral.GroupName + "_" + register.Name
				clusters = append(clusters, clusterInfo{regType, register.Description, register.Address, uint64(register.ElementSize), register.Registers})
				regType = regType + "_Type"
				subaddress := register.Address
				for _, subregister := range register.Registers {

					if subaddress != subregister.Address {
						bytesNeeded := subregister.Address - subaddress
						subaddress += bytesNeeded
					}
					var subregSize uint64
					if subregister.Array != -1 {
						subregSize = uint64(subregister.Array * subregister.ElementSize)
					} else {
						subregSize = uint64(subregister.ElementSize)
					}
					subaddress += subregSize
				}
				if register.Array == -1 {
					lastCluster = true
				}
				address = subaddress
			}

			if register.Array != -1 {
				regType = fmt.Sprintf("[%d]%s", register.Array, regType)
			}
			fmt.Fprintf(w, "\t%s %s // 0x%X\n", register.Name, regType, register.Address-peripheral.BaseAddress)

			// next address
			if lastCluster {
				lastCluster = false
			} else if register.Array != -1 {
				address = register.Address + uint64(register.ElementSize*register.Array)
			} else {
				address = register.Address + uint64(register.ElementSize)
			}
		}
		w.WriteString("}\n")

		for _, register := range peripheral.Registers {
			regName := register.Name
			writeGoRegisterBitfieldType(w, register, peripheral.GroupName, regName)
		}

		// Define clusters
		for i := 0; i < len(clusters); i++ {
			cluster := clusters[i]
			if len(cluster.registers) == 0 {
				continue
			}

			if _, ok := device.PeripheralDict[cluster.name]; ok {
				continue
			}

			fmt.Fprintln(w)
			if cluster.description != "" {
				for _, l := range splitLine(cluster.description) {
					fmt.Fprintf(w, "// %s\n", l)
				}
			}
			fmt.Fprintf(w, "type %s_Type struct {\n", cluster.name)

			address := cluster.address

			for _, register := range cluster.registers {
				if register.Registers == nil && address > register.Address {
					// In Nordic SVD files, these registers are deprecated or
					// duplicates, so can be ignored.
					//fmt.Fprintf(os.Stderr, "skip: %s.%s 0x%x - 0x%x %d\n", peripheral.Name, register.name, address, register.address, register.elementSize)
					continue
				}
				var regType string
				switch register.ElementSize {
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
				if address < register.Address {
					bytesNeeded := register.Address - address
					if bytesNeeded == 1 {
						w.WriteString("\t_ byte\n")
					} else {
						fmt.Fprintf(w, "\t_ [%d]byte\n", bytesNeeded)
					}
					address = register.Address
				}

				lastCluster := false
				if register.Registers != nil {
					// This is a cluster, not a register. Create the cluster type.
					regType = peripheral.GroupName + "_" + register.Name
					clusters = append(clusters, clusterInfo{regType, register.Description, register.Address, uint64(register.ElementSize), register.Registers})
					regType = regType + "_Type"

					subaddress := register.Address
					for _, subregister := range register.Registers {
						if subaddress != subregister.Address {
							bytesNeeded := subregister.Address - subaddress
							subaddress += bytesNeeded
						}
						var subregSize uint64
						if subregister.Array != -1 {
							subregSize = uint64(subregister.Array * subregister.ElementSize)
						} else {
							subregSize = uint64(subregister.ElementSize)
						}
						subaddress += subregSize
					}
					if register.Array == -1 {
						lastCluster = true
					}
					address = subaddress
				}

				if register.Array != -1 {
					regType = fmt.Sprintf("[%d]%s", register.Array, regType)
				}
				fmt.Fprintf(w, "\t%s %s // 0x%X\n", register.Name, regType, register.Address-peripheral.BaseAddress)

				// next address
				if lastCluster {
					lastCluster = false
				} else if register.Array != -1 {
					address = register.Address + uint64(register.ElementSize*register.Array)
				} else {
					address = register.Address + uint64(register.ElementSize)
				}
			}
			// make sure the structure is full
			if cluster.size > (address - cluster.registers[0].Address) {
				bytesNeeded := cluster.size - (address - cluster.registers[0].Address)
				if bytesNeeded == 1 {
					w.WriteString("\t_ byte\n")
				} else {
					fmt.Fprintf(w, "\t_ [%d]byte\n", bytesNeeded)
				}
			} else if cluster.size != (address - cluster.registers[0].Address) {
				println("peripheral:", peripheral.Name, "cluster:", cluster.name, "size:", cluster.size, "struct size:", (address - cluster.registers[0].Address))
			}
			w.WriteString("}\n")

			for _, register := range cluster.registers {
				regName := register.Name
				if register.Array == -1 {
					writeGoRegisterBitfieldType(w, register, cluster.name, regName)
				}
			}
		}
	}

	// Define bitfields.
	for _, peripheral := range device.Peripherals {
		if peripheral.Registers == nil {
			// This peripheral was derived from another peripheral. Constants are
			// already defined.
			continue
		}
		fmt.Fprintf(w, "\n// Constants for %s", peripheral.Name)
		if isMultiline(peripheral.Description) {
			for _, l := range splitLine(peripheral.Description) {
				fmt.Fprintf(w, "\n// %s", l)
			}
		} else if peripheral.Description != "" {
			fmt.Fprintf(w, ": %s", peripheral.Description)
		}

		fmt.Fprint(w, "\nconst(")
		for _, register := range peripheral.Registers {
			if len(register.Constants) != 0 {
				writeGoRegisterConstants(w, register, register.Name)
			}
			if register.Registers == nil {
				continue
			}
			for _, subregister := range register.Registers {
				writeGoRegisterConstants(w, subregister, register.Name+"."+subregister.Name)
			}
		}
		w.WriteString(")\n")
	}

	return w.Flush()
}

func writeGoRegisterConstants(w *bufio.Writer, register *PeripheralField, name string) {
	w.WriteString("\n\t// " + name)
	if register.Description != "" {
		if isMultiline(register.Description) {
			for _, l := range splitLine(register.Description) {
				w.WriteString("\n\t// " + l)
			}
		} else {
			w.WriteString(": " + register.Description)
		}
	}
	w.WriteByte('\n')
	for _, bitfield := range register.Constants {
		if bitfield.Description != "" {
			for _, l := range splitLine(bitfield.Description) {
				w.WriteString("\t// " + l + "\n")
			}
		}
		fmt.Fprintf(w, "\t%s = 0x%x\n", bitfield.Name, bitfield.Value)
	}
}

func writeGoRegisterBitfieldType(w *bufio.Writer, register *PeripheralField, peripheralName, registerName string) {
	if len(register.Bitfields) == 0 {
		return
	}
	w.WriteString("\n// " + peripheralName + "." + registerName)
	if register.Description != "" {
		if isMultiline(register.Description) {
			for _, l := range splitLine(register.Description) {
				w.WriteString("\n\t// " + l)
			}
		} else {
			w.WriteString(": " + register.Description)
		}
	}
	w.WriteByte('\n')
	var bitSize int
	var maxMask uint32
	switch register.ElementSize {
	case 8:
		bitSize = 64
		maxMask = 0xffffffff
		// maxMask = 0xffffffffffffffff // TODO how to handle 64-bit fields
	case 4:
		bitSize = 32
		maxMask = 0xffffffff
	case 2:
		bitSize = 16
		maxMask = 0xffff
	case 1:
		bitSize = 8
		maxMask = 0xff
	default:
		bitSize = 32
		maxMask = 0xffffffff
	}

	typeName := fmt.Sprintf("%s_Type", peripheralName)

	for _, bitfield := range register.Bitfields {
		idxArg := ""
		regAccess := "&o." + registerName + ".Reg"
		if register.Array != -1 {
			idxArg = "idx int, "
			regAccess = "&o." + registerName + "[idx].Reg"
		}
		var funcSuffix string
		if maxMask == bitfield.Mask || registerName == bitfield.Name {
			funcSuffix = registerName
		} else {
			funcSuffix = registerName + "_" + bitfield.Name
		}
		fmt.Fprintf(w, "func (o *%s) Set%s(%s value uint%d) {\n", typeName, funcSuffix, idxArg, bitSize)
		if maxMask == bitfield.Mask {
			fmt.Fprintf(w, "\tvolatile.StoreUint%d(%s, value)\n", bitSize, regAccess)
		} else if bitfield.Offset > 0 {
			fmt.Fprintf(w, "\tvolatile.StoreUint%d(%s, volatile.LoadUint%d(%s)&^(0x%x)|value<<%d)\n", bitSize, regAccess, bitSize, regAccess, bitfield.Mask, bitfield.Offset)
		} else {
			fmt.Fprintf(w, "\tvolatile.StoreUint%d(%s, volatile.LoadUint%d(%s)&^(0x%x)|value)\n", bitSize, regAccess, bitSize, regAccess, bitfield.Mask)
		}
		w.WriteString("}\n")
		fmt.Fprintf(w, "func (o *%s) Get%s(%s) uint%d {\n", typeName, funcSuffix, idxArg, bitSize)
		if maxMask == bitfield.Mask {
			fmt.Fprintf(w, "\treturn volatile.LoadUint%d(%s)\n", bitSize, regAccess)
		} else if bitfield.Offset > 0 {
			fmt.Fprintf(w, "\treturn (volatile.LoadUint%d(%s)&0x%x) >> %d\n", bitSize, regAccess, bitfield.Mask, bitfield.Offset)
		} else {
			fmt.Fprintf(w, "\treturn volatile.LoadUint%d(%s)&0x%x\n", bitSize, regAccess, bitfield.Mask)
		}
		w.WriteString("}\n")
	}
}

// The interrupt vector, which is hard to write directly in Go.
func writeAsm(outdir string, device *Device) error {
	outf, err := os.Create(filepath.Join(outdir, device.Metadata.NameLower+".s"))
	if err != nil {
		return err
	}
	defer outf.Close()
	w := bufio.NewWriter(outf)

	t := template.Must(template.New("go").Parse(`// Automatically generated file. DO NOT EDIT.
// Generated by gen-device-svd.go from {{.File}}, see {{.DescriptorSource}}

// {{.Description}}
//
{{.LicenseBlock}}

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
__isr_vector:
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
	err = t.Execute(w, device.Metadata)
	if err != nil {
		return err
	}
	num := 0
	for _, intr := range device.Interrupts {
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
	for _, intr := range device.Interrupts {
		fmt.Fprintf(w, "    IRQ %s_IRQHandler\n", intr.Name)
	}
	return w.Flush()
}

func generate(indir, outdir, sourceURL, interruptSystem string) error {
	if _, err := os.Stat(indir); errors.Is(err, fs.ErrNotExist) {
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
