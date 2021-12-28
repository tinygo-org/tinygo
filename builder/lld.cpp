//go:build byollvm

// This file provides C wrappers for liblld.

#include <lld/Common/Driver.h>
#include <llvm/Support/Parallel.h>

extern "C" {

static void configure() {
#if _WIN64
	// This is a hack to work around a hang in the LLD linker on Windows, with
	// -DLLVM_ENABLE_THREADS=ON. It has a similar effect as the -threads=1
	// linker flag, but with support for the COFF linker.
	llvm::parallel::strategy = llvm::hardware_concurrency(1);
#endif
}

bool tinygo_link_elf(int argc, char **argv) {
	configure();
	std::vector<const char*> args(argv, argv + argc);
	return lld::elf::link(args, llvm::outs(), llvm::errs(), false, false);
}

bool tinygo_link_macho(int argc, char **argv) {
	configure();
	std::vector<const char*> args(argv, argv + argc);
	return lld::macho::link(args, llvm::outs(), llvm::errs(), false, false);
}

bool tinygo_link_mingw(int argc, char **argv) {
	configure();
	std::vector<const char*> args(argv, argv + argc);
	return lld::mingw::link(args, llvm::outs(), llvm::errs(), false, false);
}

bool tinygo_link_wasm(int argc, char **argv) {
	configure();
	std::vector<const char*> args(argv, argv + argc);
	return lld::wasm::link(args, llvm::outs(), llvm::errs(), false, false);
}

} // external "C"
