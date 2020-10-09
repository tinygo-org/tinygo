package transform

import "tinygo.org/x/go-llvm"

func AddWASMCustomSections(mod llvm.Module, sections [][2]string) {
	for _, sec := range sections {
		mod.AddNamedMetadataOperand("wasm.custom_sections",
			mod.Context().MDNode([]llvm.Metadata{
				llvm.GlobalContext().MDString(sec[0]),
				llvm.GlobalContext().MDString(sec[1]),
			}),
		)
	}
}
