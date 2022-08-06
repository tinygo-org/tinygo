package builder

// This file parses Go-flavored assembly. Specifically, it runs Go assembly
// through `go tool asm -S` and parses the output.

import (
	"bytes"
	"debug/elf"
	"fmt"
	"go/scanner"
	"go/token"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/tinygo-org/tinygo/compileopts"
	"github.com/tinygo-org/tinygo/goenv"
	"github.com/tinygo-org/tinygo/objfile"
)

// Regular expression for machine code.
var asmBytesRegexp = regexp.MustCompile(`^	0x[0-9a-f]{4} [0-9a-f]{2} `)

type goasmReloc struct {
	offset   uint64
	reloc    objfile.Reloc
	linkName string
	addend   int64
}

// Compile the given Go assembly file to a standard object file format. Returns
// the path, any definitions or references in the Go assembly file, and any
// errors encountered in the process. The output object file is stored somewhere
// in the temporary directory.
func compileAsmFile(path, tmpdir, importPath string, config *compileopts.Config) (string, map[string]string, error) {
	references := make(map[string]string)

	// We need to know the Go version to be able to understand numeric
	// relocation types.
	_, goMinor, err := goenv.GetGorootVersion()
	if err != nil {
		return "", nil, err
	}

	// Create a temporary file to store the Go object file output.
	// We won't be using this file, but it has to be stored somewhere.
	goobjfile, err := ioutil.TempFile(tmpdir, "goasm-"+filepath.Base(path)+"-*.go.o")
	if err != nil {
		return "", nil, err
	}
	goobjfile.Close()

	// Compile the assembly file, and capture stdout.
	commandName := filepath.Join(goenv.Get("GOROOT"), "bin", "go")
	args := []string{"tool", "asm", "-S", "-p", importPath, "-o", goobjfile.Name(), "-I", filepath.Join(goenv.Get("GOROOT"), "pkg", "include"), path}
	cmd := exec.Command(commandName, args...)
	if config.Options.PrintCommands != nil {
		config.Options.PrintCommands(commandName, args...)
	}
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(cmd.Env, "GOOS="+config.GOOS(), "GOARCH="+config.GOARCH())
	err = cmd.Run()
	if err != nil {
		return "", nil, fmt.Errorf("could not invoke Go assembler: %w", err)
	}

	// Split the stdout text into symbol chunks.
	var chunk []string
	var chunks [][]string
	for {
		line, err := stdout.ReadString('\n')
		if err != nil {
			break
		}
		if !strings.HasPrefix(line, "\t") {
			// Start a new chunk.
			if len(chunk) != 0 {
				chunks = append(chunks, chunk)
			}
			chunk = nil
		}
		chunk = append(chunk, line)
	}
	if len(chunk) != 0 {
		// Add last chunk.
		chunks = append(chunks, chunk)
	}

	// Determine which output file format to use based on the architecture.
	var out objfile.ObjectFile
	switch config.GOARCH() {
	case "386":
		out = objfile.NewELF(elf.EM_386)
	case "amd64":
		out = objfile.NewELF(elf.EM_X86_64)
	case "arm":
		out = objfile.NewELF(elf.EM_ARM)
	case "arm64":
		out = objfile.NewELF(elf.EM_AARCH64)
	default:
		return "", nil, scanner.Error{
			Pos: token.Position{
				Filename: path,
			},
			Msg: fmt.Sprintf("unknown GOARCH while creating ELF file: %s", config.GOARCH()),
		}
	}

	// Parse each chunk (equivalent to a single symbol).
	localSymbols := make(map[string]struct{})
	for _, chunk := range chunks {
		header := chunk[0]
		lines := chunk[1:]
		headerFields := strings.Fields(header)
		symbolName := getSymbolName(headerFields[0], importPath)
		var section string
		switch headerFields[1] {
		case "STEXT":
			section = "text"
		case "SRODATA":
			section = "rodata"
		default:
			return "", nil, fmt.Errorf("unknown section type: %s", headerFields[1])
		}
		bind := objfile.LinkageODR
		for _, flag := range headerFields[2:] {
			if flag == "static" {
				bind = objfile.LinkageLocal
				localSymbols[symbolName] = struct{}{}
			}
		}
		chunkReferences := []string{symbolName}
		var buf []byte
		var parsedRelocs []goasmReloc
		canUseSymbol := true
		for _, line := range lines {
			switch {
			case asmBytesRegexp.MatchString(line):
				values := strings.Fields(line[8:55])
				for _, value := range values {
					n, err := strconv.ParseUint(value, 16, 8)
					if err != nil {
						return "", nil, scanner.Error{
							Pos: token.Position{
								Filename: path,
							},
							Msg: fmt.Sprintf("could not parse Go assembly: %v", err),
						}
					}
					buf = append(buf, uint8(n))
				}
			case strings.HasPrefix(line, "\trel "):
				var offset, size uint64
				var typ string
				var symaddend string
				_, err := fmt.Sscanf(line, "\trel %d+%d t=%s %s", &offset, &size, &typ, &symaddend)
				if err != nil {
					return "", nil, fmt.Errorf("cannot read relocation %s: %w", strings.TrimSpace(line), err)
				}
				if size == 0 {
					// This can happen for instructions like "CALL AX", possibly
					// as a way to signal there is a function pointer call.
					continue
				}
				index := strings.LastIndexByte(symaddend, '+')
				if index < 0 {
					return "", nil, fmt.Errorf("cannot find addend for relocation %s", strings.TrimSpace(line))
				}
				symbolName := getSymbolName(symaddend[:index], importPath)
				chunkReferences = append(chunkReferences, symbolName)
				reloc := getRelocType(typ, goMinor)
				if reloc == objfile.RelocNone {
					return "", nil, scanner.Error{
						Pos: token.Position{
							Filename: path,
						},
						Msg: fmt.Sprintf("unknown relocation type %s in relocation %#v", typ, strings.TrimSpace(line)),
					}
				}
				if reloc == objfile.RelocTLS_LE {
					// This relocation seems to be used mostly for goroutine
					// stack size checks. This is not yet supported, so don't
					// emit this symbol in the output object file.
					canUseSymbol = false
					break
				}
				var addend int64
				if config.GOARCH() == "arm" {
					// The addend is a hexadecimal number on ARM.
					addend, err = strconv.ParseInt(symaddend[index+1:], 16, 64)
					if reloc == objfile.RelocCALL {
						// It appears that the instruction is encoded in the
						// addend.
						// That seems like a bad idea, so instead write the
						// instruction back into the machine code and use a
						// conventional addend of -8 (for standard 8-byte ARM PC
						// offset).
						buf[offset+0] = byte(addend >> 0)
						buf[offset+1] = byte(addend >> 8)
						buf[offset+2] = byte(addend >> 16)
						buf[offset+3] = byte(addend >> 24)
						addend = -8
					}
				} else {
					addend, err = strconv.ParseInt(symaddend[index+1:], 10, 64)
				}
				if err != nil {
					return "", nil, fmt.Errorf("cannot read addend for relocation %s: %w", strings.TrimSpace(line), err)
				}
				parsedRelocs = append(parsedRelocs, goasmReloc{
					offset:   offset,
					reloc:    reloc,
					linkName: "__GoABI0_" + symbolName,
					addend:   addend,
				})
			}
		}

		// Only add the symbol when it is usable.
		if canUseSymbol {
			symbolIndex := out.AddSymbol("__GoABI0_"+symbolName, section, bind, buf)
			for _, reloc := range parsedRelocs {
				out.AddReloc(symbolIndex, reloc.offset, reloc.reloc, reloc.linkName, reloc.addend)
			}
			for _, name := range chunkReferences {
				references[name] = "__GoABI0_" + name
			}
		}
	}

	// Some symbols are defined as local in this assembly file but are still
	// referenced. They should not be returned as references.
	for name := range localSymbols {
		delete(references, name)
	}

	// Write output object file.
	objpath := strings.TrimSuffix(goobjfile.Name(), ".go.o") + ".o"
	err = os.WriteFile(objpath, out.Bytes(), 0o666)
	if err != nil {
		return "", nil, fmt.Errorf("failed to write object file for %s: %s", path, err)
	}
	return objpath, references, err
}

