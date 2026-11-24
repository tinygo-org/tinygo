package transform

// Call sites for the heap profile on wasm, recorded in the GC stack objects
// so the runtime can walk them from runtime.stackChainStart.

import (
	"path/filepath"
	"strconv"
	"strings"

	"github.com/tinygo-org/tinygo/internal/llvmutil"
	"tinygo.org/x/go-llvm"
)

// A shadowFrame is one function in a call site, after inlining.
type shadowFrame struct {
	name, file string
	line       uint
}

// shadowSites numbers the call sites from 1, so 0 ends a stack.
type shadowSites struct {
	ids    map[string]uint32
	frames [][]shadowFrame
}

func (s *shadowSites) id(fn, call llvm.Value) uint32 {
	frames := callFrames(fn, call)
	var key strings.Builder
	for _, f := range frames {
		key.WriteString(f.name + "\x00" + f.file + "\x00" + strconv.Itoa(int(f.line)) + "\x00")
	}
	id, ok := s.ids[key.String()]
	if !ok {
		if s.ids == nil {
			s.ids = map[string]uint32{}
		}
		s.frames = append(s.frames, frames)
		id = uint32(len(s.frames))
		s.ids[key.String()] = id
	}
	return id
}

// addCallSites stores the call site before each call, and makes the stack
// object the chain start again, which keeps the chain right after a recover.
func addCallSites(mod llvm.Module, sites *shadowSites, builder llvm.Builder, stackChainStart llvm.Value, objects map[llvm.Value]llvm.Value, objectTypes map[llvm.Value]llvm.Type, createStackObject func(llvm.Value, []llvm.Value)) {
	ctx := mod.Context()
	calls := map[llvm.Value][]llvm.Value{}
	var fns []llvm.Value
	for fn := range llvmutil.Functions(mod) {
		if fn.IsDeclaration() || fn.Subprogram().C == nil {
			// Functions without debug info are compiler glue. Asyncify may run
			// their entry again on a rewind, which would break the chain.
			continue
		}
		for bb := range llvmutil.BasicBlocks(fn) {
			for inst := range llvmutil.Instructions(bb) {
				if inst.IsACallInst().IsNil() {
					continue
				}
				callee := inst.CalledValue()
				if !callee.IsAInlineAsm().IsNil() {
					continue
				}
				if !callee.IsAFunction().IsNil() && strings.HasPrefix(callee.Name(), "llvm.") {
					continue
				}
				calls[fn] = append(calls[fn], inst)
			}
		}
		if len(calls[fn]) != 0 {
			fns = append(fns, fn)
		}
	}
	for _, fn := range fns {
		if _, ok := objects[fn]; !ok {
			createStackObject(fn, nil)
		}
		object, objectType := objects[fn], objectTypes[fn]
		builder.SetInsertPointBefore(llvm.NextInstruction(object))
		sitePtr := builder.CreateGEP(objectType, object, []llvm.Value{
			llvm.ConstInt(ctx.Int32Type(), 0, false),
			llvm.ConstInt(ctx.Int32Type(), 2, false),
		}, "")
		siteType := objectType.StructElementTypes()[2]
		for _, call := range calls[fn] {
			builder.SetInsertPointBefore(call)
			builder.CreateStore(llvm.ConstInt(siteType, uint64(sites.id(fn, call)), false), sitePtr)
			builder.CreateStore(object, stackChainStart)
		}
	}
}

// callFrames returns the functions of a call site, innermost first.
func callFrames(fn, call llvm.Value) []shadowFrame {
	loc := call.InstructionDebugLoc()
	if loc.C == nil {
		return []shadowFrame{{name: fn.Name()}}
	}
	var frames []shadowFrame
	for loc.C != nil {
		frame := shadowFrame{line: loc.LocationLine()}
		if scope := loc.LocationScope(); scope.C != nil && scope.Kind() == llvm.DISubprogramMetadataKind {
			frame.name = scope.SubprogramName()
			if file := scope.ScopeFile(); file.C != nil {
				frame.file = file.FileFilename()
				if dir := file.FileDirectory(); dir != "" && !filepath.IsAbs(frame.file) {
					frame.file = filepath.Join(dir, frame.file)
				}
			}
		}
		if frame.name == "" {
			frame.name = fn.Name()
		}
		frames = append(frames, frame)
		loc = loc.LocationInlinedAt()
	}
	return frames
}

// defineShadowTables defines the call site tables the runtime declares, see
// src/runtime/shadowstack.go.
func defineShadowTables(mod llvm.Module, sites [][]shadowFrame) {
	ctx := mod.Context()
	targetData := llvm.NewTargetData(mod.DataLayout())
	defer targetData.Dispose()
	uintptrType := ctx.IntType(targetData.PointerSize() * 8)
	ptrType := llvm.PointerType(ctx.Int8Type(), 0)
	i32 := ctx.Int32Type()
	stringType := ctx.StructType([]llvm.Type{ptrType, uintptrType}, false)
	frameType := ctx.StructType([]llvm.Type{stringType, stringType, uintptrType}, false)

	stringGlobals := map[string]llvm.Value{}
	makeString := func(s string) llvm.Value {
		data, ok := stringGlobals[s]
		if !ok {
			value := ctx.ConstString(s, false)
			data = llvm.AddGlobal(mod, value.Type(), "shadow.string")
			data.SetInitializer(value)
			data.SetLinkage(llvm.PrivateLinkage)
			data.SetGlobalConstant(true)
			data.SetUnnamedAddr(true)
			stringGlobals[s] = data
		}
		return llvm.ConstNamedStruct(stringType, []llvm.Value{data, llvm.ConstInt(uintptrType, uint64(len(s)), false)})
	}

	offsets := []llvm.Value{llvm.ConstInt(i32, 0, false)}
	var frames []llvm.Value
	for _, site := range sites {
		for _, f := range site {
			frames = append(frames, llvm.ConstNamedStruct(frameType, []llvm.Value{
				makeString(f.name), makeString(f.file), llvm.ConstInt(uintptrType, uint64(f.line), false),
			}))
		}
		offsets = append(offsets, llvm.ConstInt(i32, uint64(len(frames)), false))
	}
	defineExternal(mod, "tinygo_shadow_sites", llvm.ConstArray(i32, offsets))
	defineExternal(mod, "tinygo_shadow_frames", llvm.ConstArray(frameType, frames))
	defineExternal(mod, "tinygo_shadow_nsites", llvm.ConstInt(uintptrType, uint64(len(sites)), false))
}

// defineExternal replaces the declaration of a global, if there is one, with
// a constant holding value, which may have a different type.
func defineExternal(mod llvm.Module, name string, value llvm.Value) {
	old := mod.NamedGlobal(name)
	if old.IsNil() {
		return
	}
	global := llvm.AddGlobal(mod, value.Type(), name+".def")
	global.SetInitializer(value)
	global.SetLinkage(llvm.InternalLinkage)
	global.SetGlobalConstant(true)
	old.ReplaceAllUsesWith(global)
	old.EraseFromParentAsGlobal()
	global.SetName(name)
}
