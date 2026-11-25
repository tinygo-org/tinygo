package builder

import (
	"debug/elf"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/tinygo-org/tinygo/compileopts"
	"github.com/tinygo-org/tinygo/interp"
)

// TestESP32S3PSRAMIRAMNoFlashRefs checks that the ESP32-S3 Octal PSRAM IRAM
// routines (which run with flash caches disabled) do not reference flash-mapped
// .text or .rodata sections (for example via runtime.memequal,
// runtime.lookupPanic, runtime.nilPanic, un-inlined helpers, or promoted
// constant tables).
func TestESP32S3PSRAMIRAMNoFlashRefs(t *testing.T) {
	if !hasBuiltinTools {
		// hasBuiltinTools is set when TinyGo is statically linked with LLVM,
		// which also implies it was built with Xtensa support.
		t.Skip("requires builtin LLVM tools with Xtensa support")
	}
	t.Parallel()

	for _, opt := range []string{"z", "s", "2", "1"} {
		t.Run("opt="+opt, func(t *testing.T) {
			t.Parallel()
			options := compileopts.Options{
				Target:                  "esp32s3-n16r8-devkit",
				Opt:                     opt,
				ExtLDFlags:              []string{"--emit-relocs"},
				Semaphore:               sema,
				InterpTimeout:           60 * time.Second,
				InterpMaxLoopIterations: interp.DefaultMaxInterpBlockEntries,
				Debug:                   true,
				VerifyIR:                true,
			}
			target, err := compileopts.LoadTarget(&options)
			if err != nil {
				t.Fatal(err)
			}
			config := &compileopts.Config{
				Options: &options,
				Target:  target,
			}
			result, err := Build("examples/psram", ".elf", t.TempDir(), config)
			if err != nil {
				t.Fatal(err)
			}
			if err := checkIRAMNoFlashRefs(result.Executable, "runtime.calibratePSRAMTiming", "runtime.setPSRAMClockDivider"); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func checkIRAMNoFlashRefs(elfPath string, fnNames ...string) error {
	f, err := elf.Open(elfPath)
	if err != nil {
		return err
	}
	defer f.Close()

	iramSec := f.Section(".iram")
	if iramSec == nil {
		return fmt.Errorf("missing .iram section in %s", elfPath)
	}
	var iramIdx int
	for i, s := range f.Sections {
		if s == iramSec {
			iramIdx = i
			break
		}
	}

	relaSec := f.Section(".rela.iram")
	if relaSec == nil {
		return fmt.Errorf("missing .rela.iram section in %s", elfPath)
	}
	relaData, err := relaSec.Data()
	if err != nil {
		return err
	}

	syms, err := f.Symbols()
	if err != nil {
		return err
	}

	type reloc struct {
		off     uint64
		rType   uint32
		target  uint64
		secName string
		symName string
	}

	isFlashSection := func(secName string) bool {
		return secName == ".text" || strings.HasPrefix(secName, ".rodata")
	}

	var allRelocs []reloc
	byOff := make(map[uint64][]reloc)
	for off := 0; off+12 <= len(relaData); off += 12 {
		rOff := uint64(f.ByteOrder.Uint32(relaData[off:]))
		rInfo := f.ByteOrder.Uint32(relaData[off+4:])
		rAdd := int32(f.ByteOrder.Uint32(relaData[off+8:]))
		rType := rInfo & 0xff
		if rType == 0 {
			continue
		}
		symIdx := int(rInfo >> 8)
		if symIdx == 0 || symIdx-1 >= len(syms) {
			continue
		}
		sym := syms[symIdx-1]
		secName := "ABS"
		if sym.Section != elf.SHN_ABS && int(sym.Section) < len(f.Sections) {
			secName = f.Sections[sym.Section].Name
		}
		target := uint64(int64(sym.Value) + int64(rAdd))
		r := reloc{
			off:     rOff,
			rType:   rType,
			target:  target,
			secName: secName,
			symName: sym.Name,
		}
		allRelocs = append(allRelocs, r)
		byOff[rOff] = append(byOff[rOff], r)
	}

	iramFuncByName := make(map[string]elf.Symbol)
	var iramFuncs []elf.Symbol
	for _, s := range syms {
		if int(s.Section) == iramIdx && s.Size > 0 && elf.ST_TYPE(s.Info) == elf.STT_FUNC {
			iramFuncByName[s.Name] = s
			iramFuncs = append(iramFuncs, s)
		}
	}

	findIRAMFunc := func(addr uint64) (elf.Symbol, bool) {
		for _, s := range iramFuncs {
			if addr >= s.Value && addr < s.Value+s.Size {
				return s, true
			}
		}
		return elf.Symbol{}, false
	}

	queue := make([]elf.Symbol, 0, len(fnNames))
	visited := make(map[string]bool)
	for _, name := range fnNames {
		s, ok := iramFuncByName[name]
		if !ok {
			return fmt.Errorf("expected IRAM function %s not found in .iram", name)
		}
		queue = append(queue, s)
		visited[s.Name] = true
	}

	resolveSymbol := func(addr uint64) string {
		for _, s := range syms {
			if s.Size > 0 && addr >= s.Value && addr < s.Value+s.Size {
				return fmt.Sprintf("%s+%#x", s.Name, addr-s.Value)
			}
		}
		return ""
	}

	var errs []string
	for len(queue) > 0 {
		fn := queue[0]
		queue = queue[1:]

		for _, r := range allRelocs {
			if r.off < fn.Value || r.off >= fn.Value+fn.Size {
				continue
			}
			if isFlashSection(r.secName) {
				targetName := r.symName
				if targetName == "" {
					targetName = resolveSymbol(r.target)
				}
				errs = append(errs, fmt.Sprintf("%s+%#x -> %s %s (%#x)", fn.Name, r.off-fn.Value, r.secName, targetName, r.target))
				continue
			}
			if r.secName != ".iram" {
				continue
			}
			if callee, ok := findIRAMFunc(r.target); ok && !visited[callee.Name] {
				visited[callee.Name] = true
				queue = append(queue, callee)
			}
			for _, lit := range byOff[r.target] {
				if isFlashSection(lit.secName) {
					targetName := lit.symName
					if targetName == "" {
						targetName = resolveSymbol(lit.target)
					}
					errs = append(errs, fmt.Sprintf("%s+%#x (via literal %#x) -> %s %s (%#x)", fn.Name, r.off-fn.Value, r.target, lit.secName, targetName, lit.target))
				} else if lit.secName == ".iram" {
					if callee, ok := findIRAMFunc(lit.target); ok && !visited[callee.Name] {
						visited[callee.Name] = true
						queue = append(queue, callee)
					}
				}
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("IRAM functions reference flash sections:\n  %s", strings.Join(errs, "\n  "))
	}
	return nil
}