// getSymbol converts symbol names as given by `go tool asm` to those used as
// linkname (without the __GoABI0_ prefix).
func getSymbolName(name, importPath string) string {
	symbolName := name
	if strings.HasPrefix(symbolName, `"".`) {
		symbolName = importPath + "." + symbolName[3:]
	}
	return symbolName
}

// Return the symbolic relocation type given a numeric relocation type.
// Unfortunately, numeric relocation types vary by Go version.
func getRelocType(t string, goMinor int) objfile.Reloc {
	// See: https://github.com/golang/go/blob/master/src/cmd/internal/objabi/reloctype.go
	// When adding a new Go version, check whether the relocations changed.
	switch goMinor {
	case 21, 20, 19, 18:
		switch t {
		case "1", "3": // R_ADDR, R_ADDRARM64
			return objfile.RelocADDR
		case "7", "8", "9": // R_CALL, R_CALLARM, R_CALLARM64
			return objfile.RelocCALL
		case "14": // R_PCREL
			return objfile.RelocPCREL
		case "15": // R_TLS_LE
			return objfile.RelocTLS_LE
		case "41": // R_ARM64_PCREL_LDST64 (Go 1.20 only)
			return objfile.RelocPCREL_LDST64
		}
	}
	return objfile.RelocNone
}
