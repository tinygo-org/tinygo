package transform_test

import (
	"strings"
	"testing"

	"github.com/tinygo-org/tinygo/transform"
	"tinygo.org/x/go-llvm"
)

// Functions named keep_* must still call runtime.stringFromBytes after the
// pass because the string bytes may change or disappear before the use.
// Functions named drop_* must not.
func TestOptimizeStringFromBytesStress(t *testing.T) {
	t.Parallel()
	ctx := llvm.NewContext()
	defer ctx.Dispose()
	buf, err := llvm.NewMemoryBufferFromFile("testdata/stringfrombytes-stress.ll")
	if err != nil {
		t.Fatal(err)
	}
	mod, err := ctx.ParseIR(buf)
	if err != nil {
		t.Fatal(err)
	}
	defer mod.Dispose()
	if err := llvm.VerifyModule(mod, llvm.ReturnStatusAction); err != nil {
		t.Fatal(err)
	}

	transform.OptimizeStringFromBytes(mod)
	if err := llvm.VerifyModule(mod, llvm.ReturnStatusAction); err != nil {
		t.Fatal(err)
	}

	conversion := mod.NamedFunction("runtime.stringFromBytes")
	for fn := mod.FirstFunction(); !fn.IsNil(); fn = llvm.NextFunction(fn) {
		name := fn.Name()
		keep := strings.HasPrefix(name, "keep_")
		if !keep && !strings.HasPrefix(name, "drop_") {
			continue
		}
		t.Run(name, func(t *testing.T) {
			copies := 0
			for bb := fn.FirstBasicBlock(); !bb.IsNil(); bb = llvm.NextBasicBlock(bb) {
				for inst := bb.FirstInstruction(); !inst.IsNil(); inst = llvm.NextInstruction(inst) {
					if !inst.IsACallInst().IsNil() && inst.CalledValue() == conversion {
						copies++
					}
				}
			}
			if keep && copies == 0 {
				t.Errorf("copy removed but it is still needed:\n%s", fn.String())
			}
			if !keep && copies != 0 {
				t.Errorf("copy kept but it can be removed:\n%s", fn.String())
			}
		})
	}
}
