package transform

import (
	"path/filepath"

	"github.com/tinygo-org/tinygo/compileopts"
	"github.com/tinygo-org/tinygo/compiler/llvmutil"
	"github.com/tinygo-org/tinygo/goenv"
	"tinygo.org/x/go-llvm"
)

// CreateStackSizeLoads replaces internal/task.getGoroutineStackSize calls with
// loads from internal/task.stackSizes that will be updated after linking. This
// way the stack sizes are loaded from a separate section and can easily be
// modified after linking.
func CreateStackSizeLoads(mod llvm.Module, config *compileopts.Config) []string {
	functionMap := map[llvm.Value][]llvm.Value{}
	var functions []llvm.Value // ptrtoint values of functions
	var functionNames []string
	var functionValues []llvm.Value // direct references to functions
	for _, use := range getUses(mod.NamedFunction("internal/task.getGoroutineStackSize")) {
		if use.FirstUse().IsNil() {
			// Apparently this stack size isn't used.
			use.EraseFromParentAsInstruction()
			continue
		}
		ptrtoint := use.Operand(0)
		if _, ok := functionMap[ptrtoint]; !ok {
			functions = append(functions, ptrtoint)
			functionNames = append(functionNames, ptrtoint.Operand(0).Name())
			functionValues = append(functionValues, ptrtoint.Operand(0))
		}
		functionMap[ptrtoint] = append(functionMap[ptrtoint], use)
	}

	if len(functions) == 0 {
		// Nothing to do.
		return nil
	}

	ctx := mod.Context()
	targetData := llvm.NewTargetData(mod.DataLayout())
	defer targetData.Dispose()
	uintptrType := ctx.IntType(targetData.PointerSize() * 8)

	// Create the new global with stack sizes, that will be put in a new section
	// just for itself.
	stackSizesGlobalType := llvm.ArrayType(functions[0].Type(), len(functions))
	stackSizesGlobal := llvm.AddGlobal(mod, stackSizesGlobalType, "internal/task.stackSizes")
	stackSizesGlobal.SetSection(".tinygo_stacksizes")
	defaultStackSizes := make([]llvm.Value, len(functions))
	defaultStackSize := llvm.ConstInt(functions[0].Type(), config.StackSize(), false)
	alignment := targetData.ABITypeAlignment(functions[0].Type())
	for i := range defaultStackSizes {
		defaultStackSizes[i] = defaultStackSize
	}
	stackSizesGlobal.SetInitializer(llvm.ConstArray(functions[0].Type(), defaultStackSizes))
	stackSizesGlobal.SetAlignment(alignment)
	// TODO: make this a constant. For some reason, that incrases code size though.
	if config.Debug() {
		dibuilder := llvm.NewDIBuilder(mod)
		dibuilder.CreateCompileUnit(llvm.DICompileUnit{
			Language:  0xb, // DW_LANG_C99 (0xc, off-by-one?)
			File:      "<unknown>",
			Dir:       "",
			Producer:  "TinyGo",
			Optimized: true,
		})
		ditype := dibuilder.CreateArrayType(llvm.DIArrayType{
			SizeInBits:  targetData.TypeAllocSize(stackSizesGlobalType) * 8,
			AlignInBits: uint32(alignment * 8),
			ElementType: dibuilder.CreateBasicType(llvm.DIBasicType{
				Name:       "uintptr",
				SizeInBits: targetData.TypeAllocSize(functions[0].Type()) * 8,
				Encoding:   llvm.DW_ATE_unsigned,
			}),
			Subscripts: []llvm.DISubrange{
				{
					Lo:    0,
					Count: int64(len(functions)),
				},
			},
		})
		diglobal := dibuilder.CreateGlobalVariableExpression(llvm.Metadata{}, llvm.DIGlobalVariableExpression{
			Name: "internal/task.stackSizes",
			File: dibuilder.CreateFile("internal/task/task_stack.go", filepath.Join(goenv.Get("TINYGOROOT"), "src")),
			Line: 1,
			Type: ditype,
			Expr: dibuilder.CreateExpression(nil),
		})
		stackSizesGlobal.AddMetadata(0, diglobal)

		dibuilder.Finalize()
		dibuilder.Destroy()
	}

	// Add all relevant values to llvm.used (for LTO).
	llvmutil.AppendToGlobal(mod, "llvm.used", append([]llvm.Value{stackSizesGlobal}, functionValues...)...)

	// Replace the calls with loads from the new global with stack sizes.
	irbuilder := ctx.NewBuilder()
	defer irbuilder.Dispose()
	for i, function := range functions {
		for _, use := range functionMap[function] {
			ptr := llvm.ConstGEP(stackSizesGlobalType, stackSizesGlobal, []llvm.Value{
				llvm.ConstInt(ctx.Int32Type(), 0, false),
				llvm.ConstInt(ctx.Int32Type(), uint64(i), false),
			})
			irbuilder.SetInsertPointBefore(use)
			stacksize := irbuilder.CreateLoad(uintptrType, ptr, "stacksize")
			use.ReplaceAllUsesWith(stacksize)
			use.EraseFromParentAsInstruction()
		}
	}

	return functionNames
}
